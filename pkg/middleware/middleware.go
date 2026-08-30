package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vancone/vancone-passport-sdk-go/pkg/config"
	"github.com/vancone/vancone-passport-sdk-go/pkg/constant"
	"github.com/vancone/vancone-passport-sdk-go/pkg/model"
	"github.com/vancone/vancone-passport-sdk-go/pkg/scheduler"
	"github.com/vancone/vancone-passport-sdk-go/pkg/util"
	"github.com/vancone/vancone-web-common-go/pkg/logger"
	"github.com/vancone/vancone-web-common-go/pkg/server/response"
)

var authenticatePatternParser *util.PathPatternParser
var authenticateUriAllowlist []string

func Init() {
	authenticatePatternParser = util.NewPathPatternParser()
	authenticateUriAllowlist = config.LocalConfig.Authentication.UriAllowlist
	for _, tokenMiddlewareUrl := range authenticateUriAllowlist {
		err := authenticatePatternParser.AddPattern(tokenMiddlewareUrl)
		if err != nil {
			logger.Error("Failed to add pattern to parser", err)
		}
	}
	if config.LocalConfig.AccessControl.Enabled {
		scheduler.Start()
	}
}

func AuthMiddleware(ctx *gin.Context) {
	if authenticatePatternParser == nil {
		Init()
	}
	if !Authenticate(ctx) {
		return
	}
	ctx.Next()
}

// Authenticate Validate user account's login status and, when access control
// is enabled, its permission on the requested API. Returns false after writing
// an error response and aborting the request.
func Authenticate(ctx *gin.Context) bool {
	// 如果关闭了登录校验，直接放行
	if !config.LocalConfig.Authentication.Enabled {
		return true
	}

	if len(authenticateUriAllowlist) > 0 {
		if pattern, _ := authenticatePatternParser.Match(ctx.Request.URL.Path); pattern != nil {
			return true
		}
	}

	// Java 端优先从 Header 取 token，Cookie 兜底
	token := ctx.GetHeader(constant.HeaderKeyToken)
	if token == "" {
		token, _ = ctx.Cookie(constant.CookieKeyToken)
	}

	accountInfo, err := util.ParseAccountInfo(token)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, response.Response[string]{
			Code:    http.StatusUnauthorized,
			Message: "Login required",
			Data:    config.LocalConfig.BaseUrl,
		})
		return false
	}

	if config.LocalConfig.AccessControl.Enabled && !accessControl(ctx.Request.URL.Path, ctx.Request.Method, &accountInfo) {
		logger.Error("No permission for user", accountInfo.AccountId, "path", ctx.Request.URL.Path)
		ctx.AbortWithStatusJSON(http.StatusForbidden, response.Response[any]{
			Code:    http.StatusForbidden,
			Message: "No permission",
		})
		return false
	}

	ctx.Set(constant.TokenKeyAccountId, accountInfo.AccountId)
	ctx.Set(constant.TokenKeyTenantId, accountInfo.TenantId)
	ctx.Set(constant.TokenKeyUserId, accountInfo.UserId)
	ctx.Set(constant.AccountInfo, accountInfo)
	return true
}

// accessControl mirrors Java AuthFilter#accessControl: the request must hit a
// registered API, and one of the account's permissions must contain that API.
func accessControl(path string, method string, accountInfo *model.AccountInfo) bool {
	currentApi := scheduler.MatchApi(path, method)
	if currentApi == nil {
		logger.Error("No registered API matched path", path)
		return false
	}
	for _, permissionId := range accountInfo.PermissionIds {
		for _, permission := range scheduler.GetPermissionCache() {
			if strconv.FormatInt(permissionId, 10) != permission.Id {
				continue
			}
			for _, apiId := range permission.ApiIds {
				if apiId == currentApi.Id {
					return true
				}
			}
		}
	}
	logger.Error("PermissionIds in token", accountInfo.PermissionIds)
	logger.Error("Permission in cache", scheduler.GetPermissionCache())
	return false
}
