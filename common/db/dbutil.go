package db

import (
	"go.mongodb.org/mongo-driver/bson"
)

// 当为系统级的查询时，key填0
// 注意这里有一个规则
// 1. 用cache的话，cacheId必须>0
// 2. 用cache的表，不能delete操作，现在cache只用在player相关表里面，这些表都是只插不删除的

// fieldValue有两种设置方式
// 1. addCount := bson.M{"$inc": bson.M{"Count": 1}} 将字段+1
// 2. upsertCount := bson.M{"$set": bson.M{"Count": 1}} 将字段设置为1
func MakeCacheSetField(collName string, id string, key string, fieldValue bson.M, req *DBControllerReq) error {
	req.CollectName = collName
	req.Type = OptType_UpdateFieldOpt
	req.NotUseCache = false

	c, err := bson.Marshal(bson.D{{Key: "_id", Value: key}})
	if err != nil {
		return err
	}
	req.Condition = c
	req.Key = key
	req.CacheId = id

	data, err := bson.Marshal(fieldValue)
	if err != nil {
		return err
	}
	req.RawData = append(req.Data, data)

	return nil
}

func MakeCacheUpsetId(collName string, id interface{}, data interface{}, key string, cacheId string, req *DBControllerReq) error {
	return MakeCacheUpsetCondition(collName, bson.D{{Key: "_id", Value: id}}, data, key, cacheId, req)
}

func MakeCacheUpsetCondition(collName string, condition bson.D, data interface{}, key string, cacheId string, req *DBControllerReq) error {
	req.CollectName = collName
	req.Type = OptType_Upset
	req.Condition, _ = bson.Marshal(condition)
	req.Key = key
	req.CacheId = cacheId

	out, err := bson.Marshal(data)
	if err != nil {
		return err
	}
	req.Data = append(req.Data, out)
	return nil
}

func MakeUpsetId(collName string, id interface{}, data interface{}, key string, req *DBControllerReq) error {
	return MakeCacheUpsetId(collName, id, data, key, "", req)
}

func MakeUpsetWitchCondition(collName string, condition bson.D, data interface{}, key string, req *DBControllerReq) error {
	req.CollectName = collName
	req.Type = OptType_Upset
	req.Condition, _ = bson.Marshal(condition)
	req.Key = key
	out, err := bson.Marshal(data)
	if err != nil {
		return err
	}
	req.Data = append(req.Data, out)
	return nil
}

func MakeCountWithCondition(collName string, condition bson.D, key string, req *DBControllerReq) error {
	req.CollectName = collName
	req.Type = OptType_Count
	req.Condition, _ = bson.Marshal(condition)
	req.Key = key
	return nil
}

func MakeRemoveOneId(collName string, id interface{}, key string, req *DBControllerReq) error {
	var retErr error
	req.CollectName = collName
	req.Type = OptType_DelOne
	req.Condition, retErr = bson.Marshal(bson.D{{Key: "_id", Value: id}})
	req.Key = key

	return retErr
}

func MakeRemoveWithCondition(collName string, condition bson.D, key string, req *DBControllerReq) error {
	var retErr error
	req.CollectName = collName
	req.Type = OptType_DelOne
	req.Condition, retErr = bson.Marshal(condition)
	req.Key = key

	return retErr
}

// pagesize = 每页数量
// pageNum = 第几页（从1开始）
func MakeFindWithSkipPage(collName string, condition bson.D, key string, pageSize int64, pageNum int64, sort []*Sort, req *DBControllerReq) error {
	req.QueryDocumentCount = true
	req.Skip = int32(pageSize * (pageNum - 1))
	return MakeCacheFind(collName, condition, key, int32(pageSize), sort, "", req)
}

// 这里虽然走的是MakeCacheFind流程，不过cacheId=0，就是不缓存
func MakeFind(collName string, condition bson.D, key string, limit int32, sort []*Sort, req *DBControllerReq) error {
	return MakeCacheFind(collName, condition, key, limit, sort, "", req)
}

