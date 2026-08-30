// Package msgrouter 实现 GameService 实例级的客户端消息静态路由。
package msgrouter

import (
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
	commonpb "origingame/protocol/common"
	"origingame/service/gameservice/player"
)

type routeHandler func(*Session, *player.Player, []byte) error

// Router 保存单个 GameService 实例冻结后的 MessageID 路由表。
type Router struct {
	routes map[commonpb.MessageID]routeHandler
	frozen bool
}

// NewRouter 创建尚未登记和冻结的实例路由表。
func NewRouter() *Router { return &Router{routes: make(map[commonpb.MessageID]routeHandler)} }

// Register 使用具体 Protobuf 值类型登记无反射的解码与业务 Handler。
// T 必须是生成消息的值类型，Handler 参数固定使用 *T。
func Register[T any](router *Router, messageID commonpb.MessageID, handler func(*Session, *player.Player, *T) error) error {
	if router == nil || router.frozen || messageID == commonpb.MessageID_Ok || handler == nil {
		return errors.New("消息路由登记参数无效")
	}
	if _, exists := router.routes[messageID]; exists {
		return fmt.Errorf("MessageID %d 重复登记", messageID)
	}
	if _, ok := any(new(T)).(proto.Message); !ok {
		return errors.New("消息请求类型必须是生成的 Protobuf 消息")
	}
	router.routes[messageID] = func(session *Session, current *player.Player, body []byte) error {
		request := new(T)
		message := any(request).(proto.Message)
		if err := proto.Unmarshal(body, message); err != nil {
			return err
		}
		return handler(session, current, request)
	}
	return nil
}

// Freeze 冻结路由表；之后禁止继续登记。
func (router *Router) Freeze() error {
	if router == nil || router.frozen || len(router.routes) == 0 {
		return errors.New("消息路由无法冻结")
	}
	router.frozen = true
	return nil
}

// Dispatch 解码并执行一条已经校验当前连接的玩家消息。
func (router *Router) Dispatch(
	current *player.Player,
	connectionID string,
	messageID commonpb.MessageID,
	sequence uint32,
	body []byte,
) error {
	if router == nil || !router.frozen || current == nil || connectionID == "" || sequence == 0 {
		return errors.New("玩家消息路由参数无效")
	}
	handler := router.routes[messageID]
	if handler == nil {
		return fmt.Errorf("MessageID %d 未登记", messageID)
	}
	session := &Session{player: current, connectionID: connectionID, sequence: sequence}
	return handler(session, current, body)
}
