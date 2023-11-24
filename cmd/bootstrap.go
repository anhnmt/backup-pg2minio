package cmd

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
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
	g, _ := errgroup.WithContext(context.Background())

	// POSTGRESQL
	g.Go(func() error {
		return checkEnvString(PostgresHost)
	})

	g.Go(func() error {
		return checkEnvString(PostgresUser)
	})

	g.Go(func() error {
		return checkEnvString(PostgresPassword)
	})

	g.Go(func() error {
		return checkEnvString(PostgresDatabase)
	})

	// MINIO
	g.Go(func() error {
		return checkEnvString(MinioAccessKey)
	})

	g.Go(func() error {
		return checkEnvString(MinioSecretKey)
	})

	g.Go(func() error {
		return checkEnvString(MinioServer)
	})

	g.Go(func() error {
		return checkEnvString(MinioBucket)
	})

	return g.Wait()
}
