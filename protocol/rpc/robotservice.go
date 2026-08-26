package rpcapi

import "context"

// RobotService 是隔离测试环境中的机器人运行控制入口。
//
//origin:rpc
type RobotService interface {
	// ListScenarios 返回当前实例已经加载并允许启动的场景。
	ListScenarios(context.Context, ListRobotScenariosRequest) (ListRobotScenariosResponse, error)
	// StartRun 幂等启动一个使用服务端固定负载配置的场景。
	StartRun(context.Context, StartRobotRunRequest) (RobotRunSnapshot, error)
	// StopRun 幂等请求停止运行，不等待全部机器人完成清理。
	StopRun(context.Context, StopRobotRunRequest) (RobotRunSnapshot, error)
	// GetRun 返回指定运行、当前活动运行或最近一次运行的聚合快照。
	GetRun(context.Context, GetRobotRunRequest) (RobotRunSnapshot, error)
}

// RobotRunState 是一次机器人运行的生命周期状态。
type RobotRunState uint8

const (
	RobotRunStateStarting RobotRunState = iota + 1
	RobotRunStateRunning
	RobotRunStateStopping
	RobotRunStateSucceeded
	RobotRunStateFailed
	RobotRunStateCanceled
)

// ListRobotScenariosRequest 使用有界结果数量；Limit 为零时使用服务端默认上限。
type ListRobotScenariosRequest struct{ Limit int32 }

// ListRobotScenariosResponse 返回当前实例可运行的场景摘要。
type ListRobotScenariosResponse struct {
	Scenarios []RobotScenarioSummary
}

// RobotScenarioSummary 是可由控制端选择的已编译场景。
type RobotScenarioSummary struct {
	Name       string
	EntranceID int64
	IsDefault  bool
}

// StartRobotRunRequest 只允许选择场景；负载参数由 RobotService 配置决定。
type StartRobotRunRequest struct {
	RequestID    string
	ScenarioName string
}

// StopRobotRunRequest 请求停止指定运行。
type StopRobotRunRequest struct{ RunID string }

// GetRobotRunRequest 查询指定运行；RunID 为空时查询当前或最近运行。
type GetRobotRunRequest struct{ RunID string }

// RobotRunSnapshot 是控制面使用的低频聚合快照。
type RobotRunSnapshot struct {
	RunID        string
	RequestID    string
	ScenarioName string
	State        RobotRunState
	StartedAtMS  int64
	FinishedAtMS int64

	TargetUsers     int64
	StartedRobots   int64
	ActiveRobots    int64
	SucceededRobots int64
	FailedRobots    int64
	FlakyRobots     int64
	Replacements    int64

	FailureCode    string
	FailureMessage string
}
