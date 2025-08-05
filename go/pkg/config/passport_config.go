package config

import "github.com/spf13/viper"

var Viper *viper.Viper
var SelfConfig PassportConfig

type PassportConfig struct {
	AccessControl  AccessControlConfig  `mapstructure:"access-control"`
	BaseUrl        string               `mapstructure:"base-url"`
	Cache          CacheConfig          `mapstructure:"cache"`
	ServiceAccount ServiceAccountConfig `mapstructure:"service-account"`
	Token          TokenConfig          `mapstructure:"token"`
}

type AccessControlConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

type CacheConfig struct {
	SyncPeriodSeconds int `mapstructure:"sync-period-seconds"`
}

type ServiceAccountConfig struct {
	AccessKeyId     string `mapstructure:"access-key-id"`
	SecretAccessKey string `mapstructure:"secret-access-key"`
}

type TokenConfig struct {
	Algorithm string `mapstructure:"algorithm"`
	PublicKey string `mapstructure:"public-key"`
}
