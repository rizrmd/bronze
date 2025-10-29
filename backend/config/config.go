package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server     ServerConfig     `json:"server"`
	MinIO      MinIOConfig      `json:"minio"`
	Processing ProcessingConfig `json:"processing"`
}

type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type MinIOConfig struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	UseSSL    bool   `json:"use_ssl"`
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
}

type ProcessingConfig struct {
	MaxWorkers    int                 `json:"max_workers"`
	QueueSize     int                 `json:"queue_size"`
	Decompression DecompressionConfig `json:"decompression"`
	WatchInterval time.Duration       `json:"watch_interval"`
	TempDir       string              `json:"temp_dir"`
}

type DecompressionConfig struct {
	Enabled            bool   `json:"enabled"`
	MaxExtractSize     string `json:"max_extract_size"`
	MaxFilesPerArchive int    `json:"max_files_per_archive"`
	NestedArchiveDepth int    `json:"nested_archive_depth"`
	PasswordProtected  bool   `json:"password_protected"`
	ExtractToSubfolder bool   `json:"extract_to_subfolder"`
}

func Load() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnvInt("SERVER_PORT", 8080),
		},
		MinIO: MinIOConfig{
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
			SecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
			UseSSL:    getEnvBool("MINIO_USE_SSL", false),
			Bucket:    getEnv("MINIO_BUCKET", "files"),
			Region:    getEnv("MINIO_REGION", "us-east-1"),
		},
		Processing: ProcessingConfig{
			MaxWorkers:    getEnvInt("MAX_WORKERS", 3),
			QueueSize:     getEnvInt("QUEUE_SIZE", 100),
			WatchInterval: getEnvDuration("WATCH_INTERVAL", 5*time.Second),
			TempDir:       getEnv("TEMP_DIR", "/tmp/bronze"),
			Decompression: DecompressionConfig{
				Enabled:            getEnvBool("DECOMPRESSION_ENABLED", true),
				MaxExtractSize:     getEnv("MAX_EXTRACT_SIZE", "1GB"),
				MaxFilesPerArchive: getEnvInt("MAX_FILES_PER_ARCHIVE", 1000),
				NestedArchiveDepth: getEnvInt("NESTED_ARCHIVE_DEPTH", 3),
				PasswordProtected:  getEnvBool("PASSWORD_PROTECTED", true),
				ExtractToSubfolder: getEnvBool("EXTRACT_TO_SUBFOLDER", true),
			},
		},
	}

	if err := os.MkdirAll(config.Processing.TempDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	return config, nil
}

func (c *Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
