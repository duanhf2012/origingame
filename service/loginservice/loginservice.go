// Package loginservice 只保留 LoginService 配置、生命周期装配和跨包启动检查。
package loginservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	originconfig "github.com/duanhf2012/origin/v3/config"
	"github.com/duanhf2012/origin/v3/service"
	"github.com/duanhf2012/origin/v3/sysmodule/ginmodule"
	"go.mongodb.org/mongo-driver/v2/bson"
	"origingame/internal/dbexecutor"
	"origingame/internal/mongodb"
	"origingame/internal/security"
	rpcapi "origingame/protocol/rpc"
	"origingame/service/loginservice/account"
	"origingame/service/loginservice/area"
	"origingame/service/loginservice/authentication"
	"origingame/service/loginservice/httpapi"
	"origingame/service/loginservice/ratelimit"
)

const loginReadyDispatchKey = "login-ready"

// Config 是 LoginService 完整配置；数据库连接配置只属于 AccDBService。
type Config struct {
	HTTP           HTTPConfig           `json:"http"`             // 对外 HTTP 配置。
	Token          TokenConfig          `json:"token"`            // Token 签发配置。
	Area           AreaConfig           `json:"area"`             // 区服目录配置。
	Authentication AuthenticationConfig `json:"authentication"`   // SDK 鉴权配置。
	LoginRateLimit ratelimit.Config     `json:"login_rate_limit"` // 登录限流配置。
}

// HTTPConfig 保存对外 HTTP Server 配置。
type HTTPConfig struct {
	Server ginmodule.ServerConfig `json:"server"` // HTTP 监听配置。
}

// TokenConfig 保存 Ed25519 JWT 签发配置。
type TokenConfig struct {
	Issuer     string                `json:"issuer"`      // Token 签发方。
	Audience   string                `json:"audience"`    // Token 受众。
	Expire     originconfig.Duration `json:"expire"`      // Token 有效期。
	ActiveKID  string                `json:"active_kid"`  // 当前签发公钥标识。
	PrivateKey string                `json:"private_key"` // Base64 编码私钥材料。
}

// AreaConfig 保存区服快照的自动刷新周期。
type AreaConfig struct {
	RefreshInterval originconfig.Duration `json:"refresh_interval"` // 区服目录刷新间隔。
}

// AuthenticationConfig 明确标记当前开发鉴权不会访问第三方平台。
type AuthenticationConfig struct {
	DevelopmentPassthrough bool `json:"development_passthrough"` // 是否使用开发透传鉴权。
}

// LoginService 装配公共 AccDBService 客户端和登录边界 Module。
type LoginService struct {
	service.Service                          // Origin Service 生命周期能力。
	config          Config                   // LoginService 本地配置。
	accDB           rpcapi.DBServiceClient   // 公共数据域 RPC 客户端。
	http            *httpapi.LoginHTTPModule // HTTP 登录入口 Module。
	refresh         *area.AreaRefreshModule  // 区服目录刷新 Module。
	catalog         area.Catalog             // 当前显示区服目录。
}

// OnInit 严格解析配置，并按区服快照、HTTP 的依赖顺序登记 Module。
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

	issuer, err := security.NewTokenIssuer(security.TokenConfig{
		Issuer: target.config.Token.Issuer, Audience: target.config.Token.Audience,
		Expire: target.config.Token.Expire.Duration(), ActiveKID: target.config.Token.ActiveKID,
		PrivateKey: target.config.Token.PrivateKey,
	})
	if err != nil {
		return fmt.Errorf("初始化 TokenIssuer: %w", err)
	}
	authenticator := authentication.DevelopmentAuthenticator{}
	target.accDB = rpcapi.BindDBServiceTo(target, "AccDBService").WhereLabels(map[string]string{"scope": "pub"})
	awaitDB := dbexecutor.NewAwaitExecutor(target.accDB)
	callDB := dbexecutor.NewCallExecutor(target.accDB)

	accounts := account.NewMongoRepository(awaitDB)
	areas := area.NewMongoRepository(callDB)
	limiter := ratelimit.NewLimiter(target.config.LoginRateLimit, awaitDB)

	target.refresh = area.NewAreaRefreshModule(
		target.config.Area.RefreshInterval.Duration(), areas, &target.catalog,
	)
	if err = target.AddModule(target.refresh); err != nil {
		return err
	}
	target.http = httpapi.NewLoginHTTPModule(target.config.HTTP.Server, httpapi.Dependencies{
		Authenticator: authenticator,
		Accounts:      accounts,
		Issuer:        issuer,
		Catalog:       &target.catalog,
		Limiter:       limiter,
	})
	return target.AddModule(target.http)
}

// OnStart 在 HTTP Module 启动前确认账号集合和公共 Redis 均可访问。
func (target *LoginService) OnStart(ctx context.Context) error {
	filter, err := bson.Marshal(bson.D{})
	if err != nil {
		return err
	}
	mongoResult, err := target.accDB.Route(loginReadyDispatchKey).CallExecuteMongo(ctx, rpcapi.MongoRequest{
		DispatchKey: loginReadyDispatchKey,
		ExecuteMode: rpcapi.MongoExecuteModeSequential,
		Operations: []rpcapi.MongoOperation{{
			Kind: rpcapi.MongoOperationKindCountDocuments, Collection: mongodb.AccountName,
			CountDocuments: &rpcapi.MongoCountDocuments{Filter: filter},
		}},
	})
	if err != nil || mongoResult.Failure != nil || len(mongoResult.Results) != 1 ||
		mongoResult.Results[0].Status != rpcapi.MongoOperationStatusSucceeded {
		return dependencyNotReady("AccDBService 账号集合未就绪", err)
	}
	redisResult, err := target.accDB.Route(loginReadyDispatchKey).CallExecuteRedis(ctx, rpcapi.RedisRequest{
		DispatchKey: loginReadyDispatchKey,
		ExecuteMode: rpcapi.RedisExecuteModeCommand,
		Commands:    []rpcapi.RedisCommand{{Name: "EXISTS", Args: [][]byte{[]byte("login-ready-probe")}}},
	})
	if err != nil || redisResult.Failure != nil || len(redisResult.Results) != 1 ||
		redisResult.Results[0].Status != rpcapi.RedisCommandStatusSucceeded {
		return dependencyNotReady("AccDBService Redis 未就绪", err)
	}
	return nil
}

func dependencyNotReady(message string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w", message, err)
	}
	return errors.New(message)
}

func (target *LoginService) loadConfig() error {
	target.config = defaultConfig()
	sections := []struct {
		path        string // 配置节路径。
		destination any    // 配置接收目标。
	}{
		{"http", &target.config.HTTP},
		{"token", &target.config.Token},
		{"area", &target.config.Area},
		{"authentication", &target.config.Authentication},
		{"login_rate_limit", &target.config.LoginRateLimit},
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
