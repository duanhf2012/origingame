// Package ratelimit 实现登录入口的共享滑动窗口和本地并发限制。
package ratelimit

import (
	"fmt"

	originconfig "github.com/duanhf2012/origin/v3/config"
)

// Config 保存首期三个独立启停的登录限流维度。
type Config struct {
	Enabled     bool                   `json:"enabled"`     // 是否启用登录限流。
	IP          WindowLimitConfig      `json:"ip"`          // 按客户端 IP 限制。
	Identity    WindowLimitConfig      `json:"identity"`    // 按账号身份限制。
	Concurrency ConcurrencyLimitConfig `json:"concurrency"` // 单实例并发限制。
}

// WindowLimitConfig 保存单个共享滑动窗口规则。
type WindowLimitConfig struct {
	Enabled     bool                  `json:"enabled"`      // 是否启用该窗口限制。
	Window      originconfig.Duration `json:"window"`       // 真实系统时间窗口长度。
	MaxRequests int                   `json:"max_requests"` // 窗口内最大请求数。
}

// ConcurrencyLimitConfig 保存单实例在途登录硬上限。
type ConcurrencyLimitConfig struct {
	Enabled     bool `json:"enabled"`       // 是否启用并发限制。
	MaxInFlight int  `json:"max_in_flight"` // 最大在途登录数。
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
