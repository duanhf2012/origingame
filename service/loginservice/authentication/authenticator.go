// Package authentication 定义第三方登录平台鉴权边界。
package authentication

import (
	"context"

	"origingame/service/loginservice/account"
)

// Authenticator 隔离第三方 SDK；正式平台实现必须替换开发透传实现。
type Authenticator interface {
	Authenticate(context.Context, account.LoginIdentity) error
}

// DevelopmentAuthenticator 只用于学习和本地开发，不验证 AccessToken。
type DevelopmentAuthenticator struct{}

// Authenticate 明确执行空平台鉴权；公共参数校验由 LoginService 先完成。
func (DevelopmentAuthenticator) Authenticate(context.Context, account.LoginIdentity) error {
	// TODO: 使用者在生产项目中按具体渠道接入 SDK，并删除 development_passthrough 配置。
	return nil
}
