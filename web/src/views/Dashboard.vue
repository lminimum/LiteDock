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
        <div class="stat-icon machines">
          <Globe :size="24" />
        </div>
        <div class="stat-content">
          <h3>{{ stats.machines.total }}</h3>
          <p>{{ t('nav.machines') }}</p>
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
          </div>
          <div class="card-content">
            <SystemResourcesChart
              :labels="chartLabels"
              :cpu="chartCpu"
              :memory="chartMemory"
              :disk="chartDisk"
            />
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
import { ref, reactive, onMounted, onUnmounted, markRaw } from 'vue'
import {
  Box,
  Image as ImageIcon,
  Network,
  HardDrive,
  Plus,
  Download,
  PlusCircle,
  Globe
} from 'lucide-vue-next'
import { t } from '@/i18n'
import PageHeader from '@/components/ui/PageHeader.vue'
import SystemResourcesChart from '@/components/chart/SystemResourcesChart.vue'
import api from '@/utils/api'

const stats = reactive({
  containers: { total: 0, running: 0, stopped: 0 },
  images: { total: 0, size: '0 GB' },
  networks: { total: 0, active: 0 },
  volumes: { total: 0, size: '0 GB' },
  machines: { total: 0 }
})

const TOTAL_POINTS = 60
const UPDATE_INTERVAL_MS = 2000

const chartLabels = ref<string[]>([])
const chartCpu = ref<number[]>(new Array(TOTAL_POINTS).fill(0))
const chartMemory = ref<number[]>(new Array(TOTAL_POINTS).fill(0))
const chartDisk = ref<number[]>(new Array(TOTAL_POINTS).fill(0))

const initChartLabels = () => {
  const labels: string[] = []
  const now = new Date()
  for (let i = TOTAL_POINTS - 1; i >= 0; i--) {
    const t = new Date(now.getTime() - i * UPDATE_INTERVAL_MS)
    labels.push(t.toLocaleString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false }))
  }
  chartLabels.value = labels
}

const loadHistory = async (): Promise<boolean> => {
  try {
    const res = await api.get('/dashboard/resources/history?minutes=5')
    if (res.data.success && res.data.data.length > 0) {
      const history = res.data.data.slice(-TOTAL_POINTS)
      const cpu: number[] = []
      const memory: number[] = []
      const disk: number[] = []
      const labels: string[] = []

      for (const m of history) {
        cpu.push(m.cpu ?? 0)
        memory.push(m.memory ?? 0)
        disk.push(m.disk ?? 0)
        labels.push(m.time ?? '')
      }

      const fillCount = TOTAL_POINTS - cpu.length
      chartCpu.value = [...new Array(fillCount).fill(0), ...cpu]
      chartMemory.value = [...new Array(fillCount).fill(0), ...memory]
      chartDisk.value = [...new Array(fillCount).fill(0), ...disk]
      chartLabels.value = [...new Array(fillCount).fill(''), ...labels]
      return true
    }
    return false
  } catch (e) {
    console.error('Failed to load history:', e)
    return false
  }
}

let ws: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let isMounted = false

const startWS = () => {
  if (!isMounted) return
  // Use window.location to construct WebSocket URL for proper browser connection
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = import.meta.env.VITE_API_WS_HOST || `${window.location.hostname}:8080`
  const wsUrl = `${protocol}//${host}`
  ws = new WebSocket(`${wsUrl}/v1/dashboard/resources/stream`)

  ws.onopen = () => {
    console.log('WebSocket connected')
  }

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      const newCpu = [...chartCpu.value.slice(1), data.cpu ?? 0]
      const newMemory = [...chartMemory.value.slice(1), data.memory ?? 0]
      const newDisk = [...chartDisk.value.slice(1), data.disk ?? 0]
      chartCpu.value = newCpu
      chartMemory.value = newMemory
      chartDisk.value = newDisk

      const newLabels = [...chartLabels.value.slice(1), data.time ?? '']
      chartLabels.value = newLabels
    } catch (e) {
      console.error('Failed to parse WebSocket message:', e)
    }
  }

  ws.onerror = (e) => {
    console.error('WebSocket error:', e)
  }

  ws.onclose = () => {
    console.log('WebSocket disconnected')
    if (!isMounted) return
    reconnectTimer = setTimeout(startWS, 5000)
  }
}

const refreshStats = async () => {
  try {
    const statsRes = await api.get('/dashboard/stats')
    if (statsRes.data.success) {
      const { machines, containers } = statsRes.data.data
      stats.machines.total = machines.total
      if (containers) {
        stats.containers.total = containers.total || 0
        stats.containers.running = containers.running || 0
        stats.containers.stopped = containers.stopped || 0
      }
    }
  } catch (e) {
    console.error('Failed to fetch stats:', e)
  }
}

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

const createContainer = () => console.log('createContainer')
const pullImage = () => console.log('pullImage')
const createNetwork = () => console.log('createNetwork')
const createVolume = () => console.log('createVolume')

onMounted(async () => {
  isMounted = true
  const hasHistory = await loadHistory()
  if (!hasHistory) {
    initChartLabels()
  }
  startWS()
  refreshStats()
})

onUnmounted(() => {
  isMounted = false
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  if (ws) {
    ws.close()
    ws = null
  }
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
  gap: var(--space-4);
  margin-bottom: var(--space-6);
}

.stat-card {
  background: var(--color-background-weak);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: var(--space-5);
  display: flex;
  align-items: center;
  gap: var(--space-4);
  transition: border-color var(--transition-fast);
}

.stat-card:hover {
  border-color: var(--color-text-weaker);
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

.stat-icon.containers { background: var(--color-info-bg); color: var(--color-info); }
.stat-icon.machines { background: var(--color-info-bg); color: var(--color-info); }
.stat-icon.images { background: var(--color-accent); color: #fdfcfc; }
.stat-icon.networks { background: var(--color-success-bg); color: var(--color-success); }
.stat-icon.volumes { background: var(--color-warning-bg); color: var(--color-warning); }

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

.stat-breakdown .running { color: var(--color-success); }
.stat-breakdown .stopped { color: var(--color-text-weaker); }

.dashboard-content {
  display: grid;
  grid-template-columns: 1fr 350px;
  gap: var(--space-6);
}

.card {
  background: var(--color-background-weak);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
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
  color: var(--color-accent);
  text-decoration: underline;
  text-underline-offset: 2px;
}

.view-all:hover {
  color: var(--color-accent-hover);
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

.activity-icon.container { background: var(--color-info-bg); color: var(--color-info); }
.activity-icon.image { background: var(--color-accent); color: #fdfcfc; }
.activity-icon.network { background: var(--color-success-bg); color: var(--color-success); }
.activity-icon.volume { background: var(--color-warning-bg); color: var(--color-warning); }

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
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: border-color var(--transition-fast);
  color: var(--color-text-weak);
  font-size: var(--font-size-sm);
}

.quick-action-btn:hover {
  border-color: var(--color-text-strong);
  color: var(--color-text-strong);
  background: var(--color-info-bg);
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
