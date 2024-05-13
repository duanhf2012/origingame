package collect

import (
	"go.mongodb.org/mongo-driver/bson"
	"origingame/common/db"
)

type CShowAreaInfo struct {
	BaseMultiCollection `bson:"-"`
	ShowAreaId          int    `bson:"_id"`  //显示区服ID
	AreaName            string `bson:"name"` //区服名称
	RealAreaId          int    `bson:"rId"`  //真实区服ID

	ServerMark    EServerMark   `bson:"serverMark"`    //区服标签 0:普通 1:新服 2:推荐
	ServerStatus  EServerStatus `bson:"serverStatus"`  //区服状态 0:正常 1:维护
	MaxLoginCount int32         `bson:"maxLoginCount"` //在线人数上限
	MaxRegCount   int32         `bson:"maxRegCount"`   //注册人数上线

	OpenTime    string `bson:"openTime"`    //开服时间
	DealTime    string `bson:"dealTime"`    //操作时间
	CreateTime  string `bson:"createTime"`  //操作时间
	DefaultMark int    `bson:"defaultMark"` //是否为默认显示的服务器?
	MinVersion  string `bson:"minVersion"`  //最低版本号
	MaxVersion  string `bson:"maxVersion"`  //最高版本号

	//暂时保留a
	Label int `bson:"label"` //标签
	IsGm  int `bson:"isGm"`  //gm开关 0:关闭 1:开启

	OpenTimeMilli int64 `bson:"openTimeMilli"` //查询需要
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
