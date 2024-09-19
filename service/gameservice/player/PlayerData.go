package player

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
	"time"
)

type PoolObj struct {
	ref bool
}

// 用于存于不需要持久化的数据
type DataInfo struct {
	//需要序列化的数据首字母大写
	FromGateId string //关联网关nodeId
	ClientId   string //clientId
	isOnline   bool   //是否在线状态

	pingTime        time.Time //ping时间
	saveTime        time.Time //上次存档时间
	lastEnterTime   int64     //上次进入地图时间 ms
	logoutTime      int64     //离线时长 暂时放这里，后续会根据这个时长验证某些功能
	nextDayZeroTime int64     //下一天0点时间，用于跨天计算（现实跨天）

	SessionId string //会话Id
	Ip        string //ip地址
	Os        string //操作系统
}

func (po *PoolObj) IsRef() bool {
	return po.ref
}

func (po *PoolObj) Ref() {
	po.ref = true
}

func (po *PoolObj) UnRef() {
	po.ref = false
}

func (po *PoolObj) Reset() {
	po.ref = false
}

func (df *DataInfo) GetClientId() string {
	return df.ClientId
}

func (df *DataInfo) GetGateNodeId() string {
	return df.FromGateId
}

func (df *DataInfo) AttachConn(fromGateNodeId string, clientId string, sessionId string, ip string, os string) {
	df.FromGateId = fromGateNodeId
	df.SessionId = sessionId
	ipStr := strings.Split(ip, ":")
	if len(ipStr) > 0 {
		ip = ipStr[0]
	}

	df.Ip = ip
	df.Os = os
	df.ClientId = clientId
	df.pingTime = time.Now()
	if clientId == "" {
		df.isOnline = false
	}
}

func (df *DataInfo) GenSessionId() {
	df.SessionId = primitive.NewObjectID().Hex()
}

func (df *DataInfo) GetSessionId() string {
	return df.SessionId
}

func (df *DataInfo) SetOnline(onLine bool) {
	df.isOnline = onLine
}
