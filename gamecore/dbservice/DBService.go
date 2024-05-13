package dbservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"origingame/common/db"
	"origingame/common/performance"
	"origingame/common/util"
	"runtime"

	"sync"
	"sync/atomic"
	"time"

	"github.com/pierrec/lz4/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/duanhf2012/origin/v2/sysmodule/mongodbmodule"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/duanhf2012/origin/v2/log"
	"github.com/duanhf2012/origin/v2/node"
	"github.com/duanhf2012/origin/v2/rpc"
	"github.com/duanhf2012/origin/v2/service"
	"github.com/duanhf2012/origin/v2/sysmodule/redismodule"
	"github.com/duanhf2012/origin/v2/util/timer"
	"go.mongodb.org/mongo-driver/bson"
)

const MaxKeyNum = 33

func init() {
	AccountDBService := &DBService{}
	AccountDBService.SetName(util.AccountDBService)

	node.Setup(AccountDBService, &DBService{})
}

var emptyRes [][]byte

type DBService struct {
	service.Service
	mongoModule    mongodbmodule.MongoModule
	redisModule    redismodule.RedisModule
	channelOptData []chan DBRequest
	url            string
	dbName         string
	goroutineNum   uint32

	channelNum int

	dbDealCount   int32
	dbAllCostTime int64
	dbMaxCostTime int64
	slowQueryTime int64

	//导入昵称
	importIntervalTime time.Duration //导入间隔时间
	importNum          int

	performanceAnalyzer *performance.PerformanceAnalyzer //性能分析器

	MapCache     map[string]*util.FCMap
	MapCacheLock sync.RWMutex

	CacheCompress        bool  //是否存缩存储
	MaxCacheCap          int   //最大缓存容量
	ExpirationTimeSecond int64 //最大缓存有效时间(秒)
	CheckIntervalSecond  int64 //过期检查间隔(秒)
	IntervalCheckNum     int   //检查检查最大缓存数

	configRedis redismodule.ConfigRedis
}

// ReadCfg 读取DB配置,配置在service.json中
func (dbService *DBService) ReadCfg() error {
	mapDBServiceCfg, ok := dbService.GetServiceCfg().(map[string]interface{})
	if ok == false {
		return fmt.Errorf("DBService config is error!")
	}

	goroutineNum, ok := mapDBServiceCfg["GoroutineNum"]
	if ok == false {
		return fmt.Errorf("DBService config is error!")
	}
	dbService.goroutineNum = uint32(goroutineNum.(float64))

	slowQueryTime, ok := mapDBServiceCfg["SlowQueryTime"]
	if ok == true {
		dbService.slowQueryTime = int64(slowQueryTime.(float64))
	}

	channelNum, ok := mapDBServiceCfg["ChannelNum"]
	if ok == false {
		return fmt.Errorf("DBService config is error!")
	}
	dbService.channelNum = int(channelNum.(float64))

	//加入性能统计日志
	var analyzerInterval time.Duration
	intervalTime, okOpen := mapDBServiceCfg["PerformanceIntervalTime"]
	if okOpen == true {
		analyzerInterval = time.Duration(intervalTime.(float64)) * time.Millisecond
	}
	analyzerLogLevel := performance.AnalyzerLogLevel1
	analyzerLevel, okLogLevel := mapDBServiceCfg["PerformanceLogLevel"]
	if okLogLevel == true {
		analyzerLogLevel = int(analyzerLevel.(float64))
	}

	dbService.performanceAnalyzer = &performance.PerformanceAnalyzer{}
	InitPerformanceAnalyzer(dbService.performanceAnalyzer, analyzerInterval, analyzerLogLevel)
	_, err := dbService.AddModule(dbService.performanceAnalyzer)
	if err != nil {
		log.SError("DBService.OnInit AddModule err:", err.Error())
		return err
	}

	return nil
}

// OnInit 初始化
func (dbService *DBService) OnInit() error {
	err := dbService.ReadCfg()
	if err != nil {
		return err
	}

	err = dbService.mongoModule.Init(dbService.url, time.Second*15)
	if err != nil {
		log.SError("Init dbService[", dbService.dbName, "], url[", dbService.url, "] init error:", err.Error())
		return err
	}

	err = dbService.mongoModule.Start()
	if err != nil {
		log.SError("start dbService[", dbService.dbName, "], url[", dbService.url, "] init error:", err.Error())
		return err
	}

	dbService.channelOptData = make([]chan DBRequest, dbService.goroutineNum)
	for i := uint32(0); i < dbService.goroutineNum; i++ {
		dbService.channelOptData[i] = make(chan DBRequest, dbService.channelNum)
		go dbService.ExecuteOptData(dbService.channelOptData[i])
	}

	dbService.dbDealCount = 0
	dbService.dbAllCostTime = 0
	dbService.dbMaxCostTime = 0

	//性能监控
	dbService.OpenProfiler()
	dbService.GetProfiler().SetOverTime(time.Millisecond * 500)
	dbService.GetProfiler().SetMaxOverTime(time.Second * 10)

	//同步索引
	err = dbService.SyncDBIndex()
	if err != nil {
		return err
	}

	dbService.MapCache = make(map[string]*util.FCMap, 5)
	return nil
}

func (dbService *DBService) OnStart() {

}

func (dbService *DBService) SyncDBIndex() error {
	serviceName := dbService.GetName()
	s := dbService.mongoModule.TakeSession()

	switch serviceName {
	case util.DBService:
		var IndexKey [][]string
		var keys []string
		keys = append(keys, "PlatId", "ShowAreaId")
		IndexKey = append(IndexKey, keys)
		if err := s.EnsureUniqueIndex(dbService.dbName, "UserInfo", IndexKey, true, true, true); err != nil {
			return err
		}

		//昵称唯一索引
		IndexKey = make([][]string, 0, 1)
		keys = make([]string, 0, 1)
		keys = append(keys, "NickName")
		IndexKey = append(IndexKey, keys)
		if err := s.EnsureUniqueIndex(dbService.dbName, "UserInfo", IndexKey, true, true, true); err != nil {
			return err
		}

	case util.AccountDBService:
	}

	return nil
}

// PrintDBCost 打印
func (dbService *DBService) PrintDBCost(tm *timer.Ticker) {
	averageCostTime := int64(0)
	if dbService.dbDealCount != 0 {
		averageCostTime = dbService.dbAllCostTime / int64(dbService.dbDealCount)
	}
	log.SInfo("DBService dbDealCount[", dbService.dbDealCount, "], dbMaxCostTime[", dbService.dbMaxCostTime, "], allCostTime[", dbService.dbAllCostTime, "], averageCostTime[", averageCostTime, "]")
}

