// Package dbservice 实现业务无关的 MongoDB 与 Redis 集中访问模板。
package dbservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/duanhf2012/origin/v3/errs"
	"github.com/duanhf2012/origin/v3/log"
	"github.com/duanhf2012/origin/v3/service"
	originmongo "github.com/duanhf2012/origin/v3/sysmodule/mongodbmodule"
	originredis "github.com/duanhf2012/origin/v3/sysmodule/redismodule"
	"origingame/internal/redisscripts"
	rpcapi "origingame/protocol/rpc"
	"origingame/service/dbservice/mongodbmodule"
	"origingame/service/dbservice/redismodule"
)

const (
	defaultMaxIOConcurrency    int64 = 64
	defaultMaxInflightRequests int64 = 128
	slowHandlerThreshold             = time.Second
)

// Config 是一个实际 DBService 实例的完整配置。
type Config struct {
	MaxIOConcurrency    int64
	MaxInflightRequests int64
	MongoDB             originmongo.Config
	Redis               originredis.Config
}

// DBService 是通过 Origin 模板别名实例化为 AccDBService 或 RoleDBService 的通用服务。
type DBService struct {
	service.Service
	config   Config
	executor *keyExecutor
	mongo    *mongodbmodule.Module
	redis    *redismodule.Module
}

var _ rpcapi.DBService = (*DBService)(nil)

func defaultConfig() Config {
	return Config{
		MaxIOConcurrency:    defaultMaxIOConcurrency,
		MaxInflightRequests: defaultMaxInflightRequests,
	}
}

// OnInit 读取实际 Service 配置，并按 MongoDB、Redis 顺序登记资源 Module。
func (target *DBService) OnInit() error {
	if err := target.loadConfig(); err != nil {
		return err
	}
	if err := validateConfig(target.Name(), target.config); err != nil {
		return err
	}
	executor, err := newKeyExecutor(target.config.MaxIOConcurrency, target.config.MaxInflightRequests)
	if err != nil {
		return err
	}
	target.executor = executor

	target.mongo = mongodbmodule.New(target.config.MongoDB)
	if err = target.AddModule(target.mongo); err != nil {
		return err
	}
	target.redis, err = redismodule.New(target.config.Redis, scriptDefinitions(target.Name()))
	if err != nil {
		return err
	}
	return target.AddModule(target.redis)
}

func scriptDefinitions(serviceName string) []redismodule.ScriptDefinition {
	if serviceName != "AccDBService" {
		return nil
	}
	return []redismodule.ScriptDefinition{{
		ID: redisscripts.LoginRateLimitID, Source: redisscripts.LoginRateLimitSource,
		MinKeys: 1, MaxKeys: 1, MaxArgs: 3, MaxResultNodes: 1,
	}, {
		ID: redisscripts.RegisterGameServiceID, Source: redisscripts.RegisterGameServiceSource,
		MinKeys: 3, MaxKeys: 3, MaxArgs: 5, MaxResultNodes: 1,
	}, {
		ID: redisscripts.SetGameServiceDrainingID, Source: redisscripts.SetGameServiceDrainingSource,
		MinKeys: 2, MaxKeys: 2, MaxArgs: 2, MaxResultNodes: 1,
	}, {
		ID: redisscripts.AssignOrGetPlayerID, Source: redisscripts.AssignOrGetPlayerSource,
		MinKeys: 2, MaxKeys: 2, MaxArgs: 11, MaxResultNodes: 6,
	}, {
		ID: redisscripts.BeginPlayerLoadID, Source: redisscripts.BeginPlayerLoadSource,
		MinKeys: 3, MaxKeys: 3, MaxArgs: 4, MaxResultNodes: 1,
	}, {
		ID: redisscripts.CompletePlayerLoginID, Source: redisscripts.CompletePlayerLoginSource,
		MinKeys: 3, MaxKeys: 3, MaxArgs: 4, MaxResultNodes: 1,
	}, {
		ID: redisscripts.ReleasePlayerLoadID, Source: redisscripts.ReleasePlayerLoadSource,
		MinKeys: 3, MaxKeys: 3, MaxArgs: 3, MaxResultNodes: 1,
	}, {
		ID: redisscripts.MarkPlayerResidentID, Source: redisscripts.MarkPlayerResidentSource,
		MinKeys: 3, MaxKeys: 3, MaxArgs: 3, MaxResultNodes: 1,
	}, {
		ID: redisscripts.BeginPlayerReleaseID, Source: redisscripts.BeginPlayerReleaseSource,
		MinKeys: 3, MaxKeys: 3, MaxArgs: 3, MaxResultNodes: 1,
	}, {
		ID: redisscripts.RenewPlayerRoutesID, Source: redisscripts.RenewPlayerRoutesSource,
		MinKeys: 1, MaxKeys: 256, MaxArgs: 3, MaxResultNodes: 1,
	}}
}

// OnStart 不创建额外资源；两个数据库 Module 均已完成探活后 Service 才会 Ready。
func (*DBService) OnStart(context.Context) error { return nil }

