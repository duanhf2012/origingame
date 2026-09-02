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
type ListRobotScenariosRequest struct {
	Limit int32 // 返回的场景数量上限。
}

// ListRobotScenariosResponse 返回当前实例可运行的场景摘要。
type ListRobotScenariosResponse struct {
	Scenarios []RobotScenarioSummary // 当前可启动的场景摘要。
}

// RobotScenarioSummary 是可由控制端选择的已编译场景。
type RobotScenarioSummary struct {
	Name       string // 场景名称。
	EntranceID int64  // 默认入口节点标识。
	IsDefault  bool   // 是否为默认场景。
}

// StartRobotRunRequest 只允许选择场景；负载参数由 RobotService 配置决定。
type StartRobotRunRequest struct {
	RequestID    string // 幂等启动请求标识。
	ScenarioName string // 待启动的场景名称。
}

// StopRobotRunRequest 请求停止指定运行。
type StopRobotRunRequest struct {
	RunID string // 待停止的运行标识。
}

// GetRobotRunRequest 查询指定运行；RunID 为空时查询当前或最近运行。
type GetRobotRunRequest struct {
	RunID string // 待查询的运行标识。
}

// RobotRunSnapshot 是控制面使用的低频聚合快照。
type RobotRunSnapshot struct {
	RunID        string        // 运行实例标识。
	RequestID    string        // 对应启动请求标识。
	ScenarioName string        // 运行场景名称。
	State        RobotRunState // 当前生命周期状态。
	StartedAtMS  int64         // 真实系统启动时间戳（毫秒）。
	FinishedAtMS int64         // 真实系统结束时间戳（毫秒）。

	TargetUsers     int64 // 计划启动的机器人数量。
	StartedRobots   int64 // 已创建的机器人数量。
	ActiveRobots    int64 // 当前活跃机器人数量。
	SucceededRobots int64 // 成功完成的机器人数量。
	FailedRobots    int64 // 失败的机器人数量。
	FlakyRobots     int64 // 经重试恢复的机器人数量。
	Replacements    int64 // 已启动的替补机器人数量。

	FailureCode    string // 运行失败分类。
	FailureMessage string // 首个失败说明。
}
