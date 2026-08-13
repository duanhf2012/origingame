# GatewayService 设计

> 状态：当前阶段设计基线
> 更新日期：2026-08-12
> 上位文档：[OriginGame v3 总体架构设计](../总体架构设计.md)

## 1. 服务定位

| 层级 | 名称 |
| --- | --- |
| 部署程序 | `GatewayServer` |
| Origin Service | `GatewayService` |
| Go 包 | `gatewayservice` |

GatewayService 是所有区服共享的客户端长连接网关，负责：

- 同时接入 TCP、KCP 和 WebSocket；
- 管理客户端连接、会话和登录状态；
- 验证 LoginService 签发的 Token；
- 建立和切换客户端到后端服务的路由；
- 转发客户端消息和服务端推送；
- 提供连接限流、消息限流和异常连接处理。

Gateway 不承担玩家持久化状态和具体游戏业务逻辑。客户端不能自行指定任意后端节点，实际路由目标由服务端决定。

## 2. 共享 Gateway

所有区服共用一个逻辑 Gateway 集群，不再像老版本一样为每个真实区服部署独立 Gate。

为了承载连接和实现基本高可用，集群内部可以运行多个 Gateway 实例，但当前不划分 Cell，也不建立按地域或区服隔离的 Gateway Cell 架构。Cell 化是未来容量或故障隔离需求出现后的扩展方向。

客户端登录后维持与共享 Gateway 的一条游戏长连接。玩家在不同后端业务之间切换时，优先由 Gateway 切换服务端路由，不要求客户端重新建立区服 Gate 连接。

跨服战斗当前不是最小版本必须实现的服务，但 Gateway 路由设计必须允许后续在不更换客户端连接的情况下切换后端目标。

## 3. 公网地址与监听地址

Gateway 公网地址与 GatewayServer 本地监听地址分离管理：

- MongoDB `RealAreaInfo.GateList` 保存客户端实际连接的公网地址；
- GatewayServer 部署配置保存进程实际绑定的监听地址；
- LoginService 只向客户端返回 MongoDB 中的公网地址。

公网地址和监听地址可能因为负载均衡、NAT、端口映射、反向代理或 TLS 终止而不同，不能假定二者相等。

## 4. 同时监听三种协议

一个 GatewayService 实例同时装配并启动三个 Origin v3 网络 Server Module：

```text
GatewayService
├── TCP Server Module
├── KCP Server Module
├── WebSocket Server Module
├── Connection / Session Manager
├── Token Verifier
└── Message Router
```

老版本虽然兼容三种协议，但单个 `GateService` 实例通过互斥分支只启用其中一种；v3 明确调整为同一个 GatewayService 同时监听三种协议。

三种协议共用 Token 验证、登录状态、连接与会话管理、消息编解码、路由和服务端推送逻辑。

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

当前不增加独立的登录超时定时器，使用网络层读写超时和通用限流管理连接生命周期。

当前不维护登录前消息白名单。消息处理器必须执行自己的会话、权限和业务状态验证；没有有效路由时不得把普通业务消息转发到后端。

## 7. 统一客户端消息协议

TCP、KCP 和 WebSocket 使用相同的应用层消息协议，Body 默认采用 Protobuf。第一版不引入 JSON Body 或动态 CodecType 字段。

### 7.1 传输层帧

TCP 和 Origin v3 KCP 均为字节流，由网络模块处理长度帧；WebSocket 原生具有 Message 边界：

```text
TCP:       [uint32 Length][ApplicationPacket]
KCP:       [uint32 Length][ApplicationPacket]
WebSocket:                 [ApplicationPacket]
```

TCP/KCP 的 `Length` 使用四字节无符号整数和大端序，表示后续 `ApplicationPacket` 的字节数，不包含长度字段自身。长度帧由 Origin 网络模块添加和剥离，不进入 Gateway 业务消息头。

### 7.2 请求消息

```text
[uint16 MessageID][uint32 Sequence][uint8 Flags][Protobuf Body]
```

### 7.3 响应消息

```text
[uint16 MessageID][uint32 Sequence][uint8 Flags][int32 ErrorCode][Protobuf Body]
```

### 7.4 服务端主动推送

主动推送采用请求消息相同的外层结构，并固定 `Sequence = 0`。

字段规则：

- `MessageID` 使用 `uint16`，请求和响应使用不同的 MessageID；
- `Sequence` 使用 `uint32`，由客户端请求生成，响应原样返回；
- `Flags` 当前只定义 bit 0 为 Body 压缩标识，bit 1～7 保留且必须为 0；
- `ErrorCode` 只存在于响应消息，`0` 表示成功；
- `ErrorCode` 使用共享 Protobuf 枚举，见 [错误码设计](../protocols/错误码设计.md)；
- `Body` 默认使用 MessageID 对应的 Protobuf 类型序列化，可以为空；
- 压缩只作用于 Body，不作用于消息头；
- Gateway 根据编译期注册的 MessageID 元数据判断消息方向和 Body 类型；
- 未知 MessageID、非法 Flags、长度越界或解压失败按协议错误处理。

压缩算法及启用阈值尚未确定，第一版可以只预留 Flags 定义而不实际启用压缩。

## 8. Ready 条件

GatewayServer 只有在 GatewayService 完成以下准备后才能进入 Ready：

- Gateway 本地监听配置已经完成校验；
- Token 验签公钥及当前支持的 `kid` 已加载；
- 消息协议注册表已经完成并冻结；
- 连接、会话、登录状态和路由组件已经初始化；
- Redis `RouteStore` 已连接，所需原子 Function 版本已校验，并完成初始可写性探测；
- TCP、KCP、WebSocket 三个 Server Module 均已成功监听；
- 访问后端服务所需的服务发现或其他必需依赖已经可用。

三种协议均为当前必需监听。任意一个网络模块初始化或绑定失败时，整个 GatewayService 不进入 Ready，已经启动的部分必须随启动失败一起关闭。

## 9. 更新与排空

Gateway 持有客户端长连接，程序更新采用连接排空：

1. 旧 Gateway 实例停止接收新连接；
2. 已建立连接继续保留一段排空时间；
3. 必要时通知客户端使用有效 Token 快速重连其他 Gateway 实例；
4. 连接清理完成后退出旧实例。

共享 Gateway 解决的是后端路由变化时客户端不需要换连接，不承诺 Gateway 实例自身升级或故障时连接永不中断。

## 10. 实现前待确定

1. Gateway 登录请求与响应 Protobuf 的最终字段和错误码。
2. `ShowAreaId -> RealAreaId` 映射由 Gateway 自行预加载还是由公共 AreaDirectory Module 提供。`RealAreaId -> GameService` 不再调用 CenterService，统一通过 Redis `RouteStore` 的原子协调接口完成，见 [在线玩家路由设计](../在线玩家路由设计.md)。
3. Gateway 多实例时的 `GatewayInstanceID + ConnectionID` 定义，以及后端如何定向推送到正确实例。
4. 初始路由表采用单一默认路由还是从第一版开始预留命名路由。
5. MessageID 的分配规则、请求与响应对应关系，以及协议注册代码如何生成。
6. Sequence 的生成、回绕、重复响应和超时后迟到响应处理规则。
7. 消息大小上限、连接级限流、登录尝试限流及非法协议关闭条件。
8. Flags 压缩的具体算法和阈值；第一版是否只预留标志但关闭压缩。
9. TCP、KCP、WebSocket 的监听配置结构及开发环境默认端口。
