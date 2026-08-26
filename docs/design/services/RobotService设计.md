# RobotService 设计

> 状态：首期设计已确认，进入分阶段实现
> 更新日期：2026-08-26
> 上位文档：[OriginGame v3 总体架构设计](../总体架构设计.md)
> 关联文档：[登录链路监控与测试设计](../登录链路监控与测试设计.md)

## 1. 服务定位

| 层级 | 名称 |
| --- | --- |
| 生产程序 | `OriginGame`（统一程序，按 Node 配置启动） |
| Origin Service | `RobotService` |
| Go 包 | `robotservice` |
| 部署范围 | 仅隔离的开发或压测环境，不进入生产业务拓扑 |

RobotService 是真实客户端协议的自动化执行和负载生成服务。它创建大量轻量 `VirtualPlayer`，通过 LoginService HTTP 接口取得 Token 和区服入口，再使用 TCP 连接共享 Gateway，按照 OriginBlueprint 行为图完成登录、心跳、断线和后续业务动作。

RobotService 的首要目标是同时支撑：

- 少量机器人的功能自动化和进程级 E2E；
- 固定在线人数的长连接、心跳和业务混合负载；
- 后续固定到达率的登录风暴、Spike 和 Soak Test；
- 可重复、可版本管理并能由 OriginBlueprint 编辑的机器人行为。

RobotService 不属于生产登录链路，不参与玩家权威状态、服务发现路由或数据库访问。压测结果只能说明指定硬件、部署、配置、数据规模和场景下的表现，不提供脱离环境的固定 CCU 结论。

## 2. 非目标

首期不实现：

- 图形客户端渲染、物理、寻路和客户端 UI 自动化；
- 直接调用 GatewayService、GameService、AccDBService 或 RoleDBService RPC；
- 绕过 LoginService、Gateway 或客户端协议的压测捷径；
- 由蓝图创建其他机器人或改变全局负载计划；
- 任意 Lua、JavaScript 或不受限脚本节点；
- 后台 Web、AdminService、账号/RBAC、审计、实时推送和结果持久化；
- 多 RobotService 的中心 Controller 和跨节点统一调度；
- 自动故障注入、生产环境在线压测和测试数据清理平台。

真实客户端或 Unreal 蓝图机器人可以在后续作为高保真补充，但不承担大规模 CCU。

## 3. 总体结构

```text
tests/e2e/robot OriginBlueprint 工作区
    ├── nodes/       机器人节点定义
    ├── blueprints/  单个机器人行为图
    └── functions/   可复用蓝图函数
             |
             v
RobotService
└── scenario.Module（嵌入 Origin blueprintmodule.Module）
    ├── RunController               # 单次运行控制、状态快照和有界历史
    ├── Workload                    # 只负责创建速率、数量、持续时间和停止
    ├── OriginBlueprint CompiledGraph
    ├── Robot-1 -> Instance + Execution + VirtualPlayer
    ├── Robot-2 -> Instance + Execution + VirtualPlayer
    └── ...
             |
             +--> LoginService HTTP
             +--> Gateway TCP
             +--> GameService（仅经 Gateway 客户端消息转发）
```

蓝图只描述一个机器人的行为。`Workload` 负责机器人数量、升压和测试时长，禁止蓝图节点直接创建机器人，避免行为图产生无界并发或改变测试负载语义。

一个 `CompiledGraph` 由同一场景的全部机器人共享。每个机器人拥有独立 `blueprintmodule.Instance` 和一条长期 `Execution`；执行在网络、HTTP 或 Timer 节点处 `Yield`，回调通过 `Resume/ResumeTo` 投递回 RobotService 的有界 Service 队列。

### 3.1 控制接口与未来后台边界

首期不建设后台平台，也不为 RobotService 增加公网 HTTP、WebSocket 或 SSE 入口。RobotService 只在 `protocol/rpc/robotservice.go` 暴露最小 Origin 内部 RPC 契约：

```go
//origin:rpc
type RobotService interface {
	ListScenarios(context.Context, ListRobotScenariosRequest) (ListRobotScenariosResponse, error)
	StartRun(context.Context, StartRobotRunRequest) (RobotRunSnapshot, error)
	StopRun(context.Context, StopRobotRunRequest) (RobotRunSnapshot, error)
	GetRun(context.Context, GetRobotRunRequest) (RobotRunSnapshot, error)
}
```

- `ListScenarios`：返回已经加载、编译成功且包含机器人入口的场景摘要；
- `StartRun`：只接收必填幂等 `request_id` 和已加载的 `scenario_name`，人数、升压、持续时间、重试和补充策略全部使用服务端配置，不允许调用方临时突破硬上限；
- `StopRun`：按 `run_id` 幂等发起取消并立即返回 `stopping` 或已有终态快照，不等待全部机器人清理；调用方继续通过 `GetRun` 轮询终态；
- `GetRun`：按 `run_id` 查询快照；`run_id` 为空时返回当前活动运行，没有活动运行则返回最近一次结果。

