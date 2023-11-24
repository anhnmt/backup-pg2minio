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

func removeFile(name string) error {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return nil
	}

	err = os.Remove(name)
	if err != nil {
		log.Err(err).Msg("Error remove file")
		return err
	}

	return nil
}

func checkEnvString(key string) error {
	if viper.GetString(key) == "" {
		return fmt.Errorf("You need to set the %s environment variable", key)
	}

	return nil
}

func replacePostgresql(str string) string {
	regex, err := regexp.Compile(`-d (.*) `)
	if err != nil {
		return str
	}

	secret := "******"
	matches := regex.FindStringSubmatch(str)
	if len(matches) >= 2 {
		replaced := regex.ReplaceAllString(str, fmt.Sprintf("-d %s ", secret))
		return replaced
	}

	return str
}

func replaceMinioSecret(str string) string {
	regex, err := regexp.Compile(`set minio (.*) (.*) (.*) --api`)
	if err != nil {
		return str
	}

	secret := "******"
	matches := regex.FindStringSubmatch(str)
	if len(matches) >= 4 {
		// matches[1] = host
		// matches[2] = access key
		// matches[3] = secret key
		replaced := regex.ReplaceAllString(str, fmt.Sprintf("set minio %s --api", secret))
		return replaced
	}

	return str
}
