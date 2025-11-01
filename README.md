# Bronze

A comprehensive Go backend with MinIO integration, featuring file processing, parallel job management, and a modern Vue.js frontend.

## 🚀 Quick Start

### Prerequisites

- **Go** 1.19 or higher
- **Bun** for frontend package management
- **Node.js** and **npm** for root package management
- **MinIO** (optional, for local development)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd bronze
```

2. Install dependencies:
```bash
npm run install
```

3. Set up environment:
```bash
cp .env.example .env
# Edit .env with your configuration
```

### Development

Start both backend and frontend in development mode:

```bash
# Using npm script (recommended)
npm run dev

# Or using the shell script directly
./scripts/dev.sh

# Install dependencies and start
npm run dev -- --install
```

This will start:
- **Backend**: http://localhost:8060
- **Frontend**: http://localhost:8070
- **API Docs**: http://localhost:8060/api

The frontend is configured to proxy API requests to the backend, so no CORS issues occur.

### Production

Build and run for production:

```bash
# Build both frontend and backend
npm run build

# Start production server
npm run start

# Or manually
cd dist
./bronze-backend
```

## 📁 Project Structure

```
bronze/
├── backend/                 # Go backend application
│   ├── config/            # Configuration management
│   ├── handlers/          # HTTP request handlers
│   ├── minio/            # MinIO client integration
│   ├── processor/         # Job processing and queue
│   ├── routes/            # API routes
│   └── watcher/          # File watching service
├── frontend/              # Vue.js frontend application
│   ├── src/
│   │   ├── api/         # API client
│   │   ├── components/   # Vue components
│   │   ├── composables/  # Vue composables
│   │   ├── layouts/      # Application layouts
│   │   ├── router/       # Vue Router
│   │   ├── types/        # TypeScript types
│   │   └── views/       # Page components
│   └── public/          # Static assets
├── scripts/              # Development and build scripts
└── dist/               # Production build output
```

## 🛠 Available Scripts

### Development
- `npm run dev` - Start both backend and frontend in development mode
- `npm run dev:backend` - Start only backend
- `npm run dev:frontend` - Start only frontend

### Building
- `npm run build` - Build both frontend and backend for production
- `npm run build:frontend` - Build only frontend
- `npm run build:backend` - Build only backend

### Testing
- `npm run test` - Run all tests
- `npm run test:frontend` - Run frontend tests
- `npm run test:backend` - Run backend tests

### Linting
- `npm run lint` - Lint both frontend and backend
- `npm run lint:frontend` - Lint frontend
- `npm run lint:backend` - Lint backend

### Utilities
- `npm run clean` - Clean build artifacts
- `npm run type-check` - Run TypeScript type checking
- `npm run start` - Start production server

## 🔧 Configuration

### Backend Configuration

The backend uses environment variables. See `.env.example` for all available options:

```bash
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8060

# MinIO
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_ENDPOINT=http://localhost:9000
MINIO_BUCKET=bronze-files
MINIO_REGION=us-east-1

# Processing
PROCESSING_MAX_WORKERS=3
PROCESSING_QUEUE_SIZE=100

