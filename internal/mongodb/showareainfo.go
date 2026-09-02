package mongodb

// ShowAreaInfoName 是显示区服基础集合名称。
const ShowAreaInfoName = "ShowAreaInfo"

// ShowAreaInfo 是 ShowAreaInfo 集合的 BSON 文档结构。
// OpenTime 是需要响应 GM 调时的游戏业务时间戳，其时区和精度由区服设计统一约束。
type ShowAreaInfo struct {
	ID           int64  `bson:"_id"`          // 显示区服标识。
	AreaName     string `bson:"AreaName"`     // 客户端展示名称。
	RealAreaID   int64  `bson:"RealAreaId"`   // 实际运行的真实区服标识。
	ServerMark   int32  `bson:"ServerMark"`   // 客户端展示标记。
	ServerStatus int32  `bson:"ServerStatus"` // 客户端可见区服状态。
	OpenTime     int64  `bson:"OpenTime"`     // 受 GM 调时影响的开服时间戳。
}
