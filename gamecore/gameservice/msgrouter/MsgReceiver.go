package msgrouter

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/duanhf2012/origin/v2/util/sync"
	"google.golang.org/protobuf/proto"
	"origingame/common/performance"
	"origingame/common/proto/msg"
	"origingame/common/proto/rpc"
	"origingame/gamecore/gameservice/player"
	"origingame/gamecore/interfacedef"
)

type IProtoMsg[T any] interface {
	*T
	proto.Message
}

type MsgHandler[T any, P IProtoMsg[T]] struct {
	call    func(p *player.Player, msg P)
	msgType msg.MsgType
}

type MsgReceiver struct {
	service.Module
	gs interfacedef.IGSService

	mapRegisterMsg map[msg.MsgType]interfacedef.IMsgHandler //消息注册
}

func (mh *MsgHandler[T, P]) GetMsgType() msg.MsgType {
	return mh.msgType
}

func (mh *MsgHandler[T, P]) Cb(p interfacedef.IPlayer, msg []byte) {
	var t T
	err := proto.Unmarshal(msg, P(&t))
	if err != nil {
		return
	}

	mh.call(p.(*player.Player), P(&t))
}

type protoMsg struct {
	ref bool
	msg proto.Message
}

func (m *protoMsg) Reset() {
	proto.Reset(m.msg)
}

func (m *protoMsg) IsRef() bool {
	return m.ref
}

func (m *protoMsg) Ref() {
	m.ref = true
}

func (m *protoMsg) UnRef() {
	m.ref = false
}

type RegMsgInfo struct {
	protoMsg *protoMsg
	msgPool  *sync.PoolEx
}

func (r *RegMsgInfo) NewMsg() *protoMsg {
	pMsg := r.msgPool.Get().(*protoMsg)
	return pMsg
}

func (r *RegMsgInfo) ReleaseMsg(msg *protoMsg) {
	r.msgPool.Put(msg)
}

func (mr *MsgReceiver) OnInit() error {
	mr.gs = mr.GetService().(interfacedef.IGSService)
	mr.mapRegisterMsg = make(map[msg.MsgType]interfacedef.IMsgHandler, 256)

	return nil
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
	msgHandler, ok := mr.mapRegisterMsg[msg.MsgType(rawInput.GetMsgType())]
	if ok == false {
		err = fmt.Errorf("close client %+v, message type %d is not  register.", clientIdList, rawInput.GetMsgType())
		log.SWarning(err.Error())
		for _, clientId := range clientIdList {
			mr.gs.CloseClient(clientId)
		}
		return
	}

	clientId := clientIdList[0]
	p := mr.gs.GetClientPlayer(clientId)
	if p == nil {
		log.SWarning("close client ", clientId, ",mapClientPlayer not exists clientId")
		mr.gs.CloseClient(clientId)
		return
	}

	msgType := msg.MsgType(rawInput.MsgType)
	if msgType != msg.MsgType_Ping && (p.IsLoadFinish() == false) {
		log.SWarning("close client ", clientId, ", Player data has not been loaded yet")
		return
	}

	an := mr.gs.GetAnalyzer(performance.MsgAnalyzer, int(msgType))
	if an != nil {
		an.StartStatisticalTime()
	}

	msgHandler.Cb(p, rawInput.RawData)

	if an != nil {
		an.EndStatisticalTimeEx(performance.MsgCostTimeAnalyzer)
	}
}

func NewHandler[T any, P IProtoMsg[T]](msgType msg.MsgType, call func(p *player.Player, msg P)) *MsgHandler[T, P] {
	var handler MsgHandler[T, P]
	handler.call = call
	handler.msgType = msgType

	return &handler
}

func (mr *MsgReceiver) RegMsgHandler(handler interfacedef.IMsgHandler) {
	mr.mapRegisterMsg[handler.GetMsgType()] = handler
}
