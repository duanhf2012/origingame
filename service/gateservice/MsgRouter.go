package gateservice

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/network/processor"
	"github.com/duanhf2012/origin/v2/node"
	originRpc "github.com/duanhf2012/origin/v2/rpc"
	"github.com/duanhf2012/origin/v2/service"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/protobuf/proto"
	"origingame/common/db"
	"origingame/common/performance"
	"origingame/common/proto/msg"
	"origingame/common/proto/rpc"
	"origingame/common/util"
	"time"
)

type MsgRouter struct {
	service.Module

	netModule   INetModule
	rawPackInfo processor.PBRawPackInfo

	mapRouterCache map[string]*nodeInfo //map[clientId]nodeInfo,Client连接信息

	performanceAnalyzer *performance.PerformanceAnalyzer
}

type nodeInfo struct {
	nodeId string
	gsName string
	status StatusType
}

type StatusType int

const (
	LoginStart StatusType = 0
	Logging    StatusType = 0
	Logined    StatusType = 1
)

type INetModule interface {
	SendRawMsg(clientId string, data []byte) error
	Close(clientId string)
	GetClientIp(clientId string) string
	GetProcessor() processor.IRawProcessor
}

func (mr *MsgRouter) OnInit() error {
	mr.mapRouterCache = make(map[string]*nodeInfo, 2048)

	mapCfg := mr.GetService().GetServiceCfg().(map[string]interface{})
	analyzerLogLevel := performance.AnalyzerLogLevel1
	analyzerLevel, okLogLevel := mapCfg["PerformanceLogLevel"]
	if okLogLevel == true {
		analyzerLogLevel = int(analyzerLevel.(float64))
	}

	var analyzerInterval time.Duration
	intervalTime, okOpen := mapCfg["PerformanceIntervalTime"]
	if okOpen == true {
		analyzerInterval = time.Duration(intervalTime.(float64)) * time.Millisecond
	}

	mr.performanceAnalyzer = &performance.PerformanceAnalyzer{}
	InitPerformanceAnalyzer(mr.performanceAnalyzer, analyzerInterval, analyzerLogLevel)
	_, err := mr.AddModule(mr.performanceAnalyzer)
	if err != nil {
		log.SError("Router.OnInit AddModule err:", err.Error())
		return err
	}

	return nil
}

func (mr *MsgRouter) SetNetModule(netModule INetModule) {
	mr.netModule = netModule
}

func (mr *MsgRouter) Init(process processor.IRawProcessor) {
	process.SetConnectedHandler(mr.OnConnected)
	process.SetDisConnectedHandler(mr.OnDisconnected)
	process.SetRawMsgHandler(mr.RouterMessage)
}

func (mr *MsgRouter) OnDisconnected(clientId string) {
	log.SDebug("disconnect clientId ", clientId)
	//1.查找路由
	nodeId, gsName := mr.GetRouterId(clientId)
	if nodeId == "" || gsName == "" {
		log.SDebug("cannot find clientId ", clientId)
		return
	}

	delete(mr.mapRouterCache, clientId)
	mr.performanceAnalyzer.Set(ConnectNumAnalyzer, ClientConnectAnalyzerId, ClientConnectNumColumn, int64(mr.GetClientNum()))

	//2.转发客户端连接断开
	var rawInputArgs rpc.RawInputArgs
	rawInputArgs.ClientIdList = []string{clientId}
	rawInputBytes, err := proto.Marshal(&rawInputArgs)
	if err != nil {
		log.Error("Router.OnDisconnected proto.Marshal err", log.ErrorAttr("err", err))
		return
	}
	err = mr.RawGoNode(originRpc.RpcProcessorPB, nodeId, util.RawRpcOnClose, gsName, rawInputBytes)
	if err != nil {
		log.SError("Router.OnDisconnected RawGoNode err:", err.Error())
	}
}

func (mr *MsgRouter) OnConnected(clientId string) {
}

// RouterMessage 函数返回后，msgBuff内存将被内存池回收
func (mr *MsgRouter) RouterMessage(cliId string, msgType uint16, msgBuff []byte) {
	//1.登陆消息单独处理
	switch msg.MsgType(msgType) {
	case msg.MsgType_LoginReq:
		mr.login(cliId, msgBuff[2:])
		return
	}

	//2.通过clientId获取nodeId
	nodeId, gsName := mr.GetRouterId(cliId)
	if nodeId == "" {
		log.SWarning("cannot find clientId ", cliId)
		mr.netModule.Close(cliId)
		return
	}

	//3.组装原始Rpc参数用于转发
	var inputArgs rpc.RawInputArgs
	inputArgs.ClientIdList = []string{cliId}
	inputArgs.RawData = msgBuff[2:]
	inputArgs.MsgType = uint32(msgType)

	inputArgBytes, err := proto.Marshal(&inputArgs)
	if err != nil {
		log.Error("RouterMessage proto.Marshal err", log.ErrorAttr("err", err))
		return
	}

	//4.转发消息
	err = mr.RawGoNode(originRpc.RpcProcessorPB, nodeId, util.RawRpcOnRecv, gsName, inputArgBytes)
	if err != nil {
		//关闭连接
		mr.netModule.Close(cliId)
		log.SError("RawGoNode fail ", err.Error())
	}
}

