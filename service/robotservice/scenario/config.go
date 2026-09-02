package scenario

import (
	"fmt"
	"net/url"
	"strings"

	originconfig "github.com/duanhf2012/origin/v3/config"
)

// Config 保存 RobotService 场景执行、目标和全部资源硬上限。
type Config struct {
	Blueprint BlueprintConfig `json:"blueprint"` // 蓝图加载配置。
	Control   ControlConfig   `json:"control"`   // 运行控制配置。
	Target    TargetConfig    `json:"target"`    // 外部目标配置。
	Identity  IdentityConfig  `json:"identity"`  // 测试身份配置。
	Workload  WorkloadConfig  `json:"workload"`  // 压测负载配置。
	IO        IOConfig        `json:"io"`        // 阻塞 I/O 配置。
}

// BlueprintConfig 指定机器人节点、行为图和默认入口。
type BlueprintConfig struct {
	NodeDir    string `json:"node_dir"`    // 节点定义目录。
	GraphDir   string `json:"graph_dir"`   // 蓝图图定义目录。
	GraphName  string `json:"graph_name"`  // 默认蓝图图名称。
	EntranceID int64  `json:"entrance_id"` // 默认入口节点标识。
}

// ControlConfig 决定 Ready 前是否自动启动默认运行。
type ControlConfig struct {
	StartupRun bool `json:"startup_run"` // Ready 后是否自动启动。
}

// TargetConfig 保存真实客户端入口，不包含内部 Service 地址。
type TargetConfig struct {
	LoginURL   string `json:"login_url"`    // LoginService HTTP 地址。
	ShowAreaID int64  `json:"show_area_id"` // 目标显示区服标识。
}

// IdentityConfig 保存隔离测试身份生成规则。
type IdentityConfig struct {
	PlatformType     int32  `json:"platform_type"`      // 测试平台类型。
	PlatformIDPrefix string `json:"platform_id_prefix"` // 测试账号标识前缀。
}

// WorkloadConfig 保存单实例固定在线人数闭环计划。
type WorkloadConfig struct {
	Users              int64                 `json:"users"`                // 目标在线机器人数量。
	RampUp             originconfig.Duration `json:"ramp_up"`              // 渐进启动时长。
	Duration           originconfig.Duration `json:"duration"`             // 运行总时长。
	ScenarioRetryCount int                   `json:"scenario_retry_count"` // 单机器人场景重试次数。
	ReplaceFailed      bool                  `json:"replace_failed"`       // 是否补充失败机器人。
	MaxReplacements    int64                 `json:"max_replacements"`     // 最多替补机器人数量。
}

// IOConfig 保存阻塞 I/O Executor 的固定容量。
type IOConfig struct {
	Workers       int `json:"workers"`        // 固定 I/O Worker 数。
	QueueMessages int `json:"queue_messages"` // 有界 I/O 任务队列容量。
}

// ValidateConfig 在创建任何资源前验证全部硬边界。
func ValidateConfig(config Config) error {
	blueprint := config.Blueprint
	if strings.TrimSpace(blueprint.NodeDir) == "" || strings.TrimSpace(blueprint.GraphDir) == "" ||
		strings.TrimSpace(blueprint.GraphName) == "" || blueprint.EntranceID <= 0 {
		return fmt.Errorf("blueprint 目录、graph_name 和 entrance_id 必须有效")
	}
	if len(blueprint.GraphName) > 128 {
		return fmt.Errorf("blueprint.graph_name 不能超过128字节")
	}
	parsed, err := url.Parse(strings.TrimSpace(config.Target.LoginURL))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" ||
		parsed.User != nil || parsed.Fragment != "" {
		return fmt.Errorf("target.login_url 必须是无凭证和Fragment的HTTP(S)地址")
	}
	if config.Target.ShowAreaID <= 0 {
		return fmt.Errorf("target.show_area_id 必须为正数")
	}
	if config.Identity.PlatformType < 0 || config.Identity.PlatformType > 4 {
		return fmt.Errorf("identity.platform_type 必须在0到4之间")
	}
	prefix := strings.TrimSpace(config.Identity.PlatformIDPrefix)
	if prefix == "" || len(prefix) > 64 {
		return fmt.Errorf("identity.platform_id_prefix 必须为1到64字节")
	}
	workload := config.Workload
	if workload.Users <= 0 || workload.Users > 100000 {
		return fmt.Errorf("workload.users 必须在1到100000之间")
	}
	if workload.RampUp.Duration() <= 0 || workload.Duration.Duration() <= 0 {
		return fmt.Errorf("workload.ramp_up 和 duration 必须为正数")
	}
	if workload.ScenarioRetryCount < 0 || workload.ScenarioRetryCount > 2 {
		return fmt.Errorf("workload.scenario_retry_count 只能为0到2")
	}
	if workload.ReplaceFailed {
		return fmt.Errorf("当前阶段尚未实现固定CCU补位，replace_failed必须为false")
	}
	if !workload.ReplaceFailed && workload.MaxReplacements != 0 {
		return fmt.Errorf("关闭replace_failed时max_replacements必须为0")
	}
	if workload.MaxReplacements < 0 || workload.MaxReplacements > 100000 {
		return fmt.Errorf("workload.max_replacements 必须在0到100000之间")
	}
	if config.IO.Workers <= 0 || config.IO.Workers > 1024 {
		return fmt.Errorf("io.workers 必须在1到1024之间")
	}
	if config.IO.QueueMessages <= 0 || config.IO.QueueMessages > 1000000 {
		return fmt.Errorf("io.queue_messages 必须在1到1000000之间")
	}
	return nil
}
