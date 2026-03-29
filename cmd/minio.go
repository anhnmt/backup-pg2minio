package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

func storage(cfg Minio, dbName string, format string) error {
	// Create MinIO client
	mc, err := NewMinioClient(cfg)
	if err != nil {
		log.Err(err).Msg("Failed to create MinIO client")
		return err
	}

	ctx := context.Background()

	// Check bucket exists
	if err := mc.BucketExists(ctx); err != nil {
		log.Err(err).Msg("Bucket does not exist or inaccessible")
		return err
	}

	// backupDir là subfolder bên trong bucket, KHÔNG phải bucket name
	// cfg.Bucket đã được dùng trong UploadFile/DownloadFile rồi
	backupDir := cfg.BackupDir

	// Create upload filename with timestamp
	now := time.Now().Format(time.RFC3339)
	uploadFileName := getUploadFileName(dbName, now, format)

	// Build object key: "subfolder/filename" hoặc chỉ "filename" nếu không có BackupDir
	var objectName string
	if backupDir != "" {
		objectName = fmt.Sprintf("%s/%s", backupDir, uploadFileName)
	} else {
		objectName = uploadFileName
	}
	fileName := getDumpFileName(format)

	err = mc.UploadFile(ctx, fileName, objectName)
	if err != nil {
		log.Err(err).Msg("Failed to upload")
		return err
	}

	// Clean old backups if configured
	if cfg.Clean != "" {
		// prefix dùng backupDir (subfolder), không prepend bucket name
		prefix := backupDir

		oldObjects, err := mc.GetObjectsOlderThan(ctx, prefix, cfg.Clean)
		if err != nil {
			log.Err(err).Msg("Failed to get old objects for cleaning")
		} else if len(oldObjects) > 0 {
			err = mc.DeleteObjects(ctx, oldObjects)
			if err != nil {
				log.Err(err).Msg("Failed to clean old backups")
			} else {
				log.Info().Msgf("Cleaned %d old backup(s)", len(oldObjects))
			}
		}
	}

	return nil
}

// getUploadFileName returns the appropriate upload filename based on dump format
func getUploadFileName(dbName string, timestamp string, format string) string {
	switch strings.ToLower(format) {
	case "custom":
		return fmt.Sprintf("%s_%s.custom.gz", dbName, timestamp)
	case "directory":
		return fmt.Sprintf("%s_%s.backup.tar.gz", dbName, timestamp)
	case "plain", "sql":
		return fmt.Sprintf("%s_%s.sql.gz", dbName, timestamp)
	case "tar":
		return fmt.Sprintf("%s_%s.tar.gz", dbName, timestamp)
	default:
		return fmt.Sprintf("%s_%s.custom.gz", dbName, timestamp)
	}
}
