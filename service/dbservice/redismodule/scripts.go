package redismodule

import (
	"strings"

	"github.com/duanhf2012/origin/v3/errs"
	"github.com/redis/go-redis/v9"
)

// ScriptDefinition 是装配层提供给 DBService 的受控 Lua Script 描述。
// 业务 RPC 只能携带 ID、Keys 和 Args，不能提交源码或覆盖限制。
type ScriptDefinition struct {
	ID             string // 稳定 Script 标识。
	Source         string // 受控 Lua 源码。
	MinKeys        int    // 允许的最少键参数数。
	MaxKeys        int    // 允许的最多键参数数。
	MaxArgs        int    // 允许的最多普通参数数。
	MaxResultNodes int    // 允许的最多结果节点数。
}

type registeredScript struct {
	id             string        // 稳定 Script 标识。
	minKeys        int           // 最少键参数数。
	maxKeys        int           // 最多键参数数。
	maxArgs        int           // 最多普通参数数。
	maxResultNodes int           // 最多结果节点数。
	script         *redis.Script // 已编译的 Redis Script。
}

type scriptRegistry map[string]registeredScript

func newScriptRegistry(definitions []ScriptDefinition) (scriptRegistry, error) {
	registry := make(scriptRegistry, len(definitions))
	for _, definition := range definitions {
		definition.ID = strings.TrimSpace(definition.ID)
		if definition.ID == "" || strings.TrimSpace(definition.Source) == "" ||
			definition.MinKeys < 0 || definition.MaxKeys < definition.MinKeys ||
			definition.MaxArgs < 0 || definition.MaxResultNodes <= 0 {
			return nil, errs.ErrInvalidArgument
		}
		if _, exists := registry[definition.ID]; exists {
			return nil, errs.ErrInvalidArgument
		}
		registry[definition.ID] = registeredScript{
			id:             definition.ID,
			minKeys:        definition.MinKeys,
			maxKeys:        definition.MaxKeys,
			maxArgs:        definition.MaxArgs,
			maxResultNodes: definition.MaxResultNodes,
			script:         redis.NewScript(definition.Source),
		}
	}
	return registry, nil
}
