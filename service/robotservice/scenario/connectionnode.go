package scenario

import (
	"context"
	"errors"
	"time"

	originlog "github.com/duanhf2012/origin/v3/log"
	"github.com/duanhf2012/origin/v3/sysmodule/blueprintmodule"
	"github.com/duanhf2012/origin/v3/sysmodule/network"
	"github.com/duanhf2012/origin/v3/sysmodule/network/tcp"
	"origingame/service/robotservice/virtualplayer"
)

type connectGatewayNode struct {
	blueprintmodule.BaseExecNode
	module *RobotScenarioModule
}

func (*connectGatewayNode) GetName() string { return "RobotConnectGateway" }

func (node *connectGatewayNode) Exec() (int, error) {
	robotID, ok := node.GetInPortInt(1)
	if !ok || robotID <= 0 {
		return -1, errors.New("RobotConnectGateway缺少有效robot_id")
	}
	runtime, robot, err := node.module.findRobot(int64(robotID))
	if err != nil || robot.player.State() != virtualplayer.StateAuthenticated {
		return -1, errors.New("RobotConnectGateway机器人状态无效")
	}
	handle, err := node.Yield(0)
	if err != nil {
		return -1, err
	}
	started := time.Now()
	node.module.Logger().Debug(
		"机器人开始连接 Gateway",
		originlog.String("run_id", runtime.record.snapshot.RunID),
		originlog.Int64("robot_id", robot.id),
		originlog.String("gateway_address", robot.player.GatewayAddress()),
	)
	err = node.module.executor.submit(ioJob{
		ctx: runtime.record.ctx,
		work: func(ctx context.Context) (any, error) {
			options := tcp.DefaultDialOptions(robot.player.GatewayHandler())
			options.Network.MaxMessageSize = 4 * 1024
			dialer, dialErr := tcp.NewDialer(robot.player.GatewayAddress(), options)
			if dialErr != nil {
				return nil, dialErr
			}
			return dialer.Dial(ctx, node.module.Service())
		},
		complete: func(result any, completionErr error) {
			session, _ := result.(network.Session)
			if node.module.current != runtime || runtime.robots[robot.id] != robot || robot.player == nil {
				if session != nil {
					session.Close(context.Canceled)
				}
				return
			}
			if completionErr != nil || session == nil || robot.player.BindSession(session) != nil {
				if session != nil {
					session.Close(completionErr)
				}
				if completionErr == nil {
					completionErr = errors.New("Gateway连接结果无效")
				}
				node.module.Logger().Debug(
					"机器人连接 Gateway 失败",
					originlog.String("run_id", runtime.record.snapshot.RunID),
					originlog.Int64("robot_id", robot.id),
					originlog.Duration("duration", time.Since(started)),
					originlog.Err(completionErr),
				)
				node.module.markAttemptFailure(robot, completionErr)
				node.module.resumeTo(handle, 1, blueprintmodule.PortInt(nodeErrorIO), durationMilliseconds(started))
				return
			}
			node.module.Logger().Debug(
				"机器人已连接 Gateway",
				originlog.String("run_id", runtime.record.snapshot.RunID),
				originlog.Int64("robot_id", robot.id),
				originlog.String("gateway_address", robot.player.GatewayAddress()),
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
			"机器人 Gateway 连接任务提交失败",
			originlog.String("run_id", runtime.record.snapshot.RunID),
			originlog.Int64("robot_id", robot.id),
			originlog.Err(err),
		)
		node.module.markAttemptFailure(robot, err)
		node.module.resumeTo(handle, 1, blueprintmodule.PortInt(code), durationMilliseconds(started))
	}
	return -1, blueprintmodule.ErrExecutionSuspended
}
