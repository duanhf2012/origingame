package gameservice

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/node"
	originRpc "github.com/duanhf2012/origin/v2/rpc"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/duanhf2012/origin/v2/util/timer"
	"google.golang.org/protobuf/proto"
	"origingame/common/performance"
	"origingame/common/proto/rpc"
	"origingame/common/util"
	"origingame/service/gameservice/gm"
	_ "origingame/service/gameservice/msghandler"
	"origingame/service/gameservice/msgrouter"
	factory "origingame/service/gameservice/objectfactory"
	"origingame/service/gameservice/player"
	"origingame/service/hotloadservice"
	"origingame/service/interfacedef"
	"time"
)

func init() {
	node.SetupTemplate[GameService]()
}

type GameService struct {
	service.Service

	bCloseStatus    bool
	mapClientPlayer map[string]*player.Player //map[clientId]*Player

	objectFactoryModule *factory.ObjectFactoryModule
	performanceAnalyzer *performance.PerformanceAnalyzer
	msgSender           msgrouter.MsgSender
	msgReceiver         msgrouter.MsgReceiver
	gmModule            gm.GmModule
	balance             rpc.GameServiceBalance //负载同步变量
	tableCfgModule      hotloadservice.TableCfgModule
}

func (gs *GameService) OnInit() error {
	gs.mapClientPlayer = make(map[string]*player.Player, 2048)
	gs.initBalance()
	gs.performanceAnalyzer = &performance.PerformanceAnalyzer{}
	gs.objectFactoryModule = factory.NewGameObjectFactoryModule()
	gs.objectFactoryModule.Analyzer = gs.performanceAnalyzer

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

	_, err = gs.AddModule(&gs.msgReceiver)
	if err != nil {
		return err
	}

	_, err = gs.AddModule(&gs.gmModule)
	if err != nil {
		return err
	}

	return nil
}

func (gs *GameService) initBalance() {
	gs.balance.NodeId = node.GetNodeId()
	gs.balance.GSName = gs.GetName()
}

func (gs *GameService) OnStart() {
	gs.AfterFunc(time.Second*2, gs.asyncPlayerListTimer)
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
	rawInputBytes, err := proto.Marshal(&rawInputArgs)
	if err != nil {
		log.Error("GameService.CloseClient proto.Marshal ", log.ErrorAttr("err", err), log.String("clientId", clientId))
		return
	}
	err = gs.RawGoNode(originRpc.RpcProcessorPB, nodeId, util.RawRpcCloseClient, util.GateService, rawInputBytes)
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

// 向中心服同步GameService负载
func (gs *GameService) timerUpdateBalance(timer *timer.Ticker) {
	nodeId := util.GetMasterCenterNodeId()
	if nodeId == "" {
		log.SError("cannot find best centerservice nodeid")
		return
	}

	gs.balance.Weigh = int32(gs.objectFactoryModule.GetPlayerNum())
	err := gs.GoNode(nodeId, "CenterService.RPC_UpdateBalance", &gs.balance)
	if err != nil {
		log.SError("RPC_UpdateBalance fail ", err.Error())
	}
}

func (gs *GameService) asyncPlayerList() error {
	nodeId := util.GetMasterCenterNodeId()
	if nodeId == "" {
		return fmt.Errorf("cannot find best centerservice nodeid")
	}

	if gs.IsRetire() == true {
		log.Info("I am retire...", log.Int("playerCount", gs.objectFactoryModule.GetPlayerNum()))
		return nil
	}

	var playerList rpc.UpdatePlayerList
	playerList.NodeId = node.GetNodeId()
	playerList.UList = make([]string, 0, 255)
	playerList.GSName = gs.GetName()
	for uId := range gs.objectFactoryModule.GetAllPlayer() {
		playerList.UList = append(playerList.UList, uId)
	}

	return gs.GoNode(nodeId, "CenterService.RPC_UpdateUserList", &playerList)
}

func (gs *GameService) asyncPlayerListTimer(t *timer.Timer) {
	err := gs.asyncPlayerList()
	if err != nil {
		log.SWarning("asyncPlayerList fail:", err.Error())
		gs.AfterFunc(2*time.Second, gs.asyncPlayerListTimer)
	} else {
		gs.timerUpdateBalance(nil)
		gs.NewTicker(10*time.Second, gs.timerUpdateBalance)
	}
}

func (gs *GameService) GetClientPlayer(clientID string) interfacedef.IPlayer {
	return gs.mapClientPlayer[clientID]
}

func (gs *GameService) GetAnalyzer(analyzerType int, analyzerId int) *performance.Analyzer {
	return gs.performanceAnalyzer.GetAnalyzer(analyzerType, analyzerId)
}

func (gs *GameService) GetMsgReceiver() interfacedef.IMsgReceiver {
	return &gs.msgReceiver
}

func (gs *GameService) GetMsgSender() interfacedef.IMsgSender {
	return &gs.msgSender
}

func (gs *GameService) GetPlayerTimer() interfacedef.IPlayerTimer {
	return gs.objectFactoryModule
}
