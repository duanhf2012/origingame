package client

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/duanhf2012/origin/v3/errs"
	"github.com/duanhf2012/origin/v3/log"
	"github.com/duanhf2012/origin/v3/service"
	"github.com/duanhf2012/origin/v3/sysmodule/network"
	"github.com/duanhf2012/origin/v3/sysmodule/network/kcp"
	"github.com/duanhf2012/origin/v3/sysmodule/network/tcp"
	"github.com/duanhf2012/origin/v3/sysmodule/network/websocket"
	"google.golang.org/protobuf/proto"
	"origingame/internal/playerroute"
	commonpb "origingame/protocol/common"
	rpcapi "origingame/protocol/rpc"
)

// TCPConfig 控制是否创建 TCP 客户端入口。
type TCPConfig struct {
	Enabled bool
	Server  tcp.ServerConfig
}

// KCPConfig 控制是否创建 KCP 客户端入口。
type KCPConfig struct {
	Enabled bool
	Server  kcp.ServerConfig
}

// WebSocketConfig 控制是否创建 WebSocket 客户端入口。
type WebSocketConfig struct {
	Enabled bool
	Server  websocket.ServerConfig
}

// Config 保存 Gateway 外部网络入口配置。
type Config struct {
	TCP       TCPConfig
	KCP       KCPConfig
	WebSocket WebSocketConfig
}

// DefaultConfig 从 Origin 网络层完整默认值建立可严格覆盖的配置。
func DefaultConfig() Config {
	return Config{
		TCP:       TCPConfig{Server: tcp.DefaultServerConfig()},
		KCP:       KCPConfig{Server: kcp.DefaultServerConfig()},
		WebSocket: WebSocketConfig{Server: websocket.DefaultServerConfig()},
	}
}

// TokenVerifier 只返回可信 AccountID。
type TokenVerifier interface{ Verify(string) (string, error) }

// AreaResolver 只暴露当前显示区服映射。
type AreaResolver interface{ Resolve(int64) (int64, bool) }

// RouteStore 是 Gateway 登录分配使用的最小 Redis 路由能力。
type RouteStore interface {
	AssignOrGet(context.Context, playerroute.AssignRequest) (playerroute.AssignmentResult, error)
	ReleasePlayerLoad(context.Context, playerroute.Player, playerroute.Instance, string) (bool, error)
}

// Dependencies 是客户端入口需要的已装配能力，不包含入口自己的 Handler。
type Dependencies struct {
	NodeID             string
	Verifier           TokenVerifier
	Areas              AreaResolver
	Routes             RouteStore
	LoginGame          func(context.Context, playerroute.Instance, rpcapi.LoginPlayerRequest) (*commonpb.LoginPlayerResult, error)
	NotifyGameMessage  func(playerroute.Instance, rpcapi.PlayerMessageRequest) error
	NotifyDisconnected func(playerroute.Instance, string) error
}

// Module 统一拥有三种网络入口、连接状态和客户端协议处理。
type Module struct {
	service.Module
	config         Config
	dependencies   Dependencies
	connections    map[network.SessionID]*connection
	now            func() time.Time
	slowLoginAt    time.Time
	slowSuppressed uint64
}

// NewModule 创建尚未绑定 Service 的客户端入口 Module。
func NewModule(config Config, dependencies Dependencies) *Module {
	return &Module{
		config: config, dependencies: dependencies,
		connections: make(map[network.SessionID]*connection), now: time.Now,
	}
}

