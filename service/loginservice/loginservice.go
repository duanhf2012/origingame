// Package loginservice 只保留 LoginService 主类型、配置聚合、生命周期装配和登录协调流程。
package loginservice

import (
	"context"
	"fmt"
	"net/http"
	"time"

	originconfig "github.com/duanhf2012/origin/v3/config"
	"github.com/duanhf2012/origin/v3/log"
	"github.com/duanhf2012/origin/v3/service"
	"github.com/duanhf2012/origin/v3/sysmodule/ginmodule"
	"github.com/duanhf2012/origin/v3/sysmodule/mongodbmodule"
	"github.com/duanhf2012/origin/v3/sysmodule/redismodule"
	"origingame/internal/mongodb"
	"origingame/internal/security"
	commonpb "origingame/protocol/common"
	"origingame/service/loginservice/account"
	"origingame/service/loginservice/area"
	"origingame/service/loginservice/authentication"
	"origingame/service/loginservice/database"
	"origingame/service/loginservice/httpapi"
	"origingame/service/loginservice/ratelimit"
	loginredis "origingame/service/loginservice/redis"
)

// Config 是 LoginService 完整配置；严格解析会拒绝拼写错误和未知字段。
type Config struct {
	HTTP           HTTPConfig
	MongoDB        mongodbmodule.Config
	Redis          redismodule.Config
	Token          TokenConfig
	Area           AreaConfig
	Authentication AuthenticationConfig
	LoginRateLimit ratelimit.Config
}

// HTTPConfig 保存对外 HTTP Server 配置。
type HTTPConfig struct {
	Server ginmodule.ServerConfig
}

// TokenConfig 保存 Ed25519 JWT 签发配置。
type TokenConfig struct {
	Issuer     string
	Audience   string
	Expire     originconfig.Duration
	ActiveKID  string
	PrivateKey string
}

// AreaConfig 保存区服快照的自动刷新周期。
type AreaConfig struct {
	RefreshInterval originconfig.Duration
}

// AuthenticationConfig 明确标记当前开发鉴权不会访问第三方平台。
type AuthenticationConfig struct {
	DevelopmentPassthrough bool
}

// LoginService 是 LoginServer 中唯一的业务 Service。
type LoginService struct {
	service.Service
	config        Config
	mongo         *database.MongoModule
	redis         *loginredis.Module
	http          *httpapi.Module
	refresh       *area.RefreshModule
	catalog       area.Catalog
	authenticator authentication.Authenticator
	issuer        *security.TokenIssuer
	limiter       *ratelimit.Limiter
}

// OnInit 严格解析配置，并按数据库、Redis、HTTP 的启动依赖顺序登记 Module。
func (target *LoginService) OnInit() error {
	if err := target.loadConfig(); err != nil {
		return err
	}
	if err := target.config.validate(); err != nil {
		return err
	}
	if err := target.SetDefaultAwaitTimeout(15 * time.Second); err != nil {
		return err
	}

	// Token 私钥在 HTTP 监听前完成校验；错误不得包含私钥原文。
	issuer, err := security.NewTokenIssuer(security.TokenConfig{
		Issuer: target.config.Token.Issuer, Audience: target.config.Token.Audience,
		Expire: target.config.Token.Expire.Duration(), ActiveKID: target.config.Token.ActiveKID,
		PrivateKey: target.config.Token.PrivateKey,
	})
	if err != nil {
		return fmt.Errorf("初始化 TokenIssuer: %w", err)
	}
	target.issuer = issuer
	target.authenticator = authentication.DevelopmentAuthenticator{}

	// 数据库 Module 完成 Client、索引和初始区服数据准备，是 HTTP Ready 的硬前置。
	target.mongo = database.NewMongoModule(target.config.MongoDB, &target.catalog)
	if err = target.AddModule(target.mongo); err != nil {
		return err
	}
	target.refresh = area.NewRefreshModule(
		target.config.Area.RefreshInterval.Duration(), target.mongo.Areas(), &target.catalog,
	)
	if err = target.AddModule(target.refresh); err != nil {
		return err
	}

	// 启用总限流时 Redis 提供跨实例共享窗口；禁用时不建立无用连接。
	if target.config.LoginRateLimit.Enabled {
		target.redis = loginredis.NewModule(
			target.config.Redis,
			target.config.LoginRateLimit.RedisFailureMode,
		)
		if err = target.AddModule(target.redis); err != nil {
			return err
		}
		target.limiter = ratelimit.New(target.config.LoginRateLimit, &target.redis.Module)
	} else {
		target.limiter = ratelimit.New(target.config.LoginRateLimit, nil)
	}

	// HTTP Module 最后登记，只有关键数据和依赖全部准备完成后才开始监听。
	target.http = httpapi.NewModule(
		target.config.HTTP.Server,
		target.login,
		target.limiter.AcquireConcurrency,
	)
	return target.AddModule(target.http)
}

// OnStart 无额外资源；所有依赖按 Module 顺序启动，HTTP 始终最后监听。
func (target *LoginService) OnStart(context.Context) error {
	return nil
}

