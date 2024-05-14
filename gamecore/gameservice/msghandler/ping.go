package msghandler

import (
	"google.golang.org/protobuf/proto"
	"origingame/gamecore/gameservice/player"
)

func ping(player *player.Player, message proto.Message) {
	player.Ping()
}
