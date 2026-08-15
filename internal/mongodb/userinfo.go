package mongodb

import "time"

// UserInfoName 是玩家公共基础持久化集合名称。
const UserInfoName = "UserInfo"

// UserInfo 是 UserInfo 集合的 BSON 文档结构。
// 三个时间字段都使用不受 GM 调时影响的真实系统时间。
type UserInfo struct {
	PlayerKey  string `bson:"_id"`
	AccountID  string `bson:"account_id"`
	ShowAreaID int64  `bson:"show_area_id"`

	Nickname string `bson:"nickname"`
	Level    int32  `bson:"level"`

	CreatedAt    time.Time `bson:"created_at"`
	LastLoginAt  time.Time `bson:"last_login_at"`
	LastLogoutAt time.Time `bson:"last_logout_at"`
}
