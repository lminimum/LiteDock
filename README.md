![LiteDock](docs/img/logo.svg)

# LiteDock

[🇨🇳 中文](README_CN.md)

轻量级 Docker 容器管理系统

[![License](https://img.shields.io/badge/License-MIT-success)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/evrone/go-clean-template)](https://goreportcard.com/report/github.com/evrone/go-clean-template)

## Overview

LiteDock 是一个基于 Go 和 Clean Architecture 的 Docker 容器管理平台。提供简洁的 REST API 用于管理 Docker 容器、镜像、网络和卷。

## Tech Stack

- **Web Framework**: [Fiber](https://github.com/gofiber/fiber) v2
- **Database**: PostgreSQL + pgx v5
- **ORM/Query Builder**: [Squirrel](https://github.com/Masterminds/squirrel)
- **Migrations**: [golang-migrate](https://github.com/golang-migrate/migrate)
- **Logging**: [zerolog](https://github.com/rs/zerolog)
- **Metrics**: [Prometheus](https://github.com/ansrivas/fiberprometheus)
- **Validation**: [go-playground/validator](https://github.com/go-playground/validator)
- **Testing**: [testify](https://github.com/stretchr/testify)

## Features

- 容器管理（创建、启动、停止、删除、查看日志）
- 镜像管理（列表、拉取、删除）
- 网络管理（创建、删除、连接/断开容器）
- 卷管理（创建、删除、挂载）
- 实时容器日志流
- 容器统计信息监控
- Swagger API 文档

## Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 14+

### Local Development

```sh
# Start PostgreSQL
make compose-up

# Run app with migrations
make run
```

### Full Docker Stack

```sh
make compose-up-all
```

## Services

- REST API: http://127.0.0.1:8080
- Swagger Docs: http://127.0.0.1:8080/swagger
- Prometheus Metrics: http://127.0.0.1:8080/metrics
- Health Check: http://127.0.0.1:8080/healthz
- PostgreSQL: `postgres://user:myAwEsOm3pa55@w0rd@127.0.0.1:5432/db`

## Project Structure

```
LiteDock
├── cmd/app/main.go          # Application entry point
├── config/                   # Environment configuration
├── internal/
│   ├── app/app.go           # Main application setup & DI
│   ├── controller/           # HTTP handlers
│   │   └── restapi/         # REST API (Fiber)
│   ├── entity/              # Business models
│   ├── usecase/             # Business logic
│   └── migrate.go           # Database migrations
├── migrations/              # SQL migration files
├── pkg/                     # Shared packages
├── docs/                    # Swagger documentation
├── web/                     # Frontend (optional)
└── docker-compose.yml       # Docker orchestration
```

## Architecture

LiteDock follows **Clean Architecture** principles (Robert Martin):

```
┌─────────────────────────────────────────────────┐
│  Controller (REST API)  - HTTP Handlers         │
├─────────────────────────────────────────────────┤
│  UseCase (Business Logic) - Domain Services     │
├─────────────────────────────────────────────────┤
│  Repository (Data Access) - Database Operations │
└─────────────────────────────────────────────────┘
```

**Key Principle**: Dependencies flow inward. Inner layers know nothing about outer layers.

## API Endpoints

### Containers
- `GET  /api/v1/containers` - List all containers
- `GET  /api/v1/containers/:id` - Get container details
- `POST /api/v1/containers` - Create container
- `POST /api/v1/containers/:id/start` - Start container
- `POST /api/v1/containers/:id/stop` - Stop container
- `DELETE /api/v1/containers/:id` - Remove container
- `GET  /api/v1/containers/:id/logs` - Get container logs
- `GET  /api/v1/containers/:id/stats` - Get container stats

### Images
- `GET  /api/v1/images` - List all images
- `POST /api/v1/images/pull` - Pull image
- `DELETE /api/v1/images/:id` - Remove image

### Networks
- `GET  /api/v1/networks` - List networks
- `POST /api/v1/networks` - Create network
- `DELETE /api/v1/networks/:id` - Remove network

### Volumes
- `GET  /api/v1/volumes` - List volumes
- `POST /api/v1/volumes` - Create volume
- `DELETE /api/v1/volumes/:id` - Remove volume

## Makefile Commands

```sh
make compose-up          # Start Docker Compose (DB only)
make compose-up-all      # Start full stack with reverse proxy
make compose-down        # Stop all containers
make run                 # Run app with migrations
make swag-v1            # Generate Swagger docs
make test               # Run unit tests
make format             # Format code
make linter-golangci    # Run linter
make pre-commit         # Run all checks
```

## Configuration

Configuration is via environment variables (12-factor app):

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_HOST` | API host | `0.0.0.0` |
| `APP_PORT` | API port | `8080` |
| `PG_HOST` | PostgreSQL host | `127.0.0.1` |
| `PG_PORT` | PostgreSQL port | `5432` |
| `PG_USER` | PostgreSQL user | `user` |
| `PG_PASSWORD` | PostgreSQL password | `myAwEsOm3pa55` |
| `PG_DB` | Database name | `db` |
| `DOCKER_HOST` | Docker socket path | `/var/run/docker.sock` |

## License

MIT