`RunController` 每个 RobotService 实例同时只允许一个活动运行。并发 `StartRun` 返回明确的忙错误，不排队；`run_id` 由 RobotService 生成，`request_id` 在活动运行和最近 20 次内存快照范围内去重。运行状态固定为 `starting`、`running`、`stopping`、`succeeded`、`failed`、`canceled`，状态转换和停止均由 Service 串行上下文完成。

首期只保留最近 20 次聚合快照，不保存逐机器人明细，进程重启后允许丢失；调用方使用有界频率轮询 `GetRun`，不实现事件流。未来 AdminService 是浏览器唯一访问入口，负责 HTTP API、鉴权、RBAC、审计、运行记录持久化和多 RobotService 编排，再通过上述内部 RPC 控制 RobotService；浏览器不得直接访问 RobotService。需要远程覆盖负载参数或多 Worker 时，应先补充服务端配额与调度设计，不扩张首期 RPC。

## 4. 部署与隔离

首期使用独立 Node，例如：

```yaml
nodes:
  - id: test-robot-1 # 隔离测试环境中的机器人负载节点。
    discovery_enabled: false # 纯客户端节点，不创建 Origin 发现 Provider 或系统 RPC。
    labels: {scope: test} # 仅用于运维识别，不替代环境隔离。
    allow_discovery: [] # 机器人只走客户端外部入口，不发现内部 Service。
    services: [RobotService]
```

部署必须遵守：

- RobotService 与被测 Login、Gateway、Game、DB Node 使用不同机器或容器资源，避免负载生成器争抢被测端 CPU、内存和网络；
- 开发、压测和生产使用不同服务发现 namespace 或 network，不能依赖 `scope=test` 防止误连生产；
- RobotService 不需要发现任何 Origin Service，所有业务访问都从真实客户端入口发起；
- Robot Node 必须配置 `discovery_enabled: false`；它不加入 Origin 服务发现，也不依赖 `DiscoveryService` Node。保留的 TCP Listener 不会发布 `RobotService`，也不能作为 Origin 远端控制 RPC 的发现入口；
- 当前 `runbot` 的 Robot-only 拓扑只支持配置中的 `startup_run` 自动运行或同进程测试调用。若需要 CI 或专用控制 Node 远程调用控制 RPC，必须先单独确认静态直连控制协议或独立发现拓扑，不能重新让机器人发现被测业务 Service；
- RobotService 不监听后台 Web 端口，不允许浏览器或公网直接调用控制 RPC；
- 首期一个 RobotService Node 独立执行完整计划；只有压测机 CPU、内存、端口或带宽成为瓶颈后，才设计多 Worker 协调；
- 测试开始前必须明确目标环境、最大在线机器人、最大创建速率和最长持续时间，达到任一上限立即停止新增机器人。

## 5. 代码和资源归属

```text
protocol/rpc/
└── robotservice.go                 # RobotService内部控制RPC契约

service/robotservice/
├── robotservice.go                 # Service配置、生命周期装配和跨子包协调
├── scenario/
│   ├── module.go                   # Blueprint Module、Workload和机器人实例生命周期
│   ├── runcontroller.go            # 单活动运行、状态机、幂等启动和有界快照
│   ├── loginnode.go                # HTTP登录异步节点
│   ├── connectionnode.go           # Gateway连接和断开节点
│   ├── messagenode.go              # 登录玩家、心跳和等待消息节点
│   └── waitnode.go                 # 真实系统时间等待节点
└── virtualplayer/
    ├── player.go                   # 单个机器人状态、Token、Session和Pending所有权
    ├── loginclient.go              # LoginService HTTP客户端
    ├── gatewayclient.go            # Origin TCP Dialer和客户端封包
    └── pending.go                  # 有界请求等待与响应匹配

tests/e2e/robot/
├── originblueprint.project         # OriginBlueprint工作区设置
├── nodes/robot.json                # Go节点实现对应的编辑器契约
├── blueprints/login_heartbeat.obp  # 首个真实行为图
└── functions/                      # 出现真实复用后再增加
```

机器人节点 JSON 和行为图属于测试资产，统一放在 `tests/e2e/robot`。Go 节点实现属于 RobotService，放在 `service/robotservice/scenario`。不得把机器人业务节点加入 OriginBlueprint 的内建通用节点库。

## 6. OriginBlueprint 接入

RobotService 只导入 Origin v3 的 `sysmodule/blueprintmodule`，不直接依赖 OriginBlueprint VM 内部包。`scenario.Module` 匿名嵌入 `blueprintmodule.Module`，在 `OnInit` 中完成：

