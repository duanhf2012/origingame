package player

import (
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/util/timer"
	"google.golang.org/protobuf/proto"
	"origingame/common/collect"
	"origingame/common/proto/msg"
	"origingame/service/gameservice/dbcollection"
	"origingame/service/interfacedef"
	"time"
)

type Player struct {
	PoolObj
	interfacedef.IGSService
	dbcollection.PlayerDB

	DataInfo

	checkPingTicker uint64
}

func (p *Player) Init(id string, gsService interfacedef.IGSService) {
	p.IGSService = gsService
	p.Id = id
	p.GenSessionId()

	p.PlayerDB.OnInit(p, gsService)

	p.SafeTickerTimer(&p.checkPingTicker, time.Second*30, nil, p.checkPingTimeout)
}

func (p *Player) Reset() {
	*p = Player{}
}

func (p *Player) LoadFromDB() {
	p.PlayerDB.LoadFromDB()
}

func (p *Player) SendMsg(msgType msg.MsgType, message proto.Message) int {
	return p.GetMsgSender().SendToClient(p.GetClientId(), msgType, message)
}

func (p *Player) Destroy() {
	p.SafeCancelTimer(&p.checkPingTicker)
	p.SaveToDB(true)
	p.DestroyPlayer(p.GetId())
}

func (p *Player) SendPlayerInfo() {
	//todo 向客户端同步相关信息，需要填充
	var msgLoadFinish msg.MsgLoadFinish
	now := timer.Now()
	msgLoadFinish.SysTime = uint32(now.Hour()*3600 + now.Minute()*60 + now.Second())

	p.SendMsg(msg.MsgType_LoadFinish, &msgLoadFinish)
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
		//todo 初始化各种代理数

		//发送数据
		p.SendPlayerInfo()
		return
	}
}

// OnDelayLoadMCDBEnd 多行数据加载完成
func (p *Player) OnLoadMultiDBEnd(collectType collect.MultiCollectionType) {
}

func (p *Player) Ping() {
	p.pingTime = time.Now()

	var pong msg.MsgPong
	pong.NowTime = time.Now().Unix()
	p.SendMsg(msg.MsgType_Pong, &pong)
}

func (p *Player) checkPingTimeout(uint64, interface{}) {
	//5分钟后释放
	if time.Now().Sub(p.pingTime) > 300*time.Second {
		p.Destroy()
	}
}

func (p *Player) IsGm() bool {
	return true
}
