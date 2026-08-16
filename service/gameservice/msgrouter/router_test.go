package msgrouter

import (
	"testing"

	commonpb "origingame/protocol/common"
	"origingame/service/gameservice/player"
)

func TestRegisterFreezeAndDispatchUsesConcreteMessageType(t *testing.T) {
	router := New()
	called := false
	err := Register(router, commonpb.MessageID_PlayerHeartbeatReq, func(
		_ *Session,
		_ *player.Player,
		_ *commonpb.PlayerHeartbeatRequest,
	) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = router.Freeze(); err != nil {
		t.Fatal(err)
	}
	if err = router.Dispatch(&player.Player{}, "connection-1", commonpb.MessageID_PlayerHeartbeatReq, 1, nil); err != nil {
		t.Fatalf("Dispatch() error = %v", err)
	}
	if !called {
		t.Fatal("registered handler was not called")
	}
	if err = Register(router, commonpb.MessageID_PlayerHeartbeatReq, func(*Session, *player.Player, *commonpb.PlayerHeartbeatRequest) error { return nil }); err == nil {
		t.Fatal("Register() accepted a route after Freeze")
	}
}

func BenchmarkRouterDispatch(b *testing.B) {
	router := New()
	if err := Register(router, commonpb.MessageID_PlayerHeartbeatReq, func(
		*Session,
		*player.Player,
		*commonpb.PlayerHeartbeatRequest,
	) error {
		return nil
	}); err != nil {
		b.Fatal(err)
	}
	if err := router.Freeze(); err != nil {
		b.Fatal(err)
	}
	current := &player.Player{}
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if err := router.Dispatch(current, "connection-1", commonpb.MessageID_PlayerHeartbeatReq, 1, nil); err != nil {
			b.Fatal(err)
		}
	}
}
