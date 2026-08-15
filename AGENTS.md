# OriginGame v3 开发指导原则

## 适用范围

本文件适用于 OriginGame v3 仓库中的架构设计、Service 设计、实现、重构、测试、文档和评审工作。已经由开发者确认的规则和设计结论是后续工作的共同约束；需要改变结论时，应先说明原因和影响并重新确认。

## 协作沟通规则

- 与开发者讨论、汇报和确认时默认使用简洁输出，先给结论、接口或需要确认的选项，再补充最少量的必要理由；不得重复已经确认的背景和结论；
- 接口和配置应优先直接展示，并为名称仍不足以表达语义的函数、字段和阶段添加简短中文注释，使开发者不阅读大段说明也能理解职责、调用时机和关键约束；
- 除非开发者明确要求详细分析，避免一次展开大量实现细节、边缘场景和未来扩展。确有复杂内容需要展开时，先提供摘要，再询问是否继续；
- 需要开发者决定时集中列出少量明确问题，每个问题给出推荐项和一句主要理由，不输出冗长的重复论证。

## 设计原则

### 1. 先复核已经确认的设计

讨论、设计或实现功能前，应先检查现有设计文档、相关代码和必要的历史结论，并区分已经确认、尚未确认、文档冲突和代码事实。

已经确认且仍适用的结论直接继承，不重复讨论或静默改变。发现冲突时，应先列出冲突来源、影响和建议处理方式，经开发者确认后再修改。确认后的新结论应回写到对应的单一主设计文档，避免同一主题长期保留多套说法。

### 2. 总体架构与各 Service 分别设计

总体架构、各个 Service、协议和跨服务公共机制应分别建立边界清晰的设计文档。

Service 设计文档只展开自身职责、数据、生命周期和内部实现。跨服务协作只描述必要的接口、契约、依赖和数据流，通过引用关联其他文档，不在一个文件中重复展开所有 Service 的细节。

### 3. 重大设计变更必须先确认

总体架构、Service 职责、跨服务流程、外部协议、数据库结构以及重要一致性策略的新增或变更，必须先完成设计讨论和确认，再实施代码。

局部实现细节可以在不改变已确认设计和外部契约的前提下直接完成。实现中发现现有设计无法落地或存在缺口时，应暂停受影响部分，先更新设计并确认；不得把临时决定只保留在代码中。

### 4. 默认不为未来需求增加抽象

只为当前已确认的真实需求建立接口、工厂、适配层、扩展点和配置项。类型、函数和字段默认保持最小可见范围，只有确有跨包使用需求时才公开。

接口应小而明确，优先由使用方定义。不为了以后可能替换或可能复用提前增加层级；真实需求出现后再基于现有代码重构。

### 5. 保持代码精简、清晰和可维护

避免不必要的抽象、层级、依赖和技巧性实现。可读性、可维护性与性能、兼容性或开发效率发生明显冲突时，应列出可选方案、影响和依据，由开发者确认取舍。

### 6. 游戏服务器热路径必须考虑性能和低延迟

设计时应区分冷路径和热路径。RPC、序列化、Service 调度、Timer、队列和网络收发等热路径，应检查不必要的反射、字符串查找、堆分配、接口装箱、数据复制、锁竞争、Channel 跳转和 goroutine 调度。

不得为了可能更快而提前引入 `unsafe`、复杂对象池、无界缓存或难以验证的无锁算法。复杂优化应通过 Benchmark、逃逸分析、Profile、Trace 或可重复的端到端测试验证，并关注 P95、P99 等尾延迟，而不只关注平均吞吐量。

### 7. 禁止包级可变全局状态

Application、Node、Service、连接、缓存和其他运行时状态必须由明确实例持有。包级只允许常量、只读构建信息和不会形成跨实例污染的不可变数据。

业务包不得通过 `init()` 隐式读取配置、修改运行时注册表、创建资源或启动 goroutine。最终可执行程序可以按 Origin v3 已确认的方式装配当前 Application，但不得形成业务包级隐藏状态。

### 8. 所有资源必须有明确所有者

每个 goroutine、连接、订阅、Timer、队列项和 pending 对象都必须明确由谁创建、如何取消、何时退出以及由谁等待或释放。禁止无法停止和等待的 fire-and-forget goroutine。

