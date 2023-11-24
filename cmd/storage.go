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

	err = mcCopy()
	if err != nil {
		log.Err(err).Msg("Failed to copy")
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

func mcCopy() error {
	now := time.Now().Format(time.RFC3339)
	fileName := fmt.Sprintf("%s_%s.sql.gz", viper.GetString(PostgresDatabase), now)
	bucketPath := fmt.Sprintf("minio/%s", viper.GetString(MinioBucket))

	args := []string{
		"cp",
		fmt.Sprintf("./%s", PgDumpFile),
	}

	backupDir := fmt.Sprintf("%s/%s", bucketPath, fileName)

	minioBackupDir := viper.GetString(MinioBackupDir)
	if minioBackupDir != "" {
		backupDir = fmt.Sprintf("%s/%s/%s", bucketPath, minioBackupDir, fileName)
	}
	args = append(args, backupDir)

	log.Info().Msgf("Executing: %s %s", MC, replaceMinioSecret(strings.Join(args, " ")))
	mcCmd := exec.Command(MC, args...)
	mcCmd.Stdout = os.Stdout
	mcCmd.Stderr = os.Stderr

	return mcCmd.Run()
}
