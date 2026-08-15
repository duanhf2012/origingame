package loginservice

import (
	"path/filepath"
	"reflect"
	"testing"

	originconfig "github.com/duanhf2012/origin/v3/config"
	"github.com/duanhf2012/origin/v3/sysmodule/ginmodule"
)

func TestDefaultConfigUsesGinServerDefaults(t *testing.T) {
	if got, want := defaultConfig().HTTP.Server, ginmodule.DefaultServerConfig(); !reflect.DeepEqual(got, want) {
		t.Fatalf("HTTP server default = %#v, want %#v", got, want)
	}
}

func TestLocalConfigFragmentsMergeWithServiceDefaults(t *testing.T) {
	snapshot, err := originconfig.LoadSnapshot(filepath.Join("..", "..", "config"))
	if err != nil {
		t.Fatalf("LoadSnapshot() error = %v", err)
	}

	configured := defaultConfig()
	service, err := snapshot.Root().Lookup("services.LoginService")
	if err != nil {
		t.Fatalf("LoginService 配置不存在: %v", err)
	}
	server, err := service.Lookup("http.server")
	if err != nil {
		t.Fatalf("HTTP Server 配置不存在: %v", err)
	}
	if err := server.DecodeStrict(&configured.HTTP.Server); err != nil {
		t.Fatalf("HTTP Server 严格解码失败: %v", err)
	}
	if configured.HTTP.Server.Address != "0.0.0.0:8080" {
		t.Fatalf("HTTP address = %q, want 0.0.0.0:8080", configured.HTTP.Server.Address)
	}
	if got, want := configured.HTTP.Server.RequestTimeout, ginmodule.DefaultServerConfig().RequestTimeout; got != want {
		t.Fatalf("省略的 request_timeout = %v, want framework default %v", got, want)
	}

	var gatewayToken struct {
		Issuer     string
		Audience   string
		PublicKeys map[string]string
	}
	gateway, err := snapshot.Root().Lookup("services.GatewayService.token")
	if err != nil {
		t.Fatalf("GatewayService 公钥配置不存在: %v", err)
	}
	if err := gateway.DecodeStrict(&gatewayToken); err != nil {
		t.Fatalf("GatewayService 公钥严格解码失败: %v", err)
	}
	if gatewayToken.PublicKeys["dev-key-1"] == "" {
		t.Fatal("GatewayService 缺少 dev-key-1 公钥")
	}
}
