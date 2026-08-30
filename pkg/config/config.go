package config

import (
	"github.com/spf13/viper"
	"github.com/vancone/vancone-web-common-go/pkg/logger"
)

var Viper *viper.Viper
var LocalConfig PassportConfig

type PassportConfig struct {
	AccessControl  AccessControlConfig  `mapstructure:"access-control"`
	AppAccount     AppAccountConfig     `mapstructure:"app-account"`
	Authentication AuthenticationConfig `mapstructure:"authentication"`
	BaseUrl        string               `mapstructure:"base-url"`
	Cache          CacheConfig          `mapstructure:"cache"`
	CsrfConfig     CsrfConfig           `mapstructure:"csrf"`
	Token          TokenConfig          `mapstructure:"token"`
}

type AccessControlConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	UriAllowlist []string `mapstructure:"uri-allowlist"`
}

type AppAccountConfig struct {
	AccessKeyId     string `mapstructure:"access-key-id"`
	SecretAccessKey string `mapstructure:"secret-access-key"`
	// 兼容旧版配置键名
	Ak string `mapstructure:"ak"`
	Sk string `mapstructure:"sk"`
}

type AuthenticationConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	UriAllowlist []string `mapstructure:"uri-allowlist"`
}

type CacheConfig struct {
	SyncPeriodSeconds int `mapstructure:"sync-period-seconds"`
}

type CsrfConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	SecretKey string `mapstructure:"secret-key"`
}

type TokenConfig struct {
	Algorithm string `mapstructure:"algorithm"`
	PublicKey string `mapstructure:"public-key"`
}

func Init(v *viper.Viper) {
	Viper = v
	if v.Get("passport.access-control.enabled") == nil {
		v.Set("passport.access-control.enabled", true)
	}
	if v.Get("passport.authentication.enabled") == nil {
		v.Set("passport.authentication.enabled", true)
	}

	err := v.UnmarshalKey("passport", &LocalConfig)
	if err != nil {
		logger.Error("viper unmarshal err:", err)
		return
	}
	// Set default value
	// 兼容 Java 端配置键名 access-key-id / secret-access-key
	if LocalConfig.AppAccount.AccessKeyId == "" {
		LocalConfig.AppAccount.AccessKeyId = LocalConfig.AppAccount.Ak
	}
	if LocalConfig.AppAccount.SecretAccessKey == "" {
		LocalConfig.AppAccount.SecretAccessKey = LocalConfig.AppAccount.Sk
	}
	// 兼容旧版 platform 配置节
	if LocalConfig.BaseUrl == "" {
		LocalConfig.BaseUrl = v.GetString("passport.platform.base-url")
	}
	if LocalConfig.Cache.SyncPeriodSeconds == 0 {
		LocalConfig.Cache.SyncPeriodSeconds = v.GetInt("passport.platform.cache-sync-seconds")
	}
	if LocalConfig.BaseUrl == "" {
		LocalConfig.BaseUrl = "https://passport.vancone.com"
	}
	if LocalConfig.Cache.SyncPeriodSeconds == 0 {
		LocalConfig.Cache.SyncPeriodSeconds = 60
	}
	if LocalConfig.Token.Algorithm == "" {
		LocalConfig.Token.Algorithm = "ES256"
	}
}
