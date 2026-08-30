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
	GatewayNodeID       string
	GatewayConnectionID string
	State               State
	LastHeartbeatAt     time.Time
	ResidentDeadline    time.Time
}

// CUserInfo 是 UserInfo 集合的玩家基础持久化数据。
type CUserInfo = mongodb.UserInfo

// Player 保存稳定身份、连接状态、基础数据和固定顺序的功能 Proxy。
type Player struct {
	key        string
	accountID  string
	showAreaID int64
	realAreaID int64
	dataInfo   DataInfo
	userInfo   CUserInfo

	userInfoProxy UserInfoProxy
	proxies       []PlayerProxy
	persistent    []*persistentDataEntry
	userInfoEntry *persistentDataEntry
	released      bool
	saveTimer     *time.Timer
	gateway       GatewayClient
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
