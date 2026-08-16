package playerroute

import (
	"context"
	"errors"
	"strings"
	"testing"

	"origingame/internal/redisscripts"
	rpcapi "origingame/protocol/rpc"
)

func TestAssignOrGetBuildsStableRoutedScriptRequest(t *testing.T) {
	var routeKey string
	var captured rpcapi.RedisRequest
	store := New(func(_ context.Context, key string, request rpcapi.RedisRequest) (rpcapi.RedisResult, error) {
		routeKey, captured = key, request
		return assignmentRedisResult(AssignmentDecisionAssigned, "GameService", "game-area-1-1", "session-1", "ASSIGNING"), nil
	})
	result, err := store.AssignOrGet(context.Background(), AssignRequest{
		Player:              Player{AccountID: "0123456789abcdef01234567", ShowAreaID: 10, RealAreaID: 1},
		GatewayConnectionID: "connection-1",
	})
	if err != nil {
		t.Fatalf("AssignOrGet() error = %v", err)
	}
	if result.Decision != AssignmentDecisionAssigned || result.Instance.NodeID != "game-area-1-1" {
		t.Fatalf("AssignOrGet() = %+v", result)
	}
	if routeKey != "0123456789abcdef01234567:10" || routeKey != captured.DispatchKey ||
		captured.Script == nil || captured.Script.ID != redisscripts.AssignOrGetPlayerID || len(captured.Script.Keys) != 2 {
		t.Fatalf("unexpected request: route=%q request=%+v", routeKey, captured)
	}
	for _, key := range captured.Script.Keys {
		if !strings.HasPrefix(key, "{area:1}:") {
			t.Fatalf("key %q does not share area hash tag", key)
		}
	}
}

func TestRegisterGameServiceUsesInstanceAsDispatchKey(t *testing.T) {
	var routeKey string
	var captured rpcapi.RedisRequest
	store := New(func(_ context.Context, key string, request rpcapi.RedisRequest) (rpcapi.RedisResult, error) {
		routeKey, captured = key, request
		return integerRedisResult(1), nil
	})
	err := store.RegisterGameService(context.Background(), Registration{
		RealAreaID: 1,
		Instance:   Instance{ServiceName: "GameService", NodeID: "game-area-1-1", NodeSessionID: "session-1"},
		MaxPlayers: 5000,
	})
	if err != nil {
		t.Fatalf("RegisterGameService() error = %v", err)
	}
	if routeKey == "" || routeKey != captured.DispatchKey || captured.Script == nil ||
		captured.Script.ID != redisscripts.RegisterGameServiceID || len(captured.Script.Keys) != 3 ||
		len(captured.Script.Args) != 5 || string(captured.Script.Args[4]) != "15000" {
		t.Fatalf("unexpected request: route=%q request=%+v", routeKey, captured)
	}
}

func TestAssignOrGetPropagatesRedisFailure(t *testing.T) {
	want := errors.New("redis unavailable")
	store := New(func(context.Context, string, rpcapi.RedisRequest) (rpcapi.RedisResult, error) {
		return rpcapi.RedisResult{}, want
	})
	_, err := store.AssignOrGet(context.Background(), AssignRequest{
		Player:              Player{AccountID: "0123456789abcdef01234567", ShowAreaID: 10, RealAreaID: 1},
		GatewayConnectionID: "connection-1",
	})
	if !errors.Is(err, want) {
		t.Fatalf("AssignOrGet() error = %v", err)
	}
}

func TestBeginPlayerReleaseUsesFiveSecondIsolation(t *testing.T) {
	var captured rpcapi.RedisRequest
	store := New(func(_ context.Context, _ string, request rpcapi.RedisRequest) (rpcapi.RedisResult, error) {
		captured = request
		return integerRedisResult(1), nil
	})
	accepted, err := store.BeginPlayerRelease(
		context.Background(),
		Player{AccountID: "0123456789abcdef01234567", ShowAreaID: 10, RealAreaID: 1},
		Instance{ServiceName: "GameService", NodeID: "game-area-1-1", NodeSessionID: "session-1"},
	)
	if err != nil || !accepted || captured.Script == nil ||
		captured.Script.ID != redisscripts.BeginPlayerReleaseID || len(captured.Script.Args) != 3 ||
		string(captured.Script.Args[2]) != "5000" {
		t.Fatalf("BeginPlayerRelease() accepted=%v error=%v request=%+v", accepted, err, captured)
	}
}

func TestRenewPlayerRoutesRejectsOversizedBatch(t *testing.T) {
	store := New(func(context.Context, string, rpcapi.RedisRequest) (rpcapi.RedisResult, error) {
		return integerRedisResult(0), nil
	})
	players := make([]Player, maxRenewRoutes+1)
	for index := range players {
		players[index] = Player{AccountID: "0123456789abcdef01234567", ShowAreaID: int64(index + 1), RealAreaID: 1}
	}
	_, err := store.RenewPlayerRoutes(context.Background(), 1, Instance{
		ServiceName: "GameService", NodeID: "game-area-1-1", NodeSessionID: "session-1",
	}, players)
	if err == nil {
		t.Fatal("RenewPlayerRoutes() accepted more than 256 routes")
	}
}

func assignmentRedisResult(
	decision AssignmentDecision,
	serviceName string,
	nodeID string,
	nodeSessionID string,
	state string,
) rpcapi.RedisResult {
	nodes := []rpcapi.RedisValueNode{
		{Kind: rpcapi.RedisValueKindArray, Children: []uint32{1, 2, 3, 4, 5}},
		{Kind: rpcapi.RedisValueKindInteger, Integer: int64(decision)},
		{Kind: rpcapi.RedisValueKindBytes, Bytes: []byte(serviceName)},
		{Kind: rpcapi.RedisValueKindBytes, Bytes: []byte(nodeID)},
		{Kind: rpcapi.RedisValueKindBytes, Bytes: []byte(nodeSessionID)},
		{Kind: rpcapi.RedisValueKindBytes, Bytes: []byte(state)},
	}
	return rpcapi.RedisResult{Results: []rpcapi.RedisCommandResult{{
		Status: rpcapi.RedisCommandStatusSucceeded,
		Value:  rpcapi.RedisValue{RootIndex: 0, Nodes: nodes},
	}}}
}

func integerRedisResult(value int64) rpcapi.RedisResult {
	return rpcapi.RedisResult{Results: []rpcapi.RedisCommandResult{{
		Status: rpcapi.RedisCommandStatusSucceeded,
		Value: rpcapi.RedisValue{RootIndex: 0, Nodes: []rpcapi.RedisValueNode{{
			Kind: rpcapi.RedisValueKindInteger, Integer: value,
		}}},
	}}}
}
