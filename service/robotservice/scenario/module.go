package scenario

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/duanhf2012/origin/v3/errs"
	originlog "github.com/duanhf2012/origin/v3/log"
	"github.com/duanhf2012/origin/v3/service"
	"github.com/duanhf2012/origin/v3/sysmodule/blueprintmodule"
	commonpb "origingame/protocol/common"
	rpcapi "origingame/protocol/rpc"
	"origingame/service/robotservice/virtualplayer"
)

const pushInboxCapacity = 8

type messageWaiter struct {
	messageID     commonpb.MessageID
	startedAtMS   int64
	handle        *blueprintmodule.YieldHandle
	cancelTimeout func()
}

type robotRuntime struct {
	id          int64
	attempts    int
	firstFailed bool
	attemptErr  error
	player      *virtualplayer.Player
	instance    *blueprintmodule.Instance
	execution   *blueprintmodule.Execution
	pushWaiter  *messageWaiter
	inbox       []virtualplayer.InboundMessage
}

type runRuntime struct {
	record       *runRecord
	workload     *workload
	robots       map[int64]*robotRuntime
	scheduleDone bool
	scheduleErr  error
	failureCode  string
	failureText  string
}

// Module 同时拥有Blueprint引擎、运行控制、I/O Worker、Timer和全部VirtualPlayer。
type Module struct {
	blueprintmodule.Module
	config Config

	login      *virtualplayer.LoginClient
	executor   *ioExecutor
	timers     *realTimerScheduler
	controller *runController

	rootCtx    context.Context
	rootCancel context.CancelFunc
	current    *runRuntime
	stopping   bool
}

// NewModule 创建尚未绑定RobotService的场景Module。
func NewModule(config Config) *Module { return &Module{config: config} }

// OnInit 冻结配置并登记首批节点；此阶段不读取蓝图文件或启动goroutine。
func (module *Module) OnInit() error {
	if err := ValidateConfig(module.config); err != nil {
		return err
	}
	if err := module.Setup(blueprintmodule.Config{
		NodeDir: module.config.Blueprint.NodeDir, GraphDir: module.config.Blueprint.GraphDir,
	}); err != nil {
		return err
	}
	login, err := virtualplayer.NewLoginClient(virtualplayer.LoginClientConfig{
		URL: module.config.Target.LoginURL, ShowAreaID: module.config.Target.ShowAreaID,
		PlatformType:     module.config.Identity.PlatformType,
		PlatformIDPrefix: module.config.Identity.PlatformIDPrefix,
		Workers:          module.config.IO.Workers,
	})
	if err != nil {
		return err
	}
	module.login = login
	module.executor, err = newIOExecutor(
		module.config.IO.Workers,
		module.config.IO.QueueMessages,
		func(task func(context.Context)) error { return module.DispatchAsync(task) },
	)
	if err != nil {
		module.login.Close()
		return err
	}
	module.timers, err = newRealTimerScheduler(
		func(task func(context.Context)) error { return module.DispatchAsync(task) },
	)
	if err != nil {
		module.login.Close()
		return err
	}
	module.controller = newRunController()
	return module.RegisterNodes(
		func() blueprintmodule.IExecNode { return &robotStartNode{} },
		func() blueprintmodule.IExecNode { return &httpLoginNode{module: module} },
		func() blueprintmodule.IExecNode { return &connectGatewayNode{module: module} },
		func() blueprintmodule.IExecNode { return &loginPlayerNode{module: module} },
		func() blueprintmodule.IExecNode { return &startHeartbeatNode{module: module} },
		func() blueprintmodule.IExecNode { return &heartbeatNode{module: module} },
		func() blueprintmodule.IExecNode { return &waitNode{module: module} },
		func() blueprintmodule.IExecNode { return &waitMessageNode{module: module} },
		func() blueprintmodule.IExecNode { return &disconnectNode{module: module} },
	)
}

