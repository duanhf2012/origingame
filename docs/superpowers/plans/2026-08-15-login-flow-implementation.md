# 登录链路与玩家加载实现计划

> **执行要求：** 按 `superpowers:executing-plans` 分批实施；每个行为变更先写失败测试，再写最小实现，阶段完成后执行对应验证。

**目标：** 在 OriginGame v3 中实现已确认的 DBService、LoginService、GatewayService、GameService、玩家路由及玩家数据加载/存档闭环。

**架构边界：** DBService 只提供通用 MongoDB/Redis 执行契约，并通过模板实例化为 `AccDBService` 与 `RoleDBService`。LoginService 负责账号鉴权与登录票据；GatewayService 负责客户端连接、区服映射、GS 分配和转发；GameService 负责玩家登录互斥、数据加载、消息执行、脏数据存档和断线驻留。跨服务契约统一放在 `protocol/rpc`，客户端协议统一放在 `protocol/common`。

**技术栈：** Go 1.27、Origin v3、Protobuf、MongoDB 8、Redis 7、Docker Compose。

---

## 任务 1：协议源文件与生成链路

**文件：**

- 新增：`protocol/common/messageid.proto`
- 新增：`protocol/common/playerlogin.proto`
- 修改：`protocol/common/errorcode.proto`
- 新增：`protocol/rpc/dbservice.go`
- 新增：`protocol/rpc/gameservice.go`
- 新增：`protocol/rpc/gatewayservice.go`
- 修改：`scripts/generate.bat`、`scripts/generate.sh`
- 修改：`scripts/check-generated.bat`、`scripts/check-generated.sh`
- 测试：`protocol/rpc/codec_test.go`

- [x] 用编译测试固定消息 ID、错误码、登录请求/响应和 RPC 顶层类型。
- [x] 定义 DBService 的 MongoDB/Redis 通用请求结果，以及 Gateway/Game 登录和发消息 RPC。
- [x] 生成并提交客户端 `*.pb.go` 与 RPC `*.rpc.gen.go`，首期不创建 RPC 专用 Protobuf，确保手写协议不依赖 Service 实现包。
- [x] 执行 `go test ./protocol/...` 和生成物一致性检查。

## 任务 2：DBService 有界同 Key 执行器

**文件：**

- 新增：`service/dbservice/keyexecutor.go`
- 新增：`service/dbservice/keyexecutor_test.go`
- 新增：`service/dbservice/dbservice.go`

- [x] 测试同 Key FIFO、不同 Key 并发、全局 I/O 槽限制、inflight 上限和单 Key 上限。
- [x] 实现“先预留 inflight、后进入 Origin Await”的准入方式；空 key 直接竞争全局 I/O 槽。
- [x] 用 `defer` 保障 panic、取消和普通错误都会释放执行槽并唤醒下一请求。
- [x] 执行 `go test -race ./service/dbservice`。

## 任务 3：DBService MongoDB 执行器

**文件：**

- 新增：`service/dbservice/mongodbmodule/mongodbmodule.go`
- 新增：`service/dbservice/mongodbmodule/executor.go`
- 新增：`service/dbservice/mongodbmodule/executor_test.go`
- 修改：`service/dbservice/dbservice.go`

- [x] 用表驱动测试固定标准操作校验、顺序执行、事务约束、结果映射和失败索引。
- [x] 覆盖 Insert、Find、Count、Aggregate、Update、Replace、FindOneAndXxx、Delete 及受控 `RawCommand`。
- [x] 实现请求/结果 BSON 编解码边界、10 万文档结果上限和慢操作日志。
- [x] 让 DBService 生命周期明确拥有 MongoDB Client，并在启动失败时逆序清理。

## 任务 4：DBService Redis 执行器与脚本

**文件：**

- 新增：`service/dbservice/redismodule/redismodule.go`
- 新增：`service/dbservice/redismodule/executor.go`
- 新增：`service/dbservice/redismodule/scripts.go`
- 新增：`service/dbservice/redismodule/executor_test.go`
- 修改：`service/dbservice/dbservice.go`

