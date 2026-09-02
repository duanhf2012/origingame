package virtualplayer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/duanhf2012/origin/v3/sysmodule/network"
	commonpb "origingame/protocol/common"
)

// State 是单个机器人客户端会话状态。
type State uint8

const (
	StateCreated State = iota + 1
	StateAuthenticated
	StateConnected
	StateOnline
	StateFailed
	StateClosed
)

// ScheduleFunc 创建由scenario.Module持有、回调已投递回Service的真实时间任务。
type ScheduleFunc func(time.Duration, func()) (cancel func(), err error)

// Response 是一个确定完成的请求响应或基础设施错误。
type Response struct {
	MessageID commonpb.MessageID // 响应消息标识。
	ErrorCode commonpb.ErrorCode // 响应业务错误码。
	Body      []byte             // 未解码业务负载。
	Err       error              // 基础设施或超时错误。
}

type pendingRequest struct {
	sequence          uint32             // 请求序号。
	expectedMessageID commonpb.MessageID // 预期响应消息标识。
	complete          func(Response)     // 唯一完成回调。
	cancelTimeout     func()             // 超时任务取消函数。
}

type heartbeatLoop struct {
	interval   time.Duration // 心跳发送间隔。
	timeout    time.Duration // 单次心跳超时。
	cancelNext func()        // 下一次心跳任务取消函数。
	onFailure  func(error)   // 心跳失败回调。
}

// Player 持有一个机器人的Token、Gateway Session、业务pending和后台心跳。
// 所有状态方法必须在RobotService串行上下文调用；Session.Send和Close自身并发安全。
type Player struct {
	id               int64                // 稳定机器人标识。
	state            State                // 当前客户端状态。
	token            string               // 登录取得的敏感游戏 Token。
	gatewayAddress   string               // 目标 Gateway 地址。
	session          network.Session      // 当前底层网络会话。
	nextSequence     uint32               // 下一个客户端请求序号。
	businessPending  *pendingRequest      // 当前业务请求 Pending。
	heartbeatPending *pendingRequest      // 当前心跳请求 Pending。
	heartbeat        *heartbeatLoop       // 运行中的后台心跳。
	schedule         ScheduleFunc         // 场景 Module 所有的 Timer 调度器。
	onPush           func(InboundMessage) // 主动推送回调。
	failure          error                // 终态失败原因。
}

// NewPlayer 创建尚未登录的机器人。
func NewPlayer(id int64, schedule ScheduleFunc, onPush func(InboundMessage)) (*Player, error) {
	if id <= 0 || schedule == nil {
		return nil, errors.New("机器人ID或Timer调度器无效")
	}
	return &Player{id: id, state: StateCreated, schedule: schedule, onPush: onPush}, nil
}

func (player *Player) ID() int64              { return player.id }
func (player *Player) State() State           { return player.state }
func (player *Player) Failure() error         { return player.failure }
func (player *Player) GatewayAddress() string { return player.gatewayAddress }

// ApplyLogin 保存敏感凭证但不向蓝图返回。
func (player *Player) ApplyLogin(result LoginResult) error {
	if player == nil || player.state != StateCreated || result.Token == "" || result.GatewayAddress == "" {
		return errors.New("机器人登录状态或结果无效")
	}
	player.token = result.Token
	player.gatewayAddress = result.GatewayAddress
	player.state = StateAuthenticated
	return nil
}

// BindSession 绑定一次性Dialer返回的活动Session。
func (player *Player) BindSession(session network.Session) error {
	if player == nil || player.state != StateAuthenticated || session == nil || session.Context().Err() != nil {
		return errors.New("机器人Gateway Session无效")
	}
	player.session = session
	player.nextSequence = 0
	player.state = StateConnected
	return nil
}

// GatewayHandler 把Origin网络回调交回当前Player；回调运行在Service串行上下文。
func (player *Player) GatewayHandler() network.Handler {
	return network.HandlerFuncs{
		Message: func(_ context.Context, session network.Session, payload []byte) error {
			if player.session != nil && player.session != session {
				return errors.New("机器人收到旧Gateway Session消息")
			}
			return player.HandleInbound(payload)
		},
		Close: func(_ context.Context, session network.Session, cause error) {
			player.ConnectionClosed(session, cause)
		},
	}
}

