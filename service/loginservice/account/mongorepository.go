package account

import (
	"context"
	"errors"
	"time"

	"github.com/duanhf2012/origin/v3/sysmodule/mongodbmodule"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"origingame/internal/mongodb"
)

// MongoRepository 封装 LoginService 对 Account 集合的唯一写入入口。
type MongoRepository struct {
	mongo *mongodbmodule.Module
}

// NewMongoRepository 创建账号 MongoDB 仓储。
func NewMongoRepository(module *mongodbmodule.Module) *MongoRepository {
	return &MongoRepository{mongo: module}
}

// FindOrCreate 原子查询或创建平台账号，并处理并发 upsert 的唯一键竞争。
func (repository *MongoRepository) FindOrCreate(
	ctx context.Context,
	identity LoginIdentity,
	clientIP string,
) (mongodb.Account, error) {
	// 账号更新时间属于登录事实记录，必须使用不受 GM 调时影响的真实时间。
	now := time.Now().UTC()
	createdID := bson.NewObjectID()
	filter := bson.D{{Key: "PlatType", Value: int32(identity.PlatType)}, {Key: "PlatId", Value: identity.PlatID}}
	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "UpdateTime", Value: now}, {Key: "Ip", Value: clientIP}}},
		{Key: "$setOnInsert", Value: bson.D{
			{Key: "_id", Value: createdID}, {Key: "PlatType", Value: int32(identity.PlatType)},
			{Key: "PlatId", Value: identity.PlatID}, {Key: "Gm", Value: false}, {Key: "CreateTime", Value: now},
		}},
	}
	var document mongodb.Account
	err := repository.mongo.Collection(mongodb.AccountName).FindOneAndUpdate(
		ctx, filter, update, options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&document)
	if err == nil {
		return document, nil
	}
	if !mongo.IsDuplicateKeyError(err) {
		return mongodb.Account{}, err
	}

	// 两个首次登录请求可能同时 upsert；唯一索引获胜者已创建账号，失败者重新读取同一身份。
	if findErr := repository.mongo.Collection(mongodb.AccountName).FindOne(ctx, filter).Decode(&document); findErr != nil {
		return mongodb.Account{}, errors.Join(err, findErr)
	}
	return document, nil
}
