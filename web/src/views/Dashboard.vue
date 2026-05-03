<template>
  <div class="dashboard">
    <PageHeader :title="t('dashboard.title')" />

    <!-- Stats Grid -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-accent stat-accent--info"></div>
        <div class="stat-icon">
          <Box :size="24" />
        </div>
        <div class="stat-body">
          <span class="stat-number">{{ stats.containers.total }}</span>
          <span class="stat-label">{{ t('dashboard.totalContainers') }}</span>
          <span class="stat-sub">
            <span class="stat-sub--running">{{ stats.containers.running }} {{ t('dashboard.running') }}</span>
            <span class="stat-sub--stopped">{{ stats.containers.stopped }} {{ t('dashboard.stopped') }}</span>
          </span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-accent stat-accent--accent"></div>
        <div class="stat-icon">
          <Globe :size="24" />
        </div>
        <div class="stat-body">
          <span class="stat-number">{{ stats.machines.total }}</span>
          <span class="stat-label">{{ t('nav.machines') }}</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-accent stat-accent--success"></div>
        <div class="stat-icon">
          <Network :size="24" />
        </div>
        <div class="stat-body">
          <span class="stat-number">{{ stats.networks.total }}</span>
          <span class="stat-label">{{ t('dashboard.totalNetworks') }}</span>
          <span class="stat-sub">{{ stats.networks.active }} {{ t('dashboard.active') }}</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-accent stat-accent--warning"></div>
        <div class="stat-icon">
          <HardDrive :size="24" />
        </div>
        <div class="stat-body">
          <span class="stat-number">{{ stats.volumes.total }}</span>
          <span class="stat-label">{{ t('dashboard.totalVolumes') }}</span>
          <span class="stat-sub">{{ stats.volumes.size }} {{ t('dashboard.totalSize') }}</span>
        </div>
      </div>
    </div>

    <!-- Main Content Grid: Chart as centerpiece -->
    <div class="dashboard-content">
      <!-- Left: Main area (chart + activity) -->
      <div class="dashboard-main">
        <div class="card card--chart">
          <div class="card-header">
            <h3>{{ t('dashboard.systemResources') }}</h3>
          </div>
          <div class="card-body">
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
          <div class="card-body">
            <div class="activity-list">
              <div v-for="activity in recentActivities" :key="activity.id" class="activity-item">
                <div class="activity-dot" :class="'activity-dot--' + activity.type"></div>
                <div class="activity-content">
                  <span class="activity-title">{{ activity.title }}</span>
                  <span class="activity-time">{{ formatTime(activity.time) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Right: Side panel -->
      <div class="dashboard-side">
        <div class="card">
          <div class="card-header">
            <h3>{{ t('dashboard.quickActions') }}</h3>
          </div>
          <div class="card-body">
            <div class="quick-actions">
              <button class="quick-action-btn" @click="createContainer">
                <Plus :size="18" />
                <span>{{ t('dashboard.createContainer') }}</span>
              </button>
              <button class="quick-action-btn" @click="pullImage">
                <Download :size="18" />
                <span>{{ t('dashboard.pullImage') }}</span>
              </button>
              <button class="quick-action-btn" @click="createNetwork">
                <PlusCircle :size="18" />
                <span>{{ t('dashboard.createNetwork') }}</span>
              </button>
              <button class="quick-action-btn" @click="createVolume">
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
          <div class="card-body">
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
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import {
  Box,
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
    const data: any = await api.get('/dashboard/resources/history?minutes=5')
    if (data && data.length > 0) {
      const history = data.slice(-TOTAL_POINTS)
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
    const data: any = await api.get('/dashboard/stats')
    if (data) {
      const { machines, containers } = data
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
  { id: 2, type: 'container', title: t('dashboard.imagePulled', { name: 'nginx:latest' }), time: new Date(Date.now() - 15 * 60 * 1000) },
  { id: 3, type: 'container', title: t('dashboard.containerStopped', { name: 'database' }), time: new Date(Date.now() - 30 * 60 * 1000) },
  { id: 4, type: 'network', title: t('dashboard.networkCreated', { name: 'frontend-network' }), time: new Date(Date.now() - 45 * 60 * 1000) },
  { id: 5, type: 'volume', title: t('dashboard.volumeDeleted', { name: 'data-volume' }), time: new Date(Date.now() - 60 * 60 * 1000) }
])

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
/* ============================================
   Dashboard — Flat, monospace-first, warm dark
   ============================================ */

.dashboard {
  max-width: var(--container-max);
  margin: 0 auto;
}

/* ------------------------------------------
   Stats Grid — 4-column desktop, 2-column mobile
   ------------------------------------------ */

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-4);
  margin-bottom: var(--space-6);
}

/* Stat Card — flat, border-only, accent strip on top */
.stat-card {
  position: relative;
  display: flex;
  align-items: flex-start;
  gap: var(--space-4);
  padding: var(--space-5);
  background: var(--color-background-weak);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  overflow: hidden;
  transition: border-color var(--transition-fast);
}

.stat-card:hover {
  border-color: var(--color-text-weaker);
}

/* Accent strip — thin colored bar at top of stat card */
.stat-accent {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
}

.stat-accent--info    { background: var(--color-info); }
.stat-accent--accent  { background: var(--color-accent); }
.stat-accent--success { background: var(--color-success); }
.stat-accent--warning { background: var(--color-warning); }

.stat-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  color: var(--color-text-weak);
  margin-top: var(--space-1);
}

