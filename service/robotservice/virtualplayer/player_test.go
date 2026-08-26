package virtualplayer

import (
	"context"
	"encoding/binary"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/duanhf2012/origin/v3/sysmodule/network"
	commonpb "origingame/protocol/common"
)

type fakeSession struct {
	ctx     context.Context
	cancel  context.CancelFunc
	sent    [][]byte
	sendErr error
}

func newFakeSession() *fakeSession {
	ctx, cancel := context.WithCancel(context.Background())
	return &fakeSession{ctx: ctx, cancel: cancel}
}

func (*fakeSession) ID() network.SessionID            { return "test-session" }
func (*fakeSession) Transport() network.Transport     { return network.TransportTCP }
func (*fakeSession) LocalAddr() net.Addr              { return nil }
func (*fakeSession) RemoteAddr() net.Addr             { return nil }
func (session *fakeSession) Context() context.Context { return session.ctx }
func (session *fakeSession) Done() <-chan struct{}    { return session.ctx.Done() }
func (session *fakeSession) Send(payload []byte) error {
	if session.sendErr != nil {
		return session.sendErr
	}
	session.sent = append(session.sent, append([]byte(nil), payload...))
	return nil
}
func (session *fakeSession) Close(error)         { session.cancel() }
func (*fakeSession) Writable() bool              { return true }
func (*fakeSession) Cause() error                { return nil }
func (*fakeSession) Stats() network.SessionStats { return network.SessionStats{} }

type manualTimer struct {
	callback func()
	canceled bool
}

type queuedTimer struct {
	callback func()
	canceled bool
}

type manualScheduler struct{ timers []*queuedTimer }

func (scheduler *manualScheduler) schedule(_ time.Duration, callback func()) (func(), error) {
	timer := &queuedTimer{callback: callback}
	scheduler.timers = append(scheduler.timers, timer)
	return func() { timer.canceled = true }, nil
}

func (scheduler *manualScheduler) fire(t *testing.T, index int) {
	t.Helper()
	if index < 0 || index >= len(scheduler.timers) || scheduler.timers[index].canceled {
		t.Fatalf("Timer %d无法触发", index)
	}
	scheduler.timers[index].callback()
}

