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
	regMsg(msg.MsgType_Pong, &msg.MsgPong{}, pong)
}

func loginRes(bot *Bot, msgProto proto.Message) {
	msgLoginRes := msgProto.(*msg.MsgLoginRes)
	log.Debug("loginRes", log.Any("loginRes", msgLoginRes))
	bot.setLoginFinish()
	
	// 发送gm
	//var msgGm msg.MsgGmReq
	//msgGm.Command = "TestMsg"
	//msgGm.Param = []string{"100", "{}"}
	//bot.SendMsg(msg.MsgType_GM, &msgGm)
}

func loadFinish(bot *Bot, msgProto proto.Message) {
	msgLoadFinish := msgProto.(*msg.MsgLoadFinish)
	log.Debug("msgLoadFinish", log.Any("msgLoadFinish", msgLoadFinish))
	bot.setLoginFinish()
}

func pong(bot *Bot, msgProto proto.Message) {
	msgPong := msgProto.(*msg.MsgPong)
	log.Debug("msgPong", log.Any("msgPong", msgPong))
	bot.setLoginFinish()
}
