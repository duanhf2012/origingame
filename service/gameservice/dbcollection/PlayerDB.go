package dbcollection

import (
	"container/list"
	"errors"
	"fmt"
	"origingame/common/collect"
	"origingame/common/db"
	"origingame/common/proto/msg"
	"origingame/common/util"
	"origingame/service/interfacedef"

	"github.com/duanhf2012/origin/v2/log"
	"go.mongodb.org/mongo-driver/bson"
)

// 玩家存档数据
type PlayerDB struct {
	gsService interfacedef.IGSService

	rpcSessionKey string //每次创建对象时生成的的唯一key

	Id              string //唯一标识，即UserId
	clientId        string
	collection      [collect.CTMax]collect.ICollection   //单行集合
	multiCollection [collect.MCTMax]collect.MultiRowData //多行实在存档

	playerDBCallBack interfacedef.IPlayerDBCallBack

	totalProgress int //总进度
	loadProgress  int //当前加载的进度
	errProgress   int //加载出错计数

	//新增表需要新增以下
	collect.CUserInfo //用户表
	proxyNum          int
}

func (playerDB *PlayerDB) OnInit(playerDBCallBack interfacedef.IPlayerDBCallBack, gsService interfacedef.IGSService) {
	//以下加入注册新表
	playerDB.rpcSessionKey = util.NewUnionId()
	playerDB.gsService = gsService
	playerDB.playerDBCallBack = playerDBCallBack
	playerDB.loadProgress = 0
	playerDB.errProgress = 0

	playerDB.InitCollection()
}

func (playerDB *PlayerDB) InitCollection() {
	playerDB.totalProgress = 0
	playerDB.RegCollection(&playerDB.CUserInfo, nil)
}

func (playerDB *PlayerDB) RegCollection(coll collect.ICollection, loadBack collect.LoadCallBack) {
	if coll.GetCollectionType() >= collect.CTMax {
		panic("collection error!")
	}

	playerDB.collection[coll.GetCollectionType()] = coll
	playerDB.totalProgress++
}

func (playerDB *PlayerDB) RegMultiCollection(coll collect.IMultiCollection) {

	collType := coll.GetCollectionType()
	if collType >= collect.MCTMax {
		panic("collection error!")
	}

	multiRowData := collect.MultiRowData{}
	multiRowData.Template = coll
	multiRowData.MapCollection = make(map[interface{}]*list.Element, coll.GetMaxRowLimit())

	playerDB.multiCollection[collType] = multiRowData
	playerDB.totalProgress++
}

func (playerDB *PlayerDB) LoadFromDB() {
	//1.Load单行表
	for i := collect.CollectionType(0); i < collect.CTMax; i++ {
		if playerDB.collection[i] == nil {
			continue
		}

		playerDB.load(i)
	}

	//2.Load多行表
	for i := collect.MultiCollectionType(0); i < collect.MCTMax; i++ {
		if playerDB.getMultiCollection(i).Template != nil {
			playerDB.loadMulti(i)
		}
	}
}

func (playerDB *PlayerDB) SaveToDB(bForce bool) {
	//未创建角色的数据直接丢弃
	if playerDB.CUserInfo.NickName == "" {
		return
	}

	//没加载完不允许存档
	if playerDB.IsLoadFinish() == false {
		log.SWarning("userid:", playerDB.Id, " not load finish.")
		return
	}

	for i := collect.CollectionType(0); i < collect.CTMax; i++ {
		coll := playerDB.collection[i]
		if coll == nil || (bForce == false && coll.IsDirty() == false) {
			continue
		}

		//聊天信息数据
		if playerDB.upsetCollection(coll) {
			coll.ClearDirty()
		}
	}
}

func (playerDB *PlayerDB) ExecDB(req *db.DBControllerReq) bool {
	nodeId := util.GetBestNodeIdById(util.AreaDBRequest, playerDB.Id)
	if nodeId == "" {
		log.SError("cannot find nodeId from rpcMethod ", util.AreaDBRequest, ",userid ", playerDB.Id, ",collectName ", req.CollectName)
		return false
	}

	err := playerDB.GoNode(nodeId, util.AreaDBRequest, req)
	if err != nil {
		log.SError("ExecDB fail:go node error :", err.Error(), ",userid ", playerDB.Id, ",collectName ", req.CollectName)
		return false
	}

	return true
}

