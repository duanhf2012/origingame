package area

import (
	"context"
	"time"

	"github.com/duanhf2012/origin/v3/log"
	"github.com/duanhf2012/origin/v3/service"
)

const queryTimeout = 5 * time.Second

// AreaRefreshModule 同步加载首份映射，并按真实系统时间每分钟刷新。
type AreaRefreshModule struct {
	service.Module
	interval time.Duration
	source   *MongoRepository
	catalog  *Catalog
	cancel   context.CancelFunc
	done     chan struct{}
}

// NewAreaRefreshModule 创建持有唯一刷新协程的区服刷新 Module。
func NewAreaRefreshModule(interval time.Duration, source *MongoRepository, catalog *Catalog) *AreaRefreshModule {
	// 生命周期资源在 OnStart 时按需创建。
	return &AreaRefreshModule{interval: interval, source: source, catalog: catalog}
}

// OnStart 在对外监听前取得首份有效映射。
func (module *AreaRefreshModule) OnStart(ctx context.Context) error {
	// 首次加载失败时禁止 Gateway 开始监听。
	if err := module.load(ctx); err != nil {
		return err
	}
	// 成功后创建唯一刷新协程及其取消信号。
	runCtx, cancel := context.WithCancel(context.Background())
	module.cancel = cancel
	module.done = make(chan struct{})
	if err := module.GoSafe(func() { module.run(runCtx) }); err != nil {
		cancel()
		return err
	}
	return nil
}

func (module *AreaRefreshModule) load(parent context.Context) error {
	// 每次查询均受真实系统时间超时约束。
	ctx, cancel := context.WithTimeout(parent, queryTimeout)
	defer cancel()
	// 读取完整映射后再原子替换当前快照。
	mapping, err := module.source.LoadMapping(ctx)
	if err != nil {
		return err
	}
	return module.catalog.Replace(mapping)
}

func (module *AreaRefreshModule) run(ctx context.Context) {
	// 退出前通知 Stop 已完成刷新协程回收。
	defer close(module.done)
	// 用真实系统时间周期刷新基础设施快照。
	ticker := time.NewTicker(module.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// 刷新失败时保留上一份有效快照。
			if err := module.load(ctx); err != nil {
				module.Logger().Error("刷新 Gateway 区服映射失败，继续使用上一份快照", log.Err(err))
			}
		}
	}
}

// OnStop 取消查询和 Ticker，并等待刷新协程退出。
func (module *AreaRefreshModule) OnStop(ctx context.Context) error {
	// 未成功启动时无需等待刷新协程。
	if module.cancel == nil {
		return nil
	}
	// 取消刷新并等待其释放 Timer 等资源。
	module.cancel()
	select {
	case <-module.done:
		return nil
	case <-ctx.Done():
		return context.Cause(ctx)
	}
}
