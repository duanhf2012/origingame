package ratelimit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"origingame/internal/redisscripts"
	rpcapi "origingame/protocol/rpc"
)

// RedisExecutor 通过指定路由 Key 调用公共 AccDBService。
type RedisExecutor func(context.Context, string, rpcapi.RedisRequest) (rpcapi.RedisResult, error)

// Limiter 组合两个 AccDBService 共享窗口和单实例并发限制。
type Limiter struct {
	config   Config
	execute  RedisExecutor
	inFlight chan struct{}
	now      func() time.Time
}

// New 创建有界限流状态；禁用并发限制时不分配 Channel。
func New(config Config, execute RedisExecutor) *Limiter {
	limiter := &Limiter{config: config, execute: execute, now: time.Now}
	if config.Enabled && config.Concurrency.Enabled {
		limiter.inFlight = make(chan struct{}, config.Concurrency.MaxInFlight)
	}
	return limiter
}

// AcquireConcurrency 尝试占用本实例登录槽位；满载时立即拒绝而不排队。
func (limiter *Limiter) AcquireConcurrency() (func(), bool) {
	if limiter == nil || !limiter.config.Enabled || limiter.inFlight == nil {
		return func() {}, true
	}
	select {
	case limiter.inFlight <- struct{}{}:
		var released atomic.Bool
		return func() {
			if released.CompareAndSwap(false, true) {
				<-limiter.inFlight
			}
		}, true
	default:
		return nil, false
	}
}

// AllowIP 在平台鉴权前执行来源 IP 窗口。
func (limiter *Limiter) AllowIP(ctx context.Context, clientIP string) (bool, error) {
	return limiter.allow(ctx, limitKey("ip", clientIP), limiter.config.IP)
}

// AllowIdentity 在取得可信 AccountID 后执行账号身份窗口。
func (limiter *Limiter) AllowIdentity(ctx context.Context, accountID string) (bool, error) {
	return limiter.allow(ctx, AccountKey(accountID), limiter.config.Identity)
}

func (limiter *Limiter) allow(ctx context.Context, key string, config WindowLimitConfig) (bool, error) {
	if limiter == nil || !limiter.config.Enabled || !config.Enabled {
		return true, nil
	}
	if limiter.execute == nil {
		return false, errors.New("登录限流 AccDBService 未初始化")
	}
	request := rpcapi.RedisRequest{
		DispatchKey: key,
		ExecuteMode: rpcapi.RedisExecuteModeScript,
		Script: &rpcapi.RedisScriptCall{
			ID:   redisscripts.LoginRateLimitID,
			Keys: []string{key},
			Args: [][]byte{
				[]byte(strconv.FormatInt(limiter.now().UTC().UnixMilli(), 10)),
				[]byte(strconv.FormatInt(config.Window.Duration().Milliseconds(), 10)),
				[]byte(strconv.Itoa(config.MaxRequests)),
			},
		},
	}
	result, err := limiter.execute(ctx, key, request)
	if err != nil {
		return false, err
	}
	if result.Failure != nil || len(result.Results) != 1 ||
		result.Results[0].Status != rpcapi.RedisCommandStatusSucceeded {
		return false, errors.New("登录限流 Redis 操作失败")
	}
	value := result.Results[0].Value
	if int(value.RootIndex) >= len(value.Nodes) {
		return false, errors.New("登录限流 Redis 返回结构无效")
	}
	node := value.Nodes[value.RootIndex]
	if node.Kind != rpcapi.RedisValueKindInteger || (node.Integer != 0 && node.Integer != 1) {
		return false, fmt.Errorf("登录限流 Redis 返回类型无效: kind=%d", node.Kind)
	}
	return node.Integer == 1, nil
}

// AccountKey 返回不包含原始 AccountID 的身份限流键。
func AccountKey(accountID string) string {
	return limitKey("identity", accountID)
}

func limitKey(dimension string, identity string) string {
	identity = strings.TrimSpace(identity)
	if identity == "" {
		identity = "unknown"
	}
	sum := sha256.Sum256([]byte(dimension + "\x00" + identity))
	return "login-rate:" + dimension + ":" + hex.EncodeToString(sum[:])
}