// OnInit 只为明确启用的协议创建 Origin 网络子 Module。
func (module *Module) OnInit() error {
	if err := module.validate(); err != nil {
		return err
	}
	handler := network.HandlerFuncs{
		Open: module.onOpen, Message: module.onMessage, Close: module.onClose,
	}
	enabled := 0
	if module.config.TCP.Enabled {
		options, err := module.config.TCP.Server.Options(handler)
		if err != nil {
			return err
		}
		server, err := tcp.NewServer(module.config.TCP.Server.Address, options)
		if err != nil {
			return err
		}
		if err = module.AddModule(server); err != nil {
			return err
		}
		enabled++
	}
	if module.config.KCP.Enabled {
		options, err := module.config.KCP.Server.Options(handler)
		if err != nil {
			return err
		}
		server, err := kcp.NewServer(module.config.KCP.Server.Address, options)
		if err != nil {
			return err
		}
		if err = module.AddModule(server); err != nil {
			return err
		}
		enabled++
	}
	if module.config.WebSocket.Enabled {
		options, err := module.config.WebSocket.Server.Options(handler)
		if err != nil {
			return err
		}
		server, err := websocket.NewServer(module.config.WebSocket.Server.Address, options)
		if err != nil {
			return err
		}
		if err = module.AddModule(server); err != nil {
			return err
		}
		enabled++
	}
	if enabled == 0 {
		return errs.NewMessage(errs.CodeInvalidConfig, "GatewayService 至少启用一种网络协议")
	}
	return nil
}

func (module *Module) validate() error {
	deps := module.dependencies
	if deps.NodeID == "" || deps.Verifier == nil || deps.Areas == nil || deps.Routes == nil ||
		deps.LoginGame == nil || deps.NotifyGameMessage == nil || deps.NotifyDisconnected == nil {
		return errs.NewMessage(errs.CodeInvalidConfig, "Gateway 客户端入口依赖不完整")
	}
	return nil
}

func (module *Module) onOpen(_ context.Context, session network.Session) error {
	if session == nil || session.ID() == "" {
		return errs.ErrInvalidArgument
	}
	if _, exists := module.connections[session.ID()]; exists {
		return errs.ErrTransportProtocol
	}
	module.connections[session.ID()] = newConnection(session, module.now())
	return nil
}

func (module *Module) onMessage(ctx context.Context, session network.Session, payload []byte) error {
	current := module.connections[session.ID()]
	if current == nil || current.session != session {
		return errs.ErrTransportClosed
	}
	if !current.allowMessage(module.now()) {
		return errs.ErrTransportOverloaded
	}
	request, err := decodeRequest(payload)
	if err != nil {
		return errs.Wrap(errs.CodeTransportProtocol, err)
	}
	if request.messageID == commonpb.MessageID_LoginPlayerReq {
		return module.handleLogin(ctx, current, request)
	}
	if current.state != stateOnline {
		return module.send(current, rpcapi.ClientMessage{
			MessageID: request.messageID, Sequence: request.sequence,
			ErrorCode: commonpb.ErrorCode_ERROR_CODE_GATEWAY_NOT_LOGGED_IN,
		})
	}
	return module.dependencies.NotifyGameMessage(current.instance, rpcapi.PlayerMessageRequest{
		GatewayConnectionID: string(session.ID()), MessageID: request.messageID,
		Sequence: request.sequence, Body: append([]byte(nil), request.body...),
	})
}

