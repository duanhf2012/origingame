package mongodbmodule

import (
	"context"
	"errors"

	"github.com/duanhf2012/origin/v3/errs"
	"go.mongodb.org/mongo-driver/v2/bson"
	rpcapi "origingame/protocol/rpc"
)

const (
	maxDispatchKeyBytes       = 128
	maxMongoOperations        = 128
	maxMongoDocuments         = 100000
	maxMongoPipelineStages    = 128
	maxMongoArrayFilters      = 128
	maxMongoResultPayloadSize = 4 << 20
)

type operationRunner interface {
	runOperation(context.Context, rpcapi.MongoOperation) (rpcapi.MongoOperationResult, *rpcapi.MongoExecutionFailure)
	withTransaction(context.Context, func(context.Context) error) error
}

type operationAbort struct{}

func (operationAbort) Error() string { return "mongodb operation aborted" }

type transactionFailureError struct {
	failure rpcapi.MongoExecutionFailure // 已分类的事务失败信息。
}

func (current *transactionFailureError) Error() string { return "mongodb transaction failed" }

func validateRequest(request rpcapi.MongoRequest) error {
	if len(request.DispatchKey) > maxDispatchKeyBytes ||
		(request.ExecuteMode != rpcapi.MongoExecuteModeSequential &&
			request.ExecuteMode != rpcapi.MongoExecuteModeTransaction) ||
		len(request.Operations) == 0 || len(request.Operations) > maxMongoOperations {
		return errs.ErrInvalidArgument
	}
	for index := range request.Operations {
		if err := validateOperation(request.Operations[index], request.ExecuteMode); err != nil {
			return err
		}
	}
	return nil
}

// ValidateRequest 在 DBService 预留 inflight 名额前完成全部结构与固定边界校验。
func ValidateRequest(request rpcapi.MongoRequest) error {
	return validateRequest(request)
}

func validateOperation(operation rpcapi.MongoOperation, mode rpcapi.MongoExecuteMode) error {
	if operation.Kind == rpcapi.MongoOperationKindUnspecified ||
		operation.Kind > rpcapi.MongoOperationKindRawCommand ||
		countOperationParameters(operation) != 1 {
		return errs.ErrInvalidArgument
	}
	if operation.Kind == rpcapi.MongoOperationKindRawCommand {
		// 首期没有已登记原始命令，保持入口关闭。
		return errs.ErrInvalidArgument
	}
	if operation.Collection == "" {
		return errs.ErrInvalidArgument
	}
	if err := validateExpectation(operation.Expectation); err != nil {
		return err
	}

	switch operation.Kind {
	case rpcapi.MongoOperationKindInsertOne:
		return validateBSON(operation.InsertOne.Document)
	case rpcapi.MongoOperationKindInsertMany:
		if len(operation.InsertMany.Documents) == 0 || len(operation.InsertMany.Documents) > maxMongoDocuments {
			return errs.ErrInvalidArgument
		}
		return validateBSONList(operation.InsertMany.Documents)
	case rpcapi.MongoOperationKindFindOne:
		return validateFindParts(operation.FindOne.Filter, operation.FindOne.Projection, operation.FindOne.Sort)
	case rpcapi.MongoOperationKindFindMany:
		if operation.FindMany.Limit <= 0 || operation.FindMany.Limit > maxMongoDocuments {
			return errs.ErrInvalidArgument
		}
		return validateFindParts(operation.FindMany.Filter, operation.FindMany.Projection, operation.FindMany.Sort)
	case rpcapi.MongoOperationKindCountDocuments:
		return validateBSON(operation.CountDocuments.Filter)
	case rpcapi.MongoOperationKindEstimatedDocumentCount:
		if !bool(*operation.EstimatedDocumentCount) || mode == rpcapi.MongoExecuteModeTransaction {
			return errs.ErrInvalidArgument
		}
		return nil
	case rpcapi.MongoOperationKindAggregate:
		if operation.Aggregate.MaxDocuments <= 0 || operation.Aggregate.MaxDocuments > maxMongoDocuments {
			return errs.ErrInvalidArgument
		}
		return validateReadPipeline(operation.Aggregate.Pipeline)
	case rpcapi.MongoOperationKindUpdateOne:
		return validateUpdateParts(operation.UpdateOne.Filter, operation.UpdateOne.Update, operation.UpdateOne.ArrayFilters)
	case rpcapi.MongoOperationKindUpdateMany:
		return validateUpdateParts(operation.UpdateMany.Filter, operation.UpdateMany.Update, operation.UpdateMany.ArrayFilters)
	case rpcapi.MongoOperationKindReplaceOne:
		return validateReplacement(operation.ReplaceOne.Filter, operation.ReplaceOne.Replacement)
	case rpcapi.MongoOperationKindFindOneAndUpdate:
		if operation.FindOneAndUpdate.ReturnDocument == rpcapi.MongoReturnDocumentUnspecified {
			return errs.ErrInvalidArgument
		}
		if err := validateUpdateParts(
			operation.FindOneAndUpdate.Filter,
			operation.FindOneAndUpdate.Update,
			operation.FindOneAndUpdate.ArrayFilters,
		); err != nil {
			return err
		}
		return validateOptionalBSON(operation.FindOneAndUpdate.Projection, operation.FindOneAndUpdate.Sort)
	case rpcapi.MongoOperationKindFindOneAndReplace:
		if operation.FindOneAndReplace.ReturnDocument == rpcapi.MongoReturnDocumentUnspecified {
			return errs.ErrInvalidArgument
		}
		if err := validateReplacement(
			operation.FindOneAndReplace.Filter,
			operation.FindOneAndReplace.Replacement,
		); err != nil {
			return err
		}
		return validateOptionalBSON(operation.FindOneAndReplace.Projection, operation.FindOneAndReplace.Sort)
	case rpcapi.MongoOperationKindFindOneAndDelete:
		if err := validateBSON(operation.FindOneAndDelete.Filter); err != nil {
			return err
		}
		return validateOptionalBSON(operation.FindOneAndDelete.Projection, operation.FindOneAndDelete.Sort)
	case rpcapi.MongoOperationKindDeleteOne:
		return validateBSON(operation.DeleteOne.Filter)
	case rpcapi.MongoOperationKindDeleteMany:
		return validateBSON(operation.DeleteMany.Filter)
	default:
		return errs.ErrInvalidArgument
	}
}

