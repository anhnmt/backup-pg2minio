package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Schedule `json:"schedule"`
	Postgres `json:"postgres"`
	Minio    `json:"minio"`
	Telegram `json:"telegram"`
}

type Schedule struct {
	Cron string `env:"SCHEDULE"`
}

type Postgres struct {
	Prerun    bool   `env:"POSTGRES_PRERUN" env-default:"true"`
	Host      string `env:"POSTGRES_HOST" env-default:"localhost"`
	Port      int    `env:"POSTGRES_PORT" env-default:"5432"`
	User      string `env:"POSTGRES_USER" env-default:"postgres"`
	Password  string `env:"POSTGRES_PASSWORD"`
	Database  string `env:"POSTGRES_DATABASE"`
	ExtraOpts string `env:"POSTGRES_EXTRA_OPTS"`
}

type Minio struct {
	Prerun     bool   `env:"MINIO_PRERUN" env-default:"true"`
	AccessKey  string `env:"MINIO_ACCESS_KEY" env-required:"true"`
	SecretKey  string `env:"MINIO_SECRET_KEY" env-required:"true"`
	Server     string `env:"MINIO_SERVER" env-required:"true"`
	Bucket     string `env:"MINIO_BUCKET" env-required:"true"`
	ApiVersion string `env:"MINIO_API_VERSION" env-default:"S3v4"`
	Clean      string `env:"MINIO_CLEAN"`
	BackupDir  string `env:"MINIO_BACKUP_DIR"`
	Insecure   bool   `env:"MINIO_INSECURE"`
	Debug      bool   `env:"MINIO_DEBUG"`
}

type Telegram struct {
	Enable bool   `env:"TELEGRAM_ENABLED"`
	ChatId int64  `env:"TELEGRAM_CHAT_ID"`
	Token  string `env:"TELEGRAM_TOKEN"`
}

func New() (Config, error) {
	cfg := Config{}

	dir, err := os.Getwd()
	if err != nil {
		return cfg, fmt.Errorf("getwd error: %w", err)
	}

	path := fmt.Sprintf("%s/%s", dir, ".env")
	err = cleanenv.ReadConfig(filepath.ToSlash(path), &cfg)
	if err == nil {
		return cfg, nil
	}

	if !os.IsNotExist(err) {
		return cfg, fmt.Errorf("read config error: %w", err)
	}

	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("read env error: %w", err)
	}

	return cfg, nil
}
