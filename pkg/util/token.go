package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vancone/vancone-passport-sdk-go/pkg/config"
	"github.com/vancone/vancone-passport-sdk-go/pkg/constant"
	"github.com/vancone/vancone-passport-sdk-go/pkg/model"
	"github.com/vancone/vancone-web-common-go/pkg/logger"
	"github.com/vancone/vancone-web-common-go/pkg/server/response"
)

const cacheTokenRefreshThresholdSeconds = 300

var (
	cachedToken   string
	applyTokenMux sync.Mutex
	publicKey     interface{}
)

// GenerateSignature generates a HMAC-SHA256 signature of "accessKeyId + timestamp"
// with the secret access key, the same way the Passport platform verifies it.
func GenerateSignature(accessKeyId string, secretAccessKey string) (timestamp string, signature string, err error) {
	timestamp = strconv.FormatInt(time.Now().UnixMilli(), 10)
	signature = hmacSha256Hex(secretAccessKey, accessKeyId+timestamp)
	return timestamp, signature, nil
}

func hmacSha256Hex(secret string, message string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return strings.ToUpper(hex.EncodeToString(mac.Sum(nil)))
}

func generateToken() string {
	ak := config.LocalConfig.AppAccount.AccessKeyId
	sk := config.LocalConfig.AppAccount.SecretAccessKey
	if ak == "" || sk == "" {
		logger.Error("App account ak / sk can't be empty")
		return ""
	}
	timestamp, signature, err := GenerateSignature(ak, sk)
	if err != nil {
		logger.Error("Failed to generate signature", err)
		return ""
	}
	body := map[string]string{
		"accessKeyId": ak,
		"timestamp":   timestamp,
		"signature":   signature,
	}
	bodyStr, _ := json.Marshal(body)
	respBytes := Request(config.LocalConfig.BaseUrl+constant.GenerateTokenUrl, "POST", string(bodyStr), false)
	resp := response.Response[string]{}
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		logger.Error("Failed to apply token", err)
		return ""
	}
	return resp.Data
}

// ApplyToken returns the cached app token, reapplying a new one when it is
// blank or about to expire in cacheTokenRefreshThresholdSeconds.
func ApplyToken() string {
	applyTokenMux.Lock()
	defer applyTokenMux.Unlock()
	if cachedToken == "" || GetExpiration(cachedToken) < cacheTokenRefreshThresholdSeconds {
		cachedToken = generateToken()
	}
	return cachedToken
}

func InitKeys() {
	publicKeyBytes, err := base64.StdEncoding.DecodeString(config.LocalConfig.Token.PublicKey)
	if err != nil {
		logger.Error("Failed to decode public key", err)
		return
	}
	publicKey, err = x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		logger.Error("Failed to parse public key", err)
	}
}

func parseClaims(tokenStr string) (jwt.MapClaims, error) {
	if publicKey == nil {
		InitKeys()
	}
	if publicKey == nil {
		return nil, fmt.Errorf("public key is not initialized")
	}
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if !strings.EqualFold(token.Method.Alg(), config.LocalConfig.Token.Algorithm) {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func ValidateToken(tokenStr string) bool {
	_, err := parseClaims(tokenStr)
	if err != nil {
		logger.Error("Failed to validate token", err)
		return false
	}
	return true
}

// GetExpiration returns the seconds left before the token expires, 0 if it is
// invalid or already expired.
func GetExpiration(tokenStr string) int64 {
	claims, err := parseClaims(tokenStr)
	if err != nil {
		return 0
	}
	expiration, err := claims.GetExpirationTime()
	if err != nil || expiration == nil {
		return 0
	}
	seconds := expiration.Time.Unix() - time.Now().Unix()
	if seconds < 0 {
		return 0
	}
	return seconds
}

func ParseAccountInfo(tokenStr string) (model.AccountInfo, error) {
	claims, err := parseClaims(tokenStr)
	if err != nil {
		logger.Error("Failed to parse token", err)
		return model.AccountInfo{}, err
	}
	return model.AccountInfo{
		AccountId:     claimString(claims, constant.TokenKeyAccountId),
		TenantId:      claimString(claims, constant.TokenKeyTenantId),
		UserId:        claimString(claims, constant.TokenKeyUserId),
		PermissionIds: claimInt64Slice(claims, constant.TokenKeyPermissionIds),
	}, nil
}

func claimString(claims jwt.MapClaims, key string) string {
	if value, ok := claims[key].(string); ok {
		return value
	}
	return ""
}

func claimInt64Slice(claims jwt.MapClaims, key string) []int64 {
	rawList, ok := claims[key].([]interface{})
	if !ok {
		return nil
	}
	result := make([]int64, 0, len(rawList))
	for _, rawItem := range rawList {
		switch value := rawItem.(type) {
		case float64:
			result = append(result, int64(value))
		case int64:
			result = append(result, value)
		case json.Number:
			if parsed, err := value.Int64(); err == nil {
				result = append(result, parsed)
			}
		case string:
			if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
				result = append(result, parsed)
			}
		}
	}
	return result
}
