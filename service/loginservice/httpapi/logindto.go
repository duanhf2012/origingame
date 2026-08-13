// Package httpapi 定义 LoginService 的 HTTP 接口、JSON DTO 和监听生命周期。
package httpapi

import (
	"origingame/service/loginservice/account"
	"origingame/service/loginservice/area"
)

// LoginRequest 是 POST /api/v1/login 的 JSON 请求 DTO。
type LoginRequest struct {
	PlatType    account.LoginType `json:"PlatType"`
	PlatID      string            `json:"PlatId" binding:"required"`
	AccessToken string            `json:"AccessToken"`
	GameID      string            `json:"GameId"`
	UserName    string            `json:"UserName"`
}

// Identity 转换并规范化 HTTP DTO，使业务包不依赖 JSON 契约。
func (request LoginRequest) Identity() account.LoginIdentity {
	return (account.LoginIdentity{
		PlatType: request.PlatType, PlatID: request.PlatID, AccessToken: request.AccessToken,
		GameID: request.GameID, UserName: request.UserName,
	}).Normalize()
}

// LoginResponse 是登录成功响应；omitempty 保证失败响应只包含 ECode。
type LoginResponse struct {
	ECode    int32       `json:"ECode"`
	Token    string      `json:"Token,omitempty"`
	AreaList []area.Info `json:"AreaList,omitempty"`
}
