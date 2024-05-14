package msghandler

import (
	"google.golang.org/protobuf/proto"
	"origingame/common/proto/msg"
	"origingame/gamecore/gameservice/player"
)

type CallBack func(player *player.Player, msg proto.Message)

func RegisterMessage(register func(msgType msg.MsgType, message proto.Message, cb CallBack)) {
	register(msg.MsgType_Ping, &msg.MsgNil{}, ping)
}
