package player

import (
	"testing"

	rpcapi "origingame/protocol/rpc"
)

func TestDirtyGenerationSurvivesModificationDuringSave(t *testing.T) {
	current, err := NewPlayer("0123456789abcdef01234567", 10, 1)
	if err != nil {
		t.Fatal(err)
	}
	loadRequest, err := current.BuildLoadRequest()
	if err != nil || loadRequest.DispatchKey != current.Key() || len(loadRequest.Operations) != 1 {
		t.Fatalf("BuildLoadRequest() = %+v, %v", loadRequest, err)
	}
	isNew, err := current.ApplyLoadResult(rpcapi.MongoResult{Results: []rpcapi.MongoOperationResult{{
		Status: rpcapi.MongoOperationStatusSucceeded,
	}}})
	if err != nil || !isNew {
		t.Fatalf("ApplyLoadResult() = %v, %v", isNew, err)
	}
	if err = current.FinishLoad(isNew); err != nil {
		t.Fatal(err)
	}
	first, ok, err := current.BuildSavePlan()
	if err != nil || !ok || first.Request.Operations[0].Kind != rpcapi.MongoOperationKindInsertOne {
		t.Fatalf("first BuildSavePlan() = %+v, %v, %v", first, ok, err)
	}

	current.UserInfoProxy().SetNickname("new-name")
	if err = current.ApplySaveResult(first, successfulSaveResult(len(first.Request.Operations))); err != nil {
		t.Fatal(err)
	}
	if !current.Dirty() {
		t.Fatal("modification during save was incorrectly cleared")
	}
	second, ok, err := current.BuildSavePlan()
	if err != nil || !ok || second.Request.Operations[0].Kind != rpcapi.MongoOperationKindReplaceOne {
		t.Fatalf("second BuildSavePlan() = %+v, %v, %v", second, ok, err)
	}
	if err = current.ApplySaveResult(second, successfulSaveResult(len(second.Request.Operations))); err != nil {
		t.Fatal(err)
	}
	if current.Dirty() {
		t.Fatal("unchanged saved generation remained dirty")
	}
}

func TestMongoResultBytesCountsDocumentsAndCommandResponse(t *testing.T) {
	result := rpcapi.MongoResult{Results: []rpcapi.MongoOperationResult{
		{Documents: [][]byte{{1, 2}, {3}}, CommandResponse: []byte{4, 5, 6}},
		{Documents: [][]byte{{7, 8, 9, 10}}},
	}}
	if got := mongoResultBytes(result); got != 10 {
		t.Fatalf("mongoResultBytes()=%d, want 10", got)
	}
}

func successfulSaveResult(count int) rpcapi.MongoResult {
	results := make([]rpcapi.MongoOperationResult, count)
	for index := range results {
		results[index].Status = rpcapi.MongoOperationStatusSucceeded
	}
	return rpcapi.MongoResult{Results: results}
}
