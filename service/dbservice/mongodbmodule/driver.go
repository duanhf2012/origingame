package mongodbmodule

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	rpcapi "origingame/protocol/rpc"
)

type driverRunner struct {
	module *Module
}

type resultLimitError struct{}

func (resultLimitError) Error() string { return "mongodb result exceeds fixed limit" }

func (runner *driverRunner) withTransaction(ctx context.Context, callback func(context.Context) error) error {
	err := runner.module.WithTransaction(ctx, callback)
	if err == nil {
		return nil
	}
	var aborted operationAbort
	if errors.As(err, &aborted) {
		return err
	}
	return &transactionFailureError{failure: *classifyFailure(err)}
}

func (runner *driverRunner) runOperation(
	ctx context.Context,
	operation rpcapi.MongoOperation,
) (rpcapi.MongoOperationResult, *rpcapi.MongoExecutionFailure) {
	var result rpcapi.MongoOperationResult
	var err error
	switch operation.Kind {
	case rpcapi.MongoOperationKindInsertOne:
		result, err = runner.insertOne(ctx, operation)
	case rpcapi.MongoOperationKindInsertMany:
		result, err = runner.insertMany(ctx, operation)
	case rpcapi.MongoOperationKindFindOne:
		result, err = runner.findOne(ctx, operation)
	case rpcapi.MongoOperationKindFindMany:
		result, err = runner.findMany(ctx, operation)
	case rpcapi.MongoOperationKindCountDocuments:
		result.Count, err = runner.collection(operation).CountDocuments(ctx, bson.Raw(operation.CountDocuments.Filter))
	case rpcapi.MongoOperationKindEstimatedDocumentCount:
		result.Count, err = runner.collection(operation).EstimatedDocumentCount(ctx)
	case rpcapi.MongoOperationKindAggregate:
		result, err = runner.aggregate(ctx, operation)
	case rpcapi.MongoOperationKindUpdateOne:
		result, err = runner.updateOne(ctx, operation)
	case rpcapi.MongoOperationKindUpdateMany:
		result, err = runner.updateMany(ctx, operation)
	case rpcapi.MongoOperationKindReplaceOne:
		result, err = runner.replaceOne(ctx, operation)
	case rpcapi.MongoOperationKindFindOneAndUpdate:
		result, err = runner.findOneAndUpdate(ctx, operation)
	case rpcapi.MongoOperationKindFindOneAndReplace:
		result, err = runner.findOneAndReplace(ctx, operation)
	case rpcapi.MongoOperationKindFindOneAndDelete:
		result, err = runner.findOneAndDelete(ctx, operation)
	case rpcapi.MongoOperationKindDeleteOne:
		var deleted *mongo.DeleteResult
		deleted, err = runner.collection(operation).DeleteOne(ctx, bson.Raw(operation.DeleteOne.Filter))
		if deleted != nil {
			result.DeletedCount = deleted.DeletedCount
		}
	case rpcapi.MongoOperationKindDeleteMany:
		var deleted *mongo.DeleteResult
		deleted, err = runner.collection(operation).DeleteMany(ctx, bson.Raw(operation.DeleteMany.Filter))
		if deleted != nil {
			result.DeletedCount = deleted.DeletedCount
		}
	default:
		err = resultLimitError{}
	}
	if err != nil {
		return result, classifyFailure(err)
	}
	return result, nil
}

func (runner *driverRunner) collection(operation rpcapi.MongoOperation) *mongo.Collection {
	return runner.module.Collection(operation.Collection)
}

func (runner *driverRunner) insertOne(
	ctx context.Context,
	operation rpcapi.MongoOperation,
) (rpcapi.MongoOperationResult, error) {
	inserted, err := runner.collection(operation).InsertOne(ctx, bson.Raw(operation.InsertOne.Document))
	var result rpcapi.MongoOperationResult
	if inserted != nil {
		value, marshalErr := marshalBSONValue(inserted.InsertedID)
		if marshalErr != nil {
			return result, marshalErr
		}
		result.InsertedCount = 1
		result.InsertedIDs = []rpcapi.MongoBSONValue{value}
	}
	return result, err
}

