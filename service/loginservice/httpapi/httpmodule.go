package httpapi

import (
	"context"
	"net/http"

	"github.com/duanhf2012/origin/v3/sysmodule/ginmodule"
	"github.com/gin-gonic/gin"
	"origingame/internal/security"
	commonpb "origingame/protocol/common"
	"origingame/service/loginservice/account"
	"origingame/service/loginservice/area"
	"origingame/service/loginservice/authentication"
	"origingame/service/loginservice/ratelimit"
)

// Dependencies 是 HTTP 登录流程直接使用的最小业务能力集合。
type Dependencies struct {
	Authenticator authentication.Authenticator
	Accounts      *account.MongoRepository
	Issuer        *security.TokenIssuer
	Catalog       *area.Catalog
	Limiter       *ratelimit.Limiter
}

// Module 只在所有前置 Module 启动成功后绑定监听地址。
type Module struct {
	ginmodule.Module
	config ginmodule.ServerConfig
	deps   Dependencies
}

// NewModule 注入业务能力；路由和 Handler 始终由 HTTP Module 自己拥有。
func NewModule(config ginmodule.ServerConfig, dependencies Dependencies) *Module {
	return &Module{config: config, deps: dependencies}
}

// OnInit 冻结 HTTP 安全边界并注册路由。
func (module *Module) OnInit() error {
	options, err := module.config.Options()
	if err != nil {
		return err
	}
	if err = module.Setup(module.config.Address, options); err != nil {
		return err
	}
	module.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	module.Use(module.limitConcurrency())
	module.SafePOST("/api/v1/login", module.login)
	return nil
}

func (module *Module) limitConcurrency() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 健康检查不占登录容量；登录在请求 goroutine 入队前抢占槽位，超限立即返回。
		if ctx.Request.URL.Path != "/api/v1/login" {
			ctx.Next()
			return
		}
		release, allowed := module.deps.Limiter.AcquireConcurrency()
		if !allowed {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, LoginResponse{
				ECode: int32(commonpb.ErrorCode_ERROR_CODE_TOO_MANY_REQUESTS),
			})
			return
		}
		defer release()
		ctx.Next()
	}
}

// OnStart 同步完成真实监听；Module 最后装配以保证关键依赖已经 Ready。
func (module *Module) OnStart(ctx context.Context) error {
	return module.Module.OnStart(ctx)
}
