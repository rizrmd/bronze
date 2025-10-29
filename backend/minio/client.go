package minio

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"time"

	"bronze-backend/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOClient struct {
	client     *minio.Client
	config     *config.MinIOConfig
	bucketName string
}

func NewMinIOClient(cfg *config.MinIOConfig) (*MinIOClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	minioClient := &MinIOClient{
		client:     client,
		config:     cfg,
		bucketName: cfg.Bucket,
	}

	if err := minioClient.ensureBucket(); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return minioClient, nil
}

func (m *MinIOClient) ensureBucket() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exists, err := m.client.BucketExists(ctx, m.bucketName)
	if err != nil {
		return err
	}

	if !exists {
		err = m.client.MakeBucket(ctx, m.bucketName, minio.MakeBucketOptions{
			Region: m.config.Region,
		})
		if err != nil {
			return fmt.Errorf("failed to create bucket %s: %w", m.bucketName, err)
		}
		log.Printf("Created bucket: %s", m.bucketName)
	}

	return nil
}

func (m *MinIOClient) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (minio.UploadInfo, error) {
	return m.client.PutObject(ctx, m.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
}

func (m *MinIOClient) DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	return m.client.GetObject(ctx, m.bucketName, objectName, minio.GetObjectOptions{})
}

func (m *MinIOClient) GetFileInfo(ctx context.Context, objectName string) (minio.ObjectInfo, error) {
	return m.client.StatObject(ctx, m.bucketName, objectName, minio.StatObjectOptions{})
}

func (m *MinIOClient) ListFiles(ctx context.Context, prefix string) ([]minio.ObjectInfo, error) {
	var files []minio.ObjectInfo

	objectsCh := m.client.ListObjects(ctx, m.bucketName, minio.ListObjectsOptions{
		Prefix: prefix,
	})

	for object := range objectsCh {
		if object.Err != nil {
			return nil, object.Err
		}
		files = append(files, object)
	}

	return files, nil
}

func (m *MinIOClient) DeleteFile(ctx context.Context, objectName string) error {
	return m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
}

func (m *MinIOClient) DeleteFiles(ctx context.Context, objectNames []string) error {
	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)
		for _, objectName := range objectNames {
			objectsCh <- minio.ObjectInfo{Key: objectName}
		}
	}()

	errorCh := m.client.RemoveObjects(ctx, m.bucketName, objectsCh, minio.RemoveObjectsOptions{})

	for err := range errorCh {
		if err.Err != nil {
			return fmt.Errorf("failed to delete object %s: %w", err.ObjectName, err.Err)
		}
	}

	return nil
}

func (m *MinIOClient) CopyFile(ctx context.Context, srcObjectName, destObjectName string) (minio.UploadInfo, error) {
	srcOpts := minio.CopySrcOptions{
		Bucket: m.bucketName,
		Object: srcObjectName,
	}

	destOpts := minio.CopyDestOptions{
		Bucket: m.bucketName,
		Object: destObjectName,
	}

	return m.client.CopyObject(ctx, destOpts, srcOpts)
}

func (m *MinIOClient) GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucketName, objectName, expiry, reqParams)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

func (m *MinIOClient) GetPresignedUploadURL(ctx context.Context, objectName string, expiry time.Duration) (string, map[string]string, error) {
	presignedURL, err := m.client.PresignedPutObject(ctx, m.bucketName, objectName, expiry)
	if err != nil {
		return "", nil, err
	}
	return presignedURL.String(), make(map[string]string), nil
}

func (m *MinIOClient) FileExists(ctx context.Context, objectName string) (bool, error) {
	_, err := m.client.StatObject(ctx, m.bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (m *MinIOClient) GetBucketInfo(ctx context.Context) (minio.BucketInfo, error) {
	return minio.BucketInfo{}, nil
}

func (m *MinIOClient) GetClient() *minio.Client {
	return m.client
}

func (m *MinIOClient) GetBucketName() string {
	return m.bucketName
}

func (m *MinIOClient) GetConfig() *config.MinIOConfig {
	return m.config
}

type FileInfoResponse struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	ETag         string    `json:"etag"`
	ContentType  string    `json:"content_type"`
}

type FileInfoDetail struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	ETag         string    `json:"etag"`
	ContentType  string    `json:"content_type"`
}
