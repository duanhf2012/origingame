package scenario

import (
	"context"
	"errors"
	"sync"
)

var errIOExecutorFull = errors.New("RobotService I/O队列已满")

type ioJob struct {
	ctx      context.Context
	work     func(context.Context) (any, error)
	complete func(any, error)
}

// ioExecutor 使用固定Worker执行HTTP登录和TCP拨号等阻塞冷路径。
type ioExecutor struct {
	workers  int
	queue    chan ioJob
	dispatch func(func(context.Context)) error

	mu     sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
	closed bool
	wg     sync.WaitGroup
}

func newIOExecutor(workers int, queueMessages int, dispatch func(func(context.Context)) error) (*ioExecutor, error) {
	if workers <= 0 || queueMessages <= 0 || dispatch == nil {
		return nil, errors.New("RobotService I/O Executor配置无效")
	}
	return &ioExecutor{
		workers: workers, queue: make(chan ioJob, queueMessages), dispatch: dispatch,
	}, nil
}

func (executor *ioExecutor) start(parent context.Context) error {
	if executor == nil || parent == nil {
		return errors.New("RobotService I/O Executor启动参数无效")
	}
	executor.mu.Lock()
	defer executor.mu.Unlock()
	if executor.ctx != nil || executor.closed {
		return errors.New("RobotService I/O Executor不能重复启动")
	}
	executor.ctx, executor.cancel = context.WithCancel(parent)
	for index := 0; index < executor.workers; index++ {
		executor.wg.Add(1)
		go executor.worker()
	}
	return nil
}

func (executor *ioExecutor) submit(job ioJob) error {
	if executor == nil || job.ctx == nil || job.work == nil || job.complete == nil {
		return errors.New("RobotService I/O Job无效")
	}
	executor.mu.Lock()
	defer executor.mu.Unlock()
	if executor.ctx == nil || executor.closed {
		return errors.New("RobotService I/O Executor未运行")
	}
	select {
	case executor.queue <- job:
		return nil
	default:
		return errIOExecutorFull
	}
}

func (executor *ioExecutor) worker() {
	defer executor.wg.Done()
	for {
		select {
		case <-executor.ctx.Done():
			return
		case job := <-executor.queue:
			executor.execute(job)
		}
	}
}

func (executor *ioExecutor) execute(job ioJob) {
	ctx, cancel := context.WithCancel(job.ctx)
	stop := context.AfterFunc(executor.ctx, cancel)
	result, err := job.work(ctx)
	stop()
	cancel()
	_ = executor.dispatch(func(context.Context) { job.complete(result, err) })
}

func (executor *ioExecutor) stop() {
	if executor == nil {
		return
	}
	executor.mu.Lock()
	if executor.closed {
		executor.mu.Unlock()
		return
	}
	executor.closed = true
	if executor.cancel != nil {
		executor.cancel()
	}
	executor.mu.Unlock()
	executor.wg.Wait()
}
