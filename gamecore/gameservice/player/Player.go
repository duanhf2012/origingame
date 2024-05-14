package player

import (
	"github.com/duanhf2012/origin/v2/log"
	"google.golang.org/protobuf/proto"
	"origingame/common/collect"
	"origingame/common/proto/msg"
	"origingame/gamecore/gameservice/dbcollection"
	"origingame/gamecore/interfacedef"
)

type Player struct {
	interfacedef.IGSService
	interfacedef.IMsgSender

	dbcollection.PlayerDB

	DataInfo
	PoolObj
}

func (p *Player) Init(id string, sender interfacedef.IMsgSender, gsService interfacedef.IGSService) {
	p.IMsgSender = sender
	p.IGSService = gsService
	p.Id = id
	p.GenSessionId()

	p.PlayerDB.OnInit(p, gsService)
}

func (p *Player) Reset() {
	*p = Player{}
}

func (p *Player) LoadFromDB() {
	p.PlayerDB.LoadFromDB()
}

func (p *Player) SendMsg(msgType msg.MsgType, message proto.Message) int {
	return p.IMsgSender.SendToClient(p.GetClientId(), msgType, message)
}

func (p *Player) Destroy() {
	p.DestroyPlayer(p.GetId())
}

func (p *Player) SendPlayerInfo() {
	//向客户端同步相当信息
}

// OnLoadDBEnd 单行数据加载完成
func (p *Player) OnLoadDBEnd(ok bool) {
	//1.加载失败或者对象被封，释放对象
	if ok == false {
		log.SError("player[", p.GetId(), "] load db data failed or be sealed, need to release")
		p.Destroy()
		return
	}

	//2.如果玩家是创号流程，直接返回
	if !p.GetIsInit() {
		p.SendPlayerInfo()
		return
	}
}

// OnDelayLoadMCDBEnd 多行数据加载完成
func (p *Player) OnLoadMultiDBEnd(collectType collect.MultiCollectionType) {

}
