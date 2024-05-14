package centerservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"origingame/common/collect"
	"origingame/common/db"
	"origingame/common/proto/msg"
	"origingame/common/proto/rpc"
	"origingame/common/util"
	"time"

	"github.com/duanhf2012/origin/v2/util/coroutine"

	"github.com/duanhf2012/origin/v2/cluster"
	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/node"
	rpcHelper "github.com/duanhf2012/origin/v2/rpc"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/duanhf2012/origin/v2/util/queue"
	"github.com/duanhf2012/origin/v2/util/timer"
	"go.mongodb.org/mongo-driver/bson"
	"origingame/common/env"
)

const LoginingTimeout = time.Minute * 10
const OutingTimeout = time.Minute * 1
const PreNodeCount = 5

type UserInfo struct {
	Status rpc.LoginStatus //0正在登录，1登录成功
	//Token             string
	GameServiceNodeId string
	GSName            string
}

type BalanceInfo struct {
	NodeId      string
	GSName      string
	item        *queue.Item
	refreshTime time.Time //用于健康检查
	Weight      int32     //表示GS那边同步过来的人数
	SelectNum   uint64    //选择的次数
}

type CenterService struct {
	service.Service

	mapUserInfo map[string]UserInfo
	mapLogining map[string]time.Time
	mapOuting   map[string]time.Time

	mapBalance    map[string]map[string]*BalanceInfo //map[nodeId]map[gsName]*BalanceInfo
	priorityQueue queue.PriorityQueue                //两组
	//SelectServerGroup int32

	DBCharacterNum    int32 //DB中角色数量
	NewCharacterNum   int32 //启服后新建角色数量
	dirtyNewCharacter bool

	AreaId                     int32
	GSMaxPlayerNumbersPriority int
	//startTime                  int64

	//数量限制
	loadTimerId      uint64
	reloadTickerId   uint64
	isLoadMaxFinish  bool
	maxRegisterCount int32 //注册数量上限
	maxOnlineCount   int32 //在线数量上限

	loadShowAreaTickerId   uint64
	reloadShowAreaTickerId uint64
	ShowAreaMap            map[int]*collect.CShowAreaInfo // 显示区服信息

	//关闭的消息ID
	mapCloseMsgId map[msg.MsgType]struct{}
}

func init() {
	node.Setup(&CenterService{})
}

func (cs *CenterService) OnInit() error {
	log.SInfo("release is ", env.IsRelease)

	cs.mapUserInfo = make(map[string]UserInfo, 100000)
	cs.mapBalance = make(map[string]map[string]*BalanceInfo, 1024)
	cs.mapLogining = make(map[string]time.Time, 100)
	cs.mapOuting = make(map[string]time.Time, 100)
	cs.NewTicker(time.Minute*1, cs.CheckLoginTimeout)
	cs.NewTicker(time.Minute*2, cs.SyncAreaBalance)
	cs.priorityQueue.Init(5000)

	//性能监控
	cs.OpenProfiler()
	cs.GetProfiler().SetOverTime(time.Millisecond * 100)
	cs.GetProfiler().SetMaxOverTime(time.Second * 10)

	globalCfg := cluster.GetCluster().GetGlobalCfg()
	if globalCfg == nil {
		return fmt.Errorf("Canot find Global from config.")
	}
	mapGlobal, ok := globalCfg.(map[string]interface{})
	if ok == false {
		return fmt.Errorf("Canot find Global from config.")
	}
	areaId, ok := mapGlobal["AreaId"]
	if ok == false {
		return fmt.Errorf("Canot find Global.AreaId from config.")
	}

	cs.AreaId = int32(areaId.(float64))

	mapAreaInfo, ok := cs.GetServiceCfg().(map[string]interface{})
	if mapAreaInfo == nil || ok == false {
		return errors.New("cannot find AreaId config")
	}

	//GSMaxPlayerNumbers
	iGSMaxPlayerNumbers, ok := mapAreaInfo["GSMaxPlayerNumbers"]
	if ok == false {
		return errors.New("cannot find AreaId config")
	}

	gSMaxPlayerNumbers, ok := iGSMaxPlayerNumbers.(float64)
	if ok == false || gSMaxPlayerNumbers <= 0 {
		return errors.New("AreaId config is error")
	}

	cs.GSMaxPlayerNumbersPriority = int(gSMaxPlayerNumbers)

	return nil
}

