<template>
  <div class="dashboard">
    <!-- Centerpiece — Stats (left) + 3D cube grid (right) -->
    <section class="centerpiece" aria-label="Machine connectivity visualization">
      <!-- Stats column — 4 metric cards stacked vertically -->
      <div class="stats-panel">
        <div class="stat-card stat-card--containers">
          <div class="stat-accent"></div>
          <div class="stat-body">
            <div class="stat-head">
              <Box :size="16" class="stat-icon" />
              <span class="stat-number">{{ stats.containers.total }}</span>
            </div>
            <span class="stat-label">{{ t('dashboard.totalContainers') }}</span>
            <span class="stat-sub">
              <span class="stat-sub--running">{{ stats.containers.running }} {{ t('dashboard.running') }}</span>
              <span class="stat-sub--stopped">{{ stats.containers.stopped }} {{ t('dashboard.stopped') }}</span>
            </span>
          </div>
        </div>

        <div class="stat-card stat-card--machines">
          <div class="stat-accent"></div>
          <div class="stat-body">
            <div class="stat-head">
              <Globe :size="16" class="stat-icon" />
              <span class="stat-number">{{ stats.machines.total }}</span>
            </div>
            <span class="stat-label">{{ t('nav.machines') }}</span>
          </div>
        </div>

        <div class="stat-card stat-card--networks">
          <div class="stat-accent"></div>
          <div class="stat-body">
            <div class="stat-head">
              <Network :size="16" class="stat-icon" />
              <span class="stat-number">{{ stats.networks.total }}</span>
            </div>
            <span class="stat-label">{{ t('dashboard.totalNetworks') }}</span>
            <span class="stat-sub">{{ stats.networks.active }} {{ t('dashboard.active') }}</span>
          </div>
        </div>

        <div class="stat-card stat-card--volumes">
          <div class="stat-accent"></div>
          <div class="stat-body">
            <div class="stat-head">
              <HardDrive :size="16" class="stat-icon" />
              <span class="stat-number">{{ stats.volumes.total }}</span>
            </div>
            <span class="stat-label">{{ t('dashboard.totalVolumes') }}</span>
            <span class="stat-sub">{{ stats.volumes.size }} {{ t('dashboard.totalSize') }}</span>
          </div>
        </div>
      </div>

      <!-- 3D cube visualization — right side -->
      <div class="centerpiece-visual">
        <div class="visual-header">
          <h2 class="visual-title">{{ t('dashboard.machineStatus') }}</h2>
          <p class="visual-description">{{ t('dashboard.machineStatusDesc') }}</p>
        </div>
        <CubeArray :cubes="cubeData" :max-cubes="16" />
      </div>
    </section>

    <!-- Resource Panel — CPU / Memory / Disk progress bars -->
    <section class="resource-panel" aria-label="System resource metrics">
      <div class="card">
        <div class="card-header">
          <h3>{{ t('dashboard.systemResources') }}</h3>
        </div>
        <div class="card-body">
          <div class="resource-grid">
            <MinimalProgress
              :label="t('dashboard.cpu')"
              :value="currentCpu"
              color="info"
              unit="%"
            />
            <MinimalProgress
              :label="t('dashboard.memory')"
              :value="currentMemory"
              color="warning"
              unit="%"
            />
            <MinimalProgress
              :label="t('dashboard.disk')"
              :value="currentDisk"
              color="success"
              unit="%"
            />
          </div>
        </div>
      </div>
    </section>

    <!-- Bottom Row — Quick Actions + System Status + Secondary Stats -->
    <section class="dashboard-bottom">
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
            <button class="quick-action-btn" @click="goToOrchestration">
              <Layers :size="18" />
              <span>{{ t('dashboard.composeProjects') }}</span>
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
              <span class="badge" :class="getServiceStatusClass(systemStatus.docker)">
                {{ getServiceStatusLabel(systemStatus.docker) }}
              </span>
            </div>
            <div class="status-item">
              <span class="status-label">{{ t('dashboard.apiService') }}</span>
              <span class="badge badge-success">{{ t('dashboard.online') }}</span>
            </div>
            <div class="status-item">
              <span class="status-label">{{ t('dashboard.database') }}</span>
              <span class="badge" :class="getServiceStatusClass(systemStatus.database)">
                {{ getServiceStatusLabel(systemStatus.database) }}
              </span>
            </div>
            <div class="status-item">
              <span class="status-label">{{ t('dashboard.messageQueue') }}</span>
              <span class="badge" :class="getServiceStatusClass(systemStatus.messageQueue)">
                {{ getServiceStatusLabel(systemStatus.messageQueue) }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <h3>{{ t('dashboard.stats.compose') }} &amp; {{ t('dashboard.totalImages') }}</h3>
        </div>
        <div class="card-body">
          <div class="secondary-stats">
            <div class="secondary-stat">
              <Image :size="16" class="secondary-stat-icon" />
              <span class="secondary-stat-number">{{ stats.images.total }}</span>
              <span class="secondary-stat-label">{{ t('dashboard.totalImages') }}</span>
              <span class="secondary-stat-sub">{{ stats.images.size }}</span>
            </div>
            <div class="secondary-stat">
              <Layers :size="16" class="secondary-stat-icon" />
              <span class="secondary-stat-number">{{ stats.compose.total }}</span>
              <span class="secondary-stat-label">{{ t('dashboard.stats.compose') }}</span>
              <span class="secondary-stat-sub">{{ stats.compose.running }} {{ t('dashboard.running') }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Recent Activity — kept subtle at the bottom -->
    <section class="dashboard-activity">
      <div class="card">
        <div class="card-header">
          <h3>{{ t('dashboard.recentActivity') }}</h3>
          <router-link to="/containers" class="view-all">{{ t('dashboard.viewAll') }}</router-link>
        </div>
        <div class="card-body">
          <div v-if="recentActivities.length > 0" class="activity-list">
            <div v-for="activity in recentActivities" :key="activity.id" class="activity-item">
              <div class="activity-dot" :class="'activity-dot--' + activity.type"></div>
              <div class="activity-content">
                <span class="activity-title">{{ activity.title }}</span>
                <span class="activity-time">{{ formatTime(activity.time) }}</span>
              </div>
            </div>
          </div>
          <p v-else class="activity-empty">{{ t('dashboard.noActivity') }}</p>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  Box,
  Network,
  HardDrive,
  Plus,
  Download,
  PlusCircle,
  Globe,
  Image,
  Layers
} from 'lucide-vue-next'
import { t } from '@/i18n'
import CubeArray from '@/components/dashboard/CubeArray.vue'
import type { CubeData } from '@/components/dashboard/CubeArray.vue'
import MinimalProgress from '@/components/dashboard/MinimalProgress.vue'
import api from '@/utils/api'
import { imageService } from '@/services/imageService'
import { networkService } from '@/services/networkService'
import { volumeService } from '@/services/volumeService'
import { composeService } from '@/services/composeService'
import { remoteMachineService } from '@/services/remoteMachineService'
import { formatSize } from '@/utils/format'
import type { RemoteMachine } from '@/types'
import { useChart } from '@/composables/useChart'
import { useWebSocket } from '@/composables/useWebSocket'
import { useAuthStore } from '@/stores/auth'
import { useRefreshBus } from '@/composables/useRefreshBus'

const stats = reactive({
  containers: { total: 0, running: 0, stopped: 0 },
  networks: { total: 0, active: 0 },
  volumes: { total: 0, size: '0 GB' },
  machines: { total: 0 },
  images: { total: 0, size: '0 GB' },
  compose: { total: 0, running: 0 }
})

const remoteMachines = ref<RemoteMachine[]>([])
const systemStatus = reactive({
  docker: 'unknown' as 'online' | 'offline' | 'unknown',
  database: 'unknown' as 'online' | 'offline' | 'unknown',
  messageQueue: 'unknown' as 'online' | 'offline' | 'unknown',
})

// Cube Section
const cubeData = computed<CubeData[]>(() => {
  const cubes: CubeData[] = []
  
  // Host machine — always first (hardcoded)
  cubes.push({ 
    id: 'local', 
    name: t('common.local'), 
    status: 'local' 
  })
  
  // Map remote machines to online/offline, skip local duplicates
  remoteMachines.value.forEach(m => {
    const isLocal = m.host === 'localhost' || m.host === '127.0.0.1' || m.docker_host === 'local'
    if (isLocal) return
    cubes.push({ 
      id: m.id, 
      name: m.name, 
      status: m.status === 'unknown' ? 'unknown' : (m.status === 'online' ? 'online' : 'offline')
    })
  })
  
  return cubes
})

const chart = useChart()

const currentCpu = computed<number>(() => {
  const arr = chart.cpu.value
  return arr.length > 0 ? arr[arr.length - 1]! : 0
})
const currentMemory = computed<number>(() => {
  const arr = chart.memory.value
  return arr.length > 0 ? arr[arr.length - 1]! : 0
})
const currentDisk = computed<number>(() => {
  const arr = chart.disk.value
  return arr.length > 0 ? arr[arr.length - 1]! : 0
})

const getWsUrl = () => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  // Append current JWT token if authenticated to pass through websocket auth check
  const authStore = useAuthStore()
  const tokenParam = authStore.token ? `?token=${encodeURIComponent(authStore.token)}` : ''
  // Use Vite proxy (same host as page) so we don't hardcode backend port
  return `${protocol}//${window.location.host}/v1/dashboard/resources/stream${tokenParam}`
}

