package bootstrap

import (
	"errors"

	"github.com/spf13/viper"
)

func validate() error {
	// POSTGRESQL
	if viper.GetString("POSTGRES_HOST") == "" {
		return errors.New("You need to set the POSTGRES_HOST environment variable")
	}

	if viper.GetString("POSTGRES_USER") == "" {
		return errors.New("You need to set the POSTGRES_USER environment variable")
	}

	if viper.GetString("POSTGRES_PASSWORD") == "" {
		return errors.New("You need to set the POSTGRES_PASSWORD environment variable")
	}

	if viper.GetString("POSTGRES_DATABASE") == "" {
		return errors.New("You need to set the POSTGRES_DATABASE environment variable")
	}

	// MINIO
	if viper.GetString("MINIO_ACCESS_KEY") == "" {
		return errors.New("You need to set the MINIO_ACCESS_KEY environment variable")
	}

	if viper.GetString("MINIO_SECRET_KEY") == "" {
		return errors.New("You need to set the MINIO_SECRET_KEY environment variable")
	}

	if viper.GetString("MINIO_SERVER") == "" {
		return errors.New("You need to set the MINIO_SERVER environment variable")
	}

	if viper.GetString("MINIO_BUCKET") == "" {
		return errors.New("You need to set the MINIO_BUCKET environment variable")
	}

	return nil
}
