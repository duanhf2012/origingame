package mongodbmodule

import (
	"context"
	"errors"
	"sync"

	"github.com/duanhf2012/origin/v3/errs"
	originmongo "github.com/duanhf2012/origin/v3/sysmodule/mongodbmodule"
	"go.mongodb.org/mongo-driver/v2/bson"
	rpcapi "origingame/protocol/rpc"
)

// Module 统一拥有 DBService 的 MongoDB Client 生命周期和通用执行入口。
type Module struct {
	originmongo.Module
	config originmongo.Config

	collectionsMu sync.RWMutex
	collections   map[string]struct{}
}

// NewModule 创建尚未连接的 MongoDB Module；配置在 Origin OnInit 阶段冻结。
func NewModule(config originmongo.Config) *Module {
	return &Module{config: config}
}

// OnInit 在 Module 已绑定到 DBService 后校验并冻结连接配置。
func (module *Module) OnInit() error {
	return module.Module.Setup(module.config)
}

// OnStart 连接 MongoDB 后冻结当前已有集合快照，防止业务 RPC 隐式创建集合。
func (module *Module) OnStart(ctx context.Context) error {
	if err := module.Module.OnStart(ctx); err != nil {
		return err
	}
	names, err := module.Database().ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return errors.Join(err, module.Module.OnStop(ctx))
	}
	collections := make(map[string]struct{}, len(names))
	for _, name := range names {
		collections[name] = struct{}{}
	}
	module.collectionsMu.Lock()
	module.collections = collections
	module.collectionsMu.Unlock()
	return nil
}

// Execute 校验完整请求并在调用方 Await goroutine 中同步执行数据库 I/O。
func (module *Module) Execute(ctx context.Context, request rpcapi.MongoRequest) (rpcapi.MongoResult, error) {
	if ctx == nil {
		return rpcapi.MongoResult{}, errs.ErrInvalidArgument
	}
	if err := validateRequest(request); err != nil {
		return rpcapi.MongoResult{}, err
	}
	if !module.allowsCollections(request) {
		return rpcapi.MongoResult{}, errs.ErrInvalidArgument
	}
	if module.Database() == nil {
		return rpcapi.MongoResult{}, errs.ErrServiceNotReady
	}
	return executeRequest(ctx, request, &driverRunner{module: module}), nil
}

func (module *Module) allowsCollections(request rpcapi.MongoRequest) bool {
	module.collectionsMu.RLock()
	defer module.collectionsMu.RUnlock()
	if module.collections == nil {
		return false
	}
	for _, operation := range request.Operations {
		if _, exists := module.collections[operation.Collection]; !exists {
			return false
		}
	}
	return true
}
