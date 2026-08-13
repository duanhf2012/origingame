# LoginService 设计

> 状态：MVP 已实现；HTTP Module 与 DBService 接入调整已确认，待重构
> 更新日期：2026-08-13
> 上位文档：[OriginGame v3 总体架构设计](../总体架构设计.md)

## 1. 服务定位

| 层级 | 名称 |
| --- | --- |
| 部署程序 | `OriginGame`（统一程序，按 Node 配置启动） |
| Origin Service | `LoginService` |
| Go 包 | `loginservice` |

`LoginService` 是客户端 HTTP 登录入口，负责：

- 接收客户端登录请求；
- 调用可替换的平台 SDK 鉴权实现；
- 创建或查询游戏账号；
- 签发游戏 Token；
- 加载并返回显示区服及 Gateway 公网地址。

不再沿用老版本 `HttpGateService` 名称，因为该服务不管理游戏长连接，也不承担 Gateway 职责。

## 2. 登录请求

HTTP 登录对外固定使用 JSON，不使用 Protobuf。服务端使用普通 Go 请求结构接收 JSON；该内部结构不属于客户端长连接 Protobuf 协议。

请求字段参照老版本 `origingame` 的 `LoginInfo`，当前不扩展新的客户端字段：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `PlatType` | `LoginType` | 登录平台类型 |
| `PlatId` | `string` | 玩家在登录平台的身份标识 |
| `AccessToken` | `string` | 平台 SDK 或第三方登录凭证 |
| `GameId` | `string` | 游戏或平台侧游戏标识 |
| `UserName` | `string` | 登录用户名，按具体登录类型使用 |

老版本 `LoginInfo.LoginCheckTime` 由服务端收到请求后赋值，用于内部验证流程，不作为客户端提交字段。

平台 SDK 鉴权当前只保留可替换接口和 `TODO`。第一版参照老版本的开发方式，只校验平台类型、`PlatId` 等基础参数，不向第三方平台验证 `AccessToken`。使用者按实际发行平台实现或替换 SDK 鉴权逻辑。

该实现只用于学习、开发和最小示例，必须在代码和配置中明确标注未执行真实 SDK 鉴权，不能把它描述为生产安全登录。

## 3. 登录响应

登录接口固定为：

```text
POST /api/v1/login
```

验证成功后，LoginService 返回：

- 一条游戏 Token 字符串；
- 可见区服列表；
- 每个区服对应的 Gateway 公网地址。

客户端不需要理解 Token 内部结构，只保存完整 Token 字符串，并在连接 Gateway 时原样提交。

当前不返回最近登录区服，也不提供默认区服。区服列表直接作为 JSON 数组返回，不延续老版本 `AreaGate string` 的二次 JSON 编码。

成功响应结构：

```json
{
  "ECode": 0,
  "Token": "jwt-token-string",
  "AreaList": [
    {
      "ShowAreaId": 1,
      "AreaName": "体验1服",
      "ServerMark": 0,
      "ServerStatus": 0,
      "OpenTime": 0,
      "GateList": [
        {"Protocol": "tcp", "Address": "127.0.0.1:9001"},
        {"Protocol": "kcp", "Address": "127.0.0.1:9002"},
        {"Protocol": "websocket", "Address": "ws://127.0.0.1:9003/ws"}
      ]
    }
  ]
}
```

失败响应只返回错误码，不返回 Token 和区服列表：

```json
{
  "ECode": 1003
}
```

`RealAreaId` 是服务端内部路由信息，不返回客户端。HTTP 层错误码和 Gateway 二进制协议错误码共用一份 Protobuf `ErrorCode` 枚举，见 [错误码设计](../protocols/错误码设计.md)。

### 3.1 HTTP Module 职责边界

`httpapi.Module` 是 LoginService 对外 HTTP 边界，负责：

- HTTP Server 的 Setup、启动、停止和监听资源生命周期；
- 注册 LoginService 的 HTTP 路由和 HTTP Middleware；
- 定义并解析 JSON DTO；
- 实现各路由对应的 Handler；
- 将 HTTP 状态码、业务错误码和响应 DTO 写回客户端；
- 协调只由 HTTP 请求触发的登录流程。

