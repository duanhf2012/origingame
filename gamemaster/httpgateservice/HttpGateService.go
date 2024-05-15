package httpgateservice

import (
	"encoding/json"
	"fmt"
	"github.com/duanhf2012/origin/v2/sysmodule/ginmodule"
	"math/rand"
	"origingame/common/collect"
	"origingame/common/db"
	"origingame/common/proto/msg"
	"origingame/common/proto/rpc"
	"origingame/common/util"
	"time"

	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/node"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	node.Setup(&HttpGateService{})
}

type AreaGate struct {
	AreaName   string
	AreaId     int
	ShowAreaId int

	//与客户端列表显示顺序相关
	ServerMark   collect.EServerMark
	ServerStatus collect.EServerStatus
	OpenTime     int64 //开服时间戳 ms
	DefaultMark  int   //是否为默认显示的服务器 0：否 1：是

	//IP 端口数据
	GateInfo []GateInfoResp
}

type AreaLoadInfo struct {
	LoadType msg.ServerLoadType
}

// 下面是必须要加载成功才能启动httpgateservice
const (
	AreaLoad int32 = 1 // 区服加载
	AllLoadCount
)

type HttpGateService struct {
	service.Service
	loginModule *LoginModule

	ginModule   ginmodule.GinModule
	mapAreaGate map[int]*AreaGate //map[ShowId][]*TcpGateInfo
	mapNodeArea map[int][]int     //map[nodeId]AreaId

	//加载数据库
	loadAreaTimerId    uint64
	reloadAreaTickerId uint64

	startListenCount int32
	firstLoad        bool
}

func (gate *HttpGateService) OnInit() error {
	gate.mapNodeArea = make(map[int][]int, 1024)
	gate.mapAreaGate = map[int]*AreaGate{}
	gate.loadAreaTimerId = 0
	gate.reloadAreaTickerId = 0
	gate.startListenCount = 0
	gate.firstLoad = true

	var gateCfg struct {
		HttpListen      string
		ForbidGustLogin bool
	}

	err := gate.ParseServiceCfg(&gateCfg)
	if err != nil {
		return err
	}

	//加载全局配置
	gate.ginModule.Init(gateCfg.HttpListen, time.Second*15, nil)
	gate.ginModule.AppendDataProcessor(gate)
	gate.AddModule(&gate.ginModule)

	gate.loginModule = &LoginModule{}
	gate.loginModule.SetForbidGuestLogin(gateCfg.ForbidGustLogin)
	gate.AddModule(gate.loginModule)

	//性能监控
	gate.OpenProfiler()
	gate.GetProfiler().SetOverTime(time.Millisecond * 100)
	gate.GetProfiler().SetMaxOverTime(time.Second * 10)

	//POST方法 请求url:http://127.0.0.1:9000/api/login
	//返回结果为：{"msg":"hello world"}
	gate.ginModule.SafePOST("/api/login", gate.loginModule.Login)

	gate.tryLoadAreaInfoFromDB()
	return nil
}

func (gate *HttpGateService) tryLoadAreaInfoFromDB() {
	gate.SafeAfterFunc(&gate.loadAreaTimerId, time.Second*5, nil, gate.loadAreaInfo)
}

func (gate *HttpGateService) loadAreaInfo(timerId uint64, addition interface{}) {
	gate.loadAreaTimerId = 0
	gate.loadAreaInfoFromDB(true)
}

func (gate *HttpGateService) genRandKey() string {
	return fmt.Sprintf("%d", rand.Int())
}

// 从数据库加载区服信息，参数tryAgain在出错时，重试加载
func (gate *HttpGateService) loadAreaInfoFromDB(tryAgain bool) {
	log.SInfo("start load area info from db.")

	//先查询真实区服
	var reqRealArea db.DBControllerReq
	var realAreaColl collect.CRealAreaInfo
	db.MakeFind(realAreaColl.GetCollName(), realAreaColl.GetCondition(nil), gate.genRandKey(), 65535, realAreaColl.GetSort(), &reqRealArea)
	errCall := gate.AsyncCall(util.AccDBRequest, &reqRealArea, func(res *db.DBControllerRet, err error) {
		if err != nil {
			log.SError("AsyncCall query real area error:", err.Error())
			if tryAgain {
				gate.tryLoadAreaInfoFromDB()
			}
		} else {
			gate.queryRealAreaCallBack(res, tryAgain)
		}
	})

	if errCall != nil {
		log.SError("AsyncCall query real area error:", errCall.Error())
		if tryAgain {
			gate.tryLoadAreaInfoFromDB()
		}
	}
}