`Close`、`Stop` 和 `Cancel` 应尽量幂等。启动过程中任何一步失败，都必须按已经成功创建资源的逆序清理。

### 9. 队列、并发和重试必须有界

不得创建无限队列、无限 goroutine、无限 pending、无限缓存或业务层无限重试。达到容量上限、连接失败或过载时必须返回明确错误，不得静默丢失关键业务。

重试必须具有明确退避、Context、退出条件以及次数或时间上限。Origin v3 已明确允许持续恢复的基础设施能力，可以遵循框架契约持续恢复，但仍不得无限累积资源或待处理消息。

### 10. Service 准备完成后才能对外服务

所有 Service 在启动阶段必须完成必要依赖连接、基础数据加载和业务校验，只有确认具备正常处理请求的条件后才能发布 Ready 并对外提供服务。

关键依赖不可用、必要数据为空或数据不合法时必须启动失败，并完成资源回滚。不得先开放端口或注册发现，再异步等待关键数据准备完成。非关键缓存可以在设计明确允许时后台刷新，但必须有确定的初始可用数据。

### 11. 库和 Service 代码不得控制进程

库、Service 和业务代码禁止调用 `os.Exit`、`log.Fatal` 或其他绕过清理流程的退出函数。可预期失败通过 error 返回；只有最终程序入口可以在资源清理完成后决定退出码。

错误应保留定位问题所需的上下文，但不得包含密码、Token、密钥、连接串中的凭证或其他敏感数据。

### 12. 严格管理第三方依赖和生成代码

只增加当前功能确实需要的第三方依赖。引入重要依赖前，应检查维护状态、许可证、依赖规模、性能特征、资源退出能力和替代方案。

生成代码不得手工修改。Go、Protobuf 和其他影响生成结果的工具版本必须固定，相同输入应得到相同结果。

### 13. 配置必须明确单位和安全边界

面向使用者的时间长度配置使用带单位字符串，例如 `500ms`、`15s`、`2h`；字段名使用 `timeout`、`interval`、`ttl`，不把单位编码进字段名。

字节容量配置使用明确的带单位字符串。不同方向或作用域的容量字段必须使用能表达对象和方向的名称，避免笼统的 `queue_size`。

敏感配置当前可以按已确认方案保存在示例配置中供使用者学习和替换，但生产代码不得把密码、Token、密钥等敏感值写入日志或错误信息。

### 14. 测试与实现同步完成

普通业务逻辑应提供单元测试；并发和生命周期代码应执行 `go test -race`；资源组件应覆盖重复关闭、部分初始化失败和逆序回滚；缺陷修复必须增加能够复现问题的回归测试；性能热路径应保留必要的 Benchmark。

先执行受影响包的测试，再按改动风险执行全仓测试、竞态检测、静态检查和构建。测试以真实风险和代码路径为依据，不为了覆盖率数字编写无有效断言的测试，也不为了测试方便增加不必要的生产抽象。

### 15. 公共算法只在真实复用时提炼

职责独立、语义通用且已有多个真实使用方的基础算法或数据结构，可以放入 `internal` 下的语义明确子包。公共包只实现算法本身，不依赖具体 Service 或业务类型，API 保持最小。

只被一个模块使用、强依赖模块状态或提炼后反而增加理解成本的逻辑，应保留在所属模块中。

### 16. 注释解释意图和约束

使用中文注释说明职责、设计原因、关键执行阶段、状态转换、并发约束、资源所有权、错误回滚、边界条件和不易理解的性能选择。

注释不得机械逐行翻译代码。短小且语义清晰的内部代码不要求逐行注释；公开 API 应具有准确的 GoDoc。行为变化时必须同步维护相关注释，禁止保留失效说明。

### 17. 待确认问题集中讨论

设计或实现中存在多个需要开发者决定的问题时，应按同一层级和依赖关系集中整理。每个问题给出推荐方案、主要理由、复杂度和重要影响；已经确认且无新冲突的问题不重复询问。

### 17.1 DBService 数据域与部署关系

后续设计、配置、实现和部署必须遵守以下已确认关系，不得因调用方便把公共登录协调数据写入区服 RoleDBService，也不得让业务 Service 直接创建 MongoDB 或 Redis Client：

