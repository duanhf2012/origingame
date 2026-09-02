package dbservice

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/duanhf2012/origin/v3/errs"
)

func TestKeyExecutorRejectsInvalidCapacity(t *testing.T) {
	tests := []struct {
		io       int64 // I/O 并发上限。
		inflight int64 // 在途请求上限。
	}{
		{io: 0, inflight: 1},
		{io: 1, inflight: 0},
		{io: 2, inflight: 1},
	}
	for _, test := range tests {
		if _, err := newKeyExecutor(test.io, test.inflight); !errors.Is(err, errs.ErrInvalidArgument) {
			t.Fatalf("newKeyExecutor(%d, %d) error=%v", test.io, test.inflight, err)
		}
	}
}

func TestKeyExecutorInflightReservationIsNonBlockingAndBounded(t *testing.T) {
	executor, err := newKeyExecutor(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	first, ok := executor.tryReserve()
	if !ok {
		t.Fatal("首次预留失败")
	}
	second, ok := executor.tryReserve()
	if !ok {
		t.Fatal("第二次预留失败")
	}
	if _, ok := executor.tryReserve(); ok {
		t.Fatal("超过 inflight 上限仍然预留成功")
	}
	first.release()
	reused, ok := executor.tryReserve()
	if !ok {
		t.Fatal("释放后未能重新预留")
	}
	first.release()
	second.release()
	reused.release()
	if got := executor.inflight.Load(); got != 0 {
		t.Fatalf("inflight=%d, want 0", got)
	}
}

func TestKeyExecutorRunsSameKeyInAdmissionOrder(t *testing.T) {
	executor := mustNewKeyExecutor(t, 2, 8)
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	completed := make(chan int, 3)

	start := func(id int, block bool) <-chan error {
		done := make(chan error, 1)
		go func() {
			reservation, ok := executor.tryReserve()
			if !ok {
				done <- errs.ErrServiceQueueFull
				return
			}
			defer reservation.release()
			done <- executor.execute(context.Background(), "player-1", func(context.Context) error {
				if block {
					close(firstStarted)
					<-releaseFirst
				}
				completed <- id
				return nil
			})
		}()
		return done
	}

	firstDone := start(1, true)
	<-firstStarted
	secondDone := start(2, false)
	waitKeyQueueLength(t, executor, "player-1", 2)
	thirdDone := start(3, false)
	waitKeyQueueLength(t, executor, "player-1", 3)
	close(releaseFirst)

	for _, done := range []<-chan error{firstDone, secondDone, thirdDone} {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
	for want := 1; want <= 3; want++ {
		if got := <-completed; got != want {
			t.Fatalf("完成顺序=%d, want %d", got, want)
		}
	}
	if count := executor.keyQueueCount(); count != 0 {
		t.Fatalf("空队列未清理: %d", count)
	}
}

func TestKeyExecutorLimitsDifferentKeyIOConcurrency(t *testing.T) {
	executor := mustNewKeyExecutor(t, 2, 8)
	var active atomic.Int64
	var maximum atomic.Int64
	started := make(chan struct{}, 4)
	release := make(chan struct{})
	var wait sync.WaitGroup

	for index := 0; index < 4; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			reservation, ok := executor.tryReserve()
			if !ok {
				t.Errorf("请求 %d 预留失败", index)
				return
			}
			defer reservation.release()
			err := executor.execute(context.Background(), string(rune('a'+index)), func(context.Context) error {
				current := active.Add(1)
				for {
					old := maximum.Load()
					if current <= old || maximum.CompareAndSwap(old, current) {
						break
					}
				}
				started <- struct{}{}
				<-release
				active.Add(-1)
				return nil
			})
			if err != nil {
				t.Errorf("请求 %d 执行失败: %v", index, err)
			}
		}(index)
	}

	for index := 0; index < 2; index++ {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("I/O 槽位没有并发运行")
		}
	}
	select {
	case <-started:
		t.Fatal("实际 I/O 超过配置上限")
	case <-time.After(30 * time.Millisecond):
	}
	close(release)
	wait.Wait()
	if got := maximum.Load(); got != 2 {
		t.Fatalf("最大并发=%d, want 2", got)
	}
	if got := executor.running.Load(); got != 0 {
		t.Fatalf("running=%d, want 0", got)
	}
}