1. 从 RobotService 配置取得 `node_dir` 和 `graph_dir`；
2. 调用 `Setup` 冻结目录；
3. 使用返回新节点对象的工厂调用 `RegisterNodes`；
4. 注入 HTTP、连接、消息、Timer 和 VirtualPlayer 查找的最小接口；
5. 由 `blueprintmodule.OnStart` 一次性加载、校验并编译全部蓝图；任何节点未注册、端口不匹配或图无效都使 RobotService 启动失败。

入口 JSON 与蓝图使用 `Entrance_RobotStart_1001`，其中 `_1001` 是具体 `EntranceID`；Go 工厂按 OriginBlueprint 入口规则注册基类名 `Entrance_RobotStart`，不得把带 ID 的实例名误作工厂名。

不修改 OriginBlueprint VM 的执行模型，也不增加编辑器 `semantic_type`。蓝图只使用现有 `Integer`、`String`、`Boolean`、`Float` 和必要的扁平 `Array`；编辑者通过明确的端口名称区分对象、道具和技能句柄，RobotService 在运行时校验句柄种类、所属场景版本和有效性。Token、完整 Protobuf、网络 Session、Context、锁、Timer 和 pending 对象禁止进入蓝图变量或 `Any`。

每个 `VirtualPlayer` 启动场景时，把稳定的非零 `robot_id` 作为入口参数传入。机器人节点通过显式 `robot_id` 输入查找宿主对象，不依赖包级注册表、隐式当前玩家或未公开 VM Context。蓝图可以把 `robot_id` 保存为 instance-scope 变量，避免在每条连线上重复传递。

## 7. VirtualPlayer 状态与所有权

首期状态机：

```text
Created -> Authenticated -> Connected -> Online -> Closed
    |            |              |          |
    +------------+--------------+----------+-> Failed
```

状态含义：

| 状态 | 条件 |
| --- | --- |
| `Created` | 仅具有机器人 ID 和测试身份，未取得 Token |
| `Authenticated` | HTTP 登录成功并保存 Token、区服快照和 Gateway 地址 |
| `Connected` | 已建立 Gateway TCP Session，尚未完成 LoginPlayer |
| `Online` | 已收到 LoginPlayer 成功响应，可以发送心跳和普通业务消息 |
| `Failed` | 当前场景因确定错误结束，保留最终错误摘要等待统一回收 |
| `Closed` | Session、Pending、蓝图 Instance 和外部任务已经释放 |

`VirtualPlayer` 是普通进程内对象，不是 Origin Module。`scenario.Module` 是全部 VirtualPlayer、蓝图 Instance、I/O Job、Timer 和连接的唯一生命周期所有者。

每个 VirtualPlayer 使用两个固定 pending 槽位：一个普通业务请求，一个后台心跳请求。两者共用连接内递增的非零 `uint32 Sequence`，但分别按 `Sequence + MessageID` 匹配；同一槽位不得重叠请求，上一次心跳完成后才安排下一次。连接关闭时两个 pending 全部失败，新连接从 1 重新开始。服务端主动推送进入固定有界 inbox；首期只需覆盖 `PlayerKicked`，不得建立无界消息缓存。

Token 和客户端消息 Body 不写日志、不进入蓝图 Trace、不作为指标标签。诊断机器人使用 `robot_id`，正式测试账号使用不包含凭证的稳定前缀和序号。

复杂业务状态由每个 VirtualPlayer 自己的 `RobotStateStore` 持有，至少按真实需求分为背包、场景和战斗等有类型的快照。快照记录单调递增的领域 `revision`；场景切换还必须更新 `scene_epoch`。StateStore 是服务器权威状态的机器人侧镜像，不是新的权威数据源，连接关闭、重新登录和场景切换时必须按协议语义清理失效状态。

完整协议对象、对象列表和嵌套结构只保存在 StateStore。蓝图节点只输出稳定 ID、不透明句柄、标量、小型同类数组、结果数量和 `revision`。场景对象句柄必须包含或关联创建时的 `scene_epoch`；动作和查询节点每次使用句柄时重新校验，旧场景句柄不能命中新场景复用的 ObjectID。每个领域快照、列表和单次查询结果都必须有硬上限，超过上限明确失败，不得静默截断或建立无界缓存。

### 7.1 协议等待与蓝图恢复

等待服务器协议是机器人节点的基础异步能力，统一由 VirtualPlayer 的 `ProtocolDispatcher` 实现，不允许每个节点自行轮询 Session：