// OnStart 先启动I/O资源和Blueprint引擎，再按配置预登记默认运行。
// Origin会在全部OnStart成功后才激活Service调度器，因此自动运行必须经过零延迟Module Timer
// 跨过激活屏障，不能在OnStart中直接投递普通Service任务。
func (module *Module) OnStart(ctx context.Context) error {
	module.rootCtx, module.rootCancel = context.WithCancel(context.Background())
	if err := module.executor.start(module.rootCtx); err != nil {
		module.rootCancel()
		return err
	}
	if err := module.Module.OnStart(ctx); err != nil {
		module.executor.stop()
		module.rootCancel()
		return err
	}
	if !module.config.Control.StartupRun {
		return nil
	}
	node := module.GetNode()
	requestID := "startup"
	if node != nil {
		requestID = "startup:" + node.ID() + ":" + strconv.FormatUint(node.SessionID(), 10)
	}
	request := rpcapi.StartRobotRunRequest{
		RequestID: requestID, ScenarioName: module.config.Blueprint.GraphName,
	}
	if module.AfterFunc(0, func(context.Context, service.TimerID) {
		if _, startErr := module.StartRun(request); startErr != nil {
			module.Logger().Error("自动启动机器人运行失败", originlog.Err(startErr))
		}
	}) == service.InvalidTimerID {
		_ = module.Module.OnStop(ctx)
		module.executor.stop()
		module.rootCancel()
		return errors.New("预登记机器人自动运行Timer失败")
	}
	return nil
}

// OnStop 先撤销运行和外部I/O，再关闭Blueprint引擎；重复停止安全。
func (module *Module) OnStop(ctx context.Context) error {
	if module.stopping {
		return nil
	}
	module.stopping = true
	if module.rootCancel != nil {
		module.rootCancel()
	}
	module.forceCleanupCurrent()
	if module.timers != nil {
		module.timers.stop()
	}
	if module.executor != nil {
		module.executor.stop()
	}
	if module.login != nil {
		module.login.Close()
	}
	return module.Module.OnStop(ctx)
}

// ListScenarios 返回当前首期唯一默认场景；Blueprint OnStart已保证它可编译。
func (module *Module) ListScenarios(request rpcapi.ListRobotScenariosRequest) (rpcapi.ListRobotScenariosResponse, error) {
	if request.Limit < 0 || request.Limit > 100 {
		return rpcapi.ListRobotScenariosResponse{}, errs.ErrInvalidArgument
	}
	return rpcapi.ListRobotScenariosResponse{Scenarios: []rpcapi.RobotScenarioSummary{{
		Name: module.config.Blueprint.GraphName, EntranceID: module.config.Blueprint.EntranceID, IsDefault: true,
	}}}, nil
}