func (mr *MsgRouter) login(cliId string, msgBuff []byte) {
	var msgLoginReq msg.MsgLoginReq
	err := proto.Unmarshal(msgBuff, &msgLoginReq)
	if err != nil {
		log.SError("LoginReq fail,Unmarshal error:", err.Error())
		mr.netModule.Close(cliId)
		return
	}

	if msgLoginReq.ShowAreaId <= 0 {
		log.SError("msgLoginReq.ShowAreaId is error:", msgLoginReq.ShowAreaId, " clientId ", cliId)
		mr.netModule.Close(cliId)
		return
	}

	if msgLoginReq.SessionId == "" {
		mr.loginToCenter(cliId, &msgLoginReq)
		return
	}

	mr.loginToGs(cliId, &msgLoginReq)
}

func (mr *MsgRouter) loginToCenter(cliId string, msgLoginReq *msg.MsgLoginReq) {
	platId, err := util.DecryptToken(msgLoginReq.Token)
	if err != nil {
		log.Error("LoginReq fail", log.String("Token", msgLoginReq.Token), log.String("clientId", cliId), log.ErrorAttr("err", err))
		//Token错误时，则需要重新登陆
		var loginRes msg.MsgLoginRes
		loginRes.Ret = msg.ErrCode_TokenError
		errSend := mr.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
		if errSend != nil {
			log.SError("Router.loginReq platId:", platId, " SendMsg err:", errSend.Error())
		}
		//r.gateService.Close(cliId)
		//超时将自动断开
		return
	} else {
		mr.loginToDB(cliId, platId, msgLoginReq.ShowAreaId, msgLoginReq)
	}
}

func (mr *MsgRouter) loginToDB(cliId string, platId string, showAreaId int32, msgLoginReq *msg.MsgLoginReq) {
	//2.生成数据库请求
	var req db.DBControllerReq
	req.CollectName = "UserInfo"
	req.Type = db.OptType_FindOneAndUpdate
	req.Condition, _ = bson.Marshal(bson.D{{"PlatId", platId}, {"ShowAreaId", showAreaId}})
	req.Key = platId
	newUserId := primitive.NewObjectID().Hex()
	upsert := bson.M{"PlatId": platId, "_id": newUserId, "ShowAreaId": showAreaId, "NickName": fmt.Sprintf("%s-%d", platId, msgLoginReq.ShowAreaId)}
	update := bson.M{"$setOnInsert": upsert}
	rawData, err := bson.Marshal(update)
	if err != nil {
		var loginRes msg.MsgLoginRes
		loginRes.Ret = msg.ErrCode_InterNalError
		errSend := mr.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
		if errSend != nil {
			log.SError("SendMsg fail:", errSend.Error(), ",platId:", platId)
		}
		log.SError("LoginToDB fail:", err.Error(), ",platId:", platId)
		return
	}
	req.SelectField, _ = bson.Marshal(bson.D{{"_id", 1}})
	req.RawData = append(req.RawData, rawData)
	//todo 如果需要判断注册用户已经满了，则这里为false
	req.Upsert = true

	//3.平台登陆验证成功，去DB创建或者查询账号
	err = mr.AsyncCall(util.AreaDBRequest, &req, func(res *db.DBControllerRet, err error) {
		//返回账号创建结果
		if err != nil || len(res.Res) == 0 {
			var loginRes msg.MsgLoginRes
			loginRes.Ret = msg.ErrCode_DBReturnError
			errSend := mr.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
			if errSend != nil {
				log.SError("SendMsg fail:", errSend.Error(), ",platId:", platId)
			}

			if err != nil {
				log.SError("logindb is fail :", err.Error(), "clientId:", cliId, ",platId:", platId, ",showAreaId:", showAreaId)
			} else {
				log.SError("logindb is fail :", "clientId:", cliId, ",platId:", platId, ",showAreaId:", showAreaId)
			}

			return
		}

		var userInfo struct {
			UserId string `bson:"_id"`
		}
		err = bson.Unmarshal(res.Res[0], &userInfo)
		if err != nil {
			var loginRes msg.MsgLoginRes
			loginRes.Ret = msg.ErrCode_DBReturnError
			errSend := mr.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
			if errSend != nil {
				log.SError("SendMsg fail:", errSend.Error(), ",platId:", platId)
			}

			log.SError("logindb is fail :", err.Error(), "clientId:", cliId, ",platId:", platId, ",showAreaId:", showAreaId)
			return
		}

		msgLoginReq.UserId = userInfo.UserId
		mr.loginToGs(cliId, msgLoginReq)
	})

	if err != nil {
		var loginRes msg.MsgLoginRes
		loginRes.Ret = msg.ErrCode_InterNalError
		errSend := mr.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
		if errSend != nil {
			log.SError("SendMsg fail:", errSend.Error(), ",platId:", platId)
		}

		log.SError("logindb is fail :", err.Error(), "clientId:", cliId, ",platId:", platId, ",showAreaId:", showAreaId)
	}
}

