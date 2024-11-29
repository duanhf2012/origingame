package msgrouter

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/goccy/go-json"
	"google.golang.org/protobuf/proto"
	"origingame/common/performance"
	"origingame/common/proto/msg"
	"origingame/common/proto/rpc"
	"origingame/service/gameservice/player"
	"origingame/service/interfacedef"
)

type IProtoMsg[T any] interface {
	*T
	proto.Message
}

type MsgHandler[T any, P IProtoMsg[T]] struct {
	call    func(p *player.Player, msg P)
	msgType msg.MsgType
}

var mapRegisterMsg = make(map[msg.MsgType]interfacedef.IMsgHandler, 256) //消息注册

type MsgReceiver struct {
	service.Module
	gs interfacedef.IGSService
}

func (mh *MsgHandler[T, P]) GetMsgType() msg.MsgType {
	return mh.msgType
}

func (mh *MsgHandler[T, P]) MsgCb(p interfacedef.IPlayer, msg []byte) {
	var t T
	err := proto.Unmarshal(msg, P(&t))
	if err != nil {
		return
	}

	mh.call(p.(*player.Player), P(&t))
}

func (mh *MsgHandler[T, P]) GmCb(p interfacedef.IPlayer, msgBody []byte) {
	var t T
	err := json.Unmarshal(msgBody, &t)
	if err != nil {
		return
	}

	mh.call(p.(*player.Player), P(&t))
}

func (mh *MsgHandler[T, P]) NewMsg() P {
	var t T
	return &t
}

func (mr *MsgReceiver) OnInit() error {
	mr.gs = mr.GetService().(interfacedef.IGSService)

	return nil
}

func (mr *MsgReceiver) GmReceiver(p interfacedef.IPlayer, msgType msg.MsgType, msgBody []byte) bool {
	msgHandler, ok := mapRegisterMsg[msgType]
	if ok == false {
		return false
	}

	msgHandler.GmCb(p, msgBody)
	return true
}

func (mr *MsgReceiver) RpcOnRecvCallBack(data []byte) {
	//解析转发过来的数据
	var rawInput rpc.RawInputArgs
	err := proto.Unmarshal(data, &rawInput)
	if err != nil {
		log.SError("RpcOnRecvCallBack Unmarshal is error:", err.Error())
		return
	}

	clientIdList := rawInput.GetClientIdList()
	if len(clientIdList) == 0 {
		//收消息只可能有一个clientid
		log.SError("RpcOnRecvCallBack receive client len[", len(clientIdList), "] > 1")
		return
	}

	if err != nil || len(clientIdList) != 1 {
		for _, clientId := range clientIdList {
			//断开clientId连接
			mr.gs.CloseClient(clientId)
		}
		log.SError("parse message is error:", err.Error())
		return
	}

	//反序列化数据
	msgHandler, ok := mapRegisterMsg[msg.MsgType(rawInput.GetMsgType())]
	if ok == false {
		err = fmt.Errorf("close client %+v, message type %d is not  register.", clientIdList, rawInput.GetMsgType())
		log.SWarn(err.Error())
		for _, clientId := range clientIdList {
			mr.gs.CloseClient(clientId)
		}
		return
	}

	clientId := clientIdList[0]
	p := mr.gs.GetClientPlayer(clientId)
	if p == nil {
		log.SWarn("close client ", clientId, ",mapClientPlayer not exists clientId")
		mr.gs.CloseClient(clientId)
		return
	}

	msgType := msg.MsgType(rawInput.MsgType)
	if msgType != msg.MsgType_Ping && (p.IsLoadFinish() == false) {
		log.SWarn("close client ", clientId, ", Player data has not been loaded yet")
		return
	}

	an := mr.gs.GetAnalyzer(performance.MsgAnalyzer, int(msgType))
	if an != nil {
		an.StartStatisticalTime()
	}

	msgHandler.MsgCb(p, rawInput.RawData)

	if an != nil {
		an.EndStatisticalTimeEx(performance.MsgCostTimeAnalyzer)
	}
}

func newHandler[T any, P IProtoMsg[T]](msgType msg.MsgType, call func(p *player.Player, msg P)) *MsgHandler[T, P] {
	var handler MsgHandler[T, P]
	handler.call = call
	handler.msgType = msgType

	return &handler
}

func RegMsgHandler[T any, P IProtoMsg[T]](msgType msg.MsgType, call func(p *player.Player, msg P)) {
	if _, ok := mapRegisterMsg[msgType]; ok {
		panic(fmt.Errorf("repeated register msg type %d", msgType))
	}
	mapRegisterMsg[msgType] = newHandler(msgType, call)
}
