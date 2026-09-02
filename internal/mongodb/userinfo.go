package mongodb

import "time"

// UserInfoName 是玩家公共基础持久化集合名称。
const UserInfoName = "UserInfo"

// UserInfo 是 UserInfo 集合的 BSON 文档结构。
// 三个时间字段都使用不受 GM 调时影响的真实系统时间。
type UserInfo struct {
	PlayerKey  string `bson:"_id"`          // 账号与显示区服组成的稳定主键。
	AccountID  string `bson:"account_id"`   // 所属账号标识。
	ShowAreaID int64  `bson:"show_area_id"` // 所属显示区服标识。

	Nickname string `bson:"nickname"` // 玩家昵称。
	Level    int32  `bson:"level"`    // 当前等级。

	CreatedAt    time.Time `bson:"created_at"`     // 真实系统创建时间。
	LastLoginAt  time.Time `bson:"last_login_at"`  // 最近登录真实时间。
	LastLogoutAt time.Time `bson:"last_logout_at"` // 最近登出真实时间。
}