// ExecuteOptData 执行DB操作协程
func (dbService *DBService) ExecuteOptData(channelOptData chan DBRequest) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			err := fmt.Errorf("%v: %s", r, buf[:l])
			log.SError("core dump info:", err.Error(), "\n")
			dbService.ExecuteOptData(channelOptData)
		}
	}()

	for {
		select {
		case optData := <-channelOptData:
			timeNow := time.Now()
			switch optData.request.GetType() {
			case db.OptType_DelOne:
				err := dbService.DoDelOne(optData)
				if err != nil {
					log.SError("OptType_Del DoDelOne err ", err.Error())
				}
			case db.OptType_UpdateFieldOpt:
				{
					err := dbService.DoUpdateField(optData)
					if err != nil {
						log.Error("OptType_Update DoUpdate err ", err.Error())
					}
				}
			case db.OptType_Update:
				err := dbService.DoUpdate(optData)
				if err != nil {
					log.SError("OptType_Update DoUpdate err ", err.Error())
				}
			case db.OptType_UpdateMany:
				err := dbService.DoUpdateMany(optData)
				if err != nil {
					log.SError("OptType_Update DoUpdateMany err ", err.Error())
				}
			case db.OptType_Find:
				err := dbService.DoFind(optData)
				if err != nil {
					log.SError("OptType_Find DoFind err ", err.Error())
				}
			case db.OptType_Insert:
				err := dbService.DoInsert(optData)
				if err != nil {
					log.SError("OptType_Insert DoInsert err ", err.Error())
				}
				//			case db.OptType_Insert + db.OptType_Update:
				//				dbService.DoInsertUpdate(optData)
			//case db.OptType_SetOnInsert:
			//	dbService.DoSetOnInsert(optData)
			case db.OptType_FindOneAndUpdate:
				err := dbService.DoFindOneAndUpset(optData)
				if err != nil {
					log.SError("OptType_FindOneAndUpdate DoFindOneAndUpdate err ", err.Error())
				}
			case db.OptType_Upset:
				err := dbService.DoUpSet(optData)
				if err != nil {
					log.SError("OptType_Upset DoUpSet err ", err.Error())
				}
			case db.OptType_Count:
				err := dbService.DoCount(optData)
				if err != nil {
					log.SError("OptType_Count DoCount err ", err.Error())
				}
			case db.OptType_DelMany:
				err := dbService.DoDelMany(optData)
				if err != nil {
					log.SError("OptType_Del DoDelOne err ", err.Error())
				}
			case db.OptType_FindManyKey:
				err := dbService.DoFindManyKey(optData)
				if err != nil {
					log.SError("OptType_FindManyKey DoFindManyKey err ", err.Error())
				}
			case db.OptType_Redis_SetKey:
				err := dbService.DoRedisSetKey(optData)
				if err != nil {
					log.SError("OptType_FindManyKey DoFindManyKey err ", err.Error())
				}
			case db.OptType_Redis_GetKey:
				err := dbService.DoRedisGetKey(optData)
				if err != nil {
					log.SError("OptType_FindManyKey DoFindManyKey err ", err.Error())
				}
			case db.OptType_Redis_DelKey:
				err := dbService.DoRedisDelKey(optData)
				if err != nil {
					log.SError("OptType_FindManyKey DoFindManyKey err ", err.Error())
				}
			default:
				log.SError("optype ", optData.request.GetType(), " is error.")
			}

			costTime := time.Now().Sub(timeNow).Milliseconds()
			if atomic.LoadInt64(&dbService.dbMaxCostTime) < costTime {
				atomic.StoreInt64(&dbService.dbMaxCostTime, costTime)
			}
			atomic.AddInt64(&dbService.dbAllCostTime, costTime)
			atomic.AddInt32(&dbService.dbDealCount, 1)

			if dbService.slowQueryTime > 0 && costTime >= dbService.slowQueryTime {
				var condition interface{}
				err := bson.Unmarshal(optData.request.GetCondition(), &condition)
				if err != nil {
					log.SWarning("DBService.ExecuteOptData[", int32(optData.request.GetType()), "] slow[", costTime, "]", " CollectName:",
						optData.request.CollectName, " condition unmarshal error")
				} else {
					if bsonE, ok := condition.(primitive.E); ok {
						log.SWarning("DBService.ExecuteOptData[", int32(optData.request.GetType()), "] slow[", costTime, "]", " CollectName:",
							optData.request.CollectName, " condition:", fmt.Sprintf("%v", bsonE))
					} else if bsonD, ok := condition.(primitive.D); ok {
						log.SWarning("DBService.ExecuteOptData[", int32(optData.request.GetType()), "] slow[", costTime, "]", " CollectName:",
							optData.request.CollectName, " condition:", fmt.Sprintf("%v", bsonD))
					} else {
						log.SWarning("DBService.ExecuteOptData[", int32(optData.request.GetType()), "] slow[", costTime, "]", " CollectName:",
							optData.request.CollectName, " condition type unknow")
					}
				}
			}
		}
	}
}

// DoSetOnInsert 执行方法——数据存在则更新，不存在则插入
func (dbService *DBService) DoUpSet(dbReq DBRequest) error {
	return dbService.update(dbReq, true)
}

// findCache 从缓存中查找数据指定数据并选择字段列
func (dbService *DBService) findUpdateField(collectName string, cacheId uint64, fieldValue []byte, version int32) {
	if cacheId == 0 || dbService.MaxCacheCap == 0 {
		if cacheId == 0 {
			// 这里加一个防御性报错
			mapCache, _ := dbService.GetMapCacheByCollectName(collectName, false)
			if mapCache != nil {
				log.Stack(collectName + " has cache, but cacheId is 0")
			}
		}
		return
	}

	//1.找不到时，也先建立，以免，数据协程定入数据时，对该MapCache产生读写冲突
	mapCache, ok := dbService.GetMapCacheByCollectName(collectName, false)
	if ok == false {
		return
	}

	//2.从缓存对象中查找
	data := mapCache.FindData(cacheId)
	if data == nil {
		return
	}

	// 1. addCount := bson.M{"$inc": bson.M{"Count": 1}} 将字段+1
	// 2. upsertCount := bson.M{"$set": bson.M{"Count": 1}} 将字段设置为1

	//3.数据转换
	var cacheData [][]byte
	cacheData, ok = data.([][]byte)
	if ok == false {
		log.SError("cannot convert data ", collectName)
		mapCache.RemoveCache(cacheId)
		return
	}

	if dbService.CacheCompress == false {
		//如果是没有压缩的情况，又需要选择字段
		var retData [][]byte
		for _, rowData := range cacheData {
			byteData, err := dbService.updateFieldByFieldByte(collectName, fieldValue, rowData)
			if ok == false {
				log.Error("updateFieldByFieldByte fail", log.ErrorAttr("error", err))
				mapCache.RemoveCache(cacheId)
				return
			}

			retData = append(retData, byteData)
		}

		mapCache.UpsertData(cacheId, retData, version)
		return
	}

	//4.如果缓存被压缩，则需要解压
	var retData [][]byte
	var tmpByteBuff [409600]byte

	//遍历所有的缓存数据
	for _, rowData := range cacheData {
		//dest := dbService.compressBuff
		//长度必需为1，因为第0个位置存放是否被压缩
		if len(rowData) == 0 {
			log.SError("rowData is error :", collectName, " cacheId:", cacheId)
			mapCache.RemoveCache(cacheId)
			return
		}

		var dest []byte
		if rowData[0] == 0 {
			dest = rowData[1:]
		} else {
			uncompressDest := tmpByteBuff[:] // make([]byte, 10*len(rowData))
			cnt, err := dbService.UncompressBlock(rowData[1:], uncompressDest)
			if err != nil {
				log.SError("UncompressBlock fail collection:", collectName, " cacheId:", cacheId, " error:", err.Error(), " src len:", len(rowData[1:]), " dest len:", len(uncompressDest))
				mapCache.RemoveCache(cacheId)
				return
			}

			dest = make([]byte, cnt)
			if copy(dest, uncompressDest[:cnt]) != cnt {
				log.SError("copy fail!")
				mapCache.RemoveCache(cacheId)
				return
			}
		}

		//从库存数据中选择指定字段，并序列化返回
		//dbService.updateFieldByFieldByte(fieldValue)
		byteData, err := dbService.updateFieldByFieldByte(collectName, fieldValue, dest)
		if err != nil {
			log.Error("updateFieldByFieldByte fail", log.ErrorAttr("error", err))
			mapCache.RemoveCache(cacheId)
			return
		}

		retData = append(retData, byteData)
	}

	//压缩
	var compressByte [][]byte
	for _, rowData := range retData {
		dest := make([]byte, lz4.CompressBlockBound(len(rowData))+1)
		cnt, err := dbService.CompressBlock(rowData, dest[1:])
		if err != nil || cnt == 0 {
			log.SError("compress block collectName:", collectName, " cacheId:", cacheId, " err:", err.Error())
			//异常情况，删除缓存
			mapCache.RemoveCache(cacheId)
			return
		}

		//没有压缩
		if cnt >= len(rowData) {
			dest[0] = 0 //标记是否有压缩
			if copy(dest[1:], rowData) != len(rowData) {
				log.SError("compress block collectName:", collectName, " cacheId:", cacheId, " copy error")
				//异常情况，删除缓存
				mapCache.RemoveCache(cacheId)
				return
			}
			compressByte = append(compressByte, dest[:len(rowData)+1])
		} else {
			dest[0] = 1 //有压缩
			compressByte = append(compressByte, dest[:cnt+1])
		}
	}

	mapCache.UpsertData(cacheId, compressByte, version)
	return
}

