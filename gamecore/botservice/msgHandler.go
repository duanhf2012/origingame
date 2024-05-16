package botservice

import (
	"github.com/duanhf2012/origin/v2/log"
	"google.golang.org/protobuf/proto"
	"origingame/common/proto/msg"
)

func init() {
	mapRegisterMsg = make(map[msg.MsgType]*MsgEvent)
	regMsg(msg.MsgType_LoginRes, &msg.MsgLoginRes{}, loginRes)
	regMsg(msg.MsgType_LoadFinish, &msg.MsgLoadFinish{}, loadFinish)
}

func loginRes(bot *Bot, msgProto proto.Message) {
	msgLoginRes := msgProto.(*msg.MsgLoginRes)
	log.Debug("loginRes", log.Any("loginRes", msgLoginRes))
	bot.setLoginFinish()
}

func loadFinish(bot *Bot, msgProto proto.Message) {
	msgLoadFinish := msgProto.(*msg.MsgLoadFinish)
	log.Debug("msgLoadFinish", log.Any("msgLoadFinish", msgLoadFinish))
	bot.setLoginFinish()
}
