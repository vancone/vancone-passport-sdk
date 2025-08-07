package passport_sdk

import (
	"github.com/spf13/viper"
	"github.com/vancone/vancone-passport-sdk-go/pkg/config"
)

func Init(viper *viper.Viper, selfConfig config.PassportConfig) {
	config.Viper = viper
	config.SelfConfig = selfConfig
}
