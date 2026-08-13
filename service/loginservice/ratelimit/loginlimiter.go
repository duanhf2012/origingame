package ratelimit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/duanhf2012/origin/v3/sysmodule/redismodule"
	"github.com/redis/go-redis/v9"
	"origingame/service/loginservice/account"
)

// Redis 脚本在一个键中保存所有窗口需要的时间点；毫秒相同的请求通过独立 member 避免覆盖。
const slidingWindowScript = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local longest = tonumber(ARGV[2])
local member = ARGV[3]
redis.call('ZREMRANGEBYSCORE', key, '-inf', now - longest)
for i = 4, #ARGV, 2 do
  local duration = tonumber(ARGV[i])
  local maximum = tonumber(ARGV[i + 1])
  local count = redis.call('ZCOUNT', key, now - duration, '+inf')
  if count >= maximum then
    return 0
  end
end
redis.call('ZADD', key, now, member)
redis.call('PEXPIRE', key, longest)
return 1
`

const maxLocalRateLimitKeys = 100000

// Limiter 组合 Redis 全局滑动窗口、本地降级窗口和单实例并发限制。
type Limiter struct {
	config   Config
	redis    *redismodule.Module
	sequence atomic.Uint64
	localMu  sync.Mutex
	local    map[string]localWindow
	inFlight chan struct{}
	now      func() time.Time
}

type localWindow struct {
	records []time.Time
	longest time.Duration
}

// New 创建有界限流状态；禁用并发限制时不分配 Channel。
func New(config Config, redisModule *redismodule.Module) *Limiter {
	limiter := &Limiter{config: config, redis: redisModule, local: make(map[string]localWindow), now: time.Now}
	if config.Enabled && config.Concurrency.Enabled {
		limiter.inFlight = make(chan struct{}, config.Concurrency.MaxInFlight)
	}
	return limiter
}

// AcquireConcurrency 尝试占用本实例登录槽位；返回的释放函数可以安全重复调用。
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

// AllowWindows 对指定维度执行所有已配置窗口；Redis 故障行为由配置明确决定。
func (limiter *Limiter) AllowWindows(
	ctx context.Context,
	dimension string,
	identity string,
	config WindowLimitConfig,
) (bool, error) {
	if limiter == nil || !limiter.config.Enabled || !config.Enabled {
		return true, nil
	}
	identity = strings.TrimSpace(identity)
	if identity == "" {
		identity = "unknown"
	}
	key := "origingame:login-rate:" + dimension + ":" + identity
	if limiter.redis != nil && limiter.redis.Client() != nil {
		allowed, err := limiter.allowRedis(ctx, key, config)
		if err == nil {
			return allowed, nil
		}
		if limiter.config.RedisFailureMode == "deny" {
			return false, fmt.Errorf("redis rate limit: %w", err)
		}
	}
	return limiter.allowLocal(key, config), nil
}

func (limiter *Limiter) allowRedis(ctx context.Context, key string, config WindowLimitConfig) (bool, error) {
	// 限流窗口是基础设施安全时间，必须使用不受 GM 调时影响的真实时间。
	now := limiter.now().UTC()
	longest := maximumWindow(config)
	arguments := make([]any, 0, 3+len(config.Windows)*2)
	member := strconv.FormatInt(now.UnixMilli(), 10) + "-" + strconv.FormatUint(limiter.sequence.Add(1), 10)
	arguments = append(arguments, now.UnixMilli(), longest.Milliseconds(), member)
	for _, window := range config.Windows {
		arguments = append(arguments, window.Duration.Duration().Milliseconds(), window.MaxRequests)
	}
	result, err := limiter.redis.Client().Eval(ctx, slidingWindowScript, []string{key}, arguments...).Int64()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

func (limiter *Limiter) allowLocal(key string, config WindowLimitConfig) bool {
	now := limiter.now().UTC()
	longest := maximumWindow(config)
	limiter.localMu.Lock()
	defer limiter.localMu.Unlock()
	entry, exists := limiter.local[key]
	if !exists && len(limiter.local) >= maxLocalRateLimitKeys {
		limiter.cleanupLocal(now)
		if len(limiter.local) >= maxLocalRateLimitKeys {
			return false
		}
	}
	records := entry.records
	cutoff := now.Add(-longest)
	first := 0
	for first < len(records) && !records[first].After(cutoff) {
		first++
	}
	records = append([]time.Time(nil), records[first:]...)
	for _, window := range config.Windows {
		windowCutoff := now.Add(-window.Duration.Duration())
		count := 0
		for index := len(records) - 1; index >= 0 && !records[index].Before(windowCutoff); index-- {
			count++
		}
		if count >= window.MaxRequests {
			limiter.local[key] = localWindow{records: records, longest: longest}
			return false
		}
	}
	limiter.local[key] = localWindow{records: append(records, now), longest: longest}
	return true
}

func (limiter *Limiter) cleanupLocal(now time.Time) {
	for key, entry := range limiter.local {
		if len(entry.records) == 0 || !entry.records[len(entry.records)-1].After(now.Add(-entry.longest)) {
			delete(limiter.local, key)
		}
	}
}

// IdentityKey 对平台身份做不可逆摘要，限流键不保存原始 PlatId。
func IdentityKey(clientIP string, platform account.LoginType, platformID string) string {
	sum := sha256.Sum256([]byte(strconv.FormatInt(int64(platform), 10) + "\x00" + platformID))
	return clientIP + ":" + hex.EncodeToString(sum[:])
}

// RecordAuthFailure 在真实 SDK 鉴权接入后用于记录失败并建立冷却；开发透传不会调用它。
func (limiter *Limiter) RecordAuthFailure(ctx context.Context, identity string) error {
	if limiter == nil || !limiter.config.Enabled || !limiter.config.AuthFailure.Enabled {
		return nil
	}
	if limiter.redis == nil || limiter.redis.Client() == nil {
		return errors.New("redis unavailable")
	}
	config := limiter.config.AuthFailure
	key := "origingame:login-auth-failure:" + identity
	count, err := limiter.redis.Client().Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		if err = limiter.redis.Client().Expire(ctx, key, config.Window.Duration()).Err(); err != nil {
			return err
		}
	}
	if count >= int64(config.MaxFailures) {
		return limiter.redis.Client().Set(ctx, key+":cooldown", "1", config.Cooldown.Duration()).Err()
	}
	return nil
}

// AuthCoolingDown 报告当前身份是否处于真实鉴权失败冷却期。
func (limiter *Limiter) AuthCoolingDown(ctx context.Context, identity string) (bool, error) {
	if limiter == nil || !limiter.config.Enabled || !limiter.config.AuthFailure.Enabled ||
		limiter.redis == nil || limiter.redis.Client() == nil {
		return false, nil
	}
	_, err := limiter.redis.Client().Get(ctx, "origingame:login-auth-failure:"+identity+":cooldown").Result()
	if err == nil {
		return true, nil
	}
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	return false, err
}
