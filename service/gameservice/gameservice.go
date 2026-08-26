// Package gameservice 实现区服玩家加载、消息执行和连接生命周期。
package gameservice

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/duanhf2012/origin/v3/discovery"
	"github.com/duanhf2012/origin/v3/errs"
	"github.com/duanhf2012/origin/v3/log"
	"github.com/duanhf2012/origin/v3/service"
	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/protobuf/proto"
	"origingame/internal/mongodb"
	"origingame/internal/playerroute"
	commonpb "origingame/protocol/common"
	rpcapi "origingame/protocol/rpc"
	"origingame/service/gameservice/messagehandler"
	"origingame/service/gameservice/msgrouter"
	"origingame/service/gameservice/player"
	"origingame/service/gameservice/registration"
)

// Config 保存 GameService 当前仅有的区服身份和玩家容量配置。
type Config struct {
	RealAreaID     int64
	PlayerCapacity int64
}

// GameService 是区服内 Player 的唯一 RPC 和生命周期入口。
type GameService struct {
	service.Service
	config       Config
	accDB        rpcapi.DBServiceClient
	roleDB       rpcapi.DBServiceClient
	gateway      rpcapi.GatewayServiceClient
	routes       *playerroute.Store
	players      *player.Module
	registration *registration.Module
	router       *msgrouter.Router
	instance     playerroute.Instance
}

var _ rpcapi.GameService = (*GameService)(nil)
var _ discovery.IListener = (*GameService)(nil)

// OnInit 装配本区服两个 DBService、玩家 Module、实例租约和消息路由。
func (target *GameService) OnInit() error {
	if err := target.GetServiceConfigStrict("real_area_id", &target.config.RealAreaID); err != nil {
		return fmt.Errorf("读取 real_area_id 配置: %w", err)
	}
	if err := target.GetServiceConfigStrict("player_capacity", &target.config.PlayerCapacity); err != nil {
		return fmt.Errorf("读取 player_capacity 配置: %w", err)
	}
	if target.config.RealAreaID <= 0 || target.config.PlayerCapacity <= 0 {
		return errs.NewMessage(errs.CodeInvalidConfig, "GameService 区服或玩家容量配置无效")
	}
	if err := target.SetDefaultAwaitTimeout(30 * time.Second); err != nil {
		return err
	}
	if _, err := target.AddDiscoveryListener(target); err != nil {
		return fmt.Errorf("注册 GameService 服务发现日志监听器: %w", err)
	}
	node := target.GetNode()
	if node == nil || node.ID() == "" || node.SessionID() == 0 {
		return errs.NewMessage(errs.CodeInvalidConfig, "GameService 缺少 Node 运行身份")
	}
	target.instance = playerroute.Instance{
		ServiceName: target.Name(), NodeID: node.ID(),
		NodeSessionID: strconv.FormatUint(node.SessionID(), 10),
	}
	labels := map[string]string{"scope": "area", "real_area_id": strconv.FormatInt(target.config.RealAreaID, 10)}
	target.accDB = rpcapi.BindDBServiceTo(target, "AccDBService").WhereLabels(labels)
	target.roleDB = rpcapi.BindDBServiceTo(target, "RoleDBService").WhereLabels(labels)
	target.gateway = rpcapi.BindGatewayService(target).WhereLabels(map[string]string{"scope": "pub"})
	target.routes = playerroute.New(func(
		ctx context.Context,
		key string,
		request rpcapi.RedisRequest,
	) (rpcapi.RedisResult, error) {
		return target.accDB.Route(key).AwaitExecuteRedis(ctx, request)
	})
	// 玩家流程运行在 Service 调度任务内，使用 Await 释放执行权；实例租约由
	// registration 自有协程续租，必须使用普通 Call，不能依赖 Service 任务上下文。
	registrationRoutes := playerroute.New(func(
		ctx context.Context,
		key string,
		request rpcapi.RedisRequest,
	) (rpcapi.RedisResult, error) {
		return target.accDB.Route(key).CallExecuteRedis(ctx, request)
	})
	target.players = player.NewModule(func(
		ctx context.Context,
		key string,
		request rpcapi.MongoRequest,
	) (rpcapi.MongoResult, error) {
		return target.roleDB.Route(key).AwaitExecuteMongo(ctx, request)
	}, target.routes, target.config.RealAreaID, target.instance,
		func(gatewayNodeID string, request rpcapi.SendClientMessageRequest) error {
			return target.gateway.OnNode(gatewayNodeID).NotifySendClientMessage(context.Background(), request)
		},
		func(gatewayNodeID string, connectionID string) error {
			return target.gateway.OnNode(gatewayNodeID).NotifyCloseClientConnection(
				context.Background(),
				rpcapi.CloseClientConnectionRequest{GatewayConnectionID: connectionID},
			)
		},
	)
	if err := target.AddModule(target.players); err != nil {
		return err
	}
	target.registration = registration.New(registrationRoutes, target.routes, playerroute.Registration{
		RealAreaID: target.config.RealAreaID,
		Instance:   target.instance,
		MaxPlayers: target.config.PlayerCapacity,
	})
	if err := target.AddModule(target.registration); err != nil {
		return err
	}
	target.router = msgrouter.New()
	if err := messagehandler.Register(target.router); err != nil {
		return err
	}
	return target.router.Freeze()
}

