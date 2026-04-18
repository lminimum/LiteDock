<p align="center">
  <img src="docs/img/logo.svg" width="75%" alt="LiteDock Logo">
</p>

[🇺🇸 English](README.md)

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

## 技术栈

- **后端**：Go 1.25+，整洁架构
- **前端**：Vue 3 + Vite + TypeScript
- **数据库**：PostgreSQL、MySQL、SQLite（通过抽象层支持）
- **API**：REST (Fiber)

---

## 快速开始

### 环境依赖

- Docker >= 24
- Docker Compose >= 2.18
- Go >= 1.25
- Node.js >= 20

### 启动后端

```bash
# 启动数据库和其他依赖服务
docker-compose -f docker-compose.yml up -d

# 运行 Go 服务
go run ./cmd/app
```

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

- Web 界面：`http://localhost:5173`
- API 文档：`http://localhost:8080/swagger`

---

## 项目结构

```
LiteDock/
├── cmd/app/              # Go 后端入口
├── config/               # 配置管理（环境变量）
├── internal/
│   ├── app/             # 核心应用逻辑
│   ├── controller/       # REST 控制器
│   │   └── restapi/      # Fiber HTTP handlers
│   ├── entity/           # 业务实体
│   ├── repo/             # 数据持久化逻辑
│   └── usecase/          # 业务用例层
├── pkg/                  # 辅助库 (httpserver, logger 等)
├── web/                  # Vue3 + Vite 前端
├── docs/                 # 文档、Swagger、图片
├── migrations/           # 数据库迁移文件
├── docker-compose.yml    # 开发环境 Compose 文件
└── Makefile             # 常用命令封装
```

### 核心分层说明

- **内部层 (`internal`)**：包含业务逻辑和核心功能，不直接依赖外部库
- **控制器层 (`controller`)**：处理 HTTP 请求
- **实体层 (`entity`)**：定义业务对象和数据结构
- **用例层 (`usecase`)**：封装业务流程和逻辑
- **外部工具 (`pkg`)**：HTTP 服务器、日志、数据库连接等

---

## 常用命令

```bash
# 后端开发
make run                 # 完整开发环境（依赖 + swagger + 迁移）
make deps               # 整理和验证依赖
make swag-v1           # 生成 Swagger 文档
make format            # 代码格式化（gofumpt + gci）
make linter-golangci   # 运行 golangci-lint
make test              # 运行单元测试
make integration-test  # 运行集成测试
make mock              # 生成测试 mock
make pre-commit        # 完整检查（deps → swag → mock → format → lint → test）

# Docker
make compose-up        # 启动核心服务（数据库）
make compose-up-all    # 启动完整服务
make compose-down      # 停止所有容器

# 数据库迁移
make migrate-create name=xxx  # 创建迁移文件
make migrate-up              # 运行迁移
```

---

## Git 工作流

### 分支命名规范

```
main                    # 稳定发布分支
dev                     # 开发分支
feature/<描述>         # 新功能（如：feature/user-auth）
fix/<描述>             # Bug 修复（如：fix/container-crash）
refactor/<描述>        # 代码重构
docs/<描述>            # 文档更新
```

### 提交信息规范

遵循 [Conventional Commits](https://www.conventionalcommits.org/)：

```
<类型>(<范围>): <描述>

[可选正文]

[可选脚注]
```

**类型**：
| 类型 | 描述 |
|------|------|
| `feat` | 新功能 |
| `fix` | Bug 修复 |
| `docs` | 仅文档更改 |
| `style` | 格式、分号等 |
| `refactor` | 重构（非修 bug 非加功能）|
| `perf` | 性能改进 |
| `test` | 添加或更新测试 |
| `chore` | 维护任务（依赖、构建、CI）|

**示例**：
```bash
feat(backend): add user authentication via JWT
fix(docker): handle container restart timeout
docs(api): update API endpoint documentation
```

### 推送 & PR 流程

```
1. 同步最新代码
   git checkout main
   git pull origin main

2. 创建功能分支
   git checkout -b feature/my-feature

3. 编写代码并提交（遵循提交规范）
   git add .
   git commit -m "feat(api): add new endpoint"

4. PR 前 rebase 保持历史整洁
   git fetch origin main
   git rebase origin/main

5. 推送并创建 PR
   git push -u origin feature/my-feature

6. PR 合并后，删除分支
   git branch -d feature/my-feature
   git push origin --delete feature/my-feature
```

---

## 功能示例

### 容器管理 API

- 启动容器
- 停止容器
- 重启容器
- 查看容器状态

### 镜像管理

- 列出本地镜像
- 拉取镜像
- 删除镜像

### 网络与卷管理

- 列出网络和卷
- 创建和删除网络、卷

### 前端功能

- 仪表盘查看 Docker 状态
- 容器日志实时展示
- 支持多容器操作批量管理
- AI 提示与优化建议

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

- **依赖倒置**：内层业务逻辑不依赖外层实现
- **解耦**：业务逻辑独立，易于测试
- **分层清晰**：controller -> usecase -> repository -> entity

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

- [Docker 官方文档](https://docs.docker.com/)
- [Go 官方网站](https://golang.org/)
- [Vue3 官网](https://vuejs.org/)
- [Vite 官网](https://vitejs.dev/)

---

## 开源许可

MIT License © 2026 [lminimum](https://github.com/lminimum)