- 仓库只实现一个业务无关的 `DBService` 模板，实际运行时按数据域实例化为 `AccDBService` 和各区服自己的 `RoleDBService`；DBService 仍只提供通用 MongoDB/Redis 执行契约，不增加账号、玩家、登录路由等领域 RPC；
- `AccDBService` 属于公共账号与登录协调数据域。账号数据、登录限流、GameService 实例注册、负载、玩家在线路由和登录分配使用其固定账号数据库与公共 Redis；LoginService、GatewayService、公共 Service 以及各区服 GameService 均通过 `AccDBService` RPC 使用这些能力；
- 公共数据 Node 部署 `AccDBService`，每个区服的 DBServer 也必须部署一个 `AccDBService` 实例；这些实例连接同一套账号数据库和公共 Redis，组成同一公共数据域。NodeID、实际 ServiceName、Node Labels、发现范围和配置归属必须遵循 `docs/design/部署标识与服务发现配置设计.md`，不得通过解析 NodeID 或临时改写 ServiceName 代替明确的 Label 路由；
- 每个区服的 DBServer 还必须部署本区服独立的 `RoleDBService`。GameService 只通过本区服 `RoleDBService` 读写角色 MongoDB 和区服 Redis，以隔离不同区服的 DBService 队列、并发额度、连接池、Redis 连接、过载和发布故障；即使当前全部区服角色文档混合存放在同一个 MongoDB 数据库中，也不得因此合并各区服的 RoleDBService；
- GameService 同时依赖本区服 DBServer 上的 `AccDBService` 实例和 `RoleDBService` 实例：前者仍属于公共账号数据域，只用于登录协调和在线路由；后者只用于角色数据及区服业务。本区服 AccDBService 不可用时暂停该区服新的登录协调，但已经在线的区服业务应尽量继续依赖本区服 RoleDBService 运行；
- Gateway 的 `PlayerRouteStore` 只通过 `AccDBService` 访问公共 Redis，不通过任何区服 RoleDBService。`PlayerRouteStore` 是不持有连接、Timer 和权威内存状态的普通进程内组件，不是 Origin Module 或独立 Service；GameService 注册与租约续期若需要 Timer，由 GameService 自身或其专属生命周期 Module 持有；
- 所有区服隔离字段必须显式携带。当前玩家稳定身份固定为 `PlayerKey = AccountID + ShowAreaID`；`RealAreaID` 只表示当前运行归属和 Redis 分组，不参与玩家稳定身份。

## 工程目录和命名规则

### 18. 目录职责必须清晰

新增文件或目录前，必须先按职责归位，不得只因调用方位于某个 Service 就把共享契约放进该 Service，也不得为减少 import 创建类型中转包。

#### 18.1 仓库标准目录

```text
origingame/
├── cmd/                         # 唯一生产程序入口
├── service/                     # 各业务 Service 的实现
│   └── <xxxservice>/
├── protocol/
│   ├── common/                  # 客户端与服务端共享的 Protobuf 契约
│   └── rpc/                     # Service 间 RPC 契约、参数和生成物
├── internal/
│   ├── mongodb/                 # 仓库级 MongoDB 集合契约
│   └── <明确能力名>/            # 真实跨 Service 复用的内部能力
├── config/                      # 程序运行配置和不参与加载的示例配置
├── scripts/                     # 生成、检查、构建等开发维护脚本
├── bin/                         # 开发者直接执行的启动与调试入口脚本
├── deploy/                      # 本地基础设施和部署资源
├── docs/design/                 # 总体、Service、协议和公共机制设计
└── tests/                       # 跨包集成测试与进程级 E2E 测试，按需创建
```

- OriginGame 只有一个生产程序入口，直接放在 `cmd/main.go`；运行哪个 Node 及其 Service 由配置和启动参数决定；
- 不预先创建没有真实内容的目录、空 Go 包、空协议文件或空生成文件；文档确认的未来目录只在首个真实文件出现时创建；
- `common` 只允许作为已确认的 `protocol/common` 协议边界；其他位置不得新增含义不清的 `common`、`model`、`util`、`misc` 或 `shared` 目录。

