package msghandler

import (
	"origingame/common/proto/msg"
	"origingame/gamecore/gameservice/player"
)

func ping(player *player.Player, msg *msg.MsgNil) {
	player.Ping()
}
