package redismodule

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"net"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/duanhf2012/origin/v3/errs"
	"github.com/redis/go-redis/v9"
	rpcapi "origingame/protocol/rpc"
)

const (
	maxDispatchKeyBytes       = 128
	maxRedisCommands          = 1024
	maxRedisArguments         = 10000
	maxRedisResultNodes       = 100000
	maxRedisResultDepth       = 16
	maxRedisResultPayloadSize = 4 << 20
	maxRedisRangeItems        = 100000
)

type backendResult struct {
	value any
	err   error
}

type commandBackend interface {
	executeCommand(context.Context, rpcapi.RedisCommand) backendResult
	executePipeline(context.Context, []rpcapi.RedisCommand, bool) ([]backendResult, error)
	executeScript(context.Context, registeredScript, rpcapi.RedisScriptCall) backendResult
}

func validateRequest(request rpcapi.RedisRequest, scripts scriptRegistry) error {
	if len(request.DispatchKey) > maxDispatchKeyBytes {
		return errs.ErrInvalidArgument
	}
	switch request.ExecuteMode {
	case rpcapi.RedisExecuteModeCommand:
		if len(request.Commands) != 1 || request.Script != nil {
			return errs.ErrInvalidArgument
		}
	case rpcapi.RedisExecuteModePipeline, rpcapi.RedisExecuteModeTransaction:
		if len(request.Commands) == 0 || len(request.Commands) > maxRedisCommands || request.Script != nil {
			return errs.ErrInvalidArgument
		}
	case rpcapi.RedisExecuteModeScript:
		if len(request.Commands) != 0 || request.Script == nil {
			return errs.ErrInvalidArgument
		}
		registered, exists := scripts[request.Script.ID]
		if !exists || len(request.Script.Keys) < registered.minKeys ||
			len(request.Script.Keys) > registered.maxKeys || len(request.Script.Args) > registered.maxArgs {
			return errs.ErrInvalidArgument
		}
		for _, key := range request.Script.Keys {
			if key == "" {
				return errs.ErrInvalidArgument
			}
		}
		return nil
	default:
		return errs.ErrInvalidArgument
	}

	argumentCount := 0
	for _, command := range request.Commands {
		argumentCount += len(command.Args)
		if argumentCount > maxRedisArguments || validateCommand(command) != nil {
			return errs.ErrInvalidArgument
		}
	}
	return nil
}

func validateCommand(command rpcapi.RedisCommand) error {
	name := strings.ToUpper(command.Name)
	if name != strings.ToUpper(strings.TrimSpace(command.Name)) ||
		strings.ContainsAny(name, " \t\r\n") || !allowedCommands[name] || len(command.Args) == 0 {
		return errs.ErrInvalidArgument
	}
	if name == "LRANGE" || name == "GETRANGE" {
		if len(command.Args) != 3 {
			return errs.ErrInvalidArgument
		}
		start, startErr := strconv.ParseInt(string(command.Args[1]), 10, 64)
		stop, stopErr := strconv.ParseInt(string(command.Args[2]), 10, 64)
		if startErr != nil || stopErr != nil || start < 0 || stop < start || stop-start+1 > maxRedisRangeItems {
			return errs.ErrInvalidArgument
		}
	}
	return nil
}

