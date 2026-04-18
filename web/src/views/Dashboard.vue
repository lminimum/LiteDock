<template>
  <div class="dashboard">
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon containers">
          <Box :size="24" />
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
          <ImageIcon :size="24" />
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
          <Network :size="24" />
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
          <HardDrive :size="24" />
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

    <div class="dashboard-content">
      <div class="dashboard-left">
        <div class="card">
          <div class="card-header">
            <h3>系统资源</h3>
            <button @click="refreshResources" class="btn btn-ghost btn-sm" :disabled="refreshing">
              <RefreshCw :size="16" :class="{ 'spinning': refreshing }" />
            </button>
          </div>
          <div class="card-content">
            <div class="resource-item">
              <div class="resource-info">
                <span class="resource-label">CPU 使用率</span>
                <span class="resource-value">{{ resources.cpu }}%</span>
              </div>
              <div class="progress">
                <div class="progress-bar" :style="{ width: resources.cpu + '%' }"></div>
              </div>
            </div>

            <div class="resource-item">
              <div class="resource-info">
                <span class="resource-label">内存使用率</span>
                <span class="resource-value">{{ resources.memory }}%</span>
              </div>
              <div class="progress">
                <div class="progress-bar" :style="{ width: resources.memory + '%' }"></div>
              </div>
            </div>

            <div class="resource-item">
              <div class="resource-info">
                <span class="resource-label">磁盘使用率</span>
                <span class="resource-value">{{ resources.disk }}%</span>
              </div>
              <div class="progress">
                <div class="progress-bar" :style="{ width: resources.disk + '%' }"></div>
              </div>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-header">
            <h3>最近活动</h3>
            <router-link to="/containers" class="view-all">查看全部</router-link>
          </div>
          <div class="card-content">
            <div class="activity-list">
              <div v-for="activity in recentActivities" :key="activity.id" class="activity-item">
                <div class="activity-icon" :class="activity.type">
                  <component :is="getActivityIcon(activity.type)" :size="16" />
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
        <div class="card">
          <div class="card-header">
            <h3>快速操作</h3>
          </div>
          <div class="card-content">
            <div class="quick-actions">
              <button @click="createContainer" class="quick-action-btn">
                <Plus :size="18" />
                <span>创建容器</span>
              </button>
              <button @click="pullImage" class="quick-action-btn">
                <Download :size="18" />
                <span>拉取镜像</span>
              </button>
              <button @click="createNetwork" class="quick-action-btn">
                <PlusCircle :size="18" />
                <span>创建网络</span>
              </button>
              <button @click="createVolume" class="quick-action-btn">
                <PlusCircle :size="18" />
                <span>创建存储卷</span>
              </button>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-header">
            <h3>系统状态</h3>
          </div>
          <div class="card-content">
            <div class="status-list">
              <div class="status-item">
                <span class="status-label">Docker 服务</span>
                <span class="badge" :class="systemStatus.docker ? 'badge-success' : 'badge-error'">
                  {{ systemStatus.docker ? '在线' : '离线' }}
                </span>
              </div>
              <div class="status-item">
                <span class="status-label">API 服务</span>
                <span class="badge badge-success">在线</span>
              </div>
              <div class="status-item">
                <span class="status-label">数据库</span>
                <span class="badge" :class="systemStatus.database ? 'badge-success' : 'badge-error'">
                  {{ systemStatus.database ? '在线' : '离线' }}
                </span>
              </div>
              <div class="status-item">
                <span class="status-label">消息队列</span>
                <span class="badge" :class="systemStatus.messageQueue ? 'badge-success' : 'badge-error'">
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
import { ref, reactive, onMounted, markRaw } from 'vue'
import {
  Box,
  Image as ImageIcon,
  Network,
  HardDrive,
  RefreshCw,
  Plus,
  Download,
  PlusCircle,
  Globe
} from 'lucide-vue-next'

const refreshing = ref(false)

const stats = reactive({
  containers: { total: 12, running: 8, stopped: 4 },
  images: { total: 25, size: '2.3 GB' },
  networks: { total: 6, active: 4 },
  volumes: { total: 8, size: '15.7 GB' }
})

const resources = reactive({ cpu: 45, memory: 62, disk: 38 })

const systemStatus = reactive({ docker: true, database: true, messageQueue: true })