// SendRequest 严格先登记pending和超时，再提交网络发送。
func (player *Player) SendRequest(requestID commonpb.MessageID, expectedResponseID commonpb.MessageID, body []byte, timeout time.Duration, complete func(Response)) error {
	if player == nil {
		return errors.New("机器人请求状态无效")
	}
	return player.sendRequest(
		&player.businessPending, requestID, expectedResponseID, body, timeout, complete,
	)
}

func (player *Player) sendRequest(
	slot **pendingRequest,
	requestID commonpb.MessageID,
	expectedResponseID commonpb.MessageID,
	body []byte,
	timeout time.Duration,
	complete func(Response),
) error {
	if player.session == nil || slot == nil || *slot != nil || timeout <= 0 || complete == nil {
		return errors.New("机器人请求状态无效或已有pending")
	}
	player.nextSequence++
	if player.nextSequence == 0 {
		player.nextSequence = 1
	}
	sequence := player.nextSequence
	pending := &pendingRequest{
		sequence: sequence, expectedMessageID: expectedResponseID, complete: complete,
	}
	cancelTimeout, err := player.schedule(timeout, func() {
		if *slot != pending {
			return
		}
		player.finishPending(slot, Response{MessageID: expectedResponseID, Err: context.DeadlineExceeded})
	})
	if err != nil {
		return fmt.Errorf("登记机器人请求超时: %w", err)
	}
	pending.cancelTimeout = cancelTimeout
	*slot = pending
	payload, err := EncodeRequest(requestID, sequence, body)
	if err == nil {
		err = player.session.Send(payload)
	}
	if err != nil {
		player.finishPending(slot, Response{MessageID: expectedResponseID, Err: err})
		// pending已经登记后，发送失败也通过唯一completion收口，调用方不能再次恢复同一YieldHandle。
		return nil
	}
	return nil
}

// HandleInbound 只完成有界协议解析、push分发和两个固定pending匹配。
func (player *Player) HandleInbound(payload []byte) error {
	message, err := DecodeInbound(payload)
	if err != nil {
		return err
	}
	if message.Sequence == 0 {
		if player.onPush != nil {
			player.onPush(message)
		}
		return nil
	}
	response := Response{MessageID: message.MessageID, ErrorCode: message.ErrorCode, Body: message.Body}
	if pendingMatches(player.businessPending, message) {
		player.finishPending(&player.businessPending, response)
		return nil
	}
	if pendingMatches(player.heartbeatPending, message) {
		player.finishPending(&player.heartbeatPending, response)
		return nil
	}
	return errors.New("机器人收到无法匹配的Gateway响应")
}

func pendingMatches(pending *pendingRequest, message InboundMessage) bool {
	return pending != nil && pending.sequence == message.Sequence &&
		pending.expectedMessageID == message.MessageID
}

func (player *Player) finishPending(slot **pendingRequest, response Response) {
	if slot == nil {
		return
	}
	pending := *slot
	if pending == nil {
		return
	}
	*slot = nil
	if pending.cancelTimeout != nil {
		pending.cancelTimeout()
	}
	pending.complete(response)
}

// StartHeartbeat 启动一个不占用主蓝图执行链的有界后台心跳。
func (player *Player) StartHeartbeat(interval time.Duration, timeout time.Duration, onFailure func(error)) error {
	if player == nil || player.state != StateOnline || player.session == nil || player.heartbeat != nil ||
		interval <= 0 || timeout <= 0 || onFailure == nil {
		return errors.New("机器人后台心跳状态无效")
	}
	loop := &heartbeatLoop{interval: interval, timeout: timeout, onFailure: onFailure}
	player.heartbeat = loop
	if err := player.scheduleNextHeartbeat(loop); err != nil {
		player.heartbeat = nil
		return err
	}
	return nil
}

func (player *Player) scheduleNextHeartbeat(loop *heartbeatLoop) error {
	if player.heartbeat != loop || player.state != StateOnline || player.session == nil {
		return errors.New("机器人后台心跳已停止")
	}
	cancel, err := player.schedule(loop.interval, func() { player.sendHeartbeat(loop) })
	if err != nil {
		return fmt.Errorf("登记机器人心跳Timer: %w", err)
	}
	loop.cancelNext = cancel
	return nil
}

