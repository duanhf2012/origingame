package msghandler

import (
	"origingame/common/proto/msg"
	"origingame/gamecore/gameservice/msgrouter"
)

func init() {
	msgrouter.RegMsgHandler(msg.MsgType_Ping, ping)
}