func (cs *CenterService) OnStart() {
	cs.AfterFunc(time.Second*4, cs.CountUserInfoBy)

	cs.tryLoadAreaInfoFromDB()
	cs.tickerSyncAreaInfoFromDB()

	cs.tickerLoadShowArea()
}

func (cs *CenterService) tickerLoadShowArea() {
	cs.SafeAfterFunc(&cs.loadShowAreaTickerId, time.Second*2, nil, cs.timerLoadShowAreaInfo)
	cs.SafeNewTicker(&cs.reloadShowAreaTickerId, time.Minute*2, nil, cs.tickerLoadShowAreaInfo)
}
func (cs *CenterService) timerLoadShowAreaInfo(timerId uint64, i interface{}) {
	cs.loadShowAreaInfoFromDB()
}
func (cs *CenterService) tickerLoadShowAreaInfo(timerId uint64, i interface{}) {
	cs.loadShowAreaInfoFromDB()
}

func (cs *CenterService) tryLoadAreaInfoFromDB() {
	cs.SafeAfterFunc(&cs.loadTimerId, time.Second*5, nil, cs.loadAreaInfo)
}

func (cs *CenterService) loadAreaInfo(timerId uint64, i interface{}) {
	cs.loadTimerId = 0
	cs.loadAreaInfoFromDB(true)
}

func (cs *CenterService) tickerSyncAreaInfoFromDB() {
	cs.SafeNewTicker(&cs.reloadTickerId, time.Minute*5, nil, cs.tickerLoadAreaInfo)
}

func (cs *CenterService) tickerLoadAreaInfo(timerId uint64, i interface{}) {
	cs.loadAreaInfoFromDB(false)
}

func (cs *CenterService) CountUserInfoBy(t *timer.Timer) {
	//查询数据库行数
	var req db.DBControllerReq
	req.CollectName = collect.UserInfoCollectName
	req.Type = db.OptType_Count

	err := cs.AsyncCall(util.AreaDBRequest, &req, func(res *db.DBControllerRet, err error) {
		if err != nil || res.MatchedCount == -1 {
			cs.AfterFunc(time.Second*4, cs.CountUserInfoBy)
			return
		}

		cs.DBCharacterNum = res.MatchedCount
		cs.dirtyNewCharacter = true
	})

	if err != nil {
		cs.AfterFunc(time.Second*4, cs.CountUserInfoBy)
	}
}

// 定时检查x分钟内，状态依然没有成功的玩家
func (cs *CenterService) CheckLoginTimeout(timer *timer.Ticker) {
	now := time.Now()
	for uId, v := range cs.mapLogining {
		if now.Sub(v) > LoginingTimeout {
			delete(cs.mapLogining, uId)
			v, ok := cs.mapUserInfo[uId]
			if ok == true && v.Status == rpc.LoginStatus_Logined {
				//这种状态不应该发生
				log.SError("mapUserInfo status is ", v.Status, ",but the userid still exists mapLogining")
			} else {
				cs.delUserInfo(uId)
				//delete(cs.mapUserInfo, uId)
			}
		}
	}
	for uId, v := range cs.mapOuting {
		if now.Sub(v) > OutingTimeout {
			delete(cs.mapOuting, uId)
			v, ok := cs.mapUserInfo[uId]
			if ok == true && v.Status == rpc.LoginStatus_LoginOuting {
				//这种状态不应该发生
				log.SError("mapUserInfo status is ", v.Status, ",but the userid still exists mapOuting")
			} else {
				cs.delUserInfo(uId)
				//delete(cs.mapUserInfo, uId)
			}
		}
	}
}

