// Package account 定义登录账号身份及其 MongoDB 持久化边界。
package account

import "strings"

// LoginType 与老版本登录请求的数值保持一致，便于现有客户端迁移。
type LoginType int32

const (
	LoginTypeGuest    LoginType = 0
	LoginTypeAccount  LoginType = 1
	LoginTypeTapTap   LoginType = 2
	LoginTypeFacebook LoginType = 3
	LoginTypeGoogle   LoginType = 4
	loginTypeMax      LoginType = 5
)

// LoginIdentity 保存完成 HTTP 解析后的平台登录身份。
type LoginIdentity struct {
	PlatType    LoginType
	PlatID      string
	AccessToken string
	GameID      string
	UserName    string
}

// ValidLoginType 报告登录类型是否属于当前协议约定范围。
func ValidLoginType(value LoginType) bool {
	return value >= LoginTypeGuest && value < loginTypeMax
}

// Normalize 清理来自外部请求的可见文本字段。
func (identity LoginIdentity) Normalize() LoginIdentity {
	identity.PlatID = strings.TrimSpace(identity.PlatID)
	identity.GameID = strings.TrimSpace(identity.GameID)
	identity.UserName = strings.TrimSpace(identity.UserName)
	return identity
}
