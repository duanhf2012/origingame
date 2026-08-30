package redismodule

import (
	"context"
	"strings"

	"github.com/duanhf2012/origin/v3/errs"
	originredis "github.com/duanhf2012/origin/v3/sysmodule/redismodule"
	"github.com/redis/go-redis/v9"
	rpcapi "origingame/protocol/rpc"
)

// Module 统一拥有 DBService 的 Redis Client 生命周期、脚本登记和通用执行入口。
type Module struct {
	originredis.Module
	config  originredis.Config
	scripts scriptRegistry
}

// NewModule 校验受控 Script 登记并创建尚未连接的 Redis Module。
func NewModule(config originredis.Config, definitions []ScriptDefinition) (*Module, error) {
	scripts, err := newScriptRegistry(definitions)
	if err != nil {
		return nil, err
	}
	return &Module{config: config, scripts: scripts}, nil
}

// OnInit 在 Module 已绑定到 DBService 后校验并冻结连接配置。
func (module *Module) OnInit() error {
	return module.Module.Setup(module.config)
}

// Execute 校验完整请求并在调用方 Await goroutine 中同步执行 Redis I/O。
func (module *Module) Execute(ctx context.Context, request rpcapi.RedisRequest) (rpcapi.RedisResult, error) {
	if ctx == nil {
		return rpcapi.RedisResult{}, errs.ErrInvalidArgument
	}
	if err := validateRequest(request, module.scripts); err != nil {
		return rpcapi.RedisResult{}, err
	}
	if module.Client() == nil {
		return rpcapi.RedisResult{}, errs.ErrServiceNotReady
	}
	return executeRequest(ctx, request, module.scripts, &redisBackend{module: module}), nil
}

// ValidateRequest 在 DBService 预留 inflight 名额前完成命令和 Script 边界校验。
func (module *Module) ValidateRequest(request rpcapi.RedisRequest) error {
	if module == nil {
		return errs.ErrInvalidArgument
	}
	return validateRequest(request, module.scripts)
}

type redisBackend struct {
	module *Module
}

func (backend *redisBackend) executeCommand(ctx context.Context, command rpcapi.RedisCommand) backendResult {
	value, err := backend.module.Do(ctx, redisCommandArguments(command)...)
	return backendResult{value: value, err: err}
}

func (backend *redisBackend) executePipeline(
	ctx context.Context,
	commands []rpcapi.RedisCommand,
	transaction bool,
) ([]backendResult, error) {
	collected := make([]*redis.Cmd, 0, len(commands))
	callback := func(ctx context.Context, pipe redis.Pipeliner) error {
		for _, command := range commands {
			collected = append(collected, pipe.Do(ctx, redisCommandArguments(command)...))
		}
		return nil
	}
	var overallErr error
	if transaction {
		_, overallErr = backend.module.TxPipelined(ctx, callback)
	} else {
		_, overallErr = backend.module.Pipelined(ctx, callback)
	}
	results := make([]backendResult, len(collected))
	for index, command := range collected {
		results[index].value, results[index].err = command.Result()
	}
	return results, overallErr
}

func (backend *redisBackend) executeScript(
	ctx context.Context,
	registered registeredScript,
	call rpcapi.RedisScriptCall,
) backendResult {
	arguments := make([]any, len(call.Args))
	for index := range call.Args {
		arguments[index] = call.Args[index]
	}
	value, err := backend.module.RunScript(ctx, registered.script, call.Keys, arguments...)
	return backendResult{value: value, err: err}
}

func redisCommandArguments(command rpcapi.RedisCommand) []any {
	arguments := make([]any, 1, len(command.Args)+1)
	arguments[0] = strings.ToUpper(command.Name)
	for _, argument := range command.Args {
		arguments = append(arguments, argument)
	}
	return arguments
}