func (cs *CenterService) refreshBalance(gameServiceNodeId string, gsName string, delta int) {
	log.SInfo(" gameServiceNodeId:", gameServiceNodeId, " gsName：", gsName, " delta:", delta)
	mapNodeBalance := cs.mapBalance[gameServiceNodeId]
	if mapNodeBalance == nil {
		log.SError("cannot find gameServiceNodeId ", gameServiceNodeId)
		return
	}

	balanceInfo, ok := mapNodeBalance[gsName]
	if ok == false || balanceInfo == nil {
		log.SError("cannot find serviceId ", gsName)
		return
	}

	//刷新负载，-1降低优先级  1是增加优先级
	balanceInfo.item.Priority += delta
	if delta < 0 {
		balanceInfo.SelectNum += 1
	}

	cs.priorityQueue.Update(balanceInfo.item, balanceInfo, balanceInfo.item.Priority)
}

func (cs *CenterService) addUserInfo(uId string, status rpc.LoginStatus, gameServiceNodeId string, gsName string, bUpdate bool) {
	if bUpdate == false {
		cs.refreshBalance(gameServiceNodeId, gsName, -1)
	}

	cs.mapUserInfo[uId] = UserInfo{Status: status, GameServiceNodeId: gameServiceNodeId, GSName: gsName}
}

func (cs *CenterService) delUserInfo(uId string) {
	userInfo, ok := cs.mapUserInfo[uId]
	if ok == false {
		log.SError("cannot find userInfo Uid", uId)
		return
	}

	cs.refreshBalance(userInfo.GameServiceNodeId, userInfo.GSName, 1)
	delete(cs.mapUserInfo, uId)
}

func (cs *CenterService) SyncAreaBalance(timer *timer.Ticker) {
	if cs.dirtyNewCharacter == false {
		return
	}

	cs.dirtyNewCharacter = false

	//同步到数据库
	var req db.DBControllerReq
	data := bson.M{"registerCount": cs.NewCharacterNum + cs.DBCharacterNum}
	err := db.MakeUpsetId("RealAreaInfo", cs.AreaId, data, fmt.Sprintf("%d", cs.AreaId), &req)
	if err != nil {
		log.SError("make upsetid fail ", err.Error())
		return
	}

	nodeId := util.GetBestNodeId(util.AccDBRequest, uint64(cs.AreaId))
	if nodeId == "" {
		log.SError("cannot find nodeId from rpcMethod ", util.AccDBRequest, ",areaId ", cs.AreaId, ",collectName ", req.CollectName)
		return
	}

	if cs.GoNode(nodeId, util.AccDBRequest, &req) != nil {
		log.SError("ExecDB fail:go node error :", err.Error(), ",userid ", cs.AreaId, ",collectName ", req.CollectName)
		return
	}
}

func (cs *CenterService) RemoveBalance(balanceInfo *BalanceInfo) {
	//从map中删除
	mapServiceBalance, ok := cs.mapBalance[balanceInfo.NodeId]
	if ok == false {
		log.SError("cannot find nodeid ", balanceInfo.NodeId)
		return
	}

	delete(mapServiceBalance, balanceInfo.GSName)
}

// 查找负载最低的gameServiceService
func (cs *CenterService) getBestGameServiceNodeId() (string, string) {
	now := time.Now()
	if len(cs.mapUserInfo) >= cs.GSMaxPlayerNumbersPriority {
		return "", ""
	}

	for {
		item := cs.priorityQueue.Pop()
		if item == nil {
			return "", ""
		}
		balanceInfo, ok := item.Value.(*BalanceInfo)
		if ok == false {
			log.SError("convert data error item.Value is not *BalanceInfo")
			continue
		}

		//30秒未刷新的取消分配
		if now.Sub(balanceInfo.refreshTime) > time.Second*30 {
			cs.RemoveBalance(balanceInfo)
			continue
		}

		//退休状态不再分配
		if cluster.GetCluster().IsNodeConnected(balanceInfo.NodeId) == false || cluster.GetCluster().IsNodeRetire(balanceInfo.NodeId) == true {
			cs.RemoveBalance(balanceInfo)
			continue
		}

		cs.priorityQueue.Push(item)
		return balanceInfo.NodeId, balanceInfo.GSName
	}

}

