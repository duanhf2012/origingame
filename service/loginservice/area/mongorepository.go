package area

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"origingame/internal/dbexecutor"
	"origingame/internal/mongodb"
	rpcapi "origingame/protocol/rpc"
)

const (
	areaCatalogDispatchKey = "area-catalog"
	maxAreaDocuments       = 100000
)

// MongoRepository 负责加载并严格关联两张区服基础表。
type MongoRepository struct {
	executor dbexecutor.MongoExecutor
}

// NewMongoRepository 创建不持有数据库连接的区服仓储。
func NewMongoRepository(executor dbexecutor.MongoExecutor) *MongoRepository {
	return &MongoRepository{executor: executor}
}

// LoadSnapshot 一次读取两张基础表，返回完成校验和稳定排序的独立区服快照。
func (repository *MongoRepository) LoadSnapshot(ctx context.Context) ([]Info, error) {
	if repository == nil || repository.executor == nil {
		return nil, errors.New("区服仓储未初始化")
	}
	emptyFilter, err := bson.Marshal(bson.D{})
	if err != nil {
		return nil, fmt.Errorf("编码区服查询条件: %w", err)
	}
	request := rpcapi.MongoRequest{
		DispatchKey: areaCatalogDispatchKey,
		ExecuteMode: rpcapi.MongoExecuteModeSequential,
		Operations: []rpcapi.MongoOperation{
			{
				Kind: rpcapi.MongoOperationKindFindMany, Collection: mongodb.RealAreaInfoName,
				FindMany: &rpcapi.MongoFindMany{Filter: emptyFilter, Limit: maxAreaDocuments},
			},
			{
				Kind: rpcapi.MongoOperationKindFindMany, Collection: mongodb.ShowAreaInfoName,
				FindMany: &rpcapi.MongoFindMany{Filter: emptyFilter, Limit: maxAreaDocuments},
			},
		},
	}
	result, err := repository.executor.ExecuteMongo(ctx, areaCatalogDispatchKey, request)
	if err != nil {
		return nil, err
	}
	if result.Failure != nil || len(result.Results) != 2 {
		return nil, errors.New("区服 MongoDB 返回结构无效")
	}
	realAreas, err := decodeDocuments[mongodb.RealAreaInfo](result.Results[0])
	if err != nil {
		return nil, fmt.Errorf("解析 RealAreaInfo: %w", err)
	}
	showAreas, err := decodeDocuments[mongodb.ShowAreaInfo](result.Results[1])
	if err != nil {
		return nil, fmt.Errorf("解析 ShowAreaInfo: %w", err)
	}
	return buildSnapshot(realAreas, showAreas)
}

func decodeDocuments[T any](result rpcapi.MongoOperationResult) ([]T, error) {
	if result.Status != rpcapi.MongoOperationStatusSucceeded {
		return nil, errors.New("MongoDB 操作未成功")
	}
	documents := make([]T, len(result.Documents))
	for index := range result.Documents {
		if err := bson.Unmarshal(result.Documents[index], &documents[index]); err != nil {
			return nil, err
		}
	}
	return documents, nil
}

func buildSnapshot(realAreas []mongodb.RealAreaInfo, showAreas []mongodb.ShowAreaInfo) ([]Info, error) {
	realByID := make(map[int64][]GateInfo, len(realAreas))
	for _, current := range realAreas {
		if current.ID <= 0 || len(current.GateList) == 0 {
			return nil, fmt.Errorf("RealAreaInfo[%d] 缺少有效 GateList", current.ID)
		}
		gates := make([]GateInfo, 0, len(current.GateList))
		seenProtocols := make(map[string]struct{}, len(current.GateList))
		for _, gate := range current.GateList {
			gate.Protocol = strings.ToLower(strings.TrimSpace(gate.Protocol))
			gate.Address = strings.TrimSpace(gate.Address)
			if gate.Address == "" || (gate.Protocol != "tcp" && gate.Protocol != "kcp" && gate.Protocol != "websocket") {
				return nil, fmt.Errorf("RealAreaInfo[%d] 包含无效 Gateway", current.ID)
			}
			if _, exists := seenProtocols[gate.Protocol]; exists {
				return nil, fmt.Errorf("RealAreaInfo[%d] 的协议 %s 重复", current.ID, gate.Protocol)
			}
			seenProtocols[gate.Protocol] = struct{}{}
			gates = append(gates, GateInfo{Protocol: gate.Protocol, Address: gate.Address})
		}
		realByID[current.ID] = gates
	}

	snapshot := make([]Info, 0, len(showAreas))
	for _, current := range showAreas {
		gates, exists := realByID[current.RealAreaID]
		if !exists || current.ID <= 0 || strings.TrimSpace(current.AreaName) == "" {
			return nil, fmt.Errorf("ShowAreaInfo[%d] 无法关联有效 RealAreaInfo[%d]", current.ID, current.RealAreaID)
		}
		snapshot = append(snapshot, Info{
			ShowAreaID: current.ID, AreaName: strings.TrimSpace(current.AreaName), ServerMark: current.ServerMark,
			ServerStatus: current.ServerStatus, OpenTime: current.OpenTime, GateList: append([]GateInfo(nil), gates...),
		})
	}
	if len(snapshot) == 0 {
		return nil, errors.New("没有可用显示区服")
	}
	sort.Slice(snapshot, func(left, right int) bool { return snapshot[left].ShowAreaID < snapshot[right].ShowAreaID })
	return snapshot, nil
}
