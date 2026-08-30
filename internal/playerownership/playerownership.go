// Package playerownership 封装 Gateway 与 GameService 共享的在线玩家归属协议。
package playerownership

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"origingame/internal/dbexecutor"
	"origingame/internal/redisscripts"
	rpcapi "origingame/protocol/rpc"
)

const (
	gameServiceLeaseTTL     = 15 * time.Second
	assigningTimeout        = 10 * time.Second
	loadingTimeout          = 35 * time.Second
	playerOwnershipTTL      = 60 * time.Second
	leavingTTL              = 5 * time.Second
	assignmentScanLimit     = 16
	maxExcludedGameServices = 3
	maxRenewOwnerships      = 256
)

// PlayerOwnershipStore 不保存权威状态，只负责通用 DBService 请求与在线玩家归属协议之间的转换。
type PlayerOwnershipStore struct{ executor dbexecutor.RedisExecutor }

// NewPlayerOwnershipStore 创建玩家归属协议适配器。
func NewPlayerOwnershipStore(executor dbexecutor.RedisExecutor) *PlayerOwnershipStore {
	return &PlayerOwnershipStore{executor: executor}
}

// Player 标识账号在一个显示区服中的稳定玩家。
type Player struct {
	AccountID  string
	ShowAreaID int64
	RealAreaID int64
}

// GameServiceInstance 精确标识一次 GameService 进程实例。
type GameServiceInstance struct {
	ServiceName   string
	NodeID        string
	NodeSessionID string
}

// GameServiceRegistration 是 GameService 对登录分配公开的最小实例信息。
type GameServiceRegistration struct {
	RealAreaID  int64
	GameService GameServiceInstance
	MaxPlayers  int64
}

// AssignmentDecision 是 AssignOrGet 的稳定原子决策。
type AssignmentDecision int64

const (
	AssignmentDecisionUnspecified AssignmentDecision = iota
	AssignmentDecisionExisting
	AssignmentDecisionAssigned
	AssignmentDecisionWait
	AssignmentDecisionNoCapacity
)

// AssignmentResult 返回精确 GameService 及当前归属状态。
type AssignmentResult struct {
	Decision    AssignmentDecision
	GameService GameServiceInstance
	State       string
}

// AssignRequest 是 Gateway 查询或分配玩家归属所需的全部可信参数。
type AssignRequest struct {
	Player               Player
	GatewayConnectionID  string
	ExcludedGameServices []GameServiceInstance
}

// RegisterGameService 原子登记或刷新 READY 实例和15秒租约。
func (store *PlayerOwnershipStore) RegisterGameService(ctx context.Context, registration GameServiceRegistration) error {
	if err := validateRegistration(registration); err != nil {
		return err
	}
	info := instanceKey(registration.RealAreaID, registration.GameService)
	result, err := store.run(ctx, info, redisscripts.RegisterGameServiceID,
		[]string{info, info + ":lease", candidatesKey(registration.RealAreaID)},
		registration.GameService.ServiceName, registration.GameService.NodeID, registration.GameService.NodeSessionID,
		strconv.FormatInt(registration.MaxPlayers, 10), milliseconds(gameServiceLeaseTTL))
	return requireOne(result, err)
}

// SetGameServiceDraining 条件摘除当前实例，旧进程不能修改新实例。
func (store *PlayerOwnershipStore) SetGameServiceDraining(ctx context.Context, realAreaID int64, gameService GameServiceInstance) (bool, error) {
	if err := validateInstance(realAreaID, gameService); err != nil {
		return false, err
	}
	info := instanceKey(realAreaID, gameService)
	result, err := store.run(ctx, info, redisscripts.SetGameServiceDrainingID,
		[]string{info, candidatesKey(realAreaID)}, gameService.NodeID, gameService.NodeSessionID)
	return scriptBoolean(result, err)
}