// StartRun 幂等启动服务端固定Workload配置。
func (module *Module) StartRun(request rpcapi.StartRobotRunRequest) (rpcapi.RobotRunSnapshot, error) {
	if module == nil || module.rootCtx == nil || module.stopping {
		return rpcapi.RobotRunSnapshot{}, errs.ErrServiceNotReady
	}
	if strings.TrimSpace(request.ScenarioName) != module.config.Blueprint.GraphName {
		return rpcapi.RobotRunSnapshot{}, errs.NewMessage(errs.CodeInvalidArgument, "机器人场景不存在")
	}
	snapshot, record, existing, err := module.controller.prepareStart(
		module.rootCtx, request.RequestID, request.ScenarioName, module.config.Workload.Users,
	)
	if err != nil {
		return rpcapi.RobotRunSnapshot{}, controlError(err)
	}
	if existing {
		module.Logger().Debug(
			"机器人运行启动请求命中现有运行",
			originlog.String("run_id", snapshot.RunID),
			originlog.String("scenario_name", snapshot.ScenarioName),
		)
		return snapshot, nil
	}
	module.Logger().Debug(
		"机器人运行创建",
		originlog.String("run_id", snapshot.RunID),
		originlog.String("scenario_name", snapshot.ScenarioName),
		originlog.Int64("target_users", snapshot.TargetUsers),
		originlog.Duration("ramp_up", module.config.Workload.RampUp.Duration()),
		originlog.Duration("duration", module.config.Workload.Duration.Duration()),
	)
	runtime := &runRuntime{record: record, robots: make(map[int64]*robotRuntime)}
	current, err := newWorkload(
		record.ctx,
		module.config.Workload.Users,
		module.config.Workload.RampUp.Duration(),
		module.config.Workload.Duration.Duration(),
		func(robotID int64) error {
			return module.DispatchAsync(func(context.Context) { module.launchRobot(runtime, robotID) })
		},
	)
	if err != nil {
		module.controller.finish(record, rpcapi.RobotRunStateFailed, "workload_create", err.Error())
		return rpcapi.RobotRunSnapshot{}, err
	}
	runtime.workload = current
	module.current = runtime
	err = service.DispatchAsyncCompletionResult[error](
		module.Service(), module.rootCtx,
		func(waitCtx context.Context) (error, error) {
			select {
			case scheduleErr := <-current.done:
				return scheduleErr, nil
			case <-waitCtx.Done():
				return waitCtx.Err(), waitCtx.Err()
			}
		},
		func(_ context.Context, scheduleErr error, completionErr error) {
			if completionErr != nil && scheduleErr == nil {
				scheduleErr = completionErr
			}
			module.onScheduleDone(runtime, scheduleErr)
		},
	)
	if err != nil {
		module.current = nil
		current.stop()
		module.controller.finish(record, rpcapi.RobotRunStateFailed, "completion_reserve", err.Error())
		return rpcapi.RobotRunSnapshot{}, err
	}
	current.start()
	running := module.controller.markRunning(record)
	module.Logger().Debug(
		"机器人运行已启动",
		originlog.String("run_id", running.RunID),
		originlog.String("scenario_name", running.ScenarioName),
		originlog.Int64("target_users", running.TargetUsers),
	)
	return running, nil
}

// StopRun 只发起取消并返回stopping；终态由GetRun轮询取得。
func (module *Module) StopRun(request rpcapi.StopRobotRunRequest) (rpcapi.RobotRunSnapshot, error) {
	snapshot, record, err := module.controller.requestStop(request.RunID)
	if err != nil {
		return rpcapi.RobotRunSnapshot{}, controlError(err)
	}
	if record != nil {
		module.Logger().Debug("机器人运行收到停止请求", originlog.String("run_id", snapshot.RunID))
	}
	return snapshot, nil
}

// GetRun 返回控制器的不可变聚合快照。
func (module *Module) GetRun(request rpcapi.GetRobotRunRequest) (rpcapi.RobotRunSnapshot, error) {
	snapshot, err := module.controller.get(request.RunID)
	if err != nil {
		return rpcapi.RobotRunSnapshot{}, controlError(err)
	}
	return snapshot, nil
}

func controlError(err error) error {
	switch {
	case errors.Is(err, errRunActive), errors.Is(err, errRequestConflict):
		return errs.NewMessage(errs.CodeAdminStateConflict, err.Error())
	case errors.Is(err, errRunNotFound):
		return errs.NewMessage(errs.CodeInvalidArgument, err.Error())
	default:
		return errs.NewMessage(errs.CodeInvalidArgument, err.Error())
	}
}

func (module *Module) launchRobot(runtime *runRuntime, robotID int64) {
	if module.current != runtime || runtime.scheduleDone || runtime.record.ctx.Err() != nil {
		return
	}
	if _, exists := runtime.robots[robotID]; exists {
		module.failRun(runtime, "duplicate_robot", "重复robot_id")
		return
	}
	robot := &robotRuntime{id: robotID}
	runtime.robots[robotID] = robot
	module.controller.robotStarted(runtime.record)
	if err := module.startAttempt(runtime, robot); err != nil {
		module.handleAttemptFailure(runtime, robot, err)
	}
}

