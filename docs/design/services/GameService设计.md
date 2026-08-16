# GameService 设计

> 状态：Player组织、Proxy生命周期、登录、逻辑心跳、断线驻留、存档、消息执行与注册方案已确认
> 更新日期：2026-08-15
> 上位文档：[OriginGame v3 总体架构设计](../总体架构设计.md)

## 1. 服务定位

| 层级 | 名称 |
| --- | --- |
| 生产程序 | `OriginGame`（统一程序，按Node配置启动） |
| Origin Service | `GameService` |
| Go包 | `gameservice` |

GameService负责真实区服中的在线玩家对象、玩家数据加载、游戏业务消息和玩家生命周期。GameService所在Node及其DBService发现范围遵循[部署标识与服务发现配置设计](../部署标识与服务发现配置设计.md)：使用`game-area-<real-area-id>-<ordinal>`，并只调用本区服DB Node上的AccDBService和RoleDBService。

每个GameService必须配置`player_capacity`，表示当前实例可以同时拥有的`ASSIGNING + LOADING + ONLINE + RESIDENT`玩家上限；本地样例值为`5000`，缺失或非正数时启动失败。该值注册到公共Redis的`MaxPlayers`，达到容量后实例不再参与新玩家分配。

## 2. Player组织方式

一个GameService只创建一个`PlayerModule`。PlayerModule是Origin Module，统一拥有当前GameService内全部Player：

```text
GameService
└── PlayerModule
    ├── Player(PlayerKey-1)
    │   ├── UserInfoProxy
    │   ├── BagProxy
    │   └── QuestProxy
    ├── Player(PlayerKey-2)
    └── ...
```

职责固定如下：

| 对象 | 职责 |
| --- | --- |
| `GameService` | Service生命周期装配、RPC入口和跨Module协调 |
| `PlayerModule` | 创建、查找、加载、存档、上线、离线和释放Player，并维护ConnectionID到Player的运行期索引 |
| `Player` | 保存`PlayerKey`、会话状态以及全部玩家功能Proxy |
| `XxxProxy` | 单个玩家某一功能的数据逻辑，例如账号信息、背包和任务 |

`Player`和`XxxProxy`都不是Origin Module，不单独拥有Service生命周期。PlayerModule是Player及全部Proxy的唯一生命周期所有者。

## 3. Proxy注册与顺序

`NewPlayer`时一次性创建并注册当前版本全部持久化数据和功能Proxy。Player只通过`proxies`保存生命周期注册顺序；各具体Proxy由Player直接保存类型明确的字段并通过类型明确的访问方法取得，不增加通用ID查找表。

规则：

- 全部Proxy注册完成后冻结，Player运行期间禁止新增、删除或调整顺序；
- 如果Proxy B依赖Proxy A，必须先注册A再注册B；
- 初始化、加载、关联和上线按注册顺序执行；
- 离线和释放按注册倒序执行；
- 初始化中途失败时停止后续步骤，只对已经初始化成功的Proxy倒序调用`OnRelease`；
- 提供`BasePlayerProxy`作为全部回调的默认空实现，具体Proxy嵌入它并只覆盖需要处理的阶段。

样板工程首期只注册`UserInfoProxy`，用于演示基础数据加载、默认初始化、修改标脏和存档。背包、任务等Proxy在出现真实样板需求后再增加。

注册顺序是显式生命周期依赖，不通过包级可变注册表、反射或`init()`隐式建立。`NewPlayer`直接按固定顺序调用私有`registerPersistentData`和`registerProxy`，不建立全局工厂或单独的动态注册阶段：

```go
player.registerPersistentData(mongodb.UserInfoName, &player.userInfo)
player.registerProxy(&player.userInfoProxy)
```

