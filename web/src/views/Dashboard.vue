<template>
  <div class="dashboard">
    <PageHeader :title="t('dashboard.title')" />
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon containers">
          <Box :size="24" />
        </div>
        <div class="stat-content">
          <h3>{{ stats.containers.total }}</h3>
          <p>{{ t('dashboard.totalContainers') }}</p>
          <div class="stat-breakdown">
            <span class="running">{{ stats.containers.running }} {{ t('dashboard.running') }}</span>
            <span class="stopped">{{ stats.containers.stopped }} {{ t('dashboard.stopped') }}</span>
          </div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon images">
          <ImageIcon :size="24" />
        </div>
        <div class="stat-content">
          <h3>{{ stats.images.total }}</h3>
          <p>{{ t('dashboard.totalImages') }}</p>
          <div class="stat-breakdown">
            <span>{{ stats.images.size }} {{ t('dashboard.totalSize') }}</span>
          </div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon networks">
          <Network :size="24" />
        </div>
        <div class="stat-content">
          <h3>{{ stats.networks.total }}</h3>
          <p>{{ t('dashboard.totalNetworks') }}</p>
          <div class="stat-breakdown">
            <span>{{ stats.networks.active }} {{ t('dashboard.active') }}</span>
          </div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon volumes">
          <HardDrive :size="24" />
        </div>
        <div class="stat-content">
          <h3>{{ stats.volumes.total }}</h3>
          <p>{{ t('dashboard.totalVolumes') }}</p>
          <div class="stat-breakdown">
            <span>{{ stats.volumes.size }} {{ t('dashboard.totalSize') }}</span>
          </div>
        </div>
      </div>
    </div>

    <div class="dashboard-content">
      <div class="dashboard-left">
        <div class="card">
          <div class="card-header">
            <h3>{{ t('dashboard.systemResources') }}</h3>
            <button @click="refreshResources" class="btn btn-ghost btn-sm" :disabled="refreshing">
              <RefreshCw :size="16" :class="{ 'spinning': refreshing }" />
            </button>
          </div>
          <div class="card-content">
            <div class="resource-item">
              <div class="resource-info">
                <span class="resource-label">{{ t('dashboard.cpu') }}</span>
                <span class="resource-value">{{ resources.cpu }}%</span>
              </div>
              <div class="progress">
                <div class="progress-bar" :style="{ width: resources.cpu + '%' }"></div>
              </div>
            </div>

            <div class="resource-item">
              <div class="resource-info">
                <span class="resource-label">{{ t('dashboard.memory') }}</span>
                <span class="resource-value">{{ resources.memory }}%</span>
              </div>
              <div class="progress">
                <div class="progress-bar" :style="{ width: resources.memory + '%' }"></div>
              </div>
            </div>

            <div class="resource-item">
              <div class="resource-info">
                <span class="resource-label">{{ t('dashboard.disk') }}</span>
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
            <h3>{{ t('dashboard.recentActivity') }}</h3>
            <router-link to="/containers" class="view-all">{{ t('dashboard.viewAll') }}</router-link>
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
            <h3>{{ t('dashboard.quickActions') }}</h3>
          </div>
          <div class="card-content">
            <div class="quick-actions">
              <button @click="createContainer" class="quick-action-btn">
                <Plus :size="18" />
                <span>{{ t('dashboard.createContainer') }}</span>
              </button>
              <button @click="pullImage" class="quick-action-btn">
                <Download :size="18" />
                <span>{{ t('dashboard.pullImage') }}</span>
              </button>
              <button @click="createNetwork" class="quick-action-btn">
                <PlusCircle :size="18" />
                <span>{{ t('dashboard.createNetwork') }}</span>
              </button>
              <button @click="createVolume" class="quick-action-btn">
                <PlusCircle :size="18" />
                <span>{{ t('dashboard.createVolume') }}</span>
              </button>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-header">
            <h3>{{ t('dashboard.systemStatus') }}</h3>
          </div>
          <div class="card-content">
            <div class="status-list">
              <div class="status-item">
                <span class="status-label">{{ t('dashboard.dockerService') }}</span>
                <span class="badge" :class="systemStatus.docker ? 'badge-success' : 'badge-error'">
                  {{ systemStatus.docker ? t('dashboard.online') : t('dashboard.offline') }}
                </span>
              </div>
              <div class="status-item">
                <span class="status-label">{{ t('dashboard.apiService') }}</span>
                <span class="badge badge-success">{{ t('dashboard.online') }}</span>
              </div>
              <div class="status-item">
                <span class="status-label">{{ t('dashboard.database') }}</span>
                <span class="badge" :class="systemStatus.database ? 'badge-success' : 'badge-error'">
                  {{ systemStatus.database ? t('dashboard.online') : t('dashboard.offline') }}
                </span>
              </div>
              <div class="status-item">
                <span class="status-label">{{ t('dashboard.messageQueue') }}</span>
                <span class="badge" :class="systemStatus.messageQueue ? 'badge-success' : 'badge-error'">
                  {{ systemStatus.messageQueue ? t('dashboard.online') : t('dashboard.offline') }}
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
import { t } from '@/i18n'
import PageHeader from '@/components/ui/PageHeader.vue'

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
  { id: 1, type: 'container', title: t('dashboard.containerStarted', { name: 'web-server' }), time: new Date(Date.now() - 5 * 60 * 1000) },
  { id: 2, type: 'image', title: t('dashboard.imagePulled', { name: 'nginx:latest' }), time: new Date(Date.now() - 15 * 60 * 1000) },
  { id: 3, type: 'container', title: t('dashboard.containerStopped', { name: 'database' }), time: new Date(Date.now() - 30 * 60 * 1000) },
  { id: 4, type: 'network', title: t('dashboard.networkCreated', { name: 'frontend-network' }), time: new Date(Date.now() - 45 * 60 * 1000) },
  { id: 5, type: 'volume', title: t('dashboard.volumeDeleted', { name: 'data-volume' }), time: new Date(Date.now() - 60 * 60 * 1000) }
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

  if (minutes < 1) return t('dashboard.justNow')
  if (minutes < 60) return t('dashboard.minutesAgo', { n: minutes })

  const hours = Math.floor(minutes / 60)
  if (hours < 24) return t('dashboard.hoursAgo', { n: hours })

  const days = Math.floor(hours / 24)
  return t('dashboard.daysAgo', { n: days })
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