```text
蓝图节点 Yield
  -> 先登记 WaitSpec
  -> 再发送客户端消息
  -> Gateway 收包并投递 InboundEnvelope
  -> 解码领域事件并由 StateReducer 更新 RobotStateStore
  -> ProtocolDispatcher 匹配响应或重新判断状态等待条件
  -> ResumeTo(成功/失败出口)
  -> 后续蓝图流程
```

`WaitSpec` 至少包含机器人 ID、等待类型、期望 MessageID、可选 Sequence、真实系统时间 Deadline 和一次性 YieldHandle。请求响应必须同时匹配 `Sequence + MessageID`；服务端主动推送按 `MessageID` 匹配，并由具体协议节点执行必要的字段条件判断。禁止只按 MessageID 匹配普通响应，否则并发或迟到响应可能恢复错误执行。

发送并等待响应必须封装在同一个异步节点中，并严格“先登记、后发送”，避免服务器立即响应时出现丢包窗口。`RobotLoginPlayer`、`RobotHeartbeat` 以及后续业务动作节点都复用该机制。蓝图不得使用一个 Send 节点和另一个 Wait 节点拼接普通请求响应。

`RobotWaitMessage` 只用于服务器主动推送。收包时没有活动 waiter 的推送可以进入固定容量的可消费 inbox；Wait 节点先消费尚未被当前场景处理的匹配消息，再登记新 waiter。每条 inbox 消息带单调递增的接收序号且最多消费一次，避免同一旧推送重复恢复多个节点。inbox 满时返回明确过载错误并计数，不能静默覆盖关键推送。

网络收包回调只校验帧头、限制 Body 大小并向 RobotService 有界队列投递 `InboundEnvelope`。waiter 注册、匹配、超时和取消都在 Service 串行调度中完成。协议到达、超时、场景取消和连接关闭竞争时只有第一个事件可以完成 waiter；其余事件只做幂等清理。任何完成路径最终都通过 blueprintmodule 的有界恢复队列执行 `ResumeTo`，不得从网络协程直接继续蓝图。

协议消息必须先完成解码和 StateStore 更新，再恢复响应 waiter 或状态 waiter，保证后续蓝图节点立即读取到同一消息产生的新状态。请求响应仍使用 `Sequence + MessageID` 确认完成；业务流程需要等待背包数量、场景就绪、对象出现或技能状态时，优先等待 StateStore 条件，不依赖裸 MessageID。

超时、连接关闭、协议解码失败、等待队列满和恢复队列拒绝分别使用固定低基数错误类型。日志可以记录 robot ID、MessageID、Sequence、等待时长和错误类型，但不能记录 Token 或原始消息 Body。

### 7.2 状态等待、查询与选择

状态等待统一使用有界 `StateWaitSpec`。注册操作在 Service 串行上下文中先检查当前快照，条件已满足则立即完成；否则记录领域、当前或指定 `revision`、有类型的条件、真实系统时间 Deadline 和一次性 YieldHandle。StateReducer 每次发布新 revision 后只重新判断对应领域的 waiter，避免轮询和丢失“检查与登记之间”到达的更新。

等待语义必须明确区分：

- `current_or_next`：当前状态已满足即可成功，否则等待后续更新；
- `next_update`：只接受登记后更大的 revision，用于等待本次动作之后的新状态。

查询节点只读当前不可变快照并返回有界候选句柄数组；选择节点是无副作用的确定性计算。候选集合必须按稳定字段排序，空集合走明确失败出口。需要随机选择时使用 `run_seed + robot_id` 派生的确定性随机源并记录 seed，不能依赖 Go map 遍历顺序或进程级随机状态。

## 8. 首期蓝图节点契约

所有网络、HTTP 和 Timer 节点都是有副作用的 Exec 节点，必须接入执行流。异步节点成功 `Yield` 后立即返回 `ErrExecutionSuspended`，callback 只保留普通值和一次性 `YieldHandle`，不得访问节点端口或 `BaseExecNode`。

### 8.1 `RobotHTTPLogin`

输入：`exec`、`robot_id(Integer)`。

输出：`succeeded(exec)`、`failed(exec)`、`error_code(Integer)`、`duration_ms(Integer)`。

通过 LoginService `POST /api/v1/login` 使用当前测试身份登录。成功后 Token、区服列表和 Gateway 地址只保存到 VirtualPlayer；蓝图不输出 Token。

### 8.2 `RobotConnectGateway`

输入：`exec`、`robot_id(Integer)`。

输出：`succeeded(exec)`、`failed(exec)`、`error_code(Integer)`、`duration_ms(Integer)`。

首期只支持 TCP，使用 HTTP 登录响应中当前 `show_area_id` 对应的 Gateway 地址。连接由 VirtualPlayer 持有，节点不建立自动无限重连。

### 8.3 `RobotLoginPlayer`

