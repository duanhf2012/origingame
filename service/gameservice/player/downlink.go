package player

import (
	"errors"

	"google.golang.org/protobuf/proto"
	commonpb "origingame/protocol/common"
	rpcapi "origingame/protocol/rpc"
)

// SendMsg 向当前连接发送 Sequence=0 的主动推送。
func (player *Player) SendMsg(messageID commonpb.MessageID, body proto.Message) error {
	return player.send(player.dataInfo.GatewayConnectionID, 0, messageID, commonpb.ErrorCode_ERROR_CODE_OK, body)
}

// Reply 回复指定连接上的客户端请求；连接已替换时静默忽略。
func (player *Player) Reply(
	requestConnectionID string,
	sequence uint32,
	messageID commonpb.MessageID,
	body proto.Message,
) error {
	return player.send(requestConnectionID, sequence, messageID, commonpb.ErrorCode_ERROR_CODE_OK, body)
}

// ReplyError 回复业务错误且不携带 Body；连接已替换时静默忽略。
func (player *Player) ReplyError(
	requestConnectionID string,
	sequence uint32,
	messageID commonpb.MessageID,
	code commonpb.ErrorCode,
) error {
	return player.send(requestConnectionID, sequence, messageID, code, nil)
}

func (player *Player) send(
	requestConnectionID string,
	sequence uint32,
	messageID commonpb.MessageID,
	code commonpb.ErrorCode,
	body proto.Message,
) error {
	if player == nil || requestConnectionID == "" ||
		player.dataInfo.State != StateOnline || player.dataInfo.GatewayConnectionID != requestConnectionID {
		return nil
	}
	if player.sendGateway == nil || player.dataInfo.GatewayNodeID == "" {
		return errors.New("Player Gateway 下行未初始化")
	}
	var payload []byte
	var err error
	if body != nil {
		payload, err = proto.Marshal(body)
		if err != nil {
			return err
		}
	}
	return player.sendGateway(player.dataInfo.GatewayNodeID, rpcapi.SendClientMessageRequest{
		GatewayConnectionID: player.dataInfo.GatewayConnectionID,
		Message: rpcapi.ClientMessage{
			MessageID: messageID, Sequence: sequence, ErrorCode: code, Body: payload,
		},
	})
}