const recentActivities = ref([
  { id: 1, type: 'container', title: '容器 web-server 已启动', time: new Date(Date.now() - 5 * 60 * 1000) },
  { id: 2, type: 'image', title: '镜像 nginx:latest 已拉取', time: new Date(Date.now() - 15 * 60 * 1000) },
  { id: 3, type: 'container', title: '容器 database 已停止', time: new Date(Date.now() - 30 * 60 * 1000) },
  { id: 4, type: 'network', title: '网络 frontend-network 已创建', time: new Date(Date.now() - 45 * 60 * 1000) },
  { id: 5, type: 'volume', title: '存储卷 data-volume 已删除', time: new Date(Date.now() - 60 * 60 * 1000) }
])

const iconMap: Record<string, ReturnType<typeof markRaw>> = {
  container: markRaw(Box),
  image: markRaw(ImageIcon),
  network: markRaw(Globe),
  volume: markRaw(HardDrive)
}

const getActivityIcon = (type: string) => iconMap[type] || markRaw(Box)

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
    await new Promise(resolve => setTimeout(resolve, 1000))
    resources.cpu = Math.floor(Math.random() * 100)
    resources.memory = Math.floor(Math.random() * 100)
    resources.disk = Math.floor(Math.random() * 100)
  } finally {
    refreshing.value = false
  }
}

const createContainer = () => console.log('创建容器')
const pullImage = () => console.log('拉取镜像')
const createNetwork = () => console.log('创建网络')
const createVolume = () => console.log('创建存储卷')

onMounted(() => refreshResources())
</script>

<style scoped>
.dashboard {
  max-width: 1400px;
  margin: 0 auto;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: var(--space-4);
  margin-bottom: var(--space-6);
}

.stat-card {
  background: var(--color-background);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-md);
  padding: var(--space-6);
  display: flex;
  align-items: flex-start;
  gap: var(--space-4);
  transition: box-shadow var(--transition-fast), border-color var(--transition-fast);
}

.stat-card:hover {
  box-shadow: var(--shadow-md);
  border-color: var(--color-border);
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: var(--color-background-weak);
  color: var(--color-text-strong);
}

.stat-icon.containers { background: var(--color-background-interactive-weaker); color: var(--color-background-strong); }
.stat-icon.images { background: var(--color-background-weak); color: var(--color-text); }
.stat-icon.networks { background: var(--color-background-weak); color: var(--color-text); }
.stat-icon.volumes { background: var(--color-background-weak); color: var(--color-text); }

.stat-content h3 {
  margin: 0 0 var(--space-1) 0;
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-strong);
}

.stat-content p {
  margin: 0 0 var(--space-2) 0;
  color: var(--color-text);
  font-size: var(--font-size-sm);
}

.stat-breakdown {
  display: flex;
  gap: var(--space-3);
  font-size: var(--font-size-xs);
}

.running { color: var(--color-success); }
.stopped { color: var(--color-error); }

.dashboard-content {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: var(--space-6);
}

.card {
  background: var(--color-background);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-md);
  margin-bottom: var(--space-6);
}

.card-header {
  padding: var(--space-4) var(--space-6);
  border-bottom: 1px solid var(--color-border-weak);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.card-header h3 {
  margin: 0;
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
}

.view-all {
  color: var(--color-text-strong);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
}

.view-all:hover {
  color: var(--color-text-weak);
}

.card-content {
  padding: var(--space-6);
}

.resource-item {
  margin-bottom: var(--space-4);
}

.resource-item:last-child {
  margin-bottom: 0;
}

.resource-info {
  display: flex;
  justify-content: space-between;
  margin-bottom: var(--space-2);
}

.resource-label {
  color: var(--color-text);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
}

.resource-value {
  color: var(--color-text-strong);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
}

.activity-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.activity-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.activity-icon {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: var(--color-background-weak);
  color: var(--color-text);
}

.activity-icon.container { background: var(--color-background-interactive-weaker); color: var(--color-background-strong); }
.activity-icon.image { background: var(--color-error-bg); color: var(--color-error); }
.activity-icon.network { background: var(--color-success-bg); color: var(--color-success); }
.activity-icon.volume { background: var(--color-warning-bg); color: var(--color-warning); }

.activity-content {
  flex: 1;
  min-width: 0;
}

.activity-title {
  color: var(--color-text-strong);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  margin-bottom: 2px;
}

.activity-time {
  color: var(--color-text-weaker);
  font-size: var(--font-size-xs);
}

.quick-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-3);
}

.quick-action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-4);
  background: var(--color-background-weak);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--transition-fast);
  color: var(--color-text);
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
}

.quick-action-btn:hover {
  background: var(--color-background-interactive);
  border-color: var(--color-background-interactive);
  color: var(--color-background-strong);
}

.status-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-label {
  color: var(--color-text);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
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
