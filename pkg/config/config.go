package config

import (
	"log"

	"github.com/spf13/viper"
)

var Viper *viper.Viper
var LocalConfig PassportConfig

type PassportConfig struct {
	AccessControl AccessControlConfig `mapstructure:"access-control"`
	AppAccount    AppAccountConfig    `mapstructure:"app-account"`
	BaseUrl       string              `mapstructure:"base-url"`
	Cache         CacheConfig         `mapstructure:"cache"`
	Token         TokenConfig         `mapstructure:"token"`
}

type AccessControlConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

type AppAccountConfig struct {
	AccessKeyId     string `mapstructure:"access-key-id"`
	SecretAccessKey string `mapstructure:"secret-access-key"`
}

type CacheConfig struct {
	SyncPeriodSeconds int `mapstructure:"sync-period-seconds"`
}

type TokenConfig struct {
	Algorithm string `mapstructure:"algorithm"`
	PublicKey string `mapstructure:"public-key"`
}

func Init(viper *viper.Viper) {
	Viper = viper
	err := viper.UnmarshalKey("passport", &LocalConfig)
	if err != nil {
		log.Println("viper unmarshal err:", err)
		return
	}
}
