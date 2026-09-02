// Package httpapi 定义 LoginService 的 HTTP 接口、JSON DTO 和监听生命周期。
package httpapi

import (
	"origingame/service/loginservice/account"
	"origingame/service/loginservice/area"
)

// LoginRequest 是 POST /api/v1/login 的 JSON 请求 DTO。
type LoginRequest struct {
	PlatType    account.LoginType `json:"PlatType"`                  // 客户端声明的平台类型。
	PlatID      string            `json:"PlatId" binding:"required"` // 平台账号标识。
	AccessToken string            `json:"AccessToken"`               // 平台访问凭证。
}

// Credential 转换并规范化 HTTP DTO，使业务包不依赖 JSON 契约。
func (request LoginRequest) Credential() account.LoginCredential {
	return (account.LoginCredential{
		PlatType: request.PlatType, PlatID: request.PlatID, AccessToken: request.AccessToken,
	}).Normalize()
}

// LoginResponse 是登录成功响应；omitempty 保证失败响应只包含 ECode。
type LoginResponse struct {
	ECode    int32       `json:"ECode"`              // 客户端错误码。
	Token    string      `json:"Token,omitempty"`    // 成功时签发的游戏 Token。
	AreaList []area.Info `json:"AreaList,omitempty"` // 成功时可选择的区服列表。
}
