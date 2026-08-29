// Package ratelimit 实现登录入口的共享滑动窗口和本地并发限制。
package ratelimit

import (
	"fmt"

	originconfig "github.com/duanhf2012/origin/v3/config"
)

// Config 保存首期三个独立启停的登录限流维度。
type Config struct {
	Enabled     bool                   `json:"enabled"`
	IP          WindowLimitConfig      `json:"ip"`
	Identity    WindowLimitConfig      `json:"identity"`
	Concurrency ConcurrencyLimitConfig `json:"concurrency"`
}

// WindowLimitConfig 保存单个共享滑动窗口规则。
type WindowLimitConfig struct {
	Enabled     bool                  `json:"enabled"`
	Window      originconfig.Duration `json:"window"`
	MaxRequests int                   `json:"max_requests"`
}

// ConcurrencyLimitConfig 保存单实例在途登录硬上限。
type ConcurrencyLimitConfig struct {
	Enabled     bool `json:"enabled"`
	MaxInFlight int  `json:"max_in_flight"`
}

// ValidateConfig 校验当前启用的限流边界。
func ValidateConfig(config Config) error {
	if !config.Enabled {
		return nil
	}
	for name, limit := range map[string]WindowLimitConfig{"ip": config.IP, "identity": config.Identity} {
		if limit.Enabled && (limit.Window.Duration() <= 0 || limit.MaxRequests <= 0) {
			return fmt.Errorf("login_rate_limit.%s 窗口时长和次数必须为正数", name)
		}
	}
	if config.Concurrency.Enabled && config.Concurrency.MaxInFlight <= 0 {
		return fmt.Errorf("login_rate_limit.concurrency.max_in_flight 必须为正数")
	}
	return nil
}