// GateService登陆验证token
func (cs *CenterService) RPC_Login(req *rpc.LoginGateCheckReq, res *rpc.LoginGateCheckRet) error {
	// 1.查找登陆缓存
	var bestGameNodeId string
	var bestGS string
	var userInfo UserInfo
	var ok bool
	userInfo, ok = cs.mapUserInfo[req.UserId]
	//2.已经存在，则继续上次的GameServiceNodeId
	if ok == true {
		bestGameNodeId = userInfo.GameServiceNodeId
		bestGS = userInfo.GSName
	} else {
		bestGameNodeId, bestGS = cs.getBestGameServiceNodeId()
	}

	//3.找到GameService NodeId才需要存储
	if bestGameNodeId != "" {
		cs.addUserInfo(req.UserId, rpc.LoginStatus_Logining, bestGameNodeId, bestGS, ok)
		cs.mapLogining[req.UserId] = time.Now()
	} else {
		log.SError("cannot find gameService")
	}

	//4.返回结果
	res.NodeId = bestGameNodeId
	res.GSName = bestGS
	res.Ret = 0

	return nil
}

// 登陆GameService与Player对象释放时
func (cs *CenterService) RPC_UpdateStatus(playerStatus *rpc.UpdatePlayerStatus) error {
	//只能这两种状态才能同步
	if playerStatus.Status == rpc.LoginStatus_Logining {
		return fmt.Errorf("status is error")
	}

	//如果是登出，删除相应的缓存
	//cs.mapLogining删除可能会遇到客户端正在请求登陆的情况，这样客户端登陆失败，重新走登陆流程即可
	delete(cs.mapLogining, playerStatus.UserId)
	switch playerStatus.Status {
	case rpc.LoginStatus_LoginOuted:
		delete(cs.mapOuting, playerStatus.UserId)
		cs.delUserInfo(playerStatus.UserId)
		return nil
	case rpc.LoginStatus_LoginOuting:
		cs.mapOuting[playerStatus.UserId] = time.Now()
	case rpc.LoginStatus_Logined:
		delete(cs.mapOuting, playerStatus.UserId)
	default:
		log.Error("status is error")
	}

	//如果是登陆成功，修改保存状态
	v, ok := cs.mapUserInfo[playerStatus.UserId]
	if ok == false {
		//不应该发生的情况
		log.SError("cannot find userid ", playerStatus.UserId)
		cs.addUserInfo(playerStatus.UserId, playerStatus.Status, playerStatus.NodeId, playerStatus.GSName, ok)
		//cs.mapUserInfo[playerStatus.UserId] = UserInfo{Status: rpc.LoginStatus_Logined, GameServiceNodeId: int(playerStatus.NodeId), ServiceId: playerStatus.ServiceId}
	} else {
		//v.Status = rpc.LoginStatus_Logined
		//不应该发生的情况
		if v.GameServiceNodeId != playerStatus.NodeId || v.GSName != playerStatus.GSName {
			log.SError("GameServiceNodeId ", v.GameServiceNodeId, " != NodeId ", playerStatus.NodeId, " or gsName ", v.GSName, " != playerStatus.GSName ", playerStatus.GSName)
		}

		//v.GameServiceNodeId = int(playerStatus.NodeId)
		//v.ServiceId = playerStatus.ServiceId
		cs.addUserInfo(playerStatus.UserId, playerStatus.Status, playerStatus.NodeId, playerStatus.GSName, ok)

		//cs.mapUserInfo[playerStatus.UserId] = v
	}

	return nil
}

