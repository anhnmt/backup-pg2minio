package cmd

import (
	"fmt"
	"os"
	"regexp"

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

func replaceSecret(str string) string {
	regex, err := regexp.Compile(`postgresql:\/\/(.*?)\:(.*?)@`)
	if err != nil {
		return str
	}

	secret := "******"
	matches := regex.FindStringSubmatch(str)
	if len(matches) >= 3 {
		// matches[1] = username,
		// matches[2] = password
		replaced := regex.ReplaceAllString(str, fmt.Sprintf("postgresql://$1:%s@", secret))
		return replaced
	}

	return str
}
