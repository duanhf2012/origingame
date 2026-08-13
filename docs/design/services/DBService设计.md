# DBService 设计

> 状态：职责、部署模型、有界 KeyExecutor、有序执行模型、基础配置收敛和 RPC 总体形态已确认；MongoDB/Redis 操作契约设计稿待确认
> 更新日期：2026-08-13
> 上位文档：[OriginGame v3 总体架构设计](../总体架构设计.md)

## 1. 服务定位

| 层级 | 名称 |
| --- | --- |
| 生产程序 | `OriginGame`（统一程序，按 Node 配置启动） |
| Origin Service 模板 | `DBService` |
| Go 包 | `dbservice` |

`DBService` 是业务无关的集中式数据基础设施代理，统一持有 MongoDB 和 Redis Client、连接池及其资源生命周期。LoginService、GatewayService、GameService 和其他业务 Service 均通过 Origin RPC 使用数据能力，不直接装配 `mongodbmodule` 或 `redismodule`。

集中访问的主要目标是让 MongoDB、Redis 的总连接数随少量 DBService 副本数增长，而不是随全部 GameService 和业务 Service 实例数增长，并集中实施请求容量、并发、超时、过载和基础设施健康管理。

这里的“无状态”是无业务状态。DBService 仍然持有 MongoDB/Redis Client、连接池、执行队列、运行指标和生命周期状态，但不持有玩家缓存、账号缓存、业务版本、会话归属或迁移状态。

## 2. 业务无关边界

DBService 只执行调用方明确提交的数据操作并返回真实执行结果，不理解这些数据属于账号、角色、背包、任务、在线路由还是登录限流。

DBService 不负责：

- 定义或解释玩家、账号等业务对象；
- 根据业务内容推断数据库、集合、Redis Key 或目标 DBService；
- 管理 BSON 版本、数据版本回退或兼容策略；
- 管理玩家迁移、会话栅栏、在线所有权或业务幂等；
- 为业务增加 `LoadPlayer`、`SavePlayer` 等领域 RPC；
- 保存业务缓存，或在 MongoDB 与 Redis 之间自动补偿；
- 把 MongoDB 与 Redis 操作包装成跨资源原子事务。

调用方所属的 Repository 或基础能力包负责构造集合名、BSON 条件、更新内容、Redis Key 和原子操作参数，并按业务语义解释结果。新增业务字段或业务操作时应修改业务包及其 Repository，不应要求同步修改或重启 DBService。

具体 MongoDB CRUD 是以下基础操作的统称：

- Create：新增文档；
- Read：查询文档；
- Update：更新文档；
- Delete：删除文档。

DBService 对外只声明 `ExecuteMongo` 和 `ExecuteRedis` 两个业务无关 RPC 方法。高频 MongoDB 能力使用结构化操作，特殊能力使用受登记约束的原始 BSON Command；Redis 以受限原始命令为基础单元，并为 Pipeline、MULTI/EXEC、Function 和 Script 提供明确执行模式。所有 Redis 访问，包括在线路由、登录限流和需要的分布式协调操作，都必须经过 DBService。

## 3. 模板实例与数据库归属

仓库只实现一个 `DBService` 模板，通过 Origin v3 的 Service 模板机制实例化为不同实际服务，例如：

```text
DBService 模板
├── AccDBService
└── RoleDBService
```

每个实际 DBService 配置固定的 MongoDB 数据库名，并同时装配一个 MongoDB Module 和一个 Redis Module。RPC 请求不得携带或覆盖数据库名，只能在该实际服务已经绑定的数据库内指定后续确认允许的集合与操作：

```text
AccDBService
├── MongoDB Module -> 固定账号数据库
├── Redis Module   -> 配置指定的 Redis
└── KeyExecutor

RoleDBService
├── MongoDB Module -> 固定角色数据库
├── Redis Module   -> 配置指定的 Redis
└── KeyExecutor
```

不同实际 DBService 可以配置不同 Redis 地址。当前 OriginGame 只有一套 Redis，因此 AccDBService、RoleDBService 等实例可以指向同一 Redis 部署；仍由各实际 DBService 分别持有自己的 Redis Client 和受控连接池，不额外建立专用 `RedisDBService`。

同一个实际服务可以在不同 Node 上运行多个同名副本。业务层明确选择 AccDBService、RoleDBService 等数据域，再使用 Origin RPC 的实例选择策略挑选其中一个副本。DBService 不根据集合名或 Key 自动切换数据域。

## 4. 调用方路由与 Key 契约

请求包含可选的 `dispatch_key`。非空 Key 同时承担两层稳定路由职责：

```text
业务 dispatch_key
    -> RPC Client.Route(dispatch_key)
       -> 在候选集合不变时稳定选择一个 DBService 副本
          -> 请求携带同一个 dispatch_key
             -> DBService 实例内 KeyExecutor 保证同 Key 顺序
```

调用方必须对 `Route` 和请求字段使用完全相同的 Key。只做其中一层都不能形成完整的顺序边界：只使用 `Route` 时，同一实例内的 I/O 仍可能并发；只携带 `dispatch_key` 时，请求可能先被投递到不同 DBService 副本。

空 Key 用于相互独立、不需要有序处理的请求：

- 调用方不依赖稳定的按 Key 副本选择，可以使用适合无序请求的 RPC 选择策略；
- DBService 将请求放入公共无序可运行队列；
- 任一空闲 Worker 都可以处理，不承诺不同空 Key 请求之间的先后顺序。

是否需要顺序由业务层决定。与同一业务对象有关联的 MongoDB 和 Redis 请求必须选择同一个实际 DBService，并使用同一个非空 Key；即使 AccDBService 与 RoleDBService 指向同一套 Redis，同 Key 请求跨两个实际服务也没有共同顺序保证。

## 5. Origin ServiceScheduler 的作用和限制

Origin v3 的每个 ServiceScheduler 只有一个业务执行槽。RPC 请求到达 DBService 后先成为一个受 `scheduler.max_tasks` 限制的根任务，并在 Service Ready FIFO 中等待执行。

如果 Handler 直接同步执行数据库 I/O，整个 DBService 同时只能处理一个请求；如果每个 Handler 直接调用 `Await` 执行 I/O，则当前任务进入 Waiting 并释放唯一执行槽，不同请求可以并发，但相同 Key 的后续请求也可能先完成，不能保证顺序。

Origin 的 `Await` 还会为每个 Waiting 根任务保留原任务 goroutine，并由 `scheduler.max_await_tasks` 限制数量。框架默认容量面向普通 Service，DBService 必须按单请求最大内存、RPC Payload、数据库响应和目标延迟单独核算，不能未经验证直接沿用较大的默认值。