// GameService服同步负载情况
func (cs *CenterService) RPC_UpdateBalance(balance *rpc.GameServiceBalance) error {
	nodeId := balance.NodeId
	mapServiceBalance, ok := cs.mapBalance[nodeId]
	if ok == false {
		mapServiceBalance = make(map[string]*BalanceInfo, 16)
		cs.mapBalance[nodeId] = mapServiceBalance
	}

	balanceInfo, ok := mapServiceBalance[balance.GSName]
	if ok == false {
		balanceInfo = &BalanceInfo{
			Weight:      balance.Weigh,
			refreshTime: time.Now(),
			NodeId:      balance.NodeId,
			GSName:      balance.GSName,
			item:        &queue.Item{},
		}
		mapServiceBalance[balance.GSName] = balanceInfo
		balanceInfo.item.Value = balanceInfo
		balanceInfo.item.Priority = int(-1 * balance.Weigh)
		cs.priorityQueue.Push(balanceInfo.item)
		return nil
	}

	balanceInfo.refreshTime = time.Now()
	balanceInfo.Weight = balance.Weigh

	return nil
}

// GameService服全同步所有玩家列表
func (cs *CenterService) RPC_UpdateUserList(playerList *rpc.UpdatePlayerList) error {
	//清理掉所有的player对象
	for uId, UInfo := range cs.mapUserInfo {
		if UInfo.GameServiceNodeId == playerList.NodeId {
			cs.delUserInfo(uId)
			delete(cs.mapLogining, uId)
		}
	}

	//重新同步
	nodeId := playerList.NodeId
	for _, uId := range playerList.UList {
		cs.addUserInfo(uId, rpc.LoginStatus_Logined, nodeId, playerList.GSName, false)
	}

	return nil
}

func (cs *CenterService) RPC_GetCenterServerBalance(req *struct{}, ret *rpc.MsgCenterServerBalance) error {
	ret.CenterBalanceList = make([]*rpc.CenterServerBalanceInfo, 0, 50)
	for _, mapBalance := range cs.mapBalance {
		for _, balanceInfo := range mapBalance {
			var info rpc.CenterServerBalanceInfo
			info.NodeId = balanceInfo.NodeId
			info.GSName = balanceInfo.GSName
			info.Weight = balanceInfo.Weight
			info.SelectNum = balanceInfo.SelectNum
			ret.CenterBalanceList = append(ret.CenterBalanceList, &info)
		}
	}
	return nil
}

func (cs *CenterService) RPC_GetGameServicePlayerInfo(responder rpcHelper.Responder, req *rpc.GameNodePlayerInfo) error {
	coroutine.Go(func() {
		var empty struct{}
		getNodePlayerInfoResult := rpc.GameNodePlayerInfoResult{}

		for _, gameServiceInfo := range req.GameServiceInfo {
			gameInfo := rpc.GetGameNodePlayerInfo{}
			rpcMethod := gameServiceInfo.GameServiceName + ".RPC_GetGameNodePlayerInfo"
			err := cs.CallNode(gameServiceInfo.NodeId, rpcMethod, &empty, &gameInfo)
			if err != nil {
				gameInfo.Error = err.Error()
			}
			gameInfo.NodeId = gameServiceInfo.NodeId
			gameInfo.ServiceName = gameServiceInfo.GameServiceName
			getNodePlayerInfoResult.ResultList = append(getNodePlayerInfoResult.ResultList, &gameInfo)
		}
		responder(&getNodePlayerInfoResult, rpcHelper.NilError)
	})
	return nil
}

func (cs *CenterService) RPC_GetGateServiceInfo(responder rpcHelper.Responder, req *rpc.GetGateServiceInfo) error {
	coroutine.Go(func() {
		var empty rpc.PlaceHolders
		getGateInfoResult := rpc.GetGateServiceInfoResult{}
		rpcMethod := "GateService.RPC_GetGateServiceConnect"

		for _, gateNodeId := range req.GateList {
			var gateInfo rpc.GateServiceConnect
			err := cs.CallNode(gateNodeId, rpcMethod, &empty, &gateInfo)
			if err != nil {
				gateInfo.Error = err.Error()
			}
			getGateInfoResult.GateResultList = append(getGateInfoResult.GateResultList, &gateInfo)
		}
		responder(&getGateInfoResult, rpcHelper.NilError)
	})
	return nil
}

