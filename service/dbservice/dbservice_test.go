package dbservice

import (
	"errors"
	"testing"

	"github.com/duanhf2012/origin/v3/errs"
	originmongo "github.com/duanhf2012/origin/v3/sysmodule/mongodbmodule"
	originredis "github.com/duanhf2012/origin/v3/sysmodule/redismodule"
)

func TestConfigDefaultsAndDataDomainValidation(t *testing.T) {
	defaults := defaultConfig()
	if defaults.MaxIOConcurrency != 64 || defaults.MaxInflightRequests != 128 {
		t.Fatalf("defaults=%+v", defaults)
	}

	validAcc := Config{
		MaxIOConcurrency:    64,
		MaxInflightRequests: 128,
		MongoDB:             originmongo.Config{Database: "origingame_account"},
		Redis:               originredis.Config{Database: 0, MaxRetries: 0},
	}
	if err := validateConfig("AccDBService", validAcc); err != nil {
		t.Fatalf("valid Acc config error=%v", err)
	}
	validRole := validAcc
	validRole.MongoDB.Database = "origingame_role"
	validRole.Redis.Database = 1
	if err := validateConfig("RoleDBService", validRole); err != nil {
		t.Fatalf("valid Role config error=%v", err)
	}

	invalid := []struct {
		name   string
		config Config
	}{
		{name: "unknown service", config: validAcc},
		{name: "AccDBService", config: func() Config { value := validAcc; value.Redis.Database = 1; return value }()},
		{name: "RoleDBService", config: func() Config { value := validRole; value.MongoDB.Database = "origingame_account"; return value }()},
		{name: "AccDBService", config: func() Config { value := validAcc; value.MaxIOConcurrency = 129; return value }()},
		{name: "AccDBService", config: func() Config { value := validAcc; value.Redis.MaxRetries = 1; return value }()},
	}
	for index, test := range invalid {
		if err := validateConfig(test.name, test.config); !errors.Is(err, errs.ErrInvalidConfig) {
			t.Fatalf("invalid[%d] error=%v", index, err)
		}
	}
}
