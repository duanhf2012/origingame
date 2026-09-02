// Package client 实现 Gateway 三种传输共用的客户端协议和会话入口。
package client

import (
	"encoding/binary"
	"errors"

	commonpb "origingame/protocol/common"
	rpcapi "origingame/protocol/rpc"
)

const (
	requestHeaderSize  = 6
	responseHeaderSize = 10
)

// request 是已完成通用头校验的客户端消息。
type request struct {
	messageID commonpb.MessageID
	sequence  uint32
	body      []byte
}

func decodeRequest(payload []byte) (request, error) {
	// 先确保固定长度请求头完整。
	if len(payload) < requestHeaderSize {
		return request{}, errors.New("客户端消息头不完整")
	}
	// 按已确认的大端格式读取通用请求头。
	messageID := commonpb.MessageID(binary.BigEndian.Uint16(payload[:2]))
	sequence := binary.BigEndian.Uint32(payload[2:6])
	if sequence == 0 || messageID == commonpb.MessageID_Ok {
		return request{}, errors.New("客户端 MessageID 或 Sequence 无效")
	}
	return request{messageID: messageID, sequence: sequence, body: payload[requestHeaderSize:]}, nil
}

// encodeClientMessage 编码响应或主动推送；网络 Module 再负责传输帧。
func encodeClientMessage(message rpcapi.ClientMessage) ([]byte, error) {
	// MessageID 必须能编码为协议规定的 uint16。
	if message.MessageID < 0 || message.MessageID > 65535 {
		return nil, errors.New("下行 MessageID 超出 uint16")
	}
	if message.Sequence == 0 {
		// 主动推送只携带请求头和业务负载。
		if message.ErrorCode != commonpb.ErrorCode_ERROR_CODE_OK {
			return nil, errors.New("主动推送不能携带响应错误码")
		}
		payload := make([]byte, requestHeaderSize+len(message.Body))
		binary.BigEndian.PutUint16(payload[:2], uint16(message.MessageID))
		binary.BigEndian.PutUint32(payload[2:6], 0)
		copy(payload[requestHeaderSize:], message.Body)
		return payload, nil
	}
	// 响应额外携带对应请求序号和业务错误码。
	payload := make([]byte, responseHeaderSize+len(message.Body))
	binary.BigEndian.PutUint16(payload[:2], uint16(message.MessageID))
	binary.BigEndian.PutUint32(payload[2:6], message.Sequence)
	binary.BigEndian.PutUint32(payload[6:10], uint32(message.ErrorCode))
	copy(payload[responseHeaderSize:], message.Body)
	return payload, nil
}