// OnDiscovered 输出 GameService 当前可见的新增 Node/Service。
func (target *GameService) OnDiscovered(_ context.Context, event discovery.Event) {
	target.logDiscoveryEvent("discovered", event)
}

// OnStateChanged 输出 GameService 当前可见 Service 的状态变化。
func (target *GameService) OnStateChanged(_ context.Context, event discovery.Event) {
	target.logDiscoveryEvent("state_changed", event)
}

// OnLost 输出 GameService 当前不可见的 Node/Service。
func (target *GameService) OnLost(_ context.Context, event discovery.Event) {
	target.logDiscoveryEvent("lost", event)
}

// logDiscoveryEvent 按单个 Node/Service 输出，便于在 Debug 日志中直接检索。
func (target *GameService) logDiscoveryEvent(eventName string, event discovery.Event) {
	for _, discovered := range event.Services {
		target.Logger().Debug(
			"GameService 服务发现目录更新",
			log.String("discovery_event", eventName),
			log.String("discovered_node_id", event.NodeID),
			log.String("discovered_service_name", discovered.ServiceName),
			log.String("discovered_service_state", discoveryStateName(discovered.State)),
		)
	}
}

// logVisibleDiscoverySnapshot 输出当前可见目录，避免启动早期的发现事件难以从长日志中定位。
// GameService 仅会输出 allow_discovery 允许看见的实例，不会越过部署范围打印全局目录。
func (target *GameService) logVisibleDiscoverySnapshot() {
	for _, serviceName := range []string{"AccDBService", "RoleDBService", "GatewayService"} {
		for _, discovered := range target.ListDiscoveredServices(serviceName) {
			target.Logger().Debug(
				"GameService 服务发现当前目录",
				log.String("discovery_event", "snapshot"),
				log.String("discovered_node_id", discovered.NodeID),
				log.String("discovered_node_session_id", strconv.FormatUint(discovered.SessionID, 10)),
				log.String("discovered_service_name", discovered.ServiceName),
				log.String("discovered_service_state", discoveryStateName(discovered.State)),
				log.Any("discovered_node_labels", discovered.Labels),
			)
		}
	}
}

func discoveryStateName(state discovery.State) string {
	switch state {
	case discovery.StateRunning:
		return "running"
	case discovery.StateRetired:
		return "retired"
	default:
		return "unknown"
	}
}

// OnStart 在实例进入 Redis 候选集前确认本区服 RoleDBService 可访问。
func (target *GameService) OnStart(ctx context.Context) error {
	target.logVisibleDiscoverySnapshot()
	filter, err := bson.Marshal(bson.D{})
	if err != nil {
		return err
	}
	result, err := target.roleDB.Route("game-ready").CallExecuteMongo(ctx, rpcapi.MongoRequest{
		DispatchKey: "game-ready",
		ExecuteMode: rpcapi.MongoExecuteModeSequential,
		Operations: []rpcapi.MongoOperation{{
			Kind: rpcapi.MongoOperationKindCountDocuments, Collection: mongodb.UserInfoName,
			CountDocuments: &rpcapi.MongoCountDocuments{Filter: filter},
		}},
	})
	if err != nil {
		return fmt.Errorf("RoleDBService 未就绪: %w", err)
	}
	if result.Failure != nil || len(result.Results) != 1 ||
		result.Results[0].Status != rpcapi.MongoOperationStatusSucceeded {
		return errors.New("RoleDBService UserInfo 集合未就绪")
	}
	return nil
}

