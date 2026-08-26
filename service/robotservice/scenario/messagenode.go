package scenario

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	originlog "github.com/duanhf2012/origin/v3/log"
	"github.com/duanhf2012/origin/v3/sysmodule/blueprintmodule"
	"google.golang.org/protobuf/proto"
	commonpb "origingame/protocol/common"
	"origingame/service/robotservice/virtualplayer"
)

const protocolRequestTimeout = 15 * time.Second

type startHeartbeatNode struct {
	blueprintmodule.BaseExecNode
	module *Module
}

func (*startHeartbeatNode) GetName() string { return "RobotStartHeartbeat" }

func (node *startHeartbeatNode) Exec() (int, error) {
	robotID, robotOK := node.GetInPortInt(1)
	intervalValue, intervalOK := node.GetInPortStr(2)
	if !robotOK || !intervalOK || robotID <= 0 {
		return -1, errors.New("RobotStartHeartbeat输入无效")
	}
	interval, err := time.ParseDuration(strings.TrimSpace(string(intervalValue)))
	if err != nil || interval <= 0 || interval > time.Hour {
		return -1, errors.New("RobotStartHeartbeat interval必须在0到1h之间")
	}
	runtime, robot, err := node.module.findRobot(int64(robotID))
	if err != nil || robot.player.State() != virtualplayer.StateOnline {
		return -1, errors.New("RobotStartHeartbeat机器人状态无效")
	}
	err = robot.player.StartHeartbeat(interval, protocolRequestTimeout, func(cause error) {
		if node.module.current != runtime || runtime.robots[robot.id] != robot {
			return
		}
		node.module.markAttemptFailure(robot, fmt.Errorf("后台心跳: %w", cause))
		node.module.Logger().Debug(
			"机器人后台心跳失败",
			originlog.String("run_id", runtime.record.snapshot.RunID),
			originlog.Int64("robot_id", robot.id),
			originlog.Err(cause),
		)
		if robot.execution != nil {
			robot.execution.Cancel()
		}
	})
	if err != nil {
		node.module.markAttemptFailure(robot, err)
		node.SetOutPortInt(2, blueprintmodule.PortInt(nodeErrorInvalidState))
		return 1, nil
	}
	node.SetOutPortInt(2, blueprintmodule.PortInt(0))
	node.module.Logger().Debug(
		"机器人后台心跳已启动",
		originlog.String("run_id", runtime.record.snapshot.RunID),
		originlog.Int64("robot_id", robot.id),
		originlog.Duration("interval", interval),
	)
	return 0, nil
}

type loginPlayerNode struct {
	blueprintmodule.BaseExecNode
	module *Module
}

func (*loginPlayerNode) GetName() string { return "RobotLoginPlayer" }

