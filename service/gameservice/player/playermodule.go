package player

import (
	"context"
	"errors"
	"hash/fnv"
	"time"

	"github.com/duanhf2012/origin/v3/errs"
	"github.com/duanhf2012/origin/v3/log"
	"github.com/duanhf2012/origin/v3/service"
	"origingame/internal/playerroute"
	rpcapi "origingame/protocol/rpc"
)

const (
	autoSaveInterval    = 5 * time.Minute
	residentDuration    = 15 * time.Minute
	heartbeatTimeout    = 15 * time.Second
	routeRenewBatchSize = 256
	largePlayerLoadSize = 1 * 1024 * 1024
)

// MongoExecutor 通过指定 PlayerKey 调用本区服 RoleDBService。
type MongoExecutor func(context.Context, string, rpcapi.MongoRequest) (rpcapi.MongoResult, error)

// GatewaySender 把一条已经决定好的业务消息定向发送到 Player 当前 Gateway Node。
type GatewaySender func(string, rpcapi.SendClientMessageRequest) error

// GatewayCloser 主动关闭指定 Gateway Node 上的客户端连接。
type GatewayCloser func(string, string) error

type routeStore interface {
	BeginPlayerLoad(context.Context, playerroute.Player, playerroute.Instance, string) (bool, error)
	CompletePlayerLogin(context.Context, playerroute.Player, playerroute.Instance, string) (bool, error)
	ReleasePlayerLoad(context.Context, playerroute.Player, playerroute.Instance, string) (bool, error)
	MarkPlayerResident(context.Context, playerroute.Player, playerroute.Instance) (bool, error)
	BeginPlayerRelease(context.Context, playerroute.Player, playerroute.Instance) (bool, error)
	RenewPlayerRoutes(context.Context, int64, playerroute.Instance, []playerroute.Player) (int64, error)
}

// Module 是 Player、Proxy、连接索引和存档 Timer 的唯一生命周期所有者。
type Module struct {
	service.Module
	execute               MongoExecutor
	routes                routeStore
	instance              playerroute.Instance
	realAreaID            int64
	playersByKey          map[string]*Player
	playersByConnectionID map[string]*Player
	stopping              bool
	sendGateway           GatewaySender
	closeGateway          GatewayCloser
	scanTimer             *time.Timer
	slowLoadAt            time.Time
	slowLoadSuppressed    uint64
}

// NewModule 创建不持有数据库连接的 PlayerModule。
func NewModule(
	execute MongoExecutor,
	routes routeStore,
	realAreaID int64,
	instance playerroute.Instance,
	sendGateway GatewaySender,
	closeGateway GatewayCloser,
) *Module {
	return &Module{
		execute: execute, routes: routes, realAreaID: realAreaID, instance: instance,
		sendGateway: sendGateway, closeGateway: closeGateway,
	}
}

// OnStart 启动一个实例级真实时间扫描 Timer，不为每个 Player 创建心跳 goroutine。
func (module *Module) OnStart(context.Context) error {
	module.scheduleScan()
	return nil
}

// OnInit 创建固定有界于 player_capacity 的运行期索引。
func (module *Module) OnInit() error {
	if module.execute == nil || module.routes == nil {
		return errs.NewMessage(errs.CodeInvalidConfig, "PlayerModule 数据执行依赖不完整")
	}
	module.playersByKey = make(map[string]*Player)
	module.playersByConnectionID = make(map[string]*Player)
	return nil
}

