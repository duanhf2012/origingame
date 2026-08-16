package area

import (
	"context"
	"time"

	"github.com/duanhf2012/origin/v3/log"
	"github.com/duanhf2012/origin/v3/service"
)

const queryTimeout = 5 * time.Second

// RefreshModule 同步加载首份映射，并按真实系统时间每分钟刷新。
type RefreshModule struct {
	service.Module
	interval time.Duration
	source   *MongoRepository
	catalog  *Catalog
	cancel   context.CancelFunc
	done     chan struct{}
}

// NewRefreshModule 创建持有唯一刷新协程的生命周期 Module。
func NewRefreshModule(interval time.Duration, source *MongoRepository, catalog *Catalog) *RefreshModule {
	return &RefreshModule{interval: interval, source: source, catalog: catalog}
}

// OnStart 在对外监听前取得首份有效映射。
func (module *RefreshModule) OnStart(ctx context.Context) error {
	if err := module.load(ctx); err != nil {
		return err
	}
	runCtx, cancel := context.WithCancel(context.Background())
	module.cancel = cancel
	module.done = make(chan struct{})
	if err := module.GoSafe(func() { module.run(runCtx) }); err != nil {
		cancel()
		return err
	}
	return nil
}

func (module *RefreshModule) load(parent context.Context) error {
	ctx, cancel := context.WithTimeout(parent, queryTimeout)
	defer cancel()
	mapping, err := module.source.LoadMapping(ctx)
	if err != nil {
		return err
	}
	return module.catalog.Replace(mapping)
}

func (module *RefreshModule) run(ctx context.Context) {
	defer close(module.done)
	ticker := time.NewTicker(module.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := module.load(ctx); err != nil {
				module.Logger().Error("刷新 Gateway 区服映射失败，继续使用上一份快照", log.Err(err))
			}
		}
	}
}

// OnStop 取消查询和 Ticker，并等待刷新协程退出。
func (module *RefreshModule) OnStop(ctx context.Context) error {
	if module.cancel == nil {
		return nil
	}
	module.cancel()
	select {
	case <-module.done:
		return nil
	case <-ctx.Done():
		return context.Cause(ctx)
	}
}