`registerPersistentData`登记集合名、数据指针、是否加载到数据和脏代数；重复登记同一持久化文档或传入空指针时立即创建失败。`registerProxy`只追加具体结构体对象到生命周期顺序切片。当前没有按ID动态查找Proxy的真实需求，因此删除`ProxyID`、`ProxyID()`、`proxyMap`和`PlayerProxyCount`；以后出现真实通用查找场景时再设计。

## 4. 数据与逻辑分离

Player直接持有所有Proxy经常访问的公共临时数据`DataInfo`和公共基础持久化数据`CUserInfo`；`UserInfoProxy`拥有`CUserInfo`的业务修改逻辑：

```go
type Player struct {
	key PlayerKey

	dataInfo DataInfo  // 非持久化临时数据
	userInfo CUserInfo // MongoDB持久化基础数据

	userInfoProxy UserInfoProxy

	proxies       []PlayerProxy        // 生命周期回调的固定注册顺序
	persistentData []persistentDataEntry // 自动加载、标脏和存档的数据登记
}
```

### 4.1 DataInfo

`DataInfo`只保存当前Player对象生命周期内使用的临时状态，不参与BSON序列化：

```go
type DataInfo struct {
	GatewayNodeID       string    // 当前连接所在Gateway Node
	GatewayConnectionID string    // 当前全局唯一Gateway连接标识
	State            PlayerState // Loading、Online、Resident或Releasing
	LastHeartbeatAt  time.Time   // 最近一次玩家逻辑心跳的真实系统时间
	ResidentDeadline time.Time   // 断线驻留截止的真实系统时间
}
```

所有Proxy可以读取`DataInfo`。连接和在线状态等影响生命周期的字段只由Player或PlayerModule修改；具体Proxy不得绕过Player生命周期直接改写。

`GatewayConnectionID`由Gateway网络层生成，在全部Gateway节点、TCP/KCP/WebSocket Module和进程重启之间保持全局唯一。首期不再增加`GatewayInstanceID`或玩家级`SessionFence`；GameService进程重启由Origin的`NodeSessionID`区分。

### 4.2 CUserInfo

`CUserInfo`保存所有Proxy经常读取的玩家基础持久化数据。样板工程只保留能够演示首次创建、数据库加载、修改和登录返回的最小字段：

```go
// CUserInfo是MongoDB UserInfo集合中的玩家基础数据。
type CUserInfo struct {
	PlayerKey  string `bson:"_id"`          // AccountID和ShowAreaID生成的稳定主键
	AccountID  string `bson:"account_id"`   // 账号ID
	ShowAreaID int64  `bson:"show_area_id"` // 显示区服ID

	Nickname string `bson:"nickname"` // 玩家昵称
	Level    int32  `bson:"level"`    // 玩家等级

	CreatedAt    time.Time `bson:"created_at"`     // 角色创建真实系统时间
	LastLoginAt  time.Time `bson:"last_login_at"`  // 最近登录真实系统时间
	LastLogoutAt time.Time `bson:"last_logout_at"` // 最近离线真实系统时间
}
```

集合名固定为`UserInfo`，`_id`使用无歧义编码的`PlayerKey`。首期不增加经验、战力、头像等非必要样板字段。首次创建默认值为：

```text
Nickname = ""
Level = 1
CreatedAt = 当前真实系统时间
LastLoginAt = 当前真实系统时间
LastLogoutAt = 零值
```

`CUserInfo`只保存数据，不实现业务逻辑；集合契约最终放入仓库级`internal/mongodb/userinfo.go`。

`CUserInfo`的数据所有权固定为：

- Player直接持有`CUserInfo`，使全部Proxy可以低成本读取；
- `UserInfoProxy`负责等级、昵称、基础属性等业务修改；
- 其他Proxy需要修改`CUserInfo`时必须调用`UserInfoProxy`公开方法；
- `UserInfoProxy`每次有效修改后负责标记`CUserInfo`为脏数据；
- 禁止其他Proxy直接修改`CUserInfo`字段后再依赖人工补充脏标记。

### 4.3 BasePlayerProxy访问能力

