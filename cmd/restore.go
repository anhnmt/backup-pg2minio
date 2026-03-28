package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// RestoreConfig holds the configuration for restore operation
type RestoreConfig struct {
	Postgres   Postgres
	Minio      Minio
	SourcePath string
	TargetDB   string
}

// PerformRestore restores a database from Minio backup
func PerformRestore(cfg RestoreConfig) error {
	log.Info().Msgf("Starting restore from: %s", cfg.SourcePath)

	// Determine target database
	targetDB := cfg.TargetDB
	if targetDB == "" {
		targetDB = cfg.Postgres.Database
	}

	// Get the format from the backup file extension
	format := getFormatFromBackupFile(cfg.SourcePath)
	log.Info().Msgf("Detected format: %s", format)

	// Create a temporary directory for the restore
	tempDir := fmt.Sprintf("/tmp/restore_%d", time.Now().Unix())
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download the backup file from Minio
	localFile, err := downloadFromMinio(cfg.Minio, cfg.SourcePath, tempDir)
	if err != nil {
		return fmt.Errorf("failed to download backup from Minio: %w", err)
	}
	defer os.Remove(localFile)

	// Restore based on format
	if format == "plain" || format == "sql" {
		return restorePlain(cfg.Postgres, localFile, targetDB)
	}

	return restoreBinary(cfg.Postgres, localFile, format, targetDB)
}

// downloadFromMinio downloads a file from Minio
func downloadFromMinio(cfg Minio, sourcePath, destDir string) (string, error) {
	// Extract filename from source path
	parts := strings.Split(sourcePath, "/")
	fileName := parts[len(parts)-1]
	destPath := fmt.Sprintf("%s/%s", destDir, fileName)

	args := []string{
		"cp",
		fmt.Sprintf("minio/%s", sourcePath),
		destPath,
	}

	if cfg.Insecure {
		args = append(args, "--insecure")
	}

	if cfg.Debug {
		args = append(args, "--debug")
	}

	log.Info().Msgf("Executing: %s %s", MC, strings.Join(args, " "))
	mcCmd := exec.Command(MC, args...)
	mcCmd.Stdout = os.Stdout
	mcCmd.Stderr = os.Stderr

	if err := mcCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to download from Minio: %w", err)
	}

	return destPath, nil
}

// restorePlain restores a plain SQL backup
func restorePlain(cfg Postgres, filePath, targetDB string) error {
	conn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		targetDB,
	)

	// Decompress if it's a .gz file
	if strings.HasSuffix(filePath, ".gz") {
		decompressedFile := strings.TrimSuffix(filePath, ".gz")

		gzipCmd := exec.Command(Gunzip, "-c", filePath)
		outputFile, err := os.Create(decompressedFile)
		if err != nil {
			return fmt.Errorf("failed to create decompressed file: %w", err)
		}
		defer os.Remove(decompressedFile)

		gzipCmd.Stdout = outputFile
		gzipCmd.Stderr = os.Stderr

		if err := gzipCmd.Run(); err != nil {
			return fmt.Errorf("failed to decompress: %w", err)
		}
		filePath = decompressedFile
	}

	// Execute psql to restore
	args := []string{
		"-d", conn,
		"-f", filePath,
	}

	log.Info().Msgf("Executing: %s %s", PSQL, replacePostgresql(strings.Join(args, " ")))
	psqlCmd := exec.Command(PSQL, args...)
	psqlCmd.Stderr = os.Stderr

	if err := psqlCmd.Run(); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	log.Info().Msg("Plain SQL restore completed successfully")
	return nil
}

// restoreBinary restores a binary format backup (custom, directory, tar)
func restoreBinary(cfg Postgres, filePath, format, targetDB string) error {
	conn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		targetDB,
	)

	// Decompress if needed
	if strings.HasSuffix(filePath, ".gz") {
		decompressedFile := strings.TrimSuffix(filePath, ".gz")

		gzipCmd := exec.Command(Gunzip, "-c", filePath)
		outputFile, err := os.Create(decompressedFile)
		if err != nil {
			return fmt.Errorf("failed to create decompressed file: %w", err)
		}

		gzipCmd.Stdout = outputFile
		gzipCmd.Stderr = os.Stderr

		if err := gzipCmd.Run(); err != nil {
			return fmt.Errorf("failed to decompress: %w", err)
		}
		filePath = decompressedFile
		defer os.Remove(filePath)
	}

	// Determine pg_restore format flag
	restoreFormat := getRestoreFormat(format)

	args := []string{
		"-d", conn,
		"-f", restoreFormat,
		filePath,
	}

	pgOpts := cfg.ExtraOpts
	if pgOpts != "" {
		pgOpts = strings.TrimSpace(pgOpts)
		pgOpts = strings.ReplaceAll(pgOpts, "=", " ")
		args = append(args, strings.Split(pgOpts, " ")...)
	}

	log.Info().Msgf("Executing: %s %s", PgRestore, replacePostgresql(strings.Join(args, " ")))
	pgRestoreCmd := exec.Command(PgRestore, args...)
	pgRestoreCmd.Stderr = os.Stderr

	if err := pgRestoreCmd.Run(); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	log.Info().Msg("Binary restore completed successfully")
	return nil
}

// getRestoreFormat returns the pg_restore format flag based on format
func getRestoreFormat(format string) string {
	switch strings.ToLower(format) {
	case "custom":
		return "Fc"
	case "directory":
		return "Fd"
	case "tar":
		return "Ft"
	default:
		return "Fc"
	}
}

// getFormatFromBackupFile determines the format from the backup filename
func getFormatFromBackupFile(filePath string) string {
	lower := strings.ToLower(filePath)

	if strings.Contains(lower, ".sql.gz") || strings.Contains(lower, ".sql") {
		return "plain"
	}
	if strings.Contains(lower, ".custom.gz") {
		return "custom"
	}
	if strings.Contains(lower, ".backup") || strings.Contains(lower, ".directory") {
		return "directory"
	}
	if strings.Contains(lower, ".tar.gz") || strings.Contains(lower, ".tar") {
		return "tar"
	}

	// Default to custom format
	return "custom"
}

// ListBackups lists all available backups in Minio
func ListBackups(cfg Minio, dbName string) ([]string, error) {
	bucket := fmt.Sprintf("%s/%s", Alias, cfg.Bucket)
	backupDir := bucket

	if cfg.BackupDir != "" {
		backupDir = fmt.Sprintf("%s/%s", bucket, cfg.BackupDir)
	}
	backupDir = fmt.Sprintf("%s/%s", backupDir, dbName)

	args := []string{
		"ls",
		"--json",
		backupDir,
	}

	if cfg.Insecure {
		args = append(args, "--insecure")
	}

	if cfg.Debug {
		args = append(args, "--debug")
	}

	log.Info().Msgf("Executing: %s %s", MC, strings.Join(args, " "))
	mcCmd := exec.Command(MC, args...)
	mcCmd.Stdout = os.Stdout
	mcCmd.Stderr = os.Stderr

	if err := mcCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	// Note: mc ls --json output parsing would be needed here
	// For simplicity, returning empty and user can specify path manually
	return []string{}, nil
}
