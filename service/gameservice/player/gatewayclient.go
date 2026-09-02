package player

import (
	"context"

	rpcapi "origingame/protocol/rpc"
)

// GatewayClient 定向向玩家当前 Gateway 发送消息或关闭连接。
type GatewayClient interface {
	SendClientMessage(string, rpcapi.SendClientMessageRequest) error
	CloseClientConnection(string, string) error
}

// GatewayRPCClient 通过 GatewayService RPC 实现 Player 的下行能力。
type GatewayRPCClient struct {
	client rpcapi.GatewayServiceClient // 通过服务发现路由的 Gateway 客户端。
}

// NewGatewayRPCClient 创建面向指定 GatewayService 发现范围的调用器。
func NewGatewayRPCClient(client rpcapi.GatewayServiceClient) *GatewayRPCClient {
	return &GatewayRPCClient{client: client}
}

// SendClientMessage 将下行消息投递到指定 Gateway Node。
func (client *GatewayRPCClient) SendClientMessage(nodeID string, request rpcapi.SendClientMessageRequest) error {
	return client.client.OnNode(nodeID).NotifySendClientMessage(context.Background(), request)
}

// CloseClientConnection 通知指定 Gateway Node 关闭客户端连接。
func (client *GatewayRPCClient) CloseClientConnection(nodeID string, connectionID string) error {
	return client.client.OnNode(nodeID).NotifyCloseClientConnection(
		context.Background(),
		rpcapi.CloseClientConnectionRequest{GatewayConnectionID: connectionID},
	)
}