var allowedCommands = map[string]bool{
	"DEL": true, "UNLINK": true, "EXISTS": true, "TYPE": true, "TOUCH": true, "COPY": true,
	"EXPIRE": true, "PEXPIRE": true, "EXPIREAT": true, "PEXPIREAT": true, "EXPIRETIME": true,
	"PEXPIRETIME": true, "PERSIST": true, "TTL": true, "PTTL": true, "RENAME": true, "SCAN": true,
	"GET": true, "SET": true, "GETDEL": true, "GETEX": true, "MGET": true, "MSET": true,
	"MSETNX": true, "INCR": true, "DECR": true, "INCRBY": true, "DECRBY": true,
	"INCRBYFLOAT": true, "APPEND": true, "STRLEN": true, "GETRANGE": true, "SETRANGE": true,
	"HGET": true, "HSET": true, "HSETNX": true, "HMGET": true, "HDEL": true, "HEXISTS": true,
	"HLEN": true, "HSTRLEN": true, "HINCRBY": true, "HINCRBYFLOAT": true, "HRANDFIELD": true,
	"HSCAN": true,
	"LPUSH": true, "RPUSH": true, "LPUSHX": true, "RPUSHX": true, "LPOP": true, "RPOP": true,
	"LINDEX": true, "LSET": true, "LINSERT": true, "LPOS": true, "LRANGE": true, "LLEN": true,
	"LTRIM": true, "LREM": true, "LMOVE": true,
	"SADD": true, "SREM": true, "SISMEMBER": true, "SMISMEMBER": true, "SCARD": true,
	"SPOP": true, "SRANDMEMBER": true, "SMOVE": true, "SDIFF": true, "SINTER": true,
	"SUNION": true, "SDIFFSTORE": true, "SINTERSTORE": true, "SUNIONSTORE": true, "SSCAN": true,
	"ZADD": true, "ZINCRBY": true, "ZREM": true, "ZREMRANGEBYRANK": true,
	"ZREMRANGEBYSCORE": true, "ZREMRANGEBYLEX": true, "ZSCORE": true, "ZMSCORE": true,
	"ZRANK": true, "ZREVRANK": true, "ZRANGE": true, "ZREVRANGE": true, "ZRANGEBYSCORE": true,
	"ZREVRANGEBYSCORE": true, "ZRANGEBYLEX": true, "ZREVRANGEBYLEX": true, "ZCOUNT": true,
	"ZLEXCOUNT": true, "ZCARD": true, "ZPOPMIN": true, "ZPOPMAX": true, "ZSCAN": true,
	"ZDIFF": true, "ZINTER": true, "ZUNION": true, "ZDIFFSTORE": true, "ZINTERSTORE": true,
	"ZUNIONSTORE": true,
	"SETBIT":      true, "GETBIT": true, "BITCOUNT": true, "BITPOS": true, "BITOP": true, "BITFIELD": true,
	"GEOADD": true, "GEODIST": true, "GEOHASH": true, "GEOPOS": true, "GEOSEARCH": true,
	"GEOSEARCHSTORE": true,
	"PFADD":          true, "PFCOUNT": true, "PFMERGE": true,
	"XADD": true, "XDEL": true, "XLEN": true, "XTRIM": true, "XRANGE": true, "XREVRANGE": true,
}

func executeRequest(
	ctx context.Context,
	request rpcapi.RedisRequest,
	scripts scriptRegistry,
	backend commandBackend,
) rpcapi.RedisResult {
	switch request.ExecuteMode {
	case rpcapi.RedisExecuteModeCommand:
		return rpcapi.RedisResult{Results: []rpcapi.RedisCommandResult{
			commandResult(backend.executeCommand(ctx, request.Commands[0]), maxRedisResultNodes),
		}}
	case rpcapi.RedisExecuteModePipeline, rpcapi.RedisExecuteModeTransaction:
		items, overallErr := backend.executePipeline(
			ctx,
			request.Commands,
			request.ExecuteMode == rpcapi.RedisExecuteModeTransaction,
		)
		result := rpcapi.RedisResult{Results: make([]rpcapi.RedisCommandResult, len(request.Commands))}
		if len(items) != len(request.Commands) {
			kind, stateUnknown := classifyRedisFailure(overallErr)
			result.Failure = &rpcapi.RedisExecutionFailure{Kind: kind, StateUnknown: stateUnknown}
			for index := range result.Results {
				status := rpcapi.RedisCommandStatusNotExecuted
				if stateUnknown {
					status = rpcapi.RedisCommandStatusStateUnknown
				}
				result.Results[index].Status = status
			}
			return result
		}
		for index := range items {
			result.Results[index] = commandResult(items[index], maxRedisResultNodes)
		}
		return result
	case rpcapi.RedisExecuteModeScript:
		registered := scripts[request.Script.ID]
		return rpcapi.RedisResult{Results: []rpcapi.RedisCommandResult{
			commandResult(backend.executeScript(ctx, registered, *request.Script), registered.maxResultNodes),
		}}
	default:
		return rpcapi.RedisResult{Failure: &rpcapi.RedisExecutionFailure{Kind: rpcapi.RedisFailureKindCommand}}
	}
}

