package msghandler

import (
	"origingame/common/proto/msg"
	"origingame/service/gameservice/msgrouter"
)

func init() {
	msgrouter.RegMsgHandler(msg.MsgType_Ping, ping)
}
