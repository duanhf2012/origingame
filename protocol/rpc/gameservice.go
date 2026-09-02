package rpcapi

import (
	"context"

	commonpb "origingame/protocol/common"
)

// GameService 是 Gateway 使用的玩家登录与消息入口。
//
//origin:rpc
type GameService interface {
	// LoginPlayer 完成玩家加载、连接绑定和上线回调，并返回客户端登录数据。
	LoginPlayer(context.Context, LoginPlayerRequest) (*commonpb.LoginPlayerResult, error)
	// HandlePlayerMessage 校验当前玩家连接并路由一条客户端业务消息。
	HandlePlayerMessage(context.Context, PlayerMessageRequest) error
	// PlayerDisconnected 按连接 ID 让当前玩家进入断线流程。
	PlayerDisconnected(context.Context, PlayerDisconnectedRequest) error
}

// LoginPlayerRequest 用于进入 GameService 并绑定当前 Gateway 连接。
type LoginPlayerRequest struct {
	AccountID  string // 已验签的账号标识。
	ShowAreaID int64  // 玩家选择的显示区服标识。

	ExpectedGameServiceNodeSessionID string // Redis 分配实例的 Node 会话标识。
	GatewayNodeID                    string // 当前 Gateway Node 标识。
	GatewayConnectionID              string // 当前客户端连接标识。
}

// PlayerMessageRequest 是 Gateway 转发给 GameService 的客户端玩家消息。
type PlayerMessageRequest struct {
	GatewayConnectionID string             // 当前客户端连接标识。
	MessageID           commonpb.MessageID // 客户端业务消息标识。
	Sequence            uint32             // 客户端请求序号。
	Body                []byte             // 未解码的业务负载。
}

// PlayerDisconnectedRequest 通知 GameService 某个 Gateway 连接已经关闭。
type PlayerDisconnectedRequest struct {
	GatewayConnectionID string // 已关闭的客户端连接标识。
}
