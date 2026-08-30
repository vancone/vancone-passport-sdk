package middleware

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vancone/vancone-passport-sdk-go/pkg/config"
	"github.com/vancone/vancone-passport-sdk-go/pkg/constant"
	"github.com/vancone/vancone-passport-sdk-go/pkg/model"
	"github.com/vancone/vancone-passport-sdk-go/pkg/scheduler"
	"github.com/vancone/vancone-passport-sdk-go/pkg/util"
	"github.com/vancone/vancone-web-common-go/pkg/logger"
	"github.com/vancone/vancone-web-common-go/pkg/server/response"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	// 隔离测试输出：用 nop logger 替代文件日志，同时避免未初始化 logger 引发的 panic
	logger.SugaredLogger = zap.NewNop().Sugar()
	os.Exit(m.Run())
}

func newTestSigningKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate ECDSA key: %v", err)
	}
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	config.LocalConfig = config.PassportConfig{
		Authentication: config.AuthenticationConfig{Enabled: true},
		AccessControl:  config.AccessControlConfig{Enabled: false},
		BaseUrl:        "https://passport.example.com",
		AppAccount: config.AppAccountConfig{
			AccessKeyId:     "test-ak",
			SecretAccessKey: "test-sk",
		},
		Token: config.TokenConfig{
			Algorithm: "ES256",
			PublicKey: base64.StdEncoding.EncodeToString(publicKeyBytes),
		},
	}
	// 公钥只懒加载一次，配置换新密钥后需重新加载
	util.InitKeys()
	authenticatePatternParser = nil
	authenticateUriAllowlist = nil
	return privateKey
}

func signUserToken(t *testing.T, privateKey *ecdsa.PrivateKey, permissionIds []int64) string {
	t.Helper()
	tokenStr, err := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		constant.TokenKeyAccountId:     "acc-1",
		constant.TokenKeyTenantId:      "tenant-1",
		constant.TokenKeyUserId:        "user-1",
		constant.TokenKeyPermissionIds: permissionIds,
		"exp":                          time.Now().Add(time.Hour).Unix(),
	}).SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign user token: %v", err)
	}
	return tokenStr
}

func newGinContext(t *testing.T, method string, path string, headerToken string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(method, path, nil)
	if headerToken != "" {
		ctx.Request.Header.Set(constant.HeaderKeyToken, headerToken)
	}
	return ctx, w
}

func decodeErrorResponse(t *testing.T, w *httptest.ResponseRecorder) response.Response[string] {
	t.Helper()
	var resp response.Response[string]
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error response %s: %v", w.Body.String(), err)
	}
	return resp
}

// startMockPassport serves the token and sync endpoints so the access-control
// cache can be populated without a real Passport platform.
func startMockPassport(t *testing.T, privateKey *ecdsa.PrivateKey) {
	t.Helper()
	appToken, err := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign app token: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc(constant.GenerateTokenUrl, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(response.Response[string]{Code: 0, Message: "success", Data: appToken})
	})
	mux.HandleFunc(constant.SyncApiUrl, func(w http.ResponseWriter, r *http.Request) {
		apis := []model.Api{
			{Id: "1", Path: "/users/{id}", Method: "GET"},
			{Id: "2", Path: "/users/{id}", Method: "POST"},
			{Id: "3", Path: "/static/**", Method: "GET"},
		}
		_ = json.NewEncoder(w).Encode(response.Response[[]model.Api]{Code: 0, Message: "success", Data: apis})
	})
	mux.HandleFunc(constant.SyncPermissionUrl, func(w http.ResponseWriter, r *http.Request) {
		permissions := []model.Permission{
			{Id: "1", Name: "user-manage", ApiIds: []string{"1", "2"}},
			{Id: "2", Name: "static-read", ApiIds: []string{"3"}},
		}
		_ = json.NewEncoder(w).Encode(response.Response[[]model.Permission]{Code: 0, Message: "success", Data: permissions})
	})

	server := httptest.NewServer(mux)
	config.LocalConfig.BaseUrl = server.URL
	t.Cleanup(server.Close)
}

func syncAccessControlCache(t *testing.T) {
	t.Helper()
	scheduler.SyncApis()
	scheduler.SyncPermissions()
	if len(scheduler.GetApiCache()) != 3 || len(scheduler.GetPermissionCache()) != 2 {
		t.Fatalf("access control cache not populated, apis=%d permissions=%d",
			len(scheduler.GetApiCache()), len(scheduler.GetPermissionCache()))
	}
}

