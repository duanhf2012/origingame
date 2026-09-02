// Package player 实现单个 GameService 内的玩家对象、Proxy 和持久化生命周期。
package player

import (
	"errors"
	"time"

	"origingame/internal/mongodb"
	"origingame/internal/playerownership"
)

// State 是 Player 在本地内存中的生命周期状态。
type State int32

const (
	StateLoading State = iota
	StateOnline
	StateResident
	StateReleasing
	StateReleased
)

// DataInfo 只保存当前 Player 生命周期内的非持久化状态。
type DataInfo struct {
	GatewayNodeID       string    // 当前连接所属 Gateway Node。
	GatewayConnectionID string    // 当前客户端连接标识。
	State               State     // 本地玩家生命周期状态。
	LastHeartbeatAt     time.Time // 最近业务心跳真实时间。
	ResidentDeadline    time.Time // 断线驻留到期真实时间。
}

// CUserInfo 是 UserInfo 集合的玩家基础持久化数据。
type CUserInfo = mongodb.UserInfo

// Player 保存稳定身份、连接状态、基础数据和固定顺序的功能 Proxy。
type Player struct {
	key        string    // 账号与显示区服组成的稳定键。
	accountID  string    // 已验证的账号标识。
	showAreaID int64     // 所属显示区服标识。
	realAreaID int64     // 所属真实区服标识。
	dataInfo   DataInfo  // 非持久化生命周期状态。
	userInfo   CUserInfo // 基础持久化角色数据。

	userInfoProxy UserInfoProxy          // 基础角色数据 Proxy。
	proxies       []PlayerProxy          // 固定顺序的功能 Proxy。
	persistent    []*persistentDataEntry // 所有持久化数据登记项。
	userInfoEntry *persistentDataEntry   // 基础角色数据登记项。
	released      bool                   // 是否已完成释放。
	saveTimer     *time.Timer            // 下一次持久化检查 Timer。
	gateway       GatewayClient          // Gateway 下行能力。
}

// RecordHeartbeat 只更新当前有效在线连接的逻辑心跳时间。
func (player *Player) RecordHeartbeat(connectionID string, now time.Time) bool {
	if player == nil || player.dataInfo.State != StateOnline ||
		player.dataInfo.GatewayConnectionID != connectionID {
		return false
	}
	player.dataInfo.LastHeartbeatAt = now.UTC()
	return true
}

// NewPlayer 创建并初始化当前样板版本的全部持久化数据和 Proxy。
func NewPlayer(accountID string, showAreaID int64, realAreaID int64) (*Player, error) {
	if accountID == "" || showAreaID <= 0 || realAreaID <= 0 {
		return nil, errors.New("Player 身份无效")
	}
	current := &Player{
		key: playerownership.PlayerKey(accountID, showAreaID), accountID: accountID, showAreaID: showAreaID,
		realAreaID: realAreaID,
		dataInfo:   DataInfo{State: StateLoading},
	}
	entry, err := current.registerPersistentData(mongodb.UserInfoName, &current.userInfo)
	if err != nil {
		return nil, err
	}
	current.userInfoEntry = entry
	current.registerProxy(&current.userInfoProxy)
	if err = current.initialize(); err != nil {
		return nil, err
	}
	return current, nil
}

// Key 返回 AccountID 与 ShowAreaID 组成的稳定 PlayerKey。
func (player *Player) Key() string { return player.key }

// RealAreaID 返回当前 GameService 的真实区服归属。
func (player *Player) RealAreaID() int64 { return player.realAreaID }

// DataInfo 返回只由 PlayerModule 修改的当前运行状态。
func (player *Player) DataInfo() *DataInfo { return &player.dataInfo }

// UserInfo 返回供各 Proxy 读取的基础持久化数据。
func (player *Player) UserInfo() *CUserInfo { return &player.userInfo }

// UserInfoProxy 返回基础数据唯一业务修改入口。
func (player *Player) UserInfoProxy() *UserInfoProxy { return &player.userInfoProxy }

func (player *Player) ownershipPlayer() playerownership.Player {
	return playerownership.Player{
		AccountID:  player.accountID,
		ShowAreaID: player.showAreaID,
		RealAreaID: player.realAreaID,
	}
}