func (dbService *DBService) updateField(dbReq DBRequest) error {
	//1.选择数据库与表
	s := dbService.mongoModule.TakeSession()

	//2.设置数据
	if len(dbReq.request.RawData) != 1 {
		err := fmt.Errorf("%s DoUpdate data len is error %d.", dbReq.request.CollectName, len(dbReq.request.RawData))
		log.SError(err.Error())
		dbService.responseRet(dbReq, err, 0, 0, 0, 0)
		return err
	}

	dbService.findUpdateField(dbReq.request.CollectName, dbReq.request.CacheId, dbReq.request.RawData[0], dbReq.request.Version)

	//3.设置条件
	var condition interface{}
	unmarshalErr := bson.Unmarshal(dbReq.request.GetCondition(), &condition)
	if unmarshalErr != nil {
		log.SError("bson.Unmarshal err ", unmarshalErr.Error())
		dbService.responseRet(dbReq, unmarshalErr, 0, 0, 0, 0)
		return unmarshalErr
	}

	var updateData any
	err := bson.Unmarshal(dbReq.request.RawData[0], &updateData)
	if err != nil {
		errs := fmt.Errorf("%s DoInsertUpdate data Unmarshal RawData error %s.", dbReq.request.CollectName, err.Error())
		log.SError(errs.Error())
		dbService.responseRet(dbReq, errs, 0, 0, 0, 0)
		return errs
	}

	ctx, cancel := s.GetDefaultContext()
	defer cancel()
	ret, err := s.Collection(dbService.dbName, dbReq.request.GetCollectName()).UpdateOne(ctx, condition, updateData)
	if dbReq.responder.IsInvalid() == false {
		matchedCount, modifiedCount, upsertedCount := int32(0), int32(0), int32(0)
		if ret != nil {
			matchedCount, modifiedCount, upsertedCount = int32(ret.MatchedCount), int32(ret.ModifiedCount), int32(ret.UpsertedCount)
		}

		dbService.responseRet(dbReq, err, matchedCount, modifiedCount, upsertedCount, 0)
	}
	return nil
}

func (dbService *DBService) update(dbReq DBRequest, upset bool) error {
	dbService.updateCache(dbReq)

	//1.选择数据库与表
	s := dbService.mongoModule.TakeSession()

	//2.设置数据
	if len(dbReq.request.Data) != 0 && len(dbReq.request.RawData) != 0 {
		err := fmt.Errorf("%s DoUpdate data len is error %d.", dbReq.request.CollectName, len(dbReq.request.Data))
		log.SError(err.Error())
		dbService.responseRet(dbReq, err, 0, 0, 0, 0)
		return err
	}

	//3.设置条件
	var condition interface{}
	unmarshalErr := bson.Unmarshal(dbReq.request.GetCondition(), &condition)
	if unmarshalErr != nil {
		log.SError("bson.Unmarshal err ", unmarshalErr.Error())
		dbService.responseRet(dbReq, unmarshalErr, 0, 0, 0, 0)
		return unmarshalErr
	}

	var updateData any
	if len(dbReq.request.Data) != 0 {
		var data interface{}
		err := bson.Unmarshal(dbReq.request.Data[0], &data)
		if err != nil {
			err := fmt.Errorf("%s DoInsertUpdate data Unmarshal error %s.", dbReq.request.CollectName, err.Error())
			log.SError(err.Error())
			dbService.responseRet(dbReq, err, 0, 0, 0, 0)
			return err
		}
		updateData = bson.M{"$set": data}

		// 单独修改字段，删除Cache
		//dbService.RemoveCache(dbReq)
	} else {
		err := bson.Unmarshal(dbReq.request.RawData[0], &updateData)
		if err != nil {
			errs := fmt.Errorf("%s DoInsertUpdate data Unmarshal RawData error %s.", dbReq.request.CollectName, err.Error())
			log.SError(errs.Error())
			dbService.responseRet(dbReq, errs, 0, 0, 0, 0)
			return errs
		}
	}

	var UpdateOptionsOpts []*options.UpdateOptions
	if upset == true {
		UpdateOptionsOpts = append(UpdateOptionsOpts, options.Update().SetUpsert(true))
	}

	ctx, cancel := s.GetDefaultContext()
	defer cancel()

	ret, err := s.Collection(dbService.dbName, dbReq.request.GetCollectName()).UpdateOne(ctx, condition, updateData, UpdateOptionsOpts...)
	//changeInfo, err := collect.Upsert(condition, bson.M{"$setOnInsert": data})

	if dbReq.responder.IsInvalid() == false {
		matchedCount, modifiedCount, upsertedCount := int32(0), int32(0), int32(0)
		if ret != nil {
			matchedCount, modifiedCount, upsertedCount = int32(ret.MatchedCount), int32(ret.ModifiedCount), int32(ret.UpsertedCount)
		}

		dbService.responseRet(dbReq, err, matchedCount, modifiedCount, upsertedCount, 0)
	}
	return nil
}

// 删除缓存cache
func (dbService *DBService) RemoveCache(dbReq DBRequest) {
	mapCache, _ := dbService.GetMapCacheByCollectName(dbReq.request.CollectName, false)
	if mapCache != nil {
		if dbReq.request.CacheId > 0 {
			mapCache.RemoveCache(dbReq.request.CacheId)
		} else {
			log.Stack(fmt.Sprint(dbReq.request.CollectName, " has Cache But CacheId is 0"))
		}
	}
}

// DoFindOneAndUpset 根据条件upset一行数据(有则插入，无则更新)，并查询出来
func (dbService *DBService) DoFindOneAndUpset(dbReq DBRequest) error {
	//1.选择数据库与表
	s := dbService.mongoModule.TakeSession()

	//2.设置数据
	if len(dbReq.request.RawData) == 0 && len(dbReq.request.Data) == 0 {
		err := fmt.Errorf("%s DoFindOneAndUpset data len is error %d.", dbReq.request.CollectName, len(dbReq.request.Data))
		log.SError(err.Error())
		dbService.responseRet(dbReq, err, 0, 0, 0, 0)
		return err
	}

	//3.设置条件
	var condition interface{}
	unmarshalErr := bson.Unmarshal(dbReq.request.GetCondition(), &condition)
	if unmarshalErr != nil {
		log.SError("bson.Unmarshal err ", unmarshalErr.Error())
		dbService.responseRet(dbReq, unmarshalErr, 0, 0, 0, 0)
		return unmarshalErr
	}

	var updateData any

	if len(dbReq.request.RawData) != 0 {
		unmarshalErr = bson.Unmarshal(dbReq.request.RawData[0], &updateData)
		if unmarshalErr != nil {
			log.SError("bson.Unmarshal err ", unmarshalErr.Error())
			dbService.responseRet(dbReq, unmarshalErr, 0, 0, 0, 0)
			return unmarshalErr
		}
	} else {
		var data any
		uErr := bson.Unmarshal(dbReq.request.Data[0], &data)
		if uErr != nil {
			uErr := fmt.Errorf("%s DoFindOneAndUpset data Unmarshal error %s.", dbReq.request.CollectName, uErr.Error())
			log.SError(uErr.Error())
			dbService.responseRet(dbReq, uErr, 0, 0, 0, 0)
			return uErr
		}
		//updateData = bson.M{"$set": bson.E{Key: "xxx", Value: 344}, "$setOnInsert": data}
		updateData = bson.M{"$set": data}
	}

	//之前有数据
	var sResult *mongo.SingleResult

	ctx, cancel := s.GetDefaultContext()
	defer cancel()
	after := options.After
	updateOpts := options.FindOneAndUpdateOptions{ReturnDocument: &after}
	updateOpts.SetUpsert(dbReq.request.Upsert)

	if len(dbReq.request.SelectField) > 0 {
		var selectField interface{}
		err := bson.Unmarshal(dbReq.request.SelectField, &selectField)
		if err != nil {
			uErr := fmt.Errorf("%s DoFindOneAndUpset data Unmarshal error %s.", dbReq.request.CollectName, err.Error())
			log.SError(uErr.Error())
			dbService.responseRet(dbReq, err, 0, 0, 0, 0)
			return uErr
		}
		updateOpts.SetProjection(selectField)
	}

	sResult = s.Collection(dbService.dbName, dbReq.request.GetCollectName()).FindOneAndUpdate(ctx, condition, updateData, &updateOpts)
	if sResult.Err() != nil {
		log.SError("DoFindOneAndUpset err ", sResult.Err().Error())
		dbService.responseRet(dbReq, sResult.Err(), 0, 0, 0, 0)
		return sResult.Err()
	}

	dataResult, err := sResult.DecodeBytes()
	if err != nil {
		log.SError("DoFindOneAndUpset err ", err.Error())
		dbService.responseRet(dbReq, err, 0, 0, 0, 0)
		return err
	}

	//7.获取结果集
	var dbRet db.DBControllerRet
	var rpcErr rpc.RpcError
	dbRet.Type = dbReq.request.Type
	if dataResult != nil {
		dbRet.Res = make([][]byte, 1)
		dbRet.Res[0] = dataResult
		dbRet.UpsertedCount = 1
	}

	dbReq.responder(&dbRet, rpcErr)
	return nil
}