// LoadNew 完成首次玩家路由认领、RoleDB加载、初始保存、上线和自动存档登记。
func (module *Module) LoadNew(
	ctx context.Context,
	accountID string,
	showAreaID int64,
	gatewayNodeID string,
	connectionID string,
) (*Player, error) {
	startedAt := time.Now()
	outcome, failureStage := "failure", "route"
	loadBytes := 0
	defer func() {
		module.logSlowLoad(startedAt, outcome, failureStage, showAreaID, loadBytes)
	}()
	if module.stopping {
		return nil, errs.ErrServiceStopping
	}
	key := playerroute.PlayerKey(accountID, showAreaID)
	if existing := module.playersByKey[key]; existing != nil {
		return nil, errs.ErrServiceNotReady
	}
	routePlayer := playerroute.Player{AccountID: accountID, ShowAreaID: showAreaID, RealAreaID: module.realAreaID}
	accepted, err := module.routes.BeginPlayerLoad(ctx, routePlayer, module.instance, connectionID)
	if err != nil {
		return nil, err
	}
	if !accepted {
		return nil, errs.ErrServiceNotReady
	}
	current, err := New(accountID, showAreaID, module.realAreaID)
	failureStage = "decode"
	if err != nil {
		module.releaseLoad(ctx, routePlayer, connectionID)
		return nil, err
	}
	current.sendGateway = module.sendGateway
	module.playersByKey[key] = current
	succeeded := false
	defer func() {
		if !succeeded {
			delete(module.playersByKey, key)
			if current.dataInfo.State == StateOnline {
				module.unbindConnection(current)
				current.Offline(time.Now(), 0)
			}
			current.Release()
			module.releaseLoad(ctx, routePlayer, connectionID)
		}
	}()

	loadRequest, err := current.BuildLoadRequest()
	if err != nil {
		return nil, err
	}
	failureStage = "database"
	loadResult, err := module.execute(ctx, key, loadRequest)
	if err != nil {
		return nil, err
	}
	loadBytes = mongoResultBytes(loadResult)
	if loadBytes > largePlayerLoadSize {
		module.Logger().Warn(
			"玩家登录常驻数据过大",
			log.Int64("real_area_id", module.realAreaID),
			log.Int64("show_area_id", showAreaID),
			log.Int("result_bytes", loadBytes),
		)
	}
	failureStage = "decode"
	isNew, err := current.ApplyLoadResult(loadResult)
	if err != nil {
		return nil, err
	}
	failureStage = "on_loaded"
	if err = current.FinishLoad(isNew); err != nil {
		return nil, err
	}
	failureStage = "initial_save"
	if err = module.save(ctx, current); err != nil {
		return nil, err
	}
	if err = module.bindConnection(current, gatewayNodeID, connectionID, time.Now()); err != nil {
		return nil, err
	}
	accepted, err = module.routes.CompletePlayerLogin(ctx, routePlayer, module.instance, connectionID)
	if err != nil || !accepted {
		module.unbindConnection(current)
		return nil, errors.Join(err, errs.ErrServiceNotReady)
	}
	module.startAutoSave(current)
	outcome, failureStage = "success", "none"
	succeeded = true
	return current, nil
}

func (module *Module) logSlowLoad(
	startedAt time.Time,
	outcome string,
	failureStage string,
	showAreaID int64,
	loadBytes int,
) {
	duration := time.Since(startedAt)
	if duration < time.Second {
		return
	}
	now := time.Now()
	if !module.slowLoadAt.IsZero() && now.Sub(module.slowLoadAt) < 10*time.Second {
		module.slowLoadSuppressed++
		return
	}
	module.Logger().Warn(
		"Player 加载处理缓慢",
		log.String("outcome", outcome),
		log.String("failure_stage", failureStage),
		log.Duration("duration", duration),
		log.Int("result_bytes", loadBytes),
		log.Int64("real_area_id", module.realAreaID),
		log.Int64("show_area_id", showAreaID),
		log.Uint64("suppressed", module.slowLoadSuppressed),
	)
	module.slowLoadAt = now
	module.slowLoadSuppressed = 0
}

func mongoResultBytes(result rpcapi.MongoResult) int {
	total := 0
	for _, operation := range result.Results {
		for _, document := range operation.Documents {
			total += len(document)
		}
		total += len(operation.CommandResponse)
	}
	return total
}

// OnStop 取消全部 Player Timer，不等待驻留期，并在停止 Context 内完成最终释放。
func (module *Module) OnStop(ctx context.Context) error {
	module.stopping = true
	if module.scanTimer != nil {
		module.scanTimer.Stop()
		module.scanTimer = nil
	}
	players := make([]*Player, 0, len(module.playersByKey))
	for _, current := range module.playersByKey {
		players = append(players, current)
		module.stopAutoSave(current)
	}
	for _, current := range players {
		if err := ctx.Err(); err != nil {
			return context.Cause(ctx)
		}
		if current.dataInfo.State == StateOnline {
			module.unbindConnection(current)
			current.Offline(time.Now(), 0)
		}
		module.release(ctx, current)
	}
	return nil
}