# File Watcher
WATCHER_POLL_INTERVAL=30s
```

### Frontend Configuration

The frontend uses Vite environment variables:

```bash
# API URL (empty for proxy in development)
VITE_API_URL=http://localhost:8060
```

## 📊 Features

### Backend
- ✅ **MinIO Integration**: Complete object storage support
- ✅ **File Management**: Upload, download, delete files
- ✅ **Job Queue**: Priority-based job processing
- ✅ **Worker Pool**: Configurable concurrent processing
- ✅ **Archive Support**: ZIP, TAR, TAR.GZ decompression
- ✅ **File Watching**: Real-time file change monitoring
- ✅ **REST API**: Complete OpenAPI specification
- ✅ **Health Checks**: System monitoring endpoints

## ⚙️ Job System Architecture

Bronze features a robust job processing system built on a priority queue and worker pool architecture for efficient background task management.

### Core Components

**Job Model**: Each job contains:
- Unique ID, type, and priority level (high/medium/low)
- Status tracking (pending/processing/completed/failed/cancelled)
- File metadata (path, bucket, object name)
- Timestamps (created, started, completed)
- Progress tracking and result storage
- Custom metadata for extensibility

**Priority Queue**: 
- Implemented as a priority heap that orders jobs by priority first, then creation time
- High-priority jobs are processed before medium and low ones
- Thread-safe with concurrent access support
- Configurable queue size to prevent memory overflow

**Worker Pool**:
- Configurable number of workers (default based on CPU cores)
- Workers continuously pull jobs from the queue and process them concurrently
- Dynamic worker count adjustment without restart
- Graceful shutdown with job completion guarantees

### Job Lifecycle

1. **Creation**: Jobs are created via API with file path, bucket, object name, and priority
2. **Queuing**: Jobs enter the priority queue and wait for available workers
3. **Processing**: Workers pick up jobs, mark them as processing, and execute the job processor
4. **Completion**: Jobs finish with success/failure status and results
5. **Tracking**: Real-time progress updates and status changes are maintained

### Job Processing

Jobs support various file processing operations:
- Archive decompression (ZIP, TAR, TAR.GZ)
- File analysis and metadata extraction
- Custom processing pipelines
- Progress tracking with percentage completion

### API Management

The job system provides comprehensive API operations:
- **CRUD Operations**: Create, list, get, cancel jobs
- **Priority Management**: Update job priority (only for pending jobs)
- **Monitoring**: View statistics, active jobs, and system health
- **Worker Management**: Adjust worker count dynamically
- **Progress Tracking**: Real-time job status and progress updates

### Frontend Integration

The JobsManager interface provides:
- Real-time job monitoring with auto-refresh
- Job filtering and search capabilities
- Visual progress indicators
- Priority adjustment controls
- Cancellation functionality
- Statistics dashboard with success rates and processing times

### Frontend
- ✅ **Modern UI**: Vue 3 with Composition API
- ✅ **TypeScript**: Full type safety
- ✅ **Responsive Design**: Mobile-friendly interface
- ✅ **Real-time Updates**: Auto-refreshing dashboard
- ✅ **File Management**: Drag-and-drop uploads
- ✅ **Job Monitoring**: Live job progress tracking
- ✅ **Event Watching**: File change event monitoring
- ✅ **Settings**: Worker pool configuration
- ✅ **Error Handling**: Comprehensive user feedback
- ✅ **Loading States**: Professional loading indicators

## 🔌 API Endpoints

### Health & Info
- `GET /` - Health check
- `GET /health` - Health check
- `GET /api` - API information
- `GET /openapi.json` - OpenAPI specification

### Files
- `POST /files` - Upload file
- `GET /files` - List files
- `GET /files/{filename}` - Download file
- `DELETE /files/{filename}` - Delete file
- `GET /files/{filename}/presigned` - Get presigned URL

### Jobs
- `POST /jobs` - Create job
- `GET /jobs` - List jobs
- `GET /jobs/{id}` - Get job details
- `DELETE /jobs/{id}` - Cancel job
- `PUT /jobs/{id}/priority` - Update job priority
- `GET /jobs/stats` - Get statistics
- `PUT /jobs/workers` - Update worker count
- `GET /jobs/workers/active` - Get active jobs

### Watcher
- `GET /watcher/events/unprocessed` - Get unprocessed events
- `GET /watcher/events/history` - Get event history
- `POST /watcher/events/mark-processed` - Mark event as processed

## 🐳 Docker Support

You can run Bronze with Docker (Dockerfile not included in this setup, but you can create one):

```dockerfile
# Multi-stage build for production
FROM golang:1.19-alpine AS backend-builder
WORKDIR /app/backend
COPY backend/ .
RUN go build -o bronze-backend main.go

FROM oven/bun:1-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/ .
RUN bun install && bun run build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=backend-builder /app/backend/bronze-backend .
COPY --from=frontend-builder /app/frontend/dist ./frontend
EXPOSE 8060
CMD ["./bronze-backend"]
```

## 🧪 Development Scripts

The `scripts/` directory contains comprehensive development utilities:

- `scripts/dev.sh` - Development server launcher
- `scripts/build.sh` - Production build script
- `scripts/start.sh` - Production server launcher
- `scripts/test.sh` - Test runner

Each script supports various options:

```bash
# Development with dependency installation
./scripts/dev.sh --install

# Build with cleanup
./scripts/build.sh --clean

# Test specific components
./scripts/test.sh backend
./scripts/test.sh frontend
./scripts/test.sh lint
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `npm run test`
5. Run linting: `npm run lint`
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Troubleshooting

### Common Issues

1. **Port conflicts**: Ensure ports 8060 and 8070 are available
2. **MinIO connection**: Check MinIO is running and credentials are correct
3. **CORS issues**: In development, the frontend proxies requests, so CORS shouldn't be an issue
4. **Build failures**: Ensure all dependencies are installed with `npm run install`

### Getting Help

- Check the logs in the terminal output
- Verify environment variables in `.env`
- Ensure all prerequisites are installed
- Check network connectivity for MinIO

## 📚 Documentation

- **API Documentation**: http://localhost:8060/api (when running)
- **OpenAPI Spec**: http://localhost:8060/openapi.json
- **Frontend Components**: See `frontend/src/components/` directory
- **Backend Handlers**: See `backend/handlers/` directory