// DoSetOnInsertFind 根据条件upset一行数据(有则插入，无则更新)，并查询出来
func (dbService *DBService) DoSetOnInsertFind(dbReq DBRequest) error {
	//1.选择数据库与表
	s := dbService.mongoModule.TakeSession()

	//2.设置数据
	if len(dbReq.request.RawData) != 1 {
		err := fmt.Errorf("%s DoFindOneAndUpset data len is error %d.", dbReq.request.CollectName, len(dbReq.request.Data))
		log.SError(err.Error())
		dbService.responseRet(dbReq, err, 0, 0, 0, 0)
		return err
	}

	//3.设置条件
	var condition interface{}
	unmarshalErr := bson.Unmarshal(dbReq.request.GetCondition(), &condition)
	if unmarshalErr != nil {
		log.SError("bson.Unmarshal err ", unmarshalErr.Error())
		dbService.responseRet(dbReq, unmarshalErr, 0, 0, 0, 0)
		return unmarshalErr
	}

	var rawData any
	unmarshalErr = bson.Unmarshal(dbReq.request.RawData[0], &rawData)
	if unmarshalErr != nil {
		log.SError("bson.Unmarshal err ", unmarshalErr.Error())
		dbService.responseRet(dbReq, unmarshalErr, 0, 0, 0, 0)
		return unmarshalErr
	}

	update := bson.M{"$setOnInsert": rawData}
	updateOpts := options.Update().SetUpsert(true)
	ctx, cancel := s.GetDefaultContext()
	defer cancel()
	updateResult, err := s.Collection(dbService.dbName, dbReq.request.GetCollectName()).UpdateOne(ctx, condition, update, updateOpts)
	if err != nil {
		log.SError("DoFindOneAndUpset err ", err.Error())
		dbService.responseRet(dbReq, err, 0, 0, 0, 0)
		return err
	}

	//之前有数据
	var sResult *mongo.SingleResult
	for {
		//如果已经存在数据，则查询并更新数据
		if (updateResult.MatchedCount > 0 || updateResult.ModifiedCount > 0) && len(dbReq.request.Data) != 0 {
			var data interface{}
			uErr := bson.Unmarshal(dbReq.request.Data[0], &data)
			if uErr != nil {
				uErr := fmt.Errorf("%s DoFindOneAndUpset data Unmarshal error %s.", dbReq.request.CollectName, uErr.Error())
				log.SError(uErr.Error())
				dbService.responseRet(dbReq, uErr, 0, 0, 0, 0)
				return uErr
			}
			//updateData = bson.M{"$set": bson.E{Key: "xxx", Value: 344}, "$setOnInsert": data}
			updateData := bson.M{"$set": data}

			ctx, cancel := s.GetDefaultContext()
			defer cancel()
			after := options.After
			updateOpts := options.FindOneAndUpdateOptions{ReturnDocument: &after}
			updateOpts.SetUpsert(true)
			sResult = s.Collection(dbService.dbName, dbReq.request.GetCollectName()).FindOneAndUpdate(ctx, condition, updateData, &updateOpts)
			break
		}

		//如果是插入数据，直接查询
		ctx, cancel := s.GetDefaultContext()
		defer cancel()
		sResult = s.Collection(dbService.dbName, dbReq.request.GetCollectName()).FindOne(ctx, condition)
		break
	}

	if sResult.Err() != nil {
		log.SError("DoFindOneAndUpset err ", sResult.Err().Error())
		dbService.responseRet(dbReq, sResult.Err(), 0, 0, 0, 0)
		return sResult.Err()
	}

	dataResult, err := sResult.DecodeBytes()
	if err != nil {
		log.SError("DoFindOneAndUpset err ", err.Error())
		dbService.responseRet(dbReq, err, 0, 0, 0, 0)
		return err
	}

	//7.获取结果集
	var dbRet db.DBControllerRet
	var rpcErr rpc.RpcError
	dbRet.Type = dbReq.request.Type
	dbRet.Res = make([][]byte, 1)
	dbRet.Res[0] = dataResult
	dbRet.UpsertedCount = 1
	dbReq.responder(&dbRet, rpcErr)
	return nil
}

func (dbService *DBService) DoCount(dbReq DBRequest) error {
	s := dbService.mongoModule.TakeSession()

	//设置条件
	var condition interface{}
	if dbReq.request.Condition == nil {
		dbReq.request.Condition, _ = bson.Marshal(bson.D{})
	}

	unmarshalErr := bson.Unmarshal(dbReq.request.GetCondition(), &condition)
	if unmarshalErr != nil {
		log.SError("bson.Unmarshal err ", unmarshalErr.Error())
		dbService.responseRet(dbReq, unmarshalErr, 0, 0, 0, 0)
		return unmarshalErr
	}

	num, err := s.CountDocument(dbService.dbName, dbReq.request.GetCollectName(), condition)
	var dbRet db.DBControllerRet
	var rpcErr rpc.RpcError

	if err != nil {
		dbRet.MatchedCount = -1
		rpcErr = rpc.RpcError(err.Error())
		log.SError("count collect ", dbReq.request.GetCollectName(), " is error")
		dbReq.responder(&dbRet, rpcErr)
		return err
	}

	dbRet.MatchedCount = int32(num)
	dbReq.responder(&dbRet, rpcErr)
	return nil
}

// DoDelOne 执行方法——删除一条
func (dbService *DBService) DoDelOne(dbReq DBRequest) error {
	//1.选择数据库与表
	s := dbService.mongoModule.TakeSession()

	//2.设置条件
	var condition interface{}
	unmarshalErr := bson.Unmarshal(dbReq.request.GetCondition(), &condition)
	if unmarshalErr != nil {
		log.SError("bson.Unmarshal err ", unmarshalErr.Error())
		dbService.responseRet(dbReq, unmarshalErr, 0, 0, 0, 0)
		return unmarshalErr
	}

	ctx, cancel := s.GetDefaultContext()
	defer cancel()
	ret, err := s.Collection(dbService.dbName, dbReq.request.GetCollectName()).DeleteOne(ctx, condition)
	if err != nil {
		log.SError(dbReq.request.CollectName, " DoDelOne fail error ", err.Error())
	}

	deletedCount := int32(0)
	if ret != nil {
		deletedCount = int32(ret.DeletedCount)
	}

	dbService.responseRet(dbReq, err, 0, 0, 0, deletedCount)
	return nil
}

// DoDelMany 执行方法——删除匹配条件的所有数据
func (dbService *DBService) DoDelMany(dbReq DBRequest) error {
	//1.选择数据库与表
	s := dbService.mongoModule.TakeSession()

	//2.设置条件
	var condition interface{}
	unmarshalErr := bson.Unmarshal(dbReq.request.GetCondition(), &condition)
	if unmarshalErr != nil {
		log.SError("bson.Unmarshal err ", unmarshalErr.Error())
		dbService.responseRet(dbReq, unmarshalErr, 0, 0, 0, 0)
		return unmarshalErr
	}

	ctx, cancel := s.GetDefaultContext()
	defer cancel()
	ret, err := s.Collection(dbService.dbName, dbReq.request.GetCollectName()).DeleteMany(ctx, condition)
	if err != nil {
		log.SError(dbReq.request.CollectName, " DoDelOne fail error ", err.Error())
	}

	deletedCount := int32(0)
	if ret != nil {
		deletedCount = int32(ret.DeletedCount)
	}

	dbService.responseRet(dbReq, err, 0, 0, 0, deletedCount)
	return nil
}