func (module *Module) startAttempt(runtime *runRuntime, robot *robotRuntime) error {
	robot.attempts++
	robot.attemptErr = nil
	module.Logger().Debug(
		"机器人开始场景尝试",
		originlog.String("run_id", runtime.record.snapshot.RunID),
		originlog.Int64("robot_id", robot.id),
		originlog.Int("attempt", robot.attempts),
	)
	player, err := virtualplayer.New(
		robot.id,
		module.timers.schedule,
		func(message virtualplayer.InboundMessage) { module.handlePush(runtime, robot, message) },
	)
	if err != nil {
		return err
	}
	instance, err := module.Create(
		runtime.record.snapshot.ScenarioName,
		blueprintmodule.WithKey(fmt.Sprintf("run:%s robot:%d", runtime.record.snapshot.RunID, robot.id)),
	)
	if err != nil {
		player.Close()
		return err
	}
	// Instance.Start 会内联执行直到首次Yield，节点在返回前就会查询VirtualPlayer。
	// 因此必须先把Player和Instance发布到当前robotRuntime。
	robot.player = player
	robot.instance = instance
	execution, err := instance.Start(
		runtime.record.ctx,
		module.config.Blueprint.EntranceID,
		blueprintmodule.PortInt(robot.id),
	)
	if err != nil {
		_ = instance.Close()
		player.Close()
		robot.player, robot.instance = nil, nil
		return err
	}
	robot.execution = execution
	if err = execution.OnComplete(func(_ context.Context, _ blueprintmodule.PortArray, completionErr error) {
		module.onRobotComplete(runtime, robot, execution, completionErr)
	}); err != nil {
		execution.Cancel()
		_ = instance.Close()
		player.Close()
		robot.player, robot.instance, robot.execution = nil, nil, nil
		return err
	}
	return nil
}

func (module *Module) onRobotComplete(
	runtime *runRuntime,
	robot *robotRuntime,
	execution *blueprintmodule.Execution,
	completionErr error,
) {
	if module.current != runtime || runtime.robots[robot.id] != robot || robot.execution != execution {
		return
	}
	module.closeAttempt(robot)
	if runtime.scheduleDone || runtime.record.ctx.Err() != nil {
		delete(runtime.robots, robot.id)
		module.controller.robotFinished(runtime.record, robotOutcomeCanceled)
		module.finishRunIfReady(runtime)
		return
	}
	if completionErr == nil && robot.attemptErr != nil {
		completionErr = robot.attemptErr
	}
	if completionErr == nil {
		delete(runtime.robots, robot.id)
		outcome := robotOutcomeSucceeded
		if robot.firstFailed {
			outcome = robotOutcomeFlaky
		}
		module.controller.robotFinished(runtime.record, outcome)
		module.Logger().Debug(
			"机器人场景执行完成",
			originlog.String("run_id", runtime.record.snapshot.RunID),
			originlog.Int64("robot_id", robot.id),
			originlog.Int("attempt", robot.attempts),
			originlog.String("outcome", string(outcome)),
		)
		return
	}
	module.handleAttemptFailure(runtime, robot, completionErr)
}

// markAttemptFailure 保留本次执行的首次确定失败，避免失败分支完成清理后被误计为成功。
func (module *Module) markAttemptFailure(robot *robotRuntime, failure error) {
	if robot != nil && failure != nil && robot.attemptErr == nil {
		robot.attemptErr = failure
	}
}

func (module *Module) handleAttemptFailure(runtime *runRuntime, robot *robotRuntime, failure error) {
	if robot != nil {
		module.Logger().Debug(
			"机器人场景尝试失败",
			originlog.String("run_id", runtime.record.snapshot.RunID),
			originlog.Int64("robot_id", robot.id),
			originlog.Int("attempt", robot.attempts),
			originlog.Bool("will_retry", robot.attempts <= module.config.Workload.ScenarioRetryCount),
			originlog.Err(failure),
		)
	}
	module.closeAttempt(robot)
	if module.current != runtime || runtime.record.ctx.Err() != nil {
		delete(runtime.robots, robot.id)
		module.controller.robotFinished(runtime.record, robotOutcomeCanceled)
		module.finishRunIfReady(runtime)
		return
	}
	robot.firstFailed = true
	if robot.attempts <= module.config.Workload.ScenarioRetryCount {
		if err := module.startAttempt(runtime, robot); err == nil {
			return
		} else {
			failure = errors.Join(failure, err)
		}
	}
	delete(runtime.robots, robot.id)
	module.controller.robotFinished(runtime.record, robotOutcomeFailed)
	if runtime.failureCode == "" {
		runtime.failureCode = "scenario_failed"
		runtime.failureText = failure.Error()
	}
}

