![LiteDock](docs/img/logo.svg)

[🇺🇸 English](README.md)

轻量级 Docker 容器管理系统

## 概述

LiteDock 是一个基于 Go 和 Clean Architecture 的 Docker 容器管理平台。提供简洁的 REST API 用于管理 Docker 容器、镜像、网络和卷。

## 技术栈

- **Web 框架**: [Fiber](https://github.com/gofiber/fiber) v2
- **数据库**: PostgreSQL + pgx v5
- **查询构建器**: [Squirrel](https://github.com/Masterminds/squirrel)
- **数据库迁移**: [golang-migrate](https://github.com/golang-migrate/migrate)
- **日志**: [zerolog](https://github.com/rs/zerolog)
- **监控**: [Prometheus](https://github.com/ansrivas/fiberprometheus)
- **数据验证**: [go-playground/validator](https://github.com/go-playground/validator)
- **测试**: [testify](https://github.com/stretchr/testify)

## 功能

- 容器管理（创建、启动、停止、删除、查看日志）
- 镜像管理（列表、拉取、删除）
- 网络管理（创建、删除、连接/断开容器）
- 卷管理（创建、删除、挂载）
- 实时容器日志流
- 容器统计信息监控
- Swagger API 文档

## 快速开始

### 前置条件

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 14+

### 本地开发

```sh
# 启动 PostgreSQL
make compose-up

# 运行应用（带数据库迁移）
make run
```

### 完整 Docker 部署

```sh
make compose-up-all
```

## 服务地址

- REST API: http://127.0.0.1:8080
- Swagger 文档: http://127.0.0.1:8080/swagger
- Prometheus 监控: http://127.0.0.1:8080/metrics
- 健康检查: http://127.0.0.1:8080/healthz
- PostgreSQL: `postgres://user:myAwEsOm3pa55@w0rd@127.0.0.1:5432/db`

## 项目结构

```
LiteDock
├── cmd/app/main.go          # 应用入口点
├── config/                   # 环境配置
├── internal/
│   ├── app/app.go           # 主应用逻辑 & 依赖注入
│   ├── controller/           # HTTP 处理器
│   │   └── restapi/         # REST API (Fiber)
│   ├── entity/               # 业务模型
│   ├── usecase/             # 业务逻辑
│   └── migrate.go           # 数据库迁移
├── migrations/              # SQL 迁移文件
├── pkg/                      # 公共包
├── docs/                     # Swagger 文档
├── web/                      # 前端（可选）
└── docker-compose.yml        # Docker 编排
```

## 架构设计

LiteDock 遵循 **Clean Architecture** 原则（Robert Martin）：

```
┌─────────────────────────────────────────────────┐
│  Controller (REST API)  - HTTP 处理器            │
├─────────────────────────────────────────────────┤
│  UseCase (业务逻辑) - 领域服务                   │
├─────────────────────────────────────────────────┤
│  Repository (数据访问) - 数据库操作              │
└─────────────────────────────────────────────────┘
```

**核心原则**：依赖只能向内流转。内层不知道外层的存在。

## API 接口

### 容器
- `GET  /api/v1/containers` - 列出所有容器
- `GET  /api/v1/containers/:id` - 获取容器详情
- `POST /api/v1/containers` - 创建容器
- `POST /api/v1/containers/:id/start` - 启动容器
- `POST /api/v1/containers/:id/stop` - 停止容器
- `DELETE /api/v1/containers/:id` - 删除容器
- `GET  /api/v1/containers/:id/logs` - 获取容器日志
- `GET  /api/v1/containers/:id/stats` - 获取容器统计

### 镜像
- `GET  /api/v1/images` - 列出所有镜像
- `POST /api/v1/images/pull` - 拉取镜像
- `DELETE /api/v1/images/:id` - 删除镜像

### 网络
- `GET  /api/v1/networks` - 列出网络
- `POST /api/v1/networks` - 创建网络
- `DELETE /api/v1/networks/:id` - 删除网络

### 卷
- `GET  /api/v1/volumes` - 列出卷
- `POST /api/v1/volumes` - 创建卷
- `DELETE /api/v1/volumes/:id` - 删除卷

## Makefile 命令

```sh
make compose-up          # 启动 Docker Compose（仅数据库）
make compose-up-all      # 启动完整栈
make compose-down        # 停止所有容器
make run                 # 运行应用
make swag-v1            # 生成 Swagger 文档
make test               # 运行测试
make format             # 格式化代码
make linter-golangci    # 运行 linter
make pre-commit         # 预提交检查
```

## 配置

通过环境变量配置（12-factor app）：

| 变量 | 描述 | 默认值 |
|------|------|--------|
| `APP_HOST` | API 主机 | `0.0.0.0` |
| `APP_PORT` | API 端口 | `8080` |
| `PG_HOST` | PostgreSQL 主机 | `127.0.0.1` |
| `PG_PORT` | PostgreSQL 端口 | `5432` |
| `PG_USER` | PostgreSQL 用户 | `user` |
| `PG_PASSWORD` | PostgreSQL 密码 | `myAwEsOm3pa55` |
| `PG_DB` | 数据库名 | `db` |
| `DOCKER_HOST` | Docker socket 路径 | `/var/run/docker.sock` |

## License

MIT
