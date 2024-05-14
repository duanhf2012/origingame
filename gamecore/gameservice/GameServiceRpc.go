package gameservice

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/node"
	"origingame/common/proto/msg"
	"origingame/common/proto/rpc"
	"origingame/common/util"
)

func (gs *GameService) RPC_Login(req *rpc.LoginToGameServiceReq, res *rpc.LoginToGameServiceRet) error {
	//关服状态不让登陆
	if gs.bCloseStatus == true {
		res.Ret = 5
		return nil
	}

	//1.查找Player对象
	res.Ret = 0
	log.SDebug("GS", gs.GetModuleId(), ":login player:", req.UserId)

	//验证Session
	p := gs.objectFactoryModule.GetPlayer(req.UserId)
	//3.创建或初始化玩家
	if p == nil {
		//没有玩家对象，不能重登陆，只能登陆重连
		if req.SessionId != "" {
			res.Ret = 3
			log.SError("Player need to login :", req.UserId)
			return nil
		}
	} else {
		//如果已经有玩家对象，对比Session是否一致，否则只能走登陆流程
		if req.SessionId != "" && req.SessionId != p.DataInfo.SessionId {
			res.Ret = 3
			log.SError("Player need to login :", req.UserId, " session ", req.SessionId, "!=", p.DataInfo.SessionId)
			return nil
		}
	}

	//2.先同步Player登陆状态
	var playerStatus rpc.UpdatePlayerStatus
	playerStatus.UserId = req.UserId
	playerStatus.Status = rpc.LoginStatus_Logined
	playerStatus.NodeId = node.GetNodeId()
	playerStatus.GSName = gs.GetName()

	masterNodeId := util.GetMasterCenterNodeId()
	if masterNodeId == "" {
		res.Ret = 1
		log.SError("getBestMasterNodeId is fail")
		return nil
	}
	err := gs.GoNode(masterNodeId, "CenterService.RPC_UpdateStatus", &playerStatus)
	if err != nil {
		res.Ret = 2
		log.SError("go CenterService.RPC_UpdateStatus fail ", err.Error())
		return nil
	}

	//3.创建或初始化玩家
	if p == nil {
		p = gs.objectFactoryModule.NewPlayer(req.UserId, &gs.msgSender, gs)
		gs.attachConn(req, p)
		p.LoadFromDB()
		res.SessionId = p.GetSessionId()
	} else {
		//关闭老连接->改为不关闭，发送消息给客户端
		log.SDebug("close client ", p.GetClientId(), ",player reLogin:", p.GetId())
		if p.GetClientId() != "" {
			p.SendMsg(msg.MsgType_NotifyLogout, &msg.MsgNotifyLogout{Reason: int32(msg.LogoutType_Occupy)})
			delete(gs.mapClientPlayer, p.GetClientId())
		}
		//重新关联新连接
		gs.attachConn(req, p)
	}

	// 网关发送登陆成功
	var arg rpc.GsLoginResult
	arg.SessionId = p.GetSessionId()
	arg.ClientId = req.ClientId
	arg.UserId = req.UserId

	err = gs.GoNode(req.GateNodeId, "GateService.RPC_GSLoginRet", &arg)
	if err != nil {
		log.Error("GoNode error", log.String("GateNodeId", req.GetGateNodeId()), log.ErrorAttr("err", err))
		return fmt.Errorf("GoNode error %s", err.Error())
	}

	return nil
}
