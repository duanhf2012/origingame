package msghandler

import (
	"origingame/common/proto/msg"
	"origingame/gamecore/gameservice/msgrouter"
)

type MsgHandle struct {
	*msgrouter.MsgReceiver
}

func (mh *MsgHandle) Init(mr *msgrouter.MsgReceiver) {
	mh.MsgReceiver = mr
	mh.RegMgsHandler()
}

func (mh *MsgHandle) RegMgsHandler() {
	mh.RegMsgHandler(msgrouter.NewHandler(msg.MsgType_Ping, ping))
}
