# AGENTS.md

This file contains guidelines for agentic coding agents and developers working in the LiteDock repository.

## Project Overview

LiteDock is a lightweight Docker container management platform with a Go backend + Vue3 frontend architecture. The project follows Clean Architecture principles.

### Tech Stack
- **Backend**: Go 1.25+ with Clean Architecture
- **Frontend**: Vue 3 + Vite + TypeScript
- **Database**: PostgreSQL, MySQL, SQLite (via abstraction layer)
- **Message Queues**: RabbitMQ, NATS
- **APIs**: REST (Fiber), gRPC, RPC over message queues

---

## Git Workflow

### Branch Naming Convention

```
main                    # Stable release branch (default branch)
dev                     # Development branch (base for features)
feature/<description>   # New features (e.g., feature/user-auth)
fix/<description>      # Bug fixes (e.g., fix/container-crash)
refactor/<description>  # Code refactoring
docs/<description>      # Documentation updates
```

**Rules**:
- Use kebab-case: `feature/user-authentication`
- Be descriptive: `feature/add-container-logs` not `feature/new`
- Prefix with type: `feature/`, `fix/`, `refactor/`, `chore/`, `docs/`

### Commit Message Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types**:
| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `style` | Formatting, missing semicolons, etc. |
| `refactor` | Code change that neither fixes nor adds |
| `perf` | Performance improvement |
| `test` | Adding or updating tests |
| `chore` | Maintenance tasks (deps, build, CI) |

**Scope** (optional but recommended):
- `backend`, `frontend`, `api`, `db`, `docker`, `ci`

**Examples**:
```bash
feat(backend): add user authentication via JWT
fix(docker): handle container restart timeout
docs(api): update API endpoint documentation
refactor(db): extract query builder into separate package
chore(deps): upgrade go-fiber to v2.53.0
```

**Rules**:
- Use imperative mood: "add" not "added" or "adds"
- Keep subject line under 72 characters
- Reference issues: `fix: resolve null pointer (#123)`
- **Never add co-author** in commit messages (no `Co-authored-by:`)

### Push & PR Workflow

```
1. Sync with latest
   git checkout main
   git pull origin main

2. Create feature branch
   git checkout -b feature/my-feature

3. Make changes & commit (follow commit convention)
   git add .
   git commit -m "feat(api): add new endpoint"

4. Rebase onto latest before PR (keep history clean)
   git fetch origin main
   git rebase origin/main

5. Push and create PR
   git push -u origin feature/my-feature
   # Create PR on GitHub

6. After PR merged, delete branch
   git branch -d feature/my-feature
   git push origin --delete feature/my-feature
```

### PR Guidelines

- **Title**: Follow commit convention (e.g., `feat(backend): add container stats API`)
- **Description**: Explain what and why, link to issue
- **Size**: Keep PRs focused, < 500 lines changed is ideal
- **Squash**: Rebase and squash multiple commits before merge if needed
- **CI**: Ensure all checks pass before requesting review

---

## Build Commands

### Go Backend
```bash
make run                 # Full development (deps + swag + proto + migrations)
make deps               # Tidy and verify dependencies
make swag-v1           # Generate Swagger docs
make proto-v1          # Generate protobuf files
make format            # Format code (gofumpt + gci)
make linter-golangci   # Run golangci-lint
make test              # Run unit tests
make integration-test  # Run integration tests
make mock              # Generate mocks with mockgen
make pre-commit        # Run all checks (swag → proto → mock → format → lint → test)
```

### Running Single Tests
```bash
go test -v ./internal/usecase/translation_test.go
go test -v ./internal/usecase/... -run TestHistory
go test -race -v ./internal/...
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
make compose-up        # Start core services (db, rabbitmq, nats)
make compose-up-all    # Start full stack
make compose-down      # Stop all containers
```

---

## Code Style Guidelines

### Go Code Style

#### Import Organization
- Use `gci` for import sorting (configured in Makefile)
- Import groups: standard, default, local
- Generated code is skipped during formatting

