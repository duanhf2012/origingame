package mongodbmodule

import (
	"context"
	"errors"
	"testing"

	"github.com/duanhf2012/origin/v3/errs"
	"go.mongodb.org/mongo-driver/v2/bson"
	rpcapi "origingame/protocol/rpc"
)

func TestValidateRequestRejectsInvalidMongoOperations(t *testing.T) {
	validFilter := mustBSON(t, bson.D{{Key: "_id", Value: "1"}})
	validDocument := mustBSON(t, bson.D{{Key: "_id", Value: "1"}})
	validUpdate := mustBSON(t, bson.D{{Key: "$set", Value: bson.D{{Key: "name", Value: "n"}}}})
	marker := rpcapi.MongoEstimatedDocumentCount(true)

	tests := []struct {
		name    string
		request rpcapi.MongoRequest
	}{
		{name: "unspecified mode", request: rpcapi.MongoRequest{Operations: []rpcapi.MongoOperation{{}}}},
		{name: "empty operations", request: rpcapi.MongoRequest{ExecuteMode: rpcapi.MongoExecuteModeSequential}},
		{name: "unspecified operation", request: mongoRequest(rpcapi.MongoOperation{})},
		{name: "missing union parameter", request: mongoRequest(rpcapi.MongoOperation{Kind: rpcapi.MongoOperationKindInsertOne, Collection: "users"})},
		{name: "multiple union parameters", request: mongoRequest(rpcapi.MongoOperation{
			Kind:       rpcapi.MongoOperationKindInsertOne,
			Collection: "users",
			InsertOne:  &rpcapi.MongoInsertOne{Document: validDocument},
			DeleteOne:  &rpcapi.MongoDeleteOne{Filter: validFilter},
		})},
		{name: "missing collection", request: mongoRequest(rpcapi.MongoOperation{
			Kind:      rpcapi.MongoOperationKindInsertOne,
			InsertOne: &rpcapi.MongoInsertOne{Document: validDocument},
		})},
		{name: "invalid bson", request: mongoRequest(rpcapi.MongoOperation{
			Kind:       rpcapi.MongoOperationKindInsertOne,
			Collection: "users",
			InsertOne:  &rpcapi.MongoInsertOne{Document: []byte{1}},
		})},
		{name: "find many without limit", request: mongoRequest(rpcapi.MongoOperation{
			Kind:       rpcapi.MongoOperationKindFindMany,
			Collection: "users",
			FindMany:   &rpcapi.MongoFindMany{Filter: validFilter},
		})},
		{name: "write without filter", request: mongoRequest(rpcapi.MongoOperation{
			Kind:       rpcapi.MongoOperationKindUpdateOne,
			Collection: "users",
			UpdateOne: &rpcapi.MongoUpdateOne{Update: rpcapi.MongoUpdate{
				Kind: rpcapi.MongoUpdateKindDocument, Document: validUpdate,
			}},
		})},
		{name: "update has document and pipeline", request: mongoRequest(rpcapi.MongoOperation{
			Kind:       rpcapi.MongoOperationKindUpdateOne,
			Collection: "users",
			UpdateOne: &rpcapi.MongoUpdateOne{Filter: validFilter, Update: rpcapi.MongoUpdate{
				Kind: rpcapi.MongoUpdateKindDocument, Document: validUpdate, Pipeline: [][]byte{validUpdate},
			}},
		})},
		{name: "invalid expectation", request: mongoRequest(rpcapi.MongoOperation{
			Kind:        rpcapi.MongoOperationKindInsertOne,
			Collection:  "users",
			InsertOne:   &rpcapi.MongoInsertOne{Document: validDocument},
			Expectation: &rpcapi.MongoExpectation{Inserted: &rpcapi.MongoCountRange{Min: 2, Max: 1}},
		})},
		{name: "estimated count in transaction", request: rpcapi.MongoRequest{
			ExecuteMode: rpcapi.MongoExecuteModeTransaction,
			Operations: []rpcapi.MongoOperation{{
				Kind: rpcapi.MongoOperationKindEstimatedDocumentCount, Collection: "users", EstimatedDocumentCount: &marker,
			}},
		}},
		{name: "raw command is not registered", request: mongoRequest(rpcapi.MongoOperation{
			Kind:       rpcapi.MongoOperationKindRawCommand,
			RawCommand: &rpcapi.MongoRawCommand{ID: "custom", Command: validDocument},
		})},
		{name: "aggregate write stage", request: mongoRequest(rpcapi.MongoOperation{
			Kind: rpcapi.MongoOperationKindAggregate, Collection: "users",
			Aggregate: &rpcapi.MongoAggregate{Pipeline: [][]byte{mustBSON(t, bson.D{{Key: "$out", Value: "other"}})}, MaxDocuments: 1},
		})},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateRequest(test.request); !errors.Is(err, errs.ErrInvalidArgument) {
				t.Fatalf("validateRequest() error=%v", err)
			}
		})
	}
}

