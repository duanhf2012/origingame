// Package area 负责区服数据加载、校验、快照发布和周期刷新。
package area

// GateInfo 是客户端可选择的共享 Gateway 公网入口。
type GateInfo struct {
	Protocol string `json:"Protocol"` // 客户端连接协议。
	Address  string `json:"Address"`  // Gateway 公网地址。
}

// Info 是 LoginService 返回给客户端的显示区服视图，不暴露 RealAreaId。
type Info struct {
	ShowAreaID   int64      `json:"ShowAreaId"`   // 显示区服标识。
	AreaName     string     `json:"AreaName"`     // 客户端展示名称。
	ServerMark   int32      `json:"ServerMark"`   // 客户端展示标记。
	ServerStatus int32      `json:"ServerStatus"` // 客户端可见区服状态。
	OpenTime     int64      `json:"OpenTime"`     // 游戏业务开服时间戳。
	GateList     []GateInfo `json:"GateList"`     // 可连接的 Gateway 列表。
}
