// Package area 负责区服数据加载、校验、快照发布和周期刷新。
package area

// GateInfo 是客户端可选择的共享 Gateway 公网入口。
type GateInfo struct {
	Protocol string `json:"Protocol"`
	Address  string `json:"Address"`
}

// Info 是 LoginService 返回给客户端的显示区服视图，不暴露 RealAreaId。
type Info struct {
	ShowAreaID   int64      `json:"ShowAreaId"`
	AreaName     string     `json:"AreaName"`
	ServerMark   int32      `json:"ServerMark"`
	ServerStatus int32      `json:"ServerStatus"`
	OpenTime     int64      `json:"OpenTime"`
	GateList     []GateInfo `json:"GateList"`
}
