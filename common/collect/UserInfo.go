package collect

import (
	"github.com/duanhf2012/origin/v2/log"
	"go.mongodb.org/mongo-driver/bson"
	"origingame/common/keyword"
	global "origingame/common/keyword"
	"origingame/common/proto/msg"
	"time"
)

type CUserInfo struct {
	BaseCollection `bson:"-"`
	UserId         string `bson:"_id"`
	PlatId         string `bson:"PlatId"`
	ShowAreaId     int32  `bson:"ShowAreaId"`
	RegisterTime   int64  `bson:"RegisterTime"` //玩家注册时间
	SyncTime       int64  `bson:"SyncTime"`
	IsInit         bool   `bson:"IsInit"`
	IsGmAuth       bool   `bson:"IsGmAuth"`
	NickName       string `bson:"NickName"`

	MapUserAttribute map[global.AttributeType]int64 `bson:"Attr"`

	LastLogoutTime int64 `bson:"LastLogOutTime"` //最后一次离线时间
}

var UserInfoCollectName = "UserInfo"

func (userInfo *CUserInfo) GetCollName() string {
	return UserInfoCollectName
}

func (userInfo *CUserInfo) Clean() {
	*userInfo = CUserInfo{}
}

func (userInfo *CUserInfo) GetId() interface{} {
	return userInfo.UserId
}

func (userInfo *CUserInfo) GetCollectionType() CollectionType {
	return CTUserInfo
}

func (userInfo *CUserInfo) GetSelf() ICollection {
	return userInfo
}

func (userInfo *CUserInfo) GetCacheId(cacheId string) string {
	return cacheId
}

func (userInfo *CUserInfo) OnLoadSucc(notFound bool, userID string) {
	if notFound == true {
		userInfo.UserId = userID
	} else if userInfo.UserId == "" {
		log.SError("CUserInfo OnLoadSucc err:notFound, userID[", userID, "]")
	}

	if userInfo.MapUserAttribute == nil {
		userInfo.MapUserAttribute = make(map[global.AttributeType]int64, global.AttributeTypeSaveMaxLen)
	}

	sTime := time.Now()
	nowTime := sTime.UnixMilli()
	if userInfo.LastLogoutTime == 0 {
		userInfo.SetLastLogoutTime(nowTime)
	}
}

func (userInfo *CUserInfo) GetCondition(value interface{}) bson.D {
	return bson.D{{Key: "_id", Value: value}}
}

// 获取注册时间
func (userInfo *CUserInfo) GetRegisterTime() int64 {
	return userInfo.RegisterTime
}

// 设置注册时间
func (userInfo *CUserInfo) SetRegisterTime() {
	userInfo.RegisterTime = time.Now().UnixMilli()
	userInfo.MakeDirty()
}

func (userInfo *CUserInfo) SetSyncTime(syncTime int64) {
	userInfo.SyncTime = syncTime / 1e9
	userInfo.MakeDirty()
}

func (userInfo *CUserInfo) GetIsInit() bool {
	return userInfo.IsInit
}

func (userInfo *CUserInfo) GetIsGmAuth() bool {
	return userInfo.IsGmAuth
}

func (userInfo *CUserInfo) SetInit() {
	userInfo.IsInit = true
	userInfo.IsGmAuth = false
	userInfo.MakeDirty()
}

// SetNickName 修改昵称
func (userInfo *CUserInfo) SetNickName(nickName string) {
	userInfo.NickName = nickName
	userInfo.MakeDirty()
}

func (userInfo *CUserInfo) GetMapUserAttribute() map[keyword.AttributeType]int64 {
	return userInfo.MapUserAttribute
}

func (userInfo *CUserInfo) GetUserAttributeByType(attributeType global.AttributeType) (int64, bool) {
	value, ok := userInfo.MapUserAttribute[attributeType]
	return value, ok
}

func (userInfo *CUserInfo) SetUserAttributeByType(attributeType global.AttributeType, value int64) {
	//最大数量理论值判断
	if _, ok := userInfo.MapUserAttribute[attributeType]; !ok {
		if int32(len(userInfo.MapUserAttribute)) >= (global.AttributeTypeSaveMaxLen) {
			log.StackError("SetUserAttributeByType User Save Attribute Count more than global.AttributeTypeSaveMaxLen", log.String("UserId", userInfo.UserId), log.Int32(" AttributeType:", attributeType), log.Int32("AttributeTypeSaveMaxLen", global.AttributeTypeSaveMaxLen))
			return
		}
	}
	userInfo.MapUserAttribute[attributeType] = value
	//如果是设置等级,则设置达到此等级的时间,更新排行榜需要

	userInfo.MakeDirty()
}

// 加载玩家数据,引导和功能解锁信息
func (userInfo *CUserInfo) MsgLoadFinishData(msgInfo *msg.MsgLoadFinish) {
	if msgInfo == nil {
		log.SError("CUserInfo.MsgLoadFinishData player:", userInfo.UserId, " msgInfo is nil")
		return
	}
}

func (userInfo *CUserInfo) GetLastLogoutTime() int64 {
	return userInfo.LastLogoutTime
}

func (userInfo *CUserInfo) SetLastLogoutTime(nowTime int64) {
	userInfo.LastLogoutTime = nowTime
	userInfo.MakeDirty()
}