- [x] 用测试固定单命令、Pipeline、事务、注册脚本和通用结果树转换。
- [x] 实现已确认的普通命令白名单，并拒绝阻塞、管理和任意脚本命令。
- [x] 注册登录限流与玩家路由脚本；脚本只能通过固定 ID 调用。
- [x] MongoDB 与 Redis 请求共用同一 KeyExecutor，保证相同 dispatch key 跨存储有序。

## 任务 5：DBService 模板化配置与部署

**文件：**

- 新增：`config/local-node.yaml`
- 删除：`config/login-node.yaml`
- 新增：`config/dbservice.yaml`
- 修改：`cmd/main.go`
- 修改：`deploy/compose/compose.yaml`
- 修改：`deploy/mongodb/init-mongo.js`

- [x] 配置公共 `AccDBService`、区服 `AccDBService` 和区服 `RoleDBService`，实际 ServiceName 使用模板别名。
- [x] 仅暴露 `max_io_concurrency`、`max_inflight_requests` 两个容量配置并附中文注释；单 Key 32 上限使用代码常量。
- [x] MongoDB 本地环境改为副本集，账号库、角色库和 Redis DB 按数据域隔离。
- [x] 验证 Node/Service 配置可以被 Origin 加载，并完成 DBService Ready/Stop 生命周期测试。

## 任务 6：LoginService 切换到 AccDBService

**文件：**

- 修改：`service/loginservice/loginservice.go`
- 修改：`service/loginservice/loginconfig.go`
- 修改：`service/loginservice/account/*`
- 修改：`service/loginservice/area/*`
- 修改：`service/loginservice/ratelimit/*`
- 修改：`service/loginservice/httpapi/*`
- 修改：`internal/security/token.go`
- 修改：`config/loginservice.yaml`

- [x] 更新请求、Token claims、账号创建和区服快照测试。
- [x] 删除 LoginService 直连 MongoDB/Redis 的生命周期模块，统一通过 `AccDBService` RPC。
- [x] 请求只保留 `PlatType`、`PlatID`、`AccessToken`；JWT 只保留 `iss/aud/sub/iat/exp`。
- [x] 实现 IP、身份两个 10 秒固定窗口和本地并发上限，不做 Redis 故障降级。
- [x] 区服目录每分钟刷新；失败沿用上次成功快照，首次无有效数据则 Service 启动失败。

## 任务 7：公共玩家路由与 GameService 注册

**文件：**

- 新增：`service/gatewayservice/playerroute/playerroute.go`
- 新增：`service/gatewayservice/playerroute/playerroute_test.go`
- 新增：`service/gameservice/registration/registration.go`
- 新增：`service/gameservice/registration/registration_test.go`

- [x] 测试并实现 Redis 脚本的空闲 GS 分配、已有归属复用、负载排序和 5 秒释放 TTL。
- [x] PlayerKey 固定为 `AccountID + ShowAreaID`，RealAreaID 只作为当前运行归属筛选条件。
- [x] GameService 注册信息带真实 NodeID、ServiceName、RealAreaID、连接数、负载和 15 秒租约。
- [x] Redis 不可用时阻止新登录分配，但不主动终止已在线玩家。

## 任务 8：GameService Player、Proxy 与持久化生命周期

**文件：**

- 新增：`internal/mongodb/userinfo.go`
- 新增：`service/gameservice/player/player.go`
- 新增：`service/gameservice/player/proxy.go`
- 新增：`service/gameservice/player/lifecycle.go`
- 新增：`service/gameservice/player/persistence.go`
- 新增：`service/gameservice/player/*_test.go`

- [x] 测试 Proxy 按注册顺序初始化/加载/Ready，按倒序释放，阶段失败时只回滚已完成 Proxy。
- [x] Player 持有持久化 `CUserInfo`、非持久化 `DataInfo`、当前 GatewayNodeID 和 ConnectionID。
- [x] 自动 BSON 加载与保存；业务只标脏，首个 5 分钟存档 Tick 按 PlayerKey 分散，之后每 5 分钟执行。
- [x] 断线驻留 15 分钟；期间继续尝试脏数据定时存档，下线释放前必做一次存档，失败只记错误日志。

