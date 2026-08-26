package player

import (
	"context"
	"testing"
	"time"

	"origingame/internal/playerroute"
	rpcapi "origingame/protocol/rpc"
)

type fakeRoutes struct {
	begin    int
	complete int
	release  int
	resident int
	leaving  int
	renewed  int
}

func (routes *fakeRoutes) BeginPlayerLoad(context.Context, playerroute.Player, playerroute.Instance, string) (bool, error) {
	routes.begin++
	return true, nil
}
func (routes *fakeRoutes) CompletePlayerLogin(context.Context, playerroute.Player, playerroute.Instance, string) (bool, error) {
	routes.complete++
	return true, nil
}
func (routes *fakeRoutes) ReleasePlayerLoad(context.Context, playerroute.Player, playerroute.Instance, string) (bool, error) {
	routes.release++
	return true, nil
}
func (routes *fakeRoutes) MarkPlayerResident(context.Context, playerroute.Player, playerroute.Instance) (bool, error) {
	routes.resident++
	return true, nil
}
func (routes *fakeRoutes) BeginPlayerRelease(context.Context, playerroute.Player, playerroute.Instance) (bool, error) {
	routes.leaving++
	return true, nil
}
func (routes *fakeRoutes) RenewPlayerRoutes(_ context.Context, _ int64, _ playerroute.Instance, players []playerroute.Player) (int64, error) {
	routes.renewed += len(players)
	return int64(len(players)), nil
}

func TestModuleLoadsInitialPlayerAndBuildsConnectionIndex(t *testing.T) {
	routes := &fakeRoutes{}
	mongoCalls := 0
	execute := func(_ context.Context, _ string, request rpcapi.MongoRequest) (rpcapi.MongoResult, error) {
		mongoCalls++
		results := make([]rpcapi.MongoOperationResult, len(request.Operations))
		for index := range results {
			results[index].Status = rpcapi.MongoOperationStatusSucceeded
		}
		return rpcapi.MongoResult{Results: results}, nil
	}
	module := NewModule(execute, routes, 1, playerroute.Instance{
		ServiceName: "GameService", NodeID: "area1-game-1", NodeSessionID: "session-1",
	}, nil, nil)
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
	if mongoCalls != 2 || routes.begin != 1 || routes.complete != 1 || routes.release != 0 {
		t.Fatalf("mongo=%d routes=%+v", mongoCalls, routes)
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

func TestRenewRoutesIncludesOnlineAndResidentPlayers(t *testing.T) {
	routes := &fakeRoutes{}
	execute := func(context.Context, string, rpcapi.MongoRequest) (rpcapi.MongoResult, error) {
		return rpcapi.MongoResult{}, nil
	}
	module := NewModule(execute, routes, 1, playerroute.Instance{
		ServiceName: "GameService", NodeID: "area1-game-1", NodeSessionID: "session-1",
	}, nil, nil)
	if err := module.OnInit(); err != nil {
		t.Fatal(err)
	}
	for index, state := range []State{StateOnline, StateResident, StateReleasing} {
		current, err := New("account-"+string(rune('a'+index)), int64(index+1), 1)
		if err != nil {
			t.Fatal(err)
		}
		current.dataInfo.State = state
		module.playersByKey[current.Key()] = current
	}
	module.renewRoutes(context.Background())
	if routes.renewed != 2 {
		t.Fatalf("renewed=%d, want 2", routes.renewed)
	}
}

func TestDisconnectKeepsPlayerResidentForFifteenMinutes(t *testing.T) {
	routes := &fakeRoutes{}
	execute := func(_ context.Context, _ string, request rpcapi.MongoRequest) (rpcapi.MongoResult, error) {
		results := make([]rpcapi.MongoOperationResult, len(request.Operations))
		for index := range results {
			results[index].Status = rpcapi.MongoOperationStatusSucceeded
		}
		return rpcapi.MongoResult{Results: results}, nil
	}
	module := NewModule(execute, routes, 1, playerroute.Instance{
		ServiceName: "GameService", NodeID: "area1-game-1", NodeSessionID: "session-1",
	}, nil, nil)
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
		module.FindByConnection("connection-1") != nil || routes.resident != 1 ||
		deadline.Before(now.Add(residentDuration-time.Second)) || deadline.After(now.Add(residentDuration+time.Second)) {
		t.Fatalf("resident state invalid: data=%+v routes=%+v", current.DataInfo(), routes)
	}
}