const { connect: connectWs } = useWebSocket({
  url: getWsUrl(),
  onMessage: (data) => chart.addDataPoint(data as { cpu: number; memory: number; disk: number; time: string })
})

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

  try {
    const machines = await remoteMachineService.list()
    remoteMachines.value = machines
    const machineIds = Array.from(new Set(['local', ...machines.map((m) => m.id)]))

    const imageResults = await Promise.all(machineIds.map(async (machineId) => imageService.list(machineId)))
    const allImages = imageResults.flat()
    stats.images.total = allImages.length
    stats.images.size = formatSize(allImages.reduce((sum, img) => sum + (img.size || 0), 0))

    const networkResults = await Promise.all(machineIds.map(async (machineId) => networkService.listNetworks(machineId)))
    const allNetworks = networkResults.flat()
    stats.networks.total = allNetworks.length
    stats.networks.active = allNetworks.filter((network) => (network.containers?.length ?? 0) > 0).length

    const volumeResults = await Promise.all(machineIds.map(async (machineId) => volumeService.listVolumes(machineId)))
    const allVolumes = volumeResults.flat()
    stats.volumes.total = allVolumes.length
    stats.volumes.size = formatSize(allVolumes.reduce((sum, volume) => sum + (volume.size || 0), 0))

    let totalProjs = 0
    let runningProjs = 0
    await Promise.all(machineIds.map(async (machineId) => {
      const projs = await composeService.listProjects(machineId)
      totalProjs += projs.length
      runningProjs += projs.filter(p => p.status === 'running').length
    }))
    stats.compose.total = totalProjs
    stats.compose.running = runningProjs
  } catch (e) {
    console.warn('Failed to load compose stats:', e)
  }
}

