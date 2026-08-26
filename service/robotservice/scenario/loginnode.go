package scenario

import (
	"context"
	"errors"
	"fmt"
	"time"

	originlog "github.com/duanhf2012/origin/v3/log"
	"github.com/duanhf2012/origin/v3/sysmodule/blueprintmodule"
	"origingame/service/robotservice/virtualplayer"
)

type httpLoginNode struct {
	blueprintmodule.BaseExecNode
	module *Module
}

func (*httpLoginNode) GetName() string { return "RobotHTTPLogin" }

func (node *httpLoginNode) Exec() (int, error) {
	robotID, ok := node.GetInPortInt(1)
	if !ok || robotID <= 0 {
		return -1, errors.New("RobotHTTPLogin缺少有效robot_id")
	}
	runtime, robot, err := node.module.findRobot(int64(robotID))
	if err != nil || robot.player.State() != virtualplayer.StateCreated {
		return -1, errors.New("RobotHTTPLogin机器人状态无效")
	}
	handle, err := node.Yield(0)
	if err != nil {
		return -1, err
	}
	started := time.Now()
	node.module.Logger().Debug(
		"机器人开始 HTTP 登录",
		originlog.String("run_id", runtime.record.snapshot.RunID),
		originlog.Int64("robot_id", robot.id),
	)
	err = node.module.executor.submit(ioJob{
		ctx: runtime.record.ctx,
		work: func(ctx context.Context) (any, error) {
			return node.module.login.Login(ctx, int64(robotID))
		},
		complete: func(result any, completionErr error) {
			if node.module.current != runtime || runtime.robots[robot.id] != robot || robot.player == nil {
				return
			}
			if completionErr != nil {
				node.module.Logger().Debug(
					"机器人 HTTP 登录失败",
					originlog.String("run_id", runtime.record.snapshot.RunID),
					originlog.Int64("robot_id", robot.id),
					originlog.Duration("duration", time.Since(started)),
					originlog.Err(completionErr),
				)
				node.module.markAttemptFailure(robot, fmt.Errorf("HTTP登录: %w", completionErr))
				node.module.resumeTo(handle, 1, blueprintmodule.PortInt(nodeErrorIO), durationMilliseconds(started))
				return
			}
			loginResult, valid := result.(virtualplayer.LoginResult)
			if !valid {
				node.module.Logger().Debug(
					"机器人 HTTP 登录响应类型无效",
					originlog.String("run_id", runtime.record.snapshot.RunID),
					originlog.Int64("robot_id", robot.id),
				)
				node.module.markAttemptFailure(robot, errors.New("HTTP登录响应无效"))
				node.module.resumeTo(handle, 1, blueprintmodule.PortInt(nodeErrorProtocol), durationMilliseconds(started))
				return
			}
			if applyErr := robot.player.ApplyLogin(loginResult); applyErr != nil {
				node.module.Logger().Debug(
					"机器人 HTTP 登录状态转换失败",
					originlog.String("run_id", runtime.record.snapshot.RunID),
					originlog.Int64("robot_id", robot.id),
					originlog.Err(applyErr),
				)
				node.module.markAttemptFailure(robot, errors.New("HTTP登录响应无效"))
				node.module.resumeTo(handle, 1, blueprintmodule.PortInt(nodeErrorProtocol), durationMilliseconds(started))
				return
			}
			node.module.Logger().Debug(
				"机器人 HTTP 登录成功",
				originlog.String("run_id", runtime.record.snapshot.RunID),
				originlog.Int64("robot_id", robot.id),
				originlog.String("gateway_address", loginResult.GatewayAddress),
				originlog.Duration("duration", time.Since(started)),
			)
			node.module.resumeTo(handle, 0, blueprintmodule.PortInt(0), durationMilliseconds(started))
		},
	})
	if err != nil {
		code := nodeErrorIO
		if errors.Is(err, errIOExecutorFull) {
			code = nodeErrorOverloaded
		}
		node.module.Logger().Debug(
			"机器人 HTTP 登录任务提交失败",
			originlog.String("run_id", runtime.record.snapshot.RunID),
			originlog.Int64("robot_id", robot.id),
			originlog.Err(err),
		)
		node.module.markAttemptFailure(robot, err)
		node.module.resumeTo(handle, 1, blueprintmodule.PortInt(code), durationMilliseconds(started))
	}
	return -1, blueprintmodule.ErrExecutionSuspended
}