因此两层调度各自负责不同问题：

| 层级 | 职责 |
| --- | --- |
| Origin ServiceScheduler | RPC 根任务准入、单 Service 生命周期、Await、Deadline、停止排空和根任务总上限 |
| DBService KeyExecutor | 同 Key FIFO、MongoDB/Redis 交叉有序、固定 I/O 并发、内部积压和公平调度 |
| MongoDB/Redis Module | Client、连接池、拓扑恢复和实际 I/O |

不修改 Origin 的通用 ServiceScheduler 来加入 DBService 专用的 Key 语义，也不使用 `Await + 每 Key 锁`，避免锁等待顺序不明确和大量等待锁的 goroutine 占用调度额度。

## 6. 有界 KeyExecutor

DBService 采用“动态 Key FIFO + 统一可运行队列 + 固定共享 Worker 池”方案。Key 只拥有轻量队列状态，不拥有 goroutine：

```text
Key Map                              Runnable FIFO
Key A -> A1 -> A2 -> A3 ───────┐
Key B -> B1 ────────────────────┼──> [Key A, X1, Key B, X2, Key C]
Key C -> C1 -> C2 ──────────────┘                 |
空 Key请求 X1、X2 直接作为独立项 ──────────────────┘
                                                   |
                                           固定 N 个 I/O Worker
                                                   |
                                      MongoDB Module / Redis Module
```

核心状态可以实现为：

```go
type KeyExecutor struct {
    mu        sync.Mutex
    keyQueues map[string]*keyQueue
    runnable  runnableQueue
    cond      *sync.Cond

    pendingRequests int
    runningWorkers  int
    accepting       bool
    workers         sync.WaitGroup
}

type keyQueue struct {
    requests      requestQueue
    state         keyState // ready、running
    runnableEntry *runnableEntry
}
```

以上状态由一把短临界区锁保护。锁内只执行准入检查、链表增删、状态切换和计数，不执行 BSON/Redis 编解码、数据库 I/O、日志回调或 Ticket 等待。每个 Ticket 保存自己在 `requestQueue` 中的节点引用；Ready Key 保存 `runnableEntry`，空 Key Ticket 直接保存自己的 Runnable 节点引用。因此排队取消、Key 重新排队和队列删除都能 O(1) 完成，不通过遍历长队列清理超时请求。

KeyExecutor 在启动时一次性创建恰好 `max_concurrent_requests` 个 Worker goroutine。条件变量 `cond` 与上述状态使用同一把锁；Worker 在 `runnable` 为空且仍可继续运行时通过 `for` 循环等待，Submit 每新增一个 Runnable 项至少 `Signal` 一个 Worker，停止时 `Broadcast` 全部 Worker。条件变量只负责唤醒，锁内 `runnable` 才是唯一事实来源，因此虚假唤醒和暂时没有等待者都不会丢任务。Submit 只负责入队和唤醒，不启动 goroutine；全部 Worker 在已接受任务排空或停止 Context 生效后退出，由 DBService 通过 `WaitGroup` 统一等待。

非空 Key 请求准入和执行规则：

- 非空 Key 拥有一个逻辑 FIFO，同一个 Key 同时最多只有一个请求执行；
- MongoDB 和 Redis 请求共用同一套 Key FIFO，按 KeyExecutor 成功准入顺序交叉串行；
- 新 Key 创建 `keyQueue` 并追加请求；已有 Key 只向该 Key FIFO 追加请求；
- 一个未运行的非空 Key 在统一 Runnable FIFO 中最多出现一次，不把该 Key 的每个请求都重复放入 Runnable FIFO；
- Worker 从 Runnable FIFO 取得 Key 后，只取该 Key 队首的一个请求并把 Key 标为 Running；实际 MongoDB/Redis I/O 在锁外由该 Worker 同步执行；
- 一个请求实际 I/O 返回后，Worker 才重新加锁并释放该 Key；如果仍有后续请求，把 Key 重新放到 Runnable FIFO 尾部，从而避免热点 Key 连续独占 Worker；
- Key 队列清空后立即从 Key Map 删除，不形成永久增长的 Key 状态；
- 空 Key 请求不创建 Key 状态，作为独立 Runnable 项进入同一个 FIFO，与非空 Key 共同参与 Worker 公平调度；
- 排队请求取消时从对应队列 O(1) 删除并完成 Ticket；正在运行的请求只取消 Driver Context，必须等实际 I/O 返回后才能释放 Key；
- 全部 Key 共享固定数量的 Worker，Worker 直接执行实际 MongoDB 或 Redis I/O；
- 不为每个 Key 创建 goroutine，也不为每个请求创建额外 I/O goroutine；
- Worker 数量运行期间冻结，修改后必须重启并按生命周期排空旧实例。

例如同时准入 10000 个不同 Key、`max_concurrent_requests` 为 64 时，KeyExecutor 最多只有 64 个数据库 I/O 同时执行，其余 Key 只是 Key Map 和 Runnable FIFO 中的有界状态，不会创建 10000 个 I/O goroutine。但每个等待 RPC 都会保留一个 Origin Await 任务 goroutine，所以总体等待数量必须继续由 `scheduler.max_await_tasks` 限制。

本方案不采用“对Key取模后进入固定私有Worker队列”的分片调度。固定分片虽然简单，但慢Key会阻塞碰巧落入同一分片的无关Key，也可能出现一个队列已满而其他Worker空闲的问题。动态Key FIFO保留固定资源上限，同时减少无关Key的队头阻塞；实现复杂度集中封装在DBService的专用执行能力中。

## 7. 容量、配置和过载

DBService 不再暴露 `key_executor` 和 `rpc_limits` 两组配置。单个实际 DBService 的基础配置只聚合资源连接和一个业务并发字段：

```yaml
dbservice:
  max_concurrent_requests: 64       # MongoDB和Redis共享的最大实际I/O并发数；省略时默认64
  mongodb:
    uri: mongodb://127.0.0.1:27017  # MongoDB连接地址，可携带Driver连接池和超时参数
    database: account               # 当前实际DBService固定绑定的数据库名
  redis:
    addresses: [127.0.0.1:6379]     # 当前实际DBService使用的Redis数据节点地址
```

示例只列当前职责必需字段；Redis拓扑、认证和连接池等真实需要的字段继续使用Redis Module配置，MongoDB连接池继续使用URI参数，不在DBService层复制一套同义字段。`max_concurrent_requests`必须大于零且不能超过Origin的`scheduler.max_await_tasks`，启动后冻结；调整它需要重启并排空旧实例。