func (gate *HttpGateService) queryRealAreaCallBack(realAreaRes *db.DBControllerRet, tryAgain bool) {
	//先读取显示区服
	var reqShowArea db.DBControllerReq
	var showAreaColl collect.CShowAreaInfo
	db.MakeFind(showAreaColl.GetCollName(), showAreaColl.GetCondition(nil), gate.genRandKey(), 65535, showAreaColl.GetSort(), &reqShowArea)

	errCall := gate.AsyncCall(util.AccDBRequest, &reqShowArea, func(res *db.DBControllerRet, err error) {
		if err != nil {
			log.SError("AsyncCall query show area error:", err.Error())
			if tryAgain {
				gate.tryLoadAreaInfoFromDB()
			}
		} else {
			gate.queryShowAreaCallBack(realAreaRes, res)
		}
	})

	if errCall != nil {
		log.SError("AsyncCall query show area error:", errCall.Error())
		if tryAgain {
			gate.tryLoadAreaInfoFromDB()
		}
	}
}

func (gate *HttpGateService) queryShowAreaCallBack(realAreaRes *db.DBControllerRet, showAreaRes *db.DBControllerRet) {
	log.SInfo("start process real area info")

	mapAreaGate := make(map[int]*AreaGate, 100)

	//先整理真实区服
	mapRealArea := make(map[int]collect.CRealAreaInfo, 10)
	for _, realDbData := range realAreaRes.Res {
		realInfo := collect.CRealAreaInfo{}
		err := bson.Unmarshal(realDbData, &realInfo)
		if err != nil {
			log.SError("HttpGateService queryRealAreaCallBack Unmarshal real area err:", err.Error())
			continue
		}
		mapRealArea[realInfo.RealAreaId] = realInfo
	}

	log.SInfo("start process show area info")
	//再整理显示区服
	//mapChannelShowArea := make(map[string][]*AreaGate, 10)
	for _, showDbData := range showAreaRes.Res {
		showInfo := collect.CShowAreaInfo{}
		err := bson.Unmarshal(showDbData, &showInfo)
		if err != nil {
			log.SError("HttpGateService queryRealAreaCallBack Unmarshal  err:", err.Error())
			continue
		}

		if showInfo.AreaName == "" {
			log.SWarning("show areadId ", showInfo.ShowAreaId, " areaName is empty")
			//尚未编辑的服务器就不发给客户端了
			continue
		}

		realInfo, okReal := mapRealArea[showInfo.RealAreaId]
		if okReal == false {
			log.SError("HttpGateService loadDBDataFinish, show area[", showInfo.ShowAreaId, "] this real area[", showInfo.RealAreaId, "] data is error")
			continue
		}

		areaGate := AreaGate{
			AreaName:     showInfo.AreaName,
			AreaId:       showInfo.RealAreaId,
			ShowAreaId:   showInfo.ShowAreaId,
			ServerMark:   showInfo.ServerMark,
			ServerStatus: showInfo.ServerStatus,
			DefaultMark:  showInfo.DefaultMark,
			GateInfo:     make([]GateInfoResp, 0, 2),
		}
		for _, addr := range realInfo.GateList {
			areaGate.GateInfo = append(areaGate.GateInfo, GateInfoResp{Url: addr})
		}
		mapAreaGate[showInfo.ShowAreaId] = &areaGate
	}

	if len(mapAreaGate) == 0 {
		log.SError("No area info data information")
		return
	}

	aList := make([]*AreaGate, 0, 32)
	for _, areaGate := range mapAreaGate {
		aList = append(aList, areaGate)
	}
	jsonAreaGate, err := json.Marshal(&aList)
	if err != nil {
		log.Error("Marshal mapAreaGate error", log.ErrorAttr("err", err))
		return
	}

	gate.loginModule.SetJsonAreaGate(string(jsonAreaGate))

	//区服列表存储
	gate.mapAreaGate = mapAreaGate

	if gate.firstLoad {
		gate.firstLoad = false
		gate.AddStartListCount()
	}

	log.SInfo("load area info finish")
}

func (gate *HttpGateService) AddStartListCount() {
	gate.startListenCount++
	if gate.startListenCount == AllLoadCount {
		gate.ginModule.Start()
	}
}

func (gate *HttpGateService) RPC_ReLoadAreaInfo(reqInfo *rpc.PlaceHolders, resInfo *rpc.PlaceHolders) error {
	if gate.loadAreaTimerId != 0 {
		if gate.CancelTimerId(&gate.loadAreaTimerId) == false {
			log.Stack(fmt.Sprint("HttpGateService CancelTimerId[", gate.loadAreaTimerId, "] failed!"))
		}
	}
	gate.loadAreaInfo(gate.loadAreaTimerId, nil)
	return nil
}

// 在gin协程中
func (gate *HttpGateService) Process(c *gin.Context) (*gin.Context, error) {
	return c, nil
}
