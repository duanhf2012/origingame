package redismodule

import (
	"context"
	"errors"
	"testing"

	"github.com/duanhf2012/origin/v3/errs"
	rpcapi "origingame/protocol/rpc"
)

func TestValidateRequestAllowsOnlyBoundedRegisteredRedisWork(t *testing.T) {
	registry, err := newScriptRegistry([]ScriptDefinition{{
		ID: "login_rate", Source: "return 1", MinKeys: 1, MaxKeys: 1, MaxArgs: 2, MaxResultNodes: 4,
	}})
	if err != nil {
		t.Fatal(err)
	}

	valid := []rpcapi.RedisRequest{
		{ExecuteMode: rpcapi.RedisExecuteModeCommand, Commands: []rpcapi.RedisCommand{{Name: "get", Args: [][]byte{[]byte("key")}}}},
		{ExecuteMode: rpcapi.RedisExecuteModePipeline, Commands: []rpcapi.RedisCommand{
			{Name: "HSET", Args: bytesArgs("key", "field", "value")},
			{Name: "EXPIRE", Args: bytesArgs("key", "10")},
		}},
		{ExecuteMode: rpcapi.RedisExecuteModeScript, Script: &rpcapi.RedisScriptCall{
			ID: "login_rate", Keys: []string{"key"}, Args: bytesArgs("1", "10"),
		}},
	}
	for index, request := range valid {
		if err := validateRequest(request, registry); err != nil {
			t.Fatalf("valid[%d] error=%v", index, err)
		}
	}

	invalid := []rpcapi.RedisRequest{
		{},
		{ExecuteMode: rpcapi.RedisExecuteModeCommand},
		{ExecuteMode: rpcapi.RedisExecuteModeCommand, Commands: []rpcapi.RedisCommand{{Name: "EVAL", Args: bytesArgs("return 1")}}},
		{ExecuteMode: rpcapi.RedisExecuteModeCommand, Commands: []rpcapi.RedisCommand{{Name: "FLUSHALL"}}},
		{ExecuteMode: rpcapi.RedisExecuteModePipeline, Commands: []rpcapi.RedisCommand{{Name: "GET", Args: bytesArgs("key")}}, Script: &rpcapi.RedisScriptCall{ID: "login_rate"}},
		{ExecuteMode: rpcapi.RedisExecuteModeScript, Script: &rpcapi.RedisScriptCall{ID: "missing", Keys: []string{"key"}}},
		{ExecuteMode: rpcapi.RedisExecuteModeScript, Script: &rpcapi.RedisScriptCall{ID: "login_rate", Keys: []string{"a", "b"}}},
		{ExecuteMode: rpcapi.RedisExecuteModeCommand, Commands: []rpcapi.RedisCommand{{Name: "LRANGE", Args: bytesArgs("key", "0", "-1")}}},
	}
	for index, request := range invalid {
		if err := validateRequest(request, registry); !errors.Is(err, errs.ErrInvalidArgument) {
			t.Fatalf("invalid[%d] error=%v", index, err)
		}
	}
}

func TestBuildRedisValueProducesFlatAcyclicNodeGraph(t *testing.T) {
	value, err := buildRedisValue([]any{
		int64(7),
		[]any{"name", nil},
		map[string]any{"ok": true},
	}, maxRedisResultNodes)
	if err != nil {
		t.Fatal(err)
	}
	if value.RootIndex != 0 || len(value.Nodes) != 8 {
		t.Fatalf("value=%+v", value)
	}
	root := value.Nodes[value.RootIndex]
	if root.Kind != rpcapi.RedisValueKindArray || len(root.Children) != 3 {
		t.Fatalf("root=%+v", root)
	}
	if value.Nodes[root.Children[0]].Kind != rpcapi.RedisValueKindInteger {
		t.Fatalf("first child=%+v", value.Nodes[root.Children[0]])
	}
	mapNode := value.Nodes[root.Children[2]]
	if mapNode.Kind != rpcapi.RedisValueKindMap || len(mapNode.Children) != 2 {
		t.Fatalf("map node=%+v", mapNode)
	}
}

func TestExecuteRequestPreservesPipelineItemFailures(t *testing.T) {
	registry, err := newScriptRegistry(nil)
	if err != nil {
		t.Fatal(err)
	}
	backend := &fakeBackend{pipeline: []backendResult{
		{value: "value"},
		{err: errors.New("WRONGTYPE operation against a key")},
	}}
	request := rpcapi.RedisRequest{
		ExecuteMode: rpcapi.RedisExecuteModePipeline,
		Commands: []rpcapi.RedisCommand{
			{Name: "GET", Args: bytesArgs("a")},
			{Name: "HGET", Args: bytesArgs("a", "f")},
		},
	}
	result := executeRequest(context.Background(), request, registry, backend)
	if len(result.Results) != 2 ||
		result.Results[0].Status != rpcapi.RedisCommandStatusSucceeded ||
		result.Results[1].Status != rpcapi.RedisCommandStatusFailed ||
		result.Results[1].Failure == nil ||
		result.Results[1].Failure.Kind != rpcapi.RedisFailureKindWrongType {
		t.Fatalf("result=%+v", result)
	}
}

type fakeBackend struct {
	command  backendResult
	pipeline []backendResult
	script   backendResult
}

func (backend *fakeBackend) executeCommand(context.Context, rpcapi.RedisCommand) backendResult {
	return backend.command
}

func (backend *fakeBackend) executePipeline(context.Context, []rpcapi.RedisCommand, bool) ([]backendResult, error) {
	return backend.pipeline, nil
}

func (backend *fakeBackend) executeScript(context.Context, registeredScript, rpcapi.RedisScriptCall) backendResult {
	return backend.script
}

func bytesArgs(values ...string) [][]byte {
	result := make([][]byte, len(values))
	for index := range values {
		result[index] = []byte(values[index])
	}
	return result
}
