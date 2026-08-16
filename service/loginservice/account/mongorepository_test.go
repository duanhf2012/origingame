package account

import (
	"context"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"origingame/internal/mongodb"
	rpcapi "origingame/protocol/rpc"
)

func TestFindOrCreateUsesAccDBServiceFindOneAndUpdate(t *testing.T) {
	document := mongodb.Account{
		ID: bson.NewObjectID(), PlatType: int32(LoginTypeGuest), PlatID: "guest-1",
		CreateTime: time.Now().UTC(), UpdateTime: time.Now().UTC(),
	}
	documentBSON, err := bson.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	var routeKey string
	var captured rpcapi.MongoRequest
	repository := NewMongoRepository(func(_ context.Context, key string, request rpcapi.MongoRequest) (rpcapi.MongoResult, error) {
		routeKey, captured = key, request
		return rpcapi.MongoResult{Results: []rpcapi.MongoOperationResult{{
			Status: rpcapi.MongoOperationStatusSucceeded, Documents: [][]byte{documentBSON},
		}}}, nil
	})

	result, err := repository.FindOrCreate(context.Background(), PlatformIdentity{
		PlatType: LoginTypeGuest, PlatID: "guest-1",
	}, "192.0.2.20")
	if err != nil {
		t.Fatalf("FindOrCreate() error = %v", err)
	}
	if result.ID != document.ID || routeKey == "" || routeKey != captured.DispatchKey || strings.Contains(routeKey, "guest-1") {
		t.Fatalf("result/key mismatch: result=%+v key=%q", result, routeKey)
	}
	if len(captured.Operations) != 1 || captured.Operations[0].Kind != rpcapi.MongoOperationKindFindOneAndUpdate ||
		captured.Operations[0].Collection != mongodb.AccountName || captured.Operations[0].FindOneAndUpdate == nil ||
		!captured.Operations[0].FindOneAndUpdate.Upsert ||
		captured.Operations[0].FindOneAndUpdate.ReturnDocument != rpcapi.MongoReturnDocumentAfter {
		t.Fatalf("unexpected Mongo request: %+v", captured)
	}
}
