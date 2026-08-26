// Package scenario 实现 RobotService 蓝图场景、运行控制和机器人生命周期。
package scenario

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	rpcapi "origingame/protocol/rpc"
)

const recentRunLimit = 20

var (
	errRunActive       = errors.New("RobotService 已有活动运行")
	errRunNotFound     = errors.New("RobotService 运行不存在")
	errRequestConflict = errors.New("RobotService request_id 已用于其他场景")
)

type runRecord struct {
	snapshot rpcapi.RobotRunSnapshot
	ctx      context.Context
	cancel   context.CancelFunc
}

type runController struct {
	now    func() time.Time
	newID  func() (string, error)
	active *runRecord
	recent []rpcapi.RobotRunSnapshot
}

func newRunController() *runController {
	return &runController{now: time.Now, newID: newRunID}
}

func newRunID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("生成 Robot run_id: %w", err)
	}
	return hex.EncodeToString(value[:]), nil
}

// prepareStart 在调用真实 Workload 前发布 starting，保证幂等请求立即观察同一运行。
func (controller *runController) prepareStart(
	parent context.Context,
	requestID string,
	scenarioName string,
	targetUsers int64,
) (rpcapi.RobotRunSnapshot, *runRecord, bool, error) {
	requestID = strings.TrimSpace(requestID)
	scenarioName = strings.TrimSpace(scenarioName)
	if parent == nil || requestID == "" || scenarioName == "" || targetUsers <= 0 {
		return rpcapi.RobotRunSnapshot{}, nil, false, errors.New("RobotService 启动参数无效")
	}
	if len(requestID) > 128 || len(scenarioName) > 128 {
		return rpcapi.RobotRunSnapshot{}, nil, false, errors.New("RobotService 启动参数过长")
	}
	if existing, found := controller.findByRequestID(requestID); found {
		if existing.ScenarioName != scenarioName {
			return rpcapi.RobotRunSnapshot{}, nil, false, errRequestConflict
		}
		return existing, nil, true, nil
	}
	if controller.active != nil {
		return rpcapi.RobotRunSnapshot{}, nil, false, errRunActive
	}
	runID, err := controller.newID()
	if err != nil {
		return rpcapi.RobotRunSnapshot{}, nil, false, err
	}
	ctx, cancel := context.WithCancel(parent)
	record := &runRecord{
		ctx: ctx, cancel: cancel,
		snapshot: rpcapi.RobotRunSnapshot{
			RunID: runID, RequestID: requestID, ScenarioName: scenarioName,
			State: rpcapi.RobotRunStateStarting, StartedAtMS: controller.now().UnixMilli(),
			TargetUsers: targetUsers,
		},
	}
	controller.active = record
	return record.snapshot, record, false, nil
}

func (controller *runController) markRunning(record *runRecord) rpcapi.RobotRunSnapshot {
	if controller.active == record && record.snapshot.State == rpcapi.RobotRunStateStarting {
		record.snapshot.State = rpcapi.RobotRunStateRunning
	}
	return record.snapshot
}

func (controller *runController) robotStarted(record *runRecord) {
	if controller.active != record || terminalRunState(record.snapshot.State) {
		return
	}
	record.snapshot.StartedRobots++
	record.snapshot.ActiveRobots++
}

type robotOutcome uint8

const (
	robotOutcomeSucceeded robotOutcome = iota + 1
	robotOutcomeFailed
	robotOutcomeCanceled
	robotOutcomeFlaky
)

func (controller *runController) robotFinished(record *runRecord, outcome robotOutcome) {
	if controller.active != record || terminalRunState(record.snapshot.State) {
		return
	}
	if record.snapshot.ActiveRobots > 0 {
		record.snapshot.ActiveRobots--
	}
	switch outcome {
	case robotOutcomeSucceeded:
		record.snapshot.SucceededRobots++
	case robotOutcomeFailed:
		record.snapshot.FailedRobots++
	case robotOutcomeFlaky:
		record.snapshot.FlakyRobots++
	}
}

func (controller *runController) requestStop(runID string) (rpcapi.RobotRunSnapshot, *runRecord, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return rpcapi.RobotRunSnapshot{}, nil, errors.New("run_id 不能为空")
	}
	if controller.active != nil && controller.active.snapshot.RunID == runID {
		record := controller.active
		if record.snapshot.State != rpcapi.RobotRunStateStopping {
			record.snapshot.State = rpcapi.RobotRunStateStopping
			record.cancel()
		}
		return record.snapshot, record, nil
	}
	if snapshot, found := controller.findByRunID(runID); found {
		return snapshot, nil, nil
	}
	return rpcapi.RobotRunSnapshot{}, nil, errRunNotFound
}

func (controller *runController) finish(
	record *runRecord,
	state rpcapi.RobotRunState,
	failureCode string,
	failureMessage string,
) rpcapi.RobotRunSnapshot {
	if controller.active != record || !terminalRunState(state) {
		return rpcapi.RobotRunSnapshot{}
	}
	record.cancel()
	record.snapshot.State = state
	record.snapshot.FinishedAtMS = controller.now().UnixMilli()
	record.snapshot.ActiveRobots = 0
	record.snapshot.FailureCode = failureCode
	record.snapshot.FailureMessage = failureMessage
	snapshot := record.snapshot
	controller.active = nil
	controller.recent = append(controller.recent, snapshot)
	if len(controller.recent) > recentRunLimit {
		copy(controller.recent, controller.recent[len(controller.recent)-recentRunLimit:])
		controller.recent = controller.recent[:recentRunLimit]
	}
	return snapshot
}

func (controller *runController) get(runID string) (rpcapi.RobotRunSnapshot, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		if controller.active != nil {
			return controller.active.snapshot, nil
		}
		if len(controller.recent) > 0 {
			return controller.recent[len(controller.recent)-1], nil
		}
		return rpcapi.RobotRunSnapshot{}, errRunNotFound
	}
	if snapshot, found := controller.findByRunID(runID); found {
		return snapshot, nil
	}
	return rpcapi.RobotRunSnapshot{}, errRunNotFound
}

func (controller *runController) findByRequestID(requestID string) (rpcapi.RobotRunSnapshot, bool) {
	if controller.active != nil && controller.active.snapshot.RequestID == requestID {
		return controller.active.snapshot, true
	}
	for index := len(controller.recent) - 1; index >= 0; index-- {
		if controller.recent[index].RequestID == requestID {
			return controller.recent[index], true
		}
	}
	return rpcapi.RobotRunSnapshot{}, false
}

func (controller *runController) findByRunID(runID string) (rpcapi.RobotRunSnapshot, bool) {
	if controller.active != nil && controller.active.snapshot.RunID == runID {
		return controller.active.snapshot, true
	}
	for index := len(controller.recent) - 1; index >= 0; index-- {
		if controller.recent[index].RunID == runID {
			return controller.recent[index], true
		}
	}
	return rpcapi.RobotRunSnapshot{}, false
}

func terminalRunState(state rpcapi.RobotRunState) bool {
	return state == rpcapi.RobotRunStateSucceeded || state == rpcapi.RobotRunStateFailed ||
		state == rpcapi.RobotRunStateCanceled
}