配置所有权统一如下：

| 内容 | 所有者 | DBService是否重复配置 |
| --- | --- | --- |
| MongoDB地址、固定数据库、连接池和Driver超时 | MongoDB Module | 否，DBService只聚合Module原配置 |
| Redis拓扑、地址、认证、连接池和网络超时 | Redis Module | 否，DBService只聚合Module原配置 |
| MongoDB/Redis共享I/O并发 | DBService | 是，唯一业务并发字段`max_concurrent_requests` |
| 根任务、Await总量和默认Await超时 | Origin ServiceScheduler | 否，首期直接使用Origin默认值 |
| RPC请求和响应Payload | Origin RPC | 否，当前默认4MiB |
| 单Key积压、操作/命令数量和结果数量 | DBService代码常量 | 否 |
| RawCommand、Function和Script登记 | DBService可选扩展 | 基础配置不出现，真实使用时才增加登记项 |

全局未完成请求上限直接复用 Origin `scheduler.max_await_tasks`。DBService 的 `Submit` 是私有能力，只允许在已经成功进入 `Await` 的函数内调用，因此每个排队或运行 Ticket 必然对应一个 Await 名额：

```text
KeyExecutor Pending <= scheduler.max_await_tasks
活动非空 Key数量 <= KeyExecutor Pending
Runnable FIFO项数量 <= KeyExecutor Pending
```

因此不增加`max_pending_requests`、`max_active_keys`、`max_unordered_pending_requests`和`runnable_queue_size`：它们都能由已有Await上限推导。空Key和非空Key共用全局容量，当前没有真实需求证明需要再划分隔离配额。单个非空Key的运行加排队请求固定最多128个，作为代码安全常量，不开放配置；达到上限立即返回稳定的热点Key过载错误。

`scheduler.max_tasks`、`scheduler.max_await_tasks`和`scheduler.default_await_timeout`属于DBService所在Node的Origin框架配置，不属于DBService配置，也不在DBService基础配置中重复展示。首期使用Origin默认值；以后只有全Node任务容量确需调优时，才在对应`<node>-node.yaml`调整。达到Origin Await上限时由框架返回队列满。所有过载都不等待容量释放、不临时增加Worker，也不创建补偿goroutine。

## 8. 请求执行、取消与返回

一次 RPC 的内部流程为：

```text
RPC Runtime
    -> Origin ServiceScheduler 准入根任务
    -> DBService 校验并完整拥有请求数据
    -> Handler 成功进入 Origin Await并释放Service执行槽
        -> Await函数内调用KeyExecutor.Submit
        -> Await函数等待Ticket
        -> 固定Worker执行MongoDB/Redis I/O
        -> Ticket完成，Await函数返回
    -> 原 RPC Task 进入 Origin 恢复 FIFO
    -> 返回响应
```

必须先成功进入 Await，再调用 KeyExecutor.Submit，禁止先 Submit 后尝试 Await。Origin 在调用 Await 函数前已经原子检查并占用 `max_await_tasks` 名额；因此不会出现 Ticket 已占用 KeyExecutor 容量，但 Handler 随后因 Await 满而无法等待、必须撤销 Ticket 的窗口。

Await 函数等待 Ticket 时必须区分排队和运行状态：排队请求收到取消或超时时，由当前 Await goroutine 同步调用 KeyExecutor 取消，在锁内 O(1) 删除 Ticket、释放 Pending 并返回，不创建取消回调 goroutine；如果请求已经被 Worker 取得，则取消传给 Driver 的 Context，但 Await 函数仍必须等待 Worker 报告实际 I/O 返回后才能结束。运行请求不能因调用方已经超时就提前释放 Await 名额、Ticket 或 Key，否则既会重新形成同 Key 并发，也会破坏 `KeyExecutor Pending <= scheduler.max_await_tasks` 的容量关系。

调用方可能已经因 Deadline 收到错误，而数据库操作的最终状态仍不确定。DBService 只返回实际可观察结果，不根据超时推断写入一定未发生；业务层按自身需要使用幂等 Key、数据库原子条件或查询确认。

## 9. 顺序保证与非保证

DBService 保证：

> 在同一个 DBService 实例内，具有相同非空 `dispatch_key` 且成功准入 KeyExecutor 的 MongoDB 和 Redis 请求，按准入顺序串行执行；前一个实际 I/O 返回后才开始下一个请求。

DBService 不保证：

- 不同 Key 之间的顺序；
- 空 Key 请求之间的顺序；
- MongoDB 和 Redis 之间的跨资源原子性；
- 调用方多个并发 goroutine 的代码发起顺序；
- 请求超时等于数据库端一定没有提交；
- DBService 副本扩缩容、故障或发现候选变化期间的跨实例全局顺序；
- 业务版本、玩家迁移、会话所有权、补偿和最终一致性。

Origin `Route(dispatch_key)` 只在候选副本集合不变时稳定。DBService 副本变更必须采用受控停止和排空；真正要求故障切换期间仍保持正确性的业务，必须使用 MongoDB/Redis 原子操作、幂等业务请求或另行设计的分片所有权机制，不能依赖实例内 FIFO。

## 10. 生命周期和 Ready

KeyExecutor 是 DBService 实例拥有的长期资源。所有 Worker 都必须由该实例创建、停止并等待；禁止无法停止和等待的 goroutine。

DBService 进入 Ready 前必须完成：

- 实际 Service 配置和固定 MongoDB 数据库名校验；
- MongoDB Module 连接及必要可用性检查；
- Redis Module 连接及必要可用性检查；
- KeyExecutor 容量配置校验和固定 Worker 启动；
- 对外 RPC 契约注册完成；
- 必需数据库能力、可选MongoDB RawCommand登记以及Redis Function/Script登记和版本校验。

任一步失败必须按成功创建资源的逆序回滚，DBService 不以缺少 MongoDB、Redis 或 Worker 的半初始化状态对外服务。

停止时先由 Origin 关闭新 RPC 根任务准入。KeyExecutor 必须继续接收并处理停止边界前已经进入 Service Ready FIFO 的根任务，不能让这些已接受请求因内部执行器提前关闭而失败；这些任务及其 KeyExecutor 工作在停止 Context 预算内共同排空。全部已接受根任务完成并进入 `OnStop` 后，关闭 KeyExecutor 准入并等待固定 Worker 退出，再关闭 Redis、MongoDB Module。停止预算耗尽时由生命周期 Context 取消排队和运行请求，Worker 仍须等实际 Driver 调用返回后退出。`Stop` 必须幂等。

## 11. RPC 契约总体形态

