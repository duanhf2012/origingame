package mongodb

// RealAreaInfoName 是真实区服基础集合名称。
const RealAreaInfoName = "RealAreaInfo"

// GatewayEndpoint 是 RealAreaInfo 中保存的 Gateway 公网入口。
type GatewayEndpoint struct {
	Protocol string `bson:"Protocol"` // 客户端连接协议。
	Address  string `bson:"Address"`  // Gateway 公网地址。
}

// RealAreaInfo 是 RealAreaInfo 集合的 BSON 文档结构。
type RealAreaInfo struct {
	ID       int64             `bson:"_id"`      // 真实区服标识。
	GateList []GatewayEndpoint `bson:"GateList"` // 可用 Gateway 公网入口。
}