// AssignOrGet 返回已有归属，或原子预占当前最低负载的可用实例。
func (store *PlayerOwnershipStore) AssignOrGet(ctx context.Context, request AssignRequest) (AssignmentResult, error) {
	if err := validatePlayer(request.Player); err != nil || strings.TrimSpace(request.GatewayConnectionID) == "" ||
		len(request.ExcludedGameServices) > maxExcludedGameServices {
		return AssignmentResult{}, errors.New("玩家分配参数无效")
	}
	args := []string{
		request.GatewayConnectionID,
		request.Player.AccountID,
		strconv.FormatInt(request.Player.ShowAreaID, 10),
		strconv.FormatInt(request.Player.RealAreaID, 10),
		milliseconds(playerOwnershipTTL), milliseconds(assigningTimeout), milliseconds(loadingTimeout),
		strconv.Itoa(assignmentScanLimit),
	}
	for _, excluded := range request.ExcludedGameServices {
		if err := validateInstance(request.Player.RealAreaID, excluded); err != nil {
			return AssignmentResult{}, err
		}
		args = append(args, instanceKey(request.Player.RealAreaID, excluded))
	}
	key := PlayerKey(request.Player.AccountID, request.Player.ShowAreaID)
	ownership := ownershipKey(request.Player)
	result, err := store.run(ctx, key, redisscripts.AssignOrGetPlayerID,
		[]string{ownership, candidatesKey(request.Player.RealAreaID)}, args...)
	if err != nil {
		return AssignmentResult{}, err
	}
	values, err := scriptArray(result, 5)
	if err != nil {
		return AssignmentResult{}, err
	}
	decision := AssignmentDecision(values[0].Integer)
	if values[0].Kind != rpcapi.RedisValueKindInteger || decision < AssignmentDecisionExisting ||
		decision > AssignmentDecisionNoCapacity {
		return AssignmentResult{}, errors.New("玩家分配返回决策无效")
	}
	assignment := AssignmentResult{Decision: decision}
	assignment.GameService.ServiceName, err = bytesText(values[1])
	if err != nil {
		return AssignmentResult{}, err
	}
	assignment.GameService.NodeID, err = bytesText(values[2])
	if err != nil {
		return AssignmentResult{}, err
	}
	assignment.GameService.NodeSessionID, err = bytesText(values[3])
	if err != nil {
		return AssignmentResult{}, err
	}
	assignment.State, err = bytesText(values[4])
	if err != nil {
		return AssignmentResult{}, err
	}
	if (decision == AssignmentDecisionExisting || decision == AssignmentDecisionAssigned) &&
		validateInstance(request.Player.RealAreaID, assignment.GameService) != nil {
		return AssignmentResult{}, errors.New("玩家分配返回实例无效")
	}
	return assignment, nil
}

// BeginPlayerLoad 将当前连接的 ASSIGNING 归属条件推进到 LOADING。
func (store *PlayerOwnershipStore) BeginPlayerLoad(ctx context.Context, player Player, gameService GameServiceInstance, connectionID string) (bool, error) {
	return store.playerTransition(ctx, player, gameService, redisscripts.BeginPlayerLoadID,
		[]string{ownershipKey(player), instanceKey(player.RealAreaID, gameService), instanceKey(player.RealAreaID, gameService) + ":lease"},
		gameService.NodeID, gameService.NodeSessionID, connectionID, milliseconds(playerOwnershipTTL))
}

// CompletePlayerLogin 将 LOADING 或 RESIDENT 归属推进到 ONLINE。
func (store *PlayerOwnershipStore) CompletePlayerLogin(ctx context.Context, player Player, gameService GameServiceInstance, connectionID string) (bool, error) {
	return store.playerTransition(ctx, player, gameService, redisscripts.CompletePlayerLoginID,
		transitionKeys(player, gameService), gameService.NodeID, gameService.NodeSessionID, connectionID,
		milliseconds(playerOwnershipTTL))
}

// ReleasePlayerLoad 条件回滚尚未完成的加载预占。
func (store *PlayerOwnershipStore) ReleasePlayerLoad(ctx context.Context, player Player, gameService GameServiceInstance, connectionID string) (bool, error) {
	return store.playerTransition(ctx, player, gameService, redisscripts.ReleasePlayerLoadID,
		transitionKeys(player, gameService), gameService.NodeID, gameService.NodeSessionID, connectionID)
}

