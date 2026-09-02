// Package registration 管理 GameService 在公共 Redis 中的实例租约。
package registration

import (
	"context"
	"sync"
	"time"

	"github.com/duanhf2012/origin/v3/errs"
	"github.com/duanhf2012/origin/v3/log"
	"github.com/duanhf2012/origin/v3/service"
	"origingame/internal/playerownership"
)

const renewInterval = 5 * time.Second

type registrationStore interface {
	RegisterGameService(context.Context, playerownership.GameServiceRegistration) error
	SetGameServiceDraining(context.Context, int64, playerownership.GameServiceInstance) (bool, error)
}

// GameServiceRegistrationModule 持有实例登记的真实时间续租协程和停止等待。
type GameServiceRegistrationModule struct {
	service.Module                                         // Origin Module 生命周期能力。
	store          registrationStore                       // 续租协程使用的普通调用存储。
	stopStore      registrationStore                       // 停止阶段使用的 Await 调用存储。
	registration   playerownership.GameServiceRegistration // 当前实例登记信息。
	interval       time.Duration                           // 真实系统续租间隔。
	cancel         context.CancelFunc                      // 续租协程取消函数。
	done           chan struct{}                           // 续租协程退出信号。
	stopOnce       sync.Once                               // 保证停止只执行一次。
	stopErr        error                                   // 首次停止失败。
}

// NewGameServiceRegistrationModule 创建固定每5秒续租的 GameService 注册 Module。
// store 供自有续租协程使用普通 Call；stopStore 在 Service OnStop 生命周期内使用 Await。
func NewGameServiceRegistrationModule(
	store registrationStore,
	stopStore registrationStore,
	registration playerownership.GameServiceRegistration,
) *GameServiceRegistrationModule {
	return &GameServiceRegistrationModule{
		store: store, stopStore: stopStore, registration: registration, interval: renewInterval,
	}
}

// OnStart 首次登记必须成功；之后才启动可取消的周期续租。
func (module *GameServiceRegistrationModule) OnStart(ctx context.Context) error {
	if module.store == nil || module.stopStore == nil {
		return errs.NewMessage(errs.CodeInvalidConfig, "GameService 实例登记依赖不完整")
	}
	if err := module.store.RegisterGameService(ctx, module.registration); err != nil {
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

func (module *GameServiceRegistrationModule) run(ctx context.Context) {
	defer close(module.done)
	ticker := time.NewTicker(module.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := module.store.RegisterGameService(ctx, module.registration); err != nil {
				module.Logger().Error("续租 GameService 实例失败", log.Err(err))
			}
		}
	}
}

// OnStop 停止续租，等待协程退出，再条件摘除当前实例。
func (module *GameServiceRegistrationModule) OnStop(ctx context.Context) error {
	module.stopOnce.Do(func() {
		if module.cancel != nil {
			module.cancel()
			select {
			case <-module.done:
			case <-ctx.Done():
				module.stopErr = context.Cause(ctx)
			}
		}
		if module.stopErr != nil {
			return
		}
		_, module.stopErr = module.stopStore.SetGameServiceDraining(
			ctx,
			module.registration.RealAreaID,
			module.registration.GameService,
		)
	})
	return module.stopErr
}
