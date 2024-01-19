package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Schedule `yaml:"schedule" json:"schedule"`
	Postgres `yaml:"postgres" json:"postgres"`
	Minio    `yaml:"minio" json:"minio"`
	Telegram `yaml:"telegram" json:"telegram"`
}

type Schedule struct {
	Cron string `yaml:"cron" env:"SCHEDULE"`
}

type Postgres struct {
	Host      string `yaml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
	Port      int    `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
	User      string `yaml:"user" env:"POSTGRES_USER" env-default:"postgres"`
	Password  string `env-required:"true" yaml:"user" env:"POSTGRES_PASSWORD"`
	Database  string `env-required:"true" yaml:"database" env:"POSTGRES_DATABASE"`
	ExtraOpts string `yaml:"extraOpts" env:"POSTGRES_EXTRA_OPTS"`
}

type Minio struct {
	AccessKey  string `env-required:"true" yaml:"accessKey" env:"MINIO_ACCESS_KEY"`
	SecretKey  string `env-required:"true" yaml:"secretKey" env:"MINIO_SECRET_KEY"`
	Server     string `env-required:"true" yaml:"server" env:"MINIO_SERVER"`
	Bucket     string `env-required:"true" yaml:"bucket" env:"MINIO_BUCKET"`
	ApiVersion string `yaml:"apiVersion" env:"MINIO_API_VERSION" env-default:"S3v4"`
	Clean      string `yaml:"clean" env:"MINIO_CLEAN"`
	BackupDir  string `yaml:"backupDir" env:"MINIO_BACKUP_DIR"`
}

type Telegram struct {
	Enable bool   `yaml:"enable" env:"TELEGRAM_ENABLED"`
	ChatId int64  `yaml:"chatId" env:"TELEGRAM_CHAT_ID"`
	Token  string `yaml:"token" env:"TELEGRAM_TOKEN"`
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