#### Naming Conventions
- **Packages**: lowercase, short, descriptive
- **Interfaces**: descriptive names ending with type (`TranslationRepo`)
- **Structs**: PascalCase (`UseCase`, `Translation`)
- **Methods**: PascalCase for exported, lowercase for unexported
- **Variables**: camelCase, descriptive
- **Constants**: UPPER_SNAKE_CASE for exported, camelCase for unexported

#### Error Handling
- Wrap errors: `fmt.Errorf("UseCase - Method - repo.Call: %w", err)`
- Error messages include method path
- Handle errors immediately

#### Code Structure
- Clean Architecture: `controller` → `usecase` → `repo` → `entity`
- Dependency injection through constructors
- Interface-based design for testability
- Context as first parameter for I/O methods

#### Documentation
- Exported functions/types need godoc comments
- Swagger annotations for REST endpoints

### Testing Guidelines
- Table-driven tests for multiple scenarios
- Test files: `*_test.go`
- Use `require` for assertions
- Test helpers: `t.Helper()`
- Mocks via `make mock` (gomock)

### Frontend Code Style
- Vue3 Composition API with `<script setup>`
- TypeScript everywhere
- kebab-case for component names
- Components in `src/components/`

---

## Linting and Quality

### Go Linters (golangci.yml)
- Enabled: `wsl_v5`, `errcheck`, `gosec`, `staticcheck`, etc.
- Complexity limits: cyclomatic=10, cognitive=15
- nolint directives require specific explanation

### Quality Gates
- All code must pass `make pre-commit` before commit
- Test coverage should be maintained
- Dependencies security-checked (`make deps-audit`)

---

## Environment Setup

### Required Tools
- Go 1.25+
- Node.js 20+
- Docker & Docker Compose
- Make

### Development Environment
```bash
cp .env.example .env
make compose-up        # Start dependencies
make run               # Start backend
```

Services:
- REST API: http://127.0.0.1:8080
- Swagger: http://127.0.0.1:8080/swagger
- Metrics: http://127.0.0.1:8080/metrics
- Health: http://127.0.0.1:8080/healthz
- Frontend: http://localhost:5173

---

## Key Files and Directories

```
cmd/app/main.go           # Application entry point
config/                   # Environment config
internal/
  app/app.go             # DI & server startup
  app/migrate.go         # DB migrations (build tag: migrate)
  controller/
    restapi/             # Fiber HTTP handlers
    grpc/                # gRPC handlers
    amqp_rpc/            # RabbitMQ RPC handlers
    nats_rpc/            # NATS RPC handlers
  entity/                # Business models
  usecase/               # Business logic
  repo/
    persistent/          # Database implementation
    webapi/              # External API (Google Translate)
pkg/
  httpserver/            # Fiber server wrapper
  grpcserver/            # gRPC server wrapper
  postgres/              # pgx connection
  rabbitmq/              # RabbitMQ RPC
  nats/                  # NATS RPC
  logger/                # zerolog wrapper
  database/              # DB abstraction (factory pattern)
migrations/              # SQL migration files
docs/
  restapi/               # Swagger generated docs
  proto/v1/              # gRPC proto + generated Go
web/                     # Vue3 + Vite frontend
```

---

## Database Migrations

```bash
# Create migration
make migrate-create name=create_users

# Run migrations
make migrate-up

# Migration files in migrations/*.sql
```

Supported: PostgreSQL, MySQL, SQLite

---

## CI Pipeline

GitHub Actions runs on PRs:
1. Commit lint (go-commitlinter)
2. golangci-lint (15m timeout)
3. yamllint, hadolint, dotenv-linter
4. Dependency check (nancy)
5. Unit tests + codecov
6. Integration tests (docker compose)

---

## Common Patterns

### Dependency Injection
```go
func NewUseCase(repo TranslationRepo, webAPI TranslationWebAPI) *UseCase {
    return &UseCase{repo: repo, webAPI: webAPI}
}
```

### Error Wrapping
```go
return nil, fmt.Errorf("UseCase - Method - repo.Call: %w", err)
```

### Context Usage
```go
func (uc *UseCase) Method(ctx context.Context, req Request) (Response, error)
```