`LoginService` 根包只负责顶层配置聚合、业务能力创建、Module 装配、启动顺序以及真正跨多个子包的协调流程。它不保存具体 HTTP Handler，也不向 `httpapi.NewModule` 逐个传入 `loginHandler`、`areaHandler` 等路由函数。新增同类 HTTP 接口时，应主要修改 `httpapi` 包，不应导致 `LoginService` 随接口数量增加 Handler 字段、构造参数或无业务价值的转发方法。

`httpapi.Module` 的 Handler 通过 Module 方法注册：

```go
func (module *Module) OnInit() error {
    // 完成 Server Setup 和 Middleware 注册。
    module.SafePOST("/api/v1/login", module.login)
    return nil
}
```

构造 `httpapi.Module` 时注入完成登录流程所需的最小业务能力，例如平台鉴权、账号 Repository、Token 签发、区服快照和登录限流；注入的是能力，不是属于 HTTP Module 自身的 Handler。具体依赖优先直接使用现有最小类型，只有参数数量或可读性确有需要时才使用依赖聚合结构，不为未来接口预建 Controller、UseCase 或注册器层。

当前登录请求处理顺序保持不变：

```text
httpapi.Module
    -> 解析并校验 LoginRequest
    -> 执行登录限流
    -> 调用平台鉴权能力
    -> 调用账号 Repository
    -> 调用 Token 签发能力
    -> 读取区服快照
    -> 返回 LoginResponse
```

HTTP Handler 可以调用 Module 继承的 `Await`，在数据库、Redis 等真实 I/O 期间协作式释放所属 Service 的执行权；请求仍通过 `SafePOST` 进入所属 Service 工作协程，不得在 Handler 中创建无法停止和等待的业务 goroutine。

文件按具体职责拆分：

```text
service/loginservice/httpapi/
├── httpmodule.go       # Module、HTTP 生命周期、依赖和路由/Middleware 注册
├── loginhandler.go     # POST /api/v1/login Handler 及其专属辅助逻辑
└── logindto.go         # 登录请求和响应 JSON DTO
```

如果后续出现新的 HTTP 业务域，应先判断它是否仍属于 LoginService 对外登录边界。属于同一边界时在 `httpapi` 内按具体 Handler 和 DTO 文件扩展；具有不同监听地址、鉴权、安全策略或生命周期的 GM、运维和内部管理接口，应建立独立边界 Module，不与公开登录接口混合。

## 4. 登录路径

客户端首次登录或者游戏 Token 已失效时，通过 LoginService 完成平台鉴权。游戏 Token 在有效期内时，客户端直接连接 Gateway，由 Gateway 本地验签，不重复执行 HTTP 登录和平台 SDK 鉴权。

```text
游戏 Token 有效   -> 直接连接 Gateway
游戏 Token 已失效 -> 请求 LoginService -> 平台鉴权 -> 签发新 Token
```

当前不要求客户端每次进入区服前额外获取一次性 JoinTicket。以后如果排队、跨服实例或高安全场景需要，再作为独立机制增加。

## 5. Token 契约

游戏 Token 采用 Ed25519 非对称签名的 JWT：

- LoginService 持有私钥并签发 Token；
- GatewayService 只持有公钥并在本地验签；
- Gateway 验证 Token 时不查询 MongoDB，也不调用 LoginService；
- JWT Header 使用 `kid` 标识签名密钥，为密钥轮换保留能力；
- Token 默认有效期为 24 小时，准确时长允许通过安全配置调整。

JWT Header 和 Claims 只是服务端生成及验证 Token 时使用的内部结构。JWT 编码后是一条字符串，客户端不会收到两个 JSON 对象。

### 5.1 JWT Header

| 字段 | 值或含义 |
| --- | --- |
| `typ` | `JWT` |
| `alg` | 固定为 `EdDSA` |
| `kid` | 当前签名密钥标识 |

### 5.2 JWT Claims

