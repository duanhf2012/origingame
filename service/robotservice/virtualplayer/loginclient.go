package virtualplayer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const maxLoginResponseBytes = 1024 * 1024

// LoginClientConfig 保存真实LoginService HTTP契约所需的最小配置。
type LoginClientConfig struct {
	URL              string // LoginService HTTP 地址。
	ShowAreaID       int64  // 目标显示区服标识。
	PlatformType     int32  // 测试平台类型。
	PlatformIDPrefix string // 测试账号标识前缀。
	Workers          int    // 连接池容量。
}

// LoginResult 只保存后续连接必需的数据；Token禁止进入蓝图和日志。
type LoginResult struct {
	Token          string // 登录成功后取得的游戏 Token。
	GatewayAddress string // 目标区服的 TCP Gateway 地址。
}

// LoginClient 复用一个有界Transport执行机器人HTTP登录。
type LoginClient struct {
	config LoginClientConfig // 登录协议配置。
	client *http.Client      // 复用的有界 HTTP 客户端。
}

// NewLoginClient 创建不自动重试的HTTP客户端。
func NewLoginClient(config LoginClientConfig) (*LoginClient, error) {
	if strings.TrimSpace(config.URL) == "" || config.ShowAreaID <= 0 ||
		strings.TrimSpace(config.PlatformIDPrefix) == "" || config.Workers <= 0 {
		return nil, errors.New("机器人LoginClient配置无效")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = config.Workers
	transport.MaxIdleConnsPerHost = config.Workers
	transport.MaxConnsPerHost = config.Workers
	return &LoginClient{
		config: config,
		client: &http.Client{
			Transport: transport,
			Timeout:   15 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return errors.New("机器人登录不允许HTTP重定向")
			},
		},
	}, nil
}

type loginRequest struct {
	PlatformType int32  `json:"PlatType"`    // 测试平台类型。
	PlatformID   string `json:"PlatId"`      // 测试平台账号标识。
	AccessToken  string `json:"AccessToken"` // 平台访问凭证。
}

type loginResponse struct {
	ErrorCode int32      `json:"ECode"`    // LoginService 错误码。
	Token     string     `json:"Token"`    // 成功时的游戏 Token。
	Areas     []areaInfo `json:"AreaList"` // 可选区服列表。
}

type areaInfo struct {
	ShowAreaID int64      `json:"ShowAreaId"` // 显示区服标识。
	Gates      []gateInfo `json:"GateList"`   // Gateway 列表。
}

type gateInfo struct {
	Protocol string `json:"Protocol"` // 连接协议。
	Address  string `json:"Address"`  // Gateway 地址。
}

// Login 使用稳定robot_id生成测试身份，并选择配置区服的TCP Gateway入口。
func (client *LoginClient) Login(ctx context.Context, robotID int64) (LoginResult, error) {
	if client == nil || client.client == nil || ctx == nil || robotID <= 0 {
		return LoginResult{}, errors.New("机器人HTTP登录参数无效")
	}
	body, err := json.Marshal(loginRequest{
		PlatformType: client.config.PlatformType,
		PlatformID:   client.config.PlatformIDPrefix + strconv.FormatInt(robotID, 10),
	})
	if err != nil {
		return LoginResult{}, fmt.Errorf("编码机器人登录请求: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.config.URL, bytes.NewReader(body))
	if err != nil {
		return LoginResult{}, fmt.Errorf("创建机器人登录请求: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.client.Do(request)
	if err != nil {
		return LoginResult{}, fmt.Errorf("请求LoginService: %w", err)
	}
	defer response.Body.Close()
	limited := io.LimitReader(response.Body, maxLoginResponseBytes+1)
	responseBody, err := io.ReadAll(limited)
	if err != nil {
		return LoginResult{}, fmt.Errorf("读取LoginService响应: %w", err)
	}
	if len(responseBody) > maxLoginResponseBytes {
		return LoginResult{}, errors.New("LoginService响应超过1MiB上限")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return LoginResult{}, fmt.Errorf("LoginService HTTP状态码%d", response.StatusCode)
	}
	var decoded loginResponse
	if err = json.Unmarshal(responseBody, &decoded); err != nil {
		return LoginResult{}, fmt.Errorf("解码LoginService响应: %w", err)
	}
	if decoded.ErrorCode != 0 {
		return LoginResult{}, fmt.Errorf("LoginService返回错误码%d", decoded.ErrorCode)
	}
	if strings.TrimSpace(decoded.Token) == "" {
		return LoginResult{}, errors.New("LoginService成功响应缺少Token")
	}
	for _, area := range decoded.Areas {
		if area.ShowAreaID != client.config.ShowAreaID {
			continue
		}
		for _, gate := range area.Gates {
			if strings.EqualFold(strings.TrimSpace(gate.Protocol), "tcp") && strings.TrimSpace(gate.Address) != "" {
				return LoginResult{Token: decoded.Token, GatewayAddress: strings.TrimSpace(gate.Address)}, nil
			}
		}
		return LoginResult{}, errors.New("目标区服没有TCP Gateway入口")
	}
	return LoginResult{}, errors.New("LoginService响应不包含目标显示区服")
}

// Close 幂等关闭HTTP空闲连接。
func (client *LoginClient) Close() {
	if client == nil || client.client == nil {
		return
	}
	if transport, ok := client.client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
}