每个具体Proxy嵌入`BasePlayerProxy`，统一访问所属Player、临时数据和基础持久化数据：

```go
type BasePlayerProxy struct {
	player *Player
}

// Player返回当前Proxy所属的Player对象。
func (proxy *BasePlayerProxy) Player() *Player

// DataInfo返回非持久化临时数据。
func (proxy *BasePlayerProxy) DataInfo() *DataInfo

// UserInfo返回玩家基础持久化数据，供其他Proxy读取。
func (proxy *BasePlayerProxy) UserInfo() *CUserInfo

// UserInfoProxy返回基础信息代理；修改CUserInfo必须通过该代理完成。
func (proxy *BasePlayerProxy) UserInfoProxy() *UserInfoProxy
```

`BasePlayerProxy.OnInit`保存Player引用。具体Proxy如果覆盖`OnInit`，必须先完成Base初始化，再执行自己的初始化。

### 4.4 其他功能Proxy

其他功能继续使用“持久化数据 + 临时数据 + Proxy逻辑”分离方式：

```go
type BagProxy struct {
	BasePlayerProxy

	data    CBagData    // 背包持久化数据
	runtime BagDataInfo // 背包临时索引和缓存
}
```

- `CXxxData`只保存MongoDB持久化数据；`XxxDataInfo`只保存运行期索引、缓存和临时状态；
- Proxy负责围绕两类数据实现业务规则，不把业务方法放进纯数据结构；
- PlayerModule统一完成已注册持久化数据的MongoDB加载、反序列化、序列化和存档；
- DBService仍只执行通用MongoDB/Redis操作，不理解Player或Proxy；
- Proxy不实现`BuildPlayerLoad`、`BuildPlayerSave`等数据库构造生命周期回调；
- 某个Proxy查询不到已有数据时，在`OnLoaded`中完成自己的默认初始化；这不等同于整个Player一定是新角色。

Proxy与MongoDB集合不要求一一对应。是否拆分集合由数据大小、访问方式、更新频率和原子性决定，不因新增Proxy机械新增集合。

## 5. PlayerProxy生命周期接口

首期只保留覆盖Player完整加载和在线生命周期的必要回调：

```go
// PlayerProxy 是注册到单个Player上的功能代理。
type PlayerProxy interface {
	// OnInit 在NewPlayer注册全部Proxy后按注册顺序调用。
	// 此时数据库数据尚未加载，只初始化运行状态和绑定Player。
	OnInit(player *Player) error

	// OnLoaded 在全部已登记持久化数据完成反序列化后调用。
	// 负责本Proxy使用数据的初始化、兼容、校验和索引构建。
	OnLoaded(ctx PlayerLoadContext) error

	// OnAllLoaded 在全部Proxy完成OnLoaded后按注册顺序调用。
	// 此时可以访问其他Proxy，建立跨Proxy关联和派生数据。
	OnAllLoaded(ctx PlayerLoadContext) error

	// OnOnline 在玩家首次上线或断线重连时按注册顺序调用。
	OnOnline(ctx PlayerOnlineContext)

	// OnOffline 在玩家断线、顶号或服务器停止时按注册倒序调用。
	// Offline不代表Player立即释放，Player可以继续驻留一段时间。
	OnOffline(ctx PlayerOfflineContext)

	// OnRelease 在Player最终存档完成后按注册倒序调用。
	// 只释放运行期状态，禁止继续修改持久化数据。
	OnRelease()
}
```

`PlayerLoadContext`至少提供：

```go
type PlayerLoadContext struct {
	IsNewPlayer bool // 整个Player是否为首次创建角色
}
```

首期只有`UserInfo`持久化文档，`IsNewPlayer`由它是否存在决定。Proxy直接通过Player访问对应数据，不在公共Context中增加无法明确对应某个持久化文档的`DataFound`。以后出现一个Proxy拥有独立持久化文档并确实需要区分“文档不存在”和“字段为零值”时，再为该登记项增加明确的查询方式。