| 字段 | 含义 |
| --- | --- |
| `ver` | Token 契约版本 |
| `iss` | 签发者，固定为 OriginGame LoginService |
| `aud` | 使用范围，固定为 OriginGame GatewayService |
| `sub` | 游戏账号唯一标识 AccountID |
| `pt` | 登录平台类型 |
| `iat` | 签发时间 |
| `nbf` | Token 开始生效时间 |
| `exp` | Token 过期时间 |
| `jti` | Token 唯一标识，使用安全随机值 |

Token 不包含平台 `AccessToken`、密码、SDK 凭证、区服 ID、Gateway 地址和 GameService 地址。JWT 用于签名校验而不是数据加密，Claims 中不得写入敏感数据。

第一版只要求自然过期、Gateway 本地验签和密钥轮换能力，不引入 Token 黑名单、Redis 在线校验或每次连接回调 LoginService。

### 5.3 开发密钥配置

第一版为了方便使用者学习，将 Token 密钥相关内容直接放在配置文件中：

```yaml
token:
  issuer: "origingame-login"
  audience: "origingame-gateway"
  expire: 24h
  active_kid: "dev-key-1"
  private_key: "<base64-ed25519-private-key>"
```

GatewayServer 配置相同 `kid` 对应的公钥。示例配置中的开发密钥只用于学习和本地运行，使用者可以自行替换。以后生产化时可以把私钥来源改为 Secret、环境变量或密钥文件，但当前不增加这层复杂度。

## 6. 数据访问设计

LoginService 不再直接组合 MongoDB 和 Redis Module，账号、区服和登录限流所需的全部 MongoDB/Redis 操作统一通过 AccDBService RPC 执行。AccDBService 是业务无关 DBService 模板面向账号数据域的实际实例，固定账号数据库名并集中持有 MongoDB、Redis 连接池，见 [DBService 设计](DBService设计.md)。

数据库操作仍由 LoginService 内部 Repository 封装，避免 HTTP Handler 散布集合名、BSON 查询和 Redis 命令：

```text
LoginService
├── PlatformAuthenticator
├── TokenIssuer
├── AccountStore
├── AreaCatalogStore
└── AccDBService RPC Client
```

- `AccountStore` 负责账号查询、原子创建和历史区服更新；
- `AreaCatalogStore` 负责加载真实区服和显示区服；
- 登录限流组件负责构造 Redis 原子操作并通过 AccDBService 执行；
- Repository 和限流组件选择 AccDBService，并对同一业务身份使用一致的非空 `dispatch_key` 和 `Route(key)`；DBService 只执行操作，不理解登录业务。

MongoDB 集合名、BSON 文档结构及集合内嵌字段类型统一定义在仓库级 `internal/mongodb`，不再增加 `collection` 子目录。该包是 OriginGame 各 Service 共享的持久化契约，不包含 Client、Repository、查询更新、生命周期或业务逻辑。LoginService 的 `account`、`area` 包负责 Repository、数据校验以及持久化结构到业务视图的转换；HTTP DTO 不复用 BSON 文档结构。

```text
internal/mongodb/
├── account.go
├── realareainfo.go
└── showareainfo.go
```

## 7. MongoDB 基础表

当前固定使用以下三张基础表：

### 7.1 Account

保存账号身份、平台类型、历史区服及账号阶段必要信息。

`Account._id` 使用 MongoDB `ObjectID`，由 LoginService 在创建账号时生成；对外表示为 24 位十六进制字符串，并作为 JWT `sub`。不再使用 `PlatId` 或 `PlatId + PlatType` 作为账号主键。

平台身份使用 `(PlatType, PlatId)` 复合唯一索引：

```text
Account._id                 -> 游戏内部稳定 AccountID
Account.PlatType + PlatId   -> 外部平台身份唯一约束
JWT.sub                     -> Account._id.Hex()
```

这样平台字段不参与游戏内部主键，后续修改平台接入方式时不会改变 AccountID。第一版一个 Account 对应一个平台身份；如果以后需要多平台账号绑定，再增加独立身份表，不在当前三张基础表中预建。