func (module *Module) scheduleScan() {
	var timer *time.Timer
	timer = time.AfterFunc(5*time.Second, func() {
		if err := module.DispatchAsync(func(ctx context.Context) {
			if module.stopping {
				return
			}
			now := time.Now().UTC()
			for _, current := range module.HeartbeatExpired(now) {
				gatewayNodeID := current.dataInfo.GatewayNodeID
				connectionID := current.dataInfo.GatewayConnectionID
				_ = module.Disconnect(ctx, connectionID)
				if module.closeGateway != nil {
					if err := module.closeGateway(gatewayNodeID, connectionID); err != nil {
						module.Logger().Error("关闭心跳超时 Gateway 连接失败", log.Err(err))
					}
				}
			}
			module.ReleaseExpired(ctx, now)
			module.renewRoutes(ctx)
			if !module.stopping {
				module.scheduleScan()
			}
		}); err != nil {
			// 调度队列瞬时满时复用同一个 Timer 有界重试；Service 已停止等终态不再重启。
			if errors.Is(err, errs.ErrServiceQueueFull) {
				timer.Reset(time.Second)
				return
			}
			module.Logger().Error("提交 Player 扫描任务失败", log.Err(err))
		}
	})
	module.scanTimer = timer
}

func (module *Module) renewRoutes(ctx context.Context) {
	batch := make([]playerroute.Player, 0, routeRenewBatchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		renewed, err := module.routes.RenewPlayerRoutes(ctx, module.realAreaID, module.instance, batch)
		if err != nil || renewed != int64(len(batch)) {
			module.Logger().Error("续租玩家在线路由失败", log.Int("requested", len(batch)), log.Int64("renewed", renewed), log.Err(err))
		}
		batch = batch[:0]
	}
	for _, current := range module.playersByKey {
		if current.dataInfo.State != StateOnline && current.dataInfo.State != StateResident {
			continue
		}
		batch = append(batch, current.routePlayer())
		if len(batch) == routeRenewBatchSize {
			flush()
		}
	}
	flush()
}

// Reconnect 复用当前实例内的 Resident Player；Online 顶号由调用方先处理旧连接通知。
func (module *Module) Reconnect(
	ctx context.Context,
	current *Player,
	gatewayNodeID string,
	connectionID string,
) error {
	if current == nil || module.playersByKey[current.Key()] != current || module.stopping {
		return errs.ErrServiceNotReady
	}
	if current.dataInfo.State == StateOnline && current.dataInfo.GatewayConnectionID == connectionID {
		return nil
	}
	if current.dataInfo.State != StateResident && current.dataInfo.State != StateOnline {
		return errs.ErrServiceNotReady
	}
	if current.dataInfo.State == StateOnline {
		module.unbindConnection(current)
		current.Offline(time.Now(), 0)
	}
	if err := module.bindConnection(current, gatewayNodeID, connectionID, time.Now()); err != nil {
		return err
	}
	accepted, err := module.routes.CompletePlayerLogin(ctx, current.routePlayer(), module.instance, connectionID)
	if err != nil || !accepted {
		module.unbindConnection(current)
		current.Offline(time.Now(), residentDuration)
		return errors.Join(err, errs.ErrServiceNotReady)
	}
	return nil
}

// Disconnect 按当前连接索引幂等地进入 Resident，并立即尝试保存脏数据。
func (module *Module) Disconnect(ctx context.Context, connectionID string) error {
	current := module.playersByConnectionID[connectionID]
	if current == nil || current.dataInfo.GatewayConnectionID != connectionID || current.dataInfo.State != StateOnline {
		return nil
	}
	module.unbindConnection(current)
	current.Offline(time.Now(), residentDuration)
	accepted, err := module.routes.MarkPlayerResident(ctx, current.routePlayer(), module.instance)
	if err != nil || !accepted {
		module.Logger().Error("玩家路由转为 Resident 失败", log.Err(err))
	}
	if saveErr := module.save(ctx, current); saveErr != nil {
		module.Logger().Error("玩家离线存档失败", log.Err(saveErr))
	}
	return err
}

// FindByKey 返回当前实例内的 Player；调用方不得跨 Service goroutine 保存结果。
func (module *Module) FindByKey(key string) *Player { return module.playersByKey[key] }

// FindByConnection 返回当前连接仍绑定的 Player。
func (module *Module) FindByConnection(connectionID string) *Player {
	return module.playersByConnectionID[connectionID]
}

// Heartbeat 更新当前连接最后一次逻辑心跳的真实系统时间。
func (module *Module) Heartbeat(connectionID string, now time.Time) bool {
	current := module.playersByConnectionID[connectionID]
	if current == nil || current.dataInfo.State != StateOnline || current.dataInfo.GatewayConnectionID != connectionID {
		return false
	}
	current.dataInfo.LastHeartbeatAt = now.UTC()
	return true
}