func (target *LoginService) login(ctx *ginmodule.SafeContext) {
	// 1. 严格绑定 JSON 并校验老版本请求字段，不把详细解析错误返回客户端。
	var request httpapi.LoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		target.respondError(ctx, http.StatusBadRequest, commonpb.ErrorCode_ERROR_CODE_INVALID_REQUEST)
		return
	}
	identity := request.Identity()
	if !account.ValidLoginType(identity.PlatType) {
		target.respondError(ctx, http.StatusOK, commonpb.ErrorCode_ERROR_CODE_LOGIN_PLATFORM_TYPE_INVALID)
		return
	}
	if identity.PlatID == "" {
		target.respondError(ctx, http.StatusOK, commonpb.ErrorCode_ERROR_CODE_LOGIN_PLATFORM_ID_INVALID)
		return
	}

	// 2. 对 IP 和不可逆平台身份摘要分别执行 Redis 全局滑动窗口。
	requestCtx := ctx.Context()
	ok, err := target.allowWindow(requestCtx, "ip", ctx.ClientIP(), target.config.LoginRateLimit.IP)
	if err != nil || !ok {
		target.respondError(ctx, http.StatusTooManyRequests, commonpb.ErrorCode_ERROR_CODE_TOO_MANY_REQUESTS)
		return
	}
	limitIdentity := ratelimit.IdentityKey(ctx.ClientIP(), identity.PlatType, identity.PlatID)
	ok, err = target.allowWindow(requestCtx, "identity", limitIdentity, target.config.LoginRateLimit.Identity)
	if err != nil || !ok {
		target.respondError(ctx, http.StatusTooManyRequests, commonpb.ErrorCode_ERROR_CODE_TOO_MANY_REQUESTS)
		return
	}
	var cooling bool
	err = target.Await(requestCtx, func(waitCtx context.Context) error {
		var lookupErr error
		cooling, lookupErr = target.limiter.AuthCoolingDown(waitCtx, limitIdentity)
		return lookupErr
	})
	if err != nil && target.config.LoginRateLimit.RedisFailureMode == "local" {
		cooling, err = false, nil
	}
	if err != nil || cooling {
		target.respondError(ctx, http.StatusTooManyRequests, commonpb.ErrorCode_ERROR_CODE_TOO_MANY_REQUESTS)
		return
	}

	// 3. 当前开发实现明确不访问第三方 SDK；正式接入失败时记录失败冷却。
	if err = target.authenticator.Authenticate(requestCtx, identity); err != nil {
		_ = target.Await(requestCtx, func(waitCtx context.Context) error {
			return target.limiter.RecordAuthFailure(waitCtx, limitIdentity)
		})
		target.respondError(ctx, http.StatusOK, commonpb.ErrorCode_ERROR_CODE_LOGIN_AUTH_FAILED)
		return
	}

	// 4. 原子获取账号并签发 JWT。数据库等待释放 Service 执行权，避免串行队列被 I/O 阻塞。
	var document mongodb.Account
	err = target.Await(requestCtx, func(waitCtx context.Context) error {
		var repositoryErr error
		document, repositoryErr = target.mongo.Accounts().FindOrCreate(waitCtx, identity, ctx.ClientIP())
		return repositoryErr
	})
	if err != nil {
		target.Logger().Error("登录账号查询或创建失败", log.Int32("platform_type", int32(identity.PlatType)), log.Err(err))
		target.respondError(ctx, http.StatusOK, commonpb.ErrorCode_ERROR_CODE_LOGIN_ACCOUNT_FAILED)
		return
	}
	token, err := target.issuer.Issue(document.ID.Hex(), int32(identity.PlatType))
	if err != nil {
		target.Logger().Error("签发游戏 Token 失败", log.Err(err))
		target.respondError(ctx, http.StatusInternalServerError, commonpb.ErrorCode_ERROR_CODE_INTERNAL)
		return
	}

	// 5. 区服快照启动时已准备完成；运行期异常不能把它替换成空列表。
	areas := target.catalog.Snapshot()
	if len(areas) == 0 {
		target.respondError(ctx, http.StatusServiceUnavailable, commonpb.ErrorCode_ERROR_CODE_LOGIN_NO_AVAILABLE_AREA)
		return
	}
	ctx.JSON(http.StatusOK, httpapi.LoginResponse{
		ECode: int32(commonpb.ErrorCode_ERROR_CODE_OK), Token: token, AreaList: areas,
	})
}

func (target *LoginService) respondError(
	ctx *ginmodule.SafeContext,
	status int,
	code commonpb.ErrorCode,
) {
	ctx.JSON(status, httpapi.LoginResponse{ECode: int32(code)})
}

func (target *LoginService) allowWindow(
	ctx context.Context,
	dimension string,
	identity string,
	config ratelimit.WindowLimitConfig,
) (bool, error) {
	allowed := false
	err := target.Await(ctx, func(waitCtx context.Context) error {
		var limitErr error
		allowed, limitErr = target.limiter.AllowWindows(waitCtx, dimension, identity, config)
		return limitErr
	})
	return allowed, err
}

func (target *LoginService) loadConfig() error {
	target.config = defaultConfig()
	sections := []struct {
		path        string
		destination any
	}{
		{"http", &target.config.HTTP}, {"mongodb", &target.config.MongoDB},
		{"redis", &target.config.Redis}, {"token", &target.config.Token}, {"area", &target.config.Area},
		{"authentication", &target.config.Authentication}, {"login_rate_limit", &target.config.LoginRateLimit},
	}
	for _, section := range sections {
		if err := target.GetServiceConfigStrict(section.path, section.destination); err != nil {
			return fmt.Errorf("读取 %s 配置: %w", section.path, err)
		}
	}
	return nil
}

func (config Config) validate() error {
	if config.Area.RefreshInterval.Duration() <= 0 {
		return fmt.Errorf("area.refresh_interval 必须为正数")
	}
	if !config.Authentication.DevelopmentPassthrough {
		return fmt.Errorf("当前只实现 development_passthrough 鉴权；关闭后必须先接入真实 SDK")
	}
	if config.Token.Expire.Duration() <= 0 {
		return fmt.Errorf("token.expire 必须为正数")
	}
	return ratelimit.ValidateConfig(config.LoginRateLimit)
}