// MarkPlayerResident 将当前实例的 ONLINE 归属转为断线驻留。
func (store *PlayerOwnershipStore) MarkPlayerResident(ctx context.Context, player Player, gameService GameServiceInstance) (bool, error) {
	return store.playerTransition(ctx, player, gameService, redisscripts.MarkPlayerResidentID,
		transitionKeys(player, gameService), gameService.NodeID, gameService.NodeSessionID, milliseconds(playerOwnershipTTL))
}

// BeginPlayerRelease 将到期 RESIDENT 归属改为 LEAVING，并保留5秒隔离窗口。
func (store *PlayerOwnershipStore) BeginPlayerRelease(ctx context.Context, player Player, gameService GameServiceInstance) (bool, error) {
	return store.playerTransition(ctx, player, gameService, redisscripts.BeginPlayerReleaseID,
		transitionKeys(player, gameService), gameService.NodeID, gameService.NodeSessionID, milliseconds(leavingTTL))
}

// RenewPlayerOwnerships 条件续租当前实例拥有的最多256条在线或驻留归属。
func (store *PlayerOwnershipStore) RenewPlayerOwnerships(ctx context.Context, realAreaID int64, gameService GameServiceInstance, players []Player) (int64, error) {
	if err := validateInstance(realAreaID, gameService); err != nil || len(players) == 0 || len(players) > maxRenewOwnerships {
		return 0, errors.New("玩家归属续租参数无效")
	}
	keys := make([]string, len(players))
	for index, player := range players {
		if err := validatePlayer(player); err != nil || player.RealAreaID != realAreaID {
			return 0, errors.New("玩家归属续租包含无效玩家")
		}
		keys[index] = ownershipKey(player)
	}
	dispatchKey := instanceKey(realAreaID, gameService)
	result, err := store.run(ctx, dispatchKey, redisscripts.RenewPlayerOwnershipsID, keys,
		gameService.NodeID, gameService.NodeSessionID, milliseconds(playerOwnershipTTL))
	return scriptInteger(result, err)
}

func (store *PlayerOwnershipStore) playerTransition(
	ctx context.Context,
	player Player,
	gameService GameServiceInstance,
	scriptID string,
	keys []string,
	args ...string,
) (bool, error) {
	if err := validatePlayer(player); err != nil {
		return false, err
	}
	if err := validateInstance(player.RealAreaID, gameService); err != nil {
		return false, err
	}
	result, err := store.run(ctx, PlayerKey(player.AccountID, player.ShowAreaID), scriptID, keys, args...)
	return scriptBoolean(result, err)
}

func (store *PlayerOwnershipStore) run(
	ctx context.Context,
	dispatchKey string,
	scriptID string,
	keys []string,
	args ...string,
) (rpcapi.RedisResult, error) {
	if store == nil || store.executor == nil {
		return rpcapi.RedisResult{}, errors.New("playerownership.PlayerOwnershipStore 未初始化")
	}
	encoded := make([][]byte, len(args))
	for index := range args {
		encoded[index] = []byte(args[index])
	}
	request := rpcapi.RedisRequest{
		DispatchKey: dispatchKey,
		ExecuteMode: rpcapi.RedisExecuteModeScript,
		Script:      &rpcapi.RedisScriptCall{ID: scriptID, Keys: keys, Args: encoded},
	}
	return store.executor.ExecuteRedis(ctx, dispatchKey, request)
}

func transitionKeys(player Player, gameService GameServiceInstance) []string {
	return []string{
		ownershipKey(player),
		instanceKey(player.RealAreaID, gameService),
		candidatesKey(player.RealAreaID),
	}
}

// PlayerKey 是 AccountID 与 ShowAreaID 组成的稳定玩家业务 Key。
func PlayerKey(accountID string, showAreaID int64) string {
	return strings.TrimSpace(accountID) + ":" + strconv.FormatInt(showAreaID, 10)
}

func ownershipKey(player Player) string {
	return areaPrefix(player.RealAreaID) + ":player-route:" + strconv.FormatInt(player.ShowAreaID, 10) + ":" + player.AccountID
}

