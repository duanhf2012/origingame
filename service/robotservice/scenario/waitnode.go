package scenario

import (
	"errors"
	"strings"
	"time"

	"github.com/duanhf2012/origin/v3/sysmodule/blueprintmodule"
	commonpb "origingame/protocol/common"
)

type waitNode struct {
	blueprintmodule.BaseExecNode
	module *Module
}

func (*waitNode) GetName() string { return "RobotWait" }

func (node *waitNode) Exec() (int, error) {
	value, ok := node.GetInPortStr(1)
	if !ok {
		return -1, errors.New("RobotWait缺少duration")
	}
	duration, err := time.ParseDuration(strings.TrimSpace(string(value)))
	if err != nil || duration <= 0 || duration > time.Hour {
		return -1, errors.New("RobotWait duration必须在0到1h之间")
	}
	handle, err := node.Yield(0)
	if err != nil {
		return -1, err
	}
	_, err = node.module.timers.schedule(duration, func() {
		node.module.resumeTo(handle, 0, blueprintmodule.PortInt(0))
	})
	if err != nil {
		node.module.resumeTo(handle, 1, blueprintmodule.PortInt(nodeErrorIO))
	}
	return -1, blueprintmodule.ErrExecutionSuspended
}

type waitMessageNode struct {
	blueprintmodule.BaseExecNode
	module *Module
}

func (*waitMessageNode) GetName() string { return "RobotWaitMessage" }

func (node *waitMessageNode) Exec() (int, error) {
	robotID, robotOK := node.GetInPortInt(1)
	messageID, messageOK := node.GetInPortInt(2)
	timeoutValue, timeoutOK := node.GetInPortStr(3)
	if !robotOK || !messageOK || !timeoutOK || robotID <= 0 || messageID < 0 || messageID > 65535 {
		return -1, errors.New("RobotWaitMessage输入无效")
	}
	timeout, err := time.ParseDuration(strings.TrimSpace(string(timeoutValue)))
	if err != nil || timeout <= 0 || timeout > time.Hour {
		return -1, errors.New("RobotWaitMessage timeout必须在0到1h之间")
	}
	_, robot, err := node.module.findRobot(int64(robotID))
	if err != nil {
		return -1, err
	}
	handle, err := node.Yield(0)
	if err != nil {
		return -1, err
	}
	if err = node.module.waitForMessage(robot, commonpb.MessageID(messageID), timeout, handle); err != nil {
		node.module.markAttemptFailure(robot, err)
		node.module.resumeTo(handle, 2, blueprintmodule.PortInt(nodeErrorInvalidState), blueprintmodule.PortInt(0))
	}
	return -1, blueprintmodule.ErrExecutionSuspended
}

type disconnectNode struct {
	blueprintmodule.BaseExecNode
	module *Module
}

func (*disconnectNode) GetName() string { return "RobotDisconnect" }

func (node *disconnectNode) Exec() (int, error) {
	robotID, ok := node.GetInPortInt(1)
	if !ok || robotID <= 0 {
		return -1, errors.New("RobotDisconnect缺少有效robot_id")
	}
	_, robot, err := node.module.findRobot(int64(robotID))
	if err != nil {
		return -1, err
	}
	robot.player.Close()
	return 0, nil
}
