package msgrouter

import (
	"google.golang.org/protobuf/proto"
	commonpb "origingame/protocol/common"
	"origingame/service/gameservice/player"
)

// Session 仅描述当前入站请求，不拥有 Gateway 连接关系。
type Session struct {
	player       *player.Player
	connectionID string
	sequence     uint32
}

// ConnectionID 返回本次请求来源，用于跨 Await 后的连接校验。
func (session *Session) ConnectionID() string { return session.connectionID }

// Reply 回复当前请求。
func (session *Session) Reply(messageID commonpb.MessageID, body proto.Message) error {
	return session.player.Reply(session.connectionID, session.sequence, messageID, body)
}

// ReplyError 回复当前请求的业务错误。
func (session *Session) ReplyError(messageID commonpb.MessageID, code commonpb.ErrorCode) error {
	return session.player.ReplyError(session.connectionID, session.sequence, messageID, code)
}