// DoUpdateMany 执行方法--批量更新
func (dbService *DBService) DoUpdateMany(dbReq DBRequest) error {
	//1.选择数据库与表
	s := dbService.mongoModule.TakeSession()

	//2.设置数据
	if len(dbReq.request.Data) != 0 && len(dbReq.request.RawData) != 0 {
		err := fmt.Errorf("%s DoUpdateMany data len is error %d.", dbReq.request.CollectName, len(dbReq.request.Data))
		log.SError(err.Error())
		dbService.responseRet(dbReq, err, 0, 0, 0, 0)
		return err
	}

	//3.设置条件
	var condition interface{}
	unmarshalErr := bson.Unmarshal(dbReq.request.GetCondition(), &condition)
	if unmarshalErr != nil {
		log.SError("bson.Unmarshal err ", unmarshalErr.Error())
		dbService.responseRet(dbReq, unmarshalErr, 0, 0, 0, 0)
		return unmarshalErr
	}

	var updateData any
	if len(dbReq.request.Data) != 0 {
		var data interface{}
		err := bson.Unmarshal(dbReq.request.Data[0], &data)
		if err != nil {
			err := fmt.Errorf("%s DoUpdateMany data Unmarshal error %s.", dbReq.request.CollectName, err.Error())
			log.SError(err.Error())
			dbService.responseRet(dbReq, err, 0, 0, 0, 0)
			return err
		}
		updateData = bson.M{"$set": data}
	} else {
		err := bson.Unmarshal(dbReq.request.RawData[0], &updateData)
		if err != nil {
			errs := fmt.Errorf("%s DoUpdateMany data Unmarshal RawData error %s.", dbReq.request.CollectName, err.Error())
			log.SError(errs.Error())
			dbService.responseRet(dbReq, errs, 0, 0, 0, 0)
			return errs
		}
	}

	var UpdateOptionsOpts []*options.UpdateOptions
	//if upset == true {
	//	UpdateOptionsOpts = append(UpdateOptionsOpts, options.Update().SetUpsert(true))
	//}

	ctx, cancel := s.GetDefaultContext()
	defer cancel()

	ret, err := s.Collection(dbService.dbName, dbReq.request.GetCollectName()).UpdateMany(ctx, condition, updateData, UpdateOptionsOpts...)
	//changeInfo, err := collect.Upsert(condition, bson.M{"$setOnInsert": data})

	if dbReq.responder.IsInvalid() == false {
		matchedCount, modifiedCount, upsertedCount := int32(0), int32(0), int32(0)
		if ret != nil {
			matchedCount, modifiedCount, upsertedCount = int32(ret.MatchedCount), int32(ret.ModifiedCount), int32(ret.UpsertedCount)
		}

		dbService.responseRet(dbReq, err, matchedCount, modifiedCount, upsertedCount, 0)
	}
	return nil
}

// DoUpdate 执行方法——更新数据
func (dbService *DBService) DoUpdate(dbReq DBRequest) error {
	return dbService.update(dbReq, false)
}

func (dbService *DBService) DoUpdateField(dbReq DBRequest) error {
	return dbService.updateField(dbReq)
}

func (dbService *DBService) CompressBlock(src, dst []byte) (cnt int, err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			errString := fmt.Sprint(r)
			err = errors.New("core dump info[" + errString + "]\n" + string(buf[:l]))
		}
	}()

	var c lz4.Compressor
	cnt, err = c.CompressBlock(src, dst)
	return
}

func (dbService *DBService) UncompressBlock(src, dst []byte) (cnt int, err error) {
	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			l := runtime.Stack(buf, false)
			errString := fmt.Sprint(r)
			err = errors.New("core dump info[" + errString + "]\n" + string(buf[:l]))
		}
	}()

	cnt, err = lz4.UncompressBlock(src, dst)
	return
}

func (dbService *DBService) updateDBDataToCache(collectName string, cacheId uint64, data [][]byte, version int32) {
	if cacheId == 0 || dbService.MaxCacheCap == 0 {
		// 这里加一个防御性报错
		if cacheId == 0 {
			mapCache, _ := dbService.GetMapCacheByCollectName(collectName, false)
			if mapCache != nil {
				log.Stack(collectName + " has cache, but cacheId is 0")
			}
		}
		return
	}

	mapCache, _ := dbService.GetMapCacheByCollectName(collectName, true)

	//压缩
	if dbService.CacheCompress == false {
		mapCache.UpsertData(cacheId, data, version)
		return
	}

	var compressByte [][]byte
	for _, rowData := range data {
		dest := make([]byte, lz4.CompressBlockBound(len(rowData))+1)
		cnt, err := dbService.CompressBlock(rowData, dest[1:])
		if err != nil || cnt == 0 {
			log.SError("compress block collectName:", collectName, " cacheId:", cacheId, " err:", err.Error())
			//异常情况，删除缓存
			mapCache.RemoveCache(cacheId)
			return
		}

		//没有压缩
		if cnt >= len(rowData) {
			dest[0] = 0 //标记是否有压缩
			if copy(dest[1:], rowData) != len(rowData) {
				log.SError("compress block collectName:", collectName, " cacheId:", cacheId, " copy error")
				//异常情况，删除缓存
				mapCache.RemoveCache(cacheId)
				return
			}
			compressByte = append(compressByte, dest[:len(rowData)+1])
		} else {
			dest[0] = 1 //有压缩
			compressByte = append(compressByte, dest[:cnt+1])
			/*
				var bys [409600]byte
				cnt, err = dbService.UncompressBlock(dest[1:cnt+1], bys[:])
				if err != nil {
					log.SError("UncompressBlock collectName ", collectName, " cacheId ", cacheId, " error:", err.Error(), " src len:", len(dest[1:cnt+1]))
				}*/
		}
	}

	mapCache.UpsertData(cacheId, compressByte, version)
	//log.SDebug(">+++============Update-:", collectName, "  cacheId", cacheId)
}

func (dbService *DBService) updateCache(dbReq DBRequest) {
	//只有缓存Id>0时，才需要缓存
	dbService.updateDBDataToCache(dbReq.request.CollectName, dbReq.request.CacheId, dbReq.request.Data, dbReq.request.Version)
}

// 对doc里面字段有fields有设置相同的值进行inc或者set
func (dbService *DBService) setDoc(doc primitive.D, fields bson.M, inc bool) error {
	findCount := 0
	var err error
	for i := 0; i < len(doc); i++ {
		findVal, bfind := fields[doc[i].Key]
		if bfind == false {
			continue
		}

		findCount++
		if inc == true {
			doc[i].Value, err = NumberAdd(doc[i].Value, findVal)
			if err != nil {
				return err
			}
			continue
		}

		doc[i].Value = findVal
	}

	if findCount != len(fields) {
		return errors.New("Non-existent fields in the cache")
	}

	return nil
}

func NumberAdd(srcVal any, addVal any) (any, error) {

	switch v := srcVal.(type) {
	case int:
		av, ok := addVal.(int)
		if ok == false {
			break
		}
		return v + av, nil
	case uint:
		av, ok := addVal.(uint)
		if ok == false {
			break
		}
		return v + av, nil
	case int64:
		av, ok := addVal.(int64)
		if ok == false {
			break
		}
		return v + av, nil
	case uint64:
		av, ok := addVal.(uint64)
		if ok == false {
			break
		}
		return v + av, nil
	case uint8:
		av, ok := addVal.(uint8)
		if ok == false {
			break
		}
		return v + av, nil
	case uint16:
		av, ok := addVal.(uint16)
		if ok == false {
			break
		}
		return v + av, nil
	case uint32:
		av, ok := addVal.(uint32)
		if ok == false {
			break
		}
		return v + av, nil
	case int8:
		av, ok := addVal.(int8)
		if ok == false {
			break
		}
		return v + av, nil
	case int16:
		av, ok := addVal.(int16)
		if ok == false {
			break
		}
		return v + av, nil
	case int32:
		av, ok := addVal.(int32)
		if ok == false {
			break
		}
		return v + av, nil
	case float64:
		av, ok := addVal.(float64)
		if ok == false {
			break
		}
		return v + av, nil
	case float32:
		av, ok := addVal.(float32)
		if ok == false {
			break
		}
		return v + av, nil
	default:
		return nil, errors.New("type not supported")
	}

	return nil, errors.New("type not supported")
}

