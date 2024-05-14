package gameservice

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/util/sync"
	"google.golang.org/protobuf/proto"
	"origingame/common/performance"
	"origingame/common/proto/msg"
	"origingame/common/proto/rpc"
	"origingame/gamecore/gameservice/msghandler"
)

type MsgReceiver struct {
	*GameService
	mapRegisterMsg map[msg.MsgType]*RegMsgInfo //消息注册
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
	protoMsg    *protoMsg
	msgPool     *sync.PoolEx
	msgCallBack msghandler.CallBack

	//retMsgType msg.MsgType //返回的消息ID
}

func (r *RegMsgInfo) NewMsg() *protoMsg {
	pMsg := r.msgPool.Get().(*protoMsg)
	return pMsg
}

func (r *RegMsgInfo) ReleaseMsg(msg *protoMsg) {
	r.msgPool.Put(msg)
}

func (mr *MsgReceiver) Init(gs *GameService) {
	mr.GameService = gs
	mr.mapRegisterMsg = make(map[msg.MsgType]*RegMsgInfo, 256)
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
			mr.GameService.CloseClient(clientId)
		}
		log.SError("parse message is error:", err.Error())
		return
	}

	//反序列化数据
	msgInfo, ok := mr.mapRegisterMsg[msg.MsgType(rawInput.GetMsgType())]
	if ok == false {
		err = fmt.Errorf("close client %+v, message type %d is not  register.", clientIdList, rawInput.GetMsgType())
		log.SWarning(err.Error())
		for _, clientId := range clientIdList {
			mr.CloseClient(clientId)
		}
		return
	}

	pMsg := msgInfo.NewMsg()
	err = proto.Unmarshal(rawInput.RawData, pMsg.msg)
	if err != nil {
		err = fmt.Errorf("close client %+v, message type %d Unmarshal is fail.", clientIdList, rawInput.GetMsgType())
		log.SWarning(err.Error())

		for _, clientId := range clientIdList {
			mr.GameService.CloseClient(clientId)
		}
		return
	}

	clientId := clientIdList[0]
	p, ok := mr.GameService.mapClientPlayer[clientId]
	if ok == false {
		log.SWarning("close client ", clientId, ",mapClientPlayer not exists clientId")
		mr.GameService.CloseClient(clientId)
		return
	}

	msgType := msg.MsgType(rawInput.MsgType)
	if msgType != msg.MsgType_Ping && (p.IsLoadFinish() == false) {
		log.SWarning("close client ", clientId, ", Player data has not been loaded yet")
		return
	}

	an := mr.GameService.performanceAnalyzer.GetAnalyzer(performance.MsgAnalyzer, int(msgType))
	if an != nil {
		an.StartStatisticalTime()
	}

	msgInfo.msgCallBack(p, pMsg.msg)
	msgInfo.ReleaseMsg(pMsg)
	if an != nil {
		an.EndStatisticalTimeEx(performance.MsgCostTimeAnalyzer)
	}
}

func (mr *MsgReceiver) register(msgType msg.MsgType, message proto.Message, cb msghandler.CallBack) {
	var regMsgInfo RegMsgInfo
	regMsgInfo.protoMsg = &protoMsg{}
	regMsgInfo.protoMsg.msg = message
	regMsgInfo.msgPool = sync.NewPoolEx(make(chan sync.IPoolData, 1000), func() sync.IPoolData {
		pMsg := protoMsg{}
		pMsg.msg = proto.Clone(regMsgInfo.protoMsg.msg)
		return &pMsg
	})

	regMsgInfo.msgCallBack = cb
	mr.mapRegisterMsg[msgType] = &regMsgInfo
}

func (mr *MsgReceiver) RegisterMessage() {
	msghandler.RegisterMessage(mr.register)
}
