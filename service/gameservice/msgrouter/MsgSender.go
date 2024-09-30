package msgrouter

import (
	"fmt"
	"github.com/duanhf2012/origin/v2/log"
	originRpc "github.com/duanhf2012/origin/v2/rpc"
	"github.com/duanhf2012/origin/v2/service"
	"google.golang.org/protobuf/proto"
	"origingame/common/proto/msg"
	"origingame/common/proto/rpc"
	"origingame/common/util"
	"origingame/service/interfacedef"
)

type MsgSender struct {
	service.Module
	gs interfacedef.IGSService

	buff ProtoBuff
}

type ProtoBuff struct {
	buff []byte
}

func (pb *ProtoBuff) Reset() {
	pb.buff = pb.buff[:0]
}

func (pb *ProtoBuff) Marshal(message proto.Message) ([]byte, error) {
	var err error
	pb.buff, err = proto.MarshalOptions{}.MarshalAppend(pb.buff, message)
	return pb.buff, err
}

func (ms *MsgSender) OnInit() error {
	ms.gs = ms.GetService().(interfacedef.IGSService)

	return nil
}

func (ms *MsgSender) SendToClient(clientId string, msgType msg.MsgType, message proto.Message) int {
	if clientId == "" {
		return 0
	}

	return ms.send([]string{clientId}, msgType, message, nil)
}

func (ms *MsgSender) send(clients []string, msgType msg.MsgType, message proto.Message, bytesRawMsg []byte) int {
	var err error
	mapNodeClient := make(map[string][]string, 2)
	succNum := 0

	//1.分析发送的Node
	for _, clientId := range clients {
		nodeId := ms.gs.GetGateNodeIdByClientId(clientId)
		if nodeId == "" {
			log.SError("GetGateNodeIdByClientId fail", log.String("clientID", clientId))
			continue
		}

		if _, ok := mapNodeClient[nodeId]; ok == false {
			mapNodeClient[nodeId] = []string{clientId}
		} else {
			mapNodeClient[nodeId] = append(mapNodeClient[nodeId], clientId)
		}
	}

	//2.组装返回消息
	var rawBytes []byte
	if message != nil {
		ms.buff.Reset()
		rawBytes, err = ms.buff.Marshal(message)
		if err != nil {
			log.SError("Marshal fail,msgType ", msgType, " clientId list err:.", err.Error())
			return succNum
		}
	} else if bytesRawMsg != nil {
		rawBytes = bytesRawMsg
	}

	//3.发送
	for nodeId, clientList := range mapNodeClient {
		var rawInputArgs rpc.RawInputArgs
		rawInputArgs.MsgType = uint32(msgType)
		rawInputArgs.RawData = rawBytes
		rawInputArgs.ClientIdList = clientList

		var rawInputBytes []byte
		rawInputBytes, err = proto.Marshal(&rawInputArgs)
		if err != nil {
			log.Error("Marshal fail", log.Uint32("msgType", rawInputArgs.MsgType), log.ErrorAttr("err", err))
			continue
		}

		err = ms.RawGoNode(originRpc.RpcProcessorPB, nodeId, util.RawRpcMsgDispatch, util.GateService, rawInputBytes)
		if err != nil {
			log.Error(fmt.Sprint("RawGoNode fail :", err.Error(), ",msgType ", int32(msgType), " clientId list err:", err.Error()))
			continue
		}

		succNum += len(clientList)
	}

	return succNum
}

// CastToClient 广播消息
func (ms *MsgSender) CastToClient(clientIdList []string, msgType msg.MsgType, message proto.Message) int {
	return ms.send(clientIdList, msgType, message, nil)
}

func (ms *MsgSender) SendToPlayer(playerUserId string, msgType msg.MsgType, message proto.Message) int {
	//1.获取玩家clientid
	clientId := ms.gs.GetClientIdByPlayerId(playerUserId)
	if clientId == "" {
		log.SDebug("player[", playerUserId, "] is disconnected")
		return 0
	}

	return ms.send([]string{clientId}, msgType, message, nil)
}

func (ms *MsgSender) CastToPlayer(playerUserId []string, msgType msg.MsgType, message proto.Message) int {
	if len(playerUserId) == 0 {
		return 0
	}

	cIdList := make([]string, 0, 16)
	for _, playerId := range playerUserId {
		clientId := ms.gs.GetClientIdByPlayerId(playerId)
		if clientId == "" {
			continue
		}
		cIdList = append(cIdList, clientId)
	}

	return ms.send(cIdList, msgType, message, nil)
}
