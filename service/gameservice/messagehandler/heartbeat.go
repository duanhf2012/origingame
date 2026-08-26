package messagehandler

import (
	"time"

	commonpb "origingame/protocol/common"
	"origingame/service/gameservice/msgrouter"
	"origingame/service/gameservice/player"
)

func handlePlayerHeartbeat(session *msgrouter.Session, current *player.Player, _ *commonpb.PlayerHeartbeatRequest) error {
	if !current.RecordHeartbeat(session.ConnectionID(), time.Now()) {
		return nil
	}
	return session.Reply(commonpb.MessageID_Ok, nil)
}