func (playerDB *PlayerDB) RemoveMultiRow(collType collect.MultiCollectionType, key interface{}) bool {
	collection := playerDB.getMultiCollection(collType)
	if collection.Template != nil {
		elem, ok := collection.MapCollection[key]
		if ok {
			collection.ListICollection.Remove(elem)
			delete(collection.MapCollection, key)
		}

		var req db.DBControllerReq
		err := db.MakeRemoveOneId(collection.Template.GetCollName(), key, playerDB.Id, &req)
		if err != nil {
			log.SError("RemoveMultiRow MakeRemoveOneId err ", err.Error())
			return false
		}
		return playerDB.ExecDB(&req)
	}

	return false
}

func (playerDB *PlayerDB) GetRowByKey(collType collect.MultiCollectionType, key interface{}) collect.IMultiCollection {
	collection := playerDB.getMultiCollection(collType)
	elem, ok := collection.MapCollection[key]
	if ok == false {
		log.SError("cannot find key userid ", playerDB.Id, ",collType %d", collType)
		return nil
	}

	return elem.Value.(collect.IMultiCollection)
}

func (playerDB *PlayerDB) ApplyMultiRow(coll collect.ICollection) bool {
	return playerDB.upsetCollection(coll)
}

func (playerDB *PlayerDB) InsertMultiRow(coll collect.IMultiCollection, needSave bool) bool {
	collType := coll.GetCollectionType()
	collection := playerDB.getMultiCollection(collType)
	if collection.Template == nil {
		log.SError("template is nil,userid ", playerDB.Id, ",collType ", collType)
		return false
	}

	elem := collection.ListICollection.PushBack(coll)
	collection.MapCollection[coll.GetId()] = elem

	//超过最大条数，自动删除超出的
	if collection.ListICollection.Len() > int(coll.GetMaxRowLimit()) {
		frontElem := collection.ListICollection.Front()
		coll := frontElem.Value.(collect.IMultiCollection)
		playerDB.RemoveMultiRow(coll.GetCollectionType(), coll.GetId())
	}

	//存档
	if needSave {
		return playerDB.upsetMultiCollection(coll)
	}

	return true
}

func (playerDB *PlayerDB) UpdateMultiCollectionData(coll collect.IMultiCollection) bool {
	var req db.DBControllerReq
	updateData := coll.GetUpdateData()
	if len(updateData) == 0 {
		//无需更新数据
		return true
	}
	err := db.MakeUpsetId(coll.GetCollName(), coll.GetId(), updateData, playerDB.Id, &req)
	if err != nil {
		log.SError("make multi updateMultiCollection fail ", err.Error())
		return false
	}

	return playerDB.ExecDB(&req)
}

func (playerDB *PlayerDB) GetMultiRow(collType collect.MultiCollectionType) list.List {
	return playerDB.multiCollection[collType].ListICollection
}

func (playerDB *PlayerDB) IsLoadFinish() bool {
	return playerDB.loadProgress >= playerDB.totalProgress && playerDB.errProgress == 0
}

func (playerDB *PlayerDB) Clear() {
	playerDB.Id = ""
	playerDB.totalProgress = 0
	playerDB.loadProgress = 0
	playerDB.errProgress = 0
	playerDB.proxyNum = 0

	for i := collect.MultiCollectionType(0); i < collect.MCTMax; i++ {
		playerDB.getMultiCollection(i).Clean()
	}

	for i := collect.CollectionType(0); i < collect.CTMax; i++ {
		if playerDB.collection[i] == nil {
			continue
		}
		playerDB.collection[i].Clean()
	}
}

func (playerDB *PlayerDB) MarshalCollection(marshalData map[int32]*msg.Bytes) error {
	var err error
	for i := collect.CTUserInfo; i < collect.CTMax; i++ {
		if playerDB.collection[i] == nil {
			continue
		}

		marshalItem := &msg.Bytes{
			Value: nil,
		}

		marshalItem.Value, err = bson.Marshal(playerDB.collection[i])
		if err != nil {
			return err
		}
		marshalData[i] = marshalItem
	}

	return nil
}

func (playerDB *PlayerDB) UnmarshalCollection(pbData map[int32]*msg.Bytes) error {
	for i := collect.CTUserInfo; i < collect.CTMax; i++ {
		if playerDB.collection[i] == nil {
			continue
		}

		err := bson.Unmarshal(pbData[i].Value, playerDB.collection[i])
		if err != nil {
			return err
		}
		playerDB.loadProgress++
		playerDB.collection[i].OnLoadSucc(false, playerDB.Id)
	}

	return nil
}