func (runner *driverRunner) insertMany(
	ctx context.Context,
	operation rpcapi.MongoOperation,
) (rpcapi.MongoOperationResult, error) {
	documents := make([]any, len(operation.InsertMany.Documents))
	for index := range operation.InsertMany.Documents {
		documents[index] = bson.Raw(operation.InsertMany.Documents[index])
	}
	inserted, err := runner.collection(operation).InsertMany(
		ctx,
		documents,
		options.InsertMany().SetOrdered(!operation.InsertMany.Unordered),
	)
	var result rpcapi.MongoOperationResult
	if inserted != nil {
		ids, marshalErr := marshalBSONValues(inserted.InsertedIDs)
		if marshalErr != nil {
			return result, marshalErr
		}
		result.InsertedIDs = ids
		result.InsertedCount = int64(len(ids))
	}
	if err != nil {
		var writeException mongo.BulkWriteException
		if errors.As(err, &writeException) {
			result.WriteFailures = make([]rpcapi.MongoWriteFailure, 0, len(writeException.WriteErrors))
			for _, writeError := range writeException.WriteErrors {
				result.WriteFailures = append(result.WriteFailures, rpcapi.MongoWriteFailure{
					DocumentIndex: int32(writeError.Index),
					Kind:          failureKindFromServerCode(writeError.Code),
					ServerCode:    int32(writeError.Code),
				})
			}
		}
	}
	return result, err
}

func (runner *driverRunner) findOne(
	ctx context.Context,
	operation rpcapi.MongoOperation,
) (rpcapi.MongoOperationResult, error) {
	current := operation.FindOne
	builder := options.FindOne()
	applyProjectionAndSort(builder.SetProjection, builder.SetSort, current.Projection, current.Sort)
	raw, err := runner.collection(operation).FindOne(ctx, bson.Raw(current.Filter), builder).Raw()
	return singleDocumentResult(raw, err)
}

func (runner *driverRunner) findMany(
	ctx context.Context,
	operation rpcapi.MongoOperation,
) (rpcapi.MongoOperationResult, error) {
	current := operation.FindMany
	limit := current.Limit
	readLimit := limit
	if limit == maxMongoDocuments {
		readLimit++
	}
	builder := options.Find().SetLimit(readLimit)
	if len(current.Projection) != 0 {
		builder.SetProjection(bson.Raw(current.Projection))
	}
	if len(current.Sort) != 0 {
		builder.SetSort(bson.Raw(current.Sort))
	}
	cursor, err := runner.collection(operation).Find(ctx, bson.Raw(current.Filter), builder)
	if err != nil {
		return rpcapi.MongoOperationResult{}, err
	}
	documents, err := readCursor(ctx, cursor, limit, limit == maxMongoDocuments)
	return rpcapi.MongoOperationResult{Documents: documents}, err
}

func (runner *driverRunner) aggregate(
	ctx context.Context,
	operation rpcapi.MongoOperation,
) (rpcapi.MongoOperationResult, error) {
	current := operation.Aggregate
	pipeline, err := decodePipeline(current.Pipeline)
	if err != nil {
		return rpcapi.MongoOperationResult{}, err
	}
	cursor, err := runner.collection(operation).Aggregate(
		ctx,
		pipeline,
		options.Aggregate().SetAllowDiskUse(current.AllowDiskUse),
	)
	if err != nil {
		return rpcapi.MongoOperationResult{}, err
	}
	documents, err := readCursor(ctx, cursor, current.MaxDocuments, true)
	return rpcapi.MongoOperationResult{Documents: documents}, err
}

func (runner *driverRunner) updateOne(
	ctx context.Context,
	operation rpcapi.MongoOperation,
) (rpcapi.MongoOperationResult, error) {
	current := operation.UpdateOne
	update, err := decodeUpdate(current.Update)
	if err != nil {
		return rpcapi.MongoOperationResult{}, err
	}
	builder := options.UpdateOne().SetUpsert(current.Upsert)
	if len(current.ArrayFilters) != 0 {
		builder.SetArrayFilters(rawDocuments(current.ArrayFilters))
	}
	updated, err := runner.collection(operation).UpdateOne(ctx, bson.Raw(current.Filter), update, builder)
	return updateResult(updated, err)
}

func (runner *driverRunner) updateMany(
	ctx context.Context,
	operation rpcapi.MongoOperation,
) (rpcapi.MongoOperationResult, error) {
	current := operation.UpdateMany
	update, err := decodeUpdate(current.Update)
	if err != nil {
		return rpcapi.MongoOperationResult{}, err
	}
	builder := options.UpdateMany()
	if len(current.ArrayFilters) != 0 {
		builder.SetArrayFilters(rawDocuments(current.ArrayFilters))
	}
	updated, err := runner.collection(operation).UpdateMany(ctx, bson.Raw(current.Filter), update, builder)
	return updateResult(updated, err)
}

