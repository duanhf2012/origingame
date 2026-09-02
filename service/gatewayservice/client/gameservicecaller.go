package client

import (
	"context"

	"github.com/duanhf2012/origin/v3/service"
	"origingame/internal/playerownership"
	commonpb "origingame/protocol/common"
	rpcapi "origingame/protocol/rpc"
)

// GameServiceCaller 是 Gateway 客户端入口所需的 GameService 调用能力。
type GameServiceCaller interface {
	LoginPlayer(context.Context, playerownership.GameServiceInstance, rpcapi.LoginPlayerRequest) (*commonpb.LoginPlayerResult, error)
	HandlePlayerMessage(playerownership.GameServiceInstance, rpcapi.PlayerMessageRequest) error
	PlayerDisconnected(playerownership.GameServiceInstance, string) error
}

// RPCGameServiceCaller 通过 Origin 服务发现定向调用指定的 GameService 实例。
type RPCGameServiceCaller struct {
	owner service.IService // 用于绑定 GameService RPC 客户端的 GatewayService。
}

// NewRPCGameServiceCaller 创建由 GatewayService 装配的 GameService 调用器。
func NewRPCGameServiceCaller(owner service.IService) *RPCGameServiceCaller {
	// 保存 GatewayService，供后续按玩家归属绑定 RPC 客户端。
	return &RPCGameServiceCaller{owner: owner}
}

// LoginPlayer 让指定 GameService 完成玩家加载和上线。
func (caller *RPCGameServiceCaller) LoginPlayer(
	ctx context.Context,
	gameService playerownership.GameServiceInstance,
	request rpcapi.LoginPlayerRequest,
) (*commonpb.LoginPlayerResult, error) {
	// 定向调用已分配的 GameService 完成玩家上线。
	return caller.gameClient(gameService).CallLoginPlayer(ctx, request)
}

// HandlePlayerMessage 将一条客户端业务消息投递到当前玩家归属实例。
func (caller *RPCGameServiceCaller) HandlePlayerMessage(
	gameService playerownership.GameServiceInstance,
	request rpcapi.PlayerMessageRequest,
) error {
	// 使用 Notify 投递业务消息，不阻塞 Gateway 网络入口。
	return caller.gameClient(gameService).NotifyHandlePlayerMessage(context.Background(), request)
}

// PlayerDisconnected 通知当前归属实例回收客户端连接状态。
func (caller *RPCGameServiceCaller) PlayerDisconnected(
	gameService playerownership.GameServiceInstance,
	connectionID string,
) error {
	// 使用 Notify 让归属实例异步回收连接状态。
	return caller.gameClient(gameService).NotifyPlayerDisconnected(
		context.Background(), rpcapi.PlayerDisconnectedRequest{GatewayConnectionID: connectionID},
	)
}

func (caller *RPCGameServiceCaller) gameClient(gameService playerownership.GameServiceInstance) rpcapi.GameServiceClient {
	// 以区服标签和节点标识精确路由到玩家归属实例。
	return rpcapi.BindGameServiceTo(caller.owner, gameService.ServiceName).
		WhereLabels(map[string]string{"scope": "area"}).OnNode(gameService.NodeID)
}