DBService 使用一个与模板名称一致的 RPC 接口，不按实际服务 AccDBService、RoleDBService 重复声明契约：

```go
//origin:rpc
type DBService interface {
    ExecuteMongo(context.Context, MongoRequest) (MongoResult, error)
    ExecuteRedis(context.Context, RedisRequest) (RedisResult, error)
}
```

RPC 契约方法按数据动作命名，不增加 `Rpc` 或 `RPC` 前缀。`origingen` 已经生成 `Await`、`Call`、`Async`、`Notify` 和 `Broadcast` 调用外观，重复增加前缀会形成 `AwaitRpcExecuteMongo` 等冗余名称。该规则适用于 OriginGame 后续全部 RPC 契约。

调用方按是否关注远端执行结果选择调用外观：

| 调用方式 | 语义 |
| --- | --- |
| `AwaitExecuteMongo/Redis` | Service Task 协作式等待远端执行结果 |
| `CallExecuteMongo/Redis` | 普通 goroutine 同步等待远端执行结果 |
| `AsyncExecuteMongo/Redis` | 提交后在调用方 Service 后续串行任务中取得远端执行结果 |
| `NotifyExecuteMongo/Redis` | 只保证本地编码、路由和投递阶段，不等待也不返回远端执行结果 |

不增加只返回 `error` 的 `ExecuteAck` 契约。关注执行是否成功时使用请求—响应调用并取得完整结果；完全不关注远端结果时使用生成的 `NotifyXxx`。`NotifyXxx` 返回 `nil` 不代表 DBService 已经准入 KeyExecutor，更不代表数据库执行成功，目标端的后续失败只能记录日志和指标。

`BroadcastExecuteMongo/Redis` 必须禁止使用；它会把同一数据操作发送到多个 DBService 副本。关键数据写入、扣费、发奖、角色保存、账号创建和在线归属修改也禁止使用 Notify。Notify 只适用于允许丢失、可由其他事实源重新生成或者已有独立可靠投递机制的数据。

请求的 `DispatchKey` 必须与生成客户端的 `Route(request.DispatchKey)` 完全一致。结构化请求和结果类型统一放在 `protocol/rpc/dbservice.go`；当前数据以普通 Go 类型和 `[]byte` 表示 BSON/二进制值，不额外引入 Protobuf。

## 12. MongoDB 请求与执行模式

```go
type MongoRequest struct {
    DispatchKey string
    ExecuteMode MongoExecuteMode
    Operations  []MongoOperation
}
```

一个请求至少包含一个操作。单操作仍使用长度为一的 `Operations`，不再维护另一套单操作契约。执行模式固定为：

| 模式 | 执行与失败语义 |
| --- | --- |
| `Sequential` | 按数组顺序执行；首个失败后停止，之前已经成功的操作不回滚 |
| `Transaction` | 全部操作在一次 `mongodbmodule.WithTransaction` 中执行；任一执行错误或结果断言失败使事务回滚 |

Transaction 不支持时必须返回明确错误，禁止静默降级为 Sequential。MongoDB Driver 可能重新执行事务回调，因此请求必须能够安全重放；事务内部不得混入 Redis、RPC、HTTP、消息发送或其他外部副作用。一个请求作为 KeyExecutor 的一个 Ticket，在请求结束前同 Key 的下一个 MongoDB/Redis 请求不能插入其中。

后一个 Operation 首期不能引用前一个 Operation 的返回值，也不支持条件分支和循环。业务应优先使用条件 Filter、Update Pipeline、`FindOneAndUpdate`、事务和结果断言表达并发约束，避免把 DBService 演变成业务流程解释器。

## 13. MongoDB 标准操作

结合当前锁定的 `mongo-driver/v2 v2.8.0` Collection 数据 API、Origin MongoDB Module 示例和常见游戏项目场景，首期结构化操作覆盖：

| 分类 | OperationKind | 常见场景与结果 |
| --- | --- | --- |
| 插入 | `InsertOne` | 创建账号、角色、幂等流水；返回 InsertedID |
| 插入 | `InsertMany` | 批量初始化或导入同一集合；返回各 InsertedID |
| 查询 | `FindOne` | 按主键或唯一条件加载；返回零或一个文档 |
| 查询 | `FindMany` | 有界列表、分页、排行榜快照；返回文档集 |
| 查询 | `CountDocuments` | 按 Filter 精确计数 |
| 查询 | `EstimatedDocumentCount` | 无 Filter 的集合近似总量统计 |
| 查询 | `Distinct` | 有界的字段去重值 |
| 聚合 | `Aggregate` | 有界 Pipeline 聚合结果 |
| 修改 | `UpdateOne` | 条件扣减、乐观锁、普通 Upsert |
| 修改 | `UpdateMany` | 同条件批量修改 |
| 替换 | `ReplaceOne` | 整文档替换与可选 Upsert |
| 修改并返回 | `FindOneAndUpdate` | 原子修改或 Upsert，并返回修改前/后文档 |
| 替换并返回 | `FindOneAndReplace` | 原子整文档替换，并返回替换前/后文档 |
| 删除 | `DeleteOne` | 精确条件安全删除 |
| 删除 | `DeleteMany` | 有明确 Filter 的批量删除 |
| 删除并返回 | `FindOneAndDelete` | 原子删除并返回原文档 |
| 混合批量写 | `BulkWrite` | 同一集合混合 Insert、Update、Replace、Delete |
| 原始扩展 | `RawCommand` | 结构化操作尚未覆盖且已登记的特殊数据库命令 |

`UpdateByID` 不单独建立 OperationKind，它只是 `UpdateOne` 使用 `{_id: ...}` Filter 的便利外观。普通业务数据 RPC 不开放 Change Stream、IndexView、SearchIndexView、Drop Collection/Database、用户权限、复制集或分片管理；这些属于长生命周期或管理能力。

MongoDB v2 Driver 还提供跨 Namespace 的 Client BulkWrite。首期不把它列为常用标准操作：每个实际 DBService 已固定数据库，常见游戏写入集中在单集合 BulkWrite 或多 Operation Transaction，跨集合批量写可以先由这两种能力表达。出现明确的跨集合高吞吐、非事务写需求后，再基于真实压测决定是否增加 `MultiCollectionBulkWrite`；不能为了追齐 Driver API 预先扩大协议。

`MongoOperation` 使用判别联合结构，每种标准操作拥有专属参数类型：