.stat-body {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  min-width: 0;
}

.stat-number {
  font-family: var(--font-mono);
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-strong);
  line-height: var(--line-height-tight);
}

.stat-label {
  font-size: var(--font-size-sm);
  color: var(--color-text-weak);
  line-height: var(--line-height-tight);
}

.stat-sub {
  display: flex;
  gap: var(--space-3);
  font-size: var(--font-size-xs);
  color: var(--color-text-weaker);
  line-height: var(--line-height-tight);
}

.stat-sub--running { color: var(--color-success); }
.stat-sub--stopped { color: var(--color-text-weaker); }

/* ------------------------------------------
   Dashboard Content — chart-centered layout
   ------------------------------------------ */

.dashboard-content {
  display: grid;
  grid-template-columns: 2fr 1fr;
  gap: var(--space-6);
  align-items: start;
}

.dashboard-main {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

.dashboard-side {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}

/* ------------------------------------------
   Card — flat scoped override
   Global .card has padding; we use internal sub-layout
   ------------------------------------------ */

.card {
  background: var(--color-background-weak);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: 0;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-4) var(--space-5);
  border-bottom: 1px solid var(--color-border);
}

.card-header h3 {
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
  letter-spacing: 0.02em;
  text-transform: uppercase;
  margin: 0;
}

.card-body {
  padding: var(--space-5);
}

/* Chart card — tighter padding for chart */
.card--chart .card-body {
  padding: var(--space-4) var(--space-3) var(--space-3);
}

/* ------------------------------------------
   View All link
   ------------------------------------------ */

.view-all {
  font-size: var(--font-size-xs);
  color: var(--color-accent);
  text-decoration: none;
  transition: color var(--transition-fast), opacity var(--transition-fast);
}

.view-all:hover {
  color: var(--color-accent-hover);
}

/* ------------------------------------------
   Activity List — minimal dots, no icon boxes
   ------------------------------------------ */

.activity-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.activity-item {
  display: flex;
  align-items: flex-start;
  gap: var(--space-3);
  padding: var(--space-3) 0;
}

.activity-item + .activity-item {
  border-top: 1px solid var(--color-border-weak);
}

.activity-dot {
  width: 8px;
  height: 8px;
  border-radius: var(--radius-full);
  flex-shrink: 0;
  margin-top: 6px;
}

.activity-dot--container { background: var(--color-info); }
.activity-dot--network   { background: var(--color-success); }
.activity-dot--volume    { background: var(--color-warning); }