账号不存在时使用 `FindOneAndUpdate + upsert + $setOnInsert + ReturnDocument(After)` 原子创建。并发请求触发唯一键冲突时，重新按 `(PlatType, PlatId)` 查询现有账号，不使用“先查询、再插入”的竞态流程。

当前不返回最近登录区服和默认区服，`AreaHis` 不进入第一版登录响应。只有玩家在 GameService 完成角色加载并成功进入真实区服后，才记录该真实区服的最近进入时间；HTTP 登录成功和 Gateway 验签成功均不得提前更新。具体写入者随 GameService 数据职责一并设计。JWT 本身不保存到 Account 表。

### 7.2 RealAreaInfo

保存真实区服以及客户端实际访问的 Gateway 公网地址。`GateList` 元素从老版本的字符串调整为协议化结构：

```json
{
  "_id": 1,
  "GateList": [
    {"Protocol": "tcp", "Address": "game.example.com:9001"},
    {"Protocol": "kcp", "Address": "game.example.com:9002"},
    {"Protocol": "websocket", "Address": "wss://game.example.com/ws"}
  ]
}
```

当前所有区服共享 Gateway，不同 `RealAreaInfo` 可以配置相同的 `GateList`。当前不增加 GatewayGroup 或 Cell 数据模型。

### 7.3 ShowAreaInfo

保存客户端显示区服及其到真实区服的映射。

三张表继续使用老版本的数据关系：

```text
Account.AreaHis
ShowAreaInfo.RealAreaId
    -> RealAreaInfo.GateList
```

三张表除上述已经确定的主键和平台身份唯一索引外，其余最终字段、BSON 命名、默认数据和索引仍待确定。

## 8. 区服列表快照

LoginService 启动时从 MongoDB 查询 `RealAreaInfo` 和 `ShowAreaInfo`，完成关联校验并生成可直接返回给客户端的不可变内存快照。登录请求读取快照，不在每次登录时重复查询完整区服数据。

进入 Ready 后采用配置化的定时轮询重新加载两张区服表。区服数据量小、变更频率低，当前不引入 MongoDB Change Stream、Redis 或独立配置中心。

```yaml
area:
  refresh_interval: 30s
```

`refresh_interval` 必须为正数。每次刷新都读取完整的 `RealAreaInfo` 和 `ShowAreaInfo`，在临时对象中完成关联和校验，全部成功后再原子替换当前快照。

进入 Ready 后刷新失败时保留最后一次完整、有效的快照，不得用失败或不完整的数据覆盖当前快照。

启动时查询不到任何有效区服，或者显示区服无法关联真实区服时，LoginService 启动失败并退出。运行期间刷新得到空列表或无效关联时视为刷新失败，继续使用上一份有效快照。

第一版只提供自动定时刷新，不提供内部手动刷新区服列表接口。后续只有在运维场景证明有必要时再增加，并单独设计认证、审计和并发刷新规则。

## 9. 登录限流

所有限流维度必须分别支持配置和关闭。按次数限制的规则统一采用滑动时间窗口，不使用固定整点窗口，避免相邻窗口边界允许瞬间通过接近两倍请求。

第一版建议配置：

```yaml
login_rate_limit:
  enabled: true

  ip:
    enabled: true
    windows:
      - duration: 1s
        max_requests: 10
      - duration: 10s
        max_requests: 30
      - duration: 1m
        max_requests: 120

  identity:
    enabled: true
    windows:
      - duration: 10s
        max_requests: 5
      - duration: 5m
        max_requests: 20

  auth_failure:
    enabled: true
    window: 5m
    max_failures: 5
    cooldown: 30s

  concurrency:
    enabled: true
    max_in_flight: 256
```

规则说明：