func MakeFindConditionSelectField(collName string, condition bson.D, key string, selectField bson.D, limit int32, sort []*Sort, req *DBControllerReq) error {
	req.CollectName = collName
	req.Type = OptType_Find
	c, err := bson.Marshal(condition)
	if err != nil {
		return err
	}
	req.Condition = c
	req.Key = key
	req.SelectField, err = bson.Marshal(selectField)
	if err != nil {
		return err
	}
	req.MaxRow = limit
	req.Sort = sort
	return nil
}

func MakeCacheFindManyKey(collName string, conditionField string, key string, IdList []string, selectField map[string]*Placeholder, limit int32, req *DBControllerReq) error {
	req.CollectName = collName
	req.Type = OptType_FindManyKey

	req.Key = key
	req.MaxRow = limit
	req.ManyKeyCondition = &FindManyKeyOption{ConditionField: conditionField, Key: IdList, SelectField: selectField}

	return nil
}

func MakeCacheFind(collName string, condition bson.D, key string, limit int32, sort []*Sort, cacheId string, req *DBControllerReq) error {
	req.CollectName = collName
	req.Type = OptType_Find
	c, err := bson.Marshal(condition)
	if err != nil {
		return err
	}
	req.Condition = c
	req.Key = key
	req.MaxRow = limit
	req.CacheId = cacheId
	req.Sort = sort

	return nil
}

func MakeUpdateId(collName string, data bson.D, id interface{}, key string, req *DBControllerReq) error {
	return MakeUpdate(collName, bson.D{{Key: "_id", Value: id}}, data, key, req)
}

func MakeUpdate(collName string, condition bson.D, data bson.D, key string, req *DBControllerReq) error {
	req.CollectName = collName
	req.Type = OptType_Update
	c, err := bson.Marshal(condition)
	if err != nil {
		return err
	}
	req.Condition = c
	req.Key = key
	out, err := bson.Marshal(data)
	if err != nil {
		return err
	}
	req.Data = append(req.Data, out)

	return nil
}

// 更新多行
func MakeUpdateMany(collName string, condition bson.D, data bson.M, key string, req *DBControllerReq) error {
	req.CollectName = collName
	req.Type = OptType_UpdateMany
	c, err := bson.Marshal(condition)
	if err != nil {
		return err
	}
	req.Condition = c
	out, err := bson.Marshal(data)
	if err != nil {
		return err
	}
	req.Data = make([][]byte, 0, 1)
	req.Data = append(req.Data, out)
	req.Key = key

	return nil
}

// 删除多行
func MakeDeleteMany(collName string, condition bson.D, key string, req *DBControllerReq) error {
	req.CollectName = collName
	req.Type = OptType_DelMany
	c, err := bson.Marshal(condition)
	if err != nil {
		return err
	}
	req.Condition = c
	req.Key = key

	return nil
}

func MakeSetOnInsertAndFind(collName string, id interface{}, updateInfo interface{}, key string, req *DBControllerReq) error {
	req.CollectName = collName
	req.Type = OptType_Upset
	req.Condition, _ = bson.Marshal(bson.D{{Key: "_id", Value: id}})
	req.Key = key
	out, err := bson.Marshal(updateInfo)
	if err != nil {
		return err
	}
	req.Data = append(req.Data, out)
	return nil
}

func MakeInsertId(collName string, data interface{}, key string, req *DBControllerReq) error {
	req.CollectName = collName
	req.Type = OptType_Insert
	req.Key = key
	out, err := bson.Marshal(data)
	if err != nil {
		return err
	}
	req.Data = append(req.Data, out)
	return nil
}

func MakeMultiInsertId(collName string, data [][]byte, key string, req *DBControllerReq) error {
	req.CollectName = collName
	req.Type = OptType_Insert
	req.Key = key
	req.Data = data
	return nil
}
