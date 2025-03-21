package gateservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/network"
	"github.com/duanhf2012/origin/v2/network/processor"
	"github.com/duanhf2012/origin/v2/node"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/duanhf2012/origin/v2/sysmodule/netmodule/kcpmodule"
	"github.com/duanhf2012/origin/v2/sysmodule/netmodule/tcpmodule"
	"github.com/duanhf2012/origin/v2/sysmodule/netmodule/wsmodule"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"origingame/common/proto/msg"
	"origingame/common/proto/rpc"
	"origingame/common/util"
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

func (gate *GateService) readWSCfg() (*wsmodule.WSCfg, error) {
	//1解析配置
	iConfig := gate.GetService().GetServiceCfg()
	if iConfig == nil {
		return nil, fmt.Errorf("%s config is error", gate.GetService().GetName())
	}

	mapWSCfg := iConfig.(map[string]interface{})
	sWSCfg, ok := mapWSCfg["WSCfg"]
	if ok == false {
		return nil, fmt.Errorf("%s.WSCfg config is error", gate.GetService().GetName())
	}

	var wsCfg wsmodule.WSCfg
	byteWSCfg, err := json.Marshal(sWSCfg)
	if err != nil {
		return nil, fmt.Errorf("%s.WSCfg config is error:%s", gate.GetService().GetName(), err.Error())
	}
	err = json.Unmarshal(byteWSCfg, &wsCfg)
	if err != nil {
		return nil, fmt.Errorf("%s.WSCfg config is error:%s", gate.GetService().GetName(), err.Error())
	}

	return &wsCfg, nil
}

func (gate *GateService) readTcpCfg() (*tcpmodule.TcpCfg, error) {
	//1解析配置
	iConfig := gate.GetService().GetServiceCfg()
	if iConfig == nil {
		return nil, fmt.Errorf("%s config is error", gate.GetService().GetName())
	}
	mapTcpCfg := iConfig.(map[string]interface{})
	iTcpCfg, ok := mapTcpCfg["TcpCfg"]
	if ok == false {
		return nil, fmt.Errorf("%s.TcpCfg config is error", gate.GetService().GetName())
	}

	var tcpCfg tcpmodule.TcpCfg
	byteTcpCfg, err := json.Marshal(iTcpCfg)
	if err != nil {
		return nil, fmt.Errorf("%s.TcpCfg config is error:%s", gate.GetService().GetName(), err.Error())
	}
	err = json.Unmarshal(byteTcpCfg, &tcpCfg)
	if err != nil {
		return nil, fmt.Errorf("%s.TcpCfg config is error:%s", gate.GetService().GetName(), err.Error())
	}

	return &tcpCfg, nil
}

func (gate *GateService) readKcpCfg() (*network.KcpCfg, error) {
	//1解析配置
	iConfig := gate.GetService().GetServiceCfg()
	if iConfig == nil {
		return nil, fmt.Errorf("%s config is error", gate.GetService().GetName())
	}

	mapKcpCfg := iConfig.(map[string]interface{})
	iKcpCfg, ok := mapKcpCfg["KcpCfg"]
	if ok == false {
		return nil, fmt.Errorf("%s.TcpCfg config is error", gate.GetService().GetName())
	}

	byteKcpCfg, err := json.Marshal(iKcpCfg)
	if err != nil {
		return nil, fmt.Errorf("%s.TcpCfg config is error:%s", gate.GetService().GetName(), err.Error())
	}
	var kcpCfg network.KcpCfg
	err = json.Unmarshal(byteKcpCfg, &kcpCfg)
	if err != nil {
		return nil, fmt.Errorf("%s.TcpCfg config is error:%s", gate.GetService().GetName(), err.Error())
	}

	return &kcpCfg, nil
}

func (gate *GateService) OnInit() error {
	gate.msgRouter.Init(&gate.pbRawProcessor)

	gate.AddModule(&gate.msgRouter)

	iConfig := gate.GetService().GetServiceCfg()
	if iConfig == nil {
		return fmt.Errorf("%s config is error", gate.GetService().GetName())
	}

	//TcpCfg与KcpCfg与WSCfg取其一
	mapTcpCfg := iConfig.(map[string]interface{})
	_, tcpOk := mapTcpCfg["TcpCfg"]
	if tcpOk == true {
		var tcpModule tcpmodule.TcpModule
		tcpCfg, err := gate.readTcpCfg()
		if err != nil {
			return err
		}
		tcpModule.Init(tcpCfg, &gate.pbRawProcessor)

		gate.AddModule(&tcpModule)
		gate.msgRouter.SetNetModule(&tcpModule)
		gate.netModule = &tcpModule
	} else if _, kcpOK := mapTcpCfg["KcpCfg"]; kcpOK == true {
		var kcpModule kcpmodule.KcpModule
		kcpCfg, err := gate.readKcpCfg()
		if err != nil {
			return err
		}

		kcpModule.Init(kcpCfg, &gate.pbRawProcessor)
		gate.AddModule(&kcpModule)
		gate.msgRouter.SetNetModule(&kcpModule)
		gate.netModule = &kcpModule
	} else {
		_, wsOk := mapTcpCfg["WSCfg"]
		if wsOk == true {
			wsCfg, err := gate.readWSCfg()
			if err != nil {
				return err
			}
			var wsModule wsmodule.WSModule
			wsModule.Init(wsCfg, &gate.pbRawProcessor)
			wsModule.SetMessageType(websocket.BinaryMessage)
			gate.AddModule(&wsModule)
			gate.msgRouter.SetNetModule(&wsModule)
			gate.netModule = &wsModule
		} else {
			return errors.New("WSCfg and TcpCfg are not configured")
		}
	}

	gate.RegRawRpc(util.RawRpcMsgDispatch, gate.RawRpcDispatch)
	gate.RegRawRpc(util.RawRpcCloseClient, gate.RawCloseClient)

	return gate.netModule.Start()
}

func (gate *GateService) RPC_GSLoginRet(arg *rpc.GsLoginResult, ret *rpc.PlaceHolders) error {
	v, ok := gate.msgRouter.mapRouterCache[arg.ClientId]
	if ok == false {
		log.SWarn("Client is close cancel login ")
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
		log.Debug("SendMsg fail ", log.ErrorField("err", err), log.String("clientId", clientId))
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
		log.Error("msg is error", log.ErrorField("err", err))
		return
	}

	for _, clientId := range rawInputArgs.ClientIdList {
		gate.netModule.Close(clientId)
	}
}
