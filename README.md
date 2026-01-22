````markdown
![LiteDock Logo](docs/img/logo.svg)

# LiteDock

轻量级 Docker 容器管理平台，提供可视化界面和 AI 辅助功能，用于个人用户和开发者快速管理容器服务。

[![Release](https://img.shields.io/github/v/release/lminimum/LiteDock.svg)](https://github.com/lminimum/LiteDock/releases/)
[![License](https://img.shields.io/badge/License-MIT-success)](https://github.com/lminimum/LiteDock/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/lminimum/LiteDock)](https://goreportcard.com/report/github.com/lminimum/LiteDock)
[![Vue](https://img.shields.io/badge/Vue-3.3-blue)](https://vuejs.org/)
[![Vite](https://img.shields.io/badge/Vite-4.6-blue)](https://vitejs.dev/)
[![Docker](https://img.shields.io/badge/Docker-Container%20Management-blue)](https://www.docker.com/)

---

## 项目概述

LiteDock 是一个基于 **Go 后端 + Vue3 前端** 的 Docker 管理平台，核心特点包括：

- **容器管理**：启动、停止、重启容器
- **镜像管理**：查看、拉取、删除镜像
- **网络和卷管理**：支持 Docker 网络和卷可视化操作
- **AI 辅助**：提供简单的命令提示、容器优化建议
- **轻量部署**：通过 Docker Compose 快速启动整个系统

LiteDock 遵循 **整洁架构** 原则，将业务逻辑、服务层和界面解耦，确保可维护性和可扩展性。

---

## 快速开始

### 环境依赖

- Docker >= 24
- Docker Compose >= 2.18
- Go >= 1.25
- Node.js >= 20

### 启动后端

```bash
# 启动数据库、RabbitMQ 等依赖服务
docker-compose -f docker-compose.yml up -d

# 运行 Go 服务
go run ./cmd/app
````

### 启动前端

```bash
cd web
npm install
npm run dev
```

前端默认运行在 `http://localhost:5173`

### 全套 Docker 部署

```bash
docker-compose -f docker-compose-full.yml up -d
```

访问界面：

* Web 界面：`http://localhost:5173`
* API 文档：`http://localhost:8080/swagger`

---

## 项目结构

```
LiteDock/
├── cmd/app/             # Go 后端入口
├── config/              # 配置管理（环境变量）
├── internal/
│   ├── app/             # 核心应用逻辑
│   ├── controller/      # REST & RPC 控制器
│   ├── entity/          # 业务实体
│   ├── repo/            # 数据持久化逻辑
│   └── usecase/         # 业务用例层
├── pkg/                 # 辅助库 (docker client, logger 等)
├── web/                 # Vue3 + Vite 前端
├── docs/                # 文档、Swagger、图片
├── docker-compose.yml   # 开发环境 Compose 文件
├── docker-compose-full.yml # 生产/完整服务 Compose 文件
└── Makefile             # 常用命令封装
```

### 核心分层说明

* **内部层 (`internal`)**：包含业务逻辑和核心功能，不直接依赖外部库
* **控制器层 (`controller`)**：处理 HTTP / RPC 请求，将数据传递给业务逻辑
* **实体层 (`entity`)**：定义业务对象和数据结构
* **用例层 (`usecase`)**：封装业务流程和逻辑
* **外部工具 (`pkg`)**：Docker 客户端、日志、辅助工具

---

## 功能示例

### 容器管理 API

* 启动容器
* 停止容器
* 重启容器
* 查看容器状态

### 镜像管理

* 列出本地镜像
* 拉取镜像
* 删除镜像

### 网络与卷管理

* 列出网络和卷
* 创建和删除网络、卷

### 前端功能

* 仪表盘查看 Docker 状态
* 容器日志实时展示
* 支持多容器操作批量管理
* AI 提示与优化建议

---

## 依赖注入

LiteDock 后端使用依赖注入解耦服务，核心逻辑通过构造函数注入依赖，方便测试与 Mock：

```go
type ContainerUseCase struct {
    repo ContainerRepository
}

func NewContainerUseCase(repo ContainerRepository) *ContainerUseCase {
    return &ContainerUseCase{repo: repo}
}
```

---

## Docker 整合

LiteDock 使用 Docker API 与本机 Docker 引擎交互，通过 **Docker Go SDK** 实现容器、镜像、网络和卷操作。

---

## 整洁架构原则

* **依赖倒置**：内层业务逻辑不依赖外层实现
* **解耦**：业务逻辑独立，易于测试
* **分层清晰**：controller -> usecase -> repository -> entity

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

## 相关链接

* [Docker 官方文档](https://docs.docker.com/)
* [Go 官方网站](https://golang.org/)
* [Vue3 官网](https://vuejs.org/)
* [Vite 官网](https://vitejs.dev/)

---

## 开源许可

MIT License © 2026 [lminimum](https://github.com/lminimum)