func (dbService *DBService) updateFieldByFieldByte(collectName string, fieldVal []byte, src []byte) ([]byte, error) {
	var val interface{}
	err := bson.Unmarshal(src, &val)
	if err != nil {
		log.SError("Unmarshal fail collection ", collectName)
		//mapCache.RemoveCache(cacheId)
		return nil, errors.New("Unmarshal fail collection ")
	}

	document, okD := val.(primitive.D)
	if okD == false {
		log.SError("convert fail collection ", collectName)
		//mapCache.RemoveCache(cacheId)
		return nil, errors.New("convert fail collection ")
	}

	var fields bson.M
	bson.Unmarshal(fieldVal, &fields)

	mapFields, ok := fields["$set"]
	if ok == true {
		mfield, cok := mapFields.(bson.M)
		if cok == false {
			log.Error("$set not support data format", collectName)
		} else {
			err = dbService.setDoc(document, mfield, false)
			if err != nil {
				log.Error("setDoc fail", log.ErrorAttr("error", err), log.Any("field", mfield), "collection", collectName)
				return nil, err
			}
		}
	}

	mapFields, ok = fields["$inc"]
	if ok == true {
		mfield, cok := mapFields.(bson.M)
		if cok == false {
			log.Error("$inc not support data format", collectName)
		} else {
			err = dbService.setDoc(document, mfield, true)
			if err != nil {
				log.Error("setDoc fail", log.ErrorAttr("error", err), log.Any("field", mfield), "collection", collectName)
				return nil, err
			}
		}
	}

	byteData, err := bson.Marshal(document)
	if err != nil {
		nErr := fmt.Errorf("Marshal fail collection %s,error:%s", collectName, err.Error())
		log.Error(nErr.Error())
		//mapCache.RemoveCache(cacheId)
		return nil, nErr
	}

	return byteData, nil
}

func (dbService *DBService) selectFieldByRowByteData(collectName string, src []byte, mapSelectField map[string]*db.Placeholder) ([]byte, bool) {
	//如果需要选择列，重新解数据
	var val interface{}
	err := bson.Unmarshal(src, &val)
	if err != nil {
		log.SError("Unmarshal fail collection ", collectName)
		//mapCache.RemoveCache(cacheId)
		return nil, false
	}

	document, okD := val.(primitive.D)
	if okD == false {
		log.SError("convert fail collection ", collectName)
		//mapCache.RemoveCache(cacheId)
		return nil, false
	}

	retDocument := util.PickSlice(document, func(pickElement any) bool {
		e, okE := pickElement.(primitive.E)
		if okE == false {
			return false
		}

		_, hasKey := mapSelectField[e.Key]
		return hasKey
	})

	byteData, err := bson.Marshal(retDocument)
	if err != nil {
		log.SError("Marshal fail collection ", collectName, " err:", err.Error())
		//mapCache.RemoveCache(cacheId)
		return nil, false
	}

	return byteData, true
}

func (dbService *DBService) GetMapCacheByCollectName(collectName string, needCreate bool) (*util.FCMap, bool) {
	dbService.MapCacheLock.Lock()
	defer dbService.MapCacheLock.Unlock()

	mapCache, ok := dbService.MapCache[collectName]
	if ok == false || mapCache == nil {
		if !needCreate {
			return nil, false
		}

		mapCache = &util.FCMap{}
		mapCache.Init(dbService.MaxCacheCap, dbService.ExpirationTimeSecond, dbService.CheckIntervalSecond, dbService.IntervalCheckNum)
		dbService.MapCache[collectName] = mapCache

		return mapCache, false
	}

	return mapCache, true
}

// findCache 从缓存中查找数据指定数据并选择字段列
func (dbService *DBService) findCache(collectName string, cacheId string, mapSelectField map[string]*db.Placeholder) (bool, [][]byte) {
	if cacheId == "" || dbService.MaxCacheCap == 0 {
		if cacheId == "" {
			// 这里加一个防御性报错
			mapCache, _ := dbService.GetMapCacheByCollectName(collectName, false)
			if mapCache != nil {
				log.Stack(collectName + " has cache, but cacheId is 0")
			}
		}
		return false, nil
	}

	//1.找不到时，也先建立，以免，数据协程定入数据时，对该MapCache产生读写冲突
	mapCache, ok := dbService.GetMapCacheByCollectName(collectName, true)
	if ok == false {
		return false, nil
	}

	//2.从缓存对象中查找
	data := mapCache.FindData(cacheId)
	if data == nil {
		return false, nil
	}

	//3.数据转换
	var cacheData [][]byte
	cacheData, ok = data.([][]byte)
	if ok == false {
		log.SError("cannot convert data ", collectName)
		mapCache.RemoveCache(cacheId)
		return false, nil
	}

	//4.如果缓存被压缩，则需要解压
	if dbService.CacheCompress {
		var retData [][]byte
		var tmpByteBuff [409600]byte

		//遍历所有的缓存数据
		for _, rowData := range cacheData {
			//dest := dbService.compressBuff
			//长度必需为1，因为第0个位置存放是否被压缩
			if len(rowData) == 0 {
				log.SError("rowData is error :", collectName, " cacheId:", cacheId)
				mapCache.RemoveCache(cacheId)
				return false, nil
			}

			var dest []byte
			if rowData[0] == 0 {
				dest = rowData[1:]
			} else {
				uncompressDest := tmpByteBuff[:] // make([]byte, 10*len(rowData))
				cnt, err := dbService.UncompressBlock(rowData[1:], uncompressDest)
				if err != nil {
					log.SError("UncompressBlock fail collection:", collectName, " cacheId:", cacheId, " error:", err.Error(), " src len:", len(rowData[1:]), " dest len:", len(uncompressDest))
					mapCache.RemoveCache(cacheId)
					return false, nil
				}

				dest = make([]byte, cnt)
				if copy(dest, uncompressDest[:cnt]) != cnt {
					log.SError("copy fail!")
					mapCache.RemoveCache(cacheId)
					return false, nil
				}
			}

			//数据没有字段选择，就直接copy数据
			if len(mapSelectField) == 0 {
				retData = append(retData, dest)
				continue
			}

			//从库存数据中选择指定字段，并序列化返回
			byteData, ok := dbService.selectFieldByRowByteData(collectName, dest, mapSelectField)
			if ok == false {
				mapCache.RemoveCache(cacheId)
				return false, nil
			}

			retData = append(retData, byteData)
		}

		return true, retData
	}

	//如果是没有压缩的情况，又需要选择字段
	if len(mapSelectField) > 0 {
		var retData [][]byte
		for _, rowData := range cacheData {
			byteData, ok := dbService.selectFieldByRowByteData(collectName, rowData, mapSelectField)
			if ok == false {
				mapCache.RemoveCache(cacheId)
				return false, nil
			}

			retData = append(retData, byteData)
		}

		return true, retData
	}

	//没有压缩，也没有选择字段，直接返回
	return true, cacheData
}

func (dbService *DBService) getDocumentFieldVal(fieldName string, rowData interface{}) (error, interface{}) {
	document, ok := rowData.(primitive.D)
	if ok == false {
		return errors.New("document data format is error " + fieldName), nil
	}

	for _, elementField := range document {
		if elementField.Key == fieldName {
			return nil, elementField.Value
		}
	}

	return errors.New("cannot find field " + fieldName + " from document"), nil
}

func (dbService *DBService) marshalRowData(rowData interface{}, mapSelectField map[string]*db.Placeholder) ([]byte, error) {
	//如果没有筛选，直接返回
	if len(mapSelectField) == 0 {
		return bson.Marshal(rowData)
	}

	//
	document, ok := rowData.(primitive.D)
	if ok == false {
		return nil, errors.New("document data format is error ")
	}

	ret := util.PickSlice(document, func(pickElement any) bool {
		e, ok := pickElement.(primitive.E)
		if ok == false {
			log.Stack("cannot convert data.")
			return false
		}

		_, ok = mapSelectField[e.Key]
		return ok
	})

	return bson.Marshal(ret)
}

