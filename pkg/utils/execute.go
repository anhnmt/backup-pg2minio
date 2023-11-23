package utils

import (
	"os"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

func Execute(command string, args ...string) error {
	log.Info().Msgf("Executing: %s %s", command, strings.Join(args, " "))

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
