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

	defaultConfig()
}

func defaultConfig() {
	// POSTGRES
	viper.SetDefault("POSTGRES_PORT", 5432)
	viper.SetDefault("POSTGRES_EXTRA_OPTS", "--inserts --clean --if-exists --no-owner --no-acl --blobs --schema=public --no-sync --rows-per-insert=5000 --format=plain")

	// MINIO
	viper.SetDefault("MINIO_API_VERSION", "S3v4")
}