// HeartbeatExpired 报告在线连接是否超过15秒未收到玩家逻辑心跳。
func (module *Module) HeartbeatExpired(now time.Time) []*Player {
	result := make([]*Player, 0)
	for _, current := range module.playersByKey {
		if current.dataInfo.State == StateOnline && now.Sub(current.dataInfo.LastHeartbeatAt) > heartbeatTimeout {
			result = append(result, current)
		}
	}
	return result
}

// ReleaseExpired 释放驻留到期的 Player；最终存档失败只记录错误并继续释放。
func (module *Module) ReleaseExpired(ctx context.Context, now time.Time) {
	for _, current := range module.playersByKey {
		if current.dataInfo.State != StateResident || now.Before(current.dataInfo.ResidentDeadline) {
			continue
		}
		module.release(ctx, current)
	}
}

func (module *Module) release(ctx context.Context, current *Player) {
	current.dataInfo.State = StateReleasing
	module.stopAutoSave(current)
	if err := module.save(ctx, current); err != nil {
		module.Logger().Error("玩家最终存档失败", log.Err(err))
	}
	accepted, err := module.routes.BeginPlayerRelease(ctx, current.routePlayer(), module.instance)
	if err != nil || !accepted {
		module.Logger().Error("玩家路由进入 Leaving 失败", log.Err(err))
	}
	delete(module.playersByKey, current.Key())
	module.unbindConnection(current)
	current.Release()
}

func (module *Module) bindConnection(
	current *Player,
	gatewayNodeID string,
	connectionID string,
	now time.Time,
) error {
	if existing := module.playersByConnectionID[connectionID]; existing != nil && existing != current {
		return errors.New("GatewayConnectionID 已绑定其他 Player")
	}
	if err := current.Online(gatewayNodeID, connectionID, now); err != nil {
		return err
	}
	module.playersByConnectionID[connectionID] = current
	return nil
}

func (module *Module) unbindConnection(current *Player) {
	connectionID := current.dataInfo.GatewayConnectionID
	if module.playersByConnectionID[connectionID] == current {
		delete(module.playersByConnectionID, connectionID)
	}
}

func (module *Module) save(ctx context.Context, current *Player) error {
	plan, ok, err := current.BuildSavePlan()
	if err != nil || !ok {
		return err
	}
	result, err := module.execute(ctx, current.Key(), plan.Request)
	if err != nil {
		return err
	}
	return current.ApplySaveResult(plan, result)
}

func (module *Module) releaseLoad(ctx context.Context, routePlayer playerroute.Player, connectionID string) {
	if _, err := module.routes.ReleasePlayerLoad(ctx, routePlayer, module.instance, connectionID); err != nil {
		module.Logger().Error("回滚玩家加载预占失败", log.Err(err))
	}
}

func (module *Module) startAutoSave(current *Player) {
	module.scheduleAutoSave(current, initialSaveDelay(current.Key()))
}

func (module *Module) scheduleAutoSave(current *Player, delay time.Duration) {
	var timer *time.Timer
	timer = time.AfterFunc(delay, func() {
		if err := module.DispatchAsync(func(ctx context.Context) {
			if module.playersByKey[current.Key()] != current || current.released || module.stopping {
				return
			}
			if err := module.save(ctx, current); err != nil {
				module.Logger().Error("玩家定时存档失败", log.Err(err))
			}
			if module.playersByKey[current.Key()] == current && !current.released && !module.stopping {
				module.scheduleAutoSave(current, autoSaveInterval)
			}
		}); err != nil {
			// 只对瞬时队列满做1秒重试；停止阶段拒绝后不再复活已取消的存档Timer。
			if errors.Is(err, errs.ErrServiceQueueFull) {
				timer.Reset(time.Second)
				return
			}
			module.Logger().Error("提交玩家定时存档任务失败", log.Err(err))
		}
	})
	current.saveTimer = timer
}

func (module *Module) stopAutoSave(current *Player) {
	if current.saveTimer != nil {
		current.saveTimer.Stop()
		current.saveTimer = nil
	}
}

func initialSaveDelay(key string) time.Duration {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(key))
	span := uint64(autoSaveInterval - time.Second)
	return time.Second + time.Duration(hash.Sum64()%span)
}
