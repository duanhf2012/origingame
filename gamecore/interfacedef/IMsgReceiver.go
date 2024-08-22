package interfacedef

import (
	"origingame/common/proto/msg"
)

type IMsgHandler interface {
	Cb(p IPlayer, msg []byte)
	GetMsgType() msg.MsgType
}

type IMsgRegister interface {
}