func (module *Module) handleLogin(ctx context.Context, current *connection, request request) error {
	startedAt := module.now()
	outcome, failureStage := "failure", "token"
	defer func() { module.logSlowLogin(startedAt, outcome, failureStage) }()
	var login commonpb.LoginPlayerRequest
	if err := proto.Unmarshal(request.body, &login); err != nil {
		module.Logger().Warn("解码客户端登录请求失败", log.Int("body_bytes", len(request.body)), log.Err(err))
		return module.replyLoginError(current, request.sequence, commonpb.ErrorCode_ERROR_CODE_INVALID_REQUEST)
	}
	if login.ShowAreaId <= 0 {
		return module.replyLoginError(current, request.sequence, commonpb.ErrorCode_ERROR_CODE_INVALID_REQUEST)
	}
	accountID, err := module.dependencies.Verifier.Verify(login.Token)
	if err != nil {
		return module.replyLoginError(current, request.sequence, commonpb.ErrorCode_ERROR_CODE_GATEWAY_TOKEN_INVALID)
	}
	if current.state == stateLoggingIn {
		if current.loginSequence == request.sequence {
			return nil
		}
		return module.replyLoginError(current, request.sequence, commonpb.ErrorCode_ERROR_CODE_GATEWAY_LOGIN_IN_PROGRESS)
	}
	if current.state == stateOnline {
		if current.accountID == accountID && current.showAreaID == login.ShowAreaId {
			outcome, failureStage = "success", "none"
			return module.send(current, rpcapi.ClientMessage{
				MessageID: commonpb.MessageID_LoginPlayerRes, Sequence: request.sequence,
				ErrorCode: commonpb.ErrorCode_ERROR_CODE_OK, Body: current.loginResult,
			})
		}
		return module.replyLoginError(current, request.sequence, commonpb.ErrorCode_ERROR_CODE_GATEWAY_ALREADY_LOGGED_IN)
	}
	failureStage = "area"
	realAreaID, exists := module.dependencies.Areas.Resolve(login.ShowAreaId)
	if !exists {
		return module.replyLoginError(current, request.sequence, commonpb.ErrorCode_ERROR_CODE_GATEWAY_AREA_NOT_FOUND)
	}

	current.state = stateLoggingIn
	current.loginSequence = request.sequence
	current.loginAttempt++
	attemptID := current.loginAttempt
	loginCtx, cancel := context.WithTimeout(ctx, loginTotalTimeout)
	stopSessionCancel := context.AfterFunc(current.session.Context(), cancel)
	defer cancel()
	defer stopSessionCancel()
	var result *commonpb.LoginPlayerResult
	failureStage = "route"
	err = module.Await(loginCtx, func(waitCtx context.Context) error {
		var waitErr error
		result, waitErr = module.routeAndLogin(waitCtx, current, accountID, login.ShowAreaId, realAreaID)
		return waitErr
	})
	if module.connections[current.session.ID()] != current || current.state != stateLoggingIn ||
		current.loginAttempt != attemptID {
		return nil
	}
	if err != nil || result == nil {
		current.state = stateConnected
		current.loginSequence = 0
		return module.replyLoginError(current, request.sequence, commonpb.ErrorCode_ERROR_CODE_GATEWAY_ROUTE_UNAVAILABLE)
	}
	body, err := proto.Marshal(result)
	if err != nil {
		current.state = stateConnected
		return module.replyLoginError(current, request.sequence, commonpb.ErrorCode_ERROR_CODE_INTERNAL)
	}
	current.state = stateOnline
	current.accountID = accountID
	current.showAreaID = login.ShowAreaId
	current.realAreaID = realAreaID
	current.loginResult = body
	current.loginSequence = 0
	outcome, failureStage = "success", "none"
	return module.send(current, rpcapi.ClientMessage{
		MessageID: commonpb.MessageID_LoginPlayerRes, Sequence: request.sequence,
		ErrorCode: commonpb.ErrorCode_ERROR_CODE_OK, Body: body,
	})
}

func (module *Module) logSlowLogin(startedAt time.Time, outcome string, failureStage string) {
	duration := module.now().Sub(startedAt)
	if duration < 2*time.Second {
		return
	}
	now := module.now()
	if !module.slowLoginAt.IsZero() && now.Sub(module.slowLoginAt) < 10*time.Second {
		module.slowSuppressed++
		return
	}
	module.Logger().Warn(
		"Gateway 登录处理缓慢",
		log.String("outcome", outcome),
		log.String("failure_stage", failureStage),
		log.Duration("duration", duration),
		log.Uint64("suppressed", module.slowSuppressed),
	)
	module.slowLoginAt = now
	module.slowSuppressed = 0
}

