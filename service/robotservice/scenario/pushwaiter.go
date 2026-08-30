package scenario

import (
	"errors"
	"time"

	"github.com/duanhf2012/origin/v3/sysmodule/blueprintmodule"
	commonpb "origingame/protocol/common"
	"origingame/service/robotservice/virtualplayer"
)

func (module *RobotScenarioModule) waitForMessage(
	robot *robotRuntime,
	messageID commonpb.MessageID,
	timeout time.Duration,
	handle *blueprintmodule.YieldHandle,
) error {
	if robot == nil || robot.pushWaiter != nil || handle == nil {
		return errors.New("机器人已有活动WaitMessage")
	}
	started := time.Now()
	for index, message := range robot.inbox {
		if message.MessageID != messageID {
			continue
		}
		copy(robot.inbox[index:], robot.inbox[index+1:])
		robot.inbox = robot.inbox[:len(robot.inbox)-1]
		module.resumeTo(handle, 0, blueprintmodule.PortInt(0), durationMilliseconds(started))
		return nil
	}
	waiter := &messageWaiter{
		messageID: messageID, startedAtMS: started.UnixMilli(), handle: handle,
	}
	cancel, err := module.timers.schedule(timeout, func() {
		if robot.pushWaiter != waiter {
			return
		}
		robot.pushWaiter = nil
		module.resumeTo(
			handle, 1, blueprintmodule.PortInt(nodeErrorTimeout),
			blueprintmodule.PortInt(time.Since(started).Milliseconds()),
		)
	})
	if err != nil {
		return err
	}
	waiter.cancelTimeout = cancel
	robot.pushWaiter = waiter
	return nil
}

func (module *RobotScenarioModule) handlePush(runtime *runRuntime, robot *robotRuntime, message virtualplayer.InboundMessage) {
	if module.current != runtime || runtime.robots[robot.id] != robot {
		return
	}
	if waiter := robot.pushWaiter; waiter != nil && waiter.messageID == message.MessageID {
		robot.pushWaiter = nil
		if waiter.cancelTimeout != nil {
			waiter.cancelTimeout()
		}
		duration := time.Now().UnixMilli() - waiter.startedAtMS
		if duration < 0 {
			duration = 0
		}
		module.resumeTo(
			waiter.handle, 0, blueprintmodule.PortInt(0), blueprintmodule.PortInt(duration),
		)
		return
	}
	if len(robot.inbox) >= pushInboxCapacity {
		module.failRun(runtime, "push_inbox_full", "机器人主动推送inbox已满")
		if robot.execution != nil {
			robot.execution.Cancel()
		}
		return
	}
	robot.inbox = append(robot.inbox, message)
}
