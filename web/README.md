# LiteDock 前端路由结构

## 概述

LiteDock 前端使用 Vue 3 + Vue Router 4 + Pinia 构建，采用单页面应用架构。

## 路由设计

### 认证流程

1. **首次配置** (`/setup`) - 系统初始化配置向导
2. **登录** (`/login`) - 用户认证页面
3. **主应用** (`/`) - 需要认证的主界面

### 路由守卫

- 检查系统是否已配置 (`localStorage.getItem('litedock-configured')`)
- 检查用户认证状态 (`authStore.isAuthenticated`)
- 自动重定向到相应页面

## 页面结构

### 主要功能页面

- **概览** (`/`) - Dashboard 仪表盘
- **容器管理** (`/containers`) - Docker 容器管理
- **编排管理** (`/orchestration`) - Docker Compose/Kubernetes
- **镜像管理** (`/images`) - Docker 镜像管理
- **网络管理** (`/networks`) - Docker 网络管理
- **存储卷管理** (`/volumes`) - Docker 存储卷管理
- **系统设置** (`/settings`) - 系统配置

### 特殊页面

- **404** (`/:pathMatch(.*)*`) - 页面未找到

## 布局系统

### MainLayout

主要布局包含：
- 侧边栏导航 (可折叠)
- 顶部导航栏
- 主内容区域

#### 侧边栏功能

- 导航菜单
- 用户信息显示
- 折叠/展开控制

#### 顶部导航功能

- 面包屑导航
- 刷新按钮
- 主题切换 (暗色/亮色)
- 通知中心
- 用户菜单

## 状态管理

### Pinia Store

#### Auth Store (`stores/auth.ts`)
- 用户认证状态
- 登录/登出功能
- Token 管理
- 用户信息管理

## API 集成

### Axios 配置

- 基础 URL 配置
- 请求拦截器 (添加认证头)
- 响应拦截器 (错误处理)
- 自动重定向到登录页面

## 类型定义

### 完整的 TypeScript 类型

所有主要数据结构都有对应的 TypeScript 类型定义：

- `Container` - 容器信息
- `Image` - 镜像信息  
- `Network` - 网络信息
- `Volume` - 存储卷信息
- `User` - 用户信息
- `ApiResponse` - API 响应格式

## 环境配置

### 开发环境 (.env)
- API URL 配置
- 功能开关
- 应用基本信息

### 生产环境 (.env.production)
- 生产环境专用配置

## 开发指南

### 运行开发服务器

```bash
cd web
npm install
npm run dev
```

### 构建生产版本

```bash
npm run build
```

### 类型检查

```bash
npm run type-check
```

### 代码检查

```bash
npm run lint
```

## 组件结构

### 页面组件

所有页面组件位于 `src/views/` 目录：

```
views/
├── Dashboard.vue      # 概览仪表盘
├── Containers.vue     # 容器管理
├── Orchestration.vue  # 编排管理 (占位符)
├── Images.vue        # 镜像管理 (占位符)
├── Networks.vue      # 网络管理 (占位符)
├── Volumes.vue       # 存储卷管理 (占位符)
├── Settings.vue      # 系统设置
├── Setup.vue         # 首次配置
├── Login.vue         # 登录页面
└── NotFound.vue      # 404页面
```

### 布局组件

- `MainLayout.vue` - 主应用布局

### 工具组件

- `router/index.ts` - 路由配置
- `stores/auth.ts` - 认证状态管理
- `utils/api.ts` - API 工具

## 响应式设计

- 移动端适配
- 侧边栏自动折叠
- 网格布局自适应

## 主题支持

- 亮色/暗色主题切换
- CSS 变量实现
- 主题状态持久化

## 开发注意事项

1. **路由守卫**：确保所有需要认证的页面都有适当的保护
2. **类型安全**：充分利用 TypeScript 类型检查
3. **组件复用**：提取可复用的 UI 组件
4. **错误处理**：统一的错误处理机制
5. **性能优化**：按需加载组件，优化打包体积

## Vue 3 + TypeScript + Vite

This template should help get you started developing with Vue 3 and TypeScript in Vite. The template uses Vue 3 `<script setup>` SFCs, check out the [script setup docs](https://v3.vuejs.org/api/sfc-script-setup.html#sfc-script-setup) to learn more.

Learn more about the recommended Project Setup and IDE Support in the [Vue Docs TypeScript Guide](https://vuejs.org/guide/typescript/overview.html#project-setup).