## 6. 生命周期执行顺序

完整顺序固定为：

```text
NewPlayer
    -> 按代码中确定的固定顺序注册全部Proxy及持久化Data
    -> 按注册顺序执行OnInit
    -> PlayerModule通过RoleDBService自动加载并反序列化数据
    -> 按注册顺序执行OnLoaded
    -> 全部OnLoaded成功形成第一道屏障
    -> 按注册顺序执行OnAllLoaded
    -> 全部OnAllLoaded成功后自动插入本次初始化的缺失持久化数据
    -> 初始数据保存成功后Player进入Ready
    -> 按注册顺序执行OnOnline

普通断线
    -> 按注册倒序执行OnOffline
    -> 立即保存已标脏的持久化数据
    -> Player进入Resident并驻留15分钟

顶号
    -> 按注册倒序执行旧连接OnOffline
    -> 替换连接后按注册顺序执行OnOnline
    -> 不进入15分钟驻留

服务停止
    -> 按注册倒序执行OnOffline
    -> 不再等待驻留期，直接进入最终释放流程

最终释放
    -> PlayerModule停止该Player的新业务处理
    -> 15分钟内未重连时执行最终存档
    -> 停止并释放该Player拥有的定时器
    -> 按注册倒序执行OnRelease
    -> 从PlayerModule移除Player
```

PlayerModule将全部已登记持久化数据按登记顺序组合为一次RoleDBService Mongo请求，每项执行`FindOne({_id: PlayerKey})`；返回的`Results`与Operation下标一一对应，PlayerModule自动完成BSON反序列化，不要求Proxy构造数据库操作。样板工程只有`UserInfo`，通常远小于1KiB。

真实项目必须评估登录常驻数据总量：单次DBService响应受Origin RPC当前`4MiB`业务Payload硬上限约束，登录常驻数据的设计目标为不超过`1MiB`。超过该目标会增加DBService执行槽占用、BSON/RPC编解码、瞬时内存复制、网络往返和登录尾延迟。邮件历史、战斗记录、操作日志等非登录必需大数据不得登记为Player常驻数据，应在业务需要时分页加载；确实必须常驻且单批接近上限时，由业务按明确数据组拆成多个有界请求，不增加按数据量猜测的自动拆包机制。监控至少记录总BSON字节数、Operation数量和加载耗时。

`IsNewPlayer`只由`UserInfo`文档是否存在决定，其他Proxy缺少自身数据不改变该值。每个Proxy仍在`OnLoaded`中初始化自己的缺失数据；全部`OnAllLoaded`完成后，PlayerModule在`OnOnline`前自动插入所有本次新建的持久化文档。初始保存失败则本次登录失败，不能让尚未可靠落库的新Player进入Online。

数据库执行、反序列化、`OnLoaded`、`OnAllLoaded`或初始保存任一步失败时，立即停止后续阶段，按注册倒序执行已初始化Proxy的`OnRelease`，并按实例身份和连接ID条件释放Redis预占。Player从未进入Online，因此不调用`OnOffline`。首期不为多个MongoDB集合增加分布式事务或补偿回滚；部分已经成功插入的数据保留，下次登录按已有数据加载并继续初始化其他缺失数据。

生命周期次数：

| 回调 | 单个Player对象的调用次数 |
| --- | --- |
| `OnInit` | 一次 |
| `OnLoaded` | 一次 |
| `OnAllLoaded` | 一次 |
| `OnOnline` | 每次首次上线或重连 |
| `OnOffline` | 每次对应连接离线 |
| `OnRelease` | 一次 |

任何`OnInit`、`OnLoaded`或`OnAllLoaded`返回错误，Player都不能进入Ready。PlayerModule负责停止流程、倒序释放已经初始化的Proxy，并按登录路由设计条件释放本次加载占用。

### 6.1 登录RPC与首次加载

