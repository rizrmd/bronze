# CRUSH.md - Bronze Development Guide

This guide helps AI agents work effectively in the Bronze codebase - a Go backend with Vue.js frontend for file processing and job management.

## Quick Overview

Bronze is a full-stack application with:
- **Backend**: Go API with MinIO integration, job processing, and file management
- **Frontend**: Vue 3 + TypeScript with modern UI components
- **Architecture**: Modular by feature with separate handlers for different concerns

## Essential Commands

### Root Level (Orchestration)
```bash
# Development (both backend and frontend) - PREFERRED METHOD
bun run dev.ts                 # Starts both with dependency checking and auto-restart
npm run dev                    # Basic orchestration (no dependency checking)
npm run build                  # Builds both for production  
npm run start                  # Runs production build
npm run test                   # Tests both backend and frontend
npm run lint                   # Lints both codebases
npm run clean                  # Removes build artifacts
```

### Backend (Go)
```bash
cd backend
go run main.go                 # Start development server
go build -o bronze-backend main.go  # Build binary
go test ./...                  # Run all tests
go test ./jobs                 # Test specific package
go fmt ./... && go vet ./...   # Format and lint
```

### Frontend (Vue + TypeScript)
```bash
cd frontend
bun run dev                    # Development server (NEVER use npm run dev)
bun run build                  # Build for production
bun run preview                 # Preview production build
vue-tsc -b                     # Type checking only
bunx shadcn-vue@latest add <component>  # Add UI components
```

**CRITICAL**: Always use `bun` or `bunx` for frontend - never `npm` or `node`.

## Project Structure

```
bronze/
├── backend/                    # Go backend API
│   ├── config/                # Configuration management with defaults
│   ├── data_browser/          # Excel/CSV/MDB data processing
│   ├── files/                 # File handling and processing
│   ├── jobs/                  # Job queue and worker pool
│   ├── monitoring/            # File watching service
│   ├── routes/                # HTTP routing and middleware
│   ├── storage/               # MinIO and Nessie clients
│   ├── main.go               # Application entry point
│   └── openapi.json          # API specification
├── frontend/                  # Vue.js SPA
│   ├── src/
│   │   ├── api/              # Axios API client with types
│   │   ├── components/       # Vue components organized by domain
│   │   │   ├── ui/           # shadcn-vue reusable components
│   │   │   ├── files/        # File management components
│   │   │   └── nessie/       # Data export components
│   │   ├── composables/      # Vue composition functions
│   │   ├── types/            # TypeScript type definitions
│   │   └── views/            # Page-level components
│   └── components.json       # shadcn-vue configuration
└── AGENTS.md                 # Existing development guidelines
```

## Code Patterns & Conventions

### Go Backend
- **Package structure**: Modular by feature domain
- **Error handling**: Always check and return errors with context
- **Naming**: PascalCase for exported, camelCase for unexported
- **Configuration**: Environment variables with sensible defaults in `config/config.go`
- **HTTP handlers**: Separate handler packages with dependency injection
- **JSON responses**: Struct tags for serialization

### Vue Frontend  
- **Composition API**: Always use `<script setup>` syntax
- **TypeScript**: Strict mode enabled with comprehensive types
- **Components**: PascalCase naming, props with TypeScript interfaces
- **Imports**: Vue libraries → external → local (use `@/` alias)
- **Styling**: Tailwind v4 with shadcn-vue components and `cn()` utility
- **State**: Use composables for shared logic, API client for data

### API Design
- **RESTful routes**: Consistent patterns with `/api/` prefix
- **Error responses**: Standardized JSON format with success/error status
- **CORS**: Enabled for all origins in development
- **OpenAPI**: Complete specification available at `/openapi.json`

## Key Features & Architecture

### Job Processing System
- **Priority queue**: Jobs ordered by priority (high/medium/low) then creation time
- **Worker pool**: Configurable concurrent processing with graceful shutdown
- **Job lifecycle**: Created → Queued → Processing → Completed/Failed/Cancelled
- **Real-time tracking**: Progress updates and status monitoring

