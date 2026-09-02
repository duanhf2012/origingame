package playerownership

import (
	"context"
	"errors"
	"strings"
	"testing"

	"origingame/internal/redisscripts"
	rpcapi "origingame/protocol/rpc"
)

type testRedisExecutor struct {
	execute func(context.Context, string, rpcapi.RedisRequest) (rpcapi.RedisResult, error) // 模拟 Redis 执行。
}

func (executor testRedisExecutor) ExecuteRedis(ctx context.Context, key string, request rpcapi.RedisRequest) (rpcapi.RedisResult, error) {
	return executor.execute(ctx, key, request)
}

func newTestPlayerOwnershipStore(execute func(context.Context, string, rpcapi.RedisRequest) (rpcapi.RedisResult, error)) *PlayerOwnershipStore {
	return NewPlayerOwnershipStore(testRedisExecutor{execute: execute})
}

func TestAssignOrGetBuildsStableDispatchScriptRequest(t *testing.T) {
	var dispatchKey string
	var captured rpcapi.RedisRequest
	store := newTestPlayerOwnershipStore(func(_ context.Context, key string, request rpcapi.RedisRequest) (rpcapi.RedisResult, error) {
		dispatchKey, captured = key, request
		return assignmentRedisResult(AssignmentDecisionAssigned, "GameService", "area1-game-1", "session-1", "ASSIGNING"), nil
	})
	result, err := store.AssignOrGet(context.Background(), AssignRequest{
		Player:              Player{AccountID: "0123456789abcdef01234567", ShowAreaID: 10, RealAreaID: 1},
		GatewayConnectionID: "connection-1",
	})
	if err != nil {
		t.Fatalf("AssignOrGet() error = %v", err)
	}
	if result.Decision != AssignmentDecisionAssigned || result.GameService.NodeID != "area1-game-1" {
		t.Fatalf("AssignOrGet() = %+v", result)
	}
	if dispatchKey != "0123456789abcdef01234567:10" || dispatchKey != captured.DispatchKey ||
		captured.Script == nil || captured.Script.ID != redisscripts.AssignOrGetPlayerID || len(captured.Script.Keys) != 2 {
		t.Fatalf("unexpected request: dispatch=%q request=%+v", dispatchKey, captured)
	}
	for _, key := range captured.Script.Keys {
		if !strings.HasPrefix(key, "{area:1}:") {
			t.Fatalf("key %q does not share area hash tag", key)
		}
	}
}

func TestRegisterGameServiceUsesInstanceAsDispatchKey(t *testing.T) {
	var dispatchKey string
	var captured rpcapi.RedisRequest
	store := newTestPlayerOwnershipStore(func(_ context.Context, key string, request rpcapi.RedisRequest) (rpcapi.RedisResult, error) {
		dispatchKey, captured = key, request
		return integerRedisResult(1), nil
	})
	err := store.RegisterGameService(context.Background(), GameServiceRegistration{
		RealAreaID:  1,
		GameService: GameServiceInstance{ServiceName: "GameService", NodeID: "area1-game-1", NodeSessionID: "session-1"},
		MaxPlayers:  5000,
	})
	if err != nil {
		t.Fatalf("RegisterGameService() error = %v", err)
	}
	if dispatchKey == "" || dispatchKey != captured.DispatchKey || captured.Script == nil ||
		captured.Script.ID != redisscripts.RegisterGameServiceID || len(captured.Script.Keys) != 3 ||
		len(captured.Script.Args) != 5 || string(captured.Script.Args[4]) != "15000" {
		t.Fatalf("unexpected request: dispatch=%q request=%+v", dispatchKey, captured)
	}
}

func TestAssignOrGetPropagatesRedisFailure(t *testing.T) {
	want := errors.New("redis unavailable")
	store := newTestPlayerOwnershipStore(func(context.Context, string, rpcapi.RedisRequest) (rpcapi.RedisResult, error) {
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
	store := newTestPlayerOwnershipStore(func(_ context.Context, _ string, request rpcapi.RedisRequest) (rpcapi.RedisResult, error) {
		captured = request
		return integerRedisResult(1), nil
	})
	accepted, err := store.BeginPlayerRelease(
		context.Background(),
		Player{AccountID: "0123456789abcdef01234567", ShowAreaID: 10, RealAreaID: 1},
		GameServiceInstance{ServiceName: "GameService", NodeID: "area1-game-1", NodeSessionID: "session-1"},
	)
	if err != nil || !accepted || captured.Script == nil ||
		captured.Script.ID != redisscripts.BeginPlayerReleaseID || len(captured.Script.Args) != 3 ||
		string(captured.Script.Args[2]) != "5000" {
		t.Fatalf("BeginPlayerRelease() accepted=%v error=%v request=%+v", accepted, err, captured)
	}
}

func TestRenewPlayerOwnershipsRejectsOversizedBatch(t *testing.T) {
	store := newTestPlayerOwnershipStore(func(context.Context, string, rpcapi.RedisRequest) (rpcapi.RedisResult, error) {
		return integerRedisResult(0), nil
	})
	players := make([]Player, maxRenewOwnerships+1)
	for index := range players {
		players[index] = Player{AccountID: "0123456789abcdef01234567", ShowAreaID: int64(index + 1), RealAreaID: 1}
	}
	_, err := store.RenewPlayerOwnerships(context.Background(), 1, GameServiceInstance{
		ServiceName: "GameService", NodeID: "area1-game-1", NodeSessionID: "session-1",
	}, players)
	if err == nil {
		t.Fatal("RenewPlayerOwnerships() accepted more than 256 ownerships")
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