Gateway通过Redis取得精确GameService后调用：

```go
// LoginPlayerRequest用于进入GameService并绑定当前Gateway连接。
type LoginPlayerRequest struct {
	AccountID  string // 账号ID
	ShowAreaID int64  // 显示区服ID，与AccountID组成PlayerKey

	ExpectedGameServiceNodeSessionID string // Redis分配时记录的目标Node启动实例
	GatewayNodeID       string // GameService定向调用Gateway所需的NodeID
	GatewayConnectionID string // 全局唯一连接ID，同时作为本次登录标识
}
```

每次首次登录、重连或顶号登录，Gateway都必须在`LoginPlayerRequest`中显式携带`GatewayNodeID + GatewayConnectionID`。GameService以该请求为新连接的权威来源；登录处理不从任何`Session`对象读取这两个字段，也不能用Player当前保存的旧连接字段代替。绑定完成后，PlayerModule同时登记`GatewayConnectionID -> *Player`索引并初始化最后心跳时间。

请求不携带`RealAreaID`或玩家级路由版本。Gateway把Redis分配结果中的`NodeSessionID`写入`ExpectedGameServiceNodeSessionID`；目标GameService必须与自身当前启动实例精确比较，不匹配时拒绝旧分配请求，避免同一NodeID重启后接收上一进程的迟到登录。

首次加载顺序固定为：

```text
校验PlayerKey、GameService NodeID、NodeSessionID和GatewayConnectionID
    -> Redis路由ASSIGNING改为LOADING
    -> NewPlayer及OnInit
    -> 通过RoleDBService加载并反序列化数据
    -> OnLoaded
    -> OnAllLoaded
    -> 自动插入本次初始化的缺失持久化数据
    -> 初始保存成功后创建Player存档Timer
    -> 绑定Gateway连接并执行OnOnline
    -> Redis路由改为ONLINE
    -> 返回登录成功
```

加载失败时，只有仍匹配本次`GameService NodeID + NodeSessionID + GatewayConnectionID`的`ASSIGNING/LOADING`记录可以被释放并扣减负载；旧请求不得清理后续登录产生的新预占。

### 6.2 重复登录与顶号

同一PlayerKey的登录使用Player生命周期状态保护，不增加通用Player Key队列：

- 请求的`GatewayConnectionID`与当前绑定一致时，视为重复请求并幂等返回，不重复加载和执行`OnOnline`；
- Player已经在线且收到不同`GatewayConnectionID`时，不重新选择GameService，也不重新加载玩家数据；
- GameService先从Player当前`DataInfo`快照旧`GatewayNodeID + GatewayConnectionID`，向旧连接定向提交“账号已在其他连接登录”的错误通知和断开请求；再按倒序执行旧连接的`OnOffline`，删除旧ConnectionID索引，以本次`LoginPlayerRequest`的两个字段替换当前连接，登记新索引并按注册顺序执行`OnOnline`；
- 旧Gateway不可达或旧连接已经不存在时忽略该通知结果，不能阻止新连接接管；新登录只有在本地连接替换完成后才返回成功；
- Gateway转发的每条玩家消息都必须附带其全局`GatewayConnectionID`，GameService只处理与Player当前绑定值一致的消息，因此旧连接的迟到消息会被拒绝。

顶号通知使用生成的`NotifySendClientMessage`有界提交，不创建无法停止的独立goroutine，也不等待旧Gateway的远端处理结果。顶号下行使用专用Protobuf Body携带原因，并要求Gateway在写完后关闭旧连接。

### 6.3 登录成功数据

`LoginPlayer`成功时直接返回客户端与服务端共享的Protobuf：

