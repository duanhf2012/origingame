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
	MessageID       commonpb.MessageID // 下行消息标识。
	Sequence        uint32             // 对应请求序号，零表示主动推送。
	ErrorCode       commonpb.ErrorCode // 响应错误码。
	Body            []byte             // 未编码业务负载。
	CloseAfterWrite bool               // 写入成功后关闭连接。
}

// SendClientMessageRequest 指定接收下行消息的 Gateway 连接。
type SendClientMessageRequest struct {
	GatewayConnectionID string        // 目标客户端连接标识。
	Message             ClientMessage // 待下行消息。
}

// CloseClientConnectionRequest 指定需要关闭的客户端连接。
type CloseClientConnectionRequest struct {
	GatewayConnectionID string // 待关闭的客户端连接标识。
}
