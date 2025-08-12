package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vancone/vancone-passport-sdk-go/pkg/constant"
	"github.com/vancone/vancone-passport-sdk-go/pkg/util"
)

func AuthMiddleware(ctx *gin.Context) {
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
}