```proto
// LoginPlayerResult是玩家进入GameService成功后的客户端登录数据。
message LoginPlayerResult {
  RoleInfo role_info = 1; // 玩家常用公共基础数据

  // 其他首屏确实需要的数据按业务增加独立结构。
  // BagInfo bag_info = 2;
}

// RoleInfo保存客户端经常使用的角色公共数据。
message RoleInfo {
  string account_id = 1;   // 账号ID
  int64 show_area_id = 2;  // 显示区服ID
  string nickname = 3;     // 玩家昵称
  int32 level = 4;         // 玩家等级
  int64 created_at_ms = 5; // 角色创建Unix毫秒时间
}
```

RPC契约使用顶层Protobuf返回值：

```go
// LoginPlayer完成玩家加载、连接绑定和上线回调，并返回客户端登录数据。
LoginPlayer(context.Context, LoginPlayerRequest) (*commonpb.LoginPlayerResult, error)
```

这是对“RPC参数默认不复用客户端Protobuf”规则的明确例外：该结果本身就是客户端登录成功Body，Gateway不解析或转换业务字段，只负责转发。`role_info`固定承载公共角色数据；其他当前登录确实需要的数据可以增加独立消息字段，但不要求每个Proxy都向`LoginPlayerResult`写入数据。不属于首屏的数据在后续业务消息中按需同步。

## 7. 自动存档与断线驻留

持久化数据的修改方法必须在变更生效时立即标脏，不在定时存档时扫描比较全部数据。每个Player在初始数据保存成功后创建并持有自己的自动存档Timer，Timer回调只序列化和保存该Player已标脏的数据；Timer随Player最终释放而停止，不创建每Player goroutine，也不能在释放后继续投递任务。

自动存档周期固定为5分钟，不增加运行配置。首次Tick使用`PlayerKey`稳定Hash得到`1s～5m`的初始延迟，之后每5分钟Tick一次，避免同一时刻创建或服务器启动恢复的Player集中提交。Timer使用不受GM调时影响的真实系统时间；回调必须进入GameService调度器后再读取或修改Player状态，不能从定时器执行线程直接并发访问Player。

每份持久化数据保存脏代数：发起存档时记录当前代数，`Await`返回后只有代数未变才能清除脏标记。存档期间再次修改的数据必须保持为脏，留待该Player下一次Timer Tick。定时存档或下线立即存档失败时不增加即时重试，只保持脏标记并记录错误日志。

断线处理顺序固定为：

```text
校验GatewayConnectionID
    -> 同步切换为Resident
    -> 倒序OnOffline
    -> 更新LastLogoutAt并标脏
    -> 立即存档
    -> 驻留15分钟
```

15分钟内重连时复用原Player，取消本次释放，更新连接并执行`OnOnline`，不重新从MongoDB加载。15分钟内始终未重连时，Player先进入`Releasing`停止新业务并执行一次最终存档；失败时不重试，只记录严重错误日志并继续释放。该简化策略可能丢失最近一次成功存档后的修改，是首期明确接受的取舍。

最终存档返回后，GameService先通过AccDBService把仍属于本实例的`RESIDENT`路由原子改为`LEAVING`并设置5秒TTL，再倒序执行`OnRelease`并移除本地Player。Gateway查询到`LEAVING`时不得调用旧GameService或立即分配其他GameService，只能在登录Deadline内等待并重新查询；TTL到期后才允许重新分配。`OnRelease`不得执行I/O或长时间阻塞，必须在该保护窗口内快速完成。

GameService停服使用固定90秒优雅期限，不增加配置。期限内完成全部Player最终存档；到期仍失败时记录未保存Player数量和严重告警，`OnStop`返回错误，由最终程序入口决定退出结果。Service不得调用`os.Exit`、无限阻塞或增加本地WAL。

### 7.1 Gateway断线Notify与连接索引

PlayerModule持有实例级`playersByConnectionID map[string]*Player`，它只是加速Gateway连接事件查找的运行期索引，Player的`DataInfo`仍是连接关系的唯一所有者。绑定连接时若同一个ConnectionID已经指向其他Player，视为内部错误并拒绝覆盖；换绑、转为Resident和最终释放时必须同步删除旧索引。

