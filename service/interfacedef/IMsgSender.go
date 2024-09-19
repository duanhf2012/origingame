package interfacedef

import (
	"google.golang.org/protobuf/proto"
	"origingame/common/proto/msg"
)

type IMsgSender interface {
	SendToClient(clientId string, msgType msg.MsgType, message proto.Message) int
	CastToClient(clientIdList []string, msgType msg.MsgType, message proto.Message) int
	SendToPlayer(playerUserId string, msgType msg.MsgType, message proto.Message) int
	CastToPlayer(playerUserId []string, msgType msg.MsgType, message proto.Message) int
}
