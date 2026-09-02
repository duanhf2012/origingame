package scenario

import "github.com/duanhf2012/origin/v3/sysmodule/blueprintmodule"

type robotStartNode struct {
	blueprintmodule.BaseExecNode // 蓝图节点基础能力。
}

// GetName 返回入口工厂基类名；JSON/蓝图中的 _1001 后缀是具体 EntranceID。
func (*robotStartNode) GetName() string    { return "Entrance_RobotStart" }
func (*robotStartNode) Exec() (int, error) { return 0, nil }