- `ip` 对单个来源 IP 的全部登录请求使用短、长两个滑动窗口；短窗口允许正常客户端短时重试，长窗口限制持续请求；
- `identity` 在鉴权前使用 `IP + SHA256(PlatType, PlatId)` 作为键，避免攻击者只凭公开的 `PlatId` 锁定他人账号；鉴权成功并得到 AccountID 后可以改用 AccountID；
- `auth_failure` 只统计实际鉴权失败，达到窗口次数后进入短暂冷却，不做永久账号锁定；
- `concurrency` 限制单个 LoginService 实例正在处理的登录请求数量，使用本地信号量，不属于时间窗口限流；
- 窗口超限和并发超限均立即返回 `TooManyRequests`，HTTP 状态使用 `429`，不在服务端排队等待；
- 限流键不得保存原始 `AccessToken`、密码或其他 SDK 凭证；
- 配置中任意 `enabled: false` 只关闭对应维度；顶层 `enabled: false` 关闭全部登录限流。

多 LoginService 实例部署时，时间窗口计数优先使用 Redis 原子执行以获得全局限制。建议用一个带 TTL 的 Sorted Set 保存窗口内请求时间，通过 Redis Function 原子完成“删除最老窗口外记录、统计各窗口、判断、写入本次请求”；同一个限流键的多个窗口共用一份记录，按最长窗口清理。这样不会出现固定窗口边界的瞬间双倍流量。

Redis 暂时不可用时降级为实例本地窗口，不能因为限流依赖故障而阻止开发环境登录。生产环境是否允许该降级必须通过安全配置明确指定。

## 10. Ready 条件

LoginService 只有在完成以下准备后才能开放 HTTP 登录接口：

- AccDBService RPC 已可用，并完成账号数据库及所需 Redis 能力的可用性检查；
- 登录所需账号集合已经可访问；
- `RealAreaInfo` 已查询完成；
- `ShowAreaInfo` 已查询完成；
- 显示区服到真实区服的关联已经校验；
- `RealAreaInfo.GateList` 中的 Gateway 地址已经合并到区服列表；
- 已生成完整的初始区服列表内存快照；
- Token 签发组件已经初始化；
- 当前启用的登录鉴权实现已经初始化。

任何必需步骤失败时，LoginService 不得以空区服列表或半初始化状态对外提供登录服务。

## 11. 实现前待确定

1. 三张基础表除已确定主键和唯一索引之外的最终字段、BSON 命名及基础数据。
2. `Account.AreaHis` 的具体写入者；写入时机已确定为玩家成功进入 GameService 之后。
3. SDK 鉴权扩展接口的最终 Go 类型；第一版只保留 TODO，不实现第三方平台验证。
4. 登录限流默认参数是否需要根据压测结果调整；算法、维度和开关方式已经确定。

## 12. 实现位置与运行

当前 MVP 实现目录：

```text
cmd/main.go                      OriginGame 统一程序入口
config/application.yaml          本地开发配置
config/gatewayserver-token-public-keys.yaml.example
                                与开发私钥配对的 Gateway 公钥示例
protocol/common/                 客户端共享 Protobuf 错误码
internal/security/               Ed25519 JWT 签发
service/loginservice/            LoginService、数据访问、快照和限流
bin/run.bat、bin/run.sh           Windows 与 Linux 统一启动脚本
```

`config` 目录下的配置文件统一平铺，不再按 Server 创建子目录。Origin 会递归加载 `--config` 目录下所有 `.json`、`.yml` 和 `.yaml` 文件，因此仅供参考、不应参与当前进程加载的示例使用 `.yaml.example` 后缀。

先通过 `deploy/compose/compose.yaml` 启动 MongoDB 和 Redis，再执行：

```powershell
./bin/run.bat login-1
```

或在 Linux 下执行：

```bash
./bin/run.sh login-1
```

当前 `authentication.development_passthrough: true` 明确表示不执行真实平台 SDK 鉴权，仅供学习和本地开发。

当前工程使用 Go 1.27，并通过 `go.mod replace` 直接对接同级正式开发目录 `../origin_v3`。
截至 2026-08-13，Go 1.27 稳定版尚未发布，当前通过 `go 1.27rc2` 固定使用官方 RC2；稳定版
发布后应把 `go` 指令升级到正式的 Go 1.27 补丁版本，并重新执行完整门禁。