输入：`exec`、`robot_id(Integer)`、`show_area_id(Integer)`。

输出：`succeeded(exec)`、`failed(exec)`、`error_code(Integer)`、`duration_ms(Integer)`。

按客户端 Protobuf 契约发送 `LoginPlayerReq`，等待相同 Sequence 的 `LoginPlayerRes`。成功后 VirtualPlayer 进入 `Online`；失败时不向蓝图暴露 Token 或响应 Body。

### 8.4 `RobotStartHeartbeat`

输入：`exec`、`robot_id(Integer)`、`interval(String)`。

输出：`succeeded(exec)`、`failed(exec)`、`error_code(Integer)`。

只允许 `Online` 状态启动。节点登记 VirtualPlayer 后台心跳后立即返回，不创建第二条蓝图 Execution；后续按真实系统时间周期发送 `PlayerHeartbeatReq`，并等待相同 Sequence 的 `MessageID=Ok`。心跳超时、协议失败或连接关闭时必须标记当前 attempt 失败、关闭连接并取消主蓝图。机器人、Run 或 Module 停止时自动取消心跳，首期不增加必须显式连接的 Stop 节点。

### 8.5 `RobotHeartbeat`

输入：`exec`、`robot_id(Integer)`。

输出：`succeeded(exec)`、`failed(exec)`、`error_code(Integer)`、`duration_ms(Integer)`。

只允许 `Online` 状态发送 `PlayerHeartbeatReq`，等待 `MessageID=Ok` 和相同 Sequence。

`RobotHeartbeat` 保留为单次心跳调试节点，不再用于默认在线蓝图的长期循环。

### 8.6 `RobotWait`

输入：`exec`、`duration(String)`。

输出：`completed(exec)`、`failed(exec)`、`error_code(Integer)`。

`duration` 使用带单位字符串，例如 `500ms`、`5s`、`2m`。该等待属于测试基础设施真实系统时间，不受 Node 游戏逻辑时间调整影响。节点通过 scenario.Module 持有的专用真实时间调度器恢复；调度器可以使用 `time.NewTimer`，但必须登记、可取消并在 Module 停止时统一等待，不得使用 Origin 业务 Timer 或 `time.Sleep`。

### 8.7 `RobotWaitMessage`

输入：`exec`、`robot_id(Integer)`、`message_id(Integer)`、`timeout(String)`。

输出：`received(exec)`、`timeout(exec)`、`failed(exec)`、`error_code(Integer)`、`duration_ms(Integer)`。

等待指定服务端主动推送。首期用于验证 `PlayerKicked`；一个机器人同时最多存在一个 WaitMessage。

### 8.8 `RobotDisconnect`

输入：`exec`、`robot_id(Integer)`。

输出：`completed(exec)`。

幂等关闭当前 Session，并使全部 pending 得到确定失败；连接不存在时仍成功。

### 8.9 后续复杂业务节点扩展规则

复杂场景按以下四类增加 RobotService 工作区节点，不增加通用脚本节点：

| 类型 | 示例 | 职责 |
| --- | --- | --- |
| 动作节点 | `RobotEnterBattle`、`RobotUseItem`、`RobotCastSkill` | 发送一个业务请求并等待对应响应，成功响应先更新 StateStore 再恢复蓝图 |
| 状态等待节点 | `RobotWaitSceneReady`、`RobotWaitItemCount`、`RobotWaitEntityState` | 使用有类型条件等待当前或后续 revision，不轮询网络 Session |
| 查询节点 | `RobotQueryItems`、`RobotQueryTargets`、`RobotQueryUsableSkills` | 从 StateStore 当前快照返回有界候选句柄数组 |
| 选择节点 | `RobotSelectItem`、`RobotSelectTarget`、`RobotSelectSkill` | 按显式稳定策略从候选集合选择一个句柄，不产生外部副作用 |

动作节点只表示服务器已经接受或拒绝本次请求；如果业务状态由稍后的主动推送更新，蓝图应在动作节点后显式连接状态等待节点。查询和选择节点不能隐式发送请求；需要刷新数据时先使用对应动作节点。节点输出不得包含完整 Protobuf、嵌套对象或原始 Body。

对象、道具和技能句柄继续使用底层 `Integer` 或 `String` 端口，并在端口名称中明确写出 `entity_handle`、`item_handle` 或 `skill_handle`。OriginBlueprint 不负责区分这些业务语义；每个 RobotService 节点必须在执行前校验句柄类型和有效期，错误连接走确定的失败出口。

节点 `name`、输入输出 `port_id`、数据类型和顺序一经发布即成为兼容契约。后续不兼容调整必须增加版本化新节点，不能直接重排端口。

## 9. 首期行为图

`login_heartbeat.obp` 使用单一入口，例如 `Entrance_RobotStart_1001`：

