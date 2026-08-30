package client

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/duanhf2012/origin/v3/sysmodule/network"
	"origingame/internal/playerownership"
	commonpb "origingame/protocol/common"
	rpcapi "origingame/protocol/rpc"
)

type fakeSession struct {
	id      network.SessionID
	sent    [][]byte
	final   []byte
	closed  bool
	context context.Context
	cancel  context.CancelFunc
}

type noCapacityOwnerships struct{ calls int }

type fakeGameServiceCaller struct{ disconnected int }

func (*fakeGameServiceCaller) LoginPlayer(
	context.Context,
	playerownership.GameServiceInstance,
	rpcapi.LoginPlayerRequest,
) (*commonpb.LoginPlayerResult, error) {
	return nil, nil
}

func (*fakeGameServiceCaller) HandlePlayerMessage(playerownership.GameServiceInstance, rpcapi.PlayerMessageRequest) error {
	return nil
}

func (caller *fakeGameServiceCaller) PlayerDisconnected(playerownership.GameServiceInstance, string) error {
	caller.disconnected++
	return nil
}

func (ownerships *noCapacityOwnerships) AssignOrGet(context.Context, playerownership.AssignRequest) (playerownership.AssignmentResult, error) {
	ownerships.calls++
	return playerownership.AssignmentResult{Decision: playerownership.AssignmentDecisionNoCapacity}, nil
}

func (*noCapacityOwnerships) ReleasePlayerLoad(context.Context, playerownership.Player, playerownership.GameServiceInstance, string) (bool, error) {
	return true, nil
}

func newFakeSession(id string) *fakeSession {
	ctx, cancel := context.WithCancel(context.Background())
	return &fakeSession{id: network.SessionID(id), context: ctx, cancel: cancel}
}

func (session *fakeSession) ID() network.SessionID    { return session.id }
func (*fakeSession) Transport() network.Transport     { return network.TransportTCP }
func (*fakeSession) LocalAddr() net.Addr              { return nil }
func (*fakeSession) RemoteAddr() net.Addr             { return nil }
func (session *fakeSession) Context() context.Context { return session.context }
func (session *fakeSession) Done() <-chan struct{}    { return session.context.Done() }
func (session *fakeSession) Writable() bool           { return !session.closed }
func (*fakeSession) Cause() error                     { return nil }
func (*fakeSession) Stats() network.SessionStats      { return network.SessionStats{} }
func (session *fakeSession) Send(payload []byte) error {
	session.sent = append(session.sent, append([]byte(nil), payload...))
	return nil
}
func (session *fakeSession) Close(error) { session.closed = true; session.cancel() }
func (session *fakeSession) SendAndClose(payload []byte, _ error) error {
	session.final = append([]byte(nil), payload...)
	session.closed = true
	session.cancel()
	return nil
}

func TestConnectionMessageLimitUsesFixedTenSecondWindow(t *testing.T) {
	current := newConnection(newFakeSession("connection-1"), time.Unix(100, 0))
	for index := 0; index < messageWindowMax; index++ {
		if !current.allowMessage(time.Unix(105, 0)) {
			t.Fatalf("第 %d 条消息被错误拒绝", index+1)
		}
	}
	if current.allowMessage(time.Unix(109, 0)) {
		t.Fatal("同一窗口第 201 条消息必须拒绝")
	}
	if !current.allowMessage(time.Unix(110, 0)) {
		t.Fatal("新窗口第一条消息必须允许")
	}
}

func TestSendClientMessageUsesFinalNetworkWrite(t *testing.T) {
	session := newFakeSession("connection-1")
	module := &Module{connections: map[network.SessionID]*connection{
		session.ID(): newConnection(session, time.Now()),
	}}
	err := module.SendClientMessage(rpcapi.SendClientMessageRequest{
		GatewayConnectionID: string(session.ID()),
		Message: rpcapi.ClientMessage{
			MessageID:       commonpb.MessageID_PlayerKicked,
			ErrorCode:       commonpb.ErrorCode_ERROR_CODE_OK,
			CloseAfterWrite: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(session.final) == 0 || len(session.sent) != 0 || !session.closed {
		t.Fatalf("final=%v sent=%d closed=%v", session.final, len(session.sent), session.closed)
	}
}

func TestCloseNotifiesBoundGameServiceOnce(t *testing.T) {
	session := newFakeSession("connection-1")
	gameServices := &fakeGameServiceCaller{}
	module := &Module{
		connections:  map[network.SessionID]*connection{},
		dependencies: Dependencies{GameServices: gameServices},
	}
	current := newConnection(session, time.Now())
	current.state = stateOnline
	module.connections[session.ID()] = current
	module.onClose(context.Background(), session, nil)
	module.onClose(context.Background(), session, nil)
	if gameServices.disconnected != 1 {
		t.Fatalf("notified=%d", gameServices.disconnected)
	}
}

func TestAssignAndLoginWaitsForCapacityUntilLoginDeadline(t *testing.T) {
	ownerships := &noCapacityOwnerships{}
	session := newFakeSession("connection-1")
	module := &Module{dependencies: Dependencies{OwnershipStore: ownerships}}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err := module.assignAndLogin(ctx, newConnection(session, time.Now()), "account-1", 1, 1)
	if !errors.Is(err, context.DeadlineExceeded) || ownerships.calls != 1 {
		t.Fatalf("assignAndLogin error=%v calls=%d", err, ownerships.calls)
	}
}
