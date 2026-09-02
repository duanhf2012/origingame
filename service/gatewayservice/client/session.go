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
	session network.Session // 关联的底层网络会话。
	state   sessionState    // 当前登录阶段。

	windowStarted time.Time // 当前限流窗口起点。
	windowCount   int       // 当前窗口消息数。

	loginSequence uint32                              // 正在处理的登录请求序号。
	loginAttempt  uint64                              // 单调递增的登录尝试号。
	accountID     string                              // 已登录账号标识。
	showAreaID    int64                               // 已登录显示区服标识。
	realAreaID    int64                               // 已登录真实区服标识。
	gameService   playerownership.GameServiceInstance // 当前归属 GameService。
	loginResult   []byte                              // 幂等登录响应缓存。
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
