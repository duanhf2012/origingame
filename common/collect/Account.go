package collect

const MaxHorseRaceLampCount = 64

var AccountCollectName = "Account"

var ActiveCodeCollectName = "ActiveCode"

var HorseRaceLampCollectName = "HorseRaceLamp"
var SealCollectName = "Seal"

type CAccount struct {
	PlatId   string `bson:"_id"`      //生成最新id
	PlatType int32  `bson:"PlatType"` //平台类型
	//PlatId   string           `bson:"PlatId,omitempty"`   //平台id
	Token     string           `bson:"Token"`     //Token
	AreaHis   map[string]int64 `bson:"AreaHis"`   //历史区服
	Gm        bool             `bson:"Gm"`        //是否为GM
	Channel   string           `bson:"Channel"`   //渠道 todo 暂时没有——预留
	Ip        string           `bson:"Ip"`        //Ip todo 暂时没有——预留
	Equipment string           `bson:"Equipment"` //设备 todo 暂时没有——预留
}

type ActiveCode struct {
	Code   string `bson:"_id"`    //生成激活码
	Expire int64  `bson:"Expire"` //有效时间
	Tag    bool   `bson:"Tag"`
}

type HorseRaceLamp struct {
	Id             string `bson:"_id""`           //平台生成的唯一ID
	Type           uint8  `bson:"type"`           //跑马灯类型
	AreaId         int32  `bson:"areaId"`         //区服ID
	SendTime       int64  `bson:"sendTime"`       //发送时间
	EffectTime     int64  `bson:"effectTime"`     //生效时间
	ExpirationTime int64  `bson:"expirationTime"` //失效时间
	FrequencyTime  int64  `bson:"frequency"`      //频率时间
	Count          uint8  `bson:"count"`          //在频率时间播放次数
	Priority       int32  `bson:"priority"`       //优先级
	Channel        int32  `bson:"channel"`        //渠道
	Content        string `bson:"content"`        //内容
	Status         uint8  `bson:"status"`         //状态 0:正常 1:删除
}

type ESealType = int32

const (
	ESTalk      ESealType = 1 //禁言
	ESAccount   ESealType = 2 //封账号
	ESEquipment ESealType = 3 //封设备
	ESKickOut   ESealType = 4 //强制登出
	ESIP        ESealType = 5 //封IP
	ESChannel   ESealType = 6 //封渠道
)

type CSealInfo struct {
	Id         uint64    `bson:"_id"`        //封号ID
	SealType   ESealType `bson:"sealType"`   //封号类型
	SealValue  string    `bson:"sealValue"`  //封号值
	SealTime   int64     `bson:"sealTime"`   //封号的操作时间戳——ms
	UnSealTime int64     `bson:"unSealTime"` //解封时间戳——ms 为0表示永久不解封
}