```go
type MongoOperation struct {
    Kind MongoOperationKind

    InsertOne        *MongoInsertOne
    InsertMany       *MongoInsertMany
    FindOne          *MongoFindOne
    FindMany         *MongoFindMany
    CountDocuments   *MongoCountDocuments
    EstimatedDocumentCount *MongoEstimatedDocumentCount
    Distinct         *MongoDistinct
    Aggregate        *MongoAggregate
    UpdateOne        *MongoUpdateOne
    UpdateMany       *MongoUpdateMany
    ReplaceOne       *MongoReplaceOne
    FindOneAndUpdate *MongoFindOneAndUpdate
    FindOneAndReplace *MongoFindOneAndReplace
    FindOneAndDelete *MongoFindOneAndDelete
    DeleteOne        *MongoDeleteOne
    DeleteMany       *MongoDeleteMany
    BulkWrite        *MongoBulkWrite
    RawCommand       *MongoRawCommand
}
```

`Kind` 对应的参数字段必须恰好一个非空，其他参数字段必须为空。每种参数只开放当前契约确认的 Options，不直接传 Driver Options 类型，也不接受任意 Options BSON。

## 14. MongoDB BSON、分页和 Options

Filter、Document、Update、Projection、Sort、Hint、Pipeline 和 ArrayFilters 使用原始 BSON `[]byte` 或 `[][]byte` 表示。业务 Repository 使用 `bson.Marshal` 构造请求，DBService 只校验 BSON 合法性和操作边界，不理解字段含义。使用 BSON 而不是 JSON，可以完整保存 ObjectID、Decimal128、日期和嵌套类型。

Update 同时支持普通更新文档和 Update Pipeline：

```go
type MongoUpdate struct {
    Kind     MongoUpdateKind
    Document []byte
    Pipeline [][]byte
}
```

每种标准操作按需开放 `Projection`、`Sort`、`Skip`、`Limit`、`Hint`、`Collation`、`Upsert`、`ReturnDocument`、`ArrayFilters`、`Ordered` 和 `AllowDiskUse` 等明确字段。ReadConcern、WriteConcern、ReadPreference 和事务选项由实际 DBService 配置控制，首期不允许请求任意覆盖；所有写操作使用 acknowledged write concern。

写操作默认执行安全校验：`UpdateMany`、`DeleteMany` 必须具有非空 Filter；`UpdateOne`、`ReplaceOne`、`DeleteOne` 的空 Filter 默认拒绝，只有配置明确允许的受控集合和场景才可例外。禁止请求绕过文档校验和任意设置 WriteConcern。DBService 只能检查结构安全，唯一索引、业务条件、版本号和幂等键仍由业务 Repository 负责。

Transaction 模式具有独立的操作允许表。DBService 在开始事务前拒绝 MongoDB 服务端明确不支持或本配置未允许进入事务的操作；例如 `EstimatedDocumentCount`、带 `$out/$merge` 的 Aggregate 和未声明 `allow_in_transaction` 的 RawCommand 不能进入事务。禁止先执行一部分后才因可预检的兼容问题失败。

`FindMany` 必须带正数 `Limit`。需要分页或依赖顺序的查询由调用方提供稳定 Sort，并使用业务游标值构造 Keyset Filter；一次性读取有界快照且不依赖顺序时可以省略 Sort。DBService 不持有服务端 Cursor；`Skip` 只允许在有硬上限的低页码场景使用。大列表由调用方多次 RPC 分页，每页仍经过容量检查。DBService 不返回 Cursor ID，也不跨 RPC 保存 Cursor。

`Distinct`、`Aggregate` 和 Raw Cursor 必须声明结果数量上限；`Aggregate` 的 `$out`、`$merge` 等写入 Stage 需要通过操作配置单独允许，不能把写聚合伪装成只读查询。所有列表结果超过数量或字节上限时整体返回结果超限错误，不静默截断。

## 15. MongoDB BulkWrite 与原始 Command

`BulkWrite` 固定作用于一个集合，并包含有界 WriteModel 列表：

```go
type MongoBulkWrite struct {
    Collection string
    Ordered    bool
    Models     []MongoWriteModel
}
```

WriteModel 支持 `InsertOne`、`UpdateOne`、`UpdateMany`、`ReplaceOne`、`DeleteOne` 和 `DeleteMany`。`InsertMany` 与 `BulkWrite` 同时保留：前者表达单一批量插入并提供简单结果；后者用于减少混合写网络往返。BulkWrite 的各模型遵循 MongoDB 自身原子性，整个批次默认不是事务；需要整体原子时使用 Transaction 模式组织标准操作。

结构化操作尚未覆盖的特殊能力使用原始 BSON Command：

```go
type MongoRawCommand struct {
    Command      []byte
    ResultMode   MongoRawResultMode
    MaxDocuments int64
}
```

`Document` 模式使用 `Database.RunCommand` 返回单 BSON 响应；`Cursor` 模式使用 `Database.RunCommandCursor`，只在当前 RPC 内有界读取并关闭。RawCommand首期默认关闭；实际业务出现结构化操作无法表达的命令时，才按命令名登记并重启对应DBService，不要求修改RPC契约或重新编译DBService。

原始 Command 必须遵守：

- 只能访问实际 DBService 固定绑定的数据库；请求不得携带或覆盖 `$db`、Session 和事务号等 Driver 管理字段；
- 登记项明确命令名、结果模式和是否允许进入Transaction，并继续受MongoDB账号权限约束；
- Cursor 不跨 RPC 暴露或保存；请求和结果仍受统一容量与 Deadline 限制；
- 原始 BSON 不写入普通日志，只记录命令名、耗时、错误类别和结果大小；
- `dropDatabase`、用户权限、拓扑管理、`shutdown`、Change Stream和长期Cursor等能力硬拒绝，不能通过误登记绕过。

已有结构化`insert`、`update`、`delete`、`find`、`aggregate`等能力时，RawCommand默认不重复登记同名命令，防止绕过空Filter、结果上限和Options校验。只有真实业务需要结构化协议尚未开放的特殊Option时，才能增加具体登记项；登记项必须同时给出结果模式和事务许可，容量继续使用代码固定安全边界。

## 16. MongoDB 结果、断言与错误

```go
type MongoResult struct {
    Results []MongoOperationResult
    Failure *MongoExecutionFailure
}
```

结果与请求 Operation 下标一一对应。单项结果统一承载适用字段：`Documents`、`Values`、`Count`、`InsertedCount`、`MatchedCount`、`ModifiedCount`、`DeletedCount`、`UpsertedCount`、`InsertedIDs`、`UpsertedIDs`、`WriteFailures` 和 `CommandResponse`；不适用字段保持零值。ID 和 Distinct 值使用 BSON Type 加原始 Value 表示，不假设主键一定是 ObjectID。

