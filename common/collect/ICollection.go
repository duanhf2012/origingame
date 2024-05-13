package collect

import (
	"container/list"
	"go.mongodb.org/mongo-driver/bson"
	"origingame/common/db"
	"origingame/common/keyword"
)

type LoadCallBack func(collection ICollection) error

// ICollection MongoDB中仅一条数据使用，如用户数据
type ICollection interface {
	Clean()                                  //清理所有数据
	GetId() interface{}                      //获取主键值，一般是userID
	GetCollectionType() CollectionType       //获取数据类型
	GetCollName() string                     //获取MongoDB表名
	ClearDirty()                             //清脏
	MakeDirty()                              //数据置脏，确保每次修改数据都需要调用该接口
	IsDirty() bool                           //获取数据是否是脏数据
	GetSelf() ICollection                    //返回自身数据对象指针 ps：在MongoDB中仅一个数据使用，如玩家数据
	OnLoadSucc(notFound bool, userID string) //MongoDB数据加载成功后调用
	GetCondition(value interface{}) bson.D   //获取编辑查询条件
	GetSort() []*db.Sort                     //获取排序字段
	GetCacheId(cacheId uint64) uint64
	IsBuildInMemory() bool //表是从内存中构造的，不是从数据库载入的
	IsNeedWaitLoad() bool  //如果是玩家数据，登录的时候不用等待数据返回
}

type BaseCollection struct {
	Dirty bool `bson:"-"`
}

func (bc *BaseCollection) MakeDirty() {
	bc.Dirty = true
}

func (bc *BaseCollection) IsDirty() bool {
	return bc.Dirty
}

func (bc *BaseCollection) ClearDirty() {
	bc.Dirty = false
}

func (bc *BaseCollection) GetSort() []*db.Sort {
	return nil
}

func (bc *BaseCollection) GetCacheId(cacheId uint64) uint64 {
	return 0
}

func (bc *BaseCollection) IsBuildInMemory() bool {
	return false
}

func (bc *BaseCollection) IsNeedWaitLoad() bool {
	return true
}

// IMultiCollection MongoDB中多条数据时使用，如用户的mail数据
type IMultiCollection interface {
	Clean()                                  //清理所有数据
	GetId() interface{}                      //获取主键值，一般是userID
	GetCollectionType() MultiCollectionType  //获取数据类型
	GetCollName() string                     //获取MongoDB表名
	MakeRow() IMultiCollection               //返回一个新的数据对象指针 ps：在MongoDB中位多行数据使用，如用户的邮件数据
	OnLoadSucc(notFound bool, userID uint64) //MongoDB数据加载成功后调用
	GetCondition(value interface{}) bson.D   //获取编辑查询条件
	GetSort() []*db.Sort                     //获取排序字段
	GetUpdateData() bson.M                   //获取更新数据
	IsNeedWaitLoad() bool                    //如果是玩家数据，登录的时候不用等待数据返回
	GetMaxRowLimit() int32                   //查询返回的最大行数
}

type BaseMultiCollection struct {
}

func (bmc *BaseMultiCollection) GetMaxRowLimit() int32 {
	return keyword.MaxRowNum
}

func (bmc *BaseMultiCollection) IsNeedWaitLoad() bool {
	return true
}

// 多行数据
type MultiRowData struct {
	Template        IMultiCollection
	MapCollection   map[interface{}]*list.Element //key,value
	ListICollection list.List
}

func (multiRowData *MultiRowData) Clean() {
	multiRowData.Template = nil
	multiRowData.MapCollection = nil
	multiRowData.ListICollection = list.List{}
}

func (multiRowData *MultiRowData) LoadFromDB(res *db.DBControllerRet) error {
	if res == nil {
		return nil
	}

	for _, data := range res.Res {
		rowData := multiRowData.Template.MakeRow()
		err := bson.Unmarshal(data, rowData)
		if err != nil {
			return err
		}

		pElem := multiRowData.ListICollection.PushBack(rowData)
		multiRowData.MapCollection[rowData.GetId()] = pElem
	}

	return nil
}
