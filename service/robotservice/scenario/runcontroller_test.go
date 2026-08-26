package scenario

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	rpcapi "origingame/protocol/rpc"
)

func newTestRunController() *runController {
	controller := newRunController()
	current := time.Unix(100, 0)
	controller.now = func() time.Time {
		current = current.Add(time.Second)
		return current
	}
	sequence := 0
	controller.newID = func() (string, error) {
		sequence++
		return fmt.Sprintf("run-%d", sequence), nil
	}
	return controller
}

func TestRunControllerIdempotencyAndSingleActiveRun(t *testing.T) {
	controller := newTestRunController()
	first, record, existing, err := controller.prepareStart(context.Background(), "request-1", "login", 10)
	if err != nil || existing || first.State != rpcapi.RobotRunStateStarting {
		t.Fatalf("prepareStart() = %+v, existing=%t, err=%v", first, existing, err)
	}
	duplicate, duplicateRecord, existing, err := controller.prepareStart(context.Background(), "request-1", "login", 10)
	if err != nil || !existing || duplicateRecord != nil || duplicate.RunID != first.RunID {
		t.Fatalf("duplicate = %+v, record=%v, existing=%t, err=%v", duplicate, duplicateRecord, existing, err)
	}
	if _, _, _, err = controller.prepareStart(context.Background(), "request-1", "other", 10); !errors.Is(err, errRequestConflict) {
		t.Fatalf("conflicting request_id error = %v", err)
	}
	if _, _, _, err = controller.prepareStart(context.Background(), "request-2", "login", 10); !errors.Is(err, errRunActive) {
		t.Fatalf("parallel start error = %v", err)
	}
	controller.markRunning(record)
	controller.robotStarted(record)
	controller.robotFinished(record, robotOutcomeSucceeded)
	finished := controller.finish(record, rpcapi.RobotRunStateSucceeded, "", "")
	if finished.SucceededRobots != 1 || finished.ActiveRobots != 0 || finished.FinishedAtMS == 0 {
		t.Fatalf("finished snapshot = %+v", finished)
	}
}

func TestRunControllerStopIsIdempotent(t *testing.T) {
	controller := newTestRunController()
	started, record, _, err := controller.prepareStart(context.Background(), "request", "login", 1)
	if err != nil {
		t.Fatal(err)
	}
	controller.markRunning(record)
	first, stoppingRecord, err := controller.requestStop(started.RunID)
	if err != nil || stoppingRecord != record || first.State != rpcapi.RobotRunStateStopping || record.ctx.Err() == nil {
		t.Fatalf("first stop = %+v, record=%v, err=%v", first, stoppingRecord, err)
	}
	second, stoppingRecord, err := controller.requestStop(started.RunID)
	if err != nil || stoppingRecord != record || second.State != rpcapi.RobotRunStateStopping {
		t.Fatalf("second stop = %+v, record=%v, err=%v", second, stoppingRecord, err)
	}
	controller.finish(record, rpcapi.RobotRunStateCanceled, "", "")
	terminal, stoppingRecord, err := controller.requestStop(started.RunID)
	if err != nil || stoppingRecord != nil || terminal.State != rpcapi.RobotRunStateCanceled {
		t.Fatalf("terminal stop = %+v, record=%v, err=%v", terminal, stoppingRecord, err)
	}
}

func TestRunControllerKeepsOnlyTwentyRecentSnapshots(t *testing.T) {
	controller := newTestRunController()
	for index := 0; index < recentRunLimit+2; index++ {
		requestID := fmt.Sprintf("request-%d", index)
		_, record, _, err := controller.prepareStart(context.Background(), requestID, "login", 1)
		if err != nil {
			t.Fatal(err)
		}
		controller.finish(record, rpcapi.RobotRunStateSucceeded, "", "")
	}
	if len(controller.recent) != recentRunLimit {
		t.Fatalf("recent count = %d", len(controller.recent))
	}
	if _, found := controller.findByRequestID("request-0"); found {
		t.Fatal("oldest request_id was not evicted")
	}
	if latest, err := controller.get(""); err != nil || latest.RequestID != fmt.Sprintf("request-%d", recentRunLimit+1) {
		t.Fatalf("latest = %+v, err=%v", latest, err)
	}
}