// DoFindManyKey 通过许多key进行查找
func (dbService *DBService) DoFindManyKey(dbReq DBRequest) error {
	var dbRet db.DBControllerRet

	//1.校验不应该发生的数据错误
	if dbReq.request.ManyKeyCondition == nil || len(dbReq.request.ManyKeyCondition.Key) > MaxKeyNum {
		var err error
		err = errors.New("DoFindManyKey fail,collect name " + dbReq.request.CollectName + " param is error!")

		var rpcErr rpc.RpcError
		rpcErr = rpc.RpcError(err.Error())
		dbReq.responder(&dbRet, rpcErr)

		log.SError(err.Error())
		return err
	}

	//2.从缓存中取出所有的key，取不到存放到noCacheKey中
	noCacheKey := make([]string, 0, MaxKeyNum)
	for _, findKey := range dbReq.request.ManyKeyCondition.Key {
		findRet, findData := dbService.findCache(dbReq.request.CollectName, findKey, dbReq.request.ManyKeyCondition.SelectField)
		if findRet == false {
			noCacheKey = append(noCacheKey, findKey)
			continue
		}

		dbRet.MatchedCount += int32(len(findData))
		dbRet.ManyKeyData = append(dbRet.ManyKeyData, &db.FindManyKeyData{Key: findKey, Data: findData})
	}

	//3.组装查询所需参数
	var findOptions []*options.FindOptions
	if dbReq.request.GetMaxRow() > 0 {
		var findOption options.FindOptions
		findOption.SetLimit(int64(dbReq.request.GetMaxRow()))

		findOptions = append(findOptions, &findOption)
	}

	if len(noCacheKey) == 0 {
		var rpcErr rpc.RpcError
		dbReq.responder(&dbRet, rpcErr)
		return rpcErr
	}

	//5.从数据库查询
	condition := bson.D{{Key: dbReq.request.ManyKeyCondition.ConditionField, Value: bson.M{"$in": noCacheKey}}}
	s := dbService.mongoModule.TakeSession()
	ctx, cancel := s.GetDefaultContext()
	defer cancel()
	cursor, err := s.Collection(dbService.dbName, dbReq.request.GetCollectName()).Find(ctx, condition, findOptions...)
	if err != nil || cursor.Err() != nil {
		if err == nil {
			err = cursor.Err()
		}

		dbRet.MatchedCount = int32(len(dbRet.ManyKeyData))
		dbRet.Res = emptyRes
		dbReq.responder(&dbRet, rpc.RpcError(err.Error()))
		return err
	}

	var res []interface{}
	ctxAll, cancelAll := s.GetDefaultContext()
	defer cancelAll()
	err = cursor.All(ctxAll, &res)
	if err != nil {
		dbRet.Res = emptyRes
		dbReq.responder(&dbRet, rpc.RpcError(err.Error()))
		return err
	}

	//6.获取结果集
	//序列化结果
	mapCacheData := make(map[uint64][][]byte, MaxKeyNum)
	fullMapCacheData := make(map[uint64][][]byte, MaxKeyNum)
	dbRet.Type = dbReq.request.Type
	for i := 0; i < len(res); i++ {
		//从文档中获取条件字段值
		errD, val := dbService.getDocumentFieldVal(dbReq.request.ManyKeyCondition.ConditionField, res[i])
		//反射获取字段值
		if errD != nil {
			log.SError("field type is error ", dbReq.request.ManyKeyCondition.ConditionField)
			continue
		}

		errK, keyId := util.ConvertToNumber[uint64](val)
		if errK != nil {
			log.SError("field type is error ", dbReq.request.ManyKeyCondition.ConditionField)
			continue
		}

		allByteRet, errRet := bson.Marshal(res[i])
		if errRet != nil {
			log.SError("Marshal ", dbReq.request.CollectName, " fail,key is ", keyId)
			continue
		}

		//从文档中选择指定字段
		byteRet, errRet := dbService.marshalRowData(res[i], dbReq.request.ManyKeyCondition.SelectField)
		if errRet != nil {
			log.SError("Marshal ", dbReq.request.CollectName, " fail,key is ", keyId)
			continue
		}

		//暂存返回数据和完整数据
		rowData := mapCacheData[keyId]         //返回数据
		fullRowData := fullMapCacheData[keyId] //完整数据

		rowData = append(rowData, byteRet)
		fullRowData = append(fullRowData, allByteRet)
		mapCacheData[keyId] = rowData
		fullMapCacheData[keyId] = fullRowData
	}

	//7.组装返回数据
	for key, cacheData := range mapCacheData {
		dbRet.MatchedCount += int32(len(cacheData))
		dbRet.ManyKeyData = append(dbRet.ManyKeyData, &db.FindManyKeyData{Key: key, Data: cacheData})
	}

	//8.完整数据存缓存
	for key, cacheData := range fullMapCacheData {
		//更新到缓存
		dbService.updateDBDataToCache(dbReq.request.CollectName, key, cacheData, dbReq.request.Version)
	}

	//9.异步返回
	var rpcErr rpc.RpcError
	dbReq.responder(&dbRet, rpcErr)
	return nil
}

// DoFind 执行方法——查找
func (dbService *DBService) DoFind(dbReq DBRequest) error {
	if !dbReq.request.NotUseCache {
		// 如果使用Cache就查Cache
		findRet, findData := dbService.findCache(dbReq.request.CollectName, dbReq.request.CacheId, nil)
		if findRet == true {
			var dbRet db.DBControllerRet
			dbRet.Res = findData
			var rpcErr rpc.RpcError
			dbRet.MatchedCount = int32(len(dbRet.Res))
			dbReq.responder(&dbRet, rpcErr)
			return nil
		}
	}

	//1.选择数据库与表
	s := dbService.mongoModule.TakeSession()

	//2.设置条件
	var dbRet db.DBControllerRet
	var condition interface{}
	err := bson.Unmarshal(dbReq.request.GetCondition(), &condition)
	if err != nil {
		dbRet.Res = emptyRes
		dbReq.responder(&dbRet, rpc.RpcError(err.Error()))
		return err
	}

	var findOptions []*options.FindOptions
	if len(dbReq.request.Sort) > 0 {
		var sorts bson.D = make([]bson.E, len(dbReq.request.Sort))

		for idx, s := range dbReq.request.Sort {
			i := -1
			if s.Asc == true {
				i = 1
			}
			sorts[idx] = bson.E{Key: s.SortField, Value: i}
		}
		var findOption options.FindOptions
		findOption.SetSort(sorts)
		findOptions = append(findOptions, &findOption)
	}

	//如果选择字段
	if len(dbReq.request.SelectField) > 0 {
		var selectField interface{}
		err := bson.Unmarshal(dbReq.request.GetSelectField(), &selectField)
		if err != nil {
			dbRet.Res = emptyRes
			dbReq.responder(&dbRet, rpc.RpcError(err.Error()))
			return err
		}

		var findOption options.FindOptions
		findOption.SetProjection(selectField)
		findOptions = append(findOptions, &findOption)
	}

	if dbReq.request.GetMaxRow() > 0 {
		var findOption options.FindOptions
		findOption.SetLimit(int64(dbReq.request.GetMaxRow()))

		findOptions = append(findOptions, &findOption)
	}

	//设置跳过——分页
	if dbReq.request.GetSkip() > 0 {
		var findOption options.FindOptions
		findOption.SetSkip(int64(dbReq.request.GetSkip()))

		findOptions = append(findOptions, &findOption)
	}

	ctx, cancel := s.GetDefaultContext()
	defer cancel()
	cursor, err := s.Collection(dbService.dbName, dbReq.request.GetCollectName()).Find(ctx, condition, findOptions...)
	if err != nil || cursor.Err() != nil {
		if err == nil {
			err = cursor.Err()
		}
		dbRet.Res = emptyRes
		dbReq.responder(&dbRet, rpc.RpcError(err.Error()))
		return err
	}

	if dbReq.request.QueryDocumentCount {
		allCount, err := s.Collection(dbService.dbName, dbReq.request.GetCollectName()).EstimatedDocumentCount(ctx)
		if err == nil {
			dbRet.DocumentCount = allCount
		} else {
			log.SError(dbService.dbName, " ", dbReq.request.GetCollectName(), " EstimatedDocumentCount error:", err.Error())
		}
	}

	var res []interface{}
	ctxAll, cancelAll := s.GetDefaultContext()
	defer cancelAll()
	err = cursor.All(ctxAll, &res)
	if err != nil {
		dbRet.Res = emptyRes
		dbReq.responder(&dbRet, rpc.RpcError(err.Error()))
		return err
	}

	//5.获取结果集
	var rpcErr rpc.RpcError
	//序列化结果
	dbRet.Type = dbReq.request.Type
	dbRet.Res = make([][]byte, len(res))
	for i := 0; i < len(res); i++ {
		dbRet.Res[i], err = bson.Marshal(res[i])
		if err != nil {
			rpcErr = rpc.RpcError(err.Error())
			dbRet.Res = emptyRes
			break
		}
	}

	if !dbReq.request.NotUseCache {
		//更新到缓存
		dbService.updateDBDataToCache(dbReq.request.CollectName, dbReq.request.CacheId, dbRet.Res, dbReq.request.Version)
	}

	dbRet.MatchedCount = int32(len(res))
	dbReq.responder(&dbRet, rpcErr)
	return nil
}

