package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

func storage(cfg Minio, dbName string) error {
	err := aliasSet(cfg)
	if err != nil {
		log.Err(err).Msg("Failed to set alias")
		return err
	}

	bucket := fmt.Sprintf("%s/%s", Alias, cfg.Bucket)
	backupDir := fmt.Sprintf("%s/%s", bucket, dbName)

	minioBackupDir := cfg.BackupDir
	if minioBackupDir != "" {
		backupDir = fmt.Sprintf("%s/%s/%s", bucket, minioBackupDir, dbName)
	}

	err = mcCopy(backupDir, dbName)
	if err != nil {
		log.Err(err).Msg("Failed to copy")
		return err
	}

	if cfg.Clean == "" {
		return nil
	}

	err = mcClean(backupDir, cfg.Clean)
	if err != nil {
		log.Err(err).Msg("Failed to clean")
		return err
	}

	return nil
}

func aliasSet(cfg Minio) error {
	args := []string{
		"alias",
		"set",
		Alias,
		cfg.Server,
		cfg.AccessKey,
		cfg.SecretKey,
		"--api", cfg.ApiVersion,
	}

	log.Info().Msgf("Executing: %s %s", MC, replaceMinioSecret(strings.Join(args, " ")))
	mcCmd := exec.Command(MC, args...)
	mcCmd.Stdout = os.Stdout
	mcCmd.Stderr = os.Stderr

	return mcCmd.Run()
}

func mcCopy(backupDir string, dbName string) error {
	now := time.Now().Format(time.RFC3339)
	fileName := fmt.Sprintf("%s_%s.sql.gz", dbName, now)

	args := []string{
		"cp",
		fmt.Sprintf("./%s", PgDumpFile),
	}

	args = append(args, fmt.Sprintf("%s/%s", backupDir, fileName))

	log.Info().Msgf("Executing: %s %s", MC, replaceMinioSecret(strings.Join(args, " ")))
	mcCmd := exec.Command(MC, args...)
	mcCmd.Stdout = os.Stdout
	mcCmd.Stderr = os.Stderr

	return mcCmd.Run()
}

func mcClean(backupDir, clean string) error {
	args := []string{
		"find",
		backupDir,
		"--older-than",
		clean,
		"--exec",
		"mc rm {}",
	}

	log.Info().Msgf("Executing: %s %s", MC, replaceMinioSecret(strings.Join(args, " ")))
	mcCmd := exec.Command(MC, args...)
	mcCmd.Stdout = os.Stdout
	mcCmd.Stderr = os.Stderr

	return mcCmd.Run()
}
