// Package database 持有 LoginService 的 MongoDB 生命周期和启动数据准备流程。
package database

import (
	"context"
	"errors"

	"github.com/duanhf2012/origin/v3/sysmodule/mongodbmodule"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"origingame/internal/mongodb"
	"origingame/service/loginservice/account"
	"origingame/service/loginservice/area"
)

// MongoModule 负责 MongoDB Client、账号索引和初始区服快照的完整启动边界。
type MongoModule struct {
	mongodbmodule.Module
	config     mongodbmodule.Config
	catalog    *area.Catalog
	accounts   *account.MongoRepository
	areaSource *area.MongoRepository
}

// NewMongoModule 创建数据库生命周期，并预先建立共享同一 Client 的领域 Repository。
func NewMongoModule(config mongodbmodule.Config, catalog *area.Catalog) *MongoModule {
	module := &MongoModule{config: config, catalog: catalog}
	module.accounts = account.NewMongoRepository(&module.Module)
	module.areaSource = area.NewMongoRepository(&module.Module)
	return module
}

// Accounts 返回账号领域的 MongoDB Repository。
func (module *MongoModule) Accounts() *account.MongoRepository {
	return module.accounts
}

// Areas 返回区服领域的 MongoDB Repository。
func (module *MongoModule) Areas() *area.MongoRepository {
	return module.areaSource
}

// OnInit 冻结官方 MongoDB Module 配置。
func (module *MongoModule) OnInit() error {
	return module.Setup(module.config)
}

// OnStart 在 MongoDB Client Ready 后建立索引并加载初始区服数据。
func (module *MongoModule) OnStart(ctx context.Context) error {
	// 先完成连接和 Primary Ping，之后的索引与快照失败都必须回滚唯一 Client。
	if err := module.Module.OnStart(ctx); err != nil {
		return err
	}
	if _, err := module.EnsureUniqueIndex(ctx, mongodb.AccountName,
		bson.D{{Key: "PlatType", Value: 1}, {Key: "PlatId", Value: 1}},
		options.Index().SetName("uk_platform_identity")); err != nil {
		return errors.Join(err, module.Module.OnStop(ctx))
	}
	areas, err := module.areaSource.LoadSnapshot(ctx)
	if err != nil {
		return errors.Join(err, module.Module.OnStop(ctx))
	}
	if err = module.catalog.Replace(areas); err != nil {
		return errors.Join(err, module.Module.OnStop(ctx))
	}
	return nil
}