```go
// PlayerDisconnectedRequest通知GameService某个Gateway连接已经关闭。
type PlayerDisconnectedRequest struct {
	GatewayConnectionID string // 已关闭的全局唯一连接ID
}

//origin:rpc
type GameService interface {
	// PlayerDisconnected按连接ID让当前Player进入断线流程。
	PlayerDisconnected(context.Context, PlayerDisconnectedRequest) error
}
```

Gateway使用生成的`NotifyPlayerDisconnected`，请求不携带AccountID、ShowAreaID、GatewayNodeID或断线原因。PlayerModule查不到ConnectionID，或者索引对应Player已经换绑、Resident或Releasing时幂等忽略；只有索引与Player当前ConnectionID仍一致时才执行断线流程。

### 7.2 玩家逻辑心跳

客户端在线后每5秒发送一次`PlayerHeartbeatReq`。GameService的心跳Handler通过ConnectionID索引取得Player，在不跨`Await`的执行段中更新`LastHeartbeatAt`，并使用`MessageID = Ok`回复空Body。

PlayerModule使用一个固定5秒周期的统一扫描Timer检查在线Player，不为每个Player单独创建Timer。当前真实系统时间距离`LastHeartbeatAt`超过15秒时，GameService先删除ConnectionID索引、把Player同步切换为`Resident`并执行`OnOffline`，再通过精确Gateway节点调用`NotifyCloseClientConnection`。Gateway进程已经崩溃或连接已经关闭时通知失败不影响本地断线结果。

逻辑心跳和超时使用真实系统时间，不受Node游戏逻辑时间调整影响。Gateway连接关闭Notify与心跳超时可能重复到达，统一通过ConnectionID索引和Player状态幂等处理。

## 8. 消息执行、注册与Gateway下行

Origin Service在普通执行段内串行执行任务；调用`Await`时当前任务让出Service执行权，后续任务可能先执行。因此首期不增加通用`PlayerExecutor`、每玩家队列或额外并发配置：

- 不调用`Await`的Handler使用Origin默认串行语义，处理段内天然有序；
- 调用`Await`前先完成必要的同步状态转换；
- `Await`返回后必须确认PlayerModule中仍是原Player、`GatewayConnectionID`仍匹配且当前生命周期状态允许提交；
- 存档通过脏代数避免`Await`期间的新修改被错误清理；
- 加载、Resident和Releasing等生命周期竞态使用明确状态拒绝或让Gateway有界重试，不在Player后面积压无界等待任务；
- 只有将来出现确实需要跨`Await`严格串行的单个业务时，才在该Proxy内增加局部的有界状态保护。

### 8.1 Gateway转发与消息路由

普通玩家消息不通过通用`PlayerMessageResult{MessageID, ErrorCode, Body}`作为RPC返回。登录保留第6.3节的直接RPC返回，作为Gateway必须等待确定结果的特例。

```go
// PlayerMessageRequest是Gateway转发给GameService的一条客户端玩家消息。
type PlayerMessageRequest struct {
	GatewayConnectionID string // 本次请求来源连接，必须与Player当前绑定一致
	MessageID           commonpb.MessageID // 客户端请求消息ID
	Sequence            uint32             // 客户端请求序号，回复时原样使用
	Body                []byte             // 尚未反序列化的Protobuf Body
}

//origin:rpc
type GameService interface {
	// HandlePlayerMessage校验当前Player连接并路由一条客户端业务消息。
	// 业务响应不作为RPC返回值，由Handler通过Session或Player下行。
	HandlePlayerMessage(context.Context, PlayerMessageRequest) error
}
```

普通消息不携带AccountID、ShowAreaID或GatewayNodeID。路由器通过PlayerModule实例级ConnectionID索引直接取得Player，再比较索引键与Player当前`GatewayConnectionID`；不一致直接拒绝，旧连接消息不得进入Handler。Player当前`DataInfo`仍是唯一保存`GatewayNodeID + GatewayConnectionID`连接关系的地方。

