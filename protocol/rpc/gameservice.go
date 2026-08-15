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
	AccountID  string
	ShowAreaID int64

	GatewayNodeID       string
	GatewayConnectionID string
}

// PlayerMessageRequest 是 Gateway 转发给 GameService 的客户端玩家消息。
type PlayerMessageRequest struct {
	GatewayConnectionID string
	MessageID           commonpb.MessageID
	Sequence            uint32
	Body                []byte
}

// PlayerDisconnectedRequest 通知 GameService 某个 Gateway 连接已经关闭。
type PlayerDisconnectedRequest struct {
	GatewayConnectionID string
}
