package dbservice

import (
	"container/list"
	"context"
	"sync"
	"sync/atomic"

	"github.com/duanhf2012/origin/v3/errs"
	"golang.org/x/sync/semaphore"
)

const (
	// maxPendingRequestsPerKey 限制一个热点 Key 的运行加排队请求数量。
	maxPendingRequestsPerKey = 32
	// maxDispatchKeyBytes 是 dispatch_key 的固定字节长度上限。
	maxDispatchKeyBytes = 128
)

// keyExecutor 让 MongoDB 与 Redis 共用同 Key FIFO 和全局 I/O 并发额度。
// 它不创建 Worker goroutine，等待和 I/O 均由调用方的 Origin Await goroutine 执行。
type keyExecutor struct {
	mu        sync.Mutex
	keyQueues map[string]*keyQueue
	slots     *semaphore.Weighted

	maxInflight int64
	inflight    atomic.Int64
	running     atomic.Int64
}

type keyQueue struct {
	requests list.List
}

type keyTicket struct {
	queue   *keyQueue
	element *list.Element
	ready   chan struct{}
	granted bool
}

// inflightReservation 是进入 Origin Await 前取得的非阻塞请求名额。
type inflightReservation struct {
	executor *keyExecutor
	once     sync.Once
}

func newKeyExecutor(maxIOConcurrency, maxInflightRequests int64) (*keyExecutor, error) {
	if maxIOConcurrency <= 0 ||
		maxInflightRequests <= 0 ||
		maxIOConcurrency > maxInflightRequests {
		return nil, errs.ErrInvalidArgument
	}
	return &keyExecutor{
		keyQueues:   make(map[string]*keyQueue),
		slots:       semaphore.NewWeighted(maxIOConcurrency),
		maxInflight: maxInflightRequests,
	}, nil
}

// tryReserve 在 Service 执行槽内快速预留 inflight 名额；容量已满时立即失败。
func (executor *keyExecutor) tryReserve() (*inflightReservation, bool) {
	for {
		current := executor.inflight.Load()
		if current >= executor.maxInflight {
			return nil, false
		}
		if executor.inflight.CompareAndSwap(current, current+1) {
			return &inflightReservation{executor: executor}, true
		}
	}
}

// release 幂等归还 inflight 名额，确保 Await 失败和 panic 展开时也不会泄漏。
func (reservation *inflightReservation) release() {
	if reservation == nil || reservation.executor == nil {
		return
	}
	reservation.once.Do(func() {
		reservation.executor.inflight.Add(-1)
	})
}

// execute 先取得非空 Key 的 FIFO 执行权，再申请全局 I/O 槽并同步执行 operation。
func (executor *keyExecutor) execute(
	ctx context.Context,
	dispatchKey string,
	operation func(context.Context) error,
) error {
	if ctx == nil || operation == nil || len(dispatchKey) > maxDispatchKeyBytes {
		return errs.ErrInvalidArgument
	}

	releaseKey, err := executor.acquireKey(ctx, dispatchKey)
	if err != nil {
		return err
	}
	if releaseKey != nil {
		defer releaseKey()
	}

	if err := executor.slots.Acquire(ctx, 1); err != nil {
		return err
	}
	executor.running.Add(1)
	defer func() {
		executor.running.Add(-1)
		executor.slots.Release(1)
	}()

	return operation(ctx)
}

// acquireKey 为空 Key 时直接返回；非空 Key 只允许 FIFO 队首继续申请 I/O 槽。
func (executor *keyExecutor) acquireKey(
	ctx context.Context,
	dispatchKey string,
) (func(), error) {
	if dispatchKey == "" {
		return nil, nil
	}

	executor.mu.Lock()
	queue := executor.keyQueues[dispatchKey]
	if queue == nil {
		queue = &keyQueue{}
		executor.keyQueues[dispatchKey] = queue
	}
	if queue.requests.Len() >= maxPendingRequestsPerKey {
		executor.mu.Unlock()
		return nil, errs.ErrServiceQueueFull
	}
	ticket := &keyTicket{
		queue: queue,
		ready: make(chan struct{}),
	}
	ticket.element = queue.requests.PushBack(ticket)
	if queue.requests.Len() == 1 {
		ticket.granted = true
		close(ticket.ready)
	}
	executor.mu.Unlock()

	select {
	case <-ticket.ready:
		var once sync.Once
		return func() {
			once.Do(func() {
				executor.releaseKey(dispatchKey, ticket)
			})
		}, nil
	case <-ctx.Done():
		executor.cancelTicket(dispatchKey, ticket)
		return nil, ctx.Err()
	}
}

func (executor *keyExecutor) cancelTicket(dispatchKey string, ticket *keyTicket) {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	if ticket.element == nil {
		return
	}
	queue := ticket.queue
	wasGranted := ticket.granted
	queue.requests.Remove(ticket.element)
	ticket.element = nil
	if wasGranted {
		executor.grantFront(queue)
	}
	if queue.requests.Len() == 0 {
		delete(executor.keyQueues, dispatchKey)
	}
}

func (executor *keyExecutor) releaseKey(dispatchKey string, ticket *keyTicket) {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	if ticket.element == nil {
		return
	}
	queue := ticket.queue
	queue.requests.Remove(ticket.element)
	ticket.element = nil
	executor.grantFront(queue)
	if queue.requests.Len() == 0 {
		delete(executor.keyQueues, dispatchKey)
	}
}

func (*keyExecutor) grantFront(queue *keyQueue) {
	front := queue.requests.Front()
	if front == nil {
		return
	}
	ticket := front.Value.(*keyTicket)
	if ticket.granted {
		return
	}
	ticket.granted = true
	close(ticket.ready)
}

func (executor *keyExecutor) keyQueueLength(key string) int {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	queue := executor.keyQueues[key]
	if queue == nil {
		return 0
	}
	return queue.requests.Len()
}

func (executor *keyExecutor) keyQueueCount() int {
	executor.mu.Lock()
	defer executor.mu.Unlock()
	return len(executor.keyQueues)
}
