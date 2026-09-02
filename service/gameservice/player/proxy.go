package player

// PlayerLoadContext 描述全部持久化数据完成反序列化后的加载事实。
type PlayerLoadContext struct {
	IsNewPlayer bool // 是否首次创建玩家。
}

// PlayerOnlineContext 描述一次连接上线。
type PlayerOnlineContext struct {
	GatewayNodeID       string // 当前连接所属 Gateway Node。
	GatewayConnectionID string // 当前客户端连接标识。
}

// PlayerOfflineContext 描述当前连接离线。
type PlayerOfflineContext struct {
	GatewayConnectionID string // 断开的客户端连接标识。
}

// PlayerProxy 是注册到单个 Player 上的完整功能生命周期。
type PlayerProxy interface {
	OnInit(*Player) error
	OnLoaded(PlayerLoadContext) error
	OnAllLoaded(PlayerLoadContext) error
	OnOnline(PlayerOnlineContext)
	OnOffline(PlayerOfflineContext)
	OnRelease()
}

// BasePlayerProxy 提供默认空回调和对所属 Player 公共数据的统一访问。
type BasePlayerProxy struct {
	player *Player // 所属玩家对象。
}

func (proxy *BasePlayerProxy) OnInit(player *Player) error {
	proxy.player = player
	return nil
}
func (*BasePlayerProxy) OnLoaded(PlayerLoadContext) error    { return nil }
func (*BasePlayerProxy) OnAllLoaded(PlayerLoadContext) error { return nil }
func (*BasePlayerProxy) OnOnline(PlayerOnlineContext)        {}
func (*BasePlayerProxy) OnOffline(PlayerOfflineContext)      {}
func (*BasePlayerProxy) OnRelease()                          {}
func (proxy *BasePlayerProxy) Player() *Player               { return proxy.player }
func (proxy *BasePlayerProxy) DataInfo() *DataInfo           { return proxy.player.DataInfo() }
func (proxy *BasePlayerProxy) UserInfo() *CUserInfo          { return proxy.player.UserInfo() }
func (proxy *BasePlayerProxy) UserInfoProxy() *UserInfoProxy { return proxy.player.UserInfoProxy() }

func (player *Player) registerProxy(proxy PlayerProxy) {
	player.proxies = append(player.proxies, proxy)
}
