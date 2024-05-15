package collect

import (
	"go.mongodb.org/mongo-driver/bson"
	"origingame/common/db"
)

type CShowAreaInfo struct {
	BaseMultiCollection `bson:"-"`
	ShowAreaId          int    `bson:"_id"`        //显示区服ID
	AreaName            string `bson:"AreaName"`   //区服名称
	RealAreaId          int    `bson:"RealAreaId"` //真实区服ID

	ServerMark   EServerMark   `bson:"serverMark"`   //区服标签 0:普通 1:新服 2:推荐
	ServerStatus EServerStatus `bson:"serverStatus"` //区服状态 0:正常 1:维护

	OpenTime    string `bson:"openTime"`    //开服时间
	DefaultMark int    `bson:"defaultMark"` //是否为默认显示的服务器?

}

var ShowAreaInfoDBName = "ShowAreaInfo"

func (sa *CShowAreaInfo) GetCollName() string {
	return ShowAreaInfoDBName
}

func (sa *CShowAreaInfo) Clean() {
	sa.ShowAreaId = 0
	sa.RealAreaId = 0
	sa.AreaName = ""
}

func (sa *CShowAreaInfo) GetId() interface{} {
	return sa.ShowAreaId
}

func (sa *CShowAreaInfo) GetCollectionType() MultiCollectionType {
	return MCTShowArea
}

func (sa *CShowAreaInfo) MakeRow() IMultiCollection {
	return &CShowAreaInfo{}
}

func (sa *CShowAreaInfo) GetCondition(value interface{}) bson.D {
	return bson.D{}
}

func (sa *CShowAreaInfo) GetSort() []*db.Sort {
	return nil
}

func (sa *CShowAreaInfo) OnLoadSucc(notFound bool, userID uint64) {

}

func (sa *CShowAreaInfo) GetUpdateData() bson.M {
	return bson.M{}
}
