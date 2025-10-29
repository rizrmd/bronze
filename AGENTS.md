# Bronze Development Guidelines

## Build Commands

### Backend (Go)
- Build: `go build -o bronze-backend main.go`
- Run: `go run main.go`
- Test: `go test ./...`
- Test single package: `go test ./handlers`
- Lint: `go fmt ./... && go vet ./...`

### Frontend (Vue 3 + TypeScript)
- Dev: `bun run dev`
- Build: `bun run build`
- Preview: `bun run preview`
- Type check: `vue-tsc -b`
- Add components: `bunx shadcn-vue@latest add <component>`

## Code Style

### Go Backend
- Use standard Go formatting (`go fmt`)
- Package imports: stdlib, third-party, local (grouped)
- Error handling: always check and return errors
- Naming: PascalCase for exported, camelCase for unexported
- Struct tags for JSON responses

### Frontend
- Vue 3 Composition API with `<script setup>`
- TypeScript strict mode
- Imports: Vue libraries, external libs, local components
- Use `@/` alias for src imports
- shadcn-vue components with Tailwind v4 for styling
- Component naming: PascalCase
- Props with TypeScript interfaces
- Use `cn()` utility for conditional classes
- Always use `bun` or `bunx` instead of `npm` or `node`

## Project Structure
- Backend: modular by feature (handlers, processor, minio, config)
- Frontend: components organized by domain, shared UI components in `components/ui/`
- Use absolute imports with path aliases

## Important Notes
- **NEVER use npm** - Always use `bun` or `bunx` for frontend package management
- **NEVER use node** - Always use `bun` for running JavaScript/TypeScript
- Root package.json uses npm only for orchestrating backend and frontend with concurrently
- Frontend development should always use `bun run dev`, `bun run build`, etc.