## 任务 9：GameService 消息注册与登录互斥

**文件：**

- 新增：`service/gameservice/messagehandler/register.go`
- 新增：`service/gameservice/messagehandler/router.go`
- 新增：`service/gameservice/messagehandler/router_test.go`
- 新增：`service/gameservice/gameservice.go`
- 新增：`service/gameservice/login.go`
- 新增：`service/gameservice/login_test.go`

- [x] 用 Go 1.27 泛型注册具体 Request 消息，不使用反射解析热路径。
- [x] Session 只持有 `*Player`；`Player.Reply`、`Player.ReplyError` 统一经 Gateway RPC 发包。
- [x] 登录 RPC 显式携带 GatewayNodeID 和 ConnectionID；不同 ConnectionID 后登录者替换旧连接并通知 Gateway 踢线。
- [x] 同一 Player 默认依赖 Origin Service 调度顺序，仅在 Await 跨出执行权的有限流程中再次核对 ConnectionID。

## 任务 10：GatewayService 连接、登录与转发

**文件：**

- 新增：`service/gatewayservice/gatewayservice.go`
- 新增：`service/gatewayservice/session.go`
- 新增：`service/gatewayservice/login.go`
- 新增：`service/gatewayservice/forward.go`
- 新增：`service/gatewayservice/*_test.go`
- 新增：`config/gatewayservice.yaml`

- [x] TCP/KCP/WebSocket 统一 string ConnectionID，并维护 ConnectionID 到 Session 的唯一索引。
- [x] 验证 LoginService JWT、按 ShowAreaID 映射 RealAreaID、按 Label 筛选可用 GS，并使用 Redis 路由分配。
- [x] 客户端包体上限 4KiB；每连接 10 秒窗口最多 200 条，超限主动关闭。
- [x] 客户端心跳由 Gateway 转发；GameService 超过 15 秒未收到则调用 Gateway RPC 关闭连接。
- [x] Gateway 的统一 RPC 支持普通响应、错误响应和主动推送；`MessageID == 0` 表示成功且没有业务 Body。

## 任务 11：完整链路配置与进程级测试

**文件：**

- 新增：`config/gameservice.yaml`
- 新增：`tests/integration/login_flow_test.go`
- 新增：`tests/e2e/login_flow_test.go`
- 修改：`cmd/main.go`

- [x] 在唯一 `local-node.yaml` 中配置 LoginServer、GatewayServer、GameServer、公共/区服 DBServer 的 NodeID、ServiceName 和 Labels，避免多个可加载文件重复定义 `nodes`。
- [x] E2E 覆盖首次创建、已有路由复用、顶号和驻留重连；单元测试固定 Redis 故障不降级、GS 15秒租约、5秒 Leaving 隔离和15分钟断线驻留。
- [x] 所有 Service 关键依赖就绪后才 Ready；部分初始化失败按逆序释放已创建资源。
- [x] 在 Windows 完成单元与竞态测试，在 Ubuntu/Docker 环境完成 MongoDB 副本集、Redis 和五 Node 登录链路联调。

## 任务 12：最终验证与设计一致性复核

**文件：**

- 修改：仅修复验证中发现的代码、配置和设计偏差。

- [x] `gofmt`、生成物检查、`go vet ./...`、`go test ./...`、`go test -race ./...`、`go build ./cmd`。
- [x] 对 DBService KeyExecutor、消息路由和登录 Token 验签热路径保留 Benchmark，确认无反射注册和无界资源。
- [x] 核对设计文档、配置注释、RPC/客户端协议与实际实现；不保留空包、无使用抽象和多余配置。
- [x] 停止本次 Ubuntu 联调启动的全部业务进程，保留 Docker 基础设施供后续开发复用，并记录完成范围与验证证据。