// ExecuteMongo 执行一个通用 MongoDB 请求。
func (target *DBService) ExecuteMongo(ctx context.Context, request rpcapi.MongoRequest) (rpcapi.MongoResult, error) {
	startedAt := time.Now()
	if err := mongodbmodule.ValidateRequest(request); err != nil {
		return rpcapi.MongoResult{}, err
	}
	reservation, ok := target.executor.tryReserve()
	if !ok {
		return rpcapi.MongoResult{}, errs.ErrServiceQueueFull
	}
	defer reservation.release()

	var result rpcapi.MongoResult
	err := target.Await(ctx, func(waitCtx context.Context) error {
		return target.executor.execute(waitCtx, request.DispatchKey, func(ioCtx context.Context) error {
			var executeErr error
			result, executeErr = target.mongo.Execute(ioCtx, request)
			return executeErr
		})
	})
	target.logSlowRequest("mongo", mongoModeName(request.ExecuteMode), len(request.Operations), startedAt, err)
	return result, err
}

// ExecuteRedis 执行一个通用 Redis 请求。
func (target *DBService) ExecuteRedis(ctx context.Context, request rpcapi.RedisRequest) (rpcapi.RedisResult, error) {
	startedAt := time.Now()
	if err := target.redis.ValidateRequest(request); err != nil {
		return rpcapi.RedisResult{}, err
	}
	reservation, ok := target.executor.tryReserve()
	if !ok {
		return rpcapi.RedisResult{}, errs.ErrServiceQueueFull
	}
	defer reservation.release()

	var result rpcapi.RedisResult
	err := target.Await(ctx, func(waitCtx context.Context) error {
		return target.executor.execute(waitCtx, request.DispatchKey, func(ioCtx context.Context) error {
			var executeErr error
			result, executeErr = target.redis.Execute(ioCtx, request)
			return executeErr
		})
	})
	operationCount := len(request.Commands)
	if request.ExecuteMode == rpcapi.RedisExecuteModeScript {
		operationCount = 1
	}
	target.logSlowRequest("redis", redisModeName(request.ExecuteMode), operationCount, startedAt, err)
	return result, err
}

func (target *DBService) loadConfig() error {
	target.config = defaultConfig()
	if err := target.loadOptionalInt64("max_io_concurrency", &target.config.MaxIOConcurrency); err != nil {
		return err
	}
	if err := target.loadOptionalInt64("max_inflight_requests", &target.config.MaxInflightRequests); err != nil {
		return err
	}
	if err := target.GetServiceConfigStrict("mongodb", &target.config.MongoDB); err != nil {
		return fmt.Errorf("读取 mongodb 配置: %w", err)
	}
	if err := target.GetServiceConfigStrict("redis", &target.config.Redis); err != nil {
		return fmt.Errorf("读取 redis 配置: %w", err)
	}
	return nil
}

func (target *DBService) loadOptionalInt64(path string, destination *int64) error {
	err := target.GetServiceConfigStrict(path, destination)
	if errors.Is(err, errs.ErrConfigNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("读取 %s 配置: %w", path, err)
	}
	return nil
}

func validateConfig(serviceName string, config Config) error {
	if config.MaxIOConcurrency <= 0 || config.MaxInflightRequests <= 0 ||
		config.MaxIOConcurrency > config.MaxInflightRequests || config.Redis.MaxRetries != 0 {
		return errs.NewMessage(errs.CodeInvalidConfig, "DBService 并发配置或 Redis 重试配置无效")
	}
	switch serviceName {
	case "AccDBService":
		if config.MongoDB.Database != "origingame_account" || config.Redis.Database != 0 {
			return errs.NewMessage(errs.CodeInvalidConfig, "AccDBService 数据域配置无效")
		}
	case "RoleDBService":
		if config.MongoDB.Database != "origingame_role" || config.Redis.Database != 1 {
			return errs.NewMessage(errs.CodeInvalidConfig, "RoleDBService 数据域配置无效")
		}
	default:
		return errs.NewMessage(errs.CodeInvalidConfig, "DBService 必须使用 AccDBService 或 RoleDBService 实际名称")
	}
	return nil
}

func (target *DBService) logSlowRequest(
	resource string,
	mode string,
	operationCount int,
	startedAt time.Time,
	err error,
) {
	duration := time.Since(startedAt)
	if duration < slowHandlerThreshold {
		return
	}
	outcome := "success"
	if err != nil {
		outcome = "error"
	}
	target.Logger().Warn(
		"DBService 慢请求",
		log.String("event", "dbservice_slow_request"),
		log.String("service", target.Name()),
		log.String("resource", resource),
		log.String("mode", mode),
		log.String("operation_kind", "multi"),
		log.String("outcome", outcome),
		log.String("slow_reasons", "total"),
		log.Int64("total_duration_ms", duration.Milliseconds()),
		log.Int("operation_count", operationCount),
		log.Int64("inflight_requests", target.executor.inflight.Load()),
		log.Int64("running_io_requests", target.executor.running.Load()),
	)
}

func mongoModeName(mode rpcapi.MongoExecuteMode) string {
	if mode == rpcapi.MongoExecuteModeTransaction {
		return "transaction"
	}
	return "sequential"
}

func redisModeName(mode rpcapi.RedisExecuteMode) string {
	switch mode {
	case rpcapi.RedisExecuteModeCommand:
		return "command"
	case rpcapi.RedisExecuteModePipeline:
		return "pipeline"
	case rpcapi.RedisExecuteModeTransaction:
		return "transaction"
	case rpcapi.RedisExecuteModeScript:
		return "script"
	default:
		return "unspecified"
	}
}