func (playerDB *PlayerDB) MarshalMultiCollection(marshalData map[int32]*msg.BytesList) error {
	var err error
	for i := collect.MCTUserMail; i < collect.MCTMax; i++ {
		if playerDB.multiCollection[i].Template == nil {
			continue
		}

		marshalItem := &msg.BytesList{
			ValueList: make([]*msg.Bytes, 0, 65535),
		}

		for e := playerDB.multiCollection[i].ListICollection.Front(); e != nil; e = e.Next() {
			multiItem := &msg.Bytes{
				Value: nil,
			}
			multiItem.Value, err = bson.Marshal(e.Value.(collect.IMultiCollection))
			if err != nil {
				return err
			}

			marshalItem.ValueList = append(marshalItem.ValueList, multiItem)
		}

		marshalData[i] = marshalItem
		playerDB.loadProgress++
	}

	return nil
}

func (playerDB *PlayerDB) UnmarshalMultiCollection(pbData map[int32]*msg.BytesList) error {
	for i := collect.MCTUserMail; i < collect.MCTMax; i++ {
		if playerDB.multiCollection[i].Template == nil {
			continue
		}

		collection := playerDB.getMultiCollection(i)
		if collection == nil || collection.Template == nil {
			return fmt.Errorf("no this collection %d", i)
		}

		for _, pbItem := range pbData[i].ValueList {
			rowData := collection.Template.MakeRow()
			err := bson.Unmarshal(pbItem.Value, rowData)
			if err != nil {
				return err
			}

			pElem := collection.ListICollection.PushBack(rowData)
			collection.MapCollection[rowData.GetId()] = pElem
		}
		playerDB.loadProgress++
	}

	return nil
}

func (playerDB *PlayerDB) getMultiCollection(collType collect.MultiCollectionType) *collect.MultiRowData {
	return &playerDB.multiCollection[collType]
}

func (playerDB *PlayerDB) upsetCollection(coll collect.ICollection) bool {
	var req db.DBControllerReq
	err := db.MakeCacheUpsetId(coll.GetCollName(), coll.GetId(), coll, playerDB.Id, playerDB.Id, &req)
	if err != nil {
		log.SError("make upsetid fail ", err.Error())
		return false
	}
	return playerDB.ExecDB(&req)
}

func (playerDB *PlayerDB) upsetMultiCollection(coll collect.IMultiCollection) bool {
	var req db.DBControllerReq
	err := db.MakeUpsetId(coll.GetCollName(), coll.GetId(), coll, playerDB.Id, &req)
	if err != nil {
		log.SError("make multi upsetid fail ", err.Error())
		return false
	}
	return playerDB.ExecDB(&req)
}

func (playerDB *PlayerDB) dbMultiLoadCallBack(collType collect.MultiCollectionType, res *db.DBControllerRet, err error) {
	multiCollection := playerDB.getMultiCollection(collType)
	if multiCollection == nil || multiCollection.Template == nil {
		// 这里事实上应该不可能
		log.SError("load multiCollection fail ", err.Error(), ",userid ", playerDB.Id, " collType ", collType, " not support")
		return
	}

	if err != nil {
		log.SError("load userid ", playerDB.Id, " db multi collType ", collType, " is error!")
		playerDB.errProgress++
	}

	//加载多行数据
	err = multiCollection.LoadFromDB(res)
	if err != nil {
		log.SError("load multiCollection fail ", err.Error(), ",userid ", playerDB.Id, " collType ", collType)
		playerDB.errProgress++
	}

	playerDB.loadProgress++
	if playerDB.loadProgress >= playerDB.totalProgress {
		if playerDB.errProgress > 0 {
			playerDB.playerDBCallBack.OnLoadDBEnd(false)
		} else {
			playerDB.playerDBCallBack.OnLoadDBEnd(true)
		}
	}
}

func (playerDB *PlayerDB) loadMulti(typ collect.MultiCollectionType) {
	var coll collect.IMultiCollection
	multiCollection := playerDB.getMultiCollection(typ)
	if multiCollection != nil && multiCollection.Template != nil {
		coll = multiCollection.Template
	}

	if coll == nil {
		log.SError("loadMulti cannot find collection,load fail!")
		playerDB.dbMultiLoadCallBack(typ, nil, errors.New("loadMulti cannot find collection,load fail!"))
		return
	}

	var req db.DBControllerReq
	err := db.MakeFind(coll.GetCollName(), coll.GetCondition(playerDB.Id), playerDB.Id, coll.GetMaxRowLimit(), coll.GetSort(), &req)
	if err != nil {
		log.SError("loadMulti MakeFind err ", err.Error())
		return
	}
	nodeId := util.GetBestNodeIdById(util.AreaDBRequest, playerDB.Id)
	if nodeId == "" {
		log.SError("Cannot find ", util.AreaDBRequest, " nodeId!")
		playerDB.dbMultiLoadCallBack(typ, nil, errors.New("Cannot find DBService.RPC_DBRequest nodeId!"))
		return
	}

	err = AsyncCallNode(playerDB, nodeId, util.AreaDBRequest, &req, func(res *db.DBControllerRet, err error) {
		if err != nil {
			log.SError("AsyncCall userid:", playerDB.Id, ",type:", typ, ",error :", err.Error())
		}
		playerDB.dbMultiLoadCallBack(typ, res, err)
	})

	if err != nil {
		log.SError("loadMulti AsyncCallNode err ", err.Error())
	}
}

