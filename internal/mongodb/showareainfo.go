package mongodb

// ShowAreaInfoName 是显示区服基础集合名称。
const ShowAreaInfoName = "ShowAreaInfo"

// ShowAreaInfo 是 ShowAreaInfo 集合的 BSON 文档结构。
// OpenTime 是需要响应 GM 调时的游戏业务时间戳，其时区和精度由区服设计统一约束。
type ShowAreaInfo struct {
	ID           int64  `bson:"_id"`
	AreaName     string `bson:"AreaName"`
	RealAreaID   int64  `bson:"RealAreaId"`
	ServerMark   int32  `bson:"ServerMark"`
	ServerStatus int32  `bson:"ServerStatus"`
	OpenTime     int64  `bson:"OpenTime"`
}
