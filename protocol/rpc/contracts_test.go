package rpcapi

import (
	"context"
	"testing"

	commonpb "origingame/protocol/common"
)

type dbServiceContractStub struct{}

func (dbServiceContractStub) ExecuteMongo(context.Context, MongoRequest) (MongoResult, error) {
	return MongoResult{}, nil
}

func (dbServiceContractStub) ExecuteRedis(context.Context, RedisRequest) (RedisResult, error) {
	return RedisResult{}, nil
}

type gameServiceContractStub struct{}

func (gameServiceContractStub) LoginPlayer(context.Context, LoginPlayerRequest) (*commonpb.LoginPlayerResult, error) {
	return nil, nil
}

func (gameServiceContractStub) HandlePlayerMessage(context.Context, PlayerMessageRequest) error {
	return nil
}

func (gameServiceContractStub) PlayerDisconnected(context.Context, PlayerDisconnectedRequest) error {
	return nil
}

type gatewayServiceContractStub struct{}

func (gatewayServiceContractStub) SendClientMessage(context.Context, SendClientMessageRequest) error {
	return nil
}

func (gatewayServiceContractStub) CloseClientConnection(context.Context, CloseClientConnectionRequest) error {
	return nil
}

type robotServiceContractStub struct{}

func (robotServiceContractStub) ListScenarios(context.Context, ListRobotScenariosRequest) (ListRobotScenariosResponse, error) {
	return ListRobotScenariosResponse{}, nil
}

func (robotServiceContractStub) StartRun(context.Context, StartRobotRunRequest) (RobotRunSnapshot, error) {
	return RobotRunSnapshot{}, nil
}

func (robotServiceContractStub) StopRun(context.Context, StopRobotRunRequest) (RobotRunSnapshot, error) {
	return RobotRunSnapshot{}, nil
}

func (robotServiceContractStub) GetRun(context.Context, GetRobotRunRequest) (RobotRunSnapshot, error) {
	return RobotRunSnapshot{}, nil
}

func TestServiceContractsHaveConfirmedMethodSets(t *testing.T) {
	var _ DBService = dbServiceContractStub{}
	var _ GameService = gameServiceContractStub{}
	var _ GatewayService = gatewayServiceContractStub{}
	var _ RobotService = robotServiceContractStub{}
}

func TestProtocolEnumsReserveZeroForUnspecified(t *testing.T) {
	values := []int32{
		int32(MongoExecuteModeUnspecified),
		int32(MongoOperationKindUnspecified),
		int32(MongoUpdateKindUnspecified),
		int32(MongoReturnDocumentUnspecified),
		int32(MongoOperationStatusUnspecified),
		int32(MongoFailureKindUnspecified),
		int32(RedisExecuteModeUnspecified),
		int32(RedisValueKindUnspecified),
		int32(RedisCommandStatusUnspecified),
		int32(RedisFailureKindUnspecified),
	}
	for index, value := range values {
		if value != 0 {
			t.Fatalf("枚举 %d 的 Unspecified 不是 0: %d", index, value)
		}
	}
}

func TestRedisValueUsesFlatNodeGraph(t *testing.T) {
	value := RedisValue{
		RootIndex: 0,
		Nodes: []RedisValueNode{
			{Kind: RedisValueKindArray, Children: []uint32{1}},
			{Kind: RedisValueKindBytes, Bytes: []byte("value")},
		},
	}
	if value.Nodes[value.Nodes[value.RootIndex].Children[0]].Kind != RedisValueKindBytes {
		t.Fatalf("Redis 扁平结果图异常: %+v", value)
	}
}