### File Management
- **Object storage**: MinIO integration with bucket management
- **Archive support**: ZIP, TAR, TAR.GZ decompression without artificial limits
- **File watching**: Real-time change monitoring (currently disabled for stability)
- **Data browser**: Excel (XLSX/XLS/XLSM), CSV, MDB file processing

### Data Processing
- **Universal CSV**: Auto-detection of delimiters and headers
- **Excel support**: Multi-sheet browsing with configurable row limits
- **Streaming**: Large file support with chunked processing
- **Nessie integration**: Data export capabilities

## Configuration Management

### Environment Variables
Backend uses environment variables with defaults in `config/config.go`:
- No `.env.example` → `.env` automation (manual setup required)
- Default server: `localhost:8060`
- Default MinIO: `localhost:9000` with `minioadmin` credentials
- Processing: 3 workers, 100 queue size, unlimited decompression

### Frontend Configuration
- API URL via `VITE_API_URL` (empty for development proxy)
- Development proxy routes API calls to backend

## Critical Gotchas

1. **Use bun run dev.ts for development** - This is the preferred development method with dependency checking and auto-restart
2. **NEVER use npm/node for frontend** - Always use `bun` or `bunx`
3. **Decompression has no limits** - Any size limits are considered bugs
4. **File watcher is disabled** - Causes startup issues, don't re-enable without testing
5. **Environment setup** - Must configure manually, no automatic .env creation
6. **Port conflicts** - Backend: 8060, Frontend: 8070 (development)
7. **Import paths** - Frontend uses `@/` alias, backend uses relative imports
8. **Don't start dev server manually** - Let the dev.ts script manage both services

## Testing Approach

### Backend
- Unit tests in `*_test.go` files alongside source
- Integration tests for handlers and storage
- Use `go test ./...` for full suite
- Comprehensive data browser tests in `data_browser/data_browser_test.go`

### Frontend  
- Type checking with `vue-tsc -b` (primary validation)
- Unit tests can be added as needed
- Build validation via `bun run build`
- Test data browser functionality with backend test suite

## Development Workflow

1. **Start development**: `bun run dev.ts` from root (preferred) or `npm run dev`
2. **Backend changes**: Modify Go code, backend auto-restarts if using dev.ts
3. **Frontend changes**: Hot reload enabled, use type checking for validation
4. **API changes**: Update OpenAPI spec and frontend types together
5. **Testing**: Run tests before commits, lint both codebases
6. **Build verification**: Ensure production build works before deployment

### Development Script Benefits
The `dev.ts` script provides:
- Dependency checking (Go, Bun, Go modules, frontend packages)
- Automatic port cleanup (kills processes on 8060/8070)
- Backend auto-restart on crashes
- Colored output for both services
- Graceful shutdown handling

## Common Tasks

### Adding New API Endpoints
1. Add handler in appropriate backend package
2. Register route in `routes/routes.go`
3. Update OpenAPI specification if needed
4. Add API client methods in `frontend/src/api/index.ts`
5. Define TypeScript types in `frontend/src/types/index.ts`

### Adding New UI Components  
1. Use shadcn-vue: `bunx shadcn-vue@latest add <component>`
2. Place domain components in appropriate feature folder
3. Follow existing patterns for props and composables
4. Use Tailwind classes with `cn()` utility for conditional styling

### Configuration Changes
1. Update defaults in `backend/config/config.go`
2. Document new environment variables
3. Update API info endpoint if user-configurable
4. Test with and without `.env` file

## Storage & External Services

### MinIO (Object Storage)
- Primary file storage backend
- Supports multiple buckets with dynamic switching
- Presigned URLs for direct file access
- Archive extraction capabilities

### Nessie (Data Export)
- Optional data warehouse integration
- Batch export with configurable sizes
- Separate client from MinIO storage

## Security Considerations

- CORS enabled for all origins (development setup)
- No authentication in current implementation
- Environment variables contain sensitive data
- File upload size limits configurable
- Archive bomb protection via decompression settings

This documentation supplements the existing `AGENTS.md` file with comprehensive architecture, patterns, and operational knowledge for effective development in the Bronze codebase.