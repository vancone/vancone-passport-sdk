package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vancone/vancone-passport-sdk-go/pkg/config"
	"github.com/vancone/vancone-passport-sdk-go/pkg/constant"
	"github.com/vancone/vancone-passport-sdk-go/pkg/util"
)

var authenticatePatternParser *util.PathPatternParser
var authenticateUriAllowlist []string

func Init() {
	authenticatePatternParser = util.NewPathPatternParser()
	authenticateUriAllowlist = config.LocalConfig.Authentication.UriAllowlist
	if authenticateUriAllowlist != nil && len(authenticateUriAllowlist) > 0 {
		for _, tokenMiddlewareUrl := range config.LocalConfig.Authentication.UriAllowlist {
			err := authenticatePatternParser.AddPattern(tokenMiddlewareUrl)
			if err != nil {
				log.Println("Failed to add pattern to parser", err)
			}
		}
	}
}

func AuthMiddleware(ctx *gin.Context) {
	if authenticatePatternParser == nil {
		Init()
	}
	Authenticate(ctx)

}

// Authenticate Validate user account's login status
func Authenticate(ctx *gin.Context) bool {
	// 如果关闭了登录校验，直接放行
	if !config.LocalConfig.Authentication.Enabled {
		return true
	}

	if authenticateUriAllowlist != nil && len(authenticateUriAllowlist) > 0 {
		if pattern, _ := authenticatePatternParser.Match(ctx.Request.URL.Path); pattern != nil {
			return true
		}
	}

	token, err := ctx.Cookie(constant.CookieKeyToken)
	if err != nil {
		log.Println("Failed to get token from cookie", err)
	}
	if token == "" {
		token = ctx.GetHeader(constant.HeaderKeyToken)
	}
	if !util.ValidateToken(token) {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	} else {
		accountInfo := util.ParseAccountInfo(token)
		ctx.Set(constant.TokenKeyAccountId, accountInfo.AccountId)
		ctx.Set(constant.TokenKeyTenantId, accountInfo.TenantId)
		ctx.Set(constant.TokenKeyUserId, accountInfo.UserId)
		ctx.Set(constant.AccountInfo, accountInfo)
	}
	return false
	//return func(ctx *gin.Context) {
	//	// Allow account registration API
	//	if ctx.Request.RequestURI == "/api/passport/service/v1/account" && ctx.Request.Method == "POST" {
	//		return
	//	}
	//	if pattern, _ := parser.Match(ctx.Request.RequestURI); pattern != nil {
	//		token := ctx.GetHeader(constant.HeaderKeyToken)
	//		if token == "" {
	//			token, _ = ctx.Cookie(constant.CookieKeyToken)
	//		}
	//		if util.ValidateTokenAndWriteContext(token, ctx) {
	//			return
	//		}
	//		ctx.JSON(http.StatusUnauthorized, response.Response{
	//			Code:    11002,
	//			Message: "Not authorized",
	//		})
	//		ctx.Abort()
	//	}
	//}
}
