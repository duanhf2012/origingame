package registration

import (
	"context"
	"sync"
	"testing"
	"time"

	"origingame/internal/playerownership"
)

type fakeRegistrationStore struct {
	mu         sync.Mutex    // 保护测试状态。
	registers  int           // 注册调用数。
	drains     int           // 排空调用数。
	registered chan struct{} // 注册通知。
}

func (store *fakeRegistrationStore) RegisterGameService(context.Context, playerownership.GameServiceRegistration) error {
	store.mu.Lock()
	store.registers++
	store.mu.Unlock()
	select {
	case store.registered <- struct{}{}:
	default:
	}
	return nil
}

func (store *fakeRegistrationStore) SetGameServiceDraining(context.Context, int64, playerownership.GameServiceInstance) (bool, error) {
	store.mu.Lock()
	store.drains++
	store.mu.Unlock()
	return true, nil
}

func TestModuleRegistersRenewsAndDrains(t *testing.T) {
	store := &fakeRegistrationStore{registered: make(chan struct{}, 1)}
	module := NewGameServiceRegistrationModule(store, store, playerownership.GameServiceRegistration{
		RealAreaID:  1,
		GameService: playerownership.GameServiceInstance{ServiceName: "GameService", NodeID: "area1-game-1", NodeSessionID: "session-1"},
		MaxPlayers:  5000,
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
	runStore := &fakeRegistrationStore{registered: make(chan struct{}, 1)}
	stopStore := &fakeRegistrationStore{registered: make(chan struct{}, 1)}
	module := NewGameServiceRegistrationModule(runStore, stopStore, playerownership.GameServiceRegistration{
		RealAreaID:  1,
		GameService: playerownership.GameServiceInstance{ServiceName: "GameService", NodeID: "area1-game-1", NodeSessionID: "session-1"},
		MaxPlayers:  5000,
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