#### 18.2 新文件归位顺序

新增文件时按以下顺序判断，命中后不再放入其他层级：

1. 客户端必须共同理解的 Protobuf 枚举或消息，放 `protocol/common`；
2. Service 间 RPC 接口、普通 Go 参数、RPC 专用 Protobuf 参数及其生成物，放 `protocol/rpc`；
3. MongoDB 集合名、BSON 文档及集合内嵌字段，放 `internal/mongodb`；
4. 单个 Service 的实现、Repository、资源生命周期、内部业务对象或网络边界，放 `service/<xxxservice>` 的明确子包；
5. 已有多个真实 Service 使用、且不属于协议或数据库契约的仓库内部能力，放 `internal/<明确能力名>`；
6. 运行配置放 `config`，开发维护脚本放 `scripts`，启动入口脚本放 `bin`，部署资源放 `deploy`，设计结论放 `docs/design`；
7. 普通单元测试和 Benchmark 与源码同包；跨多个包的集成测试放 `tests/integration`，完整进程或集群级测试放 `tests/e2e`。

如一个类型同时可能落入多个目录，应以其稳定契约边界而非当前使用者决定位置。例如客户端消息属于 `protocol/common`，RPC 参数属于 `protocol/rpc`，即使当前只有一个 Service 使用也不得放入该 Service 实现包。

#### 18.3 配置、脚本与部署边界

- `config` 下的配置文件统一平铺，不按 Server 或 Service 创建子目录；按稳定职责命名：全局日志使用 `log.yaml`，Node 拓扑使用 `<node>-node.yaml`，Service 配置使用 `<service>.yaml`，独立服务能力使用 `<service>-<capability>.yaml`。禁止新增笼统的 `application.yaml`、`config.yaml` 或 `common.yaml`；
- Origin 会递归合并 `config` 下全部 `.json`、`.yml`、`.yaml` 文件，同一逻辑配置路径不得在多个可加载文件重复定义。一个 Service 的相关配置默认集中在其 `<service>.yaml`，只有独立分发、密钥轮换或权限边界确有需要时才拆为 `<service>-<capability>.yaml`；
- 仓库可提交仅含本地开发测试数据、可直接启动的 `.yaml`；生产实际配置由部署目录、Secret 或环境变量注入，参考模板使用不参与加载的 `.yaml.example` 后缀。公钥等非敏感、已确认会被当前或后续 Service 读取的配置可直接提交为 `.yaml`；
- 可加载 YAML 的叶子字段应在行末写简短中文注释，说明用途、单位、默认策略或安全边界；只列出改变框架默认值、决定外部连接/监听、容量/超时、安全和业务行为的字段。不得为了展示而复制低价值默认字段；
- `scripts` 只保存生成、检查、构建和维护脚本，不保存业务运行逻辑或生成工具二进制；Windows 与 Linux/macOS 对应脚本必须行为一致并自行定位仓库根目录；
- `bin` 只保存开发者直接执行的启动、停止或调试入口脚本，不提交编译产生的可执行文件；
- `deploy` 保存 Compose、容器、基础设施和实际采用的部署资源，不保存业务源码、协议源或仅供生产数据迁移使用的临时脚本；
- 构建产物、日志、PID、覆盖率和临时生成目录不得混入源码目录，并必须由 `.gitignore` 排除。

#### 18.4 编码前确认职责归属

新增或修改代码前，必须先确认职责应归属哪个 Service、子包、Module 和文件。归属判断以稳定职责、资源所有权、生命周期、依赖方向和主要变化原因为依据，不以当前调用位置、参数传递方便或减少 import 为依据。

编码前至少检查以下问题：

1. 代码承担的稳定职责是什么，哪些需求变化会导致它修改；
2. 它持有或使用哪些状态和资源，这些资源由谁创建、启动、停止和释放；
3. 它应与哪个 Service 或 Module 共享生命周期；
4. 放入目标包后，依赖是否仍从装配层指向具体能力，子包是否会反向依赖 Service 根包；
5. 新增同类功能时，是否只需修改职责所属的包或 Module；
6. 目标文件是否只有一个可以准确命名的主要职责。

具体边界遵循以下规则：

