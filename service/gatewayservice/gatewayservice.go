// Package gatewayservice 实现共享客户端网关的装配和 RPC 下行入口。
package gatewayservice

import (
	"context"
	"errors"
	"fmt"
	"time"

	originconfig "github.com/duanhf2012/origin/v3/config"
	"github.com/duanhf2012/origin/v3/errs"
	"github.com/duanhf2012/origin/v3/service"
	"origingame/internal/dbexecutor"
	"origingame/internal/playerownership"
	rpcapi "origingame/protocol/rpc"
	"origingame/service/gatewayservice/area"
	"origingame/service/gatewayservice/client"
	"origingame/service/gatewayservice/token"
)

const gatewayReadyDispatchKey = "gateway-ready"

// Config 保存 Gateway 自身配置；数据库与 Redis 连接只属于 AccDBService。
type Config struct {
	Token  token.Config  `json:"token"`
	Area   AreaConfig    `json:"area"`
	Client client.Config `json:"client"`
}

// AreaConfig 保存真实系统时间区服映射刷新周期。
type AreaConfig struct {
	RefreshInterval originconfig.Duration `json:"refresh_interval"`
}

// GatewayService 装配公共数据客户端、区服映射和统一网络入口。
type GatewayService struct {
	service.Service
	config         Config
	accDB          rpcapi.DBServiceClient
	ownershipStore *playerownership.PlayerOwnershipStore
	areas          area.Catalog
	refresh        *area.AreaRefreshModule
	client         *client.GatewayClientModule
}

var _ rpcapi.GatewayService = (*GatewayService)(nil)

// OnInit 按区服快照先于网络监听的顺序装配 Module。
func (target *GatewayService) OnInit() error {
	// 读取并校验 Gateway 本地配置。
	if err := target.loadConfig(); err != nil {
		return err
	}
	if target.config.Area.RefreshInterval.Duration() <= 0 {
		return errs.NewMessage(errs.CodeInvalidConfig, "Gateway area.refresh_interval 必须为正数")
	}
	if err := target.SetDefaultAwaitTimeout(30 * time.Second); err != nil {
		return err
	}
	// 初始化只在 Gateway 本地执行的 Token 验证器。
	verifier, err := token.NewVerifier(target.config.Token)
	if err != nil {
		return fmt.Errorf("初始化 Gateway TokenVerifier: %w", err)
	}
	node := target.GetNode()
	if node == nil || node.ID() == "" {
		return errs.NewMessage(errs.CodeInvalidConfig, "GatewayService 缺少 NodeID")
	}

	// 通过公共数据域装配玩家归属和区服映射读取能力。
	target.accDB = rpcapi.BindDBServiceTo(target, "AccDBService").WhereLabels(map[string]string{"scope": "pub"})
	callDB := dbexecutor.NewCallExecutor(target.accDB)
	target.ownershipStore = playerownership.NewPlayerOwnershipStore(callDB)
	repository := area.NewMongoRepository(callDB)
	target.refresh = area.NewAreaRefreshModule(
		target.config.Area.RefreshInterval.Duration(), repository, &target.areas,
	)
	if err = target.AddModule(target.refresh); err != nil {
		return err
	}
	// 在网络监听前注册统一的客户端入口。
	target.client = client.NewGatewayClientModule(target.config.Client, client.Dependencies{
		NodeID:         node.ID(),
		Verifier:       verifier,
		Areas:          &target.areas,
		OwnershipStore: target.ownershipStore,
		GameServices:   client.NewRPCGameServiceCaller(target),
	})
	return target.AddModule(target.client)
}

// OnStart 确认公共 Redis 和登记 Script 可通过 AccDBService 执行。
func (target *GatewayService) OnStart(ctx context.Context) error {
	// 探测公共 Redis，避免依赖未就绪时对外提供登录服务。
	result, err := target.accDB.Route(gatewayReadyDispatchKey).CallExecuteRedis(ctx, rpcapi.RedisRequest{
		DispatchKey: gatewayReadyDispatchKey,
		ExecuteMode: rpcapi.RedisExecuteModeCommand,
		Commands:    []rpcapi.RedisCommand{{Name: "EXISTS", Args: [][]byte{[]byte("gateway-ready-probe")}}},
	})
	if err != nil || result.Failure != nil || len(result.Results) != 1 ||
		result.Results[0].Status != rpcapi.RedisCommandStatusSucceeded {
		if err != nil {
			return fmt.Errorf("Gateway AccDBService Redis 未就绪: %w", err)
		}
		return errors.New("Gateway AccDBService Redis 未就绪")
	}
	return nil
}

// SendClientMessage 实现 GameService 的统一客户端下行入口。
func (target *GatewayService) SendClientMessage(_ context.Context, request rpcapi.SendClientMessageRequest) error {
	// 拒绝缺少本地连接标识的下行请求。
	if request.GatewayConnectionID == "" {
		return errs.ErrInvalidArgument
	}
	return target.client.SendClientMessage(request)
}

// CloseClientConnection 幂等关闭当前 Gateway Node 上的指定连接。
func (target *GatewayService) CloseClientConnection(_ context.Context, request rpcapi.CloseClientConnectionRequest) error {
	// 拒绝缺少本地连接标识的关闭请求。
	if request.GatewayConnectionID == "" {
		return errs.ErrInvalidArgument
	}
	target.client.CloseClientConnection(request.GatewayConnectionID)
	return nil
}

func (target *GatewayService) loadConfig() error {
	// 先填充 Origin 网络层默认值，再由本地配置严格覆盖。
	target.config.Client = client.DefaultConfig()
	sections := []struct {
		path string
		to   any
	}{
		{"token", &target.config.Token},
		{"area", &target.config.Area},
		{"tcp", &target.config.Client.TCP},
		{"kcp", &target.config.Client.KCP},
		{"websocket", &target.config.Client.WebSocket},
	}
	// 按职责分别读取，避免遗漏任一传输协议配置。
	for _, section := range sections {
		if err := target.GetServiceConfigStrict(section.path, section.to); err != nil {
			return fmt.Errorf("读取 Gateway %s 配置: %w", section.path, err)
		}
	}
	return nil
}
