package cmd

const (
	PgDump = "pg_dump"
	Gzip   = "gzip"
	MC     = "mc"
)

const (
	Alias          = "minio"
	PgDumpFile     = "pg_dump.sql.gz"
	BackupSchedule = "SCHEDULE"
)

const (
	PostgresHost      = "POSTGRES_HOST"
	PostgresPort      = "POSTGRES_PORT"
	PostgresUser      = "POSTGRES_USER"
	PostgresPassword  = "POSTGRES_PASSWORD"
	PostgresDatabase  = "POSTGRES_DATABASE"
	PostgresExtraOpts = "POSTGRES_EXTRA_OPTS"
)

const (
	MinioAccessKey  = "MINIO_ACCESS_KEY"
	MinioSecretKey  = "MINIO_SECRET_KEY"
	MinioServer     = "MINIO_SERVER"
	MinioBucket     = "MINIO_BUCKET"
	MinioApiVersion = "MINIO_API_VERSION"
	MinioClean      = "MINIO_CLEAN"
	MinioBackupDir  = "MINIO_BACKUP_DIR"
)

const (
	TelegramEnabled = "TELEGRAM_ENABLED"
	TelegramApiUrl  = "TELEGRAM_API_URL"
	TelegramChatId  = "TELEGRAM_CHAT_ID"
	TelegramToken   = "TELEGRAM_TOKEN"
)

const (
	SendMessage = "sendMessage"
)

type status int

const (
	StatusOK status = iota
	StatusErr
)
