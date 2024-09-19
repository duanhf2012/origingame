package gateservice

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/network/processor"
	"github.com/duanhf2012/origin/v2/node"
	"github.com/duanhf2012/origin/v2/service"
	"google.golang.org/protobuf/proto"
	"origingame/common/proto/msg"
	"origingame/common/proto/rpc"
	"origingame/common/util"
	"origingame/service/gateservice/tcpmodule"
	"origingame/service/gateservice/wsmodule"
	"time"
)

func init() {
	node.Setup(&GateService{})
}

type GateService struct {
	service.Service

	netModule      INetModule
	pbRawProcessor processor.PBRawProcessor
	msgRouter      MsgRouter
	rawPackInfo    processor.PBRawPackInfo
}

func (gate *GateService) OnInit() error {
	gate.msgRouter.Init(&gate.pbRawProcessor)

	gate.AddModule(&gate.msgRouter)

	iConfig := gate.GetService().GetServiceCfg()
	if iConfig == nil {
		return fmt.Errorf("%s config is error", gate.GetService().GetName())
	}

	//TcpCfg与WSCfg取其一
	mapTcpCfg := iConfig.(map[string]interface{})
	_, tcpOk := mapTcpCfg["TcpCfg"]
	if tcpOk == true {
		var tcpModule tcpmodule.TcpModule
		tcpModule.SetProcessor(&gate.pbRawProcessor)
		gate.AddModule(&tcpModule)
		gate.msgRouter.SetNetModule(&tcpModule)
		gate.netModule = &tcpModule
	} else {
		_, wsOk := mapTcpCfg["WSCfg"]
		if wsOk == true {
			var wsModule wsmodule.WSModule
			wsModule.SetProcessor(&gate.pbRawProcessor)
			gate.AddModule(&wsModule)
			gate.msgRouter.SetNetModule(&wsModule)
			gate.netModule = &wsModule
		} else {
			return errors.New("WSCfg and TcpCfg are not configured")
		}
	}

	gate.RegRawRpc(util.RawRpcMsgDispatch, gate.RawRpcDispatch)
	gate.RegRawRpc(util.RawRpcCloseClient, gate.RawCloseClient)
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

func (gate *GateService) SendMsg(clientId string, msgType uint16, rawMsg []byte) error {
	gate.rawPackInfo.SetPackInfo(msgType, rawMsg)
	bytes, err := gate.pbRawProcessor.Marshal(clientId, &gate.rawPackInfo)
	if err != nil {
		return err
	}

	err = gate.netModule.SendRawMsg(clientId, bytes)
	if err != nil {
		log.Debug("SendMsg fail ", log.ErrorAttr("err", err), log.String("clientId", clientId))
	}

	return err
}

func (gate *GateService) RawRpcDispatch(rawInput []byte) {
	var rawInputArgs rpc.RawInputArgs
	err := proto.Unmarshal(rawInput, &rawInputArgs)
	if err != nil {
		log.SError("msg is error:%s", err.Error())
		return
	}

	for _, clientId := range rawInputArgs.ClientIdList {
		err = gate.SendMsg(clientId, uint16(rawInputArgs.MsgType), rawInputArgs.RawData)
		if err != nil {
			log.SError("SendRawMsg fail:", err.Error())
		}
	}

	//消息统计
	gate.msgRouter.performanceAnalyzer.ChangeDeltaNum(MsgAnalyzer, int(rawInputArgs.MsgType), MsgSendNumColumn, int64(len(rawInputArgs.ClientIdList)))
	gate.msgRouter.performanceAnalyzer.ChangeDeltaNum(MsgAnalyzer, int(rawInputArgs.MsgType), MsgSendByteColumn, int64(len(rawInputArgs.ClientIdList)*(len(rawInputArgs.RawData)+2))) //排除掉消息ID长度
}

func (gate *GateService) RawCloseClient(rawInput []byte) {
	var rawInputArgs rpc.RawInputArgs
	err := proto.Unmarshal(rawInput, &rawInputArgs)
	if err != nil {
		log.Error("msg is error", log.ErrorAttr("err", err))
		return
	}

	for _, clientId := range rawInputArgs.ClientIdList {
		gate.netModule.Close(clientId)
	}
}
