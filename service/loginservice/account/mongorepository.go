package account

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"origingame/internal/dbexecutor"
	"origingame/internal/mongodb"
	rpcapi "origingame/protocol/rpc"
)

// MongoRepository 封装 LoginService 对 Account 集合的唯一写入入口。
type MongoRepository struct {
	executor dbexecutor.MongoExecutor
}

// NewMongoRepository 创建不持有数据库连接的账号仓储。
func NewMongoRepository(executor dbexecutor.MongoExecutor) *MongoRepository {
	return &MongoRepository{executor: executor}
}

// FindOrCreate 原子查询或创建平台账号，并处理并发 upsert 的唯一键竞争。
func (repository *MongoRepository) FindOrCreate(ctx context.Context, identity PlatformIdentity, clientIP string) (mongodb.Account, error) {
	if repository == nil || repository.executor == nil {
		return mongodb.Account{}, errors.New("账号仓储未初始化")
	}
	now := time.Now().UTC()
	filter, err := bson.Marshal(bson.D{
		{Key: "PlatType", Value: int32(identity.PlatType)},
		{Key: "PlatId", Value: identity.PlatID},
	})
	if err != nil {
		return mongodb.Account{}, fmt.Errorf("编码账号条件: %w", err)
	}
	update, err := bson.Marshal(bson.D{
		{Key: "$set", Value: bson.D{{Key: "UpdateTime", Value: now}, {Key: "Ip", Value: clientIP}}},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: bson.NewObjectID()},
			{Key: "PlatType", Value: int32(identity.PlatType)},
			{Key: "PlatId", Value: identity.PlatID},
			{Key: "CreateTime", Value: now},
		}},
	})
	if err != nil {
		return mongodb.Account{}, fmt.Errorf("编码账号更新: %w", err)
	}
	key := identityDispatchKey(identity)
	request := rpcapi.MongoRequest{
		DispatchKey: key,
		ExecuteMode: rpcapi.MongoExecuteModeSequential,
		Operations: []rpcapi.MongoOperation{{
			Kind: rpcapi.MongoOperationKindFindOneAndUpdate, Collection: mongodb.AccountName,
			Expectation: &rpcapi.MongoExpectation{Documents: exactCount(1)},
			FindOneAndUpdate: &rpcapi.MongoFindOneAndUpdate{
				Filter: filter,
				Update: rpcapi.MongoUpdate{Kind: rpcapi.MongoUpdateKindDocument, Document: update},
				Upsert: true, ReturnDocument: rpcapi.MongoReturnDocumentAfter,
			},
		}},
	}
	result, err := repository.executor.ExecuteMongo(ctx, key, request)
	if err != nil {
		return mongodb.Account{}, err
	}
	document, err := accountFromResult(result)
	if err == nil {
		return document, nil
	}
	if result.Failure == nil || result.Failure.Kind != rpcapi.MongoFailureKindDuplicateKey {
		return mongodb.Account{}, err
	}

	// 并发首次登录的唯一索引失败者读取获胜者创建的账号。
	findRequest := rpcapi.MongoRequest{
		DispatchKey: key,
		ExecuteMode: rpcapi.MongoExecuteModeSequential,
		Operations: []rpcapi.MongoOperation{{
			Kind: rpcapi.MongoOperationKindFindOne, Collection: mongodb.AccountName,
			Expectation: &rpcapi.MongoExpectation{Documents: exactCount(1)},
			FindOne:     &rpcapi.MongoFindOne{Filter: filter},
		}},
	}
	result, findErr := repository.executor.ExecuteMongo(ctx, key, findRequest)
	if findErr != nil {
		return mongodb.Account{}, errors.Join(err, findErr)
	}
	document, findErr = accountFromResult(result)
	if findErr != nil {
		return mongodb.Account{}, errors.Join(err, findErr)
	}
	return document, nil
}

func identityDispatchKey(identity PlatformIdentity) string {
	sum := sha256.Sum256([]byte(strconv.FormatInt(int64(identity.PlatType), 10) + "\x00" + identity.PlatID))
	return "account:" + hex.EncodeToString(sum[:])
}

func exactCount(value int64) *rpcapi.MongoCountRange {
	return &rpcapi.MongoCountRange{Min: value, Max: value}
}

func accountFromResult(result rpcapi.MongoResult) (mongodb.Account, error) {
	if result.Failure != nil {
		return mongodb.Account{}, fmt.Errorf("账号 MongoDB 操作失败: kind=%d", result.Failure.Kind)
	}
	if len(result.Results) != 1 || result.Results[0].Status != rpcapi.MongoOperationStatusSucceeded ||
		len(result.Results[0].Documents) != 1 {
		return mongodb.Account{}, errors.New("账号 MongoDB 返回结构无效")
	}
	var document mongodb.Account
	if err := bson.Unmarshal(result.Results[0].Documents[0], &document); err != nil {
		return mongodb.Account{}, fmt.Errorf("解析账号文档: %w", err)
	}
	if document.ID.IsZero() {
		return mongodb.Account{}, errors.New("账号文档缺少 _id")
	}
	return document, nil
}
