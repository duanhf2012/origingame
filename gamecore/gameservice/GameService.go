package gameservice

import (
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/node"
	"github.com/duanhf2012/origin/v2/service"
	"origingame/common/performance"
	"origingame/common/proto/rpc"
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
}

func (gs *GameService) OnInit() error {
	gs.mapClientPlayer = make(map[string]*player.Player, 2048)

	gs.performanceAnalyzer = &performance.PerformanceAnalyzer{}
	gs.objectFactoryModule = factory.NewGameObjectFactoryModule()
	gs.objectFactoryModule.Analyzer = gs.performanceAnalyzer

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
