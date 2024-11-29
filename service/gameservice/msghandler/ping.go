package msghandler

import (
	"origingame/common/proto/msg"
	"origingame/service/gameservice/player"
)

func ping(player *player.Player, msg *msg.MsgNil) {
	player.Ping()
	panic("xx")
}
