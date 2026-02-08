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
	CsrfConfig     CsrfConfig           `mapstructure:"csrf"`
	PlatformConfig PlatformConfig       `mapstructure:"platform"`
	Token          TokenConfig          `mapstructure:"token"`
}

type AccessControlConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	UriAllowlist []string `mapstructure:"uri-allowlist"`
}

type AppAccountConfig struct {
	Ak string `mapstructure:"ak"`
	Sk string `mapstructure:"sk"`
}

type AuthenticationConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	UriAllowlist []string `mapstructure:"uri-allowlist"`
}

type CsrfConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	SecretKey string `mapstructure:"secret-key"`
}

type PlatformConfig struct {
	BaseUrl          string `mapstructure:"base-url"`
	CacheSyncSeconds int    `mapstructure:"cache-sync-seconds"`
}

type TokenConfig struct {
	Algorithm string `mapstructure:"algorithm"`
	PublicKey string `mapstructure:"public-key"`
}

func Init(viper *viper.Viper) {
	Viper = viper
	accessControlEnabled := viper.Get("passport.access-control.enabled")
	if accessControlEnabled == nil {
		viper.Set("passport.access-control.enabled", true)
	}
	authenticationEnabled := viper.Get("passport.authentication.enabled")
	if authenticationEnabled == nil {
		viper.Set("passport.authentication.enabled", true)
	}

	err := viper.UnmarshalKey("passport", &LocalConfig)
	if err != nil {
		logger.Error("viper unmarshal err:", err)
		return
	}
	// Set default value
	if LocalConfig.PlatformConfig.BaseUrl == "" {
		LocalConfig.PlatformConfig.BaseUrl = "https://passport.vancone.com"
	}
	if LocalConfig.PlatformConfig.CacheSyncSeconds == 0 {
		LocalConfig.PlatformConfig.CacheSyncSeconds = 60
	}
	if LocalConfig.Token.Algorithm == "" {
		LocalConfig.Token.Algorithm = "ES256"
	}
}
