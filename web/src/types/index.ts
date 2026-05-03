// API 响应类型
export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  message?: string
  error?: string
}

// 分页类型
export interface PaginationParams {
  page: number
  limit: number
  sort?: string
  order?: 'asc' | 'desc'
}

export interface PaginationResponse<T> {
  items: T[]
  total: number
  page: number
  limit: number
  totalPages: number
}

// 容器相关类型
export interface Container {
  id: string
  name: string
  image: string
  status: 'running' | 'stopped' | 'paused' | 'restarting' | 'exited'
  ports: string[]
  createdAt: string
  startedAt?: string
  labels?: Record<string, string>
  mounts?: Array<{
    type: string
    source: string
    destination: string
  }>
}

// 镜像相关类型
export interface Image {
  id: string
  repository: string
  tag: string
  size: number
  createdAt: string
}

// 网络相关类型
export interface Network {
  id: string
  name: string
  driver: string
  scope: string
  internal: boolean
  containers?: string[]
}

// 存储卷相关类型
export interface Volume {
  name: string
  driver: string
  mountpoint: string
  createdAt: string
  size?: number
}

// 系统资源类型
export interface SystemResources {
  cpu: {
    usage: number
    cores: number
  }
  memory: {
    used: number
    total: number
    usage: number
  }
  disk: {
    used: number
    total: number
    usage: number
  }
}

// 活动日志类型
export interface ActivityLog {
  id: string
  type: 'container' | 'image' | 'network' | 'volume'
  action: string
  resource: string
  status: 'success' | 'error' | 'warning'
  message: string
  timestamp: string
}

// 用户类型
export interface User {
  id: string
  username: string
  email?: string
  role: string
  permissions: string[]
}

// 认证相关类型
export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
  expiresIn: number
}

export type AuthMethod = 'password' | 'key'

export interface RemoteMachine {
  id: string
  name: string
  host: string
  port: number
  username: string
  auth_method: AuthMethod
  ssh_key_path?: string
  docker_host: string
  status: 'online' | 'offline' | 'unknown'
  created_at: string
  updated_at: string
}

export interface CreateMachineRequest {
  name: string
  host: string
  port?: number
  username: string
  auth_method: AuthMethod
  password?: string
  ssh_key?: string
  ssh_key_path?: string
  docker_host?: string
}

export interface UpdateMachineRequest {
  name?: string
  host?: string
  port?: number
  username?: string
  auth_method?: AuthMethod
  password?: string
  ssh_key?: string
  ssh_key_path?: string
  docker_host?: string
}

export interface RemoteContainer {
  id: string
  name: string
  image: string
  status: 'running' | 'stopped' | 'paused' | 'restarting' | 'exited' | 'created'
  ports: string[]
  createdAt: string
  startedAt?: string
  labels?: Record<string, string>
  mounts?: Array<{
    type: string
    source: string
    destination: string
  }>
}

// 配置类型
export interface AppConfig {
  system: {
    name: string
    description: string
    port: number
    enableMetrics: boolean
    enableSwagger: boolean
  }
  docker: {
    type: 'local' | 'remote'
    host?: string
    tlsPath?: string
  }
  auth: {
    sessionTimeout: number
    enableTwoFactor: boolean
  }
  monitoring: {
    enabled: boolean
    interval: number
    retention: number
  }
}