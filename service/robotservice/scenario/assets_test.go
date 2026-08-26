package scenario

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/duanhf2012/origin/v3/sysmodule/blueprintmodule"
)

// TestRobotBlueprintAssetsCompile 保证提交的编辑器端口快照与 Go 节点契约没有漂移。
func TestRobotBlueprintAssetsCompile(t *testing.T) {
	root, err := robotAssetRoot()
	if err != nil {
		t.Fatal(err)
	}
	module, err := blueprintmodule.New(blueprintmodule.Config{
		NodeDir: filepath.Join(root, "nodes"), GraphDir: filepath.Join(root, "blueprints"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = module.RegisterNodes(
		func() blueprintmodule.IExecNode { return &robotStartNode{} },
		func() blueprintmodule.IExecNode { return &httpLoginNode{} },
		func() blueprintmodule.IExecNode { return &connectGatewayNode{} },
		func() blueprintmodule.IExecNode { return &loginPlayerNode{} },
		func() blueprintmodule.IExecNode { return &startHeartbeatNode{} },
		func() blueprintmodule.IExecNode { return &heartbeatNode{} },
		func() blueprintmodule.IExecNode { return &waitNode{} },
		func() blueprintmodule.IExecNode { return &waitMessageNode{} },
		func() blueprintmodule.IExecNode { return &disconnectNode{} },
	); err != nil {
		t.Fatal(err)
	}
	if err = module.OnInit(); err != nil {
		t.Fatal(err)
	}
	if err = module.OnStart(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err = module.OnStop(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestBusinessLoopCarriesEditorFallbackPorts(t *testing.T) {
	root, err := robotAssetRoot()
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(root, "blueprints", "login_heartbeat.obp"))
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Nodes []struct {
			ID         string `json:"id"`
			Properties struct {
				LegacyClass  string `json:"legacyClass"`
				LegacyInputs []struct {
					Key string `json:"key"`
				} `json:"legacyInputs"`
				LegacyOutputs []struct {
					Key string `json:"key"`
				} `json:"legacyOutputs"`
			} `json:"properties"`
		} `json:"nodes"`
	}
	if err = json.Unmarshal(content, &document); err != nil {
		t.Fatal(err)
	}
	for _, node := range document.Nodes {
		if node.ID != "business_loop" {
			continue
		}
		if node.Properties.LegacyClass != "WhileNode" || len(node.Properties.LegacyInputs) != 2 ||
			len(node.Properties.LegacyOutputs) != 2 {
			t.Fatalf("business_loop 编辑器回退端口不完整: %+v", node.Properties)
		}
		return
	}
	t.Fatal("login_heartbeat.obp 缺少 business_loop")
}

func TestBackgroundHeartbeatStartsBeforeBusinessLoop(t *testing.T) {
	root, err := robotAssetRoot()
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(root, "blueprints", "login_heartbeat.obp"))
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Connections []struct {
			Source       string `json:"source"`
			SourceOutput string `json:"sourceOutput"`
			Target       string `json:"target"`
			TargetInput  string `json:"targetInput"`
		} `json:"connections"`
	}
	if err = json.Unmarshal(content, &document); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"login_player:out0->start_heartbeat:in0":   false,
		"start_heartbeat:out0->business_loop:exec": false,
	}
	for _, connection := range document.Connections {
		key := connection.Source + ":" + connection.SourceOutput + "->" + connection.Target + ":" + connection.TargetInput
		if _, exists := want[key]; exists {
			want[key] = true
		}
	}
	for connection, found := range want {
		if !found {
			t.Fatalf("后台心跳蓝图连接缺失: %s", connection)
		}
	}
}

func robotAssetRoot() (string, error) {
	return filepath.Abs(filepath.Join("..", "..", "..", "tests", "e2e", "robot"))
}
