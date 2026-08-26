package httpapi

import (
	"net/http"

	"github.com/duanhf2012/origin/v3/log"
	"github.com/duanhf2012/origin/v3/sysmodule/ginmodule"
	commonpb "origingame/protocol/common"
	"origingame/service/loginservice/account"
)

// login 完成 POST /api/v1/login 的边界校验和登录流程协调。
func (module *Module) login(ctx *ginmodule.SafeContext) {
	var request LoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		module.logLoginRejected("request_decode", commonpb.ErrorCode_ERROR_CODE_INVALID_REQUEST)
		module.respondError(ctx, http.StatusBadRequest, commonpb.ErrorCode_ERROR_CODE_INVALID_REQUEST)
		return
	}
	credential := request.Credential()
	if !account.ValidLoginType(credential.PlatType) {
		module.logLoginRejected("platform_type", commonpb.ErrorCode_ERROR_CODE_LOGIN_PLATFORM_TYPE_INVALID)
		module.respondError(ctx, http.StatusOK, commonpb.ErrorCode_ERROR_CODE_LOGIN_PLATFORM_TYPE_INVALID)
		return
	}
	if credential.PlatID == "" {
		module.logLoginRejected("platform_id", commonpb.ErrorCode_ERROR_CODE_LOGIN_PLATFORM_ID_INVALID)
		module.respondError(ctx, http.StatusOK, commonpb.ErrorCode_ERROR_CODE_LOGIN_PLATFORM_ID_INVALID)
		return
	}
	// 平台标识与 AccessToken 都属于敏感身份信息；只记录可用于定位契约问题的类型和长度。
	module.Logger().Debug(
		"登录请求已受理",
		log.Int32("platform_type", int32(credential.PlatType)),
		log.Int("platform_id_length", len(credential.PlatID)),
	)

	requestCtx := ctx.Context()
	allowed, err := module.deps.Limiter.AllowIP(requestCtx, ctx.ClientIP())
	if err != nil {
		module.Logger().Error("登录 IP 限流不可用", log.Err(err))
		module.respondError(ctx, http.StatusServiceUnavailable, commonpb.ErrorCode_ERROR_CODE_INTERNAL)
		return
	}
	if !allowed {
		module.logLoginRejected("ip_rate_limit", commonpb.ErrorCode_ERROR_CODE_TOO_MANY_REQUESTS)
		module.respondError(ctx, http.StatusTooManyRequests, commonpb.ErrorCode_ERROR_CODE_TOO_MANY_REQUESTS)
		return
	}

	identity, err := module.deps.Authenticator.Authenticate(requestCtx, credential)
	if err != nil {
		module.logLoginRejected("authenticate", commonpb.ErrorCode_ERROR_CODE_LOGIN_AUTH_FAILED)
		module.respondError(ctx, http.StatusOK, commonpb.ErrorCode_ERROR_CODE_LOGIN_AUTH_FAILED)
		return
	}
	module.Logger().Debug("登录平台鉴权成功", log.Int32("platform_type", int32(identity.PlatType)))
	document, err := module.deps.Accounts.FindOrCreate(requestCtx, identity, ctx.ClientIP())
	if err != nil {
		module.Logger().Error("登录账号查询或创建失败", log.Int32("platform_type", int32(identity.PlatType)), log.Err(err))
		module.respondError(ctx, http.StatusOK, commonpb.ErrorCode_ERROR_CODE_LOGIN_ACCOUNT_FAILED)
		return
	}
	module.Logger().Debug("登录账号已准备", log.Int32("platform_type", int32(identity.PlatType)))

	allowed, err = module.deps.Limiter.AllowIdentity(requestCtx, document.ID.Hex())
	if err != nil {
		module.Logger().Error("登录身份限流不可用", log.Err(err))
		module.respondError(ctx, http.StatusServiceUnavailable, commonpb.ErrorCode_ERROR_CODE_INTERNAL)
		return
	}
	if !allowed {
		module.logLoginRejected("identity_rate_limit", commonpb.ErrorCode_ERROR_CODE_TOO_MANY_REQUESTS)
		module.respondError(ctx, http.StatusTooManyRequests, commonpb.ErrorCode_ERROR_CODE_TOO_MANY_REQUESTS)
		return
	}

	token, err := module.deps.Issuer.Issue(document.ID.Hex())
	if err != nil {
		module.Logger().Error("签发游戏 Token 失败", log.Err(err))
		module.respondError(ctx, http.StatusInternalServerError, commonpb.ErrorCode_ERROR_CODE_INTERNAL)
		return
	}
	module.Logger().Debug("登录游戏凭证签发完成")
	areas := module.deps.Catalog.Snapshot()
	if len(areas) == 0 {
		module.logLoginRejected("area_catalog", commonpb.ErrorCode_ERROR_CODE_LOGIN_NO_AVAILABLE_AREA)
		module.respondError(ctx, http.StatusServiceUnavailable, commonpb.ErrorCode_ERROR_CODE_LOGIN_NO_AVAILABLE_AREA)
		return
	}
	module.Logger().Debug("登录响应已完成", log.Int("area_count", len(areas)))
	ctx.JSON(http.StatusOK, LoginResponse{
		ECode: int32(commonpb.ErrorCode_ERROR_CODE_OK), Token: token, AreaList: areas,
	})
}

// logLoginRejected 仅记录安全的失败阶段和协议错误码，避免把登录凭证或请求内容写入日志。
func (module *Module) logLoginRejected(stage string, code commonpb.ErrorCode) {
	module.Logger().Debug(
		"登录请求被拒绝",
		log.String("login_stage", stage),
		log.Int32("error_code", int32(code)),
	)
}

func (module *Module) respondError(ctx *ginmodule.SafeContext, status int, code commonpb.ErrorCode) {
	ctx.JSON(status, LoginResponse{ECode: int32(code)})
}