const createContainer = () => console.log('createContainer')
const pullImage = () => console.log('pullImage')
const createNetwork = () => console.log('createNetwork')
const createVolume = () => console.log('createVolume')

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
  background: var(--color-background-weak);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-5);
  display: flex;
  align-items: center;
  gap: var(--space-4);
  transition: all var(--transition-base);
}

.stat-card:hover {
  border-color: var(--color-border);
  box-shadow: var(--shadow-md);
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-icon.containers { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
.stat-icon.images { background: rgba(168, 85, 247, 0.1); color: #a855f7; }
.stat-icon.networks { background: rgba(34, 197, 94, 0.1); color: #22c55e; }
.stat-icon.volumes { background: rgba(249, 115, 22, 0.1); color: #f97316; }

.stat-content h3 {
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-strong);
  margin: 0;
}

.stat-content p {
  font-size: var(--font-size-sm);
  color: var(--color-text-weak);
  margin: var(--space-1) 0;
}

.stat-breakdown {
  display: flex;
  gap: var(--space-3);
  font-size: var(--font-size-xs);
}

.stat-breakdown .running { color: #22c55e; }
.stat-breakdown .stopped { color: var(--color-text-weaker); }

.dashboard-content {
  display: grid;
  grid-template-columns: 1fr 350px;
  gap: var(--space-6);
}

.card {
  background: var(--color-background-weak);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  margin-bottom: var(--space-6);
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-4) var(--space-5);
  border-bottom: 1px solid var(--color-border);
}

.card-header h3 {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
  margin: 0;
}

.card-content {
  padding: var(--space-5);
}

.view-all {
  font-size: var(--font-size-sm);
  color: var(--color-text-strong);
  text-decoration: none;
}

.view-all:hover {
  text-decoration: underline;
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
  font-size: var(--font-size-sm);
  color: var(--color-text-weak);
}

.resource-value {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-strong);
}

.progress {
  height: 6px;
  background: var(--color-background-strong);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.progress-bar {
  height: 100%;
  background: var(--color-background-strong);
  border-radius: var(--radius-full);
  transition: width var(--transition-base);
}

.activity-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.activity-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.activity-icon {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.activity-icon.container { background: rgba(59, 130, 246, 0.1); color: #3b82f6; }
.activity-icon.image { background: rgba(168, 85, 247, 0.1); color: #a855f7; }
.activity-icon.network { background: rgba(34, 197, 94, 0.1); color: #22c55e; }
.activity-icon.volume { background: rgba(249, 115, 22, 0.1); color: #f97316; }

.activity-title {
  font-size: var(--font-size-sm);
  color: var(--color-text-strong);
}

.activity-time {
  font-size: var(--font-size-xs);
  color: var(--color-text-weaker);
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
  background: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all var(--transition-base);
  color: var(--color-text-weak);
  font-size: var(--font-size-sm);
}

.quick-action-btn:hover {
  border-color: var(--color-text-strong);
  color: var(--color-text-strong);
  background: rgba(59, 130, 246, 0.05);
}

.status-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.status-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.status-label {
  font-size: var(--font-size-sm);
  color: var(--color-text-weak);
}

.badge {
  font-size: var(--font-size-xs);
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-full);
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
