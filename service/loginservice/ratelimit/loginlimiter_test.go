package ratelimit

import (
	"testing"
	"time"

	originconfig "github.com/duanhf2012/origin/v3/config"
	"origingame/service/loginservice/account"
)

// TestLocalSlidingWindow 验证窗口内超限、时间推进后释放和不同键隔离。
func TestLocalSlidingWindow(t *testing.T) {
	current := time.Date(2026, 8, 12, 20, 0, 0, 0, time.UTC)
	limiter := New(Config{Enabled: true}, nil)
	limiter.now = func() time.Time { return current }
	config := WindowLimitConfig{Enabled: true, Windows: []WindowConfig{{
		Duration: originconfig.Duration(time.Second), MaxRequests: 2,
	}}}
	if !limiter.allowLocal("one", config) || !limiter.allowLocal("one", config) || limiter.allowLocal("one", config) {
		t.Fatal("local sliding window did not enforce maximum")
	}
	if !limiter.allowLocal("two", config) {
		t.Fatal("different rate limit keys interfered")
	}
	current = current.Add(time.Second + time.Millisecond)
	if !limiter.allowLocal("one", config) {
		t.Fatal("expired records were not released")
	}
}

// TestConcurrencyLimiter 验证槽位满时立即拒绝，释放后可以再次进入。
func TestConcurrencyLimiter(t *testing.T) {
	limiter := New(Config{
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

// TestIdentityKeyDoesNotContainPlatformID 保证 Redis Key 不泄露原始平台身份。
func TestIdentityKeyDoesNotContainPlatformID(t *testing.T) {
	key := IdentityKey("127.0.0.1", account.LoginTypeGuest, "secret-platform-id")
	if key == "" || key == "127.0.0.1:secret-platform-id" {
		t.Fatalf("IdentityKey() = %q", key)
	}
}