// LoginPlayer 加载或复用本实例 Player，并返回客户端可直接使用的基础角色信息。
func (target *GameService) LoginPlayer(ctx context.Context, request rpcapi.LoginPlayerRequest) (*commonpb.LoginPlayerResult, error) {
	if request.AccountID == "" || request.ShowAreaID <= 0 || request.GatewayNodeID == "" ||
		request.GatewayConnectionID == "" {
		return nil, errs.ErrInvalidArgument
	}
	if request.ExpectedGameServiceNodeSessionID != target.instance.NodeSessionID {
		return nil, errs.ErrServiceNotReady
	}
	key := playerroute.PlayerKey(request.AccountID, request.ShowAreaID)
	current := target.players.FindByKey(key)
	var err error
	if current == nil {
		target.Logger().Debug(
			"GameService 开始加载新玩家",
			log.Int64("show_area_id", request.ShowAreaID),
			log.String("gateway_node_id", request.GatewayNodeID),
			log.String("gateway_connection_id", request.GatewayConnectionID),
		)
		current, err = target.players.LoadNew(
			ctx, request.AccountID, request.ShowAreaID,
			request.GatewayNodeID, request.GatewayConnectionID,
		)
	} else {
		target.Logger().Debug(
			"GameService 开始复用在线玩家",
			log.Int64("show_area_id", request.ShowAreaID),
			log.String("gateway_node_id", request.GatewayNodeID),
			log.String("gateway_connection_id", request.GatewayConnectionID),
		)
		data := *current.DataInfo()
		if data.State == player.StateOnline && data.GatewayConnectionID != request.GatewayConnectionID {
			target.notifyKicked(data.GatewayNodeID, data.GatewayConnectionID)
		}
		err = target.players.Reconnect(ctx, current, request.GatewayNodeID, request.GatewayConnectionID)
	}
	if err != nil {
		target.Logger().Debug(
			"GameService 玩家登录失败",
			log.Int64("show_area_id", request.ShowAreaID),
			log.String("gateway_node_id", request.GatewayNodeID),
		)
		return nil, err
	}
	info := current.UserInfo()
	target.Logger().Debug(
		"GameService 玩家登录完成",
		log.Int64("show_area_id", info.ShowAreaID),
		log.Int64("real_area_id", target.config.RealAreaID),
		log.String("gateway_node_id", request.GatewayNodeID),
		log.String("gateway_connection_id", request.GatewayConnectionID),
	)
	return &commonpb.LoginPlayerResult{RoleInfo: &commonpb.RoleInfo{
		AccountId: info.AccountID, ShowAreaId: info.ShowAreaID,
		Nickname: info.Nickname, Level: info.Level, CreatedAtMs: info.CreatedAt.UnixMilli(),
	}}, nil
}

// HandlePlayerMessage 校验连接索引后把消息交给冻结的实例 Router。
func (target *GameService) HandlePlayerMessage(_ context.Context, request rpcapi.PlayerMessageRequest) error {
	if request.GatewayConnectionID == "" || request.Sequence == 0 || len(request.Body) > 4*1024 {
		return errs.ErrInvalidArgument
	}
	current := target.players.FindByConnection(request.GatewayConnectionID)
	if current == nil || current.DataInfo().GatewayConnectionID != request.GatewayConnectionID {
		return errs.ErrInvalidArgument
	}
	return target.router.Dispatch(
		current, request.GatewayConnectionID, request.MessageID, request.Sequence, request.Body,
	)
}

// PlayerDisconnected 按连接ID幂等地让 Player 进入驻留。
func (target *GameService) PlayerDisconnected(ctx context.Context, request rpcapi.PlayerDisconnectedRequest) error {
	if request.GatewayConnectionID == "" {
		return errs.ErrInvalidArgument
	}
	return target.players.Disconnect(ctx, request.GatewayConnectionID)
}

func (target *GameService) notifyKicked(gatewayNodeID string, connectionID string) {
	body, err := proto.Marshal(&commonpb.PlayerKickedNotify{
		Reason: commonpb.ErrorCode_ERROR_CODE_GATEWAY_LOGGED_IN_ELSEWHERE,
	})
	if err != nil {
		target.Logger().Error("编码顶号通知失败", log.Err(err))
		return
	}
	err = target.gateway.OnNode(gatewayNodeID).NotifySendClientMessage(
		context.Background(),
		rpcapi.SendClientMessageRequest{
			GatewayConnectionID: connectionID,
			Message: rpcapi.ClientMessage{
				MessageID: commonpb.MessageID_PlayerKicked,
				ErrorCode: commonpb.ErrorCode_ERROR_CODE_OK,
				Body:      body, CloseAfterWrite: true,
			},
		},
	)
	if err != nil {
		target.Logger().Error("提交顶号通知失败", log.Err(err))
	}
}
