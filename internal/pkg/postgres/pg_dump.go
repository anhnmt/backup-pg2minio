package postgres

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/anhnmt/backup-pg2minio/internal/pkg/config"
	"github.com/anhnmt/backup-pg2minio/internal/utils"
)

func PgDump(cfg config.Postgres) error {
	conn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	args := []string{
		"-d",
		conn,
	}

	pgOpts := cfg.ExtraOpts
	if pgOpts != "" {
		pgOpts = strings.TrimSpace(pgOpts)
		pgOpts = strings.ReplaceAll(pgOpts, "=", " ")

		args = append(args, strings.Split(pgOpts, " ")...)
	}

	return ExecutePgDump(args...)
}

func PreRunPostgres(cfg config.Postgres) error {
	conn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	args := []string{
		conn,
		"-c", "SELECT 1",
	}

	log.Info().Msgf("Executing: %s %s", utils.PSQL, utils.ReplacePostgresql(strings.Join(args, " ")))
	psqlCmd := exec.Command(utils.PSQL, args...)
	psqlCmd.Stderr = os.Stderr

	return psqlCmd.Run()
}

func ExecutePgDump(args ...string) error {
	log.Info().Msgf("Executing: %s %s", utils.PgDump, utils.ReplacePostgresql(strings.Join(args, " ")))
	pgDumpCmd := exec.Command(utils.PgDump, args...)
	pgDumpCmd.Stderr = os.Stderr

	// Create a pipe to connect the stdout of pg_dump to the stdin of gzip
	pipe, err := pgDumpCmd.StdoutPipe()
	if err != nil {
		log.Err(err).Msg("Error creating pipe")
		return err
	}

	// Start pg_dump command
	if err = pgDumpCmd.Start(); err != nil {
		log.Err(err).Msgf("Error start %s command", utils.PgDump)
		return err
	}

	// Create the gzip command and link its stdin to the output of pg_dump
	log.Info().Msgf("Executing: %s", utils.Gzip)
	gzipCmd := exec.Command(utils.Gzip)
	gzipCmd.Stdin = pipe
	gzipCmd.Stderr = os.Stderr

	// Create a file to save the output of gzip
	outputFile, err := os.Create(utils.PgDumpFile)
	if err != nil {
		log.Err(err).Msg("Error creating output file")
		return err
	}

	defer outputFile.Close()
	gzipCmd.Stdout = outputFile

	// Start gzip command
	if err = gzipCmd.Start(); err != nil {
		log.Err(err).Msgf("Error start %s command", utils.Gzip)
		return err
	}

	// Wait for both commands to finish
	if err = pgDumpCmd.Wait(); err != nil {
		log.Err(err).Msgf("Error waiting for %s command", utils.PgDump)
		return err
	}

	if err = gzipCmd.Wait(); err != nil {
		log.Err(err).Msgf("Error waiting for %s command", utils.Gzip)
		return err
	}

	return nil
}
