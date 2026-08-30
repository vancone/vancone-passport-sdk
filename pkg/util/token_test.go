package util

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vancone/vancone-passport-sdk-go/pkg/config"
	"github.com/vancone/vancone-passport-sdk-go/pkg/constant"
)

// wantTestHmac is the independently pre-computed HMAC-SHA256 of "ak123"
// with key "sk", upper-case hex (the algorithm used by the Java SDK).
const wantTestHmac = "C4DF88189239B74BC9CCF6C616D992B826205EA6BEC243F2357433C074DD483E"

func TestHmacSha256Hex(t *testing.T) {
	if got := hmacSha256Hex("sk", "ak123"); got != wantTestHmac {
		t.Fatalf("hmacSha256Hex() = %s, want %s", got, wantTestHmac)
	}
}

func TestGenerateSignature(t *testing.T) {
	timestamp, signature, err := GenerateSignature("ak", "sk")
	if err != nil {
		t.Fatalf("GenerateSignature() error = %v", err)
	}
	if timestamp == "" {
		t.Fatal("timestamp should not be empty")
	}
	if len(signature) != 64 || signature != strings.ToUpper(signature) {
		t.Fatalf("signature should be 64 upper-case hex chars, got %s", signature)
	}
	// Message must be accessKeyId + timestamp signed with the secret key
	if signature != hmacSha256Hex("sk", "ak"+timestamp) {
		t.Fatalf("signature %s does not match HMAC(sk, ak+timestamp)", signature)
	}
}

func newTestSigningKey(t *testing.T, algorithm string) (privateKey interface{}, publicKeyBytes []byte) {
	t.Helper()
	switch algorithm {
	case "ES256":
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("generate ECDSA key: %v", err)
		}
		pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if err != nil {
			t.Fatalf("marshal public key: %v", err)
		}
		return key, pubBytes
	case "RS256":
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatalf("generate RSA key: %v", err)
		}
		pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
		if err != nil {
			t.Fatalf("marshal public key: %v", err)
		}
		return key, pubBytes
	default:
		t.Fatalf("unsupported algorithm %s", algorithm)
		return nil, nil
	}
}

func setTokenTestConfig(t *testing.T, algorithm string) interface{} {
	t.Helper()
	privateKey, publicKeyBytes := newTestSigningKey(t, algorithm)
	config.LocalConfig.Token.Algorithm = algorithm
	config.LocalConfig.Token.PublicKey = base64.StdEncoding.EncodeToString(publicKeyBytes)
	publicKey = nil
	return privateKey
}

func signTestToken(t *testing.T, key interface{}, algorithm string, claims jwt.MapClaims) string {
	t.Helper()
	if claims["exp"] == nil {
		claims["exp"] = time.Now().Add(time.Hour).Unix()
	}
	tokenStr, err := jwt.NewWithClaims(jwt.GetSigningMethod(algorithm), claims).SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return tokenStr
}

func TestValidateAndParseAccountInfoES256(t *testing.T) {
	key := setTokenTestConfig(t, "ES256")
	tokenStr := signTestToken(t, key, "ES256", jwt.MapClaims{
		constant.TokenKeyAccountId:     "acc-1",
		constant.TokenKeyTenantId:      "tenant-1",
		constant.TokenKeyUserId:        "user-1",
		constant.TokenKeyPermissionIds: []interface{}{int64(1), 2, "3"},
	})

	if !ValidateToken(tokenStr) {
		t.Fatal("token should be valid")
	}
	info, err := ParseAccountInfo(tokenStr)
	if err != nil {
		t.Fatalf("ParseAccountInfo() error = %v", err)
	}
	if info.AccountId != "acc-1" || info.TenantId != "tenant-1" || info.UserId != "user-1" {
		t.Fatalf("unexpected account info: %+v", info)
	}
	if len(info.PermissionIds) != 3 || info.PermissionIds[0] != 1 || info.PermissionIds[1] != 2 || info.PermissionIds[2] != 3 {
		t.Fatalf("permission ids should be [1 2 3], got %v", info.PermissionIds)
	}
}

func TestValidateAndParseAccountInfoRS256(t *testing.T) {
	key := setTokenTestConfig(t, "RS256")
	tokenStr := signTestToken(t, key, "RS256", jwt.MapClaims{
		constant.TokenKeyAccountId: "acc-2",
	})
	if _, err := ParseAccountInfo(tokenStr); err != nil {
		t.Fatalf("RS256 token should parse, got error: %v", err)
	}
}

func TestParseAccountInfoMissingClaims(t *testing.T) {
	key := setTokenTestConfig(t, "ES256")
	tokenStr := signTestToken(t, key, "ES256", jwt.MapClaims{})

	info, err := ParseAccountInfo(tokenStr)
	if err != nil {
		t.Fatalf("token without claims should still parse, got error: %v", err)
	}
	if info.AccountId != "" || info.TenantId != "" || info.UserId != "" || info.PermissionIds != nil {
		t.Fatalf("account info should be zero value, got %+v", info)
	}
}

func TestParseAccountInfoInvalidToken(t *testing.T) {
	setTokenTestConfig(t, "ES256")
	if _, err := ParseAccountInfo("not-a-token"); err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestGetExpiration(t *testing.T) {
	key := setTokenTestConfig(t, "ES256")

	valid := signTestToken(t, key, "ES256", jwt.MapClaims{})
	if left := GetExpiration(valid); left <= 0 || left > 3600 {
		t.Fatalf("expiration of valid token should be in (0, 3600], got %d", left)
	}

	expired := signTestToken(t, key, "ES256", jwt.MapClaims{"exp": time.Now().Add(-time.Hour).Unix()})
	if left := GetExpiration(expired); left != 0 {
		t.Fatalf("expiration of expired token should be 0, got %d", left)
	}

	if left := GetExpiration("invalid-token"); left != 0 {
		t.Fatalf("expiration of invalid token should be 0, got %d", left)
	}
}

func TestApplyTokenCaching(t *testing.T) {
	privateKey, publicKeyBytes := newTestSigningKey(t, "ES256")
	config.LocalConfig.Token.Algorithm = "ES256"
	config.LocalConfig.Token.PublicKey = base64.StdEncoding.EncodeToString(publicKeyBytes)
	publicKey = nil

	claims := jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()}
	appToken, err := jwt.NewWithClaims(jwt.SigningMethodES256, claims).SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign app token: %v", err)
	}

	var hits int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode token request body: %v", err)
		}
		if body["accessKeyId"] != "test-ak" {
			t.Errorf("accessKeyId = %s, want test-ak", body["accessKeyId"])
		}
		if len(body["signature"]) != 64 || body["timestamp"] == "" {
			t.Errorf("signature / timestamp missing in request: %+v", body)
		}
		_ = json.NewEncoder(w).Encode(struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data"`
		}{Code: 0, Message: "success", Data: appToken})
	}))
	defer server.Close()

	config.LocalConfig.AppAccount = config.AppAccountConfig{AccessKeyId: "test-ak", SecretAccessKey: "test-sk"}
	config.LocalConfig.BaseUrl = server.URL
	cachedToken = ""

	if token := ApplyToken(); token != appToken {
		t.Fatalf("ApplyToken() = %s, want the token issued by the platform", token)
	}
	if token := ApplyToken(); token != appToken {
		t.Fatalf("second ApplyToken() = %s, want cached token", token)
	}
	if hits != 1 {
		t.Fatalf("platform token endpoint should be hit exactly once, got %d", hits)
	}
}