func (player *Player) sendHeartbeat(loop *heartbeatLoop) {
	if player.heartbeat != loop || player.state != StateOnline || player.session == nil {
		return
	}
	loop.cancelNext = nil
	err := player.sendRequest(
		&player.heartbeatPending,
		commonpb.MessageID_PlayerHeartbeatReq,
		commonpb.MessageID_Ok,
		nil,
		loop.timeout,
		func(response Response) {
			if player.heartbeat != loop {
				return
			}
			if response.Err != nil {
				player.failHeartbeat(loop, response.Err)
				return
			}
			if response.ErrorCode != commonpb.ErrorCode_ERROR_CODE_OK {
				player.failHeartbeat(loop, fmt.Errorf("心跳返回错误码%d", response.ErrorCode))
				return
			}
			if scheduleErr := player.scheduleNextHeartbeat(loop); scheduleErr != nil {
				player.failHeartbeat(loop, scheduleErr)
			}
		},
	)
	if err != nil {
		player.failHeartbeat(loop, err)
	}
}

// StopHeartbeat 幂等取消心跳Timer和心跳pending，不将主动停止报告为失败。
func (player *Player) StopHeartbeat() {
	if player == nil || player.heartbeat == nil {
		return
	}
	loop := player.heartbeat
	player.heartbeat = nil
	if loop.cancelNext != nil {
		loop.cancelNext()
		loop.cancelNext = nil
	}
	if player.heartbeatPending != nil {
		player.finishPending(&player.heartbeatPending, Response{Err: context.Canceled})
	}
}

func (player *Player) failHeartbeat(loop *heartbeatLoop, cause error) {
	if player.heartbeat != loop {
		return
	}
	player.StopHeartbeat()
	if cause == nil {
		cause = errors.New("机器人心跳失败")
	}
	player.state = StateFailed
	player.failure = cause
	if player.businessPending != nil {
		player.finishPending(&player.businessPending, Response{Err: cause})
	}
	if player.session != nil {
		session := player.session
		player.session = nil
		session.Close(cause)
	}
	loop.onFailure(cause)
}

// MarkOnline 在LoginPlayer成功响应完成解码后进入Online。
func (player *Player) MarkOnline() error {
	if player == nil || player.state != StateConnected {
		return errors.New("机器人不能在当前状态进入Online")
	}
	player.state = StateOnline
	return nil
}

// TokenForLoginPlayer 只供协议节点构造请求，调用方不得记录返回值。
func (player *Player) TokenForLoginPlayer() (string, error) {
	if player == nil || player.state != StateConnected || player.token == "" {
		return "", errors.New("机器人Gateway登录凭证不可用")
	}
	return player.token, nil
}

// ConnectionClosed 使两个pending得到确定失败；主动Close已经发布Closed时不改成Failed。
func (player *Player) ConnectionClosed(session network.Session, cause error) {
	if player == nil || player.session != session {
		return
	}
	player.session = nil
	closedErr := errors.New("Gateway连接已关闭")
	if cause != nil {
		closedErr = fmt.Errorf("Gateway连接已关闭: %w", cause)
	}
	if player.state != StateClosed {
		player.state = StateFailed
		player.failure = closedErr
	}
	loop := player.heartbeat
	player.StopHeartbeat()
	if player.businessPending != nil {
		player.finishPending(&player.businessPending, Response{Err: closedErr})
	}
	if loop != nil {
		loop.onFailure(closedErr)
	}
}

// Close 幂等清理心跳、pending、敏感凭证和Session。
func (player *Player) Close() {
	if player == nil || player.state == StateClosed {
		return
	}
	player.state = StateClosed
	player.StopHeartbeat()
	if player.businessPending != nil {
		player.finishPending(&player.businessPending, Response{Err: context.Canceled})
	}
	if player.session != nil {
		session := player.session
		player.session = nil
		session.Close(context.Canceled)
	}
	player.token = ""
	player.gatewayAddress = ""
}
