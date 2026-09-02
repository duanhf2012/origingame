// Package virtualplayer 实现单个机器人真实客户端状态、HTTP登录和Gateway协议。
package virtualplayer

import (
	"encoding/binary"
	"errors"

	commonpb "origingame/protocol/common"
)

const (
	requestHeaderSize  = 6
	responseHeaderSize = 10
)

// InboundMessage 是已经完成公共头校验的Gateway下行消息。
type InboundMessage struct {
	MessageID commonpb.MessageID // Gateway 下行消息标识。
	Sequence  uint32             // 对应请求序号，零表示主动推送。
	ErrorCode commonpb.ErrorCode // 响应业务错误码。
	Body      []byte             // 未解码业务负载。
}

// EncodeRequest 使用与Gateway相同的BigEndian客户端请求头。
func EncodeRequest(messageID commonpb.MessageID, sequence uint32, body []byte) ([]byte, error) {
	if messageID <= commonpb.MessageID_Ok || messageID > 65535 || sequence == 0 {
		return nil, errors.New("机器人请求MessageID或Sequence无效")
	}
	payload := make([]byte, requestHeaderSize+len(body))
	binary.BigEndian.PutUint16(payload[:2], uint16(messageID))
	binary.BigEndian.PutUint32(payload[2:6], sequence)
	copy(payload[requestHeaderSize:], body)
	return payload, nil
}

// DecodeInbound 区分Sequence为零的主动推送和带错误码的普通响应。
func DecodeInbound(payload []byte) (InboundMessage, error) {
	if len(payload) < requestHeaderSize {
		return InboundMessage{}, errors.New("Gateway下行消息头不完整")
	}
	messageID := commonpb.MessageID(binary.BigEndian.Uint16(payload[:2]))
	sequence := binary.BigEndian.Uint32(payload[2:6])
	if messageID < 0 || messageID > 65535 {
		return InboundMessage{}, errors.New("Gateway下行MessageID无效")
	}
	if sequence == 0 {
		return InboundMessage{
			MessageID: messageID,
			Body:      append([]byte(nil), payload[requestHeaderSize:]...),
		}, nil
	}
	if len(payload) < responseHeaderSize {
		return InboundMessage{}, errors.New("Gateway响应头不完整")
	}
	return InboundMessage{
		MessageID: messageID,
		Sequence:  sequence,
		ErrorCode: commonpb.ErrorCode(int32(binary.BigEndian.Uint32(payload[6:10]))),
		Body:      append([]byte(nil), payload[responseHeaderSize:]...),
	}, nil
}
