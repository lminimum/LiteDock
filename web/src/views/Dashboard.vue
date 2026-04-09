<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon containers">
          <span>📦</span>
        </div>
        <div class="stat-content">
          <h3>{{ stats.containers.total }}</h3>
          <p>容器总数</p>
          <div class="stat-breakdown">
            <span class="running">{{ stats.containers.running }} 运行中</span>
            <span class="stopped">{{ stats.containers.stopped }} 已停止</span>
          </div>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon images">
          <span>🖼️</span>
        </div>
        <div class="stat-content">
          <h3>{{ stats.images.total }}</h3>
          <p>镜像总数</p>
          <div class="stat-breakdown">
            <span>{{ stats.images.size }} 总大小</span>
          </div>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon networks">
          <span>🌐</span>
        </div>
        <div class="stat-content">
          <h3>{{ stats.networks.total }}</h3>
          <p>网络总数</p>
          <div class="stat-breakdown">
            <span>{{ stats.networks.active }} 活跃</span>
          </div>
        </div>
      </div>
      
      <div class="stat-card">
        <div class="stat-icon volumes">
          <span>💾</span>
        </div>
        <div class="stat-content">
          <h3>{{ stats.volumes.total }}</h3>
          <p>存储卷总数</p>
          <div class="stat-breakdown">
            <span>{{ stats.volumes.size }} 总大小</span>
          </div>
        </div>
      </div>
    </div>
    
    <!-- 主要内容区域 -->
    <div class="dashboard-content">
      <div class="dashboard-left">
        <!-- 系统资源 -->
        <div class="card">
          <div class="card-header">
            <h3>系统资源</h3>
            <button @click="refreshResources" class="refresh-btn" :disabled="refreshing">
              🔄
            </button>
          </div>
          <div class="card-content">
            <div class="resource-item">
              <div class="resource-info">
                <span class="resource-label">CPU 使用率</span>
                <span class="resource-value">{{ resources.cpu }}%</span>
              </div>
              <div class="progress-bar">
                <div class="progress-fill cpu" :style="{ width: resources.cpu + '%' }"></div>
              </div>
            </div>
            
            <div class="resource-item">
              <div class="resource-info">
                <span class="resource-label">内存使用率</span>
                <span class="resource-value">{{ resources.memory }}%</span>
              </div>
              <div class="progress-bar">
                <div class="progress-fill memory" :style="{ width: resources.memory + '%' }"></div>
              </div>
            </div>
            
            <div class="resource-item">
              <div class="resource-info">
                <span class="resource-label">磁盘使用率</span>
                <span class="resource-value">{{ resources.disk }}%</span>
              </div>
              <div class="progress-bar">
                <div class="progress-fill disk" :style="{ width: resources.disk + '%' }"></div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 最近活动 -->
        <div class="card">
          <div class="card-header">
            <h3>最近活动</h3>
            <router-link to="/containers" class="view-all">查看全部</router-link>
          </div>
          <div class="card-content">
            <div class="activity-list">
              <div v-for="activity in recentActivities" :key="activity.id" class="activity-item">
                <div class="activity-icon" :class="activity.type">
                  {{ getActivityIcon(activity.type) }}
                </div>
                <div class="activity-content">
                  <div class="activity-title">{{ activity.title }}</div>
                  <div class="activity-time">{{ formatTime(activity.time) }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <div class="dashboard-right">
        <!-- 快速操作 -->
        <div class="card">
          <div class="card-header">
            <h3>快速操作</h3>
          </div>
          <div class="card-content">
            <div class="quick-actions">
              <button @click="createContainer" class="quick-action-btn">
                <span class="action-icon">➕</span>
                <span>创建容器</span>
              </button>
              <button @click="pullImage" class="quick-action-btn">
                <span class="action-icon">⬇️</span>
                <span>拉取镜像</span>
              </button>
              <button @click="createNetwork" class="quick-action-btn">
                <span class="action-icon">🌐</span>
                <span>创建网络</span>
              </button>
              <button @click="createVolume" class="quick-action-btn">
                <span class="action-icon">💾</span>
                <span>创建存储卷</span>
              </button>
            </div>
          </div>
        </div>
        
        <!-- 系统状态 -->
        <div class="card">
          <div class="card-header">
            <h3>系统状态</h3>
          </div>
          <div class="card-content">
            <div class="status-list">
              <div class="status-item">
                <span class="status-label">Docker 服务</span>
                <span class="status-badge" :class="systemStatus.docker ? 'online' : 'offline'">
                  {{ systemStatus.docker ? '在线' : '离线' }}
                </span>
              </div>
              <div class="status-item">
                <span class="status-label">API 服务</span>
                <span class="status-badge online">在线</span>
              </div>
              <div class="status-item">
                <span class="status-label">数据库</span>
                <span class="status-badge" :class="systemStatus.database ? 'online' : 'offline'">
                  {{ systemStatus.database ? '在线' : '离线' }}
                </span>
              </div>
              <div class="status-item">
                <span class="status-label">消息队列</span>
                <span class="status-badge" :class="systemStatus.messageQueue ? 'online' : 'offline'">
                  {{ systemStatus.messageQueue ? '在线' : '离线' }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'

const refreshing = ref(false)

const stats = reactive({
  containers: {
    total: 12,
    running: 8,
    stopped: 4
  },
  images: {
    total: 25,
    size: '2.3 GB'
  },
  networks: {
    total: 6,
    active: 4
  },
  volumes: {
    total: 8,
    size: '15.7 GB'
  }
})

const resources = reactive({
  cpu: 45,
  memory: 62,
  disk: 38
})

const systemStatus = reactive({
  docker: true,
  database: true,
  messageQueue: true
})

const recentActivities = ref([
  {
    id: 1,
    type: 'container',
    title: '容器 web-server 已启动',
    time: new Date(Date.now() - 5 * 60 * 1000)
  },
  {
    id: 2,
    type: 'image',
    title: '镜像 nginx:latest 已拉取',
    time: new Date(Date.now() - 15 * 60 * 1000)
  },
  {
    id: 3,
    type: 'container',
    title: '容器 database 已停止',
    time: new Date(Date.now() - 30 * 60 * 1000)
  },
  {
    id: 4,
    type: 'network',
    title: '网络 frontend-network 已创建',
    time: new Date(Date.now() - 45 * 60 * 1000)
  },
  {
    id: 5,
    type: 'volume',
    title: '存储卷 data-volume 已删除',
    time: new Date(Date.now() - 60 * 60 * 1000)
  }
])

const getActivityIcon = (type: string) => {
  const icons = {
    container: '📦',
    image: '🖼️',
    network: '🌐',
    volume: '💾'
  }
  return icons[type] || '📋'
}

const formatTime = (time: Date) => {
  const now = new Date()
  const diff = now.getTime() - time.getTime()
  const minutes = Math.floor(diff / (1000 * 60))
  
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes} 分钟前`
  
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours} 小时前`
  
  const days = Math.floor(hours / 24)
  return `${days} 天前`
}

const refreshResources = async () => {
  refreshing.value = true
  try {
    // 模拟刷新资源数据
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    // 更新资源数据
    resources.cpu = Math.floor(Math.random() * 100)
    resources.memory = Math.floor(Math.random() * 100)
    resources.disk = Math.floor(Math.random() * 100)
  } finally {
    refreshing.value = false
  }
}

const createContainer = () => {
  // 导航到容器创建页面
  console.log('创建容器')
}

const pullImage = () => {
  // 打开镜像拉取对话框
  console.log('拉取镜像')
}

const createNetwork = () => {
  // 打开网络创建对话框
  console.log('创建网络')
}

const createVolume = () => {
  // 打开存储卷创建对话框
  console.log('创建存储卷')
}

onMounted(() => {
  // 初始化数据
  refreshResources()
})
</script>

<style scoped>
.dashboard {
  max-width: 1400px;
  margin: 0 auto;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 24px;
}

.stat-card {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  gap: 16px;
  transition: transform 0.2s, box-shadow 0.2s;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
}

.stat-icon.containers {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.images {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-icon.networks {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stat-icon.volumes {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.stat-content h3 {
  margin: 0 0 4px 0;
  font-size: 2rem;
  font-weight: 700;
  color: #1e293b;
}

.stat-content p {
  margin: 0 0 8px 0;
  color: #64748b;
  font-size: 0.875rem;
}

.stat-breakdown {
  display: flex;
  gap: 12px;
  font-size: 0.75rem;
}

.running {
  color: #10b981;
}

.stopped {
  color: #ef4444;
}

.dashboard-content {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: 24px;
}

.card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  margin-bottom: 24px;
}

.card-header {
  padding: 20px 24px;
  border-bottom: 1px solid #f1f5f9;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-header h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: #1e293b;
}

.refresh-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: background 0.2s;
}

.refresh-btn:hover:not(:disabled) {
  background: #f8fafc;
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.view-all {
  color: #667eea;
  text-decoration: none;
  font-size: 0.875rem;
  font-weight: 500;
}

.view-all:hover {
  text-decoration: underline;
}

.card-content {
  padding: 24px;
}

.resource-item {
  margin-bottom: 20px;
}

.resource-item:last-child {
  margin-bottom: 0;
}

.resource-info {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.resource-label {
  color: #374151;
  font-weight: 500;
}

.resource-value {
  color: #1e293b;
  font-weight: 600;
}

.progress-bar {
  height: 8px;
  background: #f1f5f9;
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.3s ease;
}

.progress-fill.cpu {
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
}

.progress-fill.memory {
  background: linear-gradient(90deg, #f093fb 0%, #f5576c 100%);
}

.progress-fill.disk {
  background: linear-gradient(90deg, #43e97b 0%, #38f9d7 100%);
}

.activity-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.activity-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.activity-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.125rem;
}

.activity-icon.container {
  background: #f0f4ff;
  color: #667eea;
}

.activity-icon.image {
  background: #fef3f2;
  color: #ef4444;
}

.activity-icon.network {
  background: #f0fdf4;
  color: #10b981;
}

.activity-icon.volume {
  background: #fefce8;
  color: #f59e0b;
}

.activity-content {
  flex: 1;
}

.activity-title {
  color: #1e293b;
  font-weight: 500;
  margin-bottom: 2px;
}

.activity-time {
  color: #64748b;
  font-size: 0.75rem;
}

.quick-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.quick-action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 20px 16px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  text-decoration: none;
  color: #374151;
}

.quick-action-btn:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
  transform: translateY(-1px);
}

.action-icon {
  font-size: 1.5rem;
}

.quick-action-btn span:last-child {
  font-size: 0.875rem;
  font-weight: 500;
}

.status-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-label {
  color: #374151;
  font-weight: 500;
}

.status-badge {
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
}

.status-badge.online {
  background: #dcfce7;
  color: #16a34a;
}

.status-badge.offline {
  background: #fef2f2;
  color: #dc2626;
}

@media (max-width: 1024px) {
  .dashboard-content {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .quick-actions {
    grid-template-columns: 1fr;
  }
}
</style>