func configuredPlayer(t *testing.T) (*Player, *fakeSession, *manualTimer) {
	t.Helper()
	timer := &manualTimer{}
	player, err := New(1, func(_ time.Duration, callback func()) (func(), error) {
		timer.callback = callback
		return func() { timer.canceled = true }, nil
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = player.ApplyLogin(LoginResult{Token: "secret", GatewayAddress: "127.0.0.1:9000"}); err != nil {
		t.Fatal(err)
	}
	session := newFakeSession()
	if err = player.BindSession(session); err != nil {
		t.Fatal(err)
	}
	return player, session, timer
}

func TestSendRequestMatchesSequenceAndMessageID(t *testing.T) {
	player, session, timer := configuredPlayer(t)
	var responses []Response
	if err := player.SendRequest(
		commonpb.MessageID_LoginPlayerReq, commonpb.MessageID_LoginPlayerRes,
		[]byte("request"), time.Second, func(response Response) { responses = append(responses, response) },
	); err != nil {
		t.Fatal(err)
	}
	if len(session.sent) != 1 || binary.BigEndian.Uint32(session.sent[0][2:6]) != 1 {
		t.Fatalf("发送帧异常: %#v", session.sent)
	}
	if err := player.HandleInbound(responsePayload(commonpb.MessageID_Ok, 1, 0, nil)); err == nil {
		t.Fatal("错误 MessageID 被匹配")
	}
	if len(responses) != 0 {
		t.Fatal("不匹配响应完成了 pending")
	}
	if err := player.HandleInbound(responsePayload(
		commonpb.MessageID_LoginPlayerRes, 1, commonpb.ErrorCode_ERROR_CODE_OK, []byte("response"),
	)); err != nil {
		t.Fatal(err)
	}
	if len(responses) != 1 || string(responses[0].Body) != "response" || !timer.canceled {
		t.Fatalf("响应完成异常: %#v canceled=%v", responses, timer.canceled)
	}
}

func TestPendingTimeoutAndSendFailureCompleteExactlyOnce(t *testing.T) {
	player, session, timer := configuredPlayer(t)
	var responses []Response
	if err := player.SendRequest(
		commonpb.MessageID_PlayerHeartbeatReq, commonpb.MessageID_Ok, nil, time.Second,
		func(response Response) { responses = append(responses, response) },
	); err != nil {
		t.Fatal(err)
	}
	timer.callback()
	timer.callback()
	if len(responses) != 1 || !errors.Is(responses[0].Err, context.DeadlineExceeded) {
		t.Fatalf("超时完成异常: %#v", responses)
	}

	session.sendErr = errors.New("send failed")
	if err := player.SendRequest(
		commonpb.MessageID_PlayerHeartbeatReq, commonpb.MessageID_Ok, nil, time.Second,
		func(response Response) { responses = append(responses, response) },
	); err != nil {
		t.Fatalf("发送失败应由 completion 收口: %v", err)
	}
	if len(responses) != 2 || !errors.Is(responses[1].Err, session.sendErr) {
		t.Fatalf("发送失败完成异常: %#v", responses)
	}
}

func TestHeartbeatAndBusinessRequestsUseIndependentPendingSlots(t *testing.T) {
	scheduler := &manualScheduler{}
	player, err := New(1, scheduler.schedule, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = player.ApplyLogin(LoginResult{Token: "secret", GatewayAddress: "127.0.0.1:9000"}); err != nil {
		t.Fatal(err)
	}
	session := newFakeSession()
	if err = player.BindSession(session); err != nil {
		t.Fatal(err)
	}
	if err = player.MarkOnline(); err != nil {
		t.Fatal(err)
	}
	var heartbeatFailure error
	if err = player.StartHeartbeat(5*time.Second, time.Second, func(cause error) {
		heartbeatFailure = cause
	}); err != nil {
		t.Fatal(err)
	}
	// Timer 0 是首次心跳周期，触发后心跳占用独立pending。
	scheduler.fire(t, 0)
	var businessResponse Response
	if err = player.SendRequest(
		commonpb.MessageID_LoginPlayerReq,
		commonpb.MessageID_LoginPlayerRes,
		nil,
		time.Second,
		func(response Response) { businessResponse = response },
	); err != nil {
		t.Fatal(err)
	}
	if len(session.sent) != 2 {
		t.Fatalf("发送帧数=%d，期望心跳和业务各2帧", len(session.sent))
	}
	heartbeatSequence := binary.BigEndian.Uint32(session.sent[0][2:6])
	businessSequence := binary.BigEndian.Uint32(session.sent[1][2:6])
	if heartbeatSequence == 0 || businessSequence == 0 || heartbeatSequence == businessSequence {
		t.Fatalf("心跳/业务Sequence=%d/%d", heartbeatSequence, businessSequence)
	}
	// 业务响应可以先于心跳返回，两者不得互相阻塞或误匹配。
	if err = player.HandleInbound(responsePayload(
		commonpb.MessageID_LoginPlayerRes,
		businessSequence,
		commonpb.ErrorCode_ERROR_CODE_OK,
		[]byte("business"),
	)); err != nil {
		t.Fatal(err)
	}
	if string(businessResponse.Body) != "business" {
		t.Fatalf("业务响应=%+v", businessResponse)
	}
	if err = player.HandleInbound(responsePayload(
		commonpb.MessageID_Ok,
		heartbeatSequence,
		commonpb.ErrorCode_ERROR_CODE_OK,
		nil,
	)); err != nil {
		t.Fatal(err)
	}
	if heartbeatFailure != nil || player.State() != StateOnline {
		t.Fatalf("心跳结果 failure=%v state=%d", heartbeatFailure, player.State())
	}
	player.StopHeartbeat()
}

func TestHeartbeatTimeoutFailsPlayerAndClosesSession(t *testing.T) {
	scheduler := &manualScheduler{}
	player, err := New(1, scheduler.schedule, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err = player.ApplyLogin(LoginResult{Token: "secret", GatewayAddress: "127.0.0.1:9000"}); err != nil {
		t.Fatal(err)
	}
	session := newFakeSession()
	if err = player.BindSession(session); err != nil {
		t.Fatal(err)
	}
	if err = player.MarkOnline(); err != nil {
		t.Fatal(err)
	}
	var failure error
	if err = player.StartHeartbeat(5*time.Second, time.Second, func(cause error) { failure = cause }); err != nil {
		t.Fatal(err)
	}
	scheduler.fire(t, 0) // 发送心跳并登记超时Timer 1。
	scheduler.fire(t, 1)
	if !errors.Is(failure, context.DeadlineExceeded) || player.State() != StateFailed ||
		!errors.Is(player.Failure(), context.DeadlineExceeded) || session.Context().Err() == nil {
		t.Fatalf("心跳超时未收口 failure=%v player_failure=%v state=%d session=%v",
			failure, player.Failure(), player.State(), session.Context().Err())
	}
}

func responsePayload(messageID commonpb.MessageID, sequence uint32, code commonpb.ErrorCode, body []byte) []byte {
	payload := make([]byte, responseHeaderSize+len(body))
	binary.BigEndian.PutUint16(payload[:2], uint16(messageID))
	binary.BigEndian.PutUint32(payload[2:6], sequence)
	binary.BigEndian.PutUint32(payload[6:10], uint32(code))
	copy(payload[responseHeaderSize:], body)
	return payload
}
