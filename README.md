<div align="center">

<p align="center">
  <img src="docs/img/logo.svg" width="75%" alt="LiteDock Logo">
</p>

# LiteDock

**Lightweight Docker Container Management Platform**

<p align="center">
  <a href="README_CN.md">简体中文</a> |
  <strong>English</strong>
</p>

<p align="center">
  <a href="https://raw.githubusercontent.com/lminimum/LiteDock/main/LICENSE">
    <img src="https://img.shields.io/github/license/lminimum/LiteDock?color=brightgreen" alt="license">
  </a><!--
  --><a href="https://github.com/lminimum/LiteDock/releases/latest">
    <img src="https://img.shields.io/github/v/release/lminimum/LiteDock?color=brightgreen&include_prereleases" alt="release">
  </a><!--
  --><a href="https://hub.docker.com/r/lminimum/litedock">
    <img src="https://img.shields.io/badge/docker-dockerHub-blue" alt="docker">
  </a><!--
  --><a href="https://goreportcard.com/report/github.com/lminimum/LiteDock">
    <img src="https://goreportcard.com/badge/github.com/lminimum/LiteDock" alt="GoReportCard">
  </a>
</p>

<p align="center">
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-key-features">Key Features</a> •
  <a href="#-deployment">Deployment</a> •
  <a href="#-tech-stack">Tech Stack</a> •
  <a href="#-project-structure">Project Structure</a>
</p>

</div>

##  Project Description

LiteDock is a lightweight Docker container management platform with a visual interface for personal users and developers to quickly manage container services.

**Core Features:**
- **Container Management**: Start, stop, restart, view logs, execute commands
- **Image Management**: View, pull, delete images
- **Network & Volume Management**: Visual Docker network and volume operations
- **Remote Machine Support**: Manage Docker containers on remote servers via SSH
- **Lightweight Deployment**: Quick start via Docker Compose

---

## ✨ Key Features

### 🎨 Core Functions

| Feature | Description |
|---------|-------------|
|  **Remote SSH Management** | Connect to remote servers via SSH to manage Docker containers |
|  **Real-time Monitoring** | View container status, logs, and resource usage |
|  **Container Operations** | Start, stop, restart, remove containers |
|  **Multi-machine Support** | Manage multiple Docker hosts from a single interface |
|  **Web-based UI** | Modern Vue 3 interface, accessible from any browser |
|  **REST API** | Full API support for automation and integration |

### 🚀 Advanced Features

**Remote Machine Integration:**
- SSH key/password authentication
- Automatic connection testing
- Real-time connection status

**Container Management:**
- Container list with status filtering
- Real-time log viewing
- Command execution in containers
- Batch operations

---

## 🚀 Quick Start

### Using Docker Compose (Recommended)

```bash
# Clone the project
git clone https://github.com/lminimum/LiteDock.git
cd LiteDock

# Start the service
docker-compose up -d
```

### Manual Setup

**Prerequisites:**
- Docker >= 24
- Docker Compose >= 2.18
- Go >= 1.25
- Node.js >= 20

```bash
# Start database and dependencies
docker-compose -f docker-compose.yml up -d

# Run backend
go run ./cmd/app

# Run frontend (in another terminal)
cd web
npm install
npm run dev
```

🎉 After startup, visit `http://localhost:5173` to start using!

---

##  Deployment

### Docker Deployment

```bash
# Pull image
docker pull lminimum/litedock:latest

# Run with SQLite
docker run --name litedock -d --restart always \
  -p 8080:8080 \
  -p 5173:5173 \
  -v ./data:/data \
  lminimum/litedock:latest
```

### Environment Configuration

| Variable | Description | Default |
|---------|-------------|---------|
| `APP_NAME` | Application name | LiteDock |
| `APP_VERSION` | Application version | 1.0.0 |
| `HTTP_PORT` | HTTP server port | 8080 |
| `LOG_LEVEL` | Log level | debug |
| `DB_TYPE` | Database type (sqlite/mysql/postgres) | sqlite |
| `DB_URL` | Database connection string | - |
| `DB_POOL_MAX` | Database connection pool size | 2 |
| `METRICS_ENABLED` | Enable metrics collection | true |
| `SWAGGER_ENABLED` | Enable Swagger documentation | false |
| `CACHE_CONTAINER_TTL` | Container cache TTL | 30s |

