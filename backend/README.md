# Bronze Backend

A comprehensive Go backend with MinIO integration, featuring file processing, parallel job management, and archive decompression capabilities.

## Features

- **MinIO Integration**: Full object storage operations (upload, download, list, delete)
- **Priority Job Queue**: Configurable job priorities (high, medium, low)
- **Parallel Processing**: Configurable worker pool (default: 3 concurrent workers)
- **Archive Decompression**: Support for ZIP, TAR, TAR.GZ formats
- **File Watching**: Monitor file changes and trigger processing
- **RESTful API**: Complete HTTP API for all operations
- **Real-time Tracking**: Job status and progress monitoring
- **Configurable**: Environment-based configuration system

## Architecture

```
bronze/
└── backend/
    ├── main.go                 # Server entry point
    ├── config/
    │   └── config.go          # Configuration management
    ├── minio/
    │   └── client.go          # MinIO client wrapper
    ├── processor/
    │   ├── job.go             # Job definitions and types
    │   ├── queue.go           # Priority job queue
    │   ├── worker.go          # Worker pool management
    │   ├── decompressor.go    # Archive extraction
    │   └── pipeline.go        # File processing pipeline
    ├── handlers/
    │   ├── file.go            # File operation handlers
    │   └── jobs.go            # Job management handlers
    ├── routes/
    │   └── routes.go          # HTTP routing
    └── README.md
```

## Quick Start

### Prerequisites

- Go 1.19+
- MinIO server (local or remote)
- Make (for build scripts)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd bronze/backend
```

2. Install dependencies:
```bash
go mod download
```

3. Configure environment variables (see Configuration section)

4. Run the server:
```bash
go run main.go
```

### Build

```bash
go build -o bronze-backend main.go
```

## Configuration

The application uses environment variables for configuration. Create a `.env` file or export the variables:

### Server Configuration
```bash
SERVER_HOST=localhost
SERVER_PORT=8080
```

### MinIO Configuration
```bash
MINIO_ENDPOINT=http://localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=files
MINIO_REGION=us-east-1
```

### Processing Configuration
```bash
MAX_WORKERS=3
QUEUE_SIZE=100
WATCH_INTERVAL=5s
TEMP_DIR=/tmp/bronze
```

### Decompression Configuration
```bash
DECOMPRESSION_ENABLED=true
MAX_EXTRACT_SIZE=1GB
MAX_FILES_PER_ARCHIVE=1000
NESTED_ARCHIVE_DEPTH=3
PASSWORD_PROTECTED=true
EXTRACT_TO_SUBFOLDER=true
```

## API Endpoints

### Health Check
- `GET /` - Health check
- `GET /health` - Health check
- `GET /api` - API documentation

### File Operations
- `POST /files` - Upload file
- `GET /files` - List files (query: `?prefix=<path>`)
- `GET /files/{filename}` - Download file
- `GET /files/{filename}` - Get file info
- `DELETE /files/{filename}` - Delete file
- `GET /files/{filename}/presigned` - Generate presigned URL (query: `?expiry=<duration>`)

### Job Management
- `POST /jobs` - Create processing job
- `GET /jobs` - List jobs (query: `?status=<status>`)
- `GET /jobs/{id}` - Get job details
- `DELETE /jobs/{id}` - Cancel job
- `PUT /jobs/{id}/priority` - Update job priority
- `GET /jobs/stats` - Get queue and worker statistics
- `PUT /jobs/workers` - Update worker count
- `GET /jobs/workers/active` - Get active jobs

## Usage Examples

### Upload a File
```bash
curl -X POST -F "file=@example.zip" http://localhost:8080/files
```

### Create a Processing Job
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{
    "type": "decompress",
    "file_path": "/tmp/example.zip",
    "bucket": "files",
    "object_name": "example.zip",
    "priority": "high"
  }' \
  http://localhost:8080/jobs
```

### List Jobs
```bash
curl http://localhost:8080/jobs
```

### Get Job Status
```bash
curl http://localhost:8080/jobs/{job-id}
```

### Update Worker Count
```bash
curl -X PUT -H "Content-Type: application/json" \
  -d '{"count": 5}' \
  http://localhost:8080/jobs/workers
```

## Supported Archive Formats

- **ZIP** - Standard ZIP archives
- **TAR** - Unix tar archives
- **TAR.GZ** - Gzip compressed tar archives

## Job Processing Pipeline

1. **File Detection**: Identify file type and if it's an archive
2. **Download**: Fetch file from MinIO to temporary storage
3. **Decompression**: Extract if archive (maintains directory structure)
4. **Processing**: Process extracted files individually
5. **Cleanup**: Remove temporary files
6. **Results**: Store processing results and metadata

## Worker Pool Configuration

- **Default Workers**: 3 concurrent workers
- **Configurable**: Update via API or environment variable
- **Priority Handling**: High priority jobs processed first
- **Graceful Shutdown**: Workers complete current jobs before stopping

## Error Handling

The API provides comprehensive error responses:
```json
{
  "success": false,
  "message": "Error description",
  "error": "Detailed error message"
}
```

## Monitoring

### Health Check
```bash
curl http://localhost:8080/health
```

### Statistics
```bash
curl http://localhost:8080/jobs/stats
```

## Development

### Project Structure
- `config/` - Configuration management
- `minio/` - MinIO client wrapper
- `processor/` - Job processing logic
- `handlers/` - HTTP request handlers
- `routes/` - HTTP routing configuration

### Running Tests
```bash
go test ./...
```

### Linting
```bash
go fmt ./...
go vet ./...
```

## Deployment

### Docker
```dockerfile
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o bronze-backend main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bronze-backend .
EXPOSE 8080
CMD ["./bronze-backend"]
```

### Docker Compose
```yaml
version: '3.8'
services:
  bronze-backend:
    build: .
    ports:
      - "8080:8080"
    environment:
      - MINIO_ENDPOINT=minio:9000
      - MINIO_ACCESS_KEY=minioadmin
      - MINIO_SECRET_KEY=minioadmin
    depends_on:
      - minio

  minio:
    image: minio/minio:latest
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      - MINIO_ACCESS_KEY=minioadmin
      - MINIO_SECRET_KEY=minioadmin
    command: server /data --console-address ":9001"
```

## License

[Add your license information here]

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Support

For issues and questions, please open an issue on the GitHub repository.