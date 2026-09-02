package area

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"origingame/internal/mongodb"
	rpcapi "origingame/protocol/rpc"
)

type testMongoExecutor struct {
	execute func(context.Context, string, rpcapi.MongoRequest) (rpcapi.MongoResult, error) // 模拟 Mongo 执行。
}

func (executor testMongoExecutor) ExecuteMongo(ctx context.Context, key string, request rpcapi.MongoRequest) (rpcapi.MongoResult, error) {
	return executor.execute(ctx, key, request)
}

func TestLoadSnapshotUsesOneSequentialAccDBServiceRequest(t *testing.T) {
	realBSON, _ := bson.Marshal(mongodb.RealAreaInfo{ID: 1, GateList: []mongodb.GatewayEndpoint{{
		Protocol: "tcp", Address: "127.0.0.1:9001",
	}}})
	showBSON, _ := bson.Marshal(mongodb.ShowAreaInfo{ID: 10, RealAreaID: 1, AreaName: "体验服"})
	var routeKey string
	var captured rpcapi.MongoRequest
	repository := NewMongoRepository(testMongoExecutor{execute: func(_ context.Context, key string, request rpcapi.MongoRequest) (rpcapi.MongoResult, error) {
		routeKey, captured = key, request
		return rpcapi.MongoResult{Results: []rpcapi.MongoOperationResult{
			{Status: rpcapi.MongoOperationStatusSucceeded, Documents: [][]byte{realBSON}},
			{Status: rpcapi.MongoOperationStatusSucceeded, Documents: [][]byte{showBSON}},
		}}, nil
	}})

	snapshot, err := repository.LoadSnapshot(context.Background())
	if err != nil {
		t.Fatalf("LoadSnapshot() error = %v", err)
	}
	if routeKey == "" || routeKey != captured.DispatchKey || len(captured.Operations) != 2 || len(snapshot) != 1 {
		t.Fatalf("unexpected request/snapshot: key=%q request=%+v snapshot=%+v", routeKey, captured, snapshot)
	}
	if captured.Operations[0].Collection != mongodb.RealAreaInfoName ||
		captured.Operations[1].Collection != mongodb.ShowAreaInfoName {
		t.Fatalf("unexpected collections: %+v", captured.Operations)
	}
}
