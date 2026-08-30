// Package dbexecutor 封装业务组件调用 DBService RPC 时的 Await 与 Call 语义。
package dbexecutor

import (
	"context"

	rpcapi "origingame/protocol/rpc"
)

// MongoExecutor 执行一次按业务 Key 稳定路由的 MongoDB 请求。
type MongoExecutor interface {
	ExecuteMongo(context.Context, string, rpcapi.MongoRequest) (rpcapi.MongoResult, error)
}

// RedisExecutor 执行一次按业务 Key 稳定路由的 Redis 请求。
type RedisExecutor interface {
	ExecuteRedis(context.Context, string, rpcapi.RedisRequest) (rpcapi.RedisResult, error)
}

// AwaitExecutor 在 Service 顺序任务中使用 Await 释放执行权。
type AwaitExecutor struct{ client rpcapi.DBServiceClient }

// NewAwaitExecutor 创建使用 Await 语义的 DBService 执行器。
func NewAwaitExecutor(client rpcapi.DBServiceClient) *AwaitExecutor {
	return &AwaitExecutor{client: client}
}

// ExecuteMongo 按业务 Key 路由并 Await MongoDB 结果。
func (executor *AwaitExecutor) ExecuteMongo(ctx context.Context, key string, request rpcapi.MongoRequest) (rpcapi.MongoResult, error) {
	return executor.client.Route(key).AwaitExecuteMongo(ctx, request)
}

// ExecuteRedis 按业务 Key 路由并 Await Redis 结果。
func (executor *AwaitExecutor) ExecuteRedis(ctx context.Context, key string, request rpcapi.RedisRequest) (rpcapi.RedisResult, error) {
	return executor.client.Route(key).AwaitExecuteRedis(ctx, request)
}

// CallExecutor 在普通 goroutine 或独立 Module 协程中阻塞等待 RPC 结果。
type CallExecutor struct{ client rpcapi.DBServiceClient }

// NewCallExecutor 创建使用 Call 语义的 DBService 执行器。
func NewCallExecutor(client rpcapi.DBServiceClient) *CallExecutor {
	return &CallExecutor{client: client}
}

// ExecuteMongo 按业务 Key 路由并 Call MongoDB 结果。
func (executor *CallExecutor) ExecuteMongo(ctx context.Context, key string, request rpcapi.MongoRequest) (rpcapi.MongoResult, error) {
	return executor.client.Route(key).CallExecuteMongo(ctx, request)
}

// ExecuteRedis 按业务 Key 路由并 Call Redis 结果。
func (executor *CallExecutor) ExecuteRedis(ctx context.Context, key string, request rpcapi.RedisRequest) (rpcapi.RedisResult, error) {
	return executor.client.Route(key).CallExecuteRedis(ctx, request)
}
