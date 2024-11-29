package botservice

import (
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/network/processor"
	"google.golang.org/protobuf/proto"
	"origingame/common/proto/msg"
)

type BotAgent struct {
	bt *Bot
}

func (ba *BotAgent) Run() {
	for {
		bytes, err := ba.bt.conn.ReadMsg()
		if err != nil {
			log.Error("rclient read msg is failed", log.ErrorField("error", err))
			return
		}

		data, err := ba.bt.pbRawProcessor.Unmarshal("", bytes)
		if err != nil {
			log.Error("data error", log.ErrorField("err", err))
			return
		}

		rawPack := data.(*processor.PBRawPackInfo)
		rawPack.GetMsg()
		msgProto := NewMsgByMsgType(msg.MsgType(rawPack.GetPackType()))
		if msgProto == nil {
			//不关注的消息忽略
			log.Warn("msg type error", log.Any("msgType", rawPack.GetPackType()))
			continue
		}

		err = proto.Unmarshal(rawPack.GetMsg()[2:], msgProto.msg)
		if err != nil {
			log.Error("Unmarshal error", log.ErrorField("err", err))
			return
		}

		ba.bt.chanMsg <- msgProto
	}
}

func (ba *BotAgent) OnClose() {
	ba.bt.tcpClient = nil
	ba.bt.setGateLogging(ba.bt.token)
}