`InsertMany` 和 `BulkWrite` 必须保留 MongoDB 的部分结果语义。`WriteFailures` 记录 Document/WriteModel 下标、稳定错误分类和 Server Code；Driver 能可靠返回的 InsertedID、UpsertedID 和影响数量同时保留。Ordered 操作在首项错误停止，Unordered 操作允许多个逐项错误。该 Operation 被视为失败并使 Sequential 停止后续 Operation，但调用方仍能看到其可信的部分结果；Transaction 失败时这些中间结果全部丢弃。

`FindOne` 和 `FindOneAndXxx` 没找到文档时返回空 `Documents` 且不形成 Failure。需要“必须找到”“必须修改一行”等语义时，由 Operation 携带通用 `MongoExpectation`，可约束 Document、Matched、Modified、Deleted 等数量范围。断言失败在 Sequential 中停止后续操作，在 Transaction 中使事务回滚；这只是数据库结果约束，不包含玩家或账号业务判断。

Sequential 失败时保留失败项之前已经成功的 Results。Transaction 失败时丢弃全部中间 Results，只返回 Failure，因为这些操作已经回滚或提交状态未知，不能作为成功结果使用。

RPC `error` 表示调用方没有取得一份可解释的执行报告，例如路由/编解码失败、DBService 未 Ready、KeyExecutor 过载、请求/BSON 非法或结果无法编码。请求进入数据库执行后发生的错误放入 `MongoResult.Failure`，记录失败 Operation 下标、稳定错误分类、MongoDB Server Code、Error Labels 和 `StateUnknown`，使 Sequential 能同时返回前面已成功操作的结果。未经处理的 Driver 错误文本不得回传。

`StateUnknown` 用于超时、断线或 UnknownTransactionCommitResult 等无法断定写入最终状态的情况。业务层不得把超时解释为一定未写入，应使用幂等键、唯一约束、条件更新或查询确认。

## 17. Redis 请求与执行模式

Redis 命令数量大、模块持续演进，若为每个 `GET`、`HSET`、`ZADD` 等命令增加独立 OperationKind，会复制 go-redis API 并造成契约频繁升级。Redis RPC 以受限原始命令为基础单元，用执行模式表达组合语义：

```go
type RedisRequest struct {
    DispatchKey string
    ExecuteMode RedisExecuteMode
    Commands    []RedisCommand
    Function    *RedisFunctionCall
    Script      *RedisScriptCall
}
```

| 模式 | 语义 |
| --- | --- |
| `Command` | 恰好执行一条命令 |
| `Pipeline` | 顺序收集多条命令并一次发送，减少往返，不保证原子性 |
| `Transaction` | 使用 MULTI/EXEC 提交多条命令；运行时命令错误不会像 MongoDB 事务一样自动回滚 |
| `Function` | 调用启动阶段已校验的 Redis Function |
| `Script` | 调用启动阶段已登记的短 Lua Script |

模式对应字段必须恰好一组有效。一个 RedisRequest 也是 KeyExecutor 的一个 Ticket；请求内的命令在结束前不会插入同 Key 的其他 MongoDB/Redis 请求。Redis 与 MongoDB 仍不存在跨资源原子事务。

## 18. Redis 命令、组合和 Cluster 约束

```go
type RedisCommand struct {
    Name string
    Args [][]byte
}
```

`Name`只保存命令名，`Args`按Redis命令语法依次保存Key、Value和选项，不重复保存命令名。参数使用二进制安全字节；整数、浮点数、时间戳和枚举选项由调用方按对应命令要求编码为十进制或约定文本。DBService通过`redismodule.Do`或Pipeline的`Do`执行命令。普通、有限时长的数据命令默认允许，不维护随业务增长的普通命令白名单；危险、管理、阻塞和长连接命令由代码固定拒绝，最终权限边界由Redis ACL控制。

首期支持的常见有界数据命令族如下；该表用于说明能力和硬限制，不是运行配置：

| 数据族 | 首期常用范围 | 主要限制 |
| --- | --- | --- |
| Key/TTL | `DEL`、`UNLINK`、`EXISTS`、`TYPE`、`TOUCH`、`COPY`、`EXPIRE/PEXPIRE`、`EXPIREAT/PEXPIREAT`、`EXPIRETIME/PEXPIRETIME`、`PERSIST`、`TTL/PTTL`、`RENAME`、`SCAN` | Scan 一次一页；全局 `SCAN` 仅支持 Standalone/Sentinel，Cluster 首期不提供跨节点全局扫描；多 Key 命令检查 Key Slot |
| String | `GET`、`SET` 及 NX/XX/TTL、`GETDEL`、`GETEX`、`MGET`、`MSET/MSETNX`、`INCR/DECR`、`INCRBYFLOAT`、`APPEND`、`STRLEN`、有界 `GETRANGE/SETRANGE` | 批量 Key/Value 数量、Range 和总字节有上限 |
| Hash | `HGET`、`HSET/HSETNX`、`HMGET`、`HDEL`、`HEXISTS`、`HLEN`、`HSTRLEN`、`HINCRBY/HINCRBYFLOAT`、`HRANDFIELD`、`HSCAN` | `HGETALL/HKEYS/HVALS` 默认关闭或要求小集合上限 |
| List | `LPUSH/RPUSH`、`LPUSHX/RPUSHX`、`LPOP/RPOP`、`LINDEX`、`LSET`、`LINSERT`、`LPOS`、有界 `LRANGE`、`LLEN`、`LTRIM`、`LREM`、`LMOVE` | 首期不允许阻塞 Pop/Move；位置、计数和 Range 有上限 |
| Set | `SADD`、`SREM`、`SISMEMBER/SMISMEMBER`、`SCARD`、`SPOP`、`SRANDMEMBER`、`SMOVE`、`SDIFF/SINTER/SUNION` 及对应 Store、`SSCAN` | 完整成员返回必须受结果上限约束；多 Key 运算检查 Slot |
| Sorted Set | `ZADD`、`ZINCRBY`、`ZREM`、`ZREMRANGEBY*`、`ZSCORE/ZMSCORE`、`ZRANK`、有界 Range/ByScore/ByLex、`ZCOUNT/ZLEXCOUNT/ZCARD`、Pop、Scan、`ZDIFF/ZINTER/ZUNION` 及 Store | Score 按命令文本编码；范围、集合数量和返回数量有上限；多 Key 运算检查 Slot |
| Bitmap/Bitfield | `SETBIT`、`GETBIT`、`BITCOUNT`、`BITPOS`、`BITOP`、受限 `BITFIELD` | Offset、Range、子命令数和结果大小有上限 |
| Geo | `GEOADD`、`GEODIST`、`GEOHASH`、`GEOPOS`、有界 `GEOSEARCH/GEOSEARCHSTORE` | 半径、返回数量和 Store 目标受限；多 Key 命令检查 Slot |
| HyperLogLog | `PFADD`、`PFCOUNT`、`PFMERGE` | 多 Key 命令检查 Slot |
| Stream | `XADD`、`XDEL`、`XLEN`、`XTRIM`、有界 `XRANGE/XREVRANGE` 等非阻塞操作 | 不把 Stream 包装成通用可靠队列；Consumer Group、阻塞消费和长期 Pending 管理另行设计 |