- Service 根包负责配置聚合、生命周期装配和真正跨多个子包的协调流程，不作为 HTTP、RPC、网络消息、Timer 等全部入口回调的集中存放处；
- 已经具有具体业务语义的 Module 应拥有自身入口注册、对应 Handler 或回调、中间件以及入口协议转换。例如业务 HTTP Module 自己注册路由并实现对应 Handler，不通过构造函数从 Service 根包逐个接收路由 Handler；
- HTTP 边界包负责路由、HTTP Middleware、JSON DTO、参数绑定、HTTP 状态码与业务错误码映射。单个接口内容较多时拆为 `<业务>handler.go` 和 `<业务>dto.go`，不把所有 Handler 堆入 `httpmodule.go`；
- TCP、KCP、WebSocket、RPC、消息订阅和 Timer Module 同样遵循“入口注册与入口回调属于同一业务边界”。入口回调需要业务能力时，构造 Module 时注入 Repository、鉴权器、签发器等最小能力，而不是注入属于该 Module 自身的回调函数；
- 通用框架 Module 为保持业务无关性可以接收 Handler 或回调；一旦 Module 已包含具体业务路由、DTO 或流程，就不得再以“通用”为由把对应 Handler 外置；
- 只被一个入口使用的协调逻辑默认留在该入口所属 Module。只有出现多个真实调用方，并且语义确实一致时，才提炼独立业务能力或公共接口；
- 禁止在 Service 根包增加只用于转发到子包、或只为满足构造参数的无业务价值方法；
- 如果新增一个同类入口必须同步修改多个没有独立变化原因的层级，应暂停编码并重新检查职责边界，不得继续堆叠构造参数、转发方法或注册器。

### 19. Service 按必要职责拆包

`service/<xxxservice>` 根目录默认只保留 `xxxservice.go`。该文件集中保存 Service 主类型、顶层配置聚合、生命周期装配以及必须由 Service 调度器串联多个子包的核心协调流程。其他职责必须进入语义明确的子包。

不为了形式统一提前创建空子包。是否拆包以职责边界、依赖方向、独立测试能力和维护成本为依据。子包不得反向依赖 Service 根包；Service 根包负责装配各子包。

#### 19.1 根包允许保存的内容

- `xxxservice.go`：Service 主类型、`OnInit`、`OnStart`、`OnStop`、顶层配置聚合、依赖装配和跨子包协调流程；
- 只有当 `xxxservice.go` 已经因多个独立协调流程过大，并且拆出的文件仍然只包含 `XxxService` 方法时，才可以经开发者确认后在根目录增加 `xxxservice_<明确职责>.go`；
- `xxxconfig.go`、`xxxhandler.go`、`xxxmodule.go` 默认不得单独留在根目录；配置归属 Service 时合入 `xxxservice.go`，DTO 和网络边界进入协议子包，资源 Module 进入其所属能力子包。

根包中的代码可以依赖子包；子包禁止导入当前 Service 根包，否则会形成反向依赖或循环依赖。根包不是公共类型中转站，不得为了少写 import 而重新导出所有子包类型。

#### 19.2 子包的划分顺序

优先按稳定的业务领域或独立能力划分，而不是单纯按文件类型横向堆放：

1. 账号、玩家、区服等业务领域分别使用 `account`、`player`、`area` 等子包；该领域的业务对象、校验和 Repository 放在同一领域包内，MongoDB 文档结构遵循 19.3 的集中定义规则；
2. HTTP、TCP、KCP、WebSocket 等外部传输契约使用 `httpapi`、`protocol` 等边界包，DTO 不放入 Service 根包或领域持久化包；
3. 限流、鉴权等职责独立并可单独测试的能力使用 `ratelimit`、`authentication` 等子包，其专属配置与实现放在同一包内；
4. 只有真正跨多个业务领域共享的持久化基础能力，才建立独立的 `repository` 包；不得把所有领域的 Repository 机械集中到一个目录；
5. `model`、`common`、`util`、`misc` 等无法表达边界的目录名禁止使用。

出现以下任一情况时应考虑拆包：一个职责已经包含多份实现或测试文件；可以定义不依赖 Service 根包的最小 API；拥有独立外部依赖或数据模型；需要被根包之外的同仓库代码复用。只有一个短文件且强依赖 Service 私有状态时，应继续留在根包，不为满足目录形式强行拆分。

