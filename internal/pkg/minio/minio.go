package minio

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/anhnmt/backup-pg2minio/internal/pkg/config"
	"github.com/anhnmt/backup-pg2minio/internal/utils"
)

func Storage(cfg config.Minio, dbName string) error {
	bucket := fmt.Sprintf("%s/%s", utils.Alias, cfg.Bucket)
	backupDir := fmt.Sprintf("%s/%s", bucket, dbName)

	if cfg.BackupDir != "" {
		backupDir = fmt.Sprintf("%s/%s/%s", bucket, cfg.BackupDir, dbName)
	}

	err := mcCopy(cfg, backupDir, dbName)
	if err != nil {
		log.Err(err).Msg("Failed to copy")
		return err
	}

	if cfg.Clean != "" {
		err = mcClean(cfg, backupDir, cfg.Clean)
		if err != nil {
			log.Err(err).Msg("Failed to clean")
			return err
		}
	}

	return nil
}

func AliasSet(cfg config.Minio) error {
	args := []string{
		"alias",
		"set",
		utils.Alias,
		cfg.Server,
		cfg.AccessKey,
		cfg.SecretKey,
		"--api", cfg.ApiVersion,
	}

	if cfg.Insecure {
		args = append(args, "--insecure")
	}

	if cfg.Debug {
		args = append(args, "--debug")
	}

	log.Info().Msgf("Executing: %s %s", utils.MC, utils.ReplaceMinioSecret(strings.Join(args, " ")))
	mcCmd := exec.Command(utils.MC, args...)
	mcCmd.Stdout = os.Stdout
	mcCmd.Stderr = os.Stderr

	return mcCmd.Run()
}

func PreRunMinio(cfg config.Minio) error {
	args := []string{
		"version",
		"info",
		fmt.Sprintf("%s/%s", utils.Alias, cfg.Bucket),
		"-q",
	}

	if cfg.Insecure {
		args = append(args, "--insecure")
	}

	if cfg.Debug {
		args = append(args, "--debug")
	}

	log.Info().Msgf("Executing: %s %s", utils.MC, strings.Join(args, " "))
	mcCmd := exec.Command(utils.MC, args...)
	mcCmd.Stdout = os.Stdout
	mcCmd.Stderr = os.Stderr

	return mcCmd.Run()
}

func mcCopy(cfg config.Minio, backupDir string, dbName string) error {
	now := time.Now().Format(time.RFC3339)
	fileName := fmt.Sprintf("%s_%s.sql.gz", dbName, now)

	args := []string{
		"cp",
		fmt.Sprintf("./%s", utils.PgDumpFile),
	}

	if cfg.Insecure {
		args = append(args, "--insecure")
	}

	if cfg.Debug {
		args = append(args, "--debug")
	}

	args = append(args, fmt.Sprintf("%s/%s", backupDir, fileName))

	log.Info().Msgf("Executing: %s %s", utils.MC, utils.ReplaceMinioSecret(strings.Join(args, " ")))
	mcCmd := exec.Command(utils.MC, args...)
	mcCmd.Stdout = os.Stdout
	mcCmd.Stderr = os.Stderr

	return mcCmd.Run()
}

func mcClean(cfg config.Minio, backupDir, clean string) error {
	args := []string{
		"find",
		backupDir,
		"--older-than",
		clean,
		"--exec",
		"mc rm {}",
	}

	if cfg.Insecure {
		args = append(args, "--insecure")
	}

	if cfg.Debug {
		args = append(args, "--debug")
	}

	log.Info().Msgf("Executing: %s %s", utils.MC, utils.ReplaceMinioSecret(strings.Join(args, " ")))
	mcCmd := exec.Command(utils.MC, args...)
	mcCmd.Stdout = os.Stdout
	mcCmd.Stderr = os.Stderr

	return mcCmd.Run()
}