Redis Module未封装但Redis已提供的模块化数据命令，例如特定部署的RedisJSON，也可以通过普通Command执行，不改变RPC契约。是否能够执行取决于实际部署能力、Redis ACL、go-redis路由能力和代码硬拒绝规则。

Pipeline 返回每条命令对应的结果或命令错误，并保持数组下标一致。Pipeline 不是事务；中间命令失败不回滚其他命令。Transaction 使用 MULTI/EXEC，但 Redis 命令运行时错误仍可能和其他成功结果同时出现，因此结果必须逐项检查，不能只返回一个整体成功标志。

Standalone/Sentinel 中，同一 Pipeline 发往同一服务端连接时按提交顺序写入；Cluster Pipeline 可能拆分到多个节点，跨节点命令之间不承诺全局执行顺序。无论何种拓扑，整个 RedisRequest 实际返回前 KeyExecutor 都不会释放其 `dispatch_key`。

首期不在 DBService 内实现 WATCH 回调，因为后续写入通常依赖 WATCH 后读取的值，会要求请求内表达结果引用、条件分支和可重入回调。需要乐观并发时优先使用 Function/Script、版本字段加条件命令或业务幂等；出现无法表达的真实需求后再单独设计受限 WATCH Plan。

Cluster模式下，Pipeline可以按go-redis路由到多个节点但不具备整体原子性；Transaction、Function、Script和要求原子的多Key命令必须由调用方使用Hash Tag保证全部Key位于同一Slot。Function/Script的Keys与Args在协议上分离；普通命令优先复用go-redis及Redis服务端的Command Metadata完成Key提取和路由。Key布局动态且Driver无法可靠识别的命令不允许进入原子模式，不为此增加普通命令白名单配置。

以下命令硬拒绝：

- `AUTH`、`HELLO`、`SELECT`、`CLIENT`、`CONFIG`、`ACL`、`MODULE`、`SHUTDOWN`、`REPLICAOF`、`CLUSTER` 等连接、权限和拓扑管理命令；
- `FLUSHDB`、`FLUSHALL`、`MIGRATE`、`RESTORE` 等大范围破坏或迁移命令；
- `SUBSCRIBE`、`PSUBSCRIBE`、`MONITOR` 和其他长期占用连接或持续流式返回的命令；
- `KEYS`、无界阻塞命令和无法在当前 RPC Deadline 内确定结束的命令；
- `MULTI`、`EXEC`、`DISCARD`、`WATCH`、`EVAL/EVALSHA` 和 `FCALL`；这些只能由 Transaction、Function 或 Script 专用模式生成；
- Function/Script 的装载、删除和刷新命令；这些由 DBService 启动配置管理，不由业务请求执行。

扫描类命令允许在命令结果中返回下一页Cursor，但一次RPC只执行一页；调用方使用新请求继续扫描。`HGETALL`、`SMEMBERS`、`LRANGE 0 -1`等可能产生大结果的命令仍受统一结果节点数、RPC Payload和Deadline限制，不为单个命令增加独立配置。

## 19. Redis Function、Script 和结果

Function 和 Script 不允许请求携带任意代码。配置保存经过审核的 Function 名称或 Script ID、源码/摘要、只读属性、Key 数量、参数上限和返回上限；DBService 在 Ready 前加载或校验版本。请求只携带登记 ID、Keys 和 Args：

```go
type RedisFunctionCall struct {
    Name string
    Keys []string
    Args [][]byte
}

type RedisScriptCall struct {
    ID   string
    Keys []string
    Args [][]byte
}
```

Function/Script 适合登录限流、在线路由 Compare-And-Set、带 TTL 的幂等写等少量短命令原子组合。脚本会阻塞 Redis 执行线程，必须短小、有界、可观测；不得把通用业务流程或长循环放入脚本。

Redis Module 的进程内 Lease `Lock` 对象不跨 RPC 暴露，也不由 DBService 保存跨请求锁状态。需要分布式锁时，使用登记的 Function/Script 以随机 Owner Token 完成获取、刷新和条件释放，并让业务层持有 Token；获取、刷新和释放请求必须选择同一个实际 DBService 和一致的 `dispatch_key`。Redis Lease 仍不能替代数据库事务、唯一约束或业务幂等。

Redis Driver 已经解析 RESP 帧，DBService 返回的是其可观察到的语义值，不承诺保留 Blob String、Simple String、Set 等已经被 Driver 归一化掉的线格式差异。为避免 RPC 类型递归，同时仍能表达嵌套 Array/Map，返回值使用“节点表 + 索引引用”的非递归结构：

```go
type RedisValue struct {
    RootIndex uint32
    Nodes     []RedisValueNode
}

type RedisValueNode struct {
    Kind     RedisValueKind
    Bytes    []byte
    Integer  int64
    Double   float64
    Boolean  bool
    Children []uint32
}

type RedisResult struct {
    Results []RedisCommandResult
    Failure *RedisExecutionFailure
}
```

Value Kind 至少支持 Null、Bytes、Integer、BigNumber、Double、Boolean、Array 和 Map；BigNumber 使用十进制字节表示，避免溢出 Go `int64`。Array 的 `Children` 按元素顺序保存子节点索引；Map 按 Key、Value 交替保存索引，因此数量必须为偶数；标量节点的 `Children` 必须为空。`RootIndex` 和所有子索引必须指向同一 `Nodes`，节点图必须从 Root 可达、无环且每个节点只出现一次。该形态没有 Go 类型递归，可由 `origingen` 生成静态 Codec。

Null 表达 Redis Miss，不作为基础设施失败。Command/Pipeline/Transaction 的结果与命令下标一一对应；Function/Script 使用一个结果项。协议拒绝无法安全表达或超过节点数量、逻辑嵌套深度、单值字节数、子索引数量和总结果字节数的返回。Pub/Sub Push 消息不属于普通请求响应，首期硬拒绝对应命令。

