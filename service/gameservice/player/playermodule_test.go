package player

import (
	"context"
	"testing"
	"time"

	"origingame/internal/playerownership"
	rpcapi "origingame/protocol/rpc"
)

type testMongoExecutor struct {
	execute func(context.Context, string, rpcapi.MongoRequest) (rpcapi.MongoResult, error)
}

func (executor testMongoExecutor) ExecuteMongo(ctx context.Context, key string, request rpcapi.MongoRequest) (rpcapi.MongoResult, error) {
	return executor.execute(ctx, key, request)
}

type fakeOwnerships struct {
	begin    int
	complete int
	release  int
	resident int
	leaving  int
	renewed  int
}

func (ownerships *fakeOwnerships) BeginPlayerLoad(context.Context, playerownership.Player, playerownership.GameServiceInstance, string) (bool, error) {
	ownerships.begin++
	return true, nil
}
func (ownerships *fakeOwnerships) CompletePlayerLogin(context.Context, playerownership.Player, playerownership.GameServiceInstance, string) (bool, error) {
	ownerships.complete++
	return true, nil
}
func (ownerships *fakeOwnerships) ReleasePlayerLoad(context.Context, playerownership.Player, playerownership.GameServiceInstance, string) (bool, error) {
	ownerships.release++
	return true, nil
}
func (ownerships *fakeOwnerships) MarkPlayerResident(context.Context, playerownership.Player, playerownership.GameServiceInstance) (bool, error) {
	ownerships.resident++
	return true, nil
}
func (ownerships *fakeOwnerships) BeginPlayerRelease(context.Context, playerownership.Player, playerownership.GameServiceInstance) (bool, error) {
	ownerships.leaving++
	return true, nil
}
func (ownerships *fakeOwnerships) RenewPlayerOwnerships(_ context.Context, _ int64, _ playerownership.GameServiceInstance, players []playerownership.Player) (int64, error) {
	ownerships.renewed += len(players)
	return int64(len(players)), nil
}

func TestModuleLoadsInitialPlayerAndBuildsConnectionIndex(t *testing.T) {
	ownerships := &fakeOwnerships{}
	mongoCalls := 0
	execute := func(_ context.Context, _ string, request rpcapi.MongoRequest) (rpcapi.MongoResult, error) {
		mongoCalls++
		results := make([]rpcapi.MongoOperationResult, len(request.Operations))
		for index := range results {
			results[index].Status = rpcapi.MongoOperationStatusSucceeded
		}
		return rpcapi.MongoResult{Results: results}, nil
	}
	module := NewPlayerModule(testMongoExecutor{execute: execute}, ownerships, 1, playerownership.GameServiceInstance{
		ServiceName: "GameService", NodeID: "area1-game-1", NodeSessionID: "session-1",
	}, nil)
	if err := module.OnInit(); err != nil {
		t.Fatal(err)
	}
	current, err := module.LoadNew(
		context.Background(), "0123456789abcdef01234567", 10, "pub-gateway-1", "connection-1",
	)
	if err != nil {
		t.Fatalf("LoadNew() error = %v", err)
	}
	defer module.stopAutoSave(current)
	if mongoCalls != 2 || ownerships.begin != 1 || ownerships.complete != 1 || ownerships.release != 0 {
		t.Fatalf("mongo=%d ownerships=%+v", mongoCalls, ownerships)
	}
	if current.DataInfo().State != StateOnline || module.FindByConnection("connection-1") != current ||
		module.FindByKey(current.Key()) != current {
		t.Fatalf("player/index state invalid: %+v", current.DataInfo())
	}
}

func TestInitialSaveDelayIsStableAndStaggered(t *testing.T) {
	first := initialSaveDelay("account-a:1")
	if first != initialSaveDelay("account-a:1") || first < time.Second || first >= autoSaveInterval {
		t.Fatalf("initialSaveDelay() = %v", first)
	}
	if first == initialSaveDelay("account-b:1") {
		t.Fatal("different PlayerKeys unexpectedly share the same initial delay")
	}
}

func TestRenewOwnershipsIncludesOnlineAndResidentPlayers(t *testing.T) {
	ownerships := &fakeOwnerships{}
	execute := func(context.Context, string, rpcapi.MongoRequest) (rpcapi.MongoResult, error) {
		return rpcapi.MongoResult{}, nil
	}
	module := NewPlayerModule(testMongoExecutor{execute: execute}, ownerships, 1, playerownership.GameServiceInstance{
		ServiceName: "GameService", NodeID: "area1-game-1", NodeSessionID: "session-1",
	}, nil)
	if err := module.OnInit(); err != nil {
		t.Fatal(err)
	}
	for index, state := range []State{StateOnline, StateResident, StateReleasing} {
		current, err := NewPlayer("account-"+string(rune('a'+index)), int64(index+1), 1)
		if err != nil {
			t.Fatal(err)
		}
		current.dataInfo.State = state
		module.playersByKey[current.Key()] = current
	}
	module.renewOwnerships(context.Background())
	if ownerships.renewed != 2 {
		t.Fatalf("renewed=%d, want 2", ownerships.renewed)
	}
}

func TestDisconnectKeepsPlayerResidentForFifteenMinutes(t *testing.T) {
	ownerships := &fakeOwnerships{}
	execute := func(_ context.Context, _ string, request rpcapi.MongoRequest) (rpcapi.MongoResult, error) {
		results := make([]rpcapi.MongoOperationResult, len(request.Operations))
		for index := range results {
			results[index].Status = rpcapi.MongoOperationStatusSucceeded
		}
		return rpcapi.MongoResult{Results: results}, nil
	}
	module := NewPlayerModule(testMongoExecutor{execute: execute}, ownerships, 1, playerownership.GameServiceInstance{
		ServiceName: "GameService", NodeID: "area1-game-1", NodeSessionID: "session-1",
	}, nil)
	if err := module.OnInit(); err != nil {
		t.Fatal(err)
	}
	current, err := module.LoadNew(
		context.Background(), "0123456789abcdef01234567", 10, "pub-gateway-1", "connection-1",
	)
	if err != nil {
		t.Fatal(err)
	}
	defer module.stopAutoSave(current)
	now := time.Now().UTC()
	if err = module.Disconnect(context.Background(), "connection-1"); err != nil {
		t.Fatal(err)
	}
	deadline := current.DataInfo().ResidentDeadline
	if current.DataInfo().State != StateResident || module.FindByKey(current.Key()) != current ||
		module.FindByConnection("connection-1") != nil || ownerships.resident != 1 ||
		deadline.Before(now.Add(residentDuration-time.Second)) || deadline.After(now.Add(residentDuration+time.Second)) {
		t.Fatalf("resident state invalid: data=%+v ownerships=%+v", current.DataInfo(), ownerships)
	}
}
