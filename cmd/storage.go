package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func storage() error {
	err := aliasSet()
	if err != nil {
		log.Err(err).Msg("Failed to set alias")
		return err
	}

	bucket := fmt.Sprintf("minio/%s", viper.GetString(MinioBucket))
	backupDir := fmt.Sprintf("%s/%s", bucket, viper.GetString(PostgresDatabase))

	minioBackupDir := viper.GetString(MinioBackupDir)
	if minioBackupDir != "" {
		backupDir = fmt.Sprintf("%s/%s/%s", bucket, minioBackupDir, viper.GetString(PostgresDatabase))
	}

	err = mcCopy(backupDir)
	if err != nil {
		log.Err(err).Msg("Failed to copy")
		return err
	}

	clean := viper.GetString(MinioClean)
	if clean == "" {
		return nil
	}

	err = mcClean(backupDir, clean)
	if err != nil {
		log.Err(err).Msg("Failed to clean")
		return err
	}

	return nil
}

func aliasSet() error {
	args := []string{
		"alias",
		"set",
		"minio",
		viper.GetString(MinioServer),
		viper.GetString(MinioAccessKey),
		viper.GetString(MinioSecretKey),
		"--api", viper.GetString(MinioApiVersion),
	}

	log.Info().Msgf("Executing: %s %s", MC, replaceMinioSecret(strings.Join(args, " ")))
	mcCmd := exec.Command(MC, args...)
	mcCmd.Stdout = os.Stdout
	mcCmd.Stderr = os.Stderr

	return mcCmd.Run()
}

func mcCopy(backupDir string) error {
	now := time.Now().Format(time.RFC3339)
	fileName := fmt.Sprintf("%s_%s.sql.gz", viper.GetString(PostgresDatabase), now)

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