#### 19.3 MongoDB 集合契约集中定义

MongoDB 持久化结构是多个 Service 共同遵守的数据契约，统一直接放在仓库级 `internal/mongodb`，不再增加 `collection` 子目录：

```text
internal/mongodb/
├── account.go               # Account 集合名和 BSON 文档结构
├── realareainfo.go          # RealAreaInfo 集合名和 BSON 文档结构
└── showareainfo.go          # ShowAreaInfo 集合名和 BSON 文档结构
```

- 包名固定使用 `mongodb`，不使用语义不准确的 `collect`，也不增加只有一层内容的 `collection` 子包；
- 每个集合使用一个以集合业务名命名的文件，不新增宽泛的 `document.go` 或 `collection.go`；
- 包内只保存集合名称常量、BSON 文档结构及其必需的内嵌字段类型，不持有 MongoDB Client、Repository、查询更新逻辑、Service 生命周期或业务流程；
- 结构体名称直接对应 MongoDB 集合，例如 `mongodb.Account`、`mongodb.RealAreaInfo`，集合名使用 `mongodb.AccountName`；
- BSON 文档与 HTTP、RPC 等传输 DTO 分离；Service 负责在持久化结构与自身业务对象或传输对象之间转换；
- 仓库根目录下的 `internal` 对本仓库所有 Service 可见，但禁止仓库外部模块导入，符合当前数据库契约只供 OriginGame 使用的边界；
- 新增或修改集合字段、BSON 名称、时间语义、主键或索引属于数据库结构变更，必须先更新对应 Service 主设计文档并确认。

#### 19.4 LoginService 标准示例

当前 LoginService 的目录结构是后续 Service 拆包的参考样例：

```text
service/loginservice/
├── loginservice.go          # 主类型、配置、生命周期、装配和登录协调流程
├── account/                 # 账号身份和账号 Repository
├── area/                    # 区服视图、快照、Repository 和刷新 Module
├── authentication/          # SDK 鉴权契约及实现
├── database/                # MongoDB 生命周期和启动数据准备
├── httpapi/                 # HTTP Module 和 JSON DTO
├── ratelimit/               # 限流配置、实现和测试
└── redis/                   # Redis 生命周期和降级策略
```

此示例表达的是依赖方向和职责边界，不要求所有 Service 创建完全相同的子目录。新 Service 只能创建自身实际需要的目录。

#### 19.5 客户端 Protobuf 协议

与客户端共同遵守的 Protobuf 源文件和服务端 Go 生成物统一平铺在 `protocol/common`：

```text
protocol/common/
├── errorcode.proto           # 客户端可见错误码
├── errorcode.pb.go           # protoc-gen-go 生成，不得手改
├── messageid.proto           # 有真实消息注册需求时创建
└── messageid.pb.go
```

- `.proto` 是协议的唯一源文件，生成的 `.pb.go` 与源文件同目录并提交仓库；
- Protobuf `package` 统一使用 `origingame.protocol`，`go_package` 统一指向 `origingame/protocol/common;commonpb`；
- Go 包名固定为 `commonpb`，调用方导入路径虽然包含 `common`，但代码中不使用含义模糊的 `common.Xxx`；
- 文件按具体协议职责命名，不新增宽泛的 `common.proto`、`message.proto` 或 `protocol.proto`；
- 已发布的字段号、枚举值和含义不得静默修改或复用；删除内容必须同时 `reserved` 原编号和名称；
- MongoDB 文档、HTTP 专用 DTO、Service 内部状态和内部 RPC 参数不得放入客户端共享协议；
- 当前不按客户端版本和业务域继续建立子目录；出现真实的不兼容协议版本或目录规模问题时再重新设计；
- 新协议文件只在出现真实消息时创建，不预建空文件或空目录。

#### 19.6 Origin RPC 契约

Service 间 RPC 使用 Origin v3 的 Go 接口与 `origingen`，统一平铺在 `protocol/rpc`，不沿用老版本 `rpcproto/*.proto`：

