// Package mongodb 定义 OriginGame 各 Service 共享的 MongoDB 集合契约。
package mongodb

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// AccountName 是账号基础集合名称。
const AccountName = "Account"

// Account 是 Account 集合的 BSON 文档结构。
// CreateTime 和 UpdateTime 是不受游戏逻辑时间影响的真实系统时间。
type Account struct {
	ID         bson.ObjectID `bson:"_id"`          // MongoDB 账号文档主键。
	PlatType   int32         `bson:"PlatType"`     // SDK 平台类型。
	PlatID     string        `bson:"PlatId"`       // 平台侧账号标识。
	IP         string        `bson:"Ip,omitempty"` // 最近登录 IP。
	CreateTime time.Time     `bson:"CreateTime"`   // 真实系统创建时间。
	UpdateTime time.Time     `bson:"UpdateTime"`   // 真实系统更新时间。
}