func (dbService *DBService) responseRet(dbReq DBRequest, err error, matchedCount int32, modifiedCount int32, upsertedCount int32, deleteCount int32) {
	var dbRet db.DBControllerRet

	dbRet.MatchedCount = matchedCount
	dbRet.ModifiedCount = modifiedCount
	dbRet.UpsertedCount = upsertedCount
	dbRet.DeletedCount = deleteCount
	if dbReq.responder.IsInvalid() == false {
		if err == nil {
			dbReq.responder(&dbRet, rpc.NilError)
		} else {
			dbReq.responder(&dbRet, rpc.RpcError(err.Error()))
		}

	}
}

// DoInsert 执行方法——插入数据
func (dbService *DBService) DoInsert(dbReq DBRequest) error {
	//1.选择数据库与表
	s := dbService.mongoModule.TakeSession()

	var data []interface{}
	data = make([]interface{}, len(dbReq.request.Data))
	for i := 0; i < len(data); i++ {
		err := bson.Unmarshal(dbReq.request.Data[i], &data[i])
		if err != nil {
			err := fmt.Errorf("%s DoInsert fail %s", dbReq.request.CollectName, err.Error())
			dbService.responseRet(dbReq, err, 0, 0, 0, 0)
			return err
		}
	}

	ctx, cancel := s.GetDefaultContext()
	defer cancel()
	res, err := s.Collection(dbService.dbName, dbReq.request.GetCollectName()).InsertMany(ctx, data, options.InsertMany().SetOrdered(dbReq.request.Ordered))
	if err != nil {
		log.SError(dbReq.request.CollectName, " DoInsert fail error ", err.Error())
	}

	insertedCount := int32(0)
	if res != nil {
		insertedCount = int32(len(res.InsertedIDs))
	}
	dbService.responseRet(dbReq, err, 0, 0, insertedCount, 0)
	return err
}

type DBRequest struct {
	request   *db.DBControllerReq
	redisReq  *db.DBControllerRedisReq
	responder rpc.Responder
}

func (dbService *DBService) RPC_DBCheckLink(arg *bool, ret *bool) error {
	*ret = true
	return nil
}

func (dbService *DBService) RPC_GsReleaseDBCheckLink(arg *bool, ret *bool) error {
	*ret = true
	return nil
}

// RPC_DBRequest 对外提供的RPC方法，接收DB操作数据后，写入管道，交给执行协程执行
func (dbService *DBService) RPC_RedisRequest(responder rpc.Responder, request *db.DBControllerRedisReq) error {
	index := uint64(0)
	if request.ModKey > 0 {
		index = request.ModKey % uint64(dbService.goroutineNum-1)
		index += 1
	}

	if len(dbService.channelOptData[index]) == cap(dbService.channelOptData[index]) {
		log.SError("channel is full ", index)

		responder(nil, rpc.RpcError("channel is full"))
		return nil
	}

	var dbRequest DBRequest
	dbRequest.redisReq = request
	dbRequest.responder = responder

	dbService.channelOptData[index] <- dbRequest
	return nil
}

// RPC_DBRequest 对外提供的RPC方法，接收DB操作数据后，写入管道，交给执行协程执行
func (dbService *DBService) RPC_DBRequest(responder rpc.Responder, request *db.DBControllerReq) error {
	index := uint64(0)
	if request.GetKey() > 0 {
		index = request.GetKey() % uint64(dbService.goroutineNum-1)
		index += 1
	}

	if len(dbService.channelOptData[index]) == cap(dbService.channelOptData[index]) {
		log.SError("channel is full ", index)

		responder(nil, rpc.RpcError("channel is full"))
		return nil
	}

	var dbRequest DBRequest
	dbRequest.request = request
	dbRequest.responder = responder

	dbService.channelOptData[index] <- dbRequest
	return nil
}

func (dbService *DBService) RPC_GetServiceInfo(inParam *[]byte, outParam *[]byte) (err error) {
	var DBInfo struct {
		DbName       string
		GoroutineNum uint32

		DBChannelCap      int
		ChannelOptDataLen int

		DbDealCount      int32
		DbAllCostTime    int64
		DbMaxCostTime    int64
		CfgSlowQueryTime int64

		ServiceChannelNum int
		ServiceTimerNum   int
	}

	DBInfo.DbName = dbService.dbName
	DBInfo.GoroutineNum = dbService.goroutineNum
	DBInfo.DBChannelCap = dbService.channelNum
	DBInfo.ChannelOptDataLen = len(dbService.channelOptData)

	DBInfo.DbDealCount = dbService.dbDealCount
	DBInfo.DbAllCostTime = dbService.dbAllCostTime
	DBInfo.DbMaxCostTime = dbService.dbMaxCostTime
	DBInfo.CfgSlowQueryTime = dbService.slowQueryTime

	DBInfo.ServiceChannelNum = dbService.GetServiceEventChannelNum()
	DBInfo.ServiceTimerNum = dbService.GetServiceTimerChannelNum()

	*outParam, err = json.Marshal(&DBInfo)
	return
}

func (dbService *DBService) DoRedisSetKey(dbReq DBRequest) error {
	for _, keyVal := range dbReq.redisReq.KeyValue {
		if keyVal == nil {
			continue
		}

		dbService.redisModule.SetString(keyVal.ReidsKey, keyVal.Data)
	}

	return nil
}

func (dbService *DBService) DoRedisGetKey(dbReq DBRequest) error {
	var dBControllerRedisRes db.DBControllerRedisRes
	dBControllerRedisRes.Ret = 0
	dBControllerRedisRes.KeyValue = make([]*db.RedisKeyValue, 0, 5)

	for _, keyVal := range dbReq.redisReq.KeyValue {
		if keyVal == nil {
			continue
		}

		val, err := dbService.redisModule.GetString(keyVal.ReidsKey)
		if err != nil {
			log.Error("DoRedisGetKey failed", log.String("ReidsKey", keyVal.ReidsKey))
			continue
		}

		var redisKeyValue db.RedisKeyValue
		redisKeyValue.ReidsKey = keyVal.ReidsKey
		redisKeyValue.Data = []byte(val)
		dBControllerRedisRes.KeyValue = append(dBControllerRedisRes.KeyValue, &redisKeyValue)
	}

	dbReq.responder(&dBControllerRedisRes, "")
	return nil
}

func (dbService *DBService) DoRedisDelKey(dbReq DBRequest) error {
	for _, keyVal := range dbReq.redisReq.KeyValue {
		if keyVal == nil {
			continue
		}

		err := dbService.redisModule.DelString(keyVal.ReidsKey)
		if err != nil {
			log.Error("DoRedisGetKey failed", log.String("ReidsKey", keyVal.ReidsKey))
		}
	}

	var dBControllerRedisRes db.DBControllerRedisRes
	dBControllerRedisRes.Ret = 0
	dbReq.responder(&dBControllerRedisRes, "")

	return nil
}