func (runner *driverRunner) replaceOne(
	ctx context.Context,
	operation rpcapi.MongoOperation,
) (rpcapi.MongoOperationResult, error) {
	current := operation.ReplaceOne
	updated, err := runner.collection(operation).ReplaceOne(
		ctx,
		bson.Raw(current.Filter),
		bson.Raw(current.Replacement),
		options.Replace().SetUpsert(current.Upsert),
	)
	return updateResult(updated, err)
}

func (runner *driverRunner) findOneAndUpdate(
	ctx context.Context,
	operation rpcapi.MongoOperation,
) (rpcapi.MongoOperationResult, error) {
	current := operation.FindOneAndUpdate
	update, err := decodeUpdate(current.Update)
	if err != nil {
		return rpcapi.MongoOperationResult{}, err
	}
	builder := options.FindOneAndUpdate().
		SetUpsert(current.Upsert).
		SetReturnDocument(returnDocument(current.ReturnDocument))
	if len(current.Projection) != 0 {
		builder.SetProjection(bson.Raw(current.Projection))
	}
	if len(current.Sort) != 0 {
		builder.SetSort(bson.Raw(current.Sort))
	}
	if len(current.ArrayFilters) != 0 {
		builder.SetArrayFilters(rawDocuments(current.ArrayFilters))
	}
	raw, err := runner.collection(operation).FindOneAndUpdate(ctx, bson.Raw(current.Filter), update, builder).Raw()
	return singleDocumentResult(raw, err)
}

func (runner *driverRunner) findOneAndReplace(
	ctx context.Context,
	operation rpcapi.MongoOperation,
) (rpcapi.MongoOperationResult, error) {
	current := operation.FindOneAndReplace
	builder := options.FindOneAndReplace().
		SetUpsert(current.Upsert).
		SetReturnDocument(returnDocument(current.ReturnDocument))
	if len(current.Projection) != 0 {
		builder.SetProjection(bson.Raw(current.Projection))
	}
	if len(current.Sort) != 0 {
		builder.SetSort(bson.Raw(current.Sort))
	}
	raw, err := runner.collection(operation).FindOneAndReplace(
		ctx,
		bson.Raw(current.Filter),
		bson.Raw(current.Replacement),
		builder,
	).Raw()
	return singleDocumentResult(raw, err)
}

func (runner *driverRunner) findOneAndDelete(
	ctx context.Context,
	operation rpcapi.MongoOperation,
) (rpcapi.MongoOperationResult, error) {
	current := operation.FindOneAndDelete
	builder := options.FindOneAndDelete()
	if len(current.Projection) != 0 {
		builder.SetProjection(bson.Raw(current.Projection))
	}
	if len(current.Sort) != 0 {
		builder.SetSort(bson.Raw(current.Sort))
	}
	raw, err := runner.collection(operation).FindOneAndDelete(ctx, bson.Raw(current.Filter), builder).Raw()
	return singleDocumentResult(raw, err)
}

func singleDocumentResult(raw bson.Raw, err error) (rpcapi.MongoOperationResult, error) {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return rpcapi.MongoOperationResult{}, nil
	}
	if err != nil {
		return rpcapi.MongoOperationResult{}, err
	}
	if len(raw) > maxMongoResultPayloadSize {
		return rpcapi.MongoOperationResult{}, resultLimitError{}
	}
	return rpcapi.MongoOperationResult{Documents: [][]byte{append([]byte(nil), raw...)}}, nil
}

