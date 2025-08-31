package util

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/vancone/vancone-passport-sdk-go/pkg/constant"
	"github.com/vancone/vancone-passport-sdk-go/pkg/model"
)

func GetAccountInfo(ctx *gin.Context) *model.AccountInfo {
	accountInfo, exists := ctx.Get(constant.AccountInfo)
	if !exists {
		log.Println("Account info not found")
		return nil
	}
	accountInfoReturnValue := accountInfo.(model.AccountInfo)
	return &accountInfoReturnValue
}

func GetTenantId(ctx *gin.Context) string {
	tenantId, exists := ctx.Get(constant.TokenKeyTenantId)
	if !exists {
		log.Println("TenantId not exists")
		return ""
	}
	return tenantId.(string)
}

func GetAccountId(ctx *gin.Context) string {
	accountId, exists := ctx.Get(constant.TokenKeyAccountId)
	if !exists {
		log.Println("AccountId not exists")
		return ""
	}
	return accountId.(string)
}

func GetUserId(ctx *gin.Context) string {
	userId, exists := ctx.Get(constant.TokenKeyUserId)
	if !exists {
		log.Println("UserId not exists")
		return ""
	}
	return userId.(string)
}

func GetToken(ctx *gin.Context) string {
	token, err := ctx.Cookie(constant.CookieKeyToken)
	if err != nil {
		log.Println("Failed to get token from cookie", err)
	}
	if token == "" {
		token = ctx.GetHeader(constant.HeaderKeyToken)
	}
	return token
}
