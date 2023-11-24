package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func bootstrap() {
	logger()
	config()

	log.Info().
		Str("goarch", runtime.GOARCH).
		Str("goos", runtime.GOOS).
		Str("version", runtime.Version()).
		Msg("Bootstrap successfully")

	err := validate()
	if err != nil {
		log.Panic().Err(err).Msg("Failed to validate config")
		return
	}
}

func logger() {
	// UNIX Time is faster and smaller than most timestamps
	consoleWriter := &zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}

	// Caller Marshal Function
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		file = short
		return file + ":" + strconv.Itoa(line)
	}

	log.Logger = zerolog.
		New(consoleWriter).
		With().
		Timestamp().
		Caller().
		Logger()
}

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
	viper.SetDefault(PostgresPort, 5432)
	viper.SetDefault(PostgresExtraOpts, "--inserts --clean --if-exists --no-owner --no-acl --blobs --schema=public --no-sync --rows-per-insert=5000 --format=plain")

	// MINIO
	viper.SetDefault(MinioApiVersion, "S3v4")
}

func validate() error {
	// POSTGRESQL
	if viper.GetString(PostgresHost) == "" {
		return fmt.Errorf("You need to set the %s environment variable", PostgresHost)
	}

	if viper.GetString(PostgresUser) == "" {
		return fmt.Errorf("You need to set the %s environment variable", PostgresUser)
	}

	if viper.GetString(PostgresPassword) == "" {
		return fmt.Errorf("You need to set the %s environment variable", PostgresPassword)
	}

	if viper.GetString(PostgresDatabase) == "" {
		return fmt.Errorf("You need to set the %s environment variable", PostgresDatabase)
	}

	// MINIO
	if viper.GetString(MinioAccessKey) == "" {
		return fmt.Errorf("You need to set the %s environment variable", MinioAccessKey)
	}

	if viper.GetString(MinioSecretKey) == "" {
		return fmt.Errorf("You need to set the %s environment variable", MinioSecretKey)
	}

	if viper.GetString(MinioServer) == "" {
		return fmt.Errorf("You need to set the %s environment variable", MinioServer)
	}

	if viper.GetString(MinioBucket) == "" {
		return fmt.Errorf("You need to set the %s environment variable", MinioBucket)
	}

	return nil
}
