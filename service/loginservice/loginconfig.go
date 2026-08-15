package loginservice

import "github.com/duanhf2012/origin/v3/sysmodule/ginmodule"

// defaultConfig 返回 LoginService 的配置基线；YAML 只覆盖当前需要调整的字段。
func defaultConfig() Config {
	return Config{HTTP: HTTPConfig{Server: ginmodule.DefaultServerConfig()}}
}