func (mr *MsgRouter) loginToGs(cliId string, msgLoginReq *msg.MsgLoginReq) {
	//3.选择主中心服
	masterNodeId := util.GetMasterCenterNodeId()
	if masterNodeId == "" || msgLoginReq.UserId == "" {
		log.SError("Cannot get center service service,userId:", msgLoginReq.UserId, " masterNodeId:", masterNodeId)

		var loginRes msg.MsgLoginRes
		loginRes.Ret = msg.ErrCode_InterNalError
		errSend := mr.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
		if errSend != nil {
			log.SError("SendMsg fail:", errSend.Error(), ",userId:", msgLoginReq.UserId)
		}

		//mr.tcpModule.Close(cliId)
		return
	}

	//4.从中心服验证Token
	var req rpc.LoginGateCheckReq
	req.UserId = msgLoginReq.UserId
	req.ShowAreaId = msgLoginReq.ShowAreaId
	req.ChannePlat = msgLoginReq.ChannePlat
	req.ChannelUUID = msgLoginReq.ChannelUUID
	req.ClientIp = mr.netModule.GetClientIp(cliId)

	err := mr.AsyncCallNode(masterNodeId, "CenterService.RPC_Login", &req, func(res *rpc.LoginGateCheckRet, err error) {
		var loginRes msg.MsgLoginRes
		if err != nil {
			loginRes.Ret = msg.ErrCode_InterNalError
			errSend := mr.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
			if errSend != nil {
				log.SError("Router.loginReq UserId:", req.UserId, " SendMsg err:", errSend.Error())
			}
			log.SError("AsyncCallNode CenterService.RPC_Login fail :", err.Error(), ",userId:", req.UserId)
			return
		}

		if res.Ret != 0 {
			// centerserver 那边调用正常，但是处理发生了错误，比如说白名单没过
			loginRes.Ret = msg.ErrCode(res.Ret)
			mr.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
			return
		}

		if res.NodeId == "" || res.GSName == "" {
			loginRes.Ret = msg.ErrCode_ServerFull
			errSend := mr.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
			if errSend != nil {
				log.SError("Router.loginReq UserId:", req.UserId, " SendMsg err:", errSend.Error())
			}
			log.SError("AsyncCallNode CenterService.RPC_Login server is full", ",userId:", req.UserId)
			return
		}

		if res.NodeId == "" {
			loginRes.Ret = msg.ErrCode_InterNalError
			errSend := mr.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
			if errSend != nil {
				log.SError("Router.loginReq UserId:", req.UserId, " SendMsg err:", errSend.Error())
			}
			log.SError("AsyncCallNode CenterService.RPC_Login fail cannot find node id", ",userId:", req.UserId)
			return
		}

		//验证通过
		mr.loginOk(cliId, res.NodeId, res.GSName, msgLoginReq)
	})

	//5.失败返回失败结果
	if err != nil {
		var loginRes msg.MsgLoginRes
		log.SError("AsyncCallNode CenterService.RPC_Login fail :", err.Error(), ",userId:", req.UserId)
		loginRes.Ret = msg.ErrCode_InterNalError
		errSend := mr.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
		if errSend != nil {
			log.SError("Router.loginReq UserId:", req.UserId, " SendMsg err:", errSend.Error())
		}
	}
}

func (mr *MsgRouter) SendMsg(clientId string, msgType msg.MsgType, msg proto.Message) error {
	byteMsg, err := proto.Marshal(msg)
	if err != nil {
		log.SError("SendMsg fail", log.String("clientId", clientId), log.ErrorAttr("error", err), log.String("msgType", fmt.Sprint(msgType)))
		return err
	}

	mr.rawPackInfo.SetPackInfo(uint16(msgType), byteMsg)
	bytes, err := mr.netModule.GetProcessor().Marshal(clientId, &mr.rawPackInfo)
	if err != nil {
		return err
	}

	err = mr.netModule.SendRawMsg(clientId, bytes)
	if err != nil {
		log.Error("SendMsg fail", log.String("clientId", clientId), log.ErrorAttr("error", err), log.String("msgType", fmt.Sprint(msgType)))
	}

	return err
}

