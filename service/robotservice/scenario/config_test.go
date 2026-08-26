package scenario

import (
	"testing"
	"time"

	originconfig "github.com/duanhf2012/origin/v3/config"
)

func validConfig() Config {
	return Config{
		Blueprint: BlueprintConfig{NodeDir: "nodes", GraphDir: "blueprints", GraphName: "login", EntranceID: 1001},
		Target:    TargetConfig{LoginURL: "http://127.0.0.1:8080/api/v1/login", ShowAreaID: 1},
		Identity:  IdentityConfig{PlatformType: 1, PlatformIDPrefix: "robot-"},
		Workload: WorkloadConfig{
			Users: 10, RampUp: originconfig.Duration(time.Second), Duration: originconfig.Duration(time.Minute),
		},
		IO: IOConfig{Workers: 2, QueueMessages: 8},
	}
}

func TestValidateConfigAcceptsBoundedPlan(t *testing.T) {
	if err := ValidateConfig(validConfig()); err != nil {
		t.Fatal(err)
	}
}

func TestValidateConfigRejectsUnboundedReplacementAndInvalidTarget(t *testing.T) {
	config := validConfig()
	config.Workload.ReplaceFailed = true
	if err := ValidateConfig(config); err == nil {
		t.Fatal("replace_failed without max_replacements was accepted")
	}
	config = validConfig()
	config.Target.LoginURL = "http://user:password@127.0.0.1/login"
	if err := ValidateConfig(config); err == nil {
		t.Fatal("credential-bearing login_url was accepted")
	}
}
