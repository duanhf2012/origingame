# DBService 设计

> 状态：职责、部署模型与有序执行模型已确认；具体 RPC 操作契约待继续设计
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

DBService 将提供业务无关的 MongoDB 与 Redis 操作契约，但具体开放哪些结构化 CRUD、批量操作、Redis 命令和 Function/Script 接口仍待后续确认。当前只确认所有 Redis 访问，包括在线路由和登录限流所需的原子操作，都必须经过 DBService；不提前确认 MongoDB 任意 `RunCommand` 等高风险能力。

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

DBService 采用“动态 Key FIFO + 固定共享 Worker 池”方案：

```text
                    ┌─ Key A FIFO: A1 -> A2 -> A3
RPC -> KeyExecutor ─┼─ Key B FIFO: B1
                    ├─ Key C FIFO: C1 -> C2
                    └─ 空 Key公共队列: X1、X2
                              |
                       固定 N 个 I/O Worker
                              |
                    MongoDB Module / Redis Module
```

执行规则：

- 非空 Key 拥有一个逻辑 FIFO，同一个 Key 同时最多只有一个请求执行；
- MongoDB 和 Redis 请求共用同一套 Key FIFO，按 KeyExecutor 成功准入顺序交叉串行；
- 不为每个 Key 创建 goroutine，也不为每个请求创建额外 I/O goroutine；
- 全部 Key 共享固定数量的 Worker，Worker 直接执行实际 MongoDB 或 Redis I/O；
- 一个请求实际 I/O 返回后，才释放该 Key 并使同 Key 下一个请求成为可运行；
- 同 Key 下一个请求重新进入公共可运行队列尾部，避免热点 Key 连续独占 Worker；
- Key 队列清空后立即删除对应状态，不形成永久增长的 Key Map；
- 空 Key 请求直接进入公共无序队列，由下一个空闲 Worker 取得；
- Worker 数量运行期间冻结，修改后必须重启并按生命周期排空旧实例。

本方案不采用 `hash(key) % worker_count` 的固定私有 Worker 队列。固定分片虽然简单，但慢 Key 会阻塞碰巧哈希到同一 Worker 的无关 Key，也可能出现一个队列已满而其他 Worker 空闲的问题。动态 Key FIFO 保留固定资源上限，同时减少无关 Key 的队头阻塞；实现复杂度集中封装在 DBService 的专用执行能力中。

## 7. 容量和过载

DBService 的所有排队和并发必须有界，至少具有：

- 固定 `worker_count`；
- 全局未完成请求上限；
- 单个非空 Key 的未完成请求上限；
- Origin `scheduler.max_tasks`；
- Origin `scheduler.max_await_tasks`；
- RPC 请求 Payload 上限；
- 单次数据库结果和批量操作上限；
- 单次请求 Deadline。

全局未完成请求包含 KeyExecutor 已接受但尚未完成的排队和运行请求。达到全局或单 Key 上限时立即返回稳定的过载错误，不等待队列腾空，不临时增加 Worker，也不创建补偿 goroutine。

容量必须作为一套预算联合核算。原则上：

```text
executor.max_pending_total <= scheduler.max_await_tasks
scheduler.max_tasks >= executor.max_pending_total + 少量准入与恢复余量
```

具体数值必须结合单请求最大内存、MongoDB/Redis 连接池、数据库可承受并发以及 P95/P99 压测确定，不在设计阶段猜测固定默认值。

## 8. 请求执行、取消与返回

一次 RPC 的内部流程为：

```text
RPC Runtime
    -> Origin ServiceScheduler 准入根任务
    -> DBService 校验并完整拥有请求数据
    -> KeyExecutor.Submit
    -> Handler 使用 Await 等待 Ticket
    -> 固定 Worker 执行 MongoDB/Redis I/O
    -> Ticket 完成
    -> 原 RPC Task 进入 Origin 恢复 FIFO
    -> 返回响应
```

KeyExecutor 成功接收请求后，如果 Handler 无法进入 Await，必须撤销尚未执行的 Ticket；已经开始执行的请求则继续由执行器持有到 I/O 实际返回并完成资源清理。

排队请求在开始前已经取消或超时时，可以从逻辑 Key FIFO 跳过并释放容量。正在执行的请求收到取消或超时时，必须取消传给 Driver 的 Context，但不能立即释放 Key；只有实际 MongoDB/Redis 调用返回后，才能启动同 Key 下一个请求，避免前一个 I/O 尚未结束时重新形成同 Key 并发。

调用方可能已经因 Deadline 收到错误，而数据库操作的最终状态仍不确定。DBService只返回实际可观察结果，不根据超时推断写入一定未发生；业务层按自身需要使用幂等 Key、数据库原子条件或查询确认。

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
- 必需数据库能力和 Redis 原子 Function/Script 版本校验，具体方式随 RPC 契约继续设计。

任一步失败必须按成功创建资源的逆序回滚，DBService 不以缺少 MongoDB、Redis 或 Worker 的半初始化状态对外服务。

停止时先由 Origin 关闭新 RPC 根任务准入。KeyExecutor 必须继续接收并处理停止边界前已经进入 Service Ready FIFO 的根任务，不能让这些已接受请求因内部执行器提前关闭而失败；这些任务及其 KeyExecutor 工作在停止 Context 预算内共同排空。全部已接受根任务完成并进入 `OnStop` 后，关闭 KeyExecutor 准入并等待固定 Worker 退出，再关闭 Redis、MongoDB Module。停止预算耗尽时由生命周期 Context 取消排队和运行请求，Worker 仍须等实际 Driver 调用返回后退出。`Stop` 必须幂等。

## 11. 监控与测试

至少记录以下指标：

- Origin ServiceScheduler 的 Accepted、Ready、Awaiting、高水位和拒绝数；
- KeyExecutor 全局 Pending、活动 Key、可运行 Key、运行 Worker 和拒绝数；
- 单 Key Pending 高水位，但日志和指标不得暴露敏感原始 Key；
- 排队时间、MongoDB/Redis 执行时间及 P95、P99；
- MongoDB/Redis 操作成功、错误、超时和结果大小；
- 停止排空耗时和被取消请求数。

测试必须覆盖：

- 相同 Key 的 MongoDB/Redis 模拟操作严格按准入顺序执行；
- 不同 Key 在 Worker 上限内并发；
- 慢 Key 不阻塞无关 Key；
- 空 Key 请求可以由任意空闲 Worker 执行；
- 全局和单 Key 容量超限立即拒绝；
- 排队取消、运行取消和超时不会提前释放 Key；
- Key 队列清空后删除；
- 重复停止、部分启动失败和逆序回滚；
- 停止期间拒绝新请求并在预算内排空；
- `go test -race` 和针对热点 Key、随机 Key、慢 I/O 的 Benchmark。

## 12. 待继续确认

1. MongoDB 结构化操作类型、请求和响应字段，以及是否开放受限批量操作。
2. Redis 通用命令、Pipeline、Transaction、Function/Script 的具体契约和安全限制。
3. MongoDB 是否开放任何形式的 `RunCommand`；当前未确认。
4. 集合名、Redis 命令、请求大小、结果大小和批量数量的允许规则。
5. DBService RPC 错误码、数据库错误映射和超时后的结果表达。
6. KeyExecutor 的公平队列细节及具体配置字段名称。
7. AccDBService、RoleDBService 的首期配置、Node 拓扑和副本数。
