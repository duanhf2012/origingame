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
		module.respondError(ctx, http.StatusBadRequest, commonpb.ErrorCode_ERROR_CODE_INVALID_REQUEST)
		return
	}
	credential := request.Credential()
	if !account.ValidLoginType(credential.PlatType) {
		module.respondError(ctx, http.StatusOK, commonpb.ErrorCode_ERROR_CODE_LOGIN_PLATFORM_TYPE_INVALID)
		return
	}
	if credential.PlatID == "" {
		module.respondError(ctx, http.StatusOK, commonpb.ErrorCode_ERROR_CODE_LOGIN_PLATFORM_ID_INVALID)
		return
	}

	requestCtx := ctx.Context()
	allowed, err := module.deps.Limiter.AllowIP(requestCtx, ctx.ClientIP())
	if err != nil {
		module.Logger().Error("登录 IP 限流不可用", log.Err(err))
		module.respondError(ctx, http.StatusServiceUnavailable, commonpb.ErrorCode_ERROR_CODE_INTERNAL)
		return
	}
	if !allowed {
		module.respondError(ctx, http.StatusTooManyRequests, commonpb.ErrorCode_ERROR_CODE_TOO_MANY_REQUESTS)
		return
	}

	identity, err := module.deps.Authenticator.Authenticate(requestCtx, credential)
	if err != nil {
		module.respondError(ctx, http.StatusOK, commonpb.ErrorCode_ERROR_CODE_LOGIN_AUTH_FAILED)
		return
	}
	document, err := module.deps.Accounts.FindOrCreate(requestCtx, identity, ctx.ClientIP())
	if err != nil {
		module.Logger().Error("登录账号查询或创建失败", log.Int32("platform_type", int32(identity.PlatType)), log.Err(err))
		module.respondError(ctx, http.StatusOK, commonpb.ErrorCode_ERROR_CODE_LOGIN_ACCOUNT_FAILED)
		return
	}

	allowed, err = module.deps.Limiter.AllowIdentity(requestCtx, document.ID.Hex())
	if err != nil {
		module.Logger().Error("登录身份限流不可用", log.Err(err))
		module.respondError(ctx, http.StatusServiceUnavailable, commonpb.ErrorCode_ERROR_CODE_INTERNAL)
		return
	}
	if !allowed {
		module.respondError(ctx, http.StatusTooManyRequests, commonpb.ErrorCode_ERROR_CODE_TOO_MANY_REQUESTS)
		return
	}

	token, err := module.deps.Issuer.Issue(document.ID.Hex())
	if err != nil {
		module.Logger().Error("签发游戏 Token 失败", log.Err(err))
		module.respondError(ctx, http.StatusInternalServerError, commonpb.ErrorCode_ERROR_CODE_INTERNAL)
		return
	}
	areas := module.deps.Catalog.Snapshot()
	if len(areas) == 0 {
		module.respondError(ctx, http.StatusServiceUnavailable, commonpb.ErrorCode_ERROR_CODE_LOGIN_NO_AVAILABLE_AREA)
		return
	}
	ctx.JSON(http.StatusOK, LoginResponse{
		ECode: int32(commonpb.ErrorCode_ERROR_CODE_OK), Token: token, AreaList: areas,
	})
}

func (module *Module) respondError(
	ctx *ginmodule.SafeContext,
	status int,
	code commonpb.ErrorCode,
) {
	ctx.JSON(status, LoginResponse{ECode: int32(code)})
}