func commandResult(current backendResult, maxNodes int) rpcapi.RedisCommandResult {
	if errors.Is(current.err, redis.Nil) {
		current.err = nil
		current.value = nil
	}
	if current.err != nil {
		kind, stateUnknown := classifyRedisFailure(current.err)
		status := rpcapi.RedisCommandStatusFailed
		if stateUnknown {
			status = rpcapi.RedisCommandStatusStateUnknown
		}
		return rpcapi.RedisCommandResult{
			Status:  status,
			Failure: &rpcapi.RedisCommandFailure{Kind: kind},
		}
	}
	value, err := buildRedisValue(current.value, maxNodes)
	if err != nil {
		return rpcapi.RedisCommandResult{
			Status:  rpcapi.RedisCommandStatusFailed,
			Failure: &rpcapi.RedisCommandFailure{Kind: rpcapi.RedisFailureKindResultLimitExceeded},
		}
	}
	return rpcapi.RedisCommandResult{Status: rpcapi.RedisCommandStatusSucceeded, Value: value}
}

func classifyRedisFailure(err error) (rpcapi.RedisFailureKind, bool) {
	if err == nil {
		return rpcapi.RedisFailureKindUnknown, false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return rpcapi.RedisFailureKindTimeout, true
	}
	if errors.Is(err, context.Canceled) {
		return rpcapi.RedisFailureKindCanceled, true
	}
	var networkError net.Error
	if errors.As(err, &networkError) {
		return rpcapi.RedisFailureKindNetwork, true
	}
	message := strings.ToUpper(err.Error())
	switch {
	case strings.Contains(message, "WRONGTYPE"):
		return rpcapi.RedisFailureKindWrongType, false
	case strings.Contains(message, "NOSCRIPT"):
		return rpcapi.RedisFailureKindNoScript, false
	case strings.Contains(message, "CROSSSLOT"):
		return rpcapi.RedisFailureKindCrossSlot, false
	case strings.Contains(message, "NOAUTH"), strings.Contains(message, "WRONGPASS"), strings.Contains(message, "NOPERM"):
		return rpcapi.RedisFailureKindUnauthorized, false
	case strings.Contains(message, "EXECABORT"), errors.Is(err, redis.TxFailedErr):
		return rpcapi.RedisFailureKindTransactionAborted, false
	default:
		return rpcapi.RedisFailureKindCommand, false
	}
}

type redisValueBuilder struct {
	nodes    []rpcapi.RedisValueNode
	maxNodes int
	bytes    int
}

func buildRedisValue(value any, maxNodes int) (rpcapi.RedisValue, error) {
	if maxNodes <= 0 || maxNodes > maxRedisResultNodes {
		return rpcapi.RedisValue{}, errs.ErrInvalidArgument
	}
	builder := &redisValueBuilder{maxNodes: maxNodes}
	root, err := builder.add(value, 1)
	if err != nil {
		return rpcapi.RedisValue{}, err
	}
	return rpcapi.RedisValue{RootIndex: root, Nodes: builder.nodes}, nil
}

