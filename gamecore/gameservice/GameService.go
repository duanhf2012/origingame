package gameservice

import (
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/node"
	originRpc "github.com/duanhf2012/origin/v2/rpc"
	"github.com/duanhf2012/origin/v2/service"
	"google.golang.org/protobuf/proto"
	"origingame/common/performance"
	"origingame/common/proto/rpc"
	"origingame/common/util"
	factory "origingame/gamecore/gameservice/objectfactory"
	"origingame/gamecore/gameservice/player"
)

func init() {
	node.SetupTemplate(func() service.IService {
		return &GameService{}
	})
}

type GameService struct {
	service.Service

	bCloseStatus    bool
	mapClientPlayer map[string]*player.Player //map[clientId]*Player

	objectFactoryModule *factory.ObjectFactoryModule
	performanceAnalyzer *performance.PerformanceAnalyzer
	msgSender           MsgSender
	msgReceiver         MsgReceiver
}

func (gs *GameService) OnInit() error {
	gs.mapClientPlayer = make(map[string]*player.Player, 2048)

	gs.performanceAnalyzer = &performance.PerformanceAnalyzer{}
	gs.objectFactoryModule = factory.NewGameObjectFactoryModule()
	gs.objectFactoryModule.Analyzer = gs.performanceAnalyzer
	gs.msgReceiver.Init(gs)

	gs.msgReceiver.RegisterMessage()
	gs.RegRawRpc(util.RawRpcOnRecv, gs.msgReceiver.RpcOnRecvCallBack)
	gs.RegRawRpc(util.RawRpcOnClose, gs.RpcOnCloseCallBack)

	_, err := gs.AddModule(gs.objectFactoryModule)
	if err != nil {
		return err
	}

	_, err = gs.AddModule(&gs.msgSender)
	if err != nil {
		return err
	}

	return nil
}

func (gs *GameService) OnStart() {
}

func (gs *GameService) OnRetire() {

}

func (gs *GameService) OnRelease() {
}

func (gs *GameService) attachConn(req *rpc.LoginToGameServiceReq, player *player.Player) {
	if req == nil {
		log.SError("GameService.ResetConn req is nil")
		return
	}

	if player.GetClientId() != "" {
		delete(gs.mapClientPlayer, player.GetClientId())
	}

	gs.mapClientPlayer[req.ClientId] = player
	player.AttachConn(req.GateNodeId, req.ClientId, req.SessionId, req.Ip, req.Os)

	//IP去掉端口

	gs.performanceAnalyzer.Set(performance.GameServerAnalyzer, performance.GameServerPlayerStatic, performance.GameServerClientPlayerNumColumn, int64(len(gs.mapClientPlayer)))
}

func (gs *GameService) GetGateNodeIdByClientId(clientId string) string {
	p := gs.mapClientPlayer[clientId]
	if p == nil {
		return ""
	}

	return p.GetGateNodeId()
}

func (gs *GameService) GetClientIdByPlayerId(playerId string) string {
	p := gs.objectFactoryModule.GetPlayer(playerId)
	if p == nil {
		return ""
	}

	return p.GetClientId()
}

func (gs *GameService) delClientPlayer(clientId string) {
	delete(gs.mapClientPlayer, clientId)
}

func (gs *GameService) RpcOnCloseCallBack(rawInput []byte) {
	var rawInputArgs rpc.RawInputArgs
	err := proto.Unmarshal(rawInput, &rawInputArgs)
	if err != nil {
		log.Error("Unmarshal fail", log.ErrorAttr("err", err))
		return
	}

	for _, clientId := range rawInputArgs.ClientIdList {
		gs.BeCloseClient(clientId)
	}

	gs.performanceAnalyzer.Set(performance.GameServerAnalyzer, performance.GameServerPlayerStatic, performance.GameServerPlayerNumColumn, int64(gs.objectFactoryModule.GetPlayerNum()))
	gs.performanceAnalyzer.Set(performance.GameServerAnalyzer, performance.GameServerPlayerStatic, performance.GameServerClientPlayerNumColumn, int64(len(gs.mapClientPlayer)))
}

// BeCloseClient 被关闭
func (gs *GameService) BeCloseClient(clientId string) {
	if clientId == "" {
		log.Error("clientId is empty")
		return
	}
	gs.closeClient(clientId)
}

func (gs *GameService) closeClient(clientId string) {
	//查找Player对象
	p, ok := gs.mapClientPlayer[clientId]
	if ok == false {
		log.SError("clientId ", clientId, " not found")
		return
	}

	//设置为离线状态
	p.SetOnline(false)
	gs.delClientPlayer(clientId)

	gs.performanceAnalyzer.Set(performance.GameServerAnalyzer, performance.GameServerPlayerStatic, performance.GameServerPlayerNumColumn, int64(gs.objectFactoryModule.GetPlayerNum()))
	gs.performanceAnalyzer.Set(performance.GameServerAnalyzer, performance.GameServerPlayerStatic, performance.GameServerClientPlayerNumColumn, int64(len(gs.mapClientPlayer)))
}

// CloseClient 主动关闭连接
func (gs *GameService) CloseClient(clientId string) {
	if clientId == "" {
		log.Error("clientId is empty")
		return
	}

	log.SWarning("GameService CloseClient:", clientId)
	nodeId := gs.GetGateNodeIdByClientId(clientId)
	var rawInputArgs rpc.RawInputArgs
	rawInputArgs.ClientIdList = []string{clientId}
	err := gs.RawGoNode(originRpc.RpcProcessorPB, nodeId, util.RawRpcCloseClient, util.GateService, rawInputArgs.GetRawData())
	if err != nil {
		log.SError("GameService.CloseClient RawGoNode err:", err.Error())
	}

	gs.closeClient(clientId)
}

func (gs *GameService) DestroyPlayer(playerId string) bool {
	p := gs.objectFactoryModule.GetPlayer(playerId)
	if p == nil {
		log.Error("cannot find player", log.String("playerId", playerId))
		return false
	}

	clientId := p.GetClientId()
	if clientId != util.IsEmpty {
		gs.CloseClient(clientId)
	}
	gs.objectFactoryModule.ReleasePlayer(p)

	return true
}
