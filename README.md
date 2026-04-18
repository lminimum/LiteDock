<p align="center">
  <img src="docs/img/logo.svg" width="75%" alt="LiteDock Logo">
</p>

[🇨🇳 中文](README_CN.md)

# LiteDock

Lightweight Docker container management platform with a visual interface and AI-assisted features for personal users and developers to quickly manage container services.

[![Release](https://img.shields.io/github/v/release/lminimum/LiteDock.svg)](https://github.com/lminimum/LiteDock/releases/)
[![License](https://img.shields.io/badge/License-MIT-success)](https://github.com/lminimum/LiteDock/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/lminimum/LiteDock)](https://goreportcard.com/report/github.com/lminimum/LiteDock)
[![Vue](https://img.shields.io/badge/Vue-3.3-blue)](https://vuejs.org/)
[![Vite](https://img.shields.io/badge/Vite-4.6-blue)](https://vitejs.dev/)
[![Docker](https://img.shields.io/badge/Docker-Container%20Management-blue)](https://www.docker.com/)

---

## Overview

LiteDock is a Docker management platform based on **Go backend + Vue3 frontend**, with core features including:

- **Container Management**: Start, stop, restart containers
- **Image Management**: View, pull, delete images
- **Network & Volume Management**: Visual Docker network and volume operations
- **AI Assistance**: Simple command suggestions, container optimization advice
- **Lightweight Deployment**: Quick start the entire system via Docker Compose

LiteDock follows **Clean Architecture** principles, decoupling business logic, service layers, and interfaces to ensure maintainability and scalability.

---

## Tech Stack

- **Backend**: Go 1.25+, Clean Architecture
- **Frontend**: Vue 3 + Vite + TypeScript
- **Database**: PostgreSQL, MySQL, SQLite (via abstraction layer)
- **API**: REST (Fiber)

---

## Quick Start

### Prerequisites

- Docker >= 24
- Docker Compose >= 2.18
- Go >= 1.25
- Node.js >= 20

### Start Backend

```bash
# Start database and other dependencies
docker-compose -f docker-compose.yml up -d

# Run Go service
go run ./cmd/app
```

### Start Frontend

```bash
cd web
npm install
npm run dev
```

Frontend runs at `http://localhost:5173`

### Full Docker Deployment

```bash
docker-compose -f docker-compose-full.yml up -d
```

Access:
- Web UI: `http://localhost:5173`
- API Docs: `http://localhost:8080/swagger`

---

## Project Structure

```
LiteDock/
├── cmd/app/              # Go backend entry point
├── config/               # Configuration (environment variables)
├── internal/
│   ├── app/              # Core application logic
│   ├── controller/       # REST controllers
│   │   └── restapi/      # Fiber HTTP handlers
│   ├── entity/           # Business entities
│   ├── repo/             # Data persistence layer
│   └── usecase/          # Business use cases
├── pkg/                  # Utility packages (httpserver, logger, etc.)
├── web/                  # Vue3 + Vite frontend
├── docs/                 # Docs, Swagger, images
├── migrations/           # Database migration files
├── docker-compose.yml    # Development environment Compose file
└── Makefile             # Common command wrappers
```

### Layer Overview

- **Internal Layer (`internal`)**: Contains business logic and core functionality
- **Controller Layer (`controller`)**: Handles HTTP requests
- **Entity Layer (`entity`)**: Defines business objects and data structures
- **UseCase Layer (`usecase`)**: Encapsulates business processes
- **External Tools (`pkg`)**: HTTP servers, logging, database connections

---

## Common Commands

```bash
# Backend development
make run                 # Full dev environment (deps + swagger + migrations)
make deps               # Tidy and verify dependencies
make swag-v1           # Generate Swagger docs
make format            # Format code (gofumpt + gci)
make linter-golangci   # Run golangci-lint
make test              # Run unit tests
make integration-test  # Run integration tests
make mock              # Generate test mocks
make pre-commit        # Full check (deps → swag → mock → format → lint → test)

# Docker
make compose-up        # Start core services (db)
make compose-up-all    # Start full stack
make compose-down      # Stop all containers

# Database migrations
make migrate-create name=xxx  # Create migration
make migrate-up              # Run migrations
```

---

## Git Workflow

### Branch Naming

```
main                    # Stable release branch
dev                     # Development branch
feature/<description>   # New features (e.g., feature/user-auth)
fix/<description>       # Bug fixes (e.g., fix/container-crash)
refactor/<description> # Code refactoring
docs/<description>      # Documentation updates
```

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

**Examples**:
```bash
feat(backend): add user authentication via JWT
fix(docker): handle container restart timeout
docs(api): update API endpoint documentation
```

### Push & PR Flow

```
1. Sync with latest
   git checkout main
   git pull origin main

2. Create feature branch
   git checkout -b feature/my-feature

3. Make changes & commit
   git add .
   git commit -m "feat(api): add new endpoint"

4. Rebase before PR
   git fetch origin main
   git rebase origin/main

5. Push and create PR
   git push -u origin feature/my-feature

6. After PR merged, delete branch
   git branch -d feature/my-feature
   git push origin --delete feature/my-feature
```

---

## Feature Examples

### Container Management API

- Start container
- Stop container
- Restart container
- View container status

### Image Management

- List local images
- Pull images
- Delete images

### Network & Volume Management

- List networks and volumes
- Create and delete networks and volumes

### Frontend Features

- Dashboard for Docker status
- Real-time container log display
- Batch multi-container operations
- AI suggestions and optimization advice

---

## Dependency Injection

LiteDock backend uses dependency injection to decouple services. Core logic is injected through constructors for easy testing and mocking:

```go
type ContainerUseCase struct {
    repo ContainerRepository
}

func NewContainerUseCase(repo ContainerRepository) *ContainerUseCase {
    return &ContainerUseCase{repo: repo}
}
```

---

## Docker Integration

LiteDock uses Docker API to interact with the local Docker engine, implementing container, image, network, and volume operations through the **Docker Go SDK**.

---

## Clean Architecture Principles

- **Dependency Inversion**: Inner layer business logic doesn't depend on outer layer implementations
- **Decoupling**: Business logic is independent and easy to test
- **Clear Layering**: controller -> usecase -> repository -> entity

```
  Frontend / API
        |
   Controller Layer
        |
     UseCase Layer
        |
  Repository / External Services
```

---

## Related Links

- [Docker Documentation](https://docs.docker.com/)
- [Go Website](https://golang.org/)
- [Vue3 Website](https://vuejs.org/)
- [Vite Website](https://vitejs.dev/)

---

## License

MIT License © 2026 [lminimum](https://github.com/lminimum)