func readCursor(
	ctx context.Context,
	cursor *mongo.Cursor,
	limit int64,
	failWhenMore bool,
) (documents [][]byte, resultErr error) {
	defer func() {
		resultErr = errors.Join(resultErr, cursor.Close(ctx))
	}()
	var bytes int
	for cursor.Next(ctx) {
		if int64(len(documents)) >= limit {
			if failWhenMore {
				return nil, resultLimitError{}
			}
			break
		}
		current := append([]byte(nil), cursor.Current...)
		bytes += len(current) + 16
		if bytes > maxMongoResultPayloadSize {
			return nil, resultLimitError{}
		}
		documents = append(documents, current)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return documents, nil
}

func updateResult(updated *mongo.UpdateResult, err error) (rpcapi.MongoOperationResult, error) {
	var result rpcapi.MongoOperationResult
	if updated != nil {
		result.MatchedCount = updated.MatchedCount
		result.ModifiedCount = updated.ModifiedCount
		result.UpsertedCount = updated.UpsertedCount
		if updated.UpsertedID != nil {
			value, marshalErr := marshalBSONValue(updated.UpsertedID)
			if marshalErr != nil {
				return result, marshalErr
			}
			result.UpsertedIDs = []rpcapi.MongoBSONValue{value}
		}
	}
	return result, err
}

func marshalBSONValue(value any) (rpcapi.MongoBSONValue, error) {
	typeValue, encoded, err := bson.MarshalValue(value)
	if err != nil {
		return rpcapi.MongoBSONValue{}, err
	}
	return rpcapi.MongoBSONValue{Type: byte(typeValue), Value: append([]byte(nil), encoded...)}, nil
}

func marshalBSONValues(values []any) ([]rpcapi.MongoBSONValue, error) {
	result := make([]rpcapi.MongoBSONValue, 0, len(values))
	for _, value := range values {
		encoded, err := marshalBSONValue(value)
		if err != nil {
			return nil, err
		}
		result = append(result, encoded)
	}
	return result, nil
}

func decodePipeline(stages [][]byte) (mongo.Pipeline, error) {
	pipeline := make(mongo.Pipeline, 0, len(stages))
	for _, stage := range stages {
		var document bson.D
		if err := bson.Unmarshal(stage, &document); err != nil {
			return nil, err
		}
		pipeline = append(pipeline, document)
	}
	return pipeline, nil
}

func decodeUpdate(update rpcapi.MongoUpdate) (any, error) {
	if update.Kind == rpcapi.MongoUpdateKindDocument {
		return bson.Raw(update.Document), nil
	}
	return decodePipeline(update.Pipeline)
}

func rawDocuments(documents [][]byte) []any {
	result := make([]any, len(documents))
	for index := range documents {
		result[index] = bson.Raw(documents[index])
	}
	return result
}

func returnDocument(value rpcapi.MongoReturnDocument) options.ReturnDocument {
	if value == rpcapi.MongoReturnDocumentAfter {
		return options.After
	}
	return options.Before
}

func applyProjectionAndSort(
	setProjection func(any) *options.FindOneOptionsBuilder,
	setSort func(any) *options.FindOneOptionsBuilder,
	projection []byte,
	sort []byte,
) {
	if len(projection) != 0 {
		setProjection(bson.Raw(projection))
	}
	if len(sort) != 0 {
		setSort(bson.Raw(sort))
	}
}

func classifyFailure(err error) *rpcapi.MongoExecutionFailure {
	failure := &rpcapi.MongoExecutionFailure{Kind: rpcapi.MongoFailureKindUnknown}
	if err == nil {
		return failure
	}
	var limit resultLimitError
	if errors.As(err, &limit) {
		failure.Kind = rpcapi.MongoFailureKindResultLimitExceeded
		return failure
	}
	if errors.Is(err, context.DeadlineExceeded) || mongo.IsTimeout(err) {
		failure.Kind = rpcapi.MongoFailureKindTimeout
		failure.StateUnknown = true
		return failure
	}
	if errors.Is(err, context.Canceled) {
		failure.Kind = rpcapi.MongoFailureKindCanceled
		failure.StateUnknown = true
		return failure
	}
	if mongo.IsNetworkError(err) {
		failure.Kind = rpcapi.MongoFailureKindNetwork
		failure.StateUnknown = true
		return failure
	}
	if mongo.IsDuplicateKeyError(err) {
		failure.Kind = rpcapi.MongoFailureKindDuplicateKey
	}
	var serverError mongo.ServerError
	if errors.As(err, &serverError) {
		codes := serverError.ErrorCodes()
		if len(codes) != 0 {
			failure.ServerCode = int32(codes[0])
			if failure.Kind == rpcapi.MongoFailureKindUnknown {
				failure.Kind = failureKindFromServerCode(codes[0])
			}
		}
		if serverError.HasErrorLabel("UnknownTransactionCommitResult") {
			failure.StateUnknown = true
		}
	}
	return failure
}

func failureKindFromServerCode(code int) rpcapi.MongoFailureKind {
	switch code {
	case 11000, 11001, 12582:
		return rpcapi.MongoFailureKindDuplicateKey
	case 112:
		return rpcapi.MongoFailureKindWriteConflict
	case 121:
		return rpcapi.MongoFailureKindDocumentValidation
	case 13, 18:
		return rpcapi.MongoFailureKindUnauthorized
	case 50:
		return rpcapi.MongoFailureKindTimeout
	default:
		return rpcapi.MongoFailureKindCommand
	}
}