func countOperationParameters(operation rpcapi.MongoOperation) int {
	count := 0
	values := []bool{
		operation.InsertOne != nil,
		operation.InsertMany != nil,
		operation.FindOne != nil,
		operation.FindMany != nil,
		operation.CountDocuments != nil,
		operation.EstimatedDocumentCount != nil,
		operation.Aggregate != nil,
		operation.UpdateOne != nil,
		operation.UpdateMany != nil,
		operation.ReplaceOne != nil,
		operation.FindOneAndUpdate != nil,
		operation.FindOneAndReplace != nil,
		operation.FindOneAndDelete != nil,
		operation.DeleteOne != nil,
		operation.DeleteMany != nil,
		operation.RawCommand != nil,
	}
	for _, present := range values {
		if present {
			count++
		}
	}
	return count
}

func validateExpectation(expectation *rpcapi.MongoExpectation) error {
	if expectation == nil {
		return nil
	}
	for _, current := range []*rpcapi.MongoCountRange{
		expectation.Documents, expectation.Inserted, expectation.Matched,
		expectation.Modified, expectation.Deleted, expectation.Upserted,
	} {
		if current == nil {
			continue
		}
		if current.Min < 0 || (current.Max != -1 && current.Max < current.Min) {
			return errs.ErrInvalidArgument
		}
	}
	return nil
}

func validateBSON(document []byte) error {
	if len(document) == 0 || bson.Raw(document).Validate() != nil {
		return errs.ErrInvalidArgument
	}
	return nil
}

func validateBSONList(documents [][]byte) error {
	for _, document := range documents {
		if err := validateBSON(document); err != nil {
			return err
		}
	}
	return nil
}