func TestAuthenticateDisabled(t *testing.T) {
	newTestSigningKey(t)
	config.LocalConfig.Authentication.Enabled = false

	ctx, w := newGinContext(t, http.MethodGet, "/api/v1/thing", "")
	if !Authenticate(ctx) {
		t.Fatal("Authenticate should pass when authentication is disabled")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestAuthenticateAllowlist(t *testing.T) {
	newTestSigningKey(t)
	config.LocalConfig.Authentication.UriAllowlist = []string{"/public/**"}
	Init()

	ctx, w := newGinContext(t, http.MethodGet, "/public/anything", "")
	if !Authenticate(ctx) {
		t.Fatal("Authenticate should pass for allowlisted URI")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestAuthenticateMissingToken(t *testing.T) {
	newTestSigningKey(t)

	ctx, w := newGinContext(t, http.MethodGet, "/api/v1/thing", "")
	if Authenticate(ctx) {
		t.Fatal("Authenticate should fail without a token")
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
	resp := decodeErrorResponse(t, w)
	if resp.Code != http.StatusUnauthorized || resp.Message != "Login required" {
		t.Fatalf("unexpected response body: %+v", resp)
	}
	if resp.Data != config.LocalConfig.BaseUrl {
		t.Fatalf("response data should carry the platform base url, got %s", resp.Data)
	}
}

func TestAuthenticateInvalidToken(t *testing.T) {
	newTestSigningKey(t)

	ctx, w := newGinContext(t, http.MethodGet, "/api/v1/thing", "invalid-token")
	if Authenticate(ctx) {
		t.Fatal("Authenticate should fail with an invalid token")
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

func TestAuthenticateValidTokenViaCookie(t *testing.T) {
	privateKey := newTestSigningKey(t)
	tokenStr := signUserToken(t, privateKey, nil)

	ctx, w := newGinContext(t, http.MethodGet, "/api/v1/thing", "")
	ctx.Request.AddCookie(&http.Cookie{Name: constant.CookieKeyToken, Value: tokenStr})
	if !Authenticate(ctx) {
		t.Fatal("Authenticate should pass with a valid cookie token")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	assertAccountInfoInContext(t, ctx, nil)
}

func TestAuthenticateValidTokenSetsContext(t *testing.T) {
	privateKey := newTestSigningKey(t)
	tokenStr := signUserToken(t, privateKey, []int64{1})

	ctx, w := newGinContext(t, http.MethodGet, "/api/v1/thing", tokenStr)
	if !Authenticate(ctx) {
		t.Fatal("Authenticate should pass with a valid token")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	assertAccountInfoInContext(t, ctx, []int64{1})
}

func assertAccountInfoInContext(t *testing.T, ctx *gin.Context, wantPermissionIds []int64) {
	t.Helper()
	if got, exists := ctx.Get(constant.TokenKeyAccountId); !exists || got != "acc-1" {
		t.Fatalf("context account id = %v", got)
	}
	if got, exists := ctx.Get(constant.TokenKeyTenantId); !exists || got != "tenant-1" {
		t.Fatalf("context tenant id = %v", got)
	}
	if got, exists := ctx.Get(constant.TokenKeyUserId); !exists || got != "user-1" {
		t.Fatalf("context user id = %v", got)
	}
	got, exists := ctx.Get(constant.AccountInfo)
	if !exists {
		t.Fatal("account info not set in context")
	}
	info, ok := got.(model.AccountInfo)
	if !ok {
		t.Fatalf("context value is not model.AccountInfo: %T", got)
	}
	if info.PermissionIds == nil && wantPermissionIds != nil {
		t.Fatalf("permission ids = nil, want %v", wantPermissionIds)
	}
	if len(info.PermissionIds) != len(wantPermissionIds) {
		t.Fatalf("permission ids = %v, want %v", info.PermissionIds, wantPermissionIds)
	}
	for i, want := range wantPermissionIds {
		if info.PermissionIds[i] != want {
			t.Fatalf("permission ids = %v, want %v", info.PermissionIds, wantPermissionIds)
		}
	}
}

func TestAccessControlAllowed(t *testing.T) {
	privateKey := newTestSigningKey(t)
	config.LocalConfig.AccessControl.Enabled = true
	startMockPassport(t, privateKey)
	syncAccessControlCache(t)

	ctx, w := newGinContext(t, http.MethodGet, "/users/123", signUserToken(t, privateKey, []int64{1}))
	if !Authenticate(ctx) {
		t.Fatal("request with granted permission should pass")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestAccessControlDeniedNoPermission(t *testing.T) {
	privateKey := newTestSigningKey(t)
	config.LocalConfig.AccessControl.Enabled = true
	startMockPassport(t, privateKey)
	syncAccessControlCache(t)

	ctx, w := newGinContext(t, http.MethodGet, "/users/123", signUserToken(t, privateKey, []int64{2}))
	if Authenticate(ctx) {
		t.Fatal("request without matching permission should be denied")
	}
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
	resp := decodeErrorResponse(t, w)
	if resp.Code != http.StatusForbidden || resp.Message != "No permission" {
		t.Fatalf("unexpected response body: %+v", resp)
	}
}

func TestAccessControlDeniedUnregisteredApi(t *testing.T) {
	privateKey := newTestSigningKey(t)
	config.LocalConfig.AccessControl.Enabled = true
	startMockPassport(t, privateKey)
	syncAccessControlCache(t)

	ctx, w := newGinContext(t, http.MethodGet, "/unknown/1", signUserToken(t, privateKey, []int64{1}))
	if Authenticate(ctx) {
		t.Fatal("request to an unregistered API should be denied")
	}
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
}

func TestAccessControlWildcardAllowed(t *testing.T) {
	privateKey := newTestSigningKey(t)
	config.LocalConfig.AccessControl.Enabled = true
	startMockPassport(t, privateKey)
	syncAccessControlCache(t)

	ctx, w := newGinContext(t, http.MethodGet, "/static/a/b.js", signUserToken(t, privateKey, []int64{2}))
	if !Authenticate(ctx) {
		t.Fatal("request covered by /** pattern with granted permission should pass")
	}
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}

func TestAuthMiddlewareAbortsOnFailure(t *testing.T) {
	newTestSigningKey(t)

	ctx, w := newGinContext(t, http.MethodGet, "/api/v1/thing", "")
	AuthMiddleware(ctx)
	if !ctx.IsAborted() {
		t.Fatal("AuthMiddleware should abort the request on authentication failure")
	}
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}
