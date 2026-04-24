<div align="center">

<p align="center">
  <img src="docs/img/logo.svg" width="75%" alt="LiteDock Logo">
</p>

# LiteDock

**轻量级 Docker 容器管理平台**

<p align="center">
  <a href="README.md">简体中文</a> |
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
  <a href="#-快速开始">快速开始</a> •
  <a href="#-核心功能">核心功能</a> •
  <a href="#-部署">部署</a> •
  <a href="#-技术栈">技术栈</a> •
  <a href="#-项目结构">项目结构</a>
</p>

</div>

##  项目描述

LiteDock 是一个轻量级 Docker 容器管理平台，提供可视化界面，支持个人用户和开发者快速管理容器服务。

**核心功能：**
- **容器管理**：启动、停止、重启、查看日志、执行命令
- **镜像管理**：查看、拉取、删除镜像
- **网络和卷管理**：Docker 网络和卷的可视化操作
- **远程机器支持**：通过 SSH 连接管理远程服务器的 Docker 容器
- **轻量部署**：通过 Docker Compose 快速启动

---

## ✨ 核心功能

### 🎨 主要功能

| 功能 | 描述 |
|---------|-------------|
|  **远程 SSH 管理** | 通过 SSH 连接到远程服务器管理 Docker 容器 |
|  **实时监控** | 查看容器状态、日志和资源使用情况 |
|  **容器操作** | 启动、停止、重启、删除容器 |
|  **多机器支持** | 从单一界面管理多个 Docker 主机 |
|  **Web 界面** | 现代 Vue 3 界面，可从任何浏览器访问 |
|  **REST API** | 支持自动化和集成的完整 API |

### 🚀 高级功能

**远程机器集成：**
- SSH 密钥/密码认证
- 自动连接测试
- 实时连接状态

**容器管理：**
- 支持状态过滤的容器列表
- 实时日志查看
- 容器内命令执行
- 批量操作

---

## 🚀 快速开始

### 使用 Docker Compose（推荐）

```bash
# 克隆项目
git clone https://github.com/lminimum/LiteDock.git
cd LiteDock

# 启动服务
docker-compose up -d
```

### 手动配置

**环境依赖：**
- Docker >= 24
- Docker Compose >= 2.18
- Go >= 1.25
- Node.js >= 20

```bash
# 启动数据库和依赖
docker-compose -f docker-compose.yml up -d

# 运行后端
go run ./cmd/app

# 运行前端（另一个终端）
cd web
npm install
npm run dev
```

🎉 启动后访问 `http://localhost:5173` 即可开始使用！

---

##  部署

### Docker 部署

```bash
# 拉取镜像
docker pull lminimum/litedock:latest

# 使用 SQLite 运行
docker run --name litedock -d --restart always \
  -p 8080:8080 \
  -p 5173:5173 \
  -v ./data:/data \
  lminimum/litedock:latest
```

### 环境变量配置

| 变量 | 描述 | 默认值 |
|---------|-------------|---------|
| `APP_NAME` | 应用名称 | LiteDock |
| `APP_VERSION` | 应用版本 | 1.0.0 |
| `HTTP_PORT` | HTTP 服务端口 | 8080 |
| `LOG_LEVEL` | 日志级别 | debug |
| `DB_TYPE` | 数据库类型 (sqlite/mysql/postgres) | sqlite |
| `DB_URL` | 数据库连接字符串 | - |
| `DB_POOL_MAX` | 数据库连接池大小 | 2 |
| `METRICS_ENABLED` | 启用指标收集 | true |
| `SWAGGER_ENABLED` | 启用 Swagger 文档 | false |
| `CACHE_CONTAINER_TTL` | 容器缓存 TTL | 30s |

### 数据库配置

**SQLite（默认）：**
```bash
DB_TYPE=sqlite
DB_URL=./data.db
```

**MySQL：**
```bash
DB_TYPE=mysql
DB_URL=mysql://user:password@tcp(localhost:3306)/litedock
```

**PostgreSQL：**
```bash
DB_TYPE=postgres
DB_URL=postgres://user:password@localhost:5432/litedock
```

---

##  技术栈

| 组件 | 技术 |
|-----------|------------|
| **后端** | Go 1.25+，整洁架构，Fiber |
| **前端** | Vue 3 + Vite + TypeScript |
| **数据库** | PostgreSQL、MySQL、SQLite（通过抽象层） |
| **API** | REST (Fiber)，Swagger 文档 |
| **容器** | 通过 SSH 的 Docker API |

### 架构

LiteDock 遵循**整洁架构**原则：

```
  前端 / API
        |
    控制器层
        |
      用例层
        |
   仓储层 / 外部服务
```

**核心原则：**
- **依赖倒置**：内层不依赖外层实现
- **解耦**：业务逻辑独立，易于测试
- **分层清晰**：controller → usecase → repository → entity

---

## 项目结构

```
LiteDock/
├── cmd/app/              # 应用入口
├── config/               # 配置（环境变量）
├── internal/
│   ├── app/            # 核心应用逻辑
│   ├── controller/      # REST 控制器（Fiber handlers）
│   ├── entity/         # 业务实体
│   ├── repo/           # 数据持久化层
│   └── usecase/        # 业务用例
├── pkg/                  # 工具包（httpserver, logger, database, sshclient, dockerclient）
├── web/                  # Vue3 + Vite 前端
├── docs/                 # 文档、Swagger
├── migrations/           # 数据库迁移文件
└── Makefile             # 构建命令
```

### 分层说明

| 分层 | 目录 | 职责 |
|-------|----------|----------------|
| 控制器 | `internal/controller/` | HTTP 请求处理 |
| 用例 | `internal/usecase/` | 业务逻辑封装 |
| 仓储 | `internal/repo/` | 数据持久化 |
| 实体 | `internal/entity/` | 业务对象 |

---

## 🔧 常用命令

```bash
# 开发
make run                 # 启动应用（含依赖 + swagger + 迁移）
make deps              # 整理和验证依赖
make swag-v1          # 生成 Swagger 文档
make format           # 格式化代码（gofumpt + gci）
make test             # 运行单元测试
make pre-commit       # 完整检查（deps → swag → format → test）

# Docker
make compose-up        # 启动核心服务（数据库）
make compose-up-all   # 启动完整服务
make compose-down     # 停止所有容器

# 数据库迁移
make migrate-create name=xxx  # 创建迁移文件
make migrate-up              # 应用迁移
```

---

##  相关链接

| 资源 | 链接 |
|----------|------|
| Docker 文档 | [docker.com](https://docs.docker.com/) |
| Go 官网 | [golang.org](https://golang.org/) |
| Vue3 官网 | [vuejs.org](https://vuejs.org/) |
| Vite 官网 | [vitejs.dev](https://vitejs.dev/) |

---

##  开源许可

本项目基于 [MIT License](./LICENSE) 发布。

---

## 🌟 Star History

<div align="center">

[![Star History Chart](https://api.star-history.com/svg?repos=lminimum/LiteDock&type=Date)](https://star-history.com/#lminimum/LiteDock&Date)

</div>

---

<div align="center">

### 💖 感谢使用 LiteDock

如果该项目对您有帮助，欢迎给我们一个 ⭐️ Star！

**[官方文档](./docs/)** • **[问题反馈](https://github.com/lminimum/LiteDock/issues)** • **[最新发布](https://github.com/lminimum/LiteDock/releases)**

<sub>Built with ❤️ by LiteDock Team</sub>

</div>