func (playerDB *PlayerDB) load(typ collect.CollectionType) {
	var coll collect.ICollection
	if playerDB.collection[typ] != nil {
		coll = playerDB.collection[typ]
	}

	if coll == nil {
		playerDB.playerDBCallBack.OnLoadDBEnd(false)
		log.SError("load cannot find collection,load fail!")
		return
	}

	var req db.DBControllerReq
	err := db.MakeCacheFind(coll.GetCollName(), coll.GetCondition(playerDB.Id), playerDB.Id, 1, coll.GetSort(), coll.GetCacheId(playerDB.Id), &req)
	if err != nil {
		log.SError("load MakeFind err ", err.Error())
		return
	}

	nodeId := util.GetBestNodeIdById(util.AreaDBRequest, playerDB.Id)
	if nodeId == "" {
		log.SError("Cannot find ", util.AreaDBRequest, " nodeId!")
		return
	}

	err = AsyncCallNode(playerDB, nodeId, util.AreaDBRequest, &req, func(res *db.DBControllerRet, err error) {
		if err != nil {
			log.SError("AsyncCall userid:", playerDB.Id, ",type:", typ, ",error :", err.Error())
		}
		playerDB.dbLoadCallBack(typ, res, err)
	})

	if err != nil {
		log.SError("load AsyncCallNode err ", err.Error())
	}
}

func (playerDB *PlayerDB) dbLoadCallBack(collType collect.CollectionType, res *db.DBControllerRet, err error) {
	if err != nil {
		log.SError("load userid ", playerDB.Id, " db collType ", collType, " is error!")
		playerDB.errProgress++
	} else {
		if playerDB.collection[collType] != nil {
			if len(res.Res) > 0 {
				//加载单行数据
				err := bson.Unmarshal(res.Res[0], playerDB.collection[collType])
				if err != nil {
					log.SError("bson.Unmarshal fail ", err.Error(), ",userid ", playerDB.Id, " collType ", collType)
					playerDB.errProgress++
				}
			}
			playerDB.collection[collType].OnLoadSucc(len(res.Res) == 0, playerDB.Id)
		}
	}
	playerDB.loadProgress++

	if playerDB.loadProgress >= playerDB.totalProgress {
		if playerDB.errProgress > 0 {
			playerDB.playerDBCallBack.OnLoadDBEnd(false)
		} else {
			playerDB.playerDBCallBack.OnLoadDBEnd(true)
		}
	}
}

func (playerDB *PlayerDB) CheckHasErrProgress() bool {
	return playerDB.errProgress > 0
}

func (playerDB *PlayerDB) GetCollect(collectType collect.CollectionType) collect.ICollection {
	if collectType >= collect.CTMax {
		log.Stack(fmt.Sprint("collectType is error ", collectType))
		return nil
	}

	return playerDB.collection[collectType]
}

func (playerDB *PlayerDB) GetCollectUserInfo() *collect.CUserInfo {
	userInfo, ok := playerDB.collection[collect.CTUserInfo].(*collect.CUserInfo)
	if ok == false {
		return nil
	}
	return userInfo
}

func (playerDB *PlayerDB) GetMultiCollectionDataMap(collectionType collect.MultiCollectionType) map[interface{}]*list.Element {
	return playerDB.multiCollection[collectionType].MapCollection
}

func (playerDB *PlayerDB) GetId() string {
	return playerDB.Id
}

func (playerDB *PlayerDB) GoNode(nodeId string, serviceMethod string, args interface{}) error {
	return playerDB.gsService.GoNode(nodeId, serviceMethod, args)
}

func AsyncCallNode[RpcMsg any](playerDB *PlayerDB, nodeId string, serviceMethod string, args interface{}, callback func(ret *RpcMsg, err error)) error {
	rpcKey := playerDB.rpcSessionKey
	return playerDB.gsService.AsyncCallNode(nodeId, serviceMethod, args, func(ret *RpcMsg, err error) {
		nowRpcSessionKey := playerDB.rpcSessionKey
		if rpcKey != nowRpcSessionKey {
			log.Stack(fmt.Sprint("serviceMethod:", serviceMethod, " is fail rpc key:", rpcKey, " now rpc key:", nowRpcSessionKey))
			return
		}

		callback(ret, err)
	})
}
