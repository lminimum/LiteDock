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

**IMPORTANT - Commit Review Rule**:
After completing any implementation or change, **DO NOT commit immediately**. Stage the changes with `git add`, but wait for either:
- The user to review and approve the changes, OR
- A code review to be completed

Only commit after receiving explicit approval. This ensures changes are reviewed before becoming part of the project history.

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

### Mock Generation
- Use `make mock` to generate mocks via mockgen (gomock)
- Generated mocks are placed in `internal/usecase/` with `mocks_*_test.go` naming
- **Existing hand-written mocks** (e.g., in test files like `volume_test.go`) **must NOT be regenerated**
- Use mockgen **only for NEW interfaces** that don't yet have hand-written mocks

#### When to Use mockgen vs Hand-written Mocks
| Scenario | Approach |
|----------|----------|
| New interface, simple or few methods | Hand-written mock (easier to control) |
| New interface, complex or many methods | mockgen |
| Existing hand-written mock | Keep as-is, do not regenerate |
| Generated mock needed for new interface | `make mock` or `mockgen -source=...` |

#### mockgen Command Examples
```bash
# Generate from source interface
mockgen -source=./pkg/dockerclient/client.go -destination=internal/usecase/mocks_dockerclient_test.go -package=usecase_test

# Generate from contracts file (as in Makefile)
mockgen -source=./internal/repo/contracts.go -package=usecase_test > ./internal/usecase/mocks_repo_test.go
mockgen -source=./internal/usecase/contracts.go -package=usecase_test > ./internal/usecase/mocks_usecase_test.go
```

#### Hand-written Mock Pattern
Hand-written mocks in test files use function pointers for configurable behavior:
```go
type mockVolumeRepo struct {
    listByMachineFn func(ctx context.Context, machineID string) ([]entity.Volume, error)
    // ...
}

func (m *mockVolumeRepo) ListByMachine(ctx context.Context, machineID string) ([]entity.Volume, error) {
    return m.listByMachineFn(ctx, machineID)
}

var _ repo.VolumeRepo = (*mockVolumeRepo)(nil) // compile-time interface check
```

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

#### CSS / Style Reuse (CRITICAL)
**复用优先，不要重复定义样式。**

1. **使用全局 Design Tokens**: 所有颜色、间距、字体、圆角、阴影、过渡动画都已在 `web/src/style.css` 中定义为 CSS 变量。优先使用：
   - `var(--color-*)` 系列（颜色）
   - `var(--space-*)` 系列（间距）
   - `var(--font-*)` 系列（字体）
   - `var(--radius-*)` 系列（圆角）
   - `var(--shadow-*)` 系列（阴影）
   - `var(--transition-*)` 系列（过渡）

2. **使用全局 Utility Classes**: `style.css` 中已定义大量全局类和组件样式，优先使用：
   - `.card` - 卡片样式
   - `.btn`, `.btn-primary`, `.btn-secondary`, `.btn-ghost`, `.btn-danger`, `.btn-sm`, `.btn-lg` - 按钮
   - `.input` - 输入框
   - `.badge`, `.badge-success`, `.badge-warning`, `.badge-error`, `.badge-info` - 徽章
   - `.flex`, `.gap-*`, `.items-center`, `.justify-between` 等 Flex/Grid 工具类
   - `.text-xs` ~ `.text-2xl`, `.font-medium`, `.font-semibold` 等文本工具类
   - `.p-*`, `.px-*`, `.py-*`, `.m-*`, `.mt-*`, `.mb-*` 等间距工具类
   - `.rounded-sm`, `.rounded-md`, `.rounded-lg`, `.rounded-full` 等圆角工具类

3. **新增样式前先检查**: 在 `<style scoped>` 中定义新样式前，先确认 `style.css` 中是否已有相同或类似的样式定义。

4. **禁止重复**: 不得在组件 `<style scoped>` 中重复定义已在 `style.css` 全局定义的相同 CSS 规则（如再次定义 `.btn`, `.input`, `.card` 等基础组件类）。

5. **响应式覆盖**: 如果全局类的响应式行为需要微调，可以只在组件 scoped style 中覆盖需要改动的属性（如 `.card` 的 padding），而不是重新定义整个卡片样式。

6. **Dashboard.vue 例外**: `Dashboard.vue` 有独特的视觉样式（如 `stat-card` 使用不同背景色），允许保留组件级样式定义，但必须使用标准 CSS 变量名（如 `--color-background-weak` 而非 `--color-bg-secondary`）。

**示例**:
```vue
<!-- ✅ 正确：使用全局 .btn 和 .btn-primary 类 -->
<button class="btn btn-primary">Click</button>

<!-- ❌ 错误：在 scoped style 中重新定义 .btn -->
<style scoped>
.btn { display: inline-flex; padding: 8px 16px; ... } /* 不要这样做 */
</style>
```

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
