package centerservice

import (
	"github.com/duanhf2012/origin/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"origingame/common/collect"
	"origingame/common/db"
	"origingame/common/util"
)

func (cs *CenterService) loadShowAreaInfoFromDB() {
	var req db.DBControllerReq
	err := db.MakeFind(collect.ShowAreaInfoDBName, bson.D{{Key: "rId", Value: cs.AreaId}}, 0, 30, nil, &req)
	if err != nil {
		log.SError("loadShowAreaInfoFromDB DB Req create err: ", err.Error())
		return
	}

	nodeId := util.GetBestNodeId(util.AccDBRequest, 0)
	if nodeId == "" {
		log.SError("Cannot find ", util.AccDBRequest, " nodeId!")
		return
	}

	errCall := cs.AsyncCallNode(nodeId, util.AccDBRequest, &req, func(res *db.DBControllerRet, err error) {
		if err != nil {
			log.SError("db operator ret error, error:", err.Error())
			return
		} else {
			cs.ShowAreaMap = make(map[int]*collect.CShowAreaInfo, 10)
			for _, dbData := range res.Res {
				showAreaInfo := &collect.CShowAreaInfo{}
				err = bson.Unmarshal(dbData, showAreaInfo)
				if err != nil {
					log.SError("loadShowAreaInfoFromDB Unmarshal  err:", err.Error())
					continue
				}

				cs.ShowAreaMap[showAreaInfo.ShowAreaId] = showAreaInfo
			}
		}
	})

	if errCall != nil {
		log.SError("call db operator error, error:", err.Error())
		return
	}
}

func (cs *CenterService) loadAreaInfoFromDB(needTryAgain bool) {
	var coll collect.CRealAreaInfo
	var req db.DBControllerReq
	db.MakeFind(coll.GetCollName(), bson.D{{Key: "_id", Value: cs.AreaId}}, 0, 1, nil, &req) //OptType_FindOneAndUpdate 必须构造一个insert数据 个人感觉不大好

	errCall := cs.AsyncCall(util.AccDBRequest, &req, func(res *db.DBControllerRet, err error) {
		//处理
		cs.dealAreaInfoUpdate(res, err, needTryAgain)
	})

	if errCall != nil {
		log.SError("CenterService.startLoadAreaInfo AsyncCallNode err:", errCall.Error())
		if needTryAgain {
			cs.tryLoadAreaInfoFromDB()
		}
	}
}

func (cs *CenterService) dealAreaInfoUpdate(res *db.DBControllerRet, err error, needTryAgain bool) {
	//查询返回结果检查
	if err != nil {
		log.SError("startDealUserMailUpdate, error :", err.Error())
		if needTryAgain {
			cs.tryLoadAreaInfoFromDB()
		}
		return
	}

	//必定查到一条数据
	if len(res.Res) != 1 {
		log.SError("no this real area data[", cs.AreaId, "]")
		if needTryAgain {
			cs.tryLoadAreaInfoFromDB()
		}
		return
	}

	realInfo := &collect.CRealAreaInfo{}
	errUnmarshal := bson.Unmarshal(res.Res[0], realInfo)
	if errUnmarshal != nil {
		//因为仅一条数据,这里解析错误
		log.SError("CenterService dealUserMailUpdate Unmarshal  err:", errUnmarshal.Error())
		if needTryAgain {
			cs.tryLoadAreaInfoFromDB()
		}
		return
	}

	cs.maxOnlineCount = realInfo.MaxLoginCount
	cs.maxRegisterCount = realInfo.MaxRegCount
	cs.isLoadMaxFinish = true
}