func (cs *CenterService) RPC_GetAllServerNode(req *struct{}, ret *rpc.MsgAllServerNode) error {
	ret.NodeIdList = make([]string, 0, 50)
	for _, mapBalance := range cs.mapBalance {
		for _, balanceInfo := range mapBalance {
			ret.NodeIdList = append(ret.NodeIdList, balanceInfo.NodeId)
		}
	}
	return nil
}

func (cs *CenterService) RPC_CreateCharacterInfo(ret *struct{}, res *struct{}) error {
	cs.NewCharacterNum++
	cs.dirtyNewCharacter = true
	return nil
}

func (cs *CenterService) RPC_SaveAreaInfo(ret *rpc.SaveAreaInfo, res *struct{}) error {
	//记录角色所在区服信息
	var req db.DBControllerReq
	field := fmt.Sprintf("AreaHis.%d", cs.AreaId)
	data := bson.M{field: time.Now().Unix()}

	err := db.MakeUpsetId(collect.AccountCollectName, ret.PlatId, data, ret.PlatId, &req)
	if err != nil {
		log.SError("make upsetid fail ", err.Error())
		return nil
	}

	nodeId := util.GetBestNodeId(util.AccDBRequest, uint64(cs.AreaId))
	if nodeId == "" {
		log.SError("cannot find nodeId from rpcMethod ", util.AccDBRequest, ",areaId ", cs.AreaId, ",collectName ", req.CollectName)
		return errors.New("cannot find AccDBService.RPC_DBRequest")
	}

	if cs.GoNode(nodeId, util.AccDBRequest, &req) != nil {
		log.SError("ExecDB fail:go node error :", err.Error(), ",userid ", cs.AreaId, ",collectName ", req.CollectName)
		return errors.New("cannot find AccDBService.RPC_DBRequest")
	}

	return nil
}

// 获取serviceName所在的node，返回有多少个service
func (cs *CenterService) getServiceNode(nodeId string, serviceName string, mapNodeService map[string][]string) int {
	if mapNodeService == nil {
		return 0
	}

	retCount := 0
	mapServiceNodeId := cluster.GetNodeByServiceName(serviceName)
	if mapServiceNodeId == nil {
		return 0
	}

	if nodeId == "" {
		for serviceNodeId := range mapServiceNodeId {
			mapNodeService[serviceNodeId] = append(mapNodeService[serviceNodeId], serviceName)
			retCount++
		}
	} else {
		if _, ok := mapServiceNodeId[nodeId]; ok == true {
			mapNodeService[nodeId] = append(mapNodeService[nodeId], serviceName)
			retCount++
		}
	}

	return retCount
}

