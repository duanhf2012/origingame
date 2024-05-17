package collect

var AccountCollectName = "Account"

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
