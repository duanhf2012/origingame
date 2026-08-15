package rpcapi

import (
	"context"

	commonpb "origingame/protocol/common"
)

// GatewayService 是 GameService 使用的统一客户端下行入口。
//
//origin:rpc
type GatewayService interface {
	// SendClientMessage 向指定客户端连接发送消息。
	SendClientMessage(context.Context, SendClientMessageRequest) error
	// CloseClientConnection 关闭指定连接；连接不存在时幂等成功。
	CloseClientConnection(context.Context, CloseClientConnectionRequest) error
}

// ClientMessage 描述 Gateway 需要写入客户端连接的一条消息。
type ClientMessage struct {
	MessageID       commonpb.MessageID
	Sequence        uint32
	ErrorCode       commonpb.ErrorCode
	Body            []byte
	CloseAfterWrite bool
}

// SendClientMessageRequest 指定接收下行消息的 Gateway 连接。
type SendClientMessageRequest struct {
	GatewayConnectionID string
	Message             ClientMessage
}

// CloseClientConnectionRequest 指定需要关闭的客户端连接。
type CloseClientConnectionRequest struct {
	GatewayConnectionID string
}
