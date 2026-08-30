// Package robotservice 装配机器人场景运行能力并提供内部控制 RPC。
package robotservice

import (
	"context"
	"fmt"
	"time"

	"github.com/duanhf2012/origin/v3/service"
	rpcapi "origingame/protocol/rpc"
	"origingame/service/robotservice/scenario"
)

const controlCompletionMargin = 2 * time.Minute

// RobotService 只持有顶层配置和场景 Module，协议连接与执行资源由 RobotScenarioModule 所有。
type RobotService struct {
	service.Service
	config    scenario.Config
	scenarios *scenario.RobotScenarioModule
}

var _ rpcapi.RobotService = (*RobotService)(nil)

// OnInit 严格加载全部配置，并在发布 Ready 前完成节点注册和蓝图编译准备。
func (target *RobotService) OnInit() error {
	if err := target.loadConfig(); err != nil {
		return err
	}
	if err := scenario.ValidateConfig(target.config); err != nil {
		return fmt.Errorf("校验 RobotService 配置: %w", err)
	}
	awaitTimeout := target.config.Workload.RampUp.Duration() +
		target.config.Workload.Duration.Duration() + controlCompletionMargin
	if err := target.SetDefaultAwaitTimeout(awaitTimeout); err != nil {
		return err
	}
	target.scenarios = scenario.NewRobotScenarioModule(target.config)
	return target.AddModule(target.scenarios)
}

func (target *RobotService) loadConfig() error {
	sections := []struct {
		path        string
		destination any
	}{
		{"blueprint", &target.config.Blueprint},
		{"control", &target.config.Control},
		{"target", &target.config.Target},
		{"identity", &target.config.Identity},
		{"workload", &target.config.Workload},
		{"io", &target.config.IO},
	}
	for _, section := range sections {
		if err := target.GetServiceConfigStrict(section.path, section.destination); err != nil {
			return fmt.Errorf("加载 RobotService.%s 配置: %w", section.path, err)
		}
	}
	return nil
}

func (target *RobotService) ListScenarios(_ context.Context, request rpcapi.ListRobotScenariosRequest) (rpcapi.ListRobotScenariosResponse, error) {
	return target.scenarios.ListScenarios(request)
}

func (target *RobotService) StartRun(_ context.Context, request rpcapi.StartRobotRunRequest) (rpcapi.RobotRunSnapshot, error) {
	return target.scenarios.StartRun(request)
}

func (target *RobotService) StopRun(_ context.Context, request rpcapi.StopRobotRunRequest) (rpcapi.RobotRunSnapshot, error) {
	return target.scenarios.StopRun(request)
}

func (target *RobotService) GetRun(_ context.Context, request rpcapi.GetRobotRunRequest) (rpcapi.RobotRunSnapshot, error) {
	return target.scenarios.GetRun(request)
}