func (module *Module) closeAttempt(robot *robotRuntime) {
	if robot == nil {
		return
	}
	if robot.pushWaiter != nil && robot.pushWaiter.cancelTimeout != nil {
		robot.pushWaiter.cancelTimeout()
	}
	robot.pushWaiter = nil
	if robot.execution != nil && !robot.execution.IsDone() {
		robot.execution.Cancel()
	}
	if robot.instance != nil {
		_ = robot.instance.Close()
	}
	if robot.player != nil {
		robot.player.Close()
	}
	robot.player, robot.instance, robot.execution = nil, nil, nil
}

func (module *Module) onScheduleDone(runtime *runRuntime, scheduleErr error) {
	if module.current != runtime || runtime.scheduleDone {
		return
	}
	runtime.scheduleDone = true
	runtime.scheduleErr = scheduleErr
	if scheduleErr != nil && !errors.Is(scheduleErr, context.Canceled) && runtime.failureCode == "" {
		runtime.failureCode = "workload_failed"
		runtime.failureText = scheduleErr.Error()
	}
	runtime.record.cancel()
	for _, robot := range runtime.robots {
		if robot.player != nil {
			robot.player.Close()
		}
		if robot.execution != nil {
			robot.execution.Cancel()
		}
	}
	module.finishRunIfReady(runtime)
}

func (module *Module) finishRunIfReady(runtime *runRuntime) {
	if module.current != runtime || !runtime.scheduleDone || len(runtime.robots) != 0 {
		return
	}
	state := rpcapi.RobotRunStateSucceeded
	if runtime.failureCode != "" {
		state = rpcapi.RobotRunStateFailed
	} else if runtime.record.snapshot.State == rpcapi.RobotRunStateStopping ||
		errors.Is(runtime.scheduleErr, context.Canceled) {
		state = rpcapi.RobotRunStateCanceled
	}
	snapshot := module.controller.finish(runtime.record, state, runtime.failureCode, runtime.failureText)
	module.Logger().Debug(
		"机器人运行结束",
		originlog.String("run_id", snapshot.RunID),
		originlog.Int32("run_state", int32(snapshot.State)),
		originlog.Int64("started_robots", snapshot.StartedRobots),
		originlog.Int64("succeeded_robots", snapshot.SucceededRobots),
		originlog.Int64("failed_robots", snapshot.FailedRobots),
		originlog.String("failure_code", snapshot.FailureCode),
	)
	if runtime.workload != nil {
		runtime.workload.stop()
	}
	module.current = nil
}

func (module *Module) failRun(runtime *runRuntime, code string, message string) {
	if module.current != runtime {
		return
	}
	if runtime.failureCode == "" {
		runtime.failureCode, runtime.failureText = code, message
	}
	runtime.record.cancel()
}

func (module *Module) forceCleanupCurrent() {
	runtime := module.current
	if runtime == nil {
		return
	}
	runtime.record.cancel()
	if runtime.workload != nil {
		runtime.workload.stop()
	}
	for id, robot := range runtime.robots {
		module.closeAttempt(robot)
		delete(runtime.robots, id)
	}
	module.controller.finish(runtime.record, rpcapi.RobotRunStateCanceled, "", "")
	module.current = nil
}

func (module *Module) findRobot(robotID int64) (*runRuntime, *robotRuntime, error) {
	if robotID <= 0 || module.current == nil {
		return nil, nil, errors.New("机器人运行不存在")
	}
	robot := module.current.robots[robotID]
	if robot == nil || robot.player == nil {
		return nil, nil, errors.New("机器人不存在或已经结束")
	}
	return module.current, robot, nil
}
