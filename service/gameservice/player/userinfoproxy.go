package player

import "time"

// UserInfoProxy 拥有 CUserInfo 的初始化和全部业务修改入口。
type UserInfoProxy struct{ BasePlayerProxy }

func (proxy *UserInfoProxy) OnLoaded(ctx PlayerLoadContext) error {
	if ctx.IsNewPlayer {
		now := time.Now().UTC()
		info := proxy.UserInfo()
		info.PlayerKey = proxy.Player().Key()
		parts := proxy.Player().routePlayer()
		info.AccountID = parts.AccountID
		info.ShowAreaID = parts.ShowAreaID
		info.Level = 1
		info.CreatedAt = now
		info.LastLoginAt = now
		proxy.Player().markUserInfoDirty()
	}
	return nil
}

func (proxy *UserInfoProxy) OnOnline(PlayerOnlineContext) {
	proxy.UserInfo().LastLoginAt = time.Now().UTC()
	proxy.Player().markUserInfoDirty()
}

func (proxy *UserInfoProxy) OnOffline(PlayerOfflineContext) {
	proxy.UserInfo().LastLogoutAt = time.Now().UTC()
	proxy.Player().markUserInfoDirty()
}

// SetNickname 修改昵称并在真实变化时标脏。
func (proxy *UserInfoProxy) SetNickname(value string) {
	if proxy.UserInfo().Nickname == value {
		return
	}
	proxy.UserInfo().Nickname = value
	proxy.Player().markUserInfoDirty()
}

// SetLevel 修改等级并在真实变化时标脏。
func (proxy *UserInfoProxy) SetLevel(value int32) {
	if value <= 0 || proxy.UserInfo().Level == value {
		return
	}
	proxy.UserInfo().Level = value
	proxy.Player().markUserInfoDirty()
}
