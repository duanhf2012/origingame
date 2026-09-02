package e2e

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	originconfig "github.com/duanhf2012/origin/v3/config"
	originlog "github.com/duanhf2012/origin/v3/log"
	"github.com/duanhf2012/origin/v3/node"
	"google.golang.org/protobuf/proto"
	commonpb "origingame/protocol/common"
	rpcapi "origingame/protocol/rpc"
	"origingame/service/robotservice"
)

const (
	robotBlueprintTestTimeout = 15 * time.Second
	robotFrameLimit           = 4 * 1024
)

// TestRobotBlueprintLoginHeartbeatFlow 使用真实Node调度器、HTTP/TCP和提交的蓝图验证机器人闭环。
// 外部Login/Gateway由确定性测试端替代，因此不需要MongoDB、Redis、NATS或etcd。
func TestRobotBlueprintLoginHeartbeatFlow(t *testing.T) {
	if os.Getenv("ORIGINGAME_ROBOT_E2E") != "1" {
		t.Skip("设置 ORIGINGAME_ROBOT_E2E=1 后执行完整机器人蓝图流程测试")
	}

	gateway := newRobotGatewayFixture(t)
	t.Cleanup(func() {
		if err := gateway.close(); err != nil {
			t.Errorf("关闭Gateway测试端: %v", err)
		}
	})
	var httpLogins atomic.Int64
	login := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		var credential struct {
			PlatformType int32  `json:"PlatType"` // 平台类型。
			PlatformID   string `json:"PlatId"`   // 平台用户标识。
		}
		if request.Method != http.MethodPost || json.NewDecoder(request.Body).Decode(&credential) != nil ||
			credential.PlatformType != 1 || credential.PlatformID != "robot-blueprint-1" {
			http.Error(writer, "invalid robot login", http.StatusBadRequest)
			return
		}
		httpLogins.Add(1)
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"ECode": 0,
			"Token": "robot-blueprint-token",
			"AreaList": []any{map[string]any{
				"ShowAreaId": 1,
				"GateList": []any{map[string]any{
					"Protocol": "tcp", "Address": gateway.address(),
				}},
			}},
		})
	}))
	t.Cleanup(login.Close)

	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	snapshot := robotE2EConfig(t, root, login.URL)
	target := &robotservice.RobotService{}
	current, err := node.New(
		node.Config{ID: "robot-e2e-1", Private: true, Services: []string{"RobotService"}},
		[]node.ServiceBinding{{Name: "RobotService", Template: "RobotService", Private: true, Service: target}},
		originlog.NewNop(),
		node.Options{Config: snapshot, MaxTimersPerNode: 1024, TimerLocation: time.UTC},
	)
	if err != nil {
		t.Fatal(err)
	}
	stopped := false
	t.Cleanup(func() {
		if stopped {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if current.State() == node.StateReady || current.State() == node.StateFailed {
			_ = current.Stop(ctx)
		} else {
			_ = current.Rollback(ctx)
		}
	})

	startCtx, cancelStart := context.WithTimeout(context.Background(), 5*time.Second)
	err = current.Start(startCtx)
	cancelStart()
	if err != nil {
		t.Fatalf("启动RobotService测试Node: %v", err)
	}

	deadline := time.Now().Add(robotBlueprintTestTimeout)
	var result rpcapi.RobotRunSnapshot
	for time.Now().Before(deadline) {
		result, err = queryRobotRun(target)
		if err == nil && (result.State == rpcapi.RobotRunStateSucceeded ||
			result.State == rpcapi.RobotRunStateFailed || result.State == rpcapi.RobotRunStateCanceled) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("查询机器人运行: %v", err)
	}
	if result.State != rpcapi.RobotRunStateSucceeded || result.StartedRobots != 1 ||
		result.ActiveRobots != 0 || result.FailedRobots != 0 || result.FailureCode != "" {
		t.Fatalf("机器人运行结果异常: %+v", result)
	}
	if httpLogins.Load() != 1 || gateway.loginPlayers.Load() != 1 || gateway.heartbeats.Load() < 1 {
		t.Fatalf(
			"蓝图阶段计数异常: http_login=%d login_player=%d heartbeat=%d",
			httpLogins.Load(), gateway.loginPlayers.Load(), gateway.heartbeats.Load(),
		)
	}

	stopCtx, cancelStop := context.WithTimeout(context.Background(), 5*time.Second)
	err = current.Stop(stopCtx)
	cancelStop()
	if err != nil {
		t.Fatalf("停止RobotService测试Node: %v", err)
	}
	stopped = true
}

func robotE2EConfig(t *testing.T, root string, loginURL string) *originconfig.Snapshot {
	t.Helper()
	configDir := t.TempDir()
	document := map[string]any{"services": map[string]any{"RobotService": map[string]any{
		"blueprint": map[string]any{
			"node_dir":   filepath.Join(root, "tests", "e2e", "robot", "nodes"),
			"graph_dir":  filepath.Join(root, "tests", "e2e", "robot", "blueprints"),
			"graph_name": "login_heartbeat", "entrance_id": 1001,
		},
		"control": map[string]any{"startup_run": true},
		"target":  map[string]any{"login_url": loginURL, "show_area_id": 1},
		"identity": map[string]any{
			"platform_type": 1, "platform_id_prefix": "robot-blueprint-",
		},
		"workload": map[string]any{
			"users": 1, "ramp_up": "50ms", "duration": "6s",
			"scenario_retry_count": 0, "replace_failed": false, "max_replacements": 0,
		},
		"io": map[string]any{"workers": 2, "queue_messages": 8},
	}}}
	content, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(configDir, "robot-e2e.json"), content, 0o600); err != nil {
		t.Fatal(err)
	}
	snapshot, err := originconfig.LoadSnapshot(configDir)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func queryRobotRun(target *robotservice.RobotService) (rpcapi.RobotRunSnapshot, error) {
	type queryResult struct {
		snapshot rpcapi.RobotRunSnapshot // 查询快照。
		err      error                   // 查询错误。
	}
	result := make(chan queryResult, 1)
	if err := target.DispatchAsync(func(ctx context.Context) {
		snapshot, queryErr := target.GetRun(ctx, rpcapi.GetRobotRunRequest{})
		result <- queryResult{snapshot: snapshot, err: queryErr}
	}); err != nil {
		return rpcapi.RobotRunSnapshot{}, err
	}
	select {
	case current := <-result:
		return current.snapshot, current.err
	case <-time.After(time.Second):
		return rpcapi.RobotRunSnapshot{}, errors.New("查询RobotService运行超时")
	}
}