func (node *loginPlayerNode) Exec() (int, error) {
	robotID, robotOK := node.GetInPortInt(1)
	showAreaID, areaOK := node.GetInPortInt(2)
	if !robotOK || !areaOK || robotID <= 0 || showAreaID <= 0 {
		return -1, errors.New("RobotLoginPlayer输入无效")
	}
	runtime, robot, err := node.module.findRobot(int64(robotID))
	if err != nil || robot.player.State() != virtualplayer.StateConnected {
		return -1, errors.New("RobotLoginPlayer机器人状态无效")
	}
	token, err := robot.player.TokenForLoginPlayer()
	if err != nil {
		return -1, err
	}
	body, err := proto.Marshal(&commonpb.LoginPlayerRequest{Token: token, ShowAreaId: int64(showAreaID)})
	if err != nil {
		return -1, err
	}
	handle, err := node.Yield(0)
	if err != nil {
		return -1, err
	}
	started := time.Now()
	node.module.Logger().Debug(
		"机器人发送 Gateway 登录请求",
		originlog.String("run_id", runtime.record.snapshot.RunID),
		originlog.Int64("robot_id", robot.id),
		originlog.Int64("show_area_id", int64(showAreaID)),
	)
	err = robot.player.SendRequest(
		commonpb.MessageID_LoginPlayerReq,
		commonpb.MessageID_LoginPlayerRes,
		body,
		protocolRequestTimeout,
		func(response virtualplayer.Response) {
			if response.Err != nil {
				node.module.Logger().Debug(
					"机器人 Gateway 登录请求失败",
					originlog.String("run_id", runtime.record.snapshot.RunID),
					originlog.Int64("robot_id", robot.id),
					originlog.Duration("duration", time.Since(started)),
					originlog.Err(response.Err),
				)
				node.module.markAttemptFailure(robot, response.Err)
				code := nodeErrorIO
				if errors.Is(response.Err, context.DeadlineExceeded) {
					code = nodeErrorTimeout
				}
				node.module.resumeTo(handle, 1, blueprintmodule.PortInt(code), durationMilliseconds(started))
				return
			}
			if response.ErrorCode != commonpb.ErrorCode_ERROR_CODE_OK {
				node.module.Logger().Debug(
					"机器人 Gateway 登录返回业务错误",
					originlog.String("run_id", runtime.record.snapshot.RunID),
					originlog.Int64("robot_id", robot.id),
					originlog.Int32("error_code", int32(response.ErrorCode)),
				)
				node.module.markAttemptFailure(robot, fmt.Errorf("LoginPlayer返回错误码%d", response.ErrorCode))
				node.module.resumeTo(handle, 1, blueprintmodule.PortInt(response.ErrorCode), durationMilliseconds(started))
				return
			}
			var result commonpb.LoginPlayerResult
			if proto.Unmarshal(response.Body, &result) != nil || result.RoleInfo == nil || robot.player.MarkOnline() != nil {
				node.module.Logger().Debug(
					"机器人 Gateway 登录响应无效",
					originlog.String("run_id", runtime.record.snapshot.RunID),
					originlog.Int64("robot_id", robot.id),
				)
				node.module.markAttemptFailure(robot, errors.New("LoginPlayer响应无效"))
				node.module.resumeTo(handle, 1, blueprintmodule.PortInt(nodeErrorProtocol), durationMilliseconds(started))
				return
			}
			node.module.Logger().Debug(
				"机器人 Gateway 登录成功",
				originlog.String("run_id", runtime.record.snapshot.RunID),
				originlog.Int64("robot_id", robot.id),
				originlog.Int64("show_area_id", int64(showAreaID)),
				originlog.Duration("duration", time.Since(started)),
			)
			node.module.resumeTo(handle, 0, blueprintmodule.PortInt(0), durationMilliseconds(started))
		},
	)
	if err != nil {
		node.module.markAttemptFailure(robot, err)
		node.module.resumeTo(handle, 1, blueprintmodule.PortInt(nodeErrorInvalidState), durationMilliseconds(started))
	}
	return -1, blueprintmodule.ErrExecutionSuspended
}

type heartbeatNode struct {
	blueprintmodule.BaseExecNode
	module *Module
}

func (*heartbeatNode) GetName() string { return "RobotHeartbeat" }

func (node *heartbeatNode) Exec() (int, error) {
	robotID, ok := node.GetInPortInt(1)
	if !ok || robotID <= 0 {
		return -1, errors.New("RobotHeartbeat缺少有效robot_id")
	}
	_, robot, err := node.module.findRobot(int64(robotID))
	if err != nil || robot.player.State() != virtualplayer.StateOnline {
		return -1, errors.New("RobotHeartbeat机器人状态无效")
	}
	body, err := proto.Marshal(&commonpb.PlayerHeartbeatRequest{})
	if err != nil {
		return -1, err
	}
	handle, err := node.Yield(0)
	if err != nil {
		return -1, err
	}
	started := time.Now()
	err = robot.player.SendRequest(
		commonpb.MessageID_PlayerHeartbeatReq,
		commonpb.MessageID_Ok,
		body,
		protocolRequestTimeout,
		func(response virtualplayer.Response) {
			if response.Err != nil {
				node.module.markAttemptFailure(robot, response.Err)
				code := nodeErrorIO
				if errors.Is(response.Err, context.DeadlineExceeded) {
					code = nodeErrorTimeout
				}
				node.module.resumeTo(handle, 1, blueprintmodule.PortInt(code), durationMilliseconds(started))
				return
			}
			if response.ErrorCode != commonpb.ErrorCode_ERROR_CODE_OK {
				node.module.markAttemptFailure(robot, fmt.Errorf("Heartbeat返回错误码%d", response.ErrorCode))
				node.module.resumeTo(handle, 1, blueprintmodule.PortInt(response.ErrorCode), durationMilliseconds(started))
				return
			}
			node.module.resumeTo(handle, 0, blueprintmodule.PortInt(0), durationMilliseconds(started))
		},
	)
	if err != nil {
		node.module.markAttemptFailure(robot, err)
		node.module.resumeTo(handle, 1, blueprintmodule.PortInt(nodeErrorInvalidState), durationMilliseconds(started))
	}
	return -1, blueprintmodule.ErrExecutionSuspended
}
