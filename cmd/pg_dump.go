package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

func pgDump(cfg Postgres) error {
	// Validate: --jobs option only works with directory format
	if strings.Contains(cfg.ExtraOpts, "--jobs") && strings.ToLower(cfg.Format) != "directory" {
		return fmt.Errorf("--jobs option requires POSTGRES_FORMAT=directory, got %s", cfg.Format)
	}

	conn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	format := getDumpFormat(cfg.Format)
	fileName := getDumpFileName(cfg.Format)

	args := []string{
		"-d",
		conn,
		"-f",
		format,
	}

	pgOpts := cfg.ExtraOpts
	if pgOpts != "" {
		pgOpts = strings.TrimSpace(pgOpts)
		args = append(args, strings.Fields(pgOpts)...)
	}

	return executePgDump(args, fileName, cfg.Format)
}

// getDumpFormat returns the pg_dump format flag based on the format string
func getDumpFormat(format string) string {
	switch strings.ToLower(format) {
	case "custom":
		return "Fc"
	case "directory":
		return "Fd"
	case "plain", "sql":
		return "Fp"
	case "tar":
		return "Ft"
	default:
		return "Fc"
	}
}

// getDumpFileName returns the appropriate filename based on dump format
func getDumpFileName(format string) string {
	switch strings.ToLower(format) {
	case "custom":
		return PgDumpFileCustom
	case "directory":
		return PgDumpFileDirectory
	case "plain", "sql":
		return PgDumpFilePlain
	case "tar":
		return PgDumpFileTar
	default:
		return PgDumpFileCustom
	}
}

func preRunPostgres(cfg Postgres) error {
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

	log.Info().Msgf("Executing: %s %s", PSQL, replacePostgresql(strings.Join(args, " ")))
	psqlCmd := exec.Command(PSQL, args...)
	psqlCmd.Stderr = os.Stderr

	return psqlCmd.Run()
}

func executePgDump(args []string, fileName string, format string) error {
	log.Info().Msgf("Executing: %s %s", PgDump, replacePostgresql(strings.Join(args, " ")))
	pgDumpCmd := exec.Command(PgDump, args...)
	pgDumpCmd.Stderr = os.Stderr

	// Handle directory format differently - it creates a directory instead of stdout
	if strings.ToLower(format) == "directory" {
		return executePgDumpDirectory(pgDumpCmd, fileName)
	}

	// Create a pipe to connect the stdout of pg_dump to the stdin of gzip
	pipe, err := pgDumpCmd.StdoutPipe()
	if err != nil {
		log.Err(err).Msg("Error creating pipe")
		return err
	}

	// Start pg_dump command
	if err = pgDumpCmd.Start(); err != nil {
		log.Err(err).Msgf("Error start %s command", PgDump)
		return err
	}

	// Create the gzip command and link its stdin to the output of pg_dump
	log.Info().Msgf("Executing: %s", Gzip)
	gzipCmd := exec.Command(Gzip)
	gzipCmd.Stdin = pipe
	gzipCmd.Stderr = os.Stderr

	// Create a file to save the output of gzip
	outputFile, err := createFile(fileName)
	if err != nil {
		return err
	}

	defer outputFile.Close()
	gzipCmd.Stdout = outputFile

	// Start gzip command
	if err = gzipCmd.Start(); err != nil {
		log.Err(err).Msgf("Error start %s command", Gzip)
		return err
	}

	// Wait for both commands to finish
	if err = pgDumpCmd.Wait(); err != nil {
		log.Err(err).Msgf("Error waiting for %s command", PgDump)
		return err
	}

	if err = gzipCmd.Wait(); err != nil {
		log.Err(err).Msgf("Error waiting for %s command", Gzip)
		return err
	}

	return nil
}

// executePgDumpDirectory handles the directory format which creates a directory
func executePgDumpDirectory(cmd *exec.Cmd, dirName string) error {
	// Create the output directory
	if err := os.MkdirAll(dirName, 0755); err != nil {
		log.Err(err).Msgf("Error creating directory %s", dirName)
		return err
	}

	// Set the output directory for pg_dump
	cmd.Dir = dirName

	// Start pg_dump command
	if err := cmd.Start(); err != nil {
		log.Err(err).Msgf("Error start %s command", PgDump)
		return err
	}

	// Wait for pg_dump to finish
	if err := cmd.Wait(); err != nil {
		log.Err(err).Msgf("Error waiting for %s command", PgDump)
		return err
	}

	// Compress the directory with tar and gzip
	if err := compressDirectory(dirName); err != nil {
		return err
	}

	return nil
}

// compressDirectory creates a tar.gz archive from a directory
func compressDirectory(dirName string) error {
	// Create the tar.gz file
	outputFile, err := createFile(dirName + ".tar.gz")
	if err != nil {
		return err
	}
	defer outputFile.Close()

	// Create tar command
	tarCmd := exec.Command("tar", "-czf", "-", ".")
	tarCmd.Dir = dirName
	tarCmd.Stdout = outputFile
	tarCmd.Stderr = os.Stderr

	if err := tarCmd.Run(); err != nil {
		log.Err(err).Msg("Error creating tar archive")
		return err
	}

	// Remove the original directory after compression
	if err := os.RemoveAll(dirName); err != nil {
		log.Err(err).Msgf("Error removing directory %s", dirName)
		return err
	}

	return nil
}