const recentActivities = ref<Array<{ id: number; type: 'container' | 'network' | 'volume'; title: string; time: Date }>>([])

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

const router = useRouter()
const createContainer = () => router.push('/containers')
const pullImage = () => router.push('/images')
const createNetwork = () => router.push('/networks')
const createVolume = () => router.push('/volumes')
const goToOrchestration = () => router.push('/orchestration')

const getServiceStatusClass = (status: 'online' | 'offline' | 'unknown') => {
  if (status === 'online') return 'badge-success'
  if (status === 'offline') return 'badge-error'
  return 'badge-warning'
}

const getServiceStatusLabel = (status: 'online' | 'offline' | 'unknown') => {
  if (status === 'online') return t('dashboard.online')
  if (status === 'offline') return t('dashboard.offline')
  return t('common.unknown')
}

onMounted(async () => {
  const hasHistory = await chart.loadHistory(api.get.bind(api))
  if (!hasHistory) {
    chart.initLabels()
  }
  connectWs()
  refreshStats()
})

useRefreshBus(refreshStats)
</script>

<style scoped>
.dashboard {
  max-width: var(--container-max);
  margin: 0 auto;
}

/* Centerpiece — Stats (left) + 3D cube (right) */
.centerpiece {
  display: flex;
  align-items: center;
  gap: var(--space-8);
  min-height: 420px;
  margin-bottom: var(--space-8);
  padding: var(--space-6);
  background: radial-gradient(circle at center, var(--color-background-weak) 0%, transparent 70%);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-md);
}

.stats-panel {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
  flex: 1;
  min-width: 0;
  max-width: 240px;
}

.stat-card {
  position: relative;
  display: flex;
  background: var(--color-background-weak);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-sm);
  overflow: hidden;
  transition: border-color var(--transition-base);
}

.stat-card:hover {
  border-color: var(--color-text-weak);
}

.stat-accent {
  flex-shrink: 0;
  width: 3px;
  border-radius: var(--radius-sm) 0 0 var(--radius-sm);
}

.stat-card--containers .stat-accent { background: var(--color-info); }
.stat-card--machines   .stat-accent { background: var(--color-accent); }
.stat-card--networks   .stat-accent { background: var(--color-success); }
.stat-card--volumes    .stat-accent { background: var(--color-warning); }

.stat-body {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: var(--space-3) var(--space-4);
  min-width: 0;
  flex: 1;
}

