package e2e

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	commonpb "origingame/protocol/common"
)

const maxTestFrameSize = 4 * 1024

type loginHTTPResponse struct {
	ECode int32  `json:"ECode"`
	Token string `json:"Token"`
}

type clientMessage struct {
	messageID commonpb.MessageID
	sequence  uint32
	errorCode commonpb.ErrorCode
	body      []byte
}

// TestLoginFlow 验证 HTTP 登录、玩家加载、心跳响应，以及新连接顶掉旧连接。
// 默认跳过；先启动本地五个 Node，再设置 ORIGINGAME_E2E=1 执行。
func TestLoginFlow(t *testing.T) {
	if os.Getenv("ORIGINGAME_E2E") != "1" {
		t.Skip("设置 ORIGINGAME_E2E=1 后执行进程级登录链路测试")
	}
	loginURL := envOrDefault("ORIGINGAME_LOGIN_URL", "http://127.0.0.1:8080/api/v1/login")
	gatewayAddress := envOrDefault("ORIGINGAME_GATEWAY_TCP", "127.0.0.1:9001")
	token := requestLoginToken(t, loginURL)

	first := dialGateway(t, gatewayAddress)
	defer first.Close()
	loginPlayer(t, first, token, 1)
	heartbeat(t, first, 2)

	second := dialGateway(t, gatewayAddress)
	loginPlayer(t, second, token, 3)

	_ = first.SetReadDeadline(time.Now().Add(5 * time.Second))
	kicked := readClientMessage(t, first)
	if kicked.messageID != commonpb.MessageID_PlayerKicked || kicked.sequence != 0 {
		t.Fatalf("旧连接收到消息 id=%d sequence=%d，期望顶号推送", kicked.messageID, kicked.sequence)
	}
	var notification commonpb.PlayerKickedNotify
	if err := proto.Unmarshal(kicked.body, &notification); err != nil {
		t.Fatalf("解码顶号通知: %v", err)
	}
	if notification.Reason != commonpb.ErrorCode_ERROR_CODE_GATEWAY_LOGGED_IN_ELSEWHERE {
		t.Fatalf("顶号原因 = %v", notification.Reason)
	}
	one := make([]byte, 1)
	if _, err := first.Read(one); !errors.Is(err, io.EOF) {
		t.Fatalf("旧连接最终消息后未关闭: %v", err)
	}
	heartbeat(t, second, 4)
	if err := second.Close(); err != nil {
		t.Fatalf("关闭第二条连接: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// 断线驻留期间重连必须复用原 GameService 与内存 Player。
	third := dialGateway(t, gatewayAddress)
	defer third.Close()
	loginPlayer(t, third, token, 5)
	heartbeat(t, third, 6)
}

func requestLoginToken(t *testing.T, loginURL string) string {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"PlatType":    1,
		"PlatId":      fmt.Sprintf("e2e-%d", time.Now().UnixNano()),
		"AccessToken": "development",
	})
	if err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequest(http.MethodPost, loginURL, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("调用 LoginService: %v", err)
	}
	defer response.Body.Close()
	var result loginHTTPResponse
	if err = json.NewDecoder(io.LimitReader(response.Body, 64*1024)).Decode(&result); err != nil {
		t.Fatalf("解码 LoginService 响应: %v", err)
	}
	if response.StatusCode != http.StatusOK || result.ECode != 0 || result.Token == "" {
		t.Fatalf("LoginService status=%d ecode=%d token_empty=%t", response.StatusCode, result.ECode, result.Token == "")
	}
	return result.Token
}

func dialGateway(t *testing.T, address string) net.Conn {
	t.Helper()
	connection, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		t.Fatalf("连接 Gateway: %v", err)
	}
	return connection
}

func loginPlayer(t *testing.T, connection net.Conn, token string, sequence uint32) {
	t.Helper()
	body, err := proto.Marshal(&commonpb.LoginPlayerRequest{Token: token, ShowAreaId: 1})
	if err != nil {
		t.Fatal(err)
	}
	writeRequest(t, connection, commonpb.MessageID_LoginPlayerReq, sequence, body)
	response := readClientMessage(t, connection)
	if response.messageID != commonpb.MessageID_LoginPlayerRes || response.sequence != sequence ||
		response.errorCode != commonpb.ErrorCode_ERROR_CODE_OK {
		t.Fatalf("玩家登录响应 id=%d sequence=%d error=%v", response.messageID, response.sequence, response.errorCode)
	}
	var result commonpb.LoginPlayerResult
	if err = proto.Unmarshal(response.body, &result); err != nil {
		t.Fatalf("解码玩家登录响应: %v", err)
	}
	if result.RoleInfo == nil || result.RoleInfo.AccountId == "" || result.RoleInfo.ShowAreaId != 1 {
		t.Fatalf("玩家基础数据无效: %+v", result.RoleInfo)
	}
}

func heartbeat(t *testing.T, connection net.Conn, sequence uint32) {
	t.Helper()
	body, err := proto.Marshal(&commonpb.PlayerHeartbeatRequest{})
	if err != nil {
		t.Fatal(err)
	}
	writeRequest(t, connection, commonpb.MessageID_PlayerHeartbeatReq, sequence, body)
	response := readClientMessage(t, connection)
	if response.messageID != commonpb.MessageID_Ok || response.sequence != sequence ||
		response.errorCode != commonpb.ErrorCode_ERROR_CODE_OK || len(response.body) != 0 {
		t.Fatalf("心跳响应 id=%d sequence=%d error=%v body=%d", response.messageID, response.sequence, response.errorCode, len(response.body))
	}
}

func writeRequest(t *testing.T, connection net.Conn, messageID commonpb.MessageID, sequence uint32, body []byte) {
	t.Helper()
	payload := make([]byte, 6+len(body))
	binary.BigEndian.PutUint16(payload[:2], uint16(messageID))
	binary.BigEndian.PutUint32(payload[2:6], sequence)
	copy(payload[6:], body)
	frame := make([]byte, 4+len(payload))
	binary.BigEndian.PutUint32(frame[:4], uint32(len(payload)))
	copy(frame[4:], payload)
	_ = connection.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if _, err := connection.Write(frame); err != nil {
		t.Fatalf("写入 Gateway 请求: %v", err)
	}
}

func readClientMessage(t *testing.T, connection net.Conn) clientMessage {
	t.Helper()
	_ = connection.SetReadDeadline(time.Now().Add(5 * time.Second))
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(connection, lengthBytes); err != nil {
		t.Fatalf("读取 Gateway 帧长度: %v", err)
	}
	length := binary.BigEndian.Uint32(lengthBytes)
	if length < 6 || length > maxTestFrameSize {
		t.Fatalf("Gateway 帧长度 = %d", length)
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(connection, payload); err != nil {
		t.Fatalf("读取 Gateway 帧: %v", err)
	}
	message := clientMessage{
		messageID: commonpb.MessageID(binary.BigEndian.Uint16(payload[:2])),
		sequence:  binary.BigEndian.Uint32(payload[2:6]),
	}
	if message.sequence == 0 {
		message.body = payload[6:]
		return message
	}
	if len(payload) < 10 {
		t.Fatalf("Gateway 响应头长度 = %d", len(payload))
	}
	message.errorCode = commonpb.ErrorCode(int32(binary.BigEndian.Uint32(payload[6:10])))
	message.body = payload[10:]
	return message
}

func envOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
