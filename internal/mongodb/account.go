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
	ID         bson.ObjectID `bson:"_id"`
	PlatType   int32         `bson:"PlatType"`
	PlatID     string        `bson:"PlatId"`
	IP         string        `bson:"Ip,omitempty"`
	CreateTime time.Time     `bson:"CreateTime"`
	UpdateTime time.Time     `bson:"UpdateTime"`
}
