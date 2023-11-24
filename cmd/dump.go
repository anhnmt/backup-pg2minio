package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func pgDump() error {
	postgresql := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		viper.GetString(PostgresUser),
		viper.GetString(PostgresPassword),
		viper.GetString(PostgresHost),
		viper.GetString(PostgresPort),
		viper.GetString(PostgresDatabase),
	)

	args := []string{
		"-d", postgresql,
	}

	pgOpts := viper.GetString(PostgresExtraOpts)
	if pgOpts != "" {
		pgOpts = strings.TrimSpace(pgOpts)
		pgOpts = strings.ReplaceAll(pgOpts, "=", " ")

		args = append(args, strings.Split(pgOpts, " ")...)
	}

	return executePgDump(args...)
}

func executePgDump(args ...string) error {
	log.Info().Msgf("Executing: %s", PgDump)
	pgDumpCmd := exec.Command(PgDump, args...)

	// Create a pipe to connect the stdout of pg_dump to the stdin of gzip
	pipe, err := pgDumpCmd.StdoutPipe()
	if err != nil {
		log.Err(err).Msg("Error creating pipe")
		return err
	}

	// Start pg_dump command
	if err = pgDumpCmd.Start(); err != nil {
		log.Err(err).Msg("Error start pg_dump command")
		return err
	}

	// Create the gzip command and link its stdin to the output of pg_dump
	log.Info().Msgf("Executing: %s", Gzip)
	gzipCmd := exec.Command(Gzip)
	gzipCmd.Stdin = pipe

	// Create a file to save the output of gzip
	outputFile, err := createFile(PgDumpFile)
	if err != nil {
		return err
	}

	defer outputFile.Close()
	gzipCmd.Stdout = outputFile

	// Start gzip command
	if err = gzipCmd.Start(); err != nil {
		log.Err(err).Msg("Error start gzip command")
		return err
	}

	// Wait for both commands to finish
	if err = pgDumpCmd.Wait(); err != nil {
		log.Err(err).Msg("Error waiting for pg_dump command")
		return err
	}

	if err = gzipCmd.Wait(); err != nil {
		log.Err(err).Msg("Error waiting for gzip command")
		return err
	}

	return nil
}
