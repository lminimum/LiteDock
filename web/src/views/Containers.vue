<template>
  <div class="containers-page">
    <div class="page-header">
      <h1>容器管理</h1>
      <div class="header-actions">
        <button @click="refreshContainers" class="action-btn" :disabled="loading">
          🔄 刷新
        </button>
        <button @click="showCreateModal = true" class="btn-primary">
          ➕ 创建容器
        </button>
      </div>
    </div>
    
    <!-- 过滤器和搜索 -->
    <div class="filters">
      <div class="search-box">
        <input 
          v-model="searchQuery" 
          placeholder="搜索容器..." 
          type="text"
        />
      </div>
      <div class="filter-options">
        <select v-model="statusFilter">
          <option value="">所有状态</option>
          <option value="running">运行中</option>
          <option value="stopped">已停止</option>
          <option value="paused">已暂停</option>
        </select>
      </div>
    </div>
    
    <!-- 容器列表 -->
    <div class="containers-grid">
      <div 
        v-for="container in filteredContainers" 
        :key="container.id"
        class="container-card"
        :class="{ 'status-running': container.status === 'running' }"
      >
        <div class="container-header">
          <div class="container-name">{{ container.name }}</div>
          <div class="container-status" :class="`status-${container.status}`">
            {{ getStatusText(container.status) }}
          </div>
        </div>
        
        <div class="container-info">
          <div class="info-item">
            <span class="label">镜像:</span>
            <span class="value">{{ container.image }}</span>
          </div>
          <div class="info-item">
            <span class="label">端口:</span>
            <span class="value">{{ container.ports || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="label">创建时间:</span>
            <span class="value">{{ formatDate(container.createdAt) }}</span>
          </div>
        </div>
        
        <div class="container-actions">
          <button 
            v-if="container.status === 'stopped'"
            @click="startContainer(container.id)"
            class="btn-success"
          >
            ▶️ 启动
          </button>
          <button 
            v-if="container.status === 'running'"
            @click="stopContainer(container.id)"
            class="btn-warning"
          >
            ⏸️ 停止
          </button>
          <button 
            v-if="container.status === 'running'"
            @click="restartContainer(container.id)"
            class="btn-info"
          >
            🔄 重启
          </button>
          <button @click="showLogs(container.id)" class="btn-secondary">
            📋 日志
          </button>
          <button @click="deleteContainer(container.id)" class="btn-danger">
            🗑️ 删除
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

interface Container {
  id: string
  name: string
  image: string
  status: 'running' | 'stopped' | 'paused'
  ports: string
  createdAt: Date
}

const loading = ref(false)
const searchQuery = ref('')
const statusFilter = ref('')
const showCreateModal = ref(false)

const containers = ref<Container[]>([
  {
    id: '1',
    name: 'web-server',
    image: 'nginx:latest',
    status: 'running',
    ports: '80:8080',
    createdAt: new Date(Date.now() - 2 * 60 * 60 * 1000)
  },
  {
    id: '2',
    name: 'database',
    image: 'postgres:15',
    status: 'running',
    ports: '5432:5432',
    createdAt: new Date(Date.now() - 4 * 60 * 60 * 1000)
  },
  {
    id: '3',
    name: 'redis-cache',
    image: 'redis:7',
    status: 'stopped',
    ports: '6379:6379',
    createdAt: new Date(Date.now() - 6 * 60 * 60 * 1000)
  }
])

const filteredContainers = computed(() => {
  let filtered = containers.value
  
  if (searchQuery.value) {
    filtered = filtered.filter(container => 
      container.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      container.image.toLowerCase().includes(searchQuery.value.toLowerCase())
    )
  }
  
  if (statusFilter.value) {
    filtered = filtered.filter(container => container.status === statusFilter.value)
  }
  
  return filtered
})

const getStatusText = (status: string) => {
  const statusMap = {
    running: '运行中',
    stopped: '已停止',
    paused: '已暂停'
  }
  return statusMap[status] || status
}

const formatDate = (date: Date) => {
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}

const refreshContainers = async () => {
  loading.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 1000))
    // 这里应该调用API获取容器列表
  } finally {
    loading.value = false
  }
}

const startContainer = async (id: string) => {
  try {
    // 调用API启动容器
    const container = containers.value.find(c => c.id === id)
    if (container) {
      container.status = 'running'
    }
  } catch (error) {
    console.error('启动容器失败:', error)
  }
}

const stopContainer = async (id: string) => {
  try {
    // 调用API停止容器
    const container = containers.value.find(c => c.id === id)
    if (container) {
      container.status = 'stopped'
    }
  } catch (error) {
    console.error('停止容器失败:', error)
  }
}

const restartContainer = async (id: string) => {
  try {
    // 调用API重启容器
    console.log('重启容器:', id)
  } catch (error) {
    console.error('重启容器失败:', error)
  }
}

const showLogs = (id: string) => {
  // 显示容器日志
  console.log('显示日志:', id)
}

const deleteContainer = async (id: string) => {
  if (confirm('确定要删除这个容器吗？')) {
    try {
      // 调用API删除容器
      containers.value = containers.value.filter(c => c.id !== id)
    } catch (error) {
      console.error('删除容器失败:', error)
    }
  }
}

onMounted(() => {
  refreshContainers()
})
</script>

<style scoped>
.containers-page {
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-header h1 {
  margin: 0;
  color: #1e293b;
  font-size: 1.875rem;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.filters {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
  padding: 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.search-box {
  flex: 1;
}

.search-box input {
  width: 100%;
  padding: 10px 16px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 0.875rem;
}

.filter-options select {
  padding: 10px 16px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 0.875rem;
  background: white;
}

.containers-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 20px;
}

.container-card {
  background: white;
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border-left: 4px solid #e2e8f0;
  transition: all 0.2s;
}

.container-card.status-running {
  border-left-color: #10b981;
}

.container-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.container-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.container-name {
  font-size: 1.125rem;
  font-weight: 600;
  color: #1e293b;
}

.container-status {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
}

.status-running {
  background: #dcfce7;
  color: #16a34a;
}

.status-stopped {
  background: #fef2f2;
  color: #dc2626;
}

.status-paused {
  background: #fef3c7;
  color: #d97706;
}

.container-info {
  margin-bottom: 16px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
  font-size: 0.875rem;
}

.info-item .label {
  color: #64748b;
}

.info-item .value {
  color: #1e293b;
  font-weight: 500;
}

.container-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.container-actions button {
  padding: 6px 12px;
  border: none;
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: #667eea;
  color: white;
}

.btn-primary:hover {
  background: #5a67d8;
}

.btn-success {
  background: #10b981;
  color: white;
}

.btn-success:hover {
  background: #059669;
}

.btn-warning {
  background: #f59e0b;
  color: white;
}

.btn-warning:hover {
  background: #d97706;
}

.btn-info {
  background: #3b82f6;
  color: white;
}

.btn-info:hover {
  background: #2563eb;
}

.btn-secondary {
  background: #64748b;
  color: white;
}

.btn-secondary:hover {
  background: #475569;
}

.btn-danger {
  background: #ef4444;
  color: white;
}

.btn-danger:hover {
  background: #dc2626;
}

.action-btn {
  padding: 8px 16px;
  background: #f8fafc;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.action-btn:hover:not(:disabled) {
  background: #f1f5f9;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .containers-grid {
    grid-template-columns: 1fr;
  }
  
  .filters {
    flex-direction: column;
  }
  
  .container-actions {
    justify-content: center;
  }
}
</style>