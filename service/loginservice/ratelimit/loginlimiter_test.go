package ratelimit

import (
	"context"
	"strings"
	"testing"
	"time"

	originconfig "github.com/duanhf2012/origin/v3/config"
	rpcapi "origingame/protocol/rpc"
)

type testRedisExecutor struct {
	execute func(context.Context, string, rpcapi.RedisRequest) (rpcapi.RedisResult, error) // 模拟 Redis 执行。
}

func (executor testRedisExecutor) ExecuteRedis(ctx context.Context, key string, request rpcapi.RedisRequest) (rpcapi.RedisResult, error) {
	return executor.execute(ctx, key, request)
}

func TestLimiterBuildsRegisteredScriptRequest(t *testing.T) {
	var routeKey string
	var captured rpcapi.RedisRequest
	limiter := NewLimiter(Config{
		Enabled: true,
		IP:      WindowLimitConfig{Enabled: true, Window: originconfig.Duration(10 * time.Second), MaxRequests: 30},
	}, testRedisExecutor{execute: func(_ context.Context, key string, request rpcapi.RedisRequest) (rpcapi.RedisResult, error) {
		routeKey, captured = key, request
		return integerResult(1), nil
	}})
	limiter.now = func() time.Time { return time.UnixMilli(123456) }

	allowed, err := limiter.AllowIP(context.Background(), "192.0.2.10")
	if err != nil || !allowed {
		t.Fatalf("AllowIP() = %v, %v", allowed, err)
	}
	if routeKey == "" || routeKey != captured.DispatchKey || strings.Contains(routeKey, "192.0.2.10") {
		t.Fatalf("route/dispatch key = %q", routeKey)
	}
	if captured.ExecuteMode != rpcapi.RedisExecuteModeScript || captured.Script == nil ||
		len(captured.Script.Keys) != 1 || len(captured.Script.Args) != 3 {
		t.Fatalf("unexpected Redis request: %+v", captured)
	}
}

func TestConcurrencyLimiterRejectsWithoutQueueing(t *testing.T) {
	limiter := NewLimiter(Config{
		Enabled: true, Concurrency: ConcurrencyLimitConfig{Enabled: true, MaxInFlight: 1},
	}, nil)
	release, ok := limiter.AcquireConcurrency()
	if !ok {
		t.Fatal("first AcquireConcurrency() rejected")
	}
	if _, second := limiter.AcquireConcurrency(); second {
		t.Fatal("second AcquireConcurrency() exceeded limit")
	}
	release()
	secondRelease, ok := limiter.AcquireConcurrency()
	if !ok {
		t.Fatal("AcquireConcurrency() did not recover after release")
	}
	secondRelease()
}

func TestAccountKeyDoesNotContainAccountID(t *testing.T) {
	key := AccountKey("0123456789abcdef01234567")
	if key == "" || strings.Contains(key, "0123456789abcdef01234567") {
		t.Fatalf("AccountKey() = %q", key)
	}
}

func integerResult(value int64) rpcapi.RedisResult {
	return rpcapi.RedisResult{Results: []rpcapi.RedisCommandResult{{
		Status: rpcapi.RedisCommandStatusSucceeded,
		Value: rpcapi.RedisValue{RootIndex: 0, Nodes: []rpcapi.RedisValueNode{{
			Kind: rpcapi.RedisValueKindInteger, Integer: value,
		}}},
	}}}
}