func (module *Module) routeAndLogin(
	ctx context.Context,
	current *connection,
	accountID string,
	showAreaID int64,
	realAreaID int64,
) (*commonpb.LoginPlayerResult, error) {
	player := playerroute.Player{AccountID: accountID, ShowAreaID: showAreaID, RealAreaID: realAreaID}
	excluded := make([]playerroute.Instance, 0, 3)
	for {
		assignment, err := module.dependencies.Routes.AssignOrGet(ctx, playerroute.AssignRequest{
			Player: player, GatewayConnectionID: string(current.session.ID()), Excluded: excluded,
		})
		if err != nil {
			return nil, err
		}
		switch assignment.Decision {
		case playerroute.AssignmentDecisionWait:
			if err = waitRetry(ctx); err != nil {
				return nil, err
			}
			continue
		case playerroute.AssignmentDecisionNoCapacity:
			// 候选实例可能正在启动，脚本也会分批清理崩溃后遗留的过期候选；
			// 继续受登录总Deadline约束重查，不在瞬时无容量时立即失败。
			if err = waitRetry(ctx); err != nil {
				return nil, errors.Join(errors.New("当前真实区服没有可用 GameService 容量"), err)
			}
			continue
		case playerroute.AssignmentDecisionAssigned, playerroute.AssignmentDecisionExisting:
		default:
			return nil, errors.New("未知玩家分配决策")
		}

		result, callErr := module.dependencies.LoginGame(ctx, assignment.Instance, rpcapi.LoginPlayerRequest{
			AccountID: accountID, ShowAreaID: showAreaID,
			ExpectedGameServiceNodeSessionID: assignment.Instance.NodeSessionID,
			GatewayNodeID:                    module.dependencies.NodeID, GatewayConnectionID: string(current.session.ID()),
		})
		if callErr == nil {
			current.instance = assignment.Instance
			return result, nil
		}
		if assignment.Decision == playerroute.AssignmentDecisionAssigned && definitelyNotExecuted(callErr) {
			_, releaseErr := module.dependencies.Routes.ReleasePlayerLoad(
				ctx, player, assignment.Instance, string(current.session.ID()),
			)
			if releaseErr != nil {
				return nil, errors.Join(callErr, releaseErr)
			}
			excluded = append(excluded, assignment.Instance)
			if len(excluded) >= 3 {
				return nil, callErr
			}
			continue
		}
		// 请求执行状态未知或已有在线归属时，只等待 Redis 重新确认，不能另选所有者。
		if err = waitRetry(ctx); err != nil {
			return nil, errors.Join(callErr, err)
		}
	}
}

func definitelyNotExecuted(err error) bool {
	return errors.Is(err, errs.ErrRPCNoRoute) || errors.Is(err, errs.ErrServiceNotReady) ||
		errors.Is(err, errs.ErrServiceStopping) || errors.Is(err, errs.ErrServiceStopped) ||
		errors.Is(err, errs.ErrServiceRetired) || errors.Is(err, errs.ErrServiceQueueFull)
}

func waitRetry(ctx context.Context) error {
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-timer.C:
		return nil
	}
}

func (module *Module) onClose(_ context.Context, session network.Session, _ error) {
	current := module.connections[session.ID()]
	if current == nil || current.session != session {
		return
	}
	delete(module.connections, session.ID())
	if current.state == stateOnline {
		_ = module.dependencies.NotifyDisconnected(current.instance, string(session.ID()))
	}
}

// SendClientMessage 编码并写入当前连接；连接不存在时幂等成功。
func (module *Module) SendClientMessage(request rpcapi.SendClientMessageRequest) error {
	current := module.connections[network.SessionID(request.GatewayConnectionID)]
	if current == nil {
		return nil
	}
	if request.Message.CloseAfterWrite {
		payload, err := encodeClientMessage(request.Message)
		if err != nil {
			return fmt.Errorf("编码客户端最终消息: %w", err)
		}
		finalSession, ok := current.session.(network.FinalMessageSession)
		if !ok {
			return errors.New("当前网络 Session 不支持完整写后关闭")
		}
		return finalSession.SendAndClose(payload, errs.ErrTransportClosed)
	}
	return module.send(current, request.Message)
}

// CloseClientConnection 幂等关闭当前 Node 上的指定连接。
func (module *Module) CloseClientConnection(connectionID string) {
	if current := module.connections[network.SessionID(connectionID)]; current != nil {
		current.session.Close(errs.ErrTransportClosed)
	}
}

func (module *Module) replyLoginError(current *connection, sequence uint32, code commonpb.ErrorCode) error {
	return module.send(current, rpcapi.ClientMessage{
		MessageID: commonpb.MessageID_LoginPlayerRes, Sequence: sequence, ErrorCode: code,
	})
}

func (module *Module) send(current *connection, message rpcapi.ClientMessage) error {
	payload, err := encodeClientMessage(message)
	if err != nil {
		return fmt.Errorf("编码客户端消息: %w", err)
	}
	return current.session.Send(payload)
}
