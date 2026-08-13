package area

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/duanhf2012/origin/v3/sysmodule/mongodbmodule"
	"go.mongodb.org/mongo-driver/v2/bson"
	"origingame/internal/mongodb"
)

// MongoRepository 负责加载并严格关联两张区服基础表。
type MongoRepository struct {
	mongo *mongodbmodule.Module
}

// NewMongoRepository 创建区服 MongoDB 仓储。
func NewMongoRepository(module *mongodbmodule.Module) *MongoRepository {
	return &MongoRepository{mongo: module}
}

// LoadSnapshot 返回完成校验和稳定排序的独立区服快照。
func (repository *MongoRepository) LoadSnapshot(ctx context.Context) ([]Info, error) {
	// 先完整读取真实区服，临时 Map 不会在失败时污染在线快照。
	realCursor, err := repository.mongo.Collection(mongodb.RealAreaInfoName).Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("查询 RealAreaInfo: %w", err)
	}
	defer realCursor.Close(ctx)
	var realAreas []mongodb.RealAreaInfo
	if err = realCursor.All(ctx, &realAreas); err != nil {
		return nil, fmt.Errorf("解析 RealAreaInfo: %w", err)
	}
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

	// 再读取显示区服；任何悬空关联都使整个新快照失败，而不是悄悄跳过错误数据。
	showCursor, err := repository.mongo.Collection(mongodb.ShowAreaInfoName).Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("查询 ShowAreaInfo: %w", err)
	}
	defer showCursor.Close(ctx)
	var showAreas []mongodb.ShowAreaInfo
	if err = showCursor.All(ctx, &showAreas); err != nil {
		return nil, fmt.Errorf("解析 ShowAreaInfo: %w", err)
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
