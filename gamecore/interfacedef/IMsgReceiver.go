package interfacedef

import (
	"origingame/common/proto/msg"
)

type IMsgHandler interface {
	MsgCb(p IPlayer, msg []byte)
	GmCb(p IPlayer, msgBody []byte)
	GetMsgType() msg.MsgType
}

type IMsgReceiver interface {
	GmReceiver(p IPlayer, msgType msg.MsgType, msgBody []byte) bool
}