```text
protocol/rpc/
├── gameservice.go                 # 手写接口和普通 Go 参数
├── gameservice.proto              # 手写 RPC 专用 Protobuf 参数，按需创建
├── gameservice.pb.go              # protoc-gen-go 生成，不得手改
├── gameservice.rpc.gen.go         # origingen 生成，不得手改
├── gatewayservice.go
├── gatewayservice.proto
├── gatewayservice.pb.go
└── gatewayservice.rpc.gen.go
```

- 包名固定为 `rpcapi`，避免与 Origin 框架 `rpc` 包混淆；
- `protocol/rpc` 不再按参数编码方式建立子目录；Go 契约、RPC Protobuf 源及两类生成物全部平铺，并用完整 Service 名前缀聚合；
- 每个对外提供 RPC 的 Service 只使用一个 `<service>.go` 手写文件，该 Service 的接口、普通 Go 请求、返回和专用枚举集中在此文件；
- 每个手写契约文件由 `origingen` 生成同名 `.rpc.gen.go`；生成物提交仓库且禁止手工修改；
- RPC 接口使用与目标 Service 模板一致的名称，并通过 `//origin:rpc` 标记；业务实现使用编译期断言确认实现该接口；
- RPC 契约方法按具体业务动作命名，不添加 `Rpc` 或 `RPC` 前缀；调用语义由 `origingen` 生成的 `Await`、`Call`、`Async`、`Notify` 和 `Broadcast` 前缀表达，避免形成 `AwaitRpcXxx`、`NotifyRpcXxx` 等重复名称；
- 普通 Go RPC 参数与接口放在同一个 `protocol/rpc/<service>.go`，由 `origingen` 生成静态 Codec；
- 需要 Protobuf 线格式时，在同层创建 `protocol/rpc/<service>.proto`，生成同层 `<service>.pb.go`；Protobuf `package` 固定为 `origingame.rpc`，`go_package` 固定为 `origingame/protocol/rpc;rpcapi`；
- 同一 Service 没有 Protobuf 参数时不得创建对应 `.proto`；普通 Go 类型与 Protobuf 消息不得重名；
- Protobuf 类型必须直接作为 RPC 方法的顶层参数或顶层业务返回值，推荐使用生成的指针类型；嵌入普通 Go 参数结构体后不按官方 Protobuf Codec 处理；
- RPC 参数不默认复用客户端 Protobuf、MongoDB Collection、HTTP DTO 或 Service 内部结构；`protocol/rpc/*.proto` 是服务端 RPC 专用协议，不得被客户端作为共享契约使用；
- `protocol/rpc` 禁止依赖任何 `service/<xxxservice>` 实现包，Service 实现包可以依赖 `protocol/rpc`；
- 修改 RPC 接口或参数后必须重新生成，并通过 `origingen rpc --check`；统一脚本直接扫描整个 `protocol/rpc` 包，不在每个 Service 文件中重复声明相同的 `go:generate`；
- `protocol/rpc` 当前允许保存 `README.md`、`*.go.example` 和 `*.proto.example` 目录样例；样例不参与编译和生成，不得未经设计确认直接改名为真实契约；
- 除上述样例外，不得只为未来 Service 预建空的 `.go` 契约或生成文件。

#### 19.7 协议生成入口

- Windows 使用 `scripts/generate.bat`，Linux/macOS 使用 `scripts/generate.sh`，两者必须执行等价的 Protobuf 和 Origin RPC 生成流程；
- `protoc` 固定为 24.0，`protoc-gen-go` 固定为 1.31.0；脚本必须先校验版本，不得使用版本不明的全局工具生成；
- Protobuf 生成同时覆盖客户端 `protocol/common/*.proto` 和 RPC 参数 `protocol/rpc/*.proto`；只有存在真实 `.proto` 源文件时才处理对应目录，`*.proto.example` 不参与生成；
- `origingen` 通过 `go run github.com/duanhf2012/origin/v3/cmd/origingen` 执行，版本由当前项目 `go.mod` 固定，不依赖全局安装；
- 提交前使用对应的 `scripts/check-generated.*` 检查 Protobuf 和 RPC 生成物是否缺失、过期或多余；检查流程不得修改工作区；
- 升级任一生成工具必须单独说明兼容性影响，并提交相同输入产生的完整生成差异。

