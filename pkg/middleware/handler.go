package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vancone/vancone-passport-sdk-go/pkg/util"
)

func AuthMiddleware(context *gin.Context) {
	token, err := context.Cookie("passport_token")
	if err != nil {
		log.Println("Failed to get token from cookie", err)
		//context.AbortWithStatus(http.StatusUnauthorized)
		//return
	}
	if token == "" {
		token = context.GetHeader("passport-token")
	}
	if !util.ValidateToken(token) {
		context.AbortWithStatus(http.StatusUnauthorized)
	} else {
		context.Set("token", token)
		accountInfo := util.GetAccountInfo(token)
		context.Set("userId", accountInfo.UserId)
		context.Set("tenantId", accountInfo.TenantId)
		context.Set("accountInfo", accountInfo)
	}
}
