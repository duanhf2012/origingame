# GatewayService 设计

> 状态：区服快照、登录分配、断线心跳、消息安全边界、下行消息与业务消息转发边界已确认
> 更新日期：2026-08-15
> 上位文档：[OriginGame v3 总体架构设计](../总体架构设计.md)

## 1. 服务定位

| 层级 | 名称 |
| --- | --- |
| 部署程序 | `GatewayServer` |
| Origin Service | `GatewayService` |
| Go 包 | `gatewayservice` |

GatewayService 是所有区服共享的客户端长连接网关，负责：

- 按配置接入 TCP、KCP 和 WebSocket；
- 管理客户端连接、会话和登录状态；
- 验证 LoginService 签发的 Token；
- 建立客户端到当前GameService的登录路由；
- 转发客户端消息和服务端推送；
- 提供连接限流、消息限流和异常连接处理。

Gateway 不承担玩家持久化状态和具体游戏业务逻辑。客户端不能自行指定任意后端节点，实际路由目标由服务端决定。

## 2. 共享 Gateway

所有区服共用一个逻辑 Gateway 集群，不再像老版本一样为每个真实区服部署独立 Gate。

为了承载连接和实现基本高可用，集群内部可以运行多个 Gateway 实例，但当前不划分 Cell，也不建立按地域或区服隔离的 Gateway Cell 架构。Cell 化是未来容量或故障隔离需求出现后的扩展方向。

客户端登录后维持与共享 Gateway 的一条游戏长连接。首期每条已登录连接只绑定一个承载Player的GameService，不预建跨服或多后端路由切换能力。

## 3. 公网地址与监听地址

Gateway 公网地址与 GatewayServer 本地监听地址分离管理：

- MongoDB `RealAreaInfo.GateList` 保存客户端实际连接的公网地址；
- GatewayServer 部署配置保存进程实际绑定的监听地址；
- LoginService 只向客户端返回 MongoDB 中的公网地址。

公网地址和监听地址可能因为负载均衡、NAT、端口映射、反向代理或 TLS 终止而不同，不能假定二者相等。

## 4. 可配置网络协议

GatewayService按实际存在的端点配置装配Origin v3网络Server Module，至少启用一种协议：

```text
GatewayService
├── TCP Server Module
├── KCP Server Module
├── WebSocket Server Module
├── Connection / Session Manager
├── Token Verifier
└── Message Router
```

启用多个协议时，它们共用Token验证、登录状态、连接与会话管理、消息编解码、路由和服务端推送逻辑，不复制业务Handler。未配置的端点不创建Module，也不参与Ready判断。

Origin v3 网络 Module 已通过 `max_message_size` 同时限制入站和出站完整逻辑消息。Gateway 的 TCP、KCP、WebSocket 端点首期统一配置为 `4KB`；长度字段声明超限时由网络层在提交业务处理前拒绝并关闭连接，Gateway 不再重复增加应用层消息大小配置。`4KB` 包含应用层消息头和 Body，后续确有大消息时必须先确认协议场景再调整。

开发样例可以同时启用TCP `:9001`、KCP `:9002`、WebSocket `:9003`且路径为`/ws`。除监听地址和已确认的`max_message_size: 4KB`外，首期继续使用Origin网络Module的有界默认容量，不在Gateway重复暴露低价值配置。

## 5. Token 验证

GatewayService 使用按 `kid` 配置的 Ed25519 公钥本地验证 JWT，不访问 MongoDB，也不回调 LoginService。

第一版将开发公钥直接放在 GatewayServer 配置文件中，方便使用者学习和替换：

```yaml
token_verifier:
  issuer: "origingame-login"
  audience: "origingame-gateway"
  public_keys:
    - kid: "dev-key-1"
      public_key: "<base64-ed25519-public-key>"
```

`kid` 必须与 LoginServer 当前签名配置匹配。示例公钥只用于学习和本地运行；以后生产化时可以改为文件、环境变量或 Secret 注入。

