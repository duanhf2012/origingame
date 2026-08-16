package player

import (
	"errors"
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/v2/bson"
	rpcapi "origingame/protocol/rpc"
)

type persistentDataEntry struct {
	collection string
	document   any
	loaded     bool
	generation uint64
	saved      uint64
}

// SavePlan 固定一次存档请求与开始时的脏代数快照。
type SavePlan struct {
	Request   rpcapi.MongoRequest
	snapshots []saveSnapshot
}

type saveSnapshot struct {
	entry      *persistentDataEntry
	generation uint64
	insert     bool
}

func (player *Player) registerPersistentData(collection string, document any) (*persistentDataEntry, error) {
	value := reflect.ValueOf(document)
	if collection == "" || !value.IsValid() || value.Kind() != reflect.Pointer || value.IsNil() {
		return nil, errors.New("持久化数据登记无效")
	}
	for _, current := range player.persistent {
		if current.collection == collection || reflect.ValueOf(current.document).Pointer() == value.Pointer() {
			return nil, errors.New("持久化数据重复登记")
		}
	}
	entry := &persistentDataEntry{collection: collection, document: document}
	player.persistent = append(player.persistent, entry)
	return entry, nil
}

// BuildLoadRequest 按登记顺序为全部常驻文档构造一次 RoleDBService 查询。
func (player *Player) BuildLoadRequest() (rpcapi.MongoRequest, error) {
	filter, err := bson.Marshal(bson.D{{Key: "_id", Value: player.key}})
	if err != nil {
		return rpcapi.MongoRequest{}, err
	}
	operations := make([]rpcapi.MongoOperation, len(player.persistent))
	for index, entry := range player.persistent {
		operations[index] = rpcapi.MongoOperation{
			Kind: rpcapi.MongoOperationKindFindOne, Collection: entry.collection,
			FindOne: &rpcapi.MongoFindOne{Filter: filter},
		}
	}
	return rpcapi.MongoRequest{
		DispatchKey: player.key,
		ExecuteMode: rpcapi.MongoExecuteModeSequential,
		Operations:  operations,
	}, nil
}

// ApplyLoadResult 按 Operation 下标自动反序列化全部已登记文档。
func (player *Player) ApplyLoadResult(result rpcapi.MongoResult) (bool, error) {
	if result.Failure != nil || len(result.Results) != len(player.persistent) {
		return false, errors.New("玩家加载 MongoDB 返回结构无效")
	}
	for index, operation := range result.Results {
		if operation.Status != rpcapi.MongoOperationStatusSucceeded || len(operation.Documents) > 1 {
			return false, fmt.Errorf("玩家加载操作 %d 未成功", index)
		}
		entry := player.persistent[index]
		entry.loaded = len(operation.Documents) == 1
		if entry.loaded {
			if err := bson.Unmarshal(operation.Documents[0], entry.document); err != nil {
				return false, fmt.Errorf("解析 %s: %w", entry.collection, err)
			}
		}
	}
	return !player.userInfoEntry.loaded, nil
}

// BuildSavePlan 只序列化当前脏数据；没有脏数据时第二个返回值为 false。
func (player *Player) BuildSavePlan() (SavePlan, bool, error) {
	filter, err := bson.Marshal(bson.D{{Key: "_id", Value: player.key}})
	if err != nil {
		return SavePlan{}, false, err
	}
	plan := SavePlan{Request: rpcapi.MongoRequest{
		DispatchKey: player.key,
		ExecuteMode: rpcapi.MongoExecuteModeSequential,
	}}
	for _, entry := range player.persistent {
		if entry.generation == entry.saved {
			continue
		}
		document, marshalErr := bson.Marshal(entry.document)
		if marshalErr != nil {
			return SavePlan{}, false, fmt.Errorf("编码 %s: %w", entry.collection, marshalErr)
		}
		snapshot := saveSnapshot{entry: entry, generation: entry.generation, insert: !entry.loaded}
		if snapshot.insert {
			plan.Request.Operations = append(plan.Request.Operations, rpcapi.MongoOperation{
				Kind: rpcapi.MongoOperationKindInsertOne, Collection: entry.collection,
				Expectation: &rpcapi.MongoExpectation{Inserted: &rpcapi.MongoCountRange{Min: 1, Max: 1}},
				InsertOne:   &rpcapi.MongoInsertOne{Document: document},
			})
		} else {
			plan.Request.Operations = append(plan.Request.Operations, rpcapi.MongoOperation{
				Kind: rpcapi.MongoOperationKindReplaceOne, Collection: entry.collection,
				Expectation: &rpcapi.MongoExpectation{Matched: &rpcapi.MongoCountRange{Min: 1, Max: 1}},
				ReplaceOne:  &rpcapi.MongoReplaceOne{Filter: filter, Replacement: document},
			})
		}
		plan.snapshots = append(plan.snapshots, snapshot)
	}
	return plan, len(plan.Request.Operations) > 0, nil
}

// ApplySaveResult 仅在对应数据未再次修改时清除脏代数。
func (player *Player) ApplySaveResult(plan SavePlan, result rpcapi.MongoResult) error {
	if result.Failure != nil || len(result.Results) != len(plan.snapshots) {
		return errors.New("玩家存档 MongoDB 返回结构无效")
	}
	for index, snapshot := range plan.snapshots {
		if result.Results[index].Status != rpcapi.MongoOperationStatusSucceeded {
			return fmt.Errorf("玩家存档操作 %d 未成功", index)
		}
		if snapshot.insert {
			snapshot.entry.loaded = true
		}
		if snapshot.entry.generation == snapshot.generation {
			snapshot.entry.saved = snapshot.generation
		}
	}
	return nil
}

// Dirty 报告当前是否至少有一份持久化数据等待保存。
func (player *Player) Dirty() bool {
	for _, entry := range player.persistent {
		if entry.generation != entry.saved {
			return true
		}
	}
	return false
}

func (player *Player) markUserInfoDirty() {
	if player != nil && player.userInfoEntry != nil && !player.released {
		player.userInfoEntry.generation++
	}
}
