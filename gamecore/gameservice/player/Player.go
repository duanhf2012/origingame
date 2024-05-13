package player

import (
	"google.golang.org/protobuf/proto"
	"origingame/common/proto/msg"
	"origingame/gamecore/gameservice/dbcollection"
	"origingame/gamecore/interfacedef"
)

type Player struct {
	dbcollection.PlayerDB
	interfacedef.IMsgSender

	DataInfo
	PoolObj
}

func (p *Player) Init(id string, sender interfacedef.IMsgSender) {
	p.IMsgSender = sender
	p.Id = id
	p.GenSessionId()
}

func (p *Player) LoadFromDB() {
	p.PlayerDB.LoadFromDB()
}

func (p *Player) SendMsg(msgType msg.MsgType, message proto.Message) int {
	return p.IMsgSender.SendToClient(p.GetClientId(), msgType, message)
}
