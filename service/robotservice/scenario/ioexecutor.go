package scenario

import (
	"context"
	"errors"
	"sync"
)

var errIOExecutorFull = errors.New("RobotService I/O队列已满")

type ioJob struct {
	ctx      context.Context                    // 可取消的 I/O 请求上下文。
	work     func(context.Context) (any, error) // 阻塞 I/O 工作函数。
	complete func(any, error)                   // 回到 Service 的完成回调。
}

// ioExecutor 使用固定Worker执行HTTP登录和TCP拨号等阻塞冷路径。
type ioExecutor struct {
	workers  int                               // 固定 Worker 数。
	queue    chan ioJob                        // 有界 I/O 任务队列。
	dispatch func(func(context.Context)) error // 将回调投递回 Service。

	mu     sync.Mutex         // 保护启动和关闭状态。
	ctx    context.Context    // Worker 共享取消上下文。
	cancel context.CancelFunc // Worker 取消函数。
	closed bool               // 是否已停止接收任务。
	wg     sync.WaitGroup     // 等待全部 Worker 退出。
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