### 8.2 Player下行与本次请求Session

Player统一负责调用GatewayService下行。`GatewayNodeID`只从Player当前`DataInfo`读取，避免同一连接关系在Player和Session中重复保存。

```go
// SendMsg向Player当前连接发送主动推送。
func (p *Player) SendMsg(messageID commonpb.MessageID, body proto.Message) error

// Reply回复指定客户端请求；连接已变化时静默不发送，绝不回复到新连接。
func (p *Player) Reply(
	requestConnectionID string,
	sequence uint32,
	messageID commonpb.MessageID,
	body proto.Message,
) error

// ReplyError回复指定客户端请求的业务错误，不携带业务Body。
func (p *Player) ReplyError(
	requestConnectionID string,
	sequence uint32,
	messageID commonpb.MessageID,
	code commonpb.ErrorCode,
) error
```

`Reply`和`ReplyError`先比较`requestConnectionID`与Player当前`GatewayConnectionID`。不相等表示该请求已被顶号或重连替代：不发送，也不把它当作处理失败。相等时才从Player取得当前`GatewayNodeID + GatewayConnectionID`，精确路由到Gateway并发送。

```go
// Session仅描述当前入站请求，不拥有也不保存Gateway连接关系。
type Session struct {
	player       *player.Player // 当前消息处理目标，只在本次处理链使用
	connectionID string         // 本次请求来源连接，用于跨Await后的回复校验
	sequence     uint32         // 本次请求序号
}

// Reply回复当前请求。
func (s *Session) Reply(messageID commonpb.MessageID, body proto.Message) error

// ReplyError回复当前请求的业务错误。
func (s *Session) ReplyError(messageID commonpb.MessageID, code commonpb.ErrorCode) error
```

`Session.Reply`和`Session.ReplyError`只转调Player对应方法。`Session -> *Player`不是循环引用，Player不得反向保存Session；Session不得被Proxy、Player或后台任务持有，生命周期仅限当前消息处理链。`Await`返回后回复仍使用本次`connectionID`校验，旧请求不会误发给新连接。

### 8.3 显式消息注册

消息注册统一放在`service/gameservice/messagehandler`包，由`messagehandler/register.go`集中登记。PlayerModule在初始化阶段创建实例化Router、调用`messagehandler.Register(router)`并冻结Router；完成冻结后才允许处理客户端业务消息。

禁止使用`init()`、包级可变Map或隐式注册。Router及其路由表由PlayerModule实例持有，一个GameService实例对应一个独立注册表。

```go
// Register统一登记当前GameService的客户端业务消息。
func Register(router *msgrouter.Router) error {
	return router.Register(
		commonpb.MessageID_RenameReq,
		handleRename,
	)
}

// handleRename处理改名请求，并通过Session回复该请求。
func handleRename(
	session *msgrouter.Session,
	player *player.Player,
	request *commonpb.RenameRequest,
) error {
	// 具体业务由UserInfoProxy完成。
	return session.Reply(
		commonpb.MessageID_RenameRes,
		&commonpb.RenameResponse{},
	)
}
```

Router的`Register`使用Go 1.27泛型接收具体Protobuf请求类型，在注册时确定反序列化类型和Handler：

```go
func (r *Router) Register[T proto.Message](
	messageID commonpb.MessageID,
	handler func(*Session, *player.Player, T) error,
) error
```

多个真实消息时，`Register`内按同样方式逐条注册并在重复MessageID时启动失败；不为单个功能额外包装`registerXxxRoutes`函数。`messagehandler`按真实业务拆分Handler文件，例如`userinfohandler.go`；`register.go`只保留统一登记。

RPC的`error`只表示GameService未能正常取得、解码或执行该处理任务，不承载客户端业务错误。客户端业务错误由Handler通过`Session.ReplyError`返回对应`MessageID + Sequence + ErrorCode`。
