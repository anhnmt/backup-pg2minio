package bootstrap

import (
	"strings"

	"github.com/spf13/viper"
)

func config() {
	viper.AutomaticEnv()
	// Replace env key
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.AddConfigPath(".")
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	_ = viper.ReadInConfig()
}
