package client

import (
	"time"

	"github.com/duanhf2012/origin/v3/sysmodule/network"
	"origingame/internal/playerownership"
)

const (
	messageWindow     = 10 * time.Second
	messageWindowMax  = 200
	loginTotalTimeout = 30 * time.Second
)

type sessionState uint8

const (
	stateConnected sessionState = iota + 1
	stateLoggingIn
	stateOnline
)

// connection 保存单条网络连接当前唯一的登录与 GameService 归属状态。
type connection struct {
	session network.Session
	state   sessionState

	windowStarted time.Time
	windowCount   int

	loginSequence uint32
	loginAttempt  uint64
	accountID     string
	showAreaID    int64
	realAreaID    int64
	gameService   playerownership.GameServiceInstance
	loginResult   []byte
}

func newConnection(session network.Session, now time.Time) *connection {
	// 新连接从未登录状态和当前限流窗口开始。
	return &connection{session: session, state: stateConnected, windowStarted: now}
}

// allowMessage 使用固定十秒窗口执行连接级有界限流。
func (connection *connection) allowMessage(now time.Time) bool {
	// 窗口到期或时钟回退时重新计数，防止旧窗口长期生效。
	if now.Sub(connection.windowStarted) >= messageWindow || now.Before(connection.windowStarted) {
		connection.windowStarted = now
		connection.windowCount = 0
	}
	// 仅允许窗口内的前 messageWindowMax 条消息。
	connection.windowCount++
	return connection.windowCount <= messageWindowMax
}
