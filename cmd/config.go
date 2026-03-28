package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Action      Action `json:"action" env:"ACTION" env-default:"backup"`
	Schedule    `json:"schedule"`
	Postgres    `json:"postgres"`
	Minio       `json:"minio"`
	Telegram    `json:"telegram"`
	Metrics     `json:"metrics"`
	HTTPTrigger `json:"http_trigger"`
	Restore     `json:"restore"`
}

type Schedule struct {
	Cron string `env:"SCHEDULE"`
}

// Action defines the operation mode: backup or restore
type Action string

const (
	ActionBackup  Action = "backup"
	ActionRestore Action = "restore"
)

// HTTPTrigger config for manual backup trigger via HTTP
type HTTPTrigger struct {
	Enable bool   `env:"HTTP_TRIGGER_ENABLED" env-default:"false"`
	Port   string `env:"HTTP_TRIGGER_PORT" env-default:"8081"`
	Path   string `env:"HTTP_TRIGGER_PATH" env-default:"/trigger"`
}

// Restore config for restore functionality
type Restore struct {
	Enable     bool   `env:"RESTORE" env-default:"false"`
	SourcePath string `env:"RESTORE_SOURCE_PATH"` // Path to backup file in Minio
	TargetDB   string `env:"RESTORE_TARGET_DB"`   // Target database name (optional, uses source name)
}

type Postgres struct {
	Prerun    bool   `env:"POSTGRES_PRERUN" env-default:"true"`
	Host      string `env:"POSTGRES_HOST" env-default:"localhost"`
	Port      int    `env:"POSTGRES_PORT" env-default:"5432"`
	User      string `env:"POSTGRES_USER" env-default:"postgres"`
	Password  string `env:"POSTGRES_PASSWORD"`
	Database  string `env:"POSTGRES_DATABASE"`
	Format    string `env:"POSTGRES_FORMAT" env-default:"custom"` // custom, directory, plain
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

type Metrics struct {
	Enable    bool   `env:"METRICS_ENABLED" env-default:"false"`
	Namespace string `env:"METRICS_NAMESPACE"`
	Subsystem string `env:"METRICS_SUBSYSTEM"`
	Port      string `env:"METRICS_PORT" env-default:"8080"`
	Path      string `env:"METRICS_PATH" env-default:"/metrics"`
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
