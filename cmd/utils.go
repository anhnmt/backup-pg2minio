package cmd

import (
	"os"

	"github.com/rs/zerolog/log"
)

func CreateFile(name string) (*os.File, error) {
	outputFile, err := os.Create(name)
	if err != nil {
		log.Err(err).Msg("Error creating output file")
		return nil, err
	}

	return outputFile, nil
}
