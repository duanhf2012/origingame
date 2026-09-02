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

// LoginCredential 是客户端提交、尚未被信任的平台登录凭证。
type LoginCredential struct {
	PlatType    LoginType // 客户端声明的平台类型。
	PlatID      string    // 客户端提交的平台账号标识。
	AccessToken string    // 供平台 SDK 验证的访问凭证。
}

// PlatformIdentity 是鉴权实现确认后的可信平台身份。
type PlatformIdentity struct {
	PlatType LoginType // 鉴权确认的平台类型。
	PlatID   string    // 鉴权确认的平台账号标识。
}

// ValidLoginType 报告登录类型是否属于当前协议约定范围。
func ValidLoginType(value LoginType) bool {
	return value >= LoginTypeGuest && value < loginTypeMax
}

// Normalize 清理来自外部请求的可见文本字段。
func (credential LoginCredential) Normalize() LoginCredential {
	credential.PlatID = strings.TrimSpace(credential.PlatID)
	return credential
}
