// Package messagehandler 集中登记当前 GameService 的全部客户端业务消息。
package messagehandler

import (
	commonpb "origingame/protocol/common"
	"origingame/service/gameservice/msgrouter"
)

// Register 逐条登记真实存在的客户端业务消息。
func Register(router *msgrouter.Router) error {
	return msgrouter.Register(router, commonpb.MessageID_PlayerHeartbeatReq, handlePlayerHeartbeat)
}
