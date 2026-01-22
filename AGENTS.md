# AGENTS.md

This file contains guidelines and commands for agentic coding agents working in the LiteDock repository.

## Project Overview

LiteDock is a lightweight Docker container management platform with a Go backend + Vue3 frontend architecture. The project follows Clean Architecture principles with clear separation between business logic, service layers, and interfaces.

### Architecture
- **Backend**: Go 1.25+ with Clean Architecture
- **Frontend**: Vue 3 + Vite + TypeScript  
- **Database**: PostgreSQL with migrations
- **Message Queues**: RabbitMQ, NATS
- **APIs**: REST (Fiber), gRPC, RPC over message queues

## Build Commands

### Go Backend
```bash
# Full development setup
make run

# Individual commands
make deps              # Tidy and verify dependencies
make swag-v1          # Generate Swagger docs
make proto-v1         # Generate protobuf files
make format           # Format code with gofumpt and gci
make linter-golangci  # Run golangci-lint
make test             # Run unit tests
make integration-test # Run integration tests
make mock             # Generate mocks with mockgen
make pre-commit       # Run all pre-commit checks
```

### Running Single Tests
```bash
# Run specific test file
go test -v ./internal/usecase/translation_test.go

# Run specific test function
go test -v ./internal/usecase/... -run TestHistory

# Run tests with race detection
go test -race -v ./internal/...

# Run tests with coverage
go test -cover -v ./internal/...
```

### Frontend (Vue3)
```bash
cd web
npm install         # Install dependencies
npm run dev         # Start development server
npm run build       # Build for production
npm run preview     # Preview production build
```

### Docker Services
```bash
# Start core services (db, rabbitmq, nats)
make compose-up

# Start all services including backend
make compose-up-all

# Stop all services
make compose-down
```

## Code Style Guidelines

### Go Code Style

#### Import Organization
- Use `gci` for import sorting (configured in Makefile)
- Import groups: standard, default, local
- Generated code is skipped during formatting

#### Naming Conventions
- **Packages**: lowercase, short, descriptive (e.g., `usecase`, `entity`, `repo`)
- **Interfaces**: descriptive names ending with the type (e.g., `TranslationRepo`, `TranslationWebAPI`)
- **Structs**: PascalCase, descriptive (e.g., `UseCase`, `Translation`)
- **Methods**: PascalCase for exported, lowercase for unexported
- **Variables**: camelCase, descriptive names
- **Constants**: UPPER_SNAKE_CASE for exported, camelCase for unexported

#### Error Handling
- Always wrap errors with context using `fmt.Errorf("...: %w", err)`
- Error messages should include the method path for debugging
- Use structured error types when appropriate
- Handle errors immediately, don't ignore them

#### Code Structure
- Follow Clean Architecture layers: `controller` → `usecase` → `repo` → `entity`
- Dependency injection through constructors
- Interface-based design for testability
- Context as first parameter for methods that do I/O

#### Documentation
- Exported functions and types must have godoc comments
- Use proper godoc format with period at the end
- Include examples in godoc when helpful
- Swagger annotations for REST endpoints

### Testing Guidelines

#### Test Structure
- Use table-driven tests for multiple scenarios
- Test files should end with `_test.go`
- Use `require` for assertions, `assert` for non-fatal checks
- Create test helpers with `t.Helper()`

#### Mock Usage
- Generate mocks using `make mock`
- Use `gomock` for mock control
- Mock interfaces, not concrete types
- Setup mocks in test helper functions

#### Test Naming
- Test functions: `Test[FunctionName]`
- Sub-tests: descriptive names using `t.Run()`
- Use `t.Parallel()` when safe (avoid data races)

### Frontend Code Style

#### Vue3 + TypeScript
- Use Composition API with `<script setup>`
- TypeScript for all components
- Follow Vue 3 style guide
- Use kebab-case for component names in templates

#### File Organization
- Components in `src/components/`
- Use `.vue` single file components
- Separate concerns: template, script, style

## Linting and Quality

### Go Linters (golangci.yml)
- Enabled linters: `wsl_v5`, `errcheck`, `gosec`, `staticcheck`, etc.
- Custom settings for complexity, line limits, etc.
- Formatters: `gci`, `gofumpt`, `goimports`

### Quality Gates
- All code must pass `make pre-commit` before commit
- Test coverage should be maintained
- No new linting violations allowed
- Dependencies must be security-checked (`make deps-audit`)

## Development Workflow

### Making Changes
1. Create feature branch from main
2. Make changes following code style guidelines
3. Add/update tests for new functionality
4. Run `make pre-commit` to verify quality
5. Test manually with `make run`
6. Submit pull request

### Adding New Features
- Follow existing architectural patterns
- Add appropriate interfaces and implementations
- Include comprehensive tests
- Update documentation and Swagger specs
- Consider both REST and RPC endpoints

### Database Changes
- Create migration files: `make migrate-create name`
- Test migrations: `make migrate-up`
- Update entity structs accordingly
- Add repository layer tests

## Environment Setup

### Required Tools
- Go 1.25+
- Node.js 20+
- Docker & Docker Compose
- Make

### Development Environment
- Copy `.env.example` to `.env` and configure
- Use `make compose-up` to start dependencies
- Backend runs on configured port (default 8080)
- Frontend runs on port 5173

## Key Files and Directories

```
cmd/app/           # Application entry point
internal/
  app/            # Application setup and configuration
  controller/     # HTTP/RPC controllers
  usecase/        # Business logic layer
  repo/           # Data access layer
  entity/         # Business entities
pkg/              # Reusable packages
web/              # Vue3 frontend
docs/             # Documentation and Swagger
migrations/       # Database migrations
config/           # Configuration management
```

## Common Patterns

### Dependency Injection
```go
func NewTranslationUseCase(repo TranslationRepo, webAPI TranslationWebAPI) *UseCase {
    return &UseCase{repo: repo, webAPI: webAPI}
}
```

### Error Wrapping
```go
return nil, fmt.Errorf("UseCase - Method - repo.Call: %w", err)
```

### Context Usage
```go
func (uc *UseCase) Method(ctx context.Context, req Request) (Response, error) {
    // Implementation
}
```

### Testing Pattern
```go
func TestFunction(t *testing.T) {
    t.Parallel()
    
    tests := []struct{
        name string
        mock func()
        want interface{}
        err  error
    }{
        // Test cases
    }
    
    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```