必须严格检查：

1. Token 格式合法；
2. `alg` 固定为 `EdDSA`；
3. `kid` 能找到可信公钥；
4. 签名有效；
5. `ver` 是当前支持的契约版本；
6. `iss` 和 `aud` 完全匹配；
7. 当前时间位于 `nbf` 和 `exp` 允许范围内；
8. `sub` 非空。

Token 验证失败时返回 Token 错误，不立即断开连接。Token 在玩家在线期间自然到期，不强制断开已经正常建立的游戏会话。

### 5.1 区服快照

Gateway通过公共AccDBService加载`ShowAreaID -> RealAreaID`区服信息，在进程内保存只读快照：

- 首次加载必须成功且数据合法，否则GatewayService不进入Ready；
- 后台定时刷新完整新快照，校验成功后一次性原子替换；
- 单次刷新失败时继续使用上一份有效快照，并记录错误和监控；
- 后台刷新周期固定为 `1m`，单次 AccDBService RPC 使用 `5s` 超时；失败后不立即重试，等待下一个一分钟周期；
- 登录时只读取当前快照，不为每次登录访问数据库；
- 刷新Timer和退出等待由GatewayService所属的专属生命周期Module管理。

## 6. 登录状态机与重入控制

Gateway 连接登录状态定义为：

```text
Connected -> LoggingIn -> Online
                   |          |
                   +-> Connected（登录失败）
```

处理规则：

| 当前状态 | 收到登录请求 | 处理方式 |
| --- | --- | --- |
| `Connected` | 合法登录请求 | 切换为 `LoggingIn` 并发起登录流程 |
| `LoggingIn` | 与当前请求相同的 Sequence | 视为同一请求的重复发送，不再次执行登录 |
| `LoggingIn` | 不同 Sequence | 返回 `LoginInProgress` |
| `Online` | 相同账号、相同显示区服 | 幂等返回当前登录成功结果，不重复调用后端 |
| `Online` | 不同账号或不同显示区服 | 返回 `AlreadyLoggedIn` |

登录失败后恢复为 `Connected`，返回具体错误并保留连接，允许客户端修正参数或更换 Token 后重试。

每次真正开始登录时递增内部 `LoginAttemptID`。异步回调只有在连接仍有效、状态仍为 `LoggingIn` 且 `LoginAttemptID` 匹配时，才允许提交结果。进入 `Online` 后保存最近一次成功登录结果，用于相同登录请求的幂等返回。

登录流程使用请求Context统一控制`30s`总Deadline，不为连接增加独立Timer。普通网络读写仍使用网络层超时，连接频率使用通用限流。

当前不维护登录前消息白名单。消息处理器必须执行自己的会话、权限和业务状态验证；没有有效路由时不得把普通业务消息转发到后端。

### 6.1 玩家归属查找与GameService分配

Gateway持有普通进程内组件`playerownership.PlayerOwnershipStore`。它不是Origin Module或独立Service，不持有Redis连接、Timer和权威内存状态，只通过公共AccDBService RPC执行在线归属登记Script；Gateway不调用任何区服RoleDBService。

首版每条已登录连接只保存一个`GameServiceInstance`，固定指向承载当前Player的GameService，不增加多目标Map或预留槽位。

Gateway验签后使用`PlayerKey = AccountID + ShowAreaID`，从当前有效区服快照解析`RealAreaID`，调用`AssignOrGet` Script在一次Redis原子操作中查询已有GameService；没有有效归属时选择并预占当前`Loading + Online + Resident`最小且未满的GameService。Gateway不能先查询再本地选择，也不能使用各Gateway之间广播的负载快照作为权威分配依据。

GameService实例通过公共AccDBService Redis注册`RealAreaID`、Service名称、NodeID、Origin每次启动生成的`NodeSessionID`、Ready/Draining状态、加载数、在线数、驻留数、容量和带TTL租约。Origin服务发现负责RPC可达性，Redis注册表负责登录候选、负载和原子预占。