func TestValidateRequestAcceptsRepresentativeOperations(t *testing.T) {
	filter := mustBSON(t, bson.D{{Key: "_id", Value: "1"}})
	document := mustBSON(t, bson.D{{Key: "_id", Value: "1"}})
	update := mustBSON(t, bson.D{{Key: "$set", Value: bson.D{{Key: "name", Value: "n"}}}})
	request := rpcapi.MongoRequest{
		DispatchKey: "player-1",
		ExecuteMode: rpcapi.MongoExecuteModeSequential,
		Operations: []rpcapi.MongoOperation{
			{Kind: rpcapi.MongoOperationKindInsertOne, Collection: "users", InsertOne: &rpcapi.MongoInsertOne{Document: document}},
			{Kind: rpcapi.MongoOperationKindFindOne, Collection: "users", FindOne: &rpcapi.MongoFindOne{Filter: filter}},
			{Kind: rpcapi.MongoOperationKindUpdateOne, Collection: "users", UpdateOne: &rpcapi.MongoUpdateOne{
				Filter: filter, Update: rpcapi.MongoUpdate{Kind: rpcapi.MongoUpdateKindDocument, Document: update},
			}},
			{Kind: rpcapi.MongoOperationKindDeleteOne, Collection: "users", DeleteOne: &rpcapi.MongoDeleteOne{Filter: filter}},
		},
	}
	if err := validateRequest(request); err != nil {
		t.Fatalf("validateRequest() error=%v", err)
	}
}

func TestExecuteSequentialStopsAfterFailureAndKeepsPartialResults(t *testing.T) {
	runner := &fakeRunner{
		results: []runResult{
			{result: rpcapi.MongoOperationResult{InsertedCount: 1}},
			{result: rpcapi.MongoOperationResult{MatchedCount: 0}},
		},
	}
	request := rpcapi.MongoRequest{
		ExecuteMode: rpcapi.MongoExecuteModeSequential,
		Operations: []rpcapi.MongoOperation{
			validInsertOperation(t),
			withExpectation(validUpdateOperation(t), &rpcapi.MongoExpectation{
				Matched: &rpcapi.MongoCountRange{Min: 1, Max: 1},
			}),
			validInsertOperation(t),
		},
	}
	result := executeRequest(context.Background(), request, runner)
	statuses := []rpcapi.MongoOperationStatus{
		rpcapi.MongoOperationStatusSucceeded,
		rpcapi.MongoOperationStatusFailed,
		rpcapi.MongoOperationStatusNotExecuted,
	}
	for index, want := range statuses {
		if result.Results[index].Status != want {
			t.Fatalf("result[%d].Status=%v, want %v", index, result.Results[index].Status, want)
		}
	}
	if result.Failure == nil || result.Failure.OperationIndex != 1 ||
		result.Failure.Kind != rpcapi.MongoFailureKindExpectationFailed {
		t.Fatalf("Failure=%+v", result.Failure)
	}
	if runner.calls != 2 {
		t.Fatalf("runner.calls=%d, want 2", runner.calls)
	}
}

