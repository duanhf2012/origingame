package scenario

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestWorkloadLaunchesBoundedUsersAndCompletes(t *testing.T) {
	var mu sync.Mutex
	var ids []int64
	current, err := newWorkload(context.Background(), 3, 3*time.Millisecond, time.Millisecond, func(id int64) error {
		mu.Lock()
		ids = append(ids, id)
		mu.Unlock()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	current.start()
	if err = <-current.done; err != nil {
		t.Fatal(err)
	}
	current.stop()
	mu.Lock()
	defer mu.Unlock()
	if len(ids) != 3 || ids[0] != 1 || ids[2] != 3 {
		t.Fatalf("launched ids = %v", ids)
	}
}

func TestWorkloadStopCancelsRamp(t *testing.T) {
	current, err := newWorkload(context.Background(), 100, time.Hour, time.Hour, func(int64) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	current.start()
	current.stop()
	if err = <-current.done; err == nil {
		t.Fatal("stopped workload returned nil")
	}
}
