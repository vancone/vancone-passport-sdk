package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestInitDefaults(t *testing.T) {
	LocalConfig = PassportConfig{}
	Init(viper.New())

	if LocalConfig.BaseUrl != "https://passport.vancone.com" {
		t.Fatalf("default base url = %s", LocalConfig.BaseUrl)
	}
	if LocalConfig.Cache.SyncPeriodSeconds != 60 {
		t.Fatalf("default sync period = %d, want 60", LocalConfig.Cache.SyncPeriodSeconds)
	}
	if LocalConfig.Token.Algorithm != "ES256" {
		t.Fatalf("default token algorithm = %s, want ES256", LocalConfig.Token.Algorithm)
	}
	if !LocalConfig.AccessControl.Enabled {
		t.Fatal("access control should be enabled by default")
	}
	if !LocalConfig.Authentication.Enabled {
		t.Fatal("authentication should be enabled by default")
	}
}

func TestInitRespectsExplicitValues(t *testing.T) {
	LocalConfig = PassportConfig{}
	v := viper.New()
	v.Set("passport.base-url", "http://localhost:9000")
	v.Set("passport.cache.sync-period-seconds", 120)
	v.Set("passport.access-control.enabled", false)
	v.Set("passport.authentication.enabled", false)
	v.Set("passport.token.algorithm", "RS256")
	v.Set("passport.token.public-key", "cHVibGljLWtleQ==")
	Init(v)

	if LocalConfig.BaseUrl != "http://localhost:9000" {
		t.Fatalf("base url = %s", LocalConfig.BaseUrl)
	}
	if LocalConfig.Cache.SyncPeriodSeconds != 120 {
		t.Fatalf("sync period = %d, want 120", LocalConfig.Cache.SyncPeriodSeconds)
	}
	if LocalConfig.AccessControl.Enabled {
		t.Fatal("explicitly disabled access control should stay disabled")
	}
	if LocalConfig.Authentication.Enabled {
		t.Fatal("explicitly disabled authentication should stay disabled")
	}
	if LocalConfig.Token.Algorithm != "RS256" || LocalConfig.Token.PublicKey != "cHVibGljLWtleQ==" {
		t.Fatalf("unexpected token config: %+v", LocalConfig.Token)
	}
}

func TestAppAccountDocumentedKeyNames(t *testing.T) {
	LocalConfig = PassportConfig{}
	v := viper.New()
	v.Set("passport.app-account.access-key-id", "doc-ak")
	v.Set("passport.app-account.secret-access-key", "doc-sk")
	Init(v)

	if LocalConfig.AppAccount.AccessKeyId != "doc-ak" || LocalConfig.AppAccount.SecretAccessKey != "doc-sk" {
		t.Fatalf("documented key names should be picked up, got %+v", LocalConfig.AppAccount)
	}
}

func TestAppAccountLegacyKeyFallback(t *testing.T) {
	LocalConfig = PassportConfig{}
	v := viper.New()
	v.Set("passport.app-account.ak", "legacy-ak")
	v.Set("passport.app-account.sk", "legacy-sk")
	Init(v)

	if LocalConfig.AppAccount.AccessKeyId != "legacy-ak" || LocalConfig.AppAccount.SecretAccessKey != "legacy-sk" {
		t.Fatalf("legacy ak / sk should fall back to access-key-id / secret-access-key, got %+v", LocalConfig.AppAccount)
	}
}

func TestLegacyPlatformSectionFallback(t *testing.T) {
	LocalConfig = PassportConfig{}
	v := viper.New()
	v.Set("passport.platform.base-url", "http://legacy:8080")
	v.Set("passport.platform.cache-sync-seconds", 90)
	Init(v)

	if LocalConfig.BaseUrl != "http://legacy:8080" {
		t.Fatalf("legacy platform base url should be honored, got %s", LocalConfig.BaseUrl)
	}
	if LocalConfig.Cache.SyncPeriodSeconds != 90 {
		t.Fatalf("legacy platform sync period should be honored, got %d", LocalConfig.Cache.SyncPeriodSeconds)
	}
}
