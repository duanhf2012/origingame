package gatewayservice

import (
	"path/filepath"
	"testing"

	originconfig "github.com/duanhf2012/origin/v3/config"
	"origingame/service/gatewayservice/client"
)

func TestLocalGatewayConfigStrictlyOverlaysNetworkDefaults(t *testing.T) {
	snapshot, err := originconfig.LoadSnapshot(filepath.Join("..", "..", "config"))
	if err != nil {
		t.Fatal(err)
	}
	serviceConfig, err := snapshot.Root().Lookup("services.GatewayService")
	if err != nil {
		t.Fatal(err)
	}
	configured := Config{Client: client.DefaultConfig()}
	sections := []struct {
		path string // 配置路径。
		to   any    // 解码目标。
	}{
		{"token", &configured.Token},
		{"area", &configured.Area},
		{"tcp", &configured.Client.TCP},
		{"kcp", &configured.Client.KCP},
		{"websocket", &configured.Client.WebSocket},
	}
	for _, section := range sections {
		value, lookupErr := serviceConfig.Lookup(section.path)
		if lookupErr != nil {
			t.Fatalf("Lookup(%s): %v", section.path, lookupErr)
		}
		if decodeErr := value.DecodeStrict(section.to); decodeErr != nil {
			t.Fatalf("DecodeStrict(%s): %v", section.path, decodeErr)
		}
	}
	if !configured.Client.TCP.Enabled || !configured.Client.KCP.Enabled || !configured.Client.WebSocket.Enabled {
		t.Fatal("本地样例必须启用三种协议")
	}
	if configured.Client.TCP.Server.MaxMessageSize.Bytes() != 4*1024 ||
		configured.Client.KCP.Server.MaxMessageSize.Bytes() != 4*1024 ||
		configured.Client.WebSocket.Server.MaxMessageSize.Bytes() != 4*1024 {
		t.Fatal("三种协议必须统一限制为 4KB")
	}
	if configured.Client.TCP.Server.SendQueueMessages <= 0 {
		t.Fatal("未显式配置的网络容量应保留 Origin 有界默认值")
	}
}
