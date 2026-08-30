package scheduler

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vancone/vancone-passport-sdk-go/pkg/config"
	"github.com/vancone/vancone-passport-sdk-go/pkg/constant"
	"github.com/vancone/vancone-passport-sdk-go/pkg/model"
	"github.com/vancone/vancone-passport-sdk-go/pkg/util"
	"github.com/vancone/vancone-web-common-go/pkg/server/response"
)

// startMockPassport spins up a fake Passport platform serving the token and
// sync endpoints, and points the SDK config at it. The sync endpoints reject
// requests without the app token, verifying that the SDK authenticates itself.
func startMockPassport(t *testing.T) string {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate ECDSA key: %v", err)
	}
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	appToken, err := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign app token: %v", err)
	}

	config.LocalConfig = config.PassportConfig{
		BaseUrl: "",
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

	mux := http.NewServeMux()
	mux.HandleFunc(constant.GenerateTokenUrl, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(response.Response[string]{Code: 0, Message: "success", Data: appToken})
	})
	mux.HandleFunc(constant.SyncApiUrl, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(constant.HeaderKeyToken) != appToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		apis := []model.Api{
			{Id: "1", Path: "/users/{id}", Method: "GET"},
			{Id: "2", Path: "/users/{id}", Method: "POST"},
			{Id: "3", Path: "/static/**", Method: "GET"},
		}
		_ = json.NewEncoder(w).Encode(response.Response[[]model.Api]{Code: 0, Message: "success", Data: apis})
	})
	mux.HandleFunc(constant.SyncPermissionUrl, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(constant.HeaderKeyToken) != appToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		permissions := []model.Permission{
			{Id: "1", Name: "user-manage", ApiIds: []string{"1", "2"}},
			{Id: "2", Name: "static-read", ApiIds: []string{"3"}},
		}
		_ = json.NewEncoder(w).Encode(response.Response[[]model.Permission]{Code: 0, Message: "success", Data: permissions})
	})

	server := httptest.NewServer(mux)
	config.LocalConfig.BaseUrl = server.URL
	t.Cleanup(server.Close)
	return appToken
}

func TestSyncApisAndMatchApi(t *testing.T) {
	startMockPassport(t)

	SyncApis()

	apis := GetApiCache()
	if len(apis) != 3 {
		t.Fatalf("api cache should contain 3 apis, got %d", len(apis))
	}

	cases := []struct {
		path   string
		method string
		wantId string
	}{
		{"/users/123", http.MethodGet, "1"},
		{"/users/123", http.MethodPost, "2"},
		{"/users/123", http.MethodDelete, ""},
		{"/static/a/b/c.js", http.MethodGet, "3"},
		{"/static", http.MethodGet, "3"},
		{"/unknown/1", http.MethodGet, ""},
	}
	for _, c := range cases {
		api := MatchApi(c.path, c.method)
		gotId := ""
		if api != nil {
			gotId = api.Id
		}
		if gotId != c.wantId {
			t.Errorf("MatchApi(%s, %s).Id = %q, want %q", c.path, c.method, gotId, c.wantId)
		}
	}
}

func TestSyncPermissions(t *testing.T) {
	startMockPassport(t)

	SyncPermissions()

	permissions := GetPermissionCache()
	if len(permissions) != 2 {
		t.Fatalf("permission cache should contain 2 permissions, got %d", len(permissions))
	}
	if permissions[0].Id != "1" || len(permissions[0].ApiIds) != 2 || permissions[0].ApiIds[0] != "1" {
		t.Fatalf("unexpected permission cache: %+v", permissions)
	}
}