func TestExecuteTransactionMarksEarlierResultsRolledBack(t *testing.T) {
	runner := &fakeRunner{results: []runResult{
		{result: rpcapi.MongoOperationResult{InsertedCount: 1}},
		{failure: &rpcapi.MongoExecutionFailure{Kind: rpcapi.MongoFailureKindDuplicateKey, ServerCode: 11000}},
	}}
	request := rpcapi.MongoRequest{
		ExecuteMode: rpcapi.MongoExecuteModeTransaction,
		Operations: []rpcapi.MongoOperation{
			validInsertOperation(t),
			validInsertOperation(t),
			validInsertOperation(t),
		},
	}
	result := executeRequest(context.Background(), request, runner)
	if !runner.transactionCalled {
		t.Fatal("事务模式没有调用 WithTransaction")
	}
	statuses := []rpcapi.MongoOperationStatus{
		rpcapi.MongoOperationStatusRolledBack,
		rpcapi.MongoOperationStatusFailed,
		rpcapi.MongoOperationStatusNotExecuted,
	}
	for index, want := range statuses {
		if result.Results[index].Status != want {
			t.Fatalf("result[%d].Status=%v, want %v", index, result.Results[index].Status, want)
		}
	}
	if result.Results[0].InsertedCount != 0 {
		t.Fatalf("已回滚结果仍保留业务字段: %+v", result.Results[0])
	}
}

type runResult struct {
	result  rpcapi.MongoOperationResult
	failure *rpcapi.MongoExecutionFailure
}

type fakeRunner struct {
	results           []runResult
	calls             int
	transactionCalled bool
	transactionErr    error
}

func (runner *fakeRunner) runOperation(context.Context, rpcapi.MongoOperation) (rpcapi.MongoOperationResult, *rpcapi.MongoExecutionFailure) {
	result := runner.results[runner.calls]
	runner.calls++
	return result.result, result.failure
}

func (runner *fakeRunner) withTransaction(ctx context.Context, callback func(context.Context) error) error {
	runner.transactionCalled = true
	if err := callback(ctx); err != nil {
		return err
	}
	return runner.transactionErr
}

func mongoRequest(operation rpcapi.MongoOperation) rpcapi.MongoRequest {
	return rpcapi.MongoRequest{ExecuteMode: rpcapi.MongoExecuteModeSequential, Operations: []rpcapi.MongoOperation{operation}}
}

func validInsertOperation(t *testing.T) rpcapi.MongoOperation {
	t.Helper()
	return rpcapi.MongoOperation{
		Kind: rpcapi.MongoOperationKindInsertOne, Collection: "users",
		InsertOne: &rpcapi.MongoInsertOne{Document: mustBSON(t, bson.D{{Key: "_id", Value: "1"}})},
	}
}

func validUpdateOperation(t *testing.T) rpcapi.MongoOperation {
	t.Helper()
	return rpcapi.MongoOperation{
		Kind: rpcapi.MongoOperationKindUpdateOne, Collection: "users",
		UpdateOne: &rpcapi.MongoUpdateOne{
			Filter: mustBSON(t, bson.D{{Key: "_id", Value: "1"}}),
			Update: rpcapi.MongoUpdate{
				Kind:     rpcapi.MongoUpdateKindDocument,
				Document: mustBSON(t, bson.D{{Key: "$set", Value: bson.D{{Key: "name", Value: "n"}}}}),
			},
		},
	}
}

func withExpectation(operation rpcapi.MongoOperation, expectation *rpcapi.MongoExpectation) rpcapi.MongoOperation {
	operation.Expectation = expectation
	return operation
}

func mustBSON(t *testing.T, value any) []byte {
	t.Helper()
	encoded, err := bson.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