func (cs *CenterService) RPC_CallAreaServiceReq(responder rpcHelper.Responder, areaServiceReq *rpc.CallAreaServiceReq) error {
	//整理下需要通知的服务
	coroutine.Go(func() {
		var callAreaServiceRes rpc.CallAreaServiceRes

		//1.预备数据
		callAreaServiceRes.MapServiceInfo = make(map[string]*rpc.ServiceNameInfo, 5)
		//2.分析数据
		type CallService struct {
			NodeId      string
			ServiceName string
			MethodName  string
			InputParam  []byte
		}
		callServiceSlice := make([]CallService, 0, 10)

		for _, serviceInfo := range areaServiceReq.AreaServiceInfo {
			mapNodeService := make(map[string][]string, PreNodeCount)
			cs.getServiceNode(serviceInfo.NodeId, serviceInfo.ServiceName, mapNodeService)

			for nodeId := range mapNodeService {
				callServiceSlice = append(callServiceSlice, CallService{NodeId: nodeId, ServiceName: serviceInfo.ServiceName,
					MethodName: serviceInfo.MethodName, InputParam: serviceInfo.InParam})
			}
		}

		for _, callService := range callServiceSlice {
			var outParam []byte
			err := cs.CallNode(callService.NodeId, callService.ServiceName+"."+callService.MethodName, &callService.InputParam, &outParam)

			//组装数据
			serviceNameInfo, ok := callAreaServiceRes.MapServiceInfo[callService.NodeId]
			if ok == false || serviceNameInfo == nil {
				serviceNameInfo = &rpc.ServiceNameInfo{}
				serviceNameInfo.MapServiceNameInfo = make(map[string]*rpc.CallRet, 3)
				callAreaServiceRes.MapServiceInfo[callService.NodeId] = serviceNameInfo
			}
			callRet, ok := serviceNameInfo.MapServiceNameInfo[callService.ServiceName]
			if ok == false || callRet == nil {
				callRet = &rpc.CallRet{}
				serviceNameInfo.MapServiceNameInfo[callService.ServiceName] = callRet
			}

			//填充结果数据
			if err != nil {
				callRet.Error = err.Error()
			} else {
				callRet.Ret = outParam
			}
		}
		responder(&callAreaServiceRes, rpcHelper.NilError)
	})
	return nil
}

func (cs *CenterService) RPC_GetServiceInfo(inParam *[]byte, outParam *[]byte) error {

	var ServiceInfo struct {
		MapUserInfoNum             int
		MapLoginingNum             int
		MapOutingNum               int
		DBCharacterNum             int32 //DB中角色数量
		NewCharacterNum            int32 //启服后新建角色数量
		AreaId                     int32
		GSMaxPlayerNumbersPriority int
		MapBalance                 map[string]map[string]*BalanceInfo

		ServiceChannelNum int
		ServiceTimerNum   int
	}

	ServiceInfo.MapUserInfoNum = len(cs.mapUserInfo)
	ServiceInfo.MapLoginingNum = len(cs.mapLogining)
	ServiceInfo.MapOutingNum = len(cs.mapOuting)
	ServiceInfo.DBCharacterNum = cs.DBCharacterNum
	ServiceInfo.NewCharacterNum = cs.NewCharacterNum
	ServiceInfo.AreaId = cs.AreaId
	ServiceInfo.GSMaxPlayerNumbersPriority = cs.GSMaxPlayerNumbersPriority
	ServiceInfo.MapBalance = cs.mapBalance

	ServiceInfo.ServiceChannelNum = cs.GetServiceEventChannelNum()
	ServiceInfo.ServiceTimerNum = cs.GetServiceTimerChannelNum()

	var err error
	*outParam, err = json.Marshal(&ServiceInfo)
	if err != nil {
		return err
	}

	return nil
}

// 获取玩家上限状态
func (cs *CenterService) RPC_GetPlayerMaxStatus(reqInfo *rpc.PlaceHolders, resInfo *rpc.PlayerMaxStatus) error {
	if cs.isLoadMaxFinish == false {
		return fmt.Errorf("CenterService is not load finishi")
	}

	resInfo.RegisterIsFull = true
	resInfo.OnlineIsFull = true

	if cs.maxRegisterCount == 0 || cs.maxRegisterCount > (cs.DBCharacterNum+cs.NewCharacterNum) {
		resInfo.RegisterIsFull = false
	} else {
		log.SWarning("maxRegisterCount:", cs.maxRegisterCount, ",DBCharacterNum:", cs.DBCharacterNum, ",NewCharacterNum:", cs.NewCharacterNum)
	}

	if cs.maxOnlineCount == 0 || cs.maxOnlineCount > int32(len(cs.mapUserInfo)) {
		resInfo.OnlineIsFull = false
	} else {
		log.SWarning("maxOnlineCount:", cs.maxOnlineCount, ",len(mapUserInfo):", len(cs.mapUserInfo))
	}

	return nil
}