```text
RobotStart(robot_id)
  -> RobotHTTPLogin
  -> RobotConnectGateway
  -> RobotLoginPlayer(show_area_id)
  -> RobotStartHeartbeat(5s)
  -> 主业务循环到场景Context结束
       -> RobotWait(1s) # 首期占位，后续替换为战斗子流程
  -> RobotDisconnect
```

心跳是 VirtualPlayer 持有的连接基础设施，不使用通用 `Fork/Parallel/Join` 节点，也不与战斗主流程共用蓝图等待链。任一登录、连接或心跳失败记录低基数阶段和错误类型后关闭连接并取消主蓝图。蓝图不得包含无退出条件的非结构化循环；长期循环必须同时受场景 Context、测试持续时间和 Execution 步数预算限制。

异步节点首次走失败出口时，RobotService 必须在当前 attempt 上保留失败原因；失败分支即使通过 `RobotDisconnect` 正常完成清理，也不能被 VM 的正常收尾误记为成功。场景重试开始时清除本次 attempt 错误，但保留首次失败标记，重试成功仍统计为 `flaky`。

后续战斗场景按状态驱动方式组合，例如：

```text
RobotEnterBattle
  -> RobotWaitSceneReady
  -> RobotQueryTargets(enemy, alive)
  -> RobotSelectTarget(lowest_hp)
  -> RobotQueryUsableSkills(target)
  -> RobotSelectSkill(highest_priority)
  -> RobotCastSkill(skill, target)
  -> RobotWaitEntityState或RobotWaitCombatEvent
  -> 判断战斗是否结束并有界循环
```

背包场景先通过请求或推送更新 InventoryState，再执行 `RobotQueryItems -> RobotSelectItem -> RobotUseItem`；需要确认数量变化时继续连接 `RobotWaitItemCount(next_update)`。`ChooseCombatAction`、`EnsureItemAndUse` 等多节点流程出现真实复用后保存为 OriginBlueprint 函数，不在单个 Go 节点中固化整段业务。

## 10. Workload 与负载语义

首期只实现固定在线人数的闭环计划：

- `users`：目标同时活动机器人数量，也是硬上限；
- `ramp_up`：从零提升到目标数量的真实系统时间；
- `duration`：达到目标后保持的真实系统时间；
- Robot 完成或失败后是否补充，由计划固定策略决定，不由蓝图决定；首期功能 E2E 不补充，固定 CCU 场景可以显式开启补充。

开启补充时必须同时配置非零 `max_replacements`。补充启动速率不得高于初始升压速率，达到替换次数、用户数或持续时间任一上限后停止新增；不能因目标持续失败形成无限快速重连。

第二阶段实现尚未包含失败补位，因此配置校验强制 `replace_failed=false`、`max_replacements=0`；第三阶段实现补位速率限制和计数后再开放该开关，禁止在未生效时静默接受配置。

功能 E2E 可以为同一个机器人配置 `scenario_retry_count`，取值范围为 `0` 到 `2`，默认 `0`。重试前必须完整关闭旧 Session、取消 pending 和挂起执行，再以同一测试身份从行为图入口重新开始。首次失败始终记录；重试后成功标记为 `flaky`，不能覆盖成纯成功。需要验证断线恢复时，不使用场景重试或替换机器人，而由蓝图显式驱动同一机器人的有界重连。

后续登录风暴必须增加独立的开放到达率计划，使新登录开始速率不受目标响应变慢影响，并同时配置 `max_active_robots`、最长测试时间和丢弃启动统计。该能力不与首期闭环计划混在一次实现中。

本地 LoginService 当前有单 IP 登录限流。限流验收场景保留真实配置；容量场景必须使用多个来源地址，或在隔离环境中显式调整限流并把配置快照写入结果，RobotService 不得绕过服务端限流。

## 11. 配置边界

RobotService 配置集中放在平铺的 `config/robotservice.yaml`，首期至少包含：

```yaml
services:
  RobotService:
    blueprint:
      node_dir: "tests/e2e/robot/nodes" # 机器人节点定义目录。
      graph_dir: "tests/e2e/robot/blueprints" # 机器人行为蓝图目录。
      graph_name: "login_heartbeat" # 首期执行的普通蓝图文件名，不含扩展名。
      entrance_id: 1001 # 机器人启动入口ID。

    control:
      startup_run: true # true时启动默认场景；false时Ready后空闲等待内部RPC。

    target:
      login_url: "http://127.0.0.1:8080/api/v1/login" # 被测LoginService真实HTTP入口。
      show_area_id: 1 # 机器人选择的显示区服。

    identity:
      platform_type: 1 # 开发鉴权平台类型；生产凭证不得写入本地配置。
      platform_id_prefix: "robot-" # 测试身份前缀，后接稳定机器人序号。

    workload:
      users: 100 # 同时活动机器人硬上限。
      ramp_up: 30s # 从零提升到目标人数的真实系统时间。
      duration: 10m # 达到目标人数后的保持时间。
      scenario_retry_count: 0 # 同一机器人完整场景失败后的重试次数，仅允许0到2。
      replace_failed: false # 首期E2E默认不补充失败机器人。
      max_replacements: 0 # 允许补充的机器人总数硬上限；开启补充时必须大于0。

    io:
      workers: 32 # HTTP登录和拨号阻塞Job的固定Worker数量。
      queue_messages: 256 # 等待执行的I/O Job数量硬上限。
```

