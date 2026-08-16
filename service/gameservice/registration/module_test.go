package registration

import (
	"context"
	"sync"
	"testing"
	"time"

	"origingame/internal/playerroute"
)

type fakeRouteStore struct {
	mu         sync.Mutex
	registers  int
	drains     int
	registered chan struct{}
}

func (store *fakeRouteStore) RegisterGameService(context.Context, playerroute.Registration) error {
	store.mu.Lock()
	store.registers++
	store.mu.Unlock()
	select {
	case store.registered <- struct{}{}:
	default:
	}
	return nil
}

func (store *fakeRouteStore) SetGameServiceDraining(context.Context, int64, playerroute.Instance) (bool, error) {
	store.mu.Lock()
	store.drains++
	store.mu.Unlock()
	return true, nil
}

func TestModuleRegistersRenewsAndDrains(t *testing.T) {
	store := &fakeRouteStore{registered: make(chan struct{}, 1)}
	module := New(store, store, playerroute.Registration{
		RealAreaID: 1,
		Instance:   playerroute.Instance{ServiceName: "GameService", NodeID: "game-area-1-1", NodeSessionID: "session-1"},
		MaxPlayers: 5000,
	})
	module.interval = 5 * time.Millisecond

	// 没有 Origin 绑定时直接运行等价的受控协程，聚焦登记生命周期。
	if err := module.store.RegisterGameService(context.Background(), module.registration); err != nil {
		t.Fatal(err)
	}
	<-store.registered
	runCtx, cancel := context.WithCancel(context.Background())
	module.cancel = cancel
	module.done = make(chan struct{})
	go module.run(runCtx)
	select {
	case <-store.registered:
	case <-time.After(time.Second):
		t.Fatal("periodic registration did not run")
	}
	if err := module.OnStop(context.Background()); err != nil {
		t.Fatalf("OnStop() error = %v", err)
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.registers < 2 || store.drains != 1 {
		t.Fatalf("registers=%d drains=%d", store.registers, store.drains)
	}
}

func TestModuleStopUsesLifecycleStore(t *testing.T) {
	runStore := &fakeRouteStore{registered: make(chan struct{}, 1)}
	stopStore := &fakeRouteStore{registered: make(chan struct{}, 1)}
	module := New(runStore, stopStore, playerroute.Registration{
		RealAreaID: 1,
		Instance:   playerroute.Instance{ServiceName: "GameService", NodeID: "game-area-1-1", NodeSessionID: "session-1"},
		MaxPlayers: 5000,
	})

	if err := module.OnStop(context.Background()); err != nil {
		t.Fatalf("OnStop() error = %v", err)
	}
	runStore.mu.Lock()
	runDrains := runStore.drains
	runStore.mu.Unlock()
	stopStore.mu.Lock()
	stopDrains := stopStore.drains
	stopStore.mu.Unlock()
	if runDrains != 0 || stopDrains != 1 {
		t.Fatalf("run store drains=%d stop store drains=%d", runDrains, stopDrains)
	}
}