### Database Configuration

**SQLite (Default):**
```bash
DB_TYPE=sqlite
DB_URL=./data.db
```

**MySQL:**
```bash
DB_TYPE=mysql
DB_URL=mysql://user:password@tcp(localhost:3306)/litedock
```

**PostgreSQL:**
```bash
DB_TYPE=postgres
DB_URL=postgres://user:password@localhost:5432/litedock
```

---

##  Tech Stack

| Component | Technology |
|-----------|------------|
| **Backend** | Go 1.25+, Clean Architecture, Fiber |
| **Frontend** | Vue 3 + Vite + TypeScript |
| **Database** | PostgreSQL, MySQL, SQLite (via abstraction layer) |
| **API** | REST (Fiber), Swagger documentation |
| **Container** | Docker API via SSH |

### Architecture

LiteDock follows **Clean Architecture** principles:

```
  Frontend / API
        |
   Controller Layer
        |
    UseCase Layer
        |
  Repository / External Services
```

**Key Principles:**
- **Dependency Inversion**: Inner layer doesn't depend on outer layer
- **Decoupling**: Business logic independent and testable
- **Clear Layering**: controller → usecase → repository → entity

---

## Project Structure

```
LiteDock/
├── cmd/app/              # Application entry point
├── config/               # Configuration (environment variables)
├── internal/
│   ├── app/             # Core application logic
│   ├── controller/      # REST controllers (Fiber handlers)
│   ├── entity/          # Business entities
│   ├── repo/            # Data persistence layer
│   └── usecase/         # Business use cases
├── pkg/                  # Utility packages (httpserver, logger, database, sshclient, dockerclient)
├── web/                  # Vue3 + Vite frontend
├── docs/                 # Documentation, Swagger
├── migrations/           # Database migration files
└── Makefile             # Build commands
```

### Layer Overview

| Layer | Directory | Responsibility |
|-------|----------|----------------|
| Controller | `internal/controller/` | HTTP request handling |
| UseCase | `internal/usecase/` | Business logic encapsulation |
| Repository | `internal/repo/` | Data persistence |
| Entity | `internal/entity/` | Business objects |

---

## 🔧 Common Commands

```bash
# Development
make run                 # Start application (with deps + swagger + migrations)
make deps               # Tidy and verify dependencies
make swag-v1           # Generate Swagger documentation
make format            # Format code (gofumpt + gci)
make test              # Run unit tests
make pre-commit        # Full check (deps → swag → format → test)

# Docker
make compose-up        # Start core services (database)
make compose-up-all    # Start full stack
make compose-down      # Stop all containers

# Database migrations
make migrate-create name=xxx  # Create new migration
make migrate-up              # Apply migrations
```

---

##  Related Links

| Resource | Link |
|----------|------|
| Docker Documentation | [docker.com](https://docs.docker.com/) |
| Go Website | [golang.org](https://golang.org/) |
| Vue3 Website | [vuejs.org](https://vuejs.org/) |
| Vite Website | [vitejs.dev](https://vitejs.dev/) |

---

##  License

This project is licensed under the [MIT License](./LICENSE).

---

## 🌟 Star History

<div align="center">

[![Star History Chart](https://api.star-history.com/svg?repos=lminimum/LiteDock&type=Date)](https://star-history.com/#lminimum/LiteDock&Date)

</div>

---

<div align="center">

### 💖 Thank you for using LiteDock

If this project is helpful to you, welcome to give us a ⭐️ Star！

**[Official Documentation](./docs/)** • **[Issue Feedback](https://github.com/lminimum/LiteDock/issues)** • **[Latest Release](https://github.com/lminimum/LiteDock/releases)**

<sub>Built with ❤️ by LiteDock Team</sub>

</div>