`RedisCommandResult` 包含一个 Value 和可选的逐命令 Failure。Pipeline/Transaction 必须允许多个命令结果与逐项错误同时返回；`RedisResult.Failure` 只表达无法归属于某一命令的整体失败，例如连接中断、EXECABORT、超时或最终状态未知。与 MongoDB 相同，RPC `error` 只表达未取得可解释报告的框架/准入/校验问题。稳定分类至少表达 WrongType、NoScript、FunctionNotFound、TransactionAborted、CrossSlot、Timeout、Canceled、Network 和 Unknown；Null 是正常 Redis Miss，不是 Failure。原始 Redis 错误文本和命令参数不得直接回传或记录。

## 20. 固定安全边界

DBService不增加请求大小、结果大小、操作数量和事务时间的运行配置。请求和响应统一受Origin RPC当前默认4MiB业务Payload限制；事务和普通请求统一继承调用Context Deadline，调用方未提供更短值时使用Origin当前默认15s。DBService不延长Deadline，也不在同一请求中重新启动一套事务计时器。

以下边界作为代码常量集中定义，防止大量极小元素绕过Payload字节限制：

| 边界 | 首期固定值 | 说明 |
| --- | ---: | --- |
| 单Key运行加排队请求 | 128 | 防止热点Key占满全局Await容量 |
| 单请求MongoDB Operation | 128 | 包含Sequential和Transaction |
| `InsertMany`或`BulkWrite`模型 | 100000 | 理论条目上限，实际通常先触及4MiB请求上限 |
| 单请求Redis命令 | 1024 | 包含Pipeline和Transaction中的命令 |
| MongoDB结果文档 | 100000 | 理论行数上限，实际通常先触及4MiB响应上限 |
| Redis结果节点 | 100000 | 包含扁平结果节点表全部节点 |
| Redis结果逻辑深度 | 16 | 防止异常嵌套结果消耗解码栈和CPU |

MongoDB查询允许调用方提供更小的Limit；未提供或超过100000时，DBService仍以100000作为读取硬上限。为区分“恰好100000行”和“还有更多数据”，Cursor类结果最多探测第100001项：发现超限时返回明确的`ResultLimitExceeded`，不静默截断。任何结果在达到文档/节点数量上限前先超过Origin RPC Payload时，按Payload超限失败；10万行不是一次RPC必然能够承载的大小承诺。

固定常量只解决协议和内存安全，不替代MongoDB/Redis账号最小权限。日志和指标只记录实际服务名、操作/命令类别、排队/执行耗时、错误类别、操作数量和结果大小，不记录DispatchKey原值、BSON、Redis Key、Value、Token或连接凭证。

## 21. 监控、慢处理与测试

每个请求从Handler取得执行权开始，记录以下独立阶段：

```text
handler_total_duration
├── validation_duration
├── executor_queue_duration
├── database_io_duration
└── response_build_duration
```

阶段时间不重叠，`handler_total_duration`还包含阶段切换等少量开销。多Operation/Command请求除记录整个数据库阶段外，还记录每一项的执行耗时，并在诊断中保留最慢项的下标、类别和耗时。

至少记录以下指标：

- Origin ServiceScheduler的Accepted、Ready、Awaiting、高水位和拒绝数；
- KeyExecutor全局Pending、活动Key、Runnable项、运行Worker、Worker利用率和拒绝数；
- 单KeyPending高水位，但日志和指标不得暴露敏感原始Key；
- 请求总耗时、KeyExecutor排队、MongoDB/Redis I/O和响应构造耗时直方图，用于观察P50、P95、P99和最大值；
- 慢排队、慢MongoDB、慢Redis和慢总请求计数，按实际Service、资源类型、执行模式和稳定操作类别聚合；
- MongoDB/Redis操作成功、错误、取消、超时、状态未知、操作/命令数量、结果文档/节点数量和结果字节数；
- 停止排空耗时和被取消请求数。

首期不增加慢处理阈值配置。代码固定参考线为：KeyExecutor排队100ms、Redis I/O 100ms、MongoDB I/O 500ms、请求总耗时1s；达到任一参考线时增加对应慢处理计数，并输出限频结构化警告。警告只包含实际Service、资源类型、执行模式、稳定操作类别、最慢项下标、各阶段耗时、Pending和结果大小，不包含原始Key或数据。生产告警阈值使用监控系统按P95/P99、持续时间和错误率配置，无需重启DBService。

测试必须覆盖：

- 相同 Key 的 MongoDB/Redis 模拟操作严格按准入顺序执行；
- 不同 Key 在 Worker 上限内并发；
- 10000个不同Key排队时，实际数据库I/O并发和KeyExecutor Worker goroutine数量始终不超过`max_concurrent_requests`；
- 慢 Key 不阻塞无关 Key；
- 空 Key 请求可以由任意空闲 Worker 执行；
- `scheduler.max_await_tasks` 全局容量和单Key容量超限时立即拒绝；
- 排队取消、运行取消和超时不会提前释放 Key；
- Await成功但Submit失败时不残留Ticket、Key状态或Pending计数；
- Key 队列清空后删除；
- 重复停止、部分启动失败和逆序回滚；
- 停止期间拒绝新请求并在预算内排空；
- MongoDB 全部标准 Operation 的参数判别、结果映射和 BSON 边界；
- Sequential 部分成功、Transaction 回滚、Driver 事务重试和结果断言失败；
- RawCommand可选登记、固定数据库、Cursor有界读取和硬拒绝命令；
- Redis Command、Pipeline、MULTI/EXEC、Function/Script 的结果映射和逐项错误；
- Redis Cluster 同 Slot 校验、Standalone/Sentinel 扫描续页、Null，以及扁平节点表的索引、无环和嵌套容量；
- 慢排队、慢I/O、慢总请求和多命令最慢项分类正确，限频日志不泄漏原始Key或数据；
- Notify 只验证提交语义，不把它误测成远端执行确认；
- `go test -race` 和针对热点 Key、随机 Key、慢 I/O 的 Benchmark。

## 22. 待继续确认

1. 本文 MongoDB 标准操作、RawCommand、Sequential/Transaction 和结果断言设计是否确认。
2. 本文 Redis 原始命令基础单元、Pipeline、MULTI/EXEC、Function/Script 和扁平结果节点表设计是否确认。
3. 各标准MongoDB Operation的最终字段及允许Options。
4. RPC 稳定错误码、MongoDB/Redis 错误分类和部分结果的最终 Go 类型。
5. AccDBService、RoleDBService的首期完整MongoDB/Redis资源配置、Node拓扑和副本数。
