# Service RPC 契约目录

本目录统一保存 OriginGame 各 Service 对其他 Service 公开的 Origin RPC 契约。

目录保持平铺：每个 Service 使用一个以完整 Service 名命名的手写 Go 文件，例如 `gameservice.go`、`gatewayservice.go`；`origingen` 在同目录生成对应的 `*.rpc.gen.go`。RPC 实现仍放在 `service/<xxxservice>`，不得放入本目录。

`gameservice.go.example` 和 `gameservice.proto.example` 是未参与编译和生成的目录样例，其中同时展示普通 Go 参数和 Protobuf 参数。设计确认第一个真实 RPC 后，将样例内容按实际契约写入对应源文件并运行仓库生成脚本；不要直接把未确认的样例改名为生产契约。

```text
protocol/rpc/
├── README.md
├── gameservice.go.example       # 目录样例，不参与编译
├── gameservice.proto.example    # Protobuf RPC 参数样例
├── gameservice.go               # 手写真实契约，确认后创建
├── gameservice.proto            # 手写 Protobuf 参数，按需创建
├── gameservice.pb.go            # protoc-gen-go 自动生成
└── gameservice.rpc.gen.go       # origingen 自动生成
```

`protocol/rpc` 整体使用 `rpcapi` 包。普通 Go 参数与所属 Service 的 RPC 接口放在 `<service>.go`；需要使用 Protobuf 线格式的 RPC 参数放在同层 `<service>.proto`，生成的 `<service>.pb.go` 也属于 `rpcapi`，接口可直接引用消息类型。Protobuf 类型必须作为 RPC 方法的顶层参数或顶层返回值使用；不得把它嵌入普通 Go 参数结构体后误认为仍会使用官方 Protobuf Codec。
