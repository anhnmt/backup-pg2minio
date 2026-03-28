package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
)

// MinioClient wraps the MinIO SDK client
type MinioClient struct {
	client *minio.Client
	cfg    Minio
}

// NewMinioClient creates a new MinIO client
func NewMinioClient(cfg Minio) (*MinioClient, error) {
	u, err := url.Parse(cfg.Server)
	if err != nil {
		return nil, fmt.Errorf("failed to parse MinIO server URL: %w", err)
	}

	// Build endpoint from URL
	endpoint := u.Host
	if u.Port() == "" {
		if u.Scheme == "https" {
			endpoint = fmt.Sprintf("%s:443", endpoint)
		} else {
			endpoint = fmt.Sprintf("%s:9000", endpoint)
		}
	}

	// Determine if secure
	secure := u.Scheme == "https"

	options := &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: secure,
	}

	if cfg.Insecure {
		customTransport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		options.Transport = customTransport
	}

	client, err := minio.New(endpoint, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	log.Info().Msgf("MinIO client connected to %s (secure: %v)", endpoint, secure)

	return &MinioClient{
		client: client,
		cfg:    cfg,
	}, nil
}

// BucketExists checks if the bucket exists
func (m *MinioClient) BucketExists(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.cfg.Bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("bucket %s does not exist", m.cfg.Bucket)
	}
	return nil
}

// UploadFile uploads a file to MinIO
func (m *MinioClient) UploadFile(ctx context.Context, filePath, objectName string) error {
	_, err := m.client.FPutObject(ctx, m.cfg.Bucket, objectName, filePath, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	log.Info().Msgf("Uploaded %s to %s/%s", filePath, m.cfg.Bucket, objectName)
	return nil
}

// DownloadFile downloads a file from MinIO
func (m *MinioClient) DownloadFile(ctx context.Context, objectName, filePath string) error {
	err := m.client.FGetObject(ctx, m.cfg.Bucket, objectName, filePath, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	log.Info().Msgf("Downloaded %s/%s to %s", m.cfg.Bucket, objectName, filePath)
	return nil
}

// ListObjects lists objects in a prefix
func (m *MinioClient) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	var objects []string

	ch := m.client.ListObjects(ctx, m.cfg.Bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for obj := range ch {
		if obj.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", obj.Err)
		}
		objects = append(objects, obj.Key)
	}

	return objects, nil
}

// DeleteObjects deletes objects from MinIO
func (m *MinioClient) DeleteObjects(ctx context.Context, objectNames []string) error {
	for _, objectName := range objectNames {
		err := m.client.RemoveObject(ctx, m.cfg.Bucket, objectName, minio.RemoveObjectOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete object %s: %w", objectName, err)
		}
		log.Info().Msgf("Deleted %s/%s", m.cfg.Bucket, objectName)
	}
	return nil
}

// GetObjectsOlderThan returns objects older than the specified duration
func (m *MinioClient) GetObjectsOlderThan(ctx context.Context, prefix, duration string) ([]string, error) {
	objects, err := m.ListObjects(ctx, prefix)
	if err != nil {
		return nil, err
	}

	// Parse duration
	d, err := time.ParseDuration(duration)
	if err != nil {
		return nil, fmt.Errorf("failed to parse duration: %w", err)
	}

	cutoffTime := time.Now().Add(-d)
	var oldObjects []string

	for _, objectName := range objects {
		objInfo, err := m.client.StatObject(ctx, m.cfg.Bucket, objectName, minio.StatObjectOptions{})
		if err != nil {
			continue
		}
		if objInfo.LastModified.Before(cutoffTime) {
			oldObjects = append(oldObjects, objectName)
		}
	}

	return oldObjects, nil
}