func validateOptionalBSON(documents ...[]byte) error {
	for _, document := range documents {
		if len(document) != 0 {
			if err := validateBSON(document); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateFindParts(filter, projection, sort []byte) error {
	if err := validateBSON(filter); err != nil {
		return err
	}
	return validateOptionalBSON(projection, sort)
}

func validateUpdateParts(filter []byte, update rpcapi.MongoUpdate, arrayFilters [][]byte) error {
	if err := validateBSON(filter); err != nil {
		return err
	}
	if len(arrayFilters) > maxMongoArrayFilters {
		return errs.ErrInvalidArgument
	}
	if err := validateBSONList(arrayFilters); err != nil {
		return err
	}
	switch update.Kind {
	case rpcapi.MongoUpdateKindDocument:
		if len(update.Pipeline) != 0 {
			return errs.ErrInvalidArgument
		}
		return validateBSON(update.Document)
	case rpcapi.MongoUpdateKindPipeline:
		if len(update.Document) != 0 || len(update.Pipeline) == 0 || len(update.Pipeline) > maxMongoPipelineStages {
			return errs.ErrInvalidArgument
		}
		return validateBSONList(update.Pipeline)
	default:
		return errs.ErrInvalidArgument
	}
}

func validateReplacement(filter, replacement []byte) error {
	if err := validateBSON(filter); err != nil {
		return err
	}
	return validateBSON(replacement)
}

func validateReadPipeline(pipeline [][]byte) error {
	if len(pipeline) == 0 || len(pipeline) > maxMongoPipelineStages {
		return errs.ErrInvalidArgument
	}
	for _, stage := range pipeline {
		if err := validateBSON(stage); err != nil {
			return err
		}
		elements, err := bson.Raw(stage).Elements()
		if err != nil || len(elements) == 0 {
			return errs.ErrInvalidArgument
		}
		if key := elements[0].Key(); key == "$out" || key == "$merge" {
			return errs.ErrInvalidArgument
		}
	}
	return nil
}

func executeRequest(ctx context.Context, request rpcapi.MongoRequest, runner operationRunner) rpcapi.MongoResult {
	result := rpcapi.MongoResult{Results: makeNotExecutedResults(len(request.Operations))}
	if request.ExecuteMode == rpcapi.MongoExecuteModeSequential {
		for index, operation := range request.Operations {
			current, failure := runner.runOperation(ctx, operation)
			if failure == nil {
				failure = checkExpectation(index, operation.Expectation, current)
			}
			if failure != nil {
				failure.OperationIndex = int32(index)
				current.Status = statusForMongoFailure(failure)
				result.Results[index] = current
				result.Failure = failure
				return result
			}
			current.Status = rpcapi.MongoOperationStatusSucceeded
			result.Results[index] = current
		}
		return result
	}

	var attemptFailure *rpcapi.MongoExecutionFailure
	err := runner.withTransaction(ctx, func(transactionCtx context.Context) error {
		result.Results = makeNotExecutedResults(len(request.Operations))
		attemptFailure = nil
		for index, operation := range request.Operations {
			current, failure := runner.runOperation(transactionCtx, operation)
			if failure == nil {
				failure = checkExpectation(index, operation.Expectation, current)
			}
			if failure != nil {
				failure.OperationIndex = int32(index)
				current.Status = statusForMongoFailure(failure)
				result.Results[index] = current
				attemptFailure = failure
				return operationAbort{}
			}
			current.Status = rpcapi.MongoOperationStatusSucceeded
			result.Results[index] = current
		}
		return nil
	})
	if err == nil {
		return result
	}
	if attemptFailure == nil {
		var transactionFailure *transactionFailureError
		if errors.As(err, &transactionFailure) {
			attemptFailure = &transactionFailure.failure
		} else {
			attemptFailure = &rpcapi.MongoExecutionFailure{
				OperationIndex: -1,
				Kind:           rpcapi.MongoFailureKindUnknown,
			}
		}
	}
	result.Failure = attemptFailure
	if attemptFailure.StateUnknown {
		for index := range result.Results {
			result.Results[index] = rpcapi.MongoOperationResult{Status: rpcapi.MongoOperationStatusStateUnknown}
		}
		return result
	}
	failedIndex := int(attemptFailure.OperationIndex)
	for index := range result.Results {
		switch {
		case failedIndex >= 0 && index < failedIndex && result.Results[index].Status == rpcapi.MongoOperationStatusSucceeded:
			result.Results[index] = rpcapi.MongoOperationResult{Status: rpcapi.MongoOperationStatusRolledBack}
		case failedIndex >= 0 && index == failedIndex:
			result.Results[index].Status = rpcapi.MongoOperationStatusFailed
		case failedIndex < 0:
			result.Results[index] = rpcapi.MongoOperationResult{Status: rpcapi.MongoOperationStatusRolledBack}
		}
	}
	return result
}

func makeNotExecutedResults(count int) []rpcapi.MongoOperationResult {
	results := make([]rpcapi.MongoOperationResult, count)
	for index := range results {
		results[index].Status = rpcapi.MongoOperationStatusNotExecuted
	}
	return results
}

func checkExpectation(
	operationIndex int,
	expectation *rpcapi.MongoExpectation,
	result rpcapi.MongoOperationResult,
) *rpcapi.MongoExecutionFailure {
	if expectation == nil {
		return nil
	}
	checks := []struct {
		rangeValue *rpcapi.MongoCountRange // 期望数量范围。
		actual     int64                   // 实际操作数量。
	}{
		{expectation.Documents, int64(len(result.Documents))},
		{expectation.Inserted, result.InsertedCount},
		{expectation.Matched, result.MatchedCount},
		{expectation.Modified, result.ModifiedCount},
		{expectation.Deleted, result.DeletedCount},
		{expectation.Upserted, result.UpsertedCount},
	}
	for _, check := range checks {
		if check.rangeValue == nil {
			continue
		}
		if check.actual < check.rangeValue.Min ||
			(check.rangeValue.Max != -1 && check.actual > check.rangeValue.Max) {
			return &rpcapi.MongoExecutionFailure{
				OperationIndex: int32(operationIndex),
				Kind:           rpcapi.MongoFailureKindExpectationFailed,
			}
		}
	}
	return nil
}

func statusForMongoFailure(failure *rpcapi.MongoExecutionFailure) rpcapi.MongoOperationStatus {
	if failure.StateUnknown {
		return rpcapi.MongoOperationStatusStateUnknown
	}
	return rpcapi.MongoOperationStatusFailed
}