func (builder *redisValueBuilder) add(value any, depth int) (uint32, error) {
	if depth > maxRedisResultDepth || len(builder.nodes) >= builder.maxNodes {
		return 0, errs.ErrServiceQueueFull
	}
	index := uint32(len(builder.nodes))
	builder.nodes = append(builder.nodes, rpcapi.RedisValueNode{})
	node := &builder.nodes[index]

	switch current := value.(type) {
	case nil:
		node.Kind = rpcapi.RedisValueKindNull
	case []byte:
		if err := builder.addBytes(node, rpcapi.RedisValueKindBytes, current); err != nil {
			return 0, err
		}
	case string:
		if err := builder.addBytes(node, rpcapi.RedisValueKindBytes, []byte(current)); err != nil {
			return 0, err
		}
	case int64:
		node.Kind, node.Integer = rpcapi.RedisValueKindInteger, current
	case int:
		node.Kind, node.Integer = rpcapi.RedisValueKindInteger, int64(current)
	case int32:
		node.Kind, node.Integer = rpcapi.RedisValueKindInteger, int64(current)
	case uint64:
		if current <= math.MaxInt64 {
			node.Kind, node.Integer = rpcapi.RedisValueKindInteger, int64(current)
		} else if err := builder.addBytes(node, rpcapi.RedisValueKindBigNumber, []byte(strconv.FormatUint(current, 10))); err != nil {
			return 0, err
		}
	case float64:
		node.Kind, node.Double = rpcapi.RedisValueKindDouble, current
	case float32:
		node.Kind, node.Double = rpcapi.RedisValueKindDouble, float64(current)
	case bool:
		node.Kind, node.Boolean = rpcapi.RedisValueKindBoolean, current
	case *big.Int:
		if current == nil {
			node.Kind = rpcapi.RedisValueKindNull
		} else if err := builder.addBytes(node, rpcapi.RedisValueKindBigNumber, []byte(current.String())); err != nil {
			return 0, err
		}
	case big.Int:
		if err := builder.addBytes(node, rpcapi.RedisValueKindBigNumber, []byte(current.String())); err != nil {
			return 0, err
		}
	default:
		return builder.addReflect(index, reflect.ValueOf(value), depth)
	}
	return index, nil
}

func (builder *redisValueBuilder) addReflect(index uint32, value reflect.Value, depth int) (uint32, error) {
	if !value.IsValid() {
		builder.nodes[index].Kind = rpcapi.RedisValueKindNull
		return index, nil
	}
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			builder.nodes[index].Kind = rpcapi.RedisValueKindNull
			return index, nil
		}
		value = value.Elem()
	}
	node := &builder.nodes[index]
	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		node.Kind = rpcapi.RedisValueKindArray
		children := make([]uint32, 0, value.Len())
		for childIndex := 0; childIndex < value.Len(); childIndex++ {
			child, err := builder.add(value.Index(childIndex).Interface(), depth+1)
			if err != nil {
				return 0, err
			}
			children = append(children, child)
		}
		builder.nodes[index].Children = children
		return index, nil
	case reflect.Map:
		node.Kind = rpcapi.RedisValueKindMap
		keys := value.MapKeys()
		sort.Slice(keys, func(left, right int) bool {
			return fmt.Sprint(keys[left].Interface()) < fmt.Sprint(keys[right].Interface())
		})
		children := make([]uint32, 0, len(keys)*2)
		for _, key := range keys {
			keyNode, err := builder.add(key.Interface(), depth+1)
			if err != nil {
				return 0, err
			}
			valueNode, err := builder.add(value.MapIndex(key).Interface(), depth+1)
			if err != nil {
				return 0, err
			}
			children = append(children, keyNode, valueNode)
		}
		builder.nodes[index].Children = children
		return index, nil
	default:
		return 0, errs.ErrInvalidArgument
	}
}

func (builder *redisValueBuilder) addBytes(
	node *rpcapi.RedisValueNode,
	kind rpcapi.RedisValueKind,
	value []byte,
) error {
	builder.bytes += len(value) + 16
	if builder.bytes > maxRedisResultPayloadSize {
		return errs.ErrServiceQueueFull
	}
	node.Kind = kind
	node.Bytes = append([]byte(nil), value...)
	return nil
}