func TestKeyExecutorCancellationRemovesQueuedTicket(t *testing.T) {
	executor := mustNewKeyExecutor(t, 1, 4)
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	firstDone := make(chan error, 1)
	go func() {
		firstDone <- executor.execute(context.Background(), "same", func(context.Context) error {
			close(firstStarted)
			<-releaseFirst
			return nil
		})
	}()
	<-firstStarted

	ctx, cancel := context.WithCancel(context.Background())
	secondDone := make(chan error, 1)
	go func() {
		secondDone <- executor.execute(ctx, "same", func(context.Context) error {
			t.Error("已取消请求不应进入 I/O")
			return nil
		})
	}()
	waitKeyQueueLength(t, executor, "same", 2)
	cancel()
	if err := <-secondDone; !errors.Is(err, context.Canceled) {
		t.Fatalf("取消错误=%v", err)
	}
	waitKeyQueueLength(t, executor, "same", 1)
	close(releaseFirst)
	if err := <-firstDone; err != nil {
		t.Fatal(err)
	}
	if count := executor.keyQueueCount(); count != 0 {
		t.Fatalf("取消后队列泄漏: %d", count)
	}
}

func TestKeyExecutorPanicStillReleasesKeyAndSlot(t *testing.T) {
	executor := mustNewKeyExecutor(t, 1, 2)
	panicked := make(chan struct{})
	go func() {
		defer func() {
			if recover() != nil {
				close(panicked)
			}
		}()
		_ = executor.execute(context.Background(), "same", func(context.Context) error {
			panic("boom")
		})
	}()
	select {
	case <-panicked:
	case <-time.After(time.Second):
		t.Fatal("panic 未返回调用边界")
	}

	if err := executor.execute(context.Background(), "same", func(context.Context) error { return nil }); err != nil {
		t.Fatalf("panic 后同 Key 无法继续执行: %v", err)
	}
	if executor.running.Load() != 0 || executor.keyQueueCount() != 0 {
		t.Fatalf("panic 后资源未释放: running=%d queues=%d", executor.running.Load(), executor.keyQueueCount())
	}
}

func TestKeyExecutorRejectsHotKeyOverflow(t *testing.T) {
	executor := mustNewKeyExecutor(t, 1, 64)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	releaseFirst := make(chan struct{})
	firstStarted := make(chan struct{})
	var wait sync.WaitGroup
	for index := 0; index < maxPendingRequestsPerKey; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			_ = executor.execute(ctx, "hot", func(context.Context) error {
				if index == 0 {
					close(firstStarted)
					<-releaseFirst
				}
				return nil
			})
		}(index)
		if index == 0 {
			<-firstStarted
		}
		waitKeyQueueLength(t, executor, "hot", index+1)
	}
	if err := executor.execute(ctx, "hot", func(context.Context) error { return nil }); !errors.Is(err, errs.ErrServiceQueueFull) {
		t.Fatalf("热点 Key 溢出错误=%v", err)
	}
	cancel()
	close(releaseFirst)
	wait.Wait()
}

func mustNewKeyExecutor(t *testing.T, maxIO, maxInflight int64) *keyExecutor {
	t.Helper()
	executor, err := newKeyExecutor(maxIO, maxInflight)
	if err != nil {
		t.Fatal(err)
	}
	return executor
}

func waitKeyQueueLength(t *testing.T, executor *keyExecutor, key string, want int) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if executor.keyQueueLength(key) == want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("key=%q queue length=%d, want %d", key, executor.keyQueueLength(key), want)
}

func BenchmarkKeyExecutorSameKey(b *testing.B) {
	executor, err := newKeyExecutor(1, 1)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	operation := func(context.Context) error { return nil }
	b.ReportAllocs()
	for b.Loop() {
		reservation, ok := executor.tryReserve()
		if !ok {
			b.Fatal("预留 inflight 失败")
		}
		if err = executor.execute(ctx, "benchmark-player", operation); err != nil {
			b.Fatal(err)
		}
		reservation.release()
	}
}