### 20. 文件名必须表达具体职责

禁止新增 `service.go`、`model.go`、`store.go`、`util.go`、`common.go` 等影响全局检索的宽泛文件名。

文件名应直接表达主要类型或职责，例如：

- `LoginService` 放在 `loginservice.go`；
- `GatewayService` 放在 `gatewayservice.go`；
- MongoDB 账号仓储使用 `mongorepository.go`；
- MongoDB 集合契约按集合名使用 `account.go`、`realareainfo.go` 等文件，并集中放在 `internal/mongodb`；
- Service RPC 契约使用 `<service>.go`，RPC Protobuf 参数使用同层 `<service>.proto`，生成物分别为 `<service>.rpc.gen.go` 和 `<service>.pb.go`；
- 区服目录使用 `areacatalog.go`；
- HTTP 登录接口使用 `loginhandler.go`。

一个文件包含多个职责、必须依赖宽泛名称才能概括时，应优先拆分职责，而不是继续扩大文件。

## 时间使用规则

### 21. 严格区分游戏业务时间和真实系统时间

Origin v3 的 Node 提供可被 GM 调整的游戏逻辑时间，并会在时间改变后重排该 Node 中已经登记的业务 Timer。业务代码必须根据时间语义选择正确的时间源，禁止混用。

#### 21.1 游戏业务时间

以下逻辑必须使用当前 Node 的游戏逻辑时间，即通过 `Service.GetNode().Now()`、`Module.GetNode().Now()` 获取，或者由调用方显式传入该时间：

- 活动开始、结束和周期刷新；
- 每日、每周任务及跨天判断；
- 体力、建筑、生产等业务恢复或完成时间；
- 游戏邮件、排行榜赛季和业务奖励有效期；
- 其他需要被 GM 调时、测试时间推进或游戏世界时间影响的逻辑。

这些业务逻辑不得直接调用 `time.Now()`。

游戏业务定时任务必须使用 Origin 提供的 Service Timer、Ticker 和 Cron。禁止使用 `time.After`、`time.AfterFunc`、`time.NewTimer`、`time.NewTicker` 或自行创建基于真实时间等待的 goroutine，否则 GM 调时不能正确重排已经登记的任务。

#### 21.2 真实系统时间

以下基础设施、安全和事实记录必须使用真实系统时间，不受 GM 调时影响：

- TCP、KCP、WebSocket 的读写和握手超时；
- HTTP、RPC、数据库请求及 Context Deadline；
- Redis 锁、租约、基础设施 TTL、服务发现和健康检查；
- 重试、退避、熔断和限流窗口；
- Token 的签发时间、生效时间和过期时间；
- 登录记录、支付流水、审计记录和运维操作记录；
- 日志时间、监控采集和性能统计。

使用真实时间的代码可以调用 `time.Now()` 或相应基础设施时钟，但不得据此实现需要响应 GM 调时的游戏业务。

#### 21.3 时间数据必须标明语义

设计数据库字段、协议字段和内部类型时，必须明确其使用业务时间还是真实时间。存在混淆风险时，名称应体现语义，例如 `business_expire_at`、`recorded_at`，或在字段注释中明确时间源、时区和精度。

跨 Service 传递绝对业务时间时，应传递明确时区和精度的时间戳，不依赖接收方当前偏移重新推导。

#### 21.4 GM 调时的作用域和一致性

游戏逻辑时间属于 Node；同一 Node 内的 Service 共享时间偏移，不同 Node 之间默认相互独立。

全服或多 GameService 调时应向全部目标 Node 发布同一个绝对目标时间并调用 `SetTime`，不应让各节点分别执行相对 `AddTime`，避免因消息到达时间不同产生额外偏差。GM 调时入口必须鉴权并记录真实时间审计日志。

临时开发调时可以不持久化。正式活动若要求重启后保持统一时间，应由统一、可持久化的活动时间配置或专门时间控制机制负责，不能依赖进程内临时偏移。

#### 21.5 测试要求

涉及业务时间的功能必须包含时间前进、时间回退、跨日或跨周期以及 Timer 重排的测试。测试应通过 Node 游戏时间能力推进时间，不使用长时间 `Sleep` 模拟业务时间经过。
