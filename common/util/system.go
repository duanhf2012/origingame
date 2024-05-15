package util

import (
	"errors"
	"fmt"
	"github.com/duanhf2012/origin/v2/cluster"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/rpc"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"

	"strconv"
	"strings"
	"time"
)

const (
	TimeLayout    = "2006-01-02 15:04:05"
	IAPKey        = "DY@!iAp&pURchASE$xoWqFGW27(KD!dP"
	TokenKey      = "!gmn$7DYu&h3bDcSfXd)fLQWhvC&dLxR" //32位Key
	LoginKey      = "kxZdFA@#fWCNaL^UIbF$YR(oD@qwfQBw!cOeJeb*ASz?@DEdZjV?KsoZ)P&N^FYb"
	TokenValidity = 24 * 3600 //10分钟

	AccDBService  = "AccDBService"
	DBService     = "DBService"
	GateService   = "GateService"
	CenterService = "CenterService"

	AreaDBRequest = "DBService.RPC_DBRequest"
	AccDBRequest  = "AccDBService.RPC_DBRequest"

	IsEmpty = ""
)

// 原始Rpc的MethodId定义
const (
	RawRpcMsgDispatch uint32 = 1 //其他服(GameService或其他)->GateService->Client,转发消息
	RawRpcCloseClient uint32 = 2 //其他服(GameService或其他)->GateService->Client,断开与Client连接
	RawRpcOnRecv      uint32 = 3 //Client->GateService->其他服(GameService或其他),转发消息
	RawRpcOnClose     uint32 = 4 //Client->GateService->其他服(GameService或其他),转发Client连接断开事件
)

var Empty struct{}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewUnionId() string {
	return primitive.NewObjectID().Hex()
}

// GetMasterCenterNodeId 获取主centerservice的nodeId
func GetMasterCenterNodeId() string {
	clientList := make([]*rpc.Client, 0, 16)
	err, clientList := cluster.GetCluster().GetNodeIdByService(CenterService, clientList, true)
	if err != nil || len(clientList) == 0 {
		return ""
	}

	minId := ""
	for i := 0; i < len(clientList); i++ {
		if clientList[i] != nil {
			if minId == "" || clientList[i].GetTargetNodeId() < minId {
				minId = clientList[i].GetTargetNodeId()
			}
		}
	}

	return minId
}

// GetBestNodeId 获取最优服务的nodeId
func GetBestNodeId(serviceMethod string, key uint64) string {
	clientList := make([]*rpc.Client, 0, 16)
	err, clientList := cluster.GetRpcClient("", serviceMethod, true, clientList)
	if err != nil || len(clientList) == 0 {
		return ""
	}

	return clientList[key%uint64(len(clientList))].GetTargetNodeId()
}

func GetBestNodeIdById(serviceMethod string, key string) string {
	clientList := make([]*rpc.Client, 0, 16)
	err, clientList := cluster.GetRpcClient("", serviceMethod, true, clientList)
	if err != nil || len(clientList) == 0 {
		return ""
	}

	return clientList[uint64(HashString2Number(key))%uint64(len(clientList))].GetTargetNodeId()
}

func GetAllNodeId(serviceMethod string, idList []string) []string {
	clientList := make([]*rpc.Client, 0, 16)
	err, clientList := cluster.GetRpcClient("", serviceMethod, true, clientList[:])
	if err != nil || len(clientList) == 0 {
		return idList
	}

	for _, clientItem := range clientList {
		if clientItem == nil {
			break
		}
		idList = append(idList, clientItem.GetTargetNodeId())
	}
	return idList
}

func EncryptToken(platId string) (string, error) {
	origToken := fmt.Sprintf("%d#%s#%d", rand.Int(), platId, time.Now().Unix()+TokenValidity)
	return SpecialAesEncrypt(origToken, TokenKey)
}

func DecryptToken(token string) (string, error) {
	if token == "" {
		return "", errors.New("token is empty")
	}
	//origToken := fmt.Sprintf("%d_%d",userId,time.Now().Unix())
	origToken, err := SpecialAesDecrypt(token, TokenKey)
	if err != nil {
		return "", err
	}

	tokenKey := strings.Split(origToken, "#")
	if len(tokenKey) != 3 {
		return "", fmt.Errorf("token is error %s", token)
	}
	platId := strings.TrimSpace(tokenKey[1])

	tokenValidity, err := strconv.ParseInt(tokenKey[2], 10, 64)
	if err != nil {
		return "", err
	}

	if time.Now().Unix() >= tokenValidity {
		validityTime := time.Unix(tokenValidity, 0)
		log.Error("token timeout", log.String("now", time.Now().Format(TimeLayout)), log.String("validity", validityTime.Format(TimeLayout)))
		return platId, fmt.Errorf("token expired")
	}
	return platId, nil
}
