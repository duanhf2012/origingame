# OriginGame RPC 契约与生成设计

> 状态：目录、正式契约边界与生成规则已确认
> 更新日期：2026-08-15

## 1. 目录与命名

Service 间 RPC 统一使用 Origin v3 的 Go 契约与 `origingen`，不使用老版本的 RPC Protobuf：

```text
protocol/rpc/
├── gameservice.go
├── gameservice.rpc.gen.go
├── gatewayservice.go
└── gatewayservice.rpc.gen.go
```

目录保持平铺且 Go 包名固定为 `rpcapi`。每个提供 RPC 的 Service 只有一个 `<service>.go` 手写文件，其中集中声明该 Service 的 `//origin:rpc` 接口、普通 Go 请求、返回和专用枚举。需要 Protobuf 参数时，同层增加 `<service>.proto`。对应的 `.rpc.gen.go` 和 `.pb.go` 由生成器产生并提交仓库。

当前正式契约固定为GameService的`LoginPlayer`、`HandlePlayerMessage`、`PlayerDisconnected`和GatewayService的`SendClientMessage`、`CloseClientConnection`；其中断线和主动关闭均由调用方使用生成的`NotifyXxx`。实现阶段创建真实`gameservice.go`、`gatewayservice.go`及对应`.rpc.gen.go`，不能直接把样例改名为生产契约。

上述请求全部使用普通Go结构和`origingen`静态Codec；`LoginPlayer`顶层返回已经存在于`protocol/common`的`*commonpb.LoginPlayerResult`。首期没有RPC专用Protobuf参数，因此不创建`gameservice.proto`、`gatewayservice.proto`或空的`.pb.go`；以后出现真实RPC专用Protobuf参数时再按本设计增加。

RPC 契约方法按具体业务动作命名，不添加 `Rpc` 或 `RPC` 前缀。调用方式由 `origingen` 生成的 `Await`、`Call`、`Async`、`Notify` 和 `Broadcast` 前缀表达；例如契约方法 `ExecuteMongo` 生成 `AwaitExecuteMongo` 和 `NotifyExecuteMongo`，不得声明成会生成 `AwaitRpcExecuteMongo` 的 `RpcExecuteMongo`。

## 2. 依赖边界

- `protocol/rpc` 不得导入 `service/<xxxservice>`；
- Service 实现导入 `protocol/rpc`，并用编译期断言确认实现对应接口；
- 普通 Go 参数与所属 Service 接口共同定义在 `protocol/rpc/<service>.go`，由 `origingen` 建立静态 Codec；
- Protobuf 参数定义在同层 `protocol/rpc/<service>.proto`，生成的 `<service>.pb.go` 与 Go 契约共同属于 `rpcapi` 包；
- Protobuf 消息必须直接用于 RPC 方法的顶层参数或顶层业务返回值，推荐使用指针类型，只有该位置会选择官方 Protobuf Codec；
- 同一个 Service 接口允许不同方法分别使用普通 Go 参数和 Protobuf 参数，但单个参数的线格式必须由其顶层类型明确决定；
- RPC 参数不默认复用客户端 Protobuf、MongoDB Collection、HTTP DTO 或 Service 私有结构；
- 只有已确认需要跨语言或直接共享既有 Protobuf 线格式时，才允许顶层 RPC 参数采用 Protobuf。

当前已确认一个明确例外：GameService的`LoginPlayer`直接返回`*commonpb.LoginPlayerResult`。该消息本身就是Gateway向客户端转发的登录成功Body，复用可以避免重复结构和无价值转换；它必须作为RPC顶层返回值使用，不能嵌入普通Go结果结构后再假定使用官方Protobuf Codec。

Gateway与GameService之间的高频消息请求固定使用普通Go结构和`origingen`静态Codec：`PlayerMessageRequest`只携带ConnectionID、MessageID、Sequence和已经序列化的Protobuf Body；`SendClientMessageRequest`只携带ConnectionID和下行消息。客户端Body不再为内部RPC外层包装Protobuf，避免额外的`proto.Size`、Marshal和Unmarshal通用路径。实现时必须使用典型小消息和1KB Body保留普通Go静态Codec与Protobuf包装的对比Benchmark。

## 3. 生成流程

统一生成脚本通过以下命令使用当前 `go.mod` 固定的 Origin 版本，不在每个平铺的 Service 契约文件中重复声明相同的 `go:generate`，避免一次生成流程重复扫描整个包：

```text
go run github.com/duanhf2012/origin/v3/cmd/origingen rpc ./protocol/rpc
```

统一生成脚本会先使用固定版本的 `protoc` 生成 `protocol/rpc/*.proto`，再运行 `origingen`，保证 RPC 契约引用的同包 Protobuf 类型已经存在。只有 `.proto.example` 样例时跳过 Protobuf 生成。

提交前执行：

```text
go run github.com/duanhf2012/origin/v3/cmd/origingen rpc --check ./protocol/rpc
```

运行期只使用生成的强类型 Client、静态编解码、Dispatcher 和描述符，不在 RPC 热路径扫描接口或使用反射。生成文件禁止手工修改。
