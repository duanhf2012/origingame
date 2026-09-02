package scenario

import (
	"context"
	"errors"
	"sync"
	"time"
)

// workload 只拥有一个可停止调度goroutine，不执行机器人业务逻辑。
type workload struct {
	ctx      context.Context    // 工作负载取消上下文。
	cancel   context.CancelFunc // 工作负载取消函数。
	users    int64              // 目标机器人数量。
	rampUp   time.Duration      // 渐进启动时长。
	duration time.Duration      // 运行总时长。
	launch   func(int64) error  // 启动单个机器人的回调。

	done chan error     // 运行完成结果。
	wg   sync.WaitGroup // 等待调度协程退出。
}

func newWorkload(
	parent context.Context,
	users int64,
	rampUp time.Duration,
	duration time.Duration,
	launch func(int64) error,
) (*workload, error) {
	if parent == nil || users <= 0 || rampUp <= 0 || duration <= 0 || launch == nil {
		return nil, errors.New("RobotService Workload参数无效")
	}
	ctx, cancel := context.WithCancel(parent)
	return &workload{
		ctx: ctx, cancel: cancel, users: users, rampUp: rampUp, duration: duration,
		launch: launch, done: make(chan error, 1),
	}, nil
}

func (workload *workload) start() {
	workload.wg.Add(1)
	go func() {
		defer workload.wg.Done()
		workload.done <- workload.run()
		close(workload.done)
	}()
}

func (workload *workload) run() error {
	interval := workload.rampUp / time.Duration(workload.users)
	if interval <= 0 {
		interval = time.Nanosecond
	}
	for robotID := int64(1); robotID <= workload.users; robotID++ {
		if err := waitRealDuration(workload.ctx, interval); err != nil {
			return err
		}
		if err := workload.launch(robotID); err != nil {
			return err
		}
	}
	return waitRealDuration(workload.ctx, workload.duration)
}

func waitRealDuration(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (workload *workload) stop() {
	if workload == nil {
		return
	}
	workload.cancel()
	workload.wg.Wait()
}
