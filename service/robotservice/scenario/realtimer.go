package scenario

import (
	"context"
	"errors"
	"sync"
	"time"
)

type scheduledTimer struct {
	owner *realTimerScheduler
	once  sync.Once
	timer *time.Timer
}

// realTimerScheduler 统一拥有机器人基础设施真实时间等待。
type realTimerScheduler struct {
	dispatch func(func(context.Context)) error

	mu      sync.Mutex
	stopped bool
	timers  map[*scheduledTimer]struct{}
	wg      sync.WaitGroup
}

func newRealTimerScheduler(dispatch func(func(context.Context)) error) (*realTimerScheduler, error) {
	if dispatch == nil {
		return nil, errors.New("RobotService Timer Dispatcher不能为空")
	}
	return &realTimerScheduler{dispatch: dispatch, timers: make(map[*scheduledTimer]struct{})}, nil
}

func (scheduler *realTimerScheduler) schedule(delay time.Duration, callback func()) (func(), error) {
	if scheduler == nil || delay <= 0 || callback == nil {
		return nil, errors.New("RobotService Timer参数无效")
	}
	scheduler.mu.Lock()
	if scheduler.stopped {
		scheduler.mu.Unlock()
		return nil, errors.New("RobotService TimerScheduler已停止")
	}
	handle := &scheduledTimer{owner: scheduler}
	scheduler.timers[handle] = struct{}{}
	scheduler.wg.Add(1)
	handle.timer = time.AfterFunc(delay, func() {
		handle.finish(func() { _ = scheduler.dispatch(func(context.Context) { callback() }) })
	})
	scheduler.mu.Unlock()
	return func() { handle.cancel() }, nil
}

func (handle *scheduledTimer) cancel() {
	if handle == nil {
		return
	}
	handle.finish(nil)
}

func (handle *scheduledTimer) finish(callback func()) {
	handle.once.Do(func() {
		// timer 在schedule持有owner.mu时赋值；到期回调可能在AfterFunc返回的同时启动。
		// 在同一把锁内读取和移除，避免极短延迟下与handle.timer赋值竞态。
		handle.owner.mu.Lock()
		if handle.timer != nil {
			handle.timer.Stop()
		}
		delete(handle.owner.timers, handle)
		handle.owner.mu.Unlock()
		if callback != nil {
			callback()
		}
		handle.owner.wg.Done()
	})
}

func (scheduler *realTimerScheduler) stop() {
	if scheduler == nil {
		return
	}
	scheduler.mu.Lock()
	if scheduler.stopped {
		scheduler.mu.Unlock()
		return
	}
	scheduler.stopped = true
	timers := make([]*scheduledTimer, 0, len(scheduler.timers))
	for timer := range scheduler.timers {
		timers = append(timers, timer)
	}
	scheduler.mu.Unlock()
	for _, timer := range timers {
		timer.cancel()
	}
	scheduler.wg.Wait()
}