.activity-content {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
  min-width: 0;
}

.activity-title {
  font-size: var(--font-size-sm);
  color: var(--color-text-strong);
  line-height: var(--line-height-tight);
}

.activity-time {
  font-size: var(--font-size-xs);
  color: var(--color-text-weaker);
}

/* ------------------------------------------
   Quick Actions
   ------------------------------------------ */

.quick-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-3);
}

.quick-action-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-2);
  padding: var(--space-5) var(--space-4);
  min-height: 72px;
  background: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  color: var(--color-text-weak);
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  cursor: pointer;
  transition: border-color var(--transition-fast), color var(--transition-fast), background var(--transition-fast);
}

.quick-action-btn:hover {
  border-color: var(--color-text-strong);
  color: var(--color-text-strong);
  background: var(--color-info-bg);
}

.quick-action-btn:active {
  border-color: var(--color-accent);
}

/* ------------------------------------------
   System Status
   ------------------------------------------ */

.status-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.status-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-3) 0;
}

.status-item + .status-item {
  border-top: 1px solid var(--color-border-weak);
}

.status-label {
  font-size: var(--font-size-sm);
  color: var(--color-text-weak);
}

/* ------------------------------------------
   Responsive: Tablet (≤1024px)
   ------------------------------------------ */

@media (max-width: 1024px) {
  .dashboard-content {
    grid-template-columns: 1fr;
  }

  /* Side-by-side layout for side cards on tablet */
  .dashboard-side {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-4);
  }
}

/* ------------------------------------------
   Responsive: Mobile (≤767px)
   ------------------------------------------ */

@media (max-width: 767px) {
  /* 2-column stat grid */
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
gap: var(--space-2);
  }

  .stat-card {
    padding: var(--space-4);
    flex-direction: column;
    gap: var(--space-2);
  }

  .stat-icon {
    width: 32px;
    height: 32px;
    margin-top: 0;
    color: var(--color-text-weaker);
  }

  .stat-icon :deep(svg) {
    width: 20px;
    height: 20px;
  }

  .stat-number {
    font-size: var(--font-size-xl);
  }

  .stat-label {
    font-size: var(--font-size-xs);
  }

  .stat-sub {
    flex-wrap: wrap;
    font-size: 11px;
  }

  .dashboard-content {
    grid-template-columns: 1fr;
    gap: var(--space-4);
  }

  .dashboard-main {
    gap: var(--space-4);
  }

  .dashboard-side {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .card-header {
    padding: var(--space-3) var(--space-4);
  }

  .card-header h3 {
    font-size: var(--font-size-xs);
  }

  .card-body {
    padding: var(--space-4);
  }

  .card--chart .card-body {
    padding: var(--space-3) var(--space-2) var(--space-2);
  }

  /* Horizontal quick actions for better touch targets */
  .quick-actions {
    grid-template-columns: 1fr 1fr;
    gap: var(--space-2);
  }

  .quick-action-btn {
    flex-direction: row;
    justify-content: flex-start;
    padding: var(--space-3) var(--space-4);
    min-height: var(--space-12);
    gap: var(--space-3);
    font-size: var(--font-size-sm);
  }

  .status-item {
    padding: var(--space-2) 0;
  }

  .status-label {
    font-size: var(--font-size-xs);
  }

  .activity-item {
    padding: var(--space-2) 0;
  }

  .activity-title {
    font-size: var(--font-size-xs);
  }

  .view-all {
    font-size: var(--font-size-xs);
  }
}

/* ------------------------------------------
   Responsive: Small mobile (≤400px)
   ------------------------------------------ */

@media (max-width: 400px) {
  .stats-grid {
    grid-template-columns: 1fr;
    gap: var(--space-3);
  }

  .stat-card {
    flex-direction: row;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-4);
  }

  .stat-body {
    gap: var(--space-1);
  }

  .stat-sub {
    display: none;
  }
}
</style>
