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
	"origingame/internal/playerownership"
	commonpb "origingame/protocol/common"
	rpcapi "origingame/protocol/rpc"
)

// TCPConfig 控制是否创建 TCP 客户端入口。
type TCPConfig struct {
	Enabled bool             `json:"enabled"`
	Server  tcp.ServerConfig `json:"server"`
}

// KCPConfig 控制是否创建 KCP 客户端入口。
type KCPConfig struct {
	Enabled bool             `json:"enabled"`
	Server  kcp.ServerConfig `json:"server"`
}

// WebSocketConfig 控制是否创建 WebSocket 客户端入口。
type WebSocketConfig struct {
	Enabled bool                   `json:"enabled"`
	Server  websocket.ServerConfig `json:"server"`
}

// Config 保存 Gateway 外部网络入口配置。
type Config struct {
	TCP       TCPConfig       `json:"tcp"`
	KCP       KCPConfig       `json:"kcp"`
	WebSocket WebSocketConfig `json:"websocket"`
}

// DefaultConfig 从 Origin 网络层完整默认值建立可严格覆盖的配置。
func DefaultConfig() Config {
	// 复制三种 Origin 网络服务的默认配置。
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

// OwnershipStore 是 Gateway 登录分配使用的最小 Redis 玩家归属能力。
type OwnershipStore interface {
	AssignOrGet(context.Context, playerownership.AssignRequest) (playerownership.AssignmentResult, error)
	ReleasePlayerLoad(context.Context, playerownership.Player, playerownership.GameServiceInstance, string) (bool, error)
}

// Dependencies 是客户端入口需要的已装配能力，不包含入口自己的 Handler。
type Dependencies struct {
	NodeID         string
	Verifier       TokenVerifier
	Areas          AreaResolver
	OwnershipStore OwnershipStore
	GameServices   GameServiceCaller
}

// GatewayClientModule 统一拥有三种网络入口、连接状态和客户端协议处理。
type GatewayClientModule struct {
	service.Module
	config         Config
	dependencies   Dependencies
	connections    map[network.SessionID]*connection
	now            func() time.Time
	slowLoginAt    time.Time
	slowSuppressed uint64
}

// NewGatewayClientModule 创建尚未绑定 Service 的客户端入口 Module。
func NewGatewayClientModule(config Config, dependencies Dependencies) *GatewayClientModule {
	// 初始化连接表，并使用真实系统时间处理网络限流。
	return &GatewayClientModule{
		config: config, dependencies: dependencies,
		connections: make(map[network.SessionID]*connection), now: time.Now,
	}
}

// OnInit 只为明确启用的协议创建 Origin 网络子 Module。
func (module *GatewayClientModule) OnInit() error {
	// 启动前确认所有跨服务依赖都已装配。
	if err := module.validate(); err != nil {
		return err
	}
	// 三种传输共用同一组会话回调。
	handler := network.HandlerFuncs{
		Open: module.onOpen, Message: module.onMessage, Close: module.onClose,
	}
	enabled := 0
	if module.config.TCP.Enabled {
		// 按配置创建 TCP 网络入口。
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
		// 按配置创建 KCP 网络入口。
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
		// 按配置创建 WebSocket 网络入口。
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

func (module *GatewayClientModule) validate() error {
	// 入口缺少任一协调依赖时禁止开始监听。
	deps := module.dependencies
	if deps.NodeID == "" || deps.Verifier == nil || deps.Areas == nil || deps.OwnershipStore == nil ||
		deps.GameServices == nil {
		return errs.NewMessage(errs.CodeInvalidConfig, "Gateway 客户端入口依赖不完整")
	}
	return nil
}

func (module *GatewayClientModule) onOpen(_ context.Context, session network.Session) error {
	// 拒绝无标识或重复注册的底层会话。
	if session == nil || session.ID() == "" {
		return errs.ErrInvalidArgument
	}
	if _, exists := module.connections[session.ID()]; exists {
		return errs.ErrTransportProtocol
	}
	// 建立连接级登录状态和限流窗口。
	module.connections[session.ID()] = newConnection(session, module.now())
	return nil
}

func (module *GatewayClientModule) onMessage(ctx context.Context, session network.Session, payload []byte) error {
	// 只处理当前仍登记的同一网络会话。
	current := module.connections[session.ID()]
	if current == nil || current.session != session {
		return errs.ErrTransportClosed
	}
	if !current.allowMessage(module.now()) {
		return errs.ErrTransportOverloaded
	}
	// 解码并校验客户端通用消息头。
	request, err := decodeRequest(payload)
	if err != nil {
		return errs.Wrap(errs.CodeTransportProtocol, err)
	}
	if request.messageID == commonpb.MessageID_LoginPlayerReq {
		// 登录请求独占完整的鉴权和归属流程。
		return module.handleLogin(ctx, current, request)
	}
	// 未登录连接不能投递游戏业务消息。
	if current.state != stateOnline {
		return module.send(current, rpcapi.ClientMessage{
			MessageID: request.messageID, Sequence: request.sequence,
			ErrorCode: commonpb.ErrorCode_ERROR_CODE_GATEWAY_NOT_LOGGED_IN,
		})
	}
	// 将在线玩家消息路由到已确认的归属实例。
	return module.dependencies.GameServices.HandlePlayerMessage(current.gameService, rpcapi.PlayerMessageRequest{
		GatewayConnectionID: string(session.ID()), MessageID: request.messageID,
		Sequence: request.sequence, Body: append([]byte(nil), request.body...),
	})
}

func (module *GatewayClientModule) handleLogin(ctx context.Context, current *connection, request request) error {
	// 统计完整登录流程耗时，并记录最终结果。
	startedAt := module.now()
	outcome, failureStage := "failure", "token"
	defer func() { module.logSlowLogin(startedAt, outcome, failureStage) }()
	// 解码客户端登录参数。
	var login commonpb.LoginPlayerRequest
	if err := proto.Unmarshal(request.body, &login); err != nil {
		module.Logger().Warn("解码客户端登录请求失败", log.Int("body_bytes", len(request.body)), log.Err(err))
		return module.replyLoginError(current, request.sequence, commonpb.ErrorCode_ERROR_CODE_INVALID_REQUEST)
	}
	// 拒绝缺少有效显示区服的请求。
	if login.ShowAreaId <= 0 {
		module.Logger().Debug(
			"Gateway 登录请求区服无效",
			log.String("connection_id", string(current.session.ID())),
			log.Int64("show_area_id", login.ShowAreaId),
		)
		return module.replyLoginError(current, request.sequence, commonpb.ErrorCode_ERROR_CODE_INVALID_REQUEST)
	}
	// 本地验签后仅保留可信账号标识。
	accountID, err := module.dependencies.Verifier.Verify(login.Token)
	if err != nil {
		module.Logger().Debug(
			"Gateway 登录凭证无效",
			log.String("connection_id", string(current.session.ID())),
			log.Int64("show_area_id", login.ShowAreaId),
		)
		return module.replyLoginError(current, request.sequence, commonpb.ErrorCode_ERROR_CODE_GATEWAY_TOKEN_INVALID)
	}
	// Token 仅在当前函数和下游 RPC 中使用，日志只保留连接与区服定位信息。
	module.Logger().Debug(
		"Gateway 收到客户端登录请求",
		log.String("connection_id", string(current.session.ID())),
		log.Uint32("sequence", request.sequence),
		log.Int64("show_area_id", login.ShowAreaId),
	)
	// 处理登录中的重传，拒绝并发登录。
	if current.state == stateLoggingIn {
		if current.loginSequence == request.sequence {
			return nil
		}
		return module.replyLoginError(current, request.sequence, commonpb.ErrorCode_ERROR_CODE_GATEWAY_LOGIN_IN_PROGRESS)
	}
	// 已登录连接只允许相同账号和区服的幂等重试。
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
	// 解析显示区服当前对应的真实区服。
	failureStage = "area"
	realAreaID, exists := module.dependencies.Areas.Resolve(login.ShowAreaId)
	if !exists {
		module.Logger().Debug(
			"Gateway 登录区服不存在",
			log.String("connection_id", string(current.session.ID())),
			log.Int64("show_area_id", login.ShowAreaId),
		)
		return module.replyLoginError(current, request.sequence, commonpb.ErrorCode_ERROR_CODE_GATEWAY_AREA_NOT_FOUND)
	}
	module.Logger().Debug(
		"Gateway 登录区服已解析",
		log.String("connection_id", string(current.session.ID())),
		log.Int64("show_area_id", login.ShowAreaId),
		log.Int64("real_area_id", realAreaID),
	)

	// 标记本次尝试，并绑定请求与连接终止的总时限。
	current.state = stateLoggingIn
	current.loginSequence = request.sequence
	current.loginAttempt++
	attemptID := current.loginAttempt
	loginCtx, cancel := context.WithTimeout(ctx, loginTotalTimeout)
	stopSessionCancel := context.AfterFunc(current.session.Context(), cancel)
	defer cancel()
	defer stopSessionCancel()
	// 串行执行 Redis 归属分配和 GameService 玩家加载。
	var result *commonpb.LoginPlayerResult
	failureStage = "ownership"
	err = module.Await(loginCtx, func(waitCtx context.Context) error {
		var waitErr error
		result, waitErr = module.assignAndLogin(waitCtx, current, accountID, login.ShowAreaId, realAreaID)
		return waitErr
	})
	// 忽略连接关闭或后续尝试已取代的过期结果。
	if module.connections[current.session.ID()] != current || current.state != stateLoggingIn ||
		current.loginAttempt != attemptID {
		return nil
	}
	// 登录失败时恢复未登录状态并返回统一路由错误。
	if err != nil || result == nil {
		current.state = stateConnected
		current.loginSequence = 0
		module.Logger().Debug(
			"Gateway 登录路由或玩家加载失败",
			log.String("connection_id", string(current.session.ID())),
			log.Int64("show_area_id", login.ShowAreaId),
			log.Int64("real_area_id", realAreaID),
		)
		return module.replyLoginError(current, request.sequence, commonpb.ErrorCode_ERROR_CODE_GATEWAY_ROUTE_UNAVAILABLE)
	}
	// 序列化 GameService 返回的登录结果。
	body, err := proto.Marshal(result)
	if err != nil {
		current.state = stateConnected
		return module.replyLoginError(current, request.sequence, commonpb.ErrorCode_ERROR_CODE_INTERNAL)
	}
	// 记录当前连接的在线归属和幂等响应缓存。
	current.state = stateOnline
	current.accountID = accountID
	current.showAreaID = login.ShowAreaId
	current.realAreaID = realAreaID
	current.loginResult = body
	current.loginSequence = 0
	outcome, failureStage = "success", "none"
	module.Logger().Debug(
		"Gateway 客户端登录完成",
		log.String("connection_id", string(current.session.ID())),
		log.Int64("show_area_id", login.ShowAreaId),
		log.Int64("real_area_id", realAreaID),
		log.String("game_service_node", current.gameService.NodeID),
	)
	return module.send(current, rpcapi.ClientMessage{
		MessageID: commonpb.MessageID_LoginPlayerRes, Sequence: request.sequence,
		ErrorCode: commonpb.ErrorCode_ERROR_CODE_OK, Body: body,
	})
}

func (module *GatewayClientModule) logSlowLogin(startedAt time.Time, outcome string, failureStage string) {
	// 两秒内的登录不产生慢请求日志。
	duration := module.now().Sub(startedAt)
	if duration < 2*time.Second {
		return
	}
	// 十秒窗口内合并重复慢日志。
	now := module.now()
	if !module.slowLoginAt.IsZero() && now.Sub(module.slowLoginAt) < 10*time.Second {
		module.slowSuppressed++
		return
	}
	// 输出本窗口第一条慢登录及已合并次数。
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

func (module *GatewayClientModule) assignAndLogin(
	ctx context.Context,
	current *connection,
	accountID string,
	showAreaID int64,
	realAreaID int64,
) (*commonpb.LoginPlayerResult, error) {
	// 固定玩家身份，并记录本次已排除的失效实例。
	player := playerownership.Player{AccountID: accountID, ShowAreaID: showAreaID, RealAreaID: realAreaID}
	excludedGameServices := make([]playerownership.GameServiceInstance, 0, 3)
	for {
		// 从公共归属存储取得已有归属或可用实例。
		assignment, err := module.dependencies.OwnershipStore.AssignOrGet(ctx, playerownership.AssignRequest{
			Player: player, GatewayConnectionID: string(current.session.ID()), ExcludedGameServices: excludedGameServices,
		})
		if err != nil {
			return nil, err
		}
		switch assignment.Decision {
		case playerownership.AssignmentDecisionWait:
			// 其他 Gateway 正在建立归属，等待其完成。
			if err = waitRetry(ctx); err != nil {
				return nil, err
			}
			continue
		case playerownership.AssignmentDecisionNoCapacity:
			// 候选实例可能正在启动，脚本也会分批清理崩溃后遗留的过期候选；
			// 继续受登录总Deadline约束重查，不在瞬时无容量时立即失败。
			if err = waitRetry(ctx); err != nil {
				return nil, errors.Join(errors.New("当前真实区服没有可用 GameService 容量"), err)
			}
			continue
		case playerownership.AssignmentDecisionAssigned, playerownership.AssignmentDecisionExisting:
		default:
			return nil, errors.New("未知玩家分配决策")
		}

		// 请求归属 GameService 加载玩家并确认上线。
		result, callErr := module.dependencies.GameServices.LoginPlayer(ctx, assignment.GameService, rpcapi.LoginPlayerRequest{
			AccountID: accountID, ShowAreaID: showAreaID,
			ExpectedGameServiceNodeSessionID: assignment.GameService.NodeSessionID,
			GatewayNodeID:                    module.dependencies.NodeID, GatewayConnectionID: string(current.session.ID()),
		})
		if callErr == nil {
			// 仅在加载成功后发布连接的 GameService 归属。
			current.gameService = assignment.GameService
			return result, nil
		}
		if assignment.Decision == playerownership.AssignmentDecisionAssigned && definitelyNotExecuted(callErr) {
			// 未执行的调用可释放预占负载后改选其他实例。
			_, releaseErr := module.dependencies.OwnershipStore.ReleasePlayerLoad(
				ctx, player, assignment.GameService, string(current.session.ID()),
			)
			if releaseErr != nil {
				return nil, errors.Join(callErr, releaseErr)
			}
			excludedGameServices = append(excludedGameServices, assignment.GameService)
			if len(excludedGameServices) >= 3 {
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
	// 仅识别调用尚未送达业务处理器的基础设施错误。
	return errors.Is(err, errs.ErrRPCNoRoute) || errors.Is(err, errs.ErrServiceNotReady) ||
		errors.Is(err, errs.ErrServiceStopping) || errors.Is(err, errs.ErrServiceStopped) ||
		errors.Is(err, errs.ErrServiceRetired) || errors.Is(err, errs.ErrServiceQueueFull)
}

func waitRetry(ctx context.Context) error {
	// 使用可取消的一秒退避，始终服从登录总时限。
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-timer.C:
		return nil
	}
}

func (module *GatewayClientModule) onClose(_ context.Context, session network.Session, _ error) {
	// 忽略已替换或未登记的会话关闭事件。
	current := module.connections[session.ID()]
	if current == nil || current.session != session {
		return
	}
	// 先删除本地状态，再通知 GameService 回收在线连接。
	delete(module.connections, session.ID())
	if current.state == stateOnline {
		module.Logger().Debug(
			"Gateway 已登录客户端断开",
			log.String("connection_id", string(session.ID())),
			log.Int64("show_area_id", current.showAreaID),
			log.Int64("real_area_id", current.realAreaID),
			log.String("game_service_node", current.gameService.NodeID),
		)
		_ = module.dependencies.GameServices.PlayerDisconnected(current.gameService, string(session.ID()))
	}
}

// SendClientMessage 编码并写入当前连接；连接不存在时幂等成功。
func (module *GatewayClientModule) SendClientMessage(request rpcapi.SendClientMessageRequest) error {
	// 连接已断开时下行请求幂等成功。
	current := module.connections[network.SessionID(request.GatewayConnectionID)]
	if current == nil {
		return nil
	}
	if request.Message.CloseAfterWrite {
		// 仅支持原子化最终写入的 Session 执行写后关闭。
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
	// 常规下行直接复用统一消息编码路径。
	return module.send(current, request.Message)
}

// CloseClientConnection 幂等关闭当前 Node 上的指定连接。
func (module *GatewayClientModule) CloseClientConnection(connectionID string) {
	// 存在连接时交由网络层触发幂等关闭流程。
	if current := module.connections[network.SessionID(connectionID)]; current != nil {
		current.session.Close(errs.ErrTransportClosed)
	}
}

func (module *GatewayClientModule) replyLoginError(current *connection, sequence uint32, code commonpb.ErrorCode) error {
	// 使用登录响应消息返回客户端可见错误码。
	return module.send(current, rpcapi.ClientMessage{
		MessageID: commonpb.MessageID_LoginPlayerRes, Sequence: sequence, ErrorCode: code,
	})
}

func (module *GatewayClientModule) send(current *connection, message rpcapi.ClientMessage) error {
	// 先完成协议编码，再写入底层网络会话。
	payload, err := encodeClientMessage(message)
	if err != nil {
		return fmt.Errorf("编码客户端消息: %w", err)
	}
	return current.session.Send(payload)
}
