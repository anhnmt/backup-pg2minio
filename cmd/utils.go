package cmd

import (
	"fmt"
	"os"
	"regexp"

	"github.com/rs/zerolog/log"
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
	err := os.Remove(name)
	if err != nil {
		log.Err(err).Msg("Error remove file")
		return err
	}

	return nil
}

func replacePostgresql(str string) string {
	regex, err := regexp.Compile(` postgresql://(.*) `)
	if err != nil {
		return str
	}

	secret := "******"
	matches := regex.FindStringSubmatch(str)
	if len(matches) >= 2 {
		replaced := regex.ReplaceAllString(str, fmt.Sprintf(" postgresql://%s ", secret))
		return replaced
	}

	return str
}

func replaceMinioSecret(str string) string {
	regex, err := regexp.Compile(`set ` + Alias + ` (.*) (.*) (.*) --api`)
	if err != nil {
		return str
	}

	secret := "******"
	matches := regex.FindStringSubmatch(str)
	if len(matches) >= 4 {
		// matches[1] = host
		// matches[2] = access key
		// matches[3] = secret key
		replaced := regex.ReplaceAllString(str, fmt.Sprintf("set %s %s --api", Alias, secret))
		return replaced
	}

	return str
}