type robotGatewayFixture struct {
	listener     net.Listener       // 测试监听器。
	ctx          context.Context    // 测试服务上下文。
	cancel       context.CancelFunc // 测试服务取消函数。
	wg           sync.WaitGroup     // 服务协程等待组。
	errors       chan error         // 异步错误队列。
	closeOnce    sync.Once          // 关闭保护。
	loginPlayers atomic.Int64       // 登录请求计数。
	heartbeats   atomic.Int64       // 心跳请求计数。
}

func newRobotGatewayFixture(t *testing.T) *robotGatewayFixture {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	fixture := &robotGatewayFixture{
		listener: listener, ctx: ctx, cancel: cancel, errors: make(chan error, 4),
	}
	fixture.wg.Add(1)
	go fixture.serve()
	return fixture
}

func (fixture *robotGatewayFixture) address() string { return fixture.listener.Addr().String() }

func (fixture *robotGatewayFixture) serve() {
	defer fixture.wg.Done()
	connection, err := fixture.listener.Accept()
	if err != nil {
		if fixture.ctx.Err() == nil {
			fixture.report(err)
		}
		return
	}
	defer connection.Close()
	stopClose := context.AfterFunc(fixture.ctx, func() { _ = connection.Close() })
	defer stopClose()
	for {
		payload, readErr := readRobotFrame(connection)
		if readErr != nil {
			if fixture.ctx.Err() == nil && !errors.Is(readErr, io.EOF) {
				fixture.report(readErr)
			}
			return
		}
		if err = fixture.handle(connection, payload); err != nil {
			fixture.report(err)
			return
		}
	}
}

func (fixture *robotGatewayFixture) handle(connection net.Conn, payload []byte) error {
	if len(payload) < 6 {
		return errors.New("机器人请求头不完整")
	}
	messageID := commonpb.MessageID(binary.BigEndian.Uint16(payload[:2]))
	sequence := binary.BigEndian.Uint32(payload[2:6])
	switch messageID {
	case commonpb.MessageID_LoginPlayerReq:
		var request commonpb.LoginPlayerRequest
		if err := proto.Unmarshal(payload[6:], &request); err != nil ||
			request.Token != "robot-blueprint-token" || request.ShowAreaId != 1 {
			return errors.New("LoginPlayer请求无效")
		}
		fixture.loginPlayers.Add(1)
		body, err := proto.Marshal(&commonpb.LoginPlayerResult{RoleInfo: &commonpb.RoleInfo{
			AccountId: "robot-account", ShowAreaId: 1, Nickname: "Robot", Level: 1,
		}})
		if err != nil {
			return err
		}
		return writeRobotResponse(connection, commonpb.MessageID_LoginPlayerRes, sequence, body)
	case commonpb.MessageID_PlayerHeartbeatReq:
		var request commonpb.PlayerHeartbeatRequest
		if err := proto.Unmarshal(payload[6:], &request); err != nil {
			return fmt.Errorf("心跳请求无效: %w", err)
		}
		fixture.heartbeats.Add(1)
		return writeRobotResponse(connection, commonpb.MessageID_Ok, sequence, nil)
	default:
		return fmt.Errorf("未预期的机器人MessageID %d", messageID)
	}
}

func (fixture *robotGatewayFixture) report(err error) {
	select {
	case fixture.errors <- err:
	default:
	}
}

func (fixture *robotGatewayFixture) close() error {
	fixture.closeOnce.Do(func() {
		fixture.cancel()
		_ = fixture.listener.Close()
		fixture.wg.Wait()
		close(fixture.errors)
	})
	var result error
	for err := range fixture.errors {
		result = errors.Join(result, err)
	}
	return result
}

func readRobotFrame(connection net.Conn) ([]byte, error) {
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(connection, lengthBytes); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBytes)
	if length < 6 || length > robotFrameLimit {
		return nil, fmt.Errorf("机器人帧长度%d无效", length)
	}
	payload := make([]byte, length)
	_, err := io.ReadFull(connection, payload)
	return payload, err
}

func writeRobotResponse(
	connection net.Conn,
	messageID commonpb.MessageID,
	sequence uint32,
	body []byte,
) error {
	payload := make([]byte, 10+len(body))
	binary.BigEndian.PutUint16(payload[:2], uint16(messageID))
	binary.BigEndian.PutUint32(payload[2:6], sequence)
	binary.BigEndian.PutUint32(payload[6:10], uint32(commonpb.ErrorCode_ERROR_CODE_OK))
	copy(payload[10:], body)
	frame := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint32(frame[:4], uint32(len(payload)))
	copy(frame[4:], payload)
	_, err := connection.Write(frame)
	return err
}
