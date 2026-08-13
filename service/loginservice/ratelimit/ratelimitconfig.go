// Package ratelimit 实现登录接口的共享滑动窗口、鉴权失败冷却和并发限制。
package ratelimit

import (
	"fmt"
	"time"

	originconfig "github.com/duanhf2012/origin/v3/config"
)

// Config 保存每个维度可独立关闭的登录限流配置。
type Config struct {
	Enabled          bool
	IP               WindowLimitConfig
	Identity         WindowLimitConfig
	AuthFailure      FailureLimitConfig
	Concurrency      ConcurrencyLimitConfig
	RedisFailureMode string
}

// WindowLimitConfig 保存一组共享滑动记录的时间窗口规则。
type WindowLimitConfig struct {
	Enabled bool
	Windows []WindowConfig
}

// WindowConfig 是单条滑动窗口限制。
type WindowConfig struct {
	Duration    originconfig.Duration
	MaxRequests int
}

// FailureLimitConfig 保存真实鉴权失败后的短期冷却规则。
type FailureLimitConfig struct {
	Enabled     bool
	Window      originconfig.Duration
	MaxFailures int
	Cooldown    originconfig.Duration
}

// ConcurrencyLimitConfig 保存单实例在途登录硬上限。
type ConcurrencyLimitConfig struct {
	Enabled     bool
	MaxInFlight int
}

// ValidateConfig 校验限流业务约束。
func ValidateConfig(config Config) error {
	if !config.Enabled {
		return nil
	}
	for name, limit := range map[string]WindowLimitConfig{"ip": config.IP, "identity": config.Identity} {
		if !limit.Enabled {
			continue
		}
		if len(limit.Windows) == 0 {
			return fmt.Errorf("login_rate_limit.%s.windows 不能为空", name)
		}
		for _, window := range limit.Windows {
			if window.Duration.Duration() <= 0 || window.MaxRequests <= 0 {
				return fmt.Errorf("login_rate_limit.%s 窗口时长和次数必须为正数", name)
			}
		}
	}
	if config.AuthFailure.Enabled && (config.AuthFailure.Window.Duration() <= 0 ||
		config.AuthFailure.Cooldown.Duration() <= 0 || config.AuthFailure.MaxFailures <= 0) {
		return fmt.Errorf("login_rate_limit.auth_failure 配置无效")
	}
	if config.Concurrency.Enabled && config.Concurrency.MaxInFlight <= 0 {
		return fmt.Errorf("login_rate_limit.concurrency.max_in_flight 必须为正数")
	}
	if config.RedisFailureMode != "local" && config.RedisFailureMode != "deny" {
		return fmt.Errorf("login_rate_limit.redis_failure_mode 只支持 local 或 deny")
	}
	return nil
}

func maximumWindow(config WindowLimitConfig) time.Duration {
	var maximum time.Duration
	for _, window := range config.Windows {
		if current := window.Duration.Duration(); current > maximum {
			maximum = current
		}
	}
	return maximum
}