func (mr *MsgRouter) loginOk(cliId string, nodeId string, gsName string, msgLoginReq *msg.MsgLoginReq) {
	if msgLoginReq == nil {
		log.SError("Router.loginOk msgLoginReq is nil")
		return
	}

	var req rpc.LoginToGameServiceReq
	req.UserId = msgLoginReq.UserId
	req.GateNodeId = node.GetNodeId()
	req.ClientId = cliId
	req.Ip = mr.netModule.GetClientIp(cliId)
	req.Os = msgLoginReq.Os
	req.SessionId = msgLoginReq.SessionId

	//同一个连接不能重入登陆
	routerNodeId, _ := mr.GetRouterId(cliId)
	if routerNodeId != "" {
		var loginRes msg.MsgLoginRes
		log.SError("status error ", routerNodeId, ",userid:", msgLoginReq.UserId)
		loginRes.Ret = msg.ErrCode_RepeatLoginReq
		errSend := mr.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
		if errSend != nil {
			log.SError("Router.loginOk UserId:", req.UserId, " SendMsg err:", errSend.Error())
		}
		return
	}

	var info nodeInfo
	info.nodeId = nodeId
	info.status = Logging
	info.gsName = gsName
	mr.AddNodeInfo(cliId, &info)

	mr.performanceAnalyzer.Set(ConnectNumAnalyzer, ClientConnectAnalyzerId, ClientConnectNumColumn, int64(len(mr.mapRouterCache)))

	//2.向选择好的GameService服发起登陆
	err := mr.AsyncCallNode(nodeId, gsName+".RPC_Login", &req, func(res *rpc.LoginToGameServiceRet, err error) {
		//登陆失败
		var loginRes msg.MsgLoginRes
		if err != nil {
			loginRes.Ret = msg.ErrCode_InterNalError
			errSend := mr.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
			if errSend != nil {
				log.SError("Router.loginOk UserId:", req.UserId, " SendMsg err:", errSend.Error())
			}

			log.SError("Router.loginOk UserId:", req.UserId, " SendMsg err:", err.Error())
			return
		}

		if res.Ret == 5 {
			loginRes.Ret = msg.ErrCode_CloseServerError
			errSend := mr.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
			if errSend != nil {
				log.SError("Router.loginOk UserId:", msgLoginReq.UserId, " SendMsg err:", errSend.Error())
			}

			log.SError("Router.loginOk ErrCode_CloseServerError UserId:", req.UserId)
		} else if res.Ret == 3 {
			loginRes.Ret = msg.ErrCode_SessionInvalid
			errSend := mr.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
			if errSend != nil {
				log.SError("Router.loginOk UserId:", msgLoginReq.UserId, " SendMsg err:", errSend.Error())
			}

			log.SError("Router.loginOk ErrCode_SessionInvalid UserId:", req.UserId)
		} else if res.Ret == 4 {
			loginRes.Ret = msg.ErrCode_LockLoginPleaseWait
			errSend := mr.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
			if errSend != nil {
				log.SError("Router.loginOk UserId:", msgLoginReq.UserId, " SendMsg err:", errSend.Error())
			}
			log.SError("Router.loginOk ErrCode_LockLoginPleaseWait UserId:", req.UserId)
		} else if res.Ret != 0 {
			loginRes.Ret = msg.ErrCode_InterNalError
			errSend := mr.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
			if errSend != nil {
				log.SError("Router.loginOk UserId:", msgLoginReq.UserId, " SendMsg err:", errSend.Error())
			}

			log.SError("Router.loginOk ErrCode_InterNalError UserId:", req.UserId)
			return
		}
	})

	if err != nil {
		log.SError("GameService.RPC_Login is error :", err.Error(), "UserId:", req.UserId)
		var loginRes msg.MsgLoginRes
		loginRes.Ret = msg.ErrCode_InterNalError
		errSend := mr.SendMsg(cliId, msg.MsgType_LoginRes, &loginRes)
		if errSend != nil {
			log.SError("Router.loginOk UserId:", msgLoginReq.UserId, " SendMsg err:", errSend.Error())
		}
	}
}

func (mr *MsgRouter) GetRouterId(clientId string) (string, string) {
	noInfo, ok := mr.mapRouterCache[clientId]
	if ok == false {
		return "", ""
	}

	return noInfo.nodeId, noInfo.gsName
}

func (mr *MsgRouter) AddNodeInfo(cliId string, info *nodeInfo) {
	mr.mapRouterCache[cliId] = info
}

func (mr *MsgRouter) GetClientNum() int {
	return len(mr.mapRouterCache)
}
