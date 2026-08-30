package area

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"origingame/internal/dbexecutor"
	"origingame/internal/mongodb"
	rpcapi "origingame/protocol/rpc"
)

const (
	dispatchKey    = "gateway-area-catalog"
	maxAreaRecords = 100000
)

// MongoRepository 只加载 Gateway 所需的显示区服映射。
type MongoRepository struct{ executor dbexecutor.MongoExecutor }

// NewMongoRepository 创建不持有数据库连接的仓储。
func NewMongoRepository(executor dbexecutor.MongoExecutor) *MongoRepository {
	return &MongoRepository{executor: executor}
}

// LoadMapping 完整读取 ShowAreaInfo，并返回独立映射。
func (repository *MongoRepository) LoadMapping(ctx context.Context) (map[int64]int64, error) {
	if repository == nil || repository.executor == nil {
		return nil, errors.New("Gateway 区服仓储未初始化")
	}
	filter, err := bson.Marshal(bson.D{})
	if err != nil {
		return nil, err
	}
	result, err := repository.executor.ExecuteMongo(ctx, dispatchKey, rpcapi.MongoRequest{
		DispatchKey: dispatchKey,
		ExecuteMode: rpcapi.MongoExecuteModeSequential,
		Operations: []rpcapi.MongoOperation{{
			Kind: rpcapi.MongoOperationKindFindMany, Collection: mongodb.ShowAreaInfoName,
			FindMany: &rpcapi.MongoFindMany{Filter: filter, Limit: maxAreaRecords},
		}},
	})
	if err != nil {
		return nil, err
	}
	if result.Failure != nil || len(result.Results) != 1 ||
		result.Results[0].Status != rpcapi.MongoOperationStatusSucceeded {
		return nil, errors.New("Gateway 区服查询返回结构无效")
	}
	mapping := make(map[int64]int64, len(result.Results[0].Documents))
	for _, encoded := range result.Results[0].Documents {
		var document mongodb.ShowAreaInfo
		if err = bson.Unmarshal(encoded, &document); err != nil {
			return nil, fmt.Errorf("解析 ShowAreaInfo: %w", err)
		}
		if document.ID <= 0 || document.RealAreaID <= 0 {
			return nil, errors.New("ShowAreaInfo 包含无效区服 ID")
		}
		if _, exists := mapping[document.ID]; exists {
			return nil, errors.New("ShowAreaInfo 包含重复显示区服")
		}
		mapping[document.ID] = document.RealAreaID
	}
	if len(mapping) == 0 {
		return nil, errors.New("没有可用显示区服")
	}
	return mapping, nil
}