.stat-head {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.stat-icon {
  color: var(--color-text-weaker);
  flex-shrink: 0;
}

.stat-number {
  font-family: var(--font-mono);
  font-size: var(--font-size-xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-strong);
  line-height: var(--line-height-tight);
}

.stat-label {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--color-text-weak);
  line-height: var(--line-height-tight);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.stat-sub {
  font-family: var(--font-mono);
  font-size: 11px;
  color: var(--color-text-weaker);
  line-height: var(--line-height-tight);
  display: flex;
  gap: var(--space-2);
  margin-top: 1px;
}

.stat-sub--running { color: var(--color-success); }
.stat-sub--stopped { color: var(--color-text-weaker); }

.centerpiece-visual {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  flex: 1.6;
  min-width: 0;
  gap: var(--space-4);
}

.visual-header {
  text-align: center;
  margin-top: var(--space-4);
}

.visual-title {
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-strong);
  text-transform: uppercase;
  letter-spacing: 0.1em;
  margin: 0 0 var(--space-1) 0;
}

.visual-description {
  font-family: var(--font-mono);
  font-size: 10px;
  color: var(--color-text-weaker);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin: 0;
}

.resource-panel {
  margin-bottom: var(--space-6);
}

.resource-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: var(--space-6);
}

.dashboard-bottom {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: var(--space-6);
  margin-bottom: var(--space-6);
}

.card {
  background: transparent;
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-sm);
  padding: 0;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--color-border-weak);
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

.view-all {
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  color: var(--color-accent);
  text-decoration: none;
  transition: color var(--transition-fast);
}

.view-all:hover {
  color: var(--color-accent-hover);
}

.secondary-stats {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.secondary-stat {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-3) 0;
}

.secondary-stat + .secondary-stat {
  border-top: 1px solid var(--color-border-weak);
}

.secondary-stat-icon {
  color: var(--color-text-weak);
  flex-shrink: 0;
}

.secondary-stat-number {
  font-family: var(--font-mono);
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-strong);
  line-height: var(--line-height-tight);
}

.secondary-stat-label {
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  color: var(--color-text-weak);
  line-height: var(--line-height-tight);
  margin-right: auto;
}

.secondary-stat-sub {
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  color: var(--color-text-weaker);
  line-height: var(--line-height-tight);
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
  justify-content: center;
  gap: var(--space-2);
  padding: var(--space-5) var(--space-4);
  min-height: 80px;
  background: var(--color-background-weak);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-sm);
  color: var(--color-text-weak);
  font-family: var(--font-mono);
  font-size: 10px;
  font-weight: var(--font-weight-bold);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  cursor: pointer;
  transition: all var(--transition-base);
}

.quick-action-btn:hover {
  border-color: var(--color-text-strong);
  color: var(--color-text-strong);
  background: var(--color-background-hover);
}

.quick-action-btn:active {
  border-color: var(--color-accent);
}

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
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
  color: var(--color-text-weak);
}

.dashboard-activity {
  margin-bottom: var(--space-6);
}

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
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
  color: var(--color-text-strong);
  line-height: var(--line-height-tight);
}

.activity-time {
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  color: var(--color-text-weaker);
}

@media (max-width: 1024px) {
  .centerpiece {
    min-height: 360px;
    gap: var(--space-6);
  }

  .stats-panel {
    max-width: 200px;
  }

  .dashboard-bottom {
    grid-template-columns: 1fr 1fr;
  }

  .resource-grid {
    gap: var(--space-4);
  }
}

@media (max-width: 767px) {
  .centerpiece {
    flex-direction: column;
    min-height: auto;
    padding: var(--space-4);
    gap: var(--space-6);
    background: var(--color-background-weak);
  }

  .centerpiece-visual {
    order: -1;
    width: 100%;
  }

  .stats-panel {
    max-width: none;
    width: 100%;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-2);
  }

  .stat-card {
    border-radius: var(--radius-sm);
  }

  .stat-body {
    padding: var(--space-2) var(--space-3);
    gap: 1px;
  }

  .stat-number {
    font-size: var(--font-size-lg);
  }

  .stat-label {
    font-size: 10px;
  }

  .stat-sub {
    font-size: 10px;
  }

  .resource-grid {
    grid-template-columns: 1fr;
    gap: var(--space-4);
  }

  .dashboard-bottom {
    grid-template-columns: 1fr;
    gap: var(--space-4);
  }

  .resource-panel {
    margin-bottom: var(--space-4);
  }

  .dashboard-activity {
    margin-bottom: var(--space-4);
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

@media (max-width: 400px) {
  .centerpiece {
    padding: var(--space-3);
    gap: var(--space-4);
  }

  .stats-panel {
    grid-template-columns: 1fr;
    gap: var(--space-2);
  }

  .stat-number {
    font-size: var(--font-size-base);
  }

  .quick-actions {
    grid-template-columns: 1fr;
  }
}
</style>