配置字段只保存测试计划和非敏感目标地址。RPC 启动只能选择已加载场景，负载参数仍以本配置为唯一来源。真实平台 Token、密码和密钥由隔离测试环境注入，不得进入蓝图文件、示例配置、日志或结果报告。

## 12. I/O 与并发边界

HTTP 登录和主动拨号可能阻塞，不得在 RobotService 工作协程或蓝图节点 `Exec` 中直接等待。`scenario.Module` 创建一个固定 Worker 数和有界 Job 队列的 I/O Executor：

- 节点先 `Yield`，再提交拥有 Context、机器人 ID 和完成回调的 Job；
- 队列满立即走失败出口，不创建额外 goroutine 或无限重试；
- Worker 只操作并发安全的 HTTP Client、Origin Dialer和 VirtualPlayer I/O 边界，不直接修改 Service 串行状态；
- 完成回调调用一次性 `ResumeTo`，后续蓝图片段由 blueprintmodule 投递回 Service；
- Module 停止时先停止准入并取消全部 Job Context，再等待全部 Worker 退出。

TCP Session 的收包回调由 Origin 网络 Runtime 投递到 RobotService。回调只完成协议解析、Sequence匹配、有界 inbox 更新和 YieldHandle 恢复，不执行长时间业务逻辑。

## 13. 生命周期与 Ready

### 13.1 启动

RobotService 只有在以下条件全部成立后才完成启动：

- 配置、测试身份规则和目标 URL 校验通过；
- I/O Executor 容量合法并已启动；
- OriginBlueprint 节点目录、行为图和函数依赖全部加载编译成功；
- 首期目标图和入口 ID 存在；
- TCP Dialer配置和客户端封包器初始化成功。

OnStart 不直接批量拨号。全部依赖准备完成后，`control.startup_run=true` 时预登记一个零延迟 Module Timer；Origin 在全部 Service 的 OnStart 成功并激活调度器后，该 Timer 才通过 `RunController` 启动默认场景。这避免在 Service 仍为 Starting 时投递普通任务。`false` 时以空闲可控状态完成 OnStart，发布 Ready 后等待内部 RPC。自动启动和 RPC 启动必须进入同一个 `RunController.Start` 校验与状态转换路径。Workload 调度器只向有界队列投递机器人启动命令，不同步创建大量连接；OnStart 阶段的资源准备或 Timer 预登记失败会使 RobotService 启动失败并执行逆序回滚，激活后的运行失败记录在该 Run 的终态快照中。

### 13.2 停止

停止顺序固定为：

1. 停止 Workload 准入并取消升压、保持和补充 Timer；
2. 取消场景根 Context，阻止新的 HTTP、拨号和消息请求；
3. 关闭所有蓝图 Instance，取消挂起 Execution；
4. 幂等关闭所有客户端 Session，使 pending 得到确定失败；
5. 关闭 I/O Job 队列，取消在途 Job 并等待全部 Worker；
6. 关闭 blueprintmodule 引擎并清空 VirtualPlayer 索引。

任一步重复执行必须安全。启动中途失败按已经成功创建资源的逆序执行同样的清理。

## 14. 指标与结果

RobotService 侧首期至少记录：

- 当前机器人状态数量；
- 各阶段开始、成功和失败数量；
- HTTP登录、连接、LoginPlayer和心跳的 P50/P95/P99；
- I/O Worker使用量、队列深度和队列拒绝数量；
- 蓝图活动 Instance、Execution失败和恢复队列拒绝；
- 实际创建速率、目标在线数、达到目标所需时间；
- 负载生成器自身CPU、内存、goroutine、网络和端口使用量。

指标标签只使用场景、阶段、结果、传输类型等固定低基数枚举。`robot_id`、账号、Token、Sequence、ConnectionID和测试运行 ID不作为指标标签。单次测试报告可以保存运行ID、Git提交、配置快照、硬件、起止时间和聚合分位数。

服务端仍使用既有 LoginService、GatewayService、GameService、DBService和Origin网络指标。容量结论必须同时检查负载生成器是否先达到资源瓶颈。

