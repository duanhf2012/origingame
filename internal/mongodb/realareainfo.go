package mongodb

// RealAreaInfoName 是真实区服基础集合名称。
const RealAreaInfoName = "RealAreaInfo"

// GatewayEndpoint 是 RealAreaInfo 中保存的 Gateway 公网入口。
type GatewayEndpoint struct {
	Protocol string `bson:"Protocol"`
	Address  string `bson:"Address"`
}

// RealAreaInfo 是 RealAreaInfo 集合的 BSON 文档结构。
type RealAreaInfo struct {
	ID       int64             `bson:"_id"`
	GateList []GatewayEndpoint `bson:"GateList"`
}