func candidatesKey(realAreaID int64) string { return areaPrefix(realAreaID) + ":game-services" }

func instanceKey(realAreaID int64, gameService GameServiceInstance) string {
	return areaPrefix(realAreaID) + ":game-service:" + gameService.NodeID + ":" + gameService.NodeSessionID
}

func areaPrefix(realAreaID int64) string { return "{area:" + strconv.FormatInt(realAreaID, 10) + "}" }

func validatePlayer(player Player) error {
	if strings.TrimSpace(player.AccountID) == "" || player.ShowAreaID <= 0 || player.RealAreaID <= 0 ||
		strings.ContainsAny(player.AccountID, "{}:") {
		return errors.New("玩家归属身份无效")
	}
	return nil
}

func validateInstance(realAreaID int64, gameService GameServiceInstance) error {
	if realAreaID <= 0 || strings.TrimSpace(gameService.ServiceName) == "" || strings.TrimSpace(gameService.NodeID) == "" ||
		strings.TrimSpace(gameService.NodeSessionID) == "" || strings.ContainsAny(gameService.NodeID+gameService.NodeSessionID, "{}:") {
		return errors.New("GameService 实例身份无效")
	}
	return nil
}

func validateRegistration(registration GameServiceRegistration) error {
	if registration.MaxPlayers <= 0 {
		return errors.New("GameService player_capacity 必须为正数")
	}
	return validateInstance(registration.RealAreaID, registration.GameService)
}

func requireOne(result rpcapi.RedisResult, err error) error {
	value, err := scriptInteger(result, err)
	if err != nil {
		return err
	}
	if value != 1 {
		return errors.New("Redis Script 未完成预期修改")
	}
	return nil
}

func scriptBoolean(result rpcapi.RedisResult, err error) (bool, error) {
	value, err := scriptInteger(result, err)
	if err != nil {
		return false, err
	}
	if value != 0 && value != 1 {
		return false, errors.New("Redis Script 返回布尔值无效")
	}
	return value == 1, nil
}

func scriptInteger(result rpcapi.RedisResult, err error) (int64, error) {
	if err != nil {
		return 0, err
	}
	if result.Failure != nil || len(result.Results) != 1 ||
		result.Results[0].Status != rpcapi.RedisCommandStatusSucceeded {
		return 0, errors.New("Redis Script 执行失败")
	}
	value := result.Results[0].Value
	if int(value.RootIndex) >= len(value.Nodes) || value.Nodes[value.RootIndex].Kind != rpcapi.RedisValueKindInteger {
		return 0, errors.New("Redis Script 返回整数结构无效")
	}
	return value.Nodes[value.RootIndex].Integer, nil
}

func scriptArray(result rpcapi.RedisResult, length int) ([]rpcapi.RedisValueNode, error) {
	if result.Failure != nil || len(result.Results) != 1 ||
		result.Results[0].Status != rpcapi.RedisCommandStatusSucceeded {
		return nil, errors.New("Redis Script 执行失败")
	}
	value := result.Results[0].Value
	if int(value.RootIndex) >= len(value.Nodes) {
		return nil, errors.New("Redis Script 返回数组结构无效")
	}
	root := value.Nodes[value.RootIndex]
	if root.Kind != rpcapi.RedisValueKindArray || len(root.Children) != length {
		return nil, errors.New("Redis Script 返回数组长度无效")
	}
	resultNodes := make([]rpcapi.RedisValueNode, length)
	for index, child := range root.Children {
		if int(child) >= len(value.Nodes) {
			return nil, errors.New("Redis Script 返回数组索引无效")
		}
		resultNodes[index] = value.Nodes[child]
	}
	return resultNodes, nil
}

func bytesText(node rpcapi.RedisValueNode) (string, error) {
	if node.Kind != rpcapi.RedisValueKindBytes {
		return "", fmt.Errorf("Redis Script 返回文本类型无效: kind=%d", node.Kind)
	}
	return string(node.Bytes), nil
}

func milliseconds(duration time.Duration) string {
	return strconv.FormatInt(duration.Milliseconds(), 10)
}