## 15. 测试与验收

### 15.1 单元测试

- VirtualPlayer状态转换、重复关闭和非法阶段拒绝；
- Sequence回绕、单pending限制、迟到响应和连接关闭；
- 客户端请求/响应/主动推送封包与服务端契约逐字节一致；
- 每个蓝图节点的成功、失败、超时、取消、重复callback和Resume队列拒绝；
- Workload升压、停止、失败补充速率、替换次数和全部硬上限；
- RunController并发启动拒绝、request_id幂等、停止幂等、状态转换和20条快照淘汰；
- I/O队列满、部分Worker启动失败、停止取消和goroutine回收。

### 15.2 集成测试

- Origin blueprintmodule真实Service调度下的Yield/Resume；
- Origin TCP Dialer连接Gateway并完成真实LoginPlayer；
- HTTP开发鉴权创建独立测试账号并取得区服入口；
- 顶号推送、心跳、断线和迟到响应；
- RobotService停止时全部蓝图Instance、Session、Timer、pending和Worker释放。
- RobotService控制RPC只能启动已加载场景，不能覆盖服务端负载硬上限；

### 15.3 E2E与性能

- 1、10、100机器人完整登录心跳流程；
- 固定CCU逐级升压，记录客户端和服务端P50/P95/P99；
- Gateway 4KB消息边界和10秒200条连接限流；
- Redis、DBService或GameService故障时机器人得到确定失败，恢复后新机器人可以登录；
- `go test -race`覆盖RobotService、blueprintmodule交互和连接生命周期；
- 保存单机负载生成器Benchmark，确认目标服务器之前没有由RobotService CPU、队列或端口先饱和。

## 16. 分阶段实施

### 第一阶段：OriginBlueprint工作区能力

- 工作区打开或切换时加载 `<workspace>/nodes`；
- F5校验使用同一工作区节点定义；
- 建立 `tests/e2e/robot` OriginBlueprint工作区和首批节点JSON。

### 第二阶段：首个RobotService闭环

- TCP、HTTP登录、LoginPlayer、Wait、Heartbeat和Disconnect节点；
- 单个RobotService Node、固定用户数和线性升压；
- 单活动运行的RunController、内部控制RPC、自动运行和空闲等待两种启动方式；
- `login_heartbeat.obp`、单元测试、集成测试和最小E2E。

### 第三阶段：真实压测能力

- 开放到达率、Spike、Soak和有界失败补充；
- 按真实协议增加有界 RobotStateStore、StateReducer、动作/状态等待/查询/选择节点与场景函数；
- 完整结果报告、服务端指标关联和容量拐点分析。

只有单机负载生成器实测成为瓶颈后，才设计多 RobotService Worker和中心协调。

后台 Web 管理平台、鉴权/RBAC、审计、结果持久化和多 RobotService 编排后续作为独立设计，不纳入 RobotService 首期实现。

## 17. 已确认结论

已确认：

- 使用 `E:\Develop\Origin开发\OriginBlueprint` 编辑机器人行为；
- 允许为真实需要修改 OriginBlueprint；
- RobotService 复用 Origin v3 `blueprintmodule`；
- 复杂业务状态保存在每个VirtualPlayer的有界RobotStateStore，协议先更新状态再恢复蓝图；
- 蓝图只传稳定ID、不透明句柄、标量和小型同类数组，复杂流程由动作、状态等待、查询和选择节点组合；
- 本轮不修改OriginBlueprint或增加`semantic_type`，由编辑者通过端口名称规避错误连接，RobotService运行时必须校验句柄种类、场景版本和有效性；
- 机器人蓝图与负载计划分离；
- 首期使用独立测试Node、TCP和登录心跳闭环；
- 首期不实现中心Controller和分布式Worker；
- 首期不实现后台Web和AdminService，只提供列场景、启动、停止和查询状态四个内部控制RPC；
- 每个RobotService实例只允许一个活动运行，调用方轮询状态，最近20次结果仅保存在内存；
- RPC只能选择已加载场景，负载参数和硬上限仍由服务端配置决定；未来后台作为唯一HTTP与鉴权边界，浏览器不得直连RobotService；
- 第8节首批节点名称、端口和入口 `Entrance_RobotStart_1001` 冻结为首期契约；
- 第11节配置结构作为首期实现基线；
- 功能E2E默认不补充失败机器人，可对同一机器人执行最多两次场景重试并保留首次失败；
- 固定CCU压测显式开启有界补充，并要求非零替换总数上限；
- 断线恢复由蓝图驱动同一机器人有界重连，不通过Workload替换；
- 测试计划结束后RobotService保持运行并输出结果，由现有进程控制入口统一停止，Service不控制进程。
