package gateservice

import (
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/network/processor"
	"github.com/duanhf2012/origin/v2/node"
	"github.com/duanhf2012/origin/v2/service"
	"origingame/common/proto/msg"
	"origingame/common/proto/rpc"
	"origingame/gamecore/gateservice/tcpmodule"
	"time"
)

func init() {
	node.Setup(&GateService{})
}

type GateService struct {
	service.Service

	tcpModule      tcpmodule.TcpModule
	pbRawProcessor processor.PBRawProcessor
	msgRouter      MsgRouter
}

func (gate *GateService) OnInit() error {
	gate.msgRouter.Init(&gate.pbRawProcessor, &gate.tcpModule)
	gate.tcpModule.SetProcessor(&gate.pbRawProcessor)
	gate.AddModule(&gate.tcpModule)

	return nil
}

func (gate *GateService) RPC_GSLoginRet(arg *rpc.GsLoginResult, ret *rpc.PlaceHolders) error {
	v, ok := gate.msgRouter.mapRouterCache[arg.ClientId]
	if ok == false {
		log.SWarning("Client is close cancel login ")
		return nil
	}

	v.status = Logined
	gate.msgRouter.mapRouterCache[arg.ClientId] = v

	var loginRes msg.MsgLoginRes
	loginRes.SessionId = arg.SessionId
	loginRes.UserId = arg.UserId
	loginRes.ServerTime = time.Now().UnixMilli()
	err := gate.msgRouter.SendMsg(arg.ClientId, msg.MsgType_LoginRes, &loginRes)
	if err != nil {
		log.SError("GateService.RPC_GSLoginRet ClientId:", arg.ClientId, " SendMsg err:", err.Error())
	}

	return nil
}
