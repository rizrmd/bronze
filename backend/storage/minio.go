package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"time"

	"bronze-backend/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOClient struct {
	client       *minio.Client
	config       *config.MinIOConfig
	bucketName   string
	bucketExists bool
	bucketError  string
}

func NewMinIOClient(cfg *config.MinIOConfig) (*MinIOClient, error) {
	// Extract host from endpoint URL
	endpoint := cfg.Endpoint
	if strings.HasPrefix(endpoint, "http://") {
		endpoint = strings.TrimPrefix(endpoint, "http://")
	} else if strings.HasPrefix(endpoint, "https://") {
		endpoint = strings.TrimPrefix(endpoint, "https://")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL(),
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	minioClient := &MinIOClient{
		client:       client,
		config:       cfg,
		bucketName:   cfg.Bucket,
		bucketExists: false, // Will be checked lazily
		bucketError:  "Bucket status not yet checked",
	}

	// Check bucket existence asynchronously to avoid blocking startup
	go func() {
		bucketExists, err := minioClient.checkBucketExists()
		if err != nil {
			log.Printf("Warning: Failed to check bucket existence: %v", err)
			minioClient.bucketExists = false
			minioClient.bucketError = fmt.Sprintf("Cannot access bucket '%s': %v", cfg.Bucket, err)
		} else if !bucketExists {
			log.Printf("Warning: Bucket '%s' does not exist", cfg.Bucket)
			minioClient.bucketExists = false
			minioClient.bucketError = fmt.Sprintf("Bucket '%s' does not exist", cfg.Bucket)
		} else {
			minioClient.bucketExists = true
			minioClient.bucketError = ""
			log.Printf("Bucket '%s' is accessible", cfg.Bucket)
		}
	}()

	return minioClient, nil
}

func (m *MinIOClient) checkBucketExists() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return m.client.BucketExists(ctx, m.bucketName)
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
	// Check if bucket is accessible first, refresh status if needed
	if !m.bucketExists {
		// Try to check bucket status again in case async check hasn't completed yet
		exists, err := m.checkBucketExists()
		if err != nil {
			m.bucketExists = false
			m.bucketError = fmt.Sprintf("Cannot access bucket '%s': %v", m.bucketName, err)
		} else if !exists {
			m.bucketExists = false
			m.bucketError = fmt.Sprintf("Bucket '%s' does not exist", m.bucketName)
		} else {
			m.bucketExists = true
			m.bucketError = ""
		}

		// If still not accessible, return error
		if !m.bucketExists {
			return minio.UploadInfo{}, fmt.Errorf("bucket '%s' is not accessible: %s", m.bucketName, m.bucketError)
		}
	}

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

func (m *MinIOClient) ListFiles(ctx context.Context, prefix string, limit int) ([]minio.ObjectInfo, error) {
	// Check if bucket is accessible first, refresh status if needed
	log.Printf("ListFiles: bucketExists=%v, bucketError=%s", m.bucketExists, m.bucketError)
	if !m.bucketExists {
		// Try to check bucket status again in case async check hasn't completed yet
		exists, err := m.checkBucketExists()
		if err != nil {
			m.bucketExists = false
			m.bucketError = fmt.Sprintf("Cannot access bucket '%s': %v", m.bucketName, err)
		} else if !exists {
			m.bucketExists = false
			m.bucketError = fmt.Sprintf("Bucket '%s' does not exist", m.bucketName)
		} else {
			m.bucketExists = true
			m.bucketError = ""
		}

		log.Printf("ListFiles: after recheck bucketExists=%v, bucketError=%s", m.bucketExists, m.bucketError)

		// If still not accessible, return error
		if !m.bucketExists {
			return nil, fmt.Errorf("bucket '%s' is not accessible: %s", m.bucketName, m.bucketError)
		}
	}

	var files []minio.ObjectInfo
	seenDirs := make(map[string]bool)

	objectsCh := m.client.ListObjects(ctx, m.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: false, // Don't recurse to get directory structure
	})

	count := 0
	for object := range objectsCh {
		if object.Err != nil {
			return nil, object.Err
		}

		// Check if we've reached the limit
		if limit > 0 && count >= limit {
			break
		}

		// Add the object itself
		files = append(files, object)
		count++

		// Check if this object represents a directory path
		// Extract directory prefixes from the object key
		if prefix != "" {
			relativePath := strings.TrimPrefix(object.Key, prefix)
			if strings.HasPrefix(relativePath, "/") {
				relativePath = relativePath[1:]
			}

			// Find directory components in the relative path
			parts := strings.Split(relativePath, "/")
			for i := 0; i < len(parts)-1; i++ {
				dirPath := strings.Join(parts[:i+1], "/")
				if dirPath != "" && !seenDirs[dirPath] {
					seenDirs[dirPath] = true
					// Create a directory object
					dirKey := prefix
					if dirKey != "" && !strings.HasSuffix(dirKey, "/") {
						dirKey += "/"
					}
					dirKey += dirPath + "/"

					dirObject := minio.ObjectInfo{
						Key:          dirKey,
						Size:         0,
						LastModified: time.Now(),
						ETag:         "",
						ContentType:  "application/x-directory",
					}
					files = append(files, dirObject)
				}
			}
		} else {
			// Handle root level directories
			parts := strings.Split(object.Key, "/")
			if len(parts) > 1 {
				dirPath := parts[0]
				// Only create synthetic directory if this object is not already a directory
				if dirPath != "" && !seenDirs[dirPath] && !strings.HasSuffix(object.Key, "/") {
					seenDirs[dirPath] = true
					dirObject := minio.ObjectInfo{
						Key:          dirPath + "/",
						Size:         0,
						LastModified: time.Now(),
						ETag:         "",
						ContentType:  "application/x-directory",
					}
					files = append(files, dirObject)
				}
			}
		}
	}

	// Remove duplicates from the final result
	uniqueFiles := make([]minio.ObjectInfo, 0, len(files))
	seenKeys := make(map[string]bool)
	for _, file := range files {
		if !seenKeys[file.Key] {
			seenKeys[file.Key] = true
			uniqueFiles = append(uniqueFiles, file)
		} else {
			log.Printf("Removing duplicate key: %s", file.Key)
		}
	}

	log.Printf("ListFiles: before deduplication %d files, after deduplication %d files", len(files), len(uniqueFiles))

	return uniqueFiles, nil
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

// Get direct MinIO client for advanced operations
func (m *MinIOClient) GetClient() *minio.Client {
	return m.client
}

// Get bucket name for advanced operations  
func (m *MinIOClient) GetBucketName() string {
	return m.bucketName
}

func (m *MinIOClient) GetBucketInfo(ctx context.Context) (minio.BucketInfo, error) {
	return minio.BucketInfo{}, nil
}

func (m *MinIOClient) SetBucket(bucketName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exists, err := m.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("bucket %s does not exist", bucketName)
	}

	m.bucketName = bucketName
	// Update bucket status to reflect the new bucket
	m.bucketExists = exists
	m.bucketError = ""
	log.Printf("Bucket changed to '%s' and status updated", bucketName)
	return nil
}

func (m *MinIOClient) GetConfig() *config.MinIOConfig {
	return m.config
}

func (m *MinIOClient) GetBucketStatus() (bool, string) {
	return m.bucketExists, m.bucketError
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