Gateway网络层生成跨全部Gateway节点、TCP/KCP/WebSocket Module和进程重启全局唯一的字符串`GatewayConnectionID`。首次分配的Redis归属保存该值，GameService进入`LOADING`和失败回滚时必须同时校验`GameService NodeID + NodeSessionID + GatewayConnectionID`。

同一次登录最多尝试三个不同GameService实例。目标不存在、请求确定尚未执行或明确返回`SERVICE_NOT_READY`、`SERVICE_DRAINING`、`SERVICE_FULL`时，Gateway使用上述条件释放`ASSIGNING`预占并排除该实例后重新分配。RPC超时或发送后断开属于状态未知，必须先查询归属；只有实例租约失效或分配状态超时后才能原子回收，不能立即选择第二个所有者。完整流程见[在线玩家归属设计](../在线玩家归属设计.md#6-gateway-原子查找分配与重试)。

归属分配结果固定为`EXISTING`、`ASSIGNED`、`WAIT`和`NO_CAPACITY`。遇到`WAIT`时每`1s`重新查询一次，且所有查询、GameService RPC和有界重试共同受本次登录`30s`总Deadline约束；到期仍无法确认时返回服务暂时不可用。

### 6.2 重复登录与顶号

Redis返回已有`ONLINE`或`RESIDENT`归属时，Gateway仍调用当前承载Player的GameService，不重新分配实例。GameService在不跨`Await`的同步执行段中完成连接校验和替换：相同`GatewayConnectionID`幂等返回；不同连接登录时，先向旧Gateway定向提交“账号已在其他连接登录”通知，再把Player绑定替换为新连接。首期不为全部玩家消息增加通用Player Key有序队列。

Gateway收到顶号请求后必须同时匹配自身`NodeID`和全局`GatewayConnectionID`：连接仍存在时先向客户端发送确定的顶号主动推送，完成写入后再关闭连接；连接不存在时幂等返回。旧Gateway通知失败不阻止GameService完成新连接接管，旧连接后续消息仍会因为连接ID不匹配而被GameService拒绝。

### 6.3 登录成功响应

客户端使用`LoginPlayerReq`发送`LoginPlayerRequest{token, show_area_id}`。Gateway不得相信客户端提供的账号或内部路由字段：`AccountID`从JWT取得，`RealAreaID`从当前区服快照取得，`GatewayNodeID`使用当前Node，`GatewayConnectionID`由网络层生成，`ExpectedGameServiceNodeSessionID`使用Redis原子分配返回的目标启动实例。Gateway再构造内部`rpcapi.LoginPlayerRequest`调用精确GameService；目标实例不匹配时按明确未执行处理，释放本次新预占并进行有界重试。

GameService的`LoginPlayer` RPC在全部玩家数据加载、生命周期回调和连接绑定成功后，返回客户端共享Protobuf `LoginPlayerResult`。该消息固定包含`RoleInfo role_info`，其他首屏确实需要的数据按业务增加独立结构字段；不要求每个Player Proxy都提供登录结果字段。

Gateway不解析、拼装或复制其中的角色业务字段。RPC成功后，Gateway把连接状态切换为`Online`，再使用`LoginPlayerRes + 原Sequence + ErrorCode OK`将`LoginPlayerResult`直接序列化为响应Body；失败时仍使用`LoginPlayerRes`，只返回对应`ErrorCode`且Body为空，不得返回部分玩家数据。

登录是明确例外：Gateway必须等待`LoginPlayer`的确定结果，才能切换连接状态、决定是否重试分配并回复原登录`Sequence`。因此登录不改为GameService独立反向推送结果。

## 7. 统一客户端消息协议

TCP、KCP 和 WebSocket 使用相同的应用层消息协议，Body 默认采用 Protobuf。第一版不引入 JSON Body 或动态 CodecType 字段。

### 7.1 协议边界

传输帧、应用层请求、响应、主动推送、MessageID、Sequence和ErrorCode的唯一契约见[客户端协议与生成设计](../protocols/客户端协议与生成设计.md#21-应用层封包)。Gateway只校验通用消息头、连接状态和消息长度，不理解普通玩家业务MessageID或Body类型；长度越界按协议错误处理，未知普通玩家MessageID由GameService实例化Router拒绝并回复业务错误。

### 7.2 连接级消息限流

Gateway 对每条客户端连接独立统计完整入站消息，登录消息、心跳和普通业务消息使用同一限流规则，不再增加登录专用频率限制：

- 使用真实系统时间划分固定 `10s` 窗口；
- 每个窗口最多接收 `200` 条消息，相当于长期平均每秒 `20` 条；
- 计数只保存在连接对象中，窗口切换时直接重置，不创建逐消息 Timer、时间戳队列或全局限流状态；
- 收到窗口内第 `201` 条消息时判定为超频并关闭连接，避免继续转发或生成错误响应放大负载。

该限制允许正常业务在窗口内短时集中发送，不使用按瞬时速率判断的令牌桶。登录流程仍由第6节状态机保证同一连接同时只有一个登录请求正在执行。

### 7.3 GameService到Gateway的统一下行RPC

普通玩家业务由GameService决定回复或推送的`MessageID`、`ErrorCode`和Protobuf Body；Gateway不理解业务消息，只负责校验连接、编码协议头和写入网络。首期不定义通用`PlayerMessageResult`作为普通消息RPC返回值。

Gateway转发普通玩家消息到GameService时只携带`GatewayConnectionID`、`MessageID`、`Sequence`和原始`Body`，不重复携带AccountID、ShowAreaID或GatewayNodeID。GameService通过ConnectionID索引直接取得Player；当前Gateway Node由Player登录时保存的连接关系确定，业务回复由Player通过该关系精确路由回来。登录、重连和顶号登录仍是例外，`LoginPlayerRequest`显式携带新`GatewayNodeID + GatewayConnectionID`以绑定Player。

```go
// ClientMessage描述Gateway需要写入客户端连接的一条消息。
type ClientMessage struct {
	MessageID       commonpb.MessageID // 客户端协议消息ID
	Sequence        uint32             // 非0为请求响应；0为服务端主动推送
	ErrorCode       commonpb.ErrorCode // 仅请求响应使用；主动推送必须为OK
	Body            []byte             // 已序列化的Protobuf Body
	CloseAfterWrite bool               // 消息完成写入后关闭连接
}

// SendClientMessageRequest指定接收下行消息的Gateway连接。
type SendClientMessageRequest struct {
	GatewayConnectionID string        // 全局唯一客户端连接ID
	Message             ClientMessage // 待发送消息
}

//origin:rpc
type GatewayService interface {
	// SendClientMessage向指定客户端连接发送消息。
	SendClientMessage(context.Context, SendClientMessageRequest) error

	// CloseClientConnection关闭指定客户端连接；连接不存在时幂等成功。
	CloseClientConnection(context.Context, CloseClientConnectionRequest) error
}

// CloseClientConnectionRequest指定需要关闭的客户端连接。
type CloseClientConnectionRequest struct {
	GatewayConnectionID string // 全局唯一客户端连接ID
}
```

调用方使用已确认的精确Node路由选中`GatewayNodeID`，请求内不重复携带NodeID。GameService只传递业务消息字段，不传入已经编码的应用层协议头。

`SendClientMessageRequest`和`CloseClientConnectionRequest`使用普通Go结构，由`origingen`生成逐字段静态Codec。客户端业务Body已经是Protobuf字节，不再为内部RPC外层包装一层Protobuf；高频路径直接计算准确大小并写入最终RPC Buffer。实现时必须为典型小消息和1KB消息保留静态Go Codec与Protobuf包装的对比Benchmark。

下行规则：

- `Sequence != 0`时编码响应头，`ErrorCode`可为非零；
- `Sequence == 0`时编码主动推送，`ErrorCode`必须为`OK`；
- 顶号原因放入专用推送Body，并使用`CloseAfterWrite = true`；
- Gateway只在当前Node上的全局唯一`GatewayConnectionID`仍存在时写入；连接已不存在时幂等忽略；
- `CloseAfterWrite`必须在整条消息完成网络写入后生效，不能先关闭连接。

GameService需要无消息直接关闭连接时，使用精确`GatewayNodeID`调用生成的`NotifyCloseClientConnection`。Gateway只按全局唯一`GatewayConnectionID`查找并关闭连接；连接已不存在时幂等忽略。顶号仍使用带`CloseAfterWrite = true`的`SendClientMessage`，确保客户端先收到顶号推送。

### 7.4 断线通知与玩家逻辑心跳

Gateway检测到客户端连接关闭后，调用精确GameService的`NotifyPlayerDisconnected`。该通知只携带`GatewayConnectionID`，不携带`AccountID`、`ShowAreaID`、`GatewayNodeID`或断线原因；GameService通过连接索引直接查找Player，并按连接ID幂等校验。

客户端在线后每5秒发送一次`PlayerHeartbeatReq`。Gateway按普通玩家消息转发，GameService更新Player最后心跳真实系统时间；成功且没有Body时回复`MessageID = Ok`。GameService超过15秒未收到心跳时，先把Player转为`Resident`并执行离线流程，再调用`NotifyCloseClientConnection`关闭Gateway连接。

Gateway进程崩溃时不增加Gateway实例租约或NodeSessionID绑定：客户端心跳停止后，GameService最多15秒主动完成本地断线处理。Gateway关闭回调产生的迟到`NotifyPlayerDisconnected`按连接ID幂等忽略。

### 7.5 Sequence与迟到响应

Sequence只用于客户端请求和服务端响应的关联，不作为通用业务幂等键。Gateway不为全部玩家请求维护Pending Map，也不按Sequence缓存或去重普通业务请求；登录仍使用第6节已经确认的连接状态机处理重复请求。

客户端负责维护当前连接尚未完成的Sequence，同一连接内不得复用仍处于Pending状态的值。客户端请求超时后丢弃对应Pending，后续收到不存在的迟到响应直接忽略。连接关闭后该连接全部Sequence失效，新连接重新从1开始；GameService回复还会校验全局唯一GatewayConnectionID，因此旧连接响应不会投递到重连后的新连接。

## 8. Ready 条件

GatewayServer 只有在 GatewayService 完成以下准备后才能进入 Ready：

- Gateway 本地监听配置已经完成校验；
- Token 验签公钥及当前支持的 `kid` 已加载；
- 已通过公共AccDBService加载并校验首份区服快照；
- Gateway自身登录入口、通用协议解析器和连接状态机已经完成初始化；
- 连接、会话、登录状态和路由组件已经初始化；
- playerownership.PlayerOwnershipStore使用的公共AccDBService RPC已可用，并已通过AccDBService完成Redis登记Script和初始可写性探测；GatewayService不直接组合Redis Module，也不依赖区服RoleDBService；
- 至少配置一个网络端点，且全部已配置的Server Module均已成功监听；
- 访问后端服务所需的服务发现或其他必需依赖已经可用。

未配置的网络协议不创建Module。任意已配置的网络Module初始化或绑定失败时，整个GatewayService不进入Ready，已经启动的资源按逆序关闭。

## 9. 更新与排空

Gateway 持有客户端长连接，程序更新采用连接排空：

1. 旧 Gateway 实例停止接收新连接；
2. 已建立连接继续保留一段排空时间；
3. 必要时通知客户端使用有效 Token 快速重连其他 Gateway 实例；
4. 连接清理完成后退出旧实例。

共享 Gateway 解决的是后端路由变化时客户端不需要换连接，不承诺 Gateway 实例自身升级或故障时连接永不中断。
