package util

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/vancone/vancone-passport-sdk-go/pkg/constant"
	"github.com/vancone/vancone-passport-sdk-go/pkg/model"
)

func newTestGinContext(t *testing.T, method string, path string) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(method, path, nil)
	return ctx, w
}

func TestAccountInfoGetters(t *testing.T) {
	ctx, _ := newTestGinContext(t, http.MethodGet, "/api/v1/thing")
	if GetAccountInfo(ctx) != nil {
		t.Fatal("GetAccountInfo should return nil when not set")
	}
	if GetTenantId(ctx) != "" || GetAccountId(ctx) != "" || GetUserId(ctx) != "" {
		t.Fatal("getters should return empty strings when values not set")
	}

	want := model.AccountInfo{AccountId: "acc-1", TenantId: "tenant-1", UserId: "user-1"}
	ctx.Set(constant.AccountInfo, want)
	ctx.Set(constant.TokenKeyAccountId, want.AccountId)
	ctx.Set(constant.TokenKeyTenantId, want.TenantId)
	ctx.Set(constant.TokenKeyUserId, want.UserId)

	got := GetAccountInfo(ctx)
	if got == nil || got.AccountId != want.AccountId || got.TenantId != want.TenantId || got.UserId != want.UserId {
		t.Fatalf("GetAccountInfo() = %+v, want %+v", got, want)
	}
	if GetTenantId(ctx) != want.TenantId || GetAccountId(ctx) != want.AccountId || GetUserId(ctx) != want.UserId {
		t.Fatal("getters should return the values set in context")
	}
}

func TestGetToken(t *testing.T) {
	ctx, _ := newTestGinContext(t, http.MethodGet, "/api/v1/thing")
	if GetToken(ctx) != "" {
		t.Fatal("GetToken should return empty when no token present")
	}

	// Header takes precedence
	ctx.Request.Header.Set(constant.HeaderKeyToken, "header-token")
	ctx.Request.AddCookie(&http.Cookie{Name: constant.CookieKeyToken, Value: "cookie-token"})
	if GetToken(ctx) != "header-token" {
		t.Fatalf("GetToken() = %s, want header-token", GetToken(ctx))
	}

	// Cookie as fallback
	ctx.Request.Header.Del(constant.HeaderKeyToken)
	if GetToken(ctx) != "cookie-token" {
		t.Fatalf("GetToken() = %s, want cookie-token", GetToken(ctx))
	}
}
