package redismodule

import (
	"strings"

	"github.com/duanhf2012/origin/v3/errs"
	"github.com/redis/go-redis/v9"
)

// ScriptDefinition 是装配层提供给 DBService 的受控 Lua Script 描述。
// 业务 RPC 只能携带 ID、Keys 和 Args，不能提交源码或覆盖限制。
type ScriptDefinition struct {
	ID             string
	Source         string
	KeyCount       int
	MaxArgs        int
	MaxResultNodes int
}

type registeredScript struct {
	id             string
	keyCount       int
	maxArgs        int
	maxResultNodes int
	script         *redis.Script
}

type scriptRegistry map[string]registeredScript

func newScriptRegistry(definitions []ScriptDefinition) (scriptRegistry, error) {
	registry := make(scriptRegistry, len(definitions))
	for _, definition := range definitions {
		definition.ID = strings.TrimSpace(definition.ID)
		if definition.ID == "" || strings.TrimSpace(definition.Source) == "" ||
			definition.KeyCount < 0 || definition.MaxArgs < 0 || definition.MaxResultNodes <= 0 {
			return nil, errs.ErrInvalidArgument
		}
		if _, exists := registry[definition.ID]; exists {
			return nil, errs.ErrInvalidArgument
		}
		registry[definition.ID] = registeredScript{
			id:             definition.ID,
			keyCount:       definition.KeyCount,
			maxArgs:        definition.MaxArgs,
			maxResultNodes: definition.MaxResultNodes,
			script:         redis.NewScript(definition.Source),
		}
	}
	return registry, nil
}
