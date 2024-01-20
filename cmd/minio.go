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

	if cfg.BackupDir != "" {
		backupDir = fmt.Sprintf("%s/%s/%s", bucket, cfg.BackupDir, dbName)
	}

	err = mcCopy(cfg, backupDir, dbName)
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

	if cfg.Insecure {
		args = append(args, "--insecure")
	}

	if cfg.Debug {
		args = append(args, "--debug")
	}

	log.Info().Msgf("Executing: %s %s", MC, replaceMinioSecret(strings.Join(args, " ")))
	mcCmd := exec.Command(MC, args...)
	mcCmd.Stdout = os.Stdout
	mcCmd.Stderr = os.Stderr

	return mcCmd.Run()
}

func mcCopy(cfg Minio, backupDir string, dbName string) error {
	now := time.Now().Format(time.RFC3339)
	fileName := fmt.Sprintf("%s_%s.sql.gz", dbName, now)

	args := []string{
		"cp",
		fmt.Sprintf("./%s", PgDumpFile),
	}

	if cfg.Insecure {
		args = append(args, "--insecure")
	}

	if cfg.Debug {
		args = append(args, "--debug")
	}

	args = append(args, fmt.Sprintf("%s/%s", backupDir, fileName))

	log.Info().Msgf("Executing: %s %s", MC, replaceMinioSecret(strings.Join(args, " ")))
	mcCmd := exec.Command(MC, args...)
	mcCmd.Stdout = os.Stdout
	mcCmd.Stderr = os.Stderr

	return mcCmd.Run()
}

func mcClean(cfg Minio, backupDir, clean string) error {
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

	log.Info().Msgf("Executing: %s %s", MC, replaceMinioSecret(strings.Join(args, " ")))
	mcCmd := exec.Command(MC, args...)
	mcCmd.Stdout = os.Stdout
	mcCmd.Stderr = os.Stderr

	return mcCmd.Run()
}
