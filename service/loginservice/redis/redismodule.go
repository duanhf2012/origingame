// Package redis 持有 LoginService 限流所需的 Redis 生命周期和降级策略。
package redis

import (
	"context"

	"github.com/duanhf2012/origin/v3/log"
	"github.com/duanhf2012/origin/v3/sysmodule/redismodule"
)

// Module 为多实例登录限流提供共享 Redis 生命周期。
type Module struct {
	redismodule.Module
	config      redismodule.Config
	failureMode string
}

// NewModule 创建 Redis 生命周期 Module。
func NewModule(config redismodule.Config, failureMode string) *Module {
	return &Module{config: config, failureMode: failureMode}
}

// OnInit 冻结 Redis 拓扑和连接池配置。
func (module *Module) OnInit() error {
	return module.Setup(module.config)
}

// OnStart 在 local 模式允许 Redis 不可用时降级为实例本地窗口。
func (module *Module) OnStart(ctx context.Context) error {
	err := module.Module.OnStart(ctx)
	if err == nil || module.failureMode == "deny" {
		return err
	}
	module.Logger().Warn("Redis 启动失败，登录限流降级为实例本地窗口", log.Err(err))
	return nil
}

// OnStop 关闭正常启动的 Redis Client；降级状态下基类安全返回。
func (module *Module) OnStop(ctx context.Context) error {
	return module.Module.OnStop(ctx)
}
