package util

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vancone/vancone-passport-sdk-go/pkg/config"
	"github.com/vancone/vancone-passport-sdk-go/pkg/constant"
	"github.com/vancone/vancone-passport-sdk-go/pkg/model"
	"github.com/vancone/vancone-web-common-go/pkg/logger"
	"github.com/vancone/vancone-web-common-go/pkg/server/response"
)

var cachedToken string
var prevCacheTime int64

var publicKey interface{}

func generateSignature(ak string, sk string) (timestamp string, signature string) {
	body := map[string]string{
		"accessKeyId":     ak,
		"secretAccessKey": sk,
	}
	bodyStr, _ := json.Marshal(body)
	respBytes := Request(config.LocalConfig.PlatformConfig.BaseUrl+constant.GenerateSignUrl, "POST", string(bodyStr), false)
	resp := response.Response[any]{}
	json.Unmarshal(respBytes, &resp)
	data := make(map[string]string)
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &data)
	return data["timestamp"], data["signature"]
}

func generateToken() string {
	ak := config.LocalConfig.AppAccount.Ak
	sk := config.LocalConfig.AppAccount.Sk
	if ak == "" || sk == "" {
		logger.Error("App account ak / sk can't be empty")
		return ""
	}
	timestamp, signature := generateSignature(ak, sk)
	body := map[string]string{
		"accessKeyId": ak,
		"timestamp":   timestamp,
		"signature":   signature,
	}
	bodyStr, _ := json.Marshal(body)
	respBytes := Request(config.LocalConfig.PlatformConfig.BaseUrl+constant.GenerateTokenUrl, "POST", string(bodyStr), false)
	resp := response.Response[any]{}
	err := json.Unmarshal(respBytes, &resp)
	if err != nil {
		return ""
	}
	return resp.Data.(string)
}

func ApplyToken() string {
	currentTime := time.Now().Unix()
	if cachedToken == "" || currentTime-prevCacheTime > 3600 {
		cachedToken = generateToken()
		prevCacheTime = currentTime
	}
	return cachedToken
}

func InitKeys() {
	publicKeyBytes, _ := base64.StdEncoding.DecodeString(config.LocalConfig.Token.PublicKey)
	var err error
	publicKey, err = x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		logger.Error("Failed to parse public key", err)
	}
}

func ValidateToken(tokenStr string) bool {
	if publicKey == nil {
		InitKeys()
	}
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		logger.Error("Failed to parse token", err)
		return false
	}
	return token.Valid
}

func ParseAccountInfo(tokenStr string) model.AccountInfo {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		logger.Error("Failed to parse token", err)
	}
	accountMap := token.Claims.(jwt.MapClaims)
	return model.AccountInfo{
		AccountId: accountMap[constant.TokenKeyAccountId].(string),
		TenantId:  accountMap[constant.TokenKeyTenantId].(string),
		UserId:    accountMap[constant.TokenKeyUserId].(string),
	}
}
