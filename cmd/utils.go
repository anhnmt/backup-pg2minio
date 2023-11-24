package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func createFile(name string) (*os.File, error) {
	outputFile, err := os.Create(name)
	if err != nil {
		log.Err(err).Msg("Error creating output file")
		return nil, err
	}

	return outputFile, nil
}

func checkEnvString(key string) error {
	if viper.GetString(key) == "" {
		return fmt.Errorf("You need to set the %s environment variable", key)
	}

	return nil
}
