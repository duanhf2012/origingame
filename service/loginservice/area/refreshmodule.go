package area

import (
	"context"
	"time"

	"github.com/duanhf2012/origin/v3/log"
	"github.com/duanhf2012/origin/v3/service"
)

// RefreshModule 持有区服目录的真实时间周期刷新任务。
type RefreshModule struct {
	service.Module
	interval time.Duration
	source   *MongoRepository
	catalog  *Catalog
	cancel   context.CancelFunc
	done     chan struct{}
}

// NewRefreshModule 创建区服周期刷新 Module。
func NewRefreshModule(interval time.Duration, source *MongoRepository, catalog *Catalog) *RefreshModule {
	return &RefreshModule{interval: interval, source: source, catalog: catalog}
}

// OnStart 同步加载首份有效快照，再启动明确可取消和等待的真实时间刷新协程。
func (module *RefreshModule) OnStart(ctx context.Context) error {
	snapshot, err := module.source.LoadSnapshot(ctx)
	if err != nil {
		return err
	}
	if err = module.catalog.Replace(snapshot); err != nil {
		return err
	}
	// 区服刷新属于基础设施同步，不能被 Node 的 GM 游戏时间重排。
	runCtx, cancel := context.WithCancel(context.Background())
	module.cancel = cancel
	module.done = make(chan struct{})
	if err = module.GoSafe(func() { module.run(runCtx) }); err != nil {
		cancel()
		return err
	}
	return nil
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
			module.refresh(ctx)
		}
	}
}

func (module *RefreshModule) refresh(ctx context.Context) {
	snapshot, err := module.source.LoadSnapshot(ctx)
	if err != nil {
		module.Logger().Error("刷新区服列表失败，继续使用上一份有效快照", log.Err(err))
		return
	}
	if err = module.catalog.Replace(snapshot); err != nil {
		module.Logger().Error("发布区服列表失败，继续使用上一份有效快照", log.Err(err))
	}
}

// OnStop 取消数据库查询和 Ticker，并等待唯一刷新协程退出。
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
