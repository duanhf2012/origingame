package scenario

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestIOExecutorBoundsQueueAndStopsWorkers(t *testing.T) {
	block := make(chan struct{})
	dispatched := make(chan struct{}, 2)
	executor, err := newIOExecutor(1, 1, func(task func(context.Context)) error {
		task(context.Background())
		dispatched <- struct{}{}
		return nil
	})
	if err != nil || executor.start(context.Background()) != nil {
		t.Fatal(err)
	}
	job := ioJob{
		ctx: context.Background(),
		work: func(context.Context) (any, error) {
			<-block
			return nil, nil
		},
		complete: func(any, error) {},
	}
	if err = executor.submit(job); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for len(executor.queue) != 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if err = executor.submit(job); err != nil {
		t.Fatal(err)
	}
	if err = executor.submit(job); !errors.Is(err, errIOExecutorFull) {
		t.Fatalf("third submit error = %v", err)
	}
	close(block)
	executor.stop()
}

func TestRealTimerSchedulerCancelAndStopAreIdempotent(t *testing.T) {
	var calls atomic.Int64
	scheduler, err := newRealTimerScheduler(func(task func(context.Context)) error {
		task(context.Background())
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	cancel, err := scheduler.schedule(time.Hour, func() { calls.Add(1) })
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	cancel()
	scheduler.stop()
	scheduler.stop()
	if calls.Load() != 0 {
		t.Fatalf("canceled callback calls = %d", calls.Load())
	}
}

func TestRealTimerSchedulerHandlesImmediateExpiry(t *testing.T) {
	const timerCount = 100
	done := make(chan struct{}, timerCount)
	scheduler, err := newRealTimerScheduler(func(task func(context.Context)) error {
		task(context.Background())
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer scheduler.stop()
	for range timerCount {
		if _, err = scheduler.schedule(time.Nanosecond, func() { done <- struct{}{} }); err != nil {
			t.Fatal(err)
		}
	}
	deadline := time.After(time.Second)
	for range timerCount {
		select {
		case <-done:
		case <-deadline:
			t.Fatal("极短延迟Timer未全部完成")
		}
	}
}
