<template>
  <div class="machine-detail-page">
    <div class="page-header">
      <div class="header-left">
        <button @click="goBack" class="btn btn-ghost">
          <ArrowLeft :size="16" />
        </button>
        <h1>{{ machine?.name || t('machines.title') }}</h1>
        <div v-if="machine" class="badge" :class="getConnectionStatusClass(connectionStatus)">
          {{ getConnectionStatusText(connectionStatus) }}
        </div>
      </div>
      <button @click="refreshAll" class="btn btn-secondary" :disabled="loading">
        <RefreshCw :size="16" :class="{ 'spinning': loading }" />
        {{ t('machines.refresh') }}
      </button>
    </div>

    <div v-if="machine" class="machine-info-bar">
      <template v-if="machine.id !== 'local'">
        <span class="info-chip">{{ machine.host }}:{{ machine.port }}</span>
        <span class="info-chip">{{ machine.username }}</span>
      </template>
      <template v-else>
        <span class="info-chip badge badge-primary">本地连接</span>
      </template>
      <span class="info-chip mono">{{ machine.docker_host }}</span>
    </div>

    <CollapsibleFilters
      v-model="searchQuery"
      :search-placeholder="t('containers.searchPlaceholder')"
      search-label="Search"
      filter-label="Filters"
      :has-filters="true"
    >
      <template #filters>
        <select v-model="statusFilter" class="input">
          <option value="">{{ t('containers.allStatuses') }}</option>
          <option value="running">{{ t('containers.running') }}</option>
          <option value="stopped">{{ t('containers.stopped') }}</option>
          <option value="paused">{{ t('containers.paused') }}</option>
        </select>
      </template>
    </CollapsibleFilters>

    <div class="containers-grid">
      <div
        v-for="container in filteredContainers"
        :key="container.id"
        class="container-card"
        :class="{ 'status-running': container.status === 'running' }"
      >
        <div class="container-header">
          <div class="container-name">{{ container.name }}</div>
          <div class="badge" :class="getContainerStatusClass(container.status)">
            {{ container.status }}
          </div>
        </div>

        <div class="container-info">
          <div class="info-item">
            <span class="label">{{ t('containers.image') }}</span>
            <span class="value mono">{{ container.image }}</span>
          </div>
          <div class="info-item">
            <span class="label">ID</span>
            <span class="value mono">{{ container.id.substring(0, 12) }}</span>
          </div>
          <div v-if="container.ports && container.ports.length" class="info-item">
            <span class="label">{{ t('containers.ports') }}</span>
            <span class="value mono">{{ container.ports.join(', ') }}</span>
          </div>
        </div>

        <div class="container-actions">
          <button
            v-if="container.status !== 'running'"
            @click="startContainer(container.id)"
            class="btn btn-sm btn-secondary"
          >
            <Play :size="14" />
            {{ t('containers.start') }}
          </button>
          <button
            v-if="container.status === 'running'"
            @click="stopContainer(container.id)"
            class="btn btn-sm btn-secondary"
          >
            <Square :size="14" />
            {{ t('containers.stop') }}
          </button>
          <button
            v-if="container.status === 'running'"
            @click="restartContainer(container.id)"
            class="btn btn-sm btn-secondary"
          >
            <RotateCcw :size="14" />
            {{ t('containers.restart') }}
          </button>
          <button
            @click="selectContainerForLogs(container)"
            class="btn btn-sm"
            :class="selectedContainerId === container.id ? 'btn-primary' : 'btn-secondary'"
          >
            <FileText :size="14" />
            {{ t('containers.logs') }}
          </button>
          <button
            @click="selectContainerForExec(container)"
            class="btn btn-sm btn-secondary"
          >
            <Terminal :size="14" />
            {{ t('machines.exec') }}
          </button>
          <button @click="removeContainer(container.id)" class="btn btn-sm btn-ghost btn-danger-text">
            <Trash2 :size="14" />
            {{ t('containers.delete') }}
          </button>
        </div>
      </div>

      <div v-if="filteredContainers.length === 0 && !loading" class="empty-state">
        <Box :size="48" class="empty-icon" />
        <p>{{ t('machines.noContainers') }}</p>
      </div>
    </div>

    <div class="bottom-panels">
      <div class="panel logs-panel">
        <div class="panel-header">
          <h3>{{ t('containers.logs') }}</h3>
          <div class="panel-controls">
            <select v-model="logsTail" class="input input-sm" @change="loadLogs">
              <option value="50">50</option>
              <option value="100">100</option>
              <option value="200">200</option>
              <option value="500">500</option>
              <option value="1000">1000</option>
            </select>
            <label class="auto-refresh">
              <input type="checkbox" v-model="autoRefreshLogs" />
              Auto
            </label>
            <button @click="loadLogs" class="btn btn-sm btn-secondary">
              <RefreshCw :size="12" />
            </button>
          </div>
        </div>
        <div class="logs-display">
          <div v-if="logsLoading" class="panel-loading">{{ t('machines.loadingLogs') }}</div>
          <pre v-else-if="logs">{{ logs }}</pre>
          <div v-else class="panel-empty">{{ t('machines.noLogs') }}</div>
        </div>
      </div>

      <div class="panel exec-panel">
        <div class="panel-header">
          <h3>{{ t('machines.exec') }}</h3>
        </div>
        <div class="exec-form">
          <input
            v-model="execCmd"
            type="text"
            class="input"
            :placeholder="t('machines.execPlaceholder')"
            @keyup.enter="runExec"
          />
          <button @click="runExec" class="btn btn-primary" :disabled="!selectedContainerIdForExec || execLoading">
            {{ t('machines.runExec') }}
          </button>
        </div>
        <div class="exec-output">
          <div v-if="execLoading" class="panel-loading">...</div>
          <pre v-else-if="execOutput">{{ execOutput }}</pre>
          <div v-else class="panel-empty">{{ t('machines.execOutput') }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  RefreshCw,
  ArrowLeft,
  Play,
  Square,
  RotateCcw,
  FileText,
  Trash2,
  Box,
  Terminal
} from 'lucide-vue-next'
import { t } from '@/i18n'
import { remoteMachineService } from '@/services/remoteMachineService'
import CollapsibleFilters from '@/components/ui/CollapsibleFilters.vue'
import type { RemoteMachine, RemoteContainer } from '@/types'

const router = useRouter()
const route = useRoute()

const machineId = route.params.id as string

const loading = ref(false)
const connectionStatus = ref<'unknown' | 'testing' | 'online' | 'offline'>('unknown')
const logsLoading = ref(false)
const execLoading = ref(false)
const searchQuery = ref('')
const statusFilter = ref('')
const machine = ref<RemoteMachine | null>(null)
const containers = ref<RemoteContainer[]>([])
const selectedContainerId = ref<string | null>(null)
const selectedContainerIdForExec = ref<string | null>(null)
const logs = ref('')
const logsTail = ref('100')
const autoRefreshLogs = ref(false)
const execCmd = ref('')
const execOutput = ref('')

let logsInterval: ReturnType<typeof setInterval> | null = null

const filteredContainers = computed(() => {
  let filtered = containers.value
  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    filtered = filtered.filter(c =>
      c.name.toLowerCase().includes(q) ||
      c.image.toLowerCase().includes(q) ||
      c.id.toLowerCase().includes(q)
    )
  }
  if (statusFilter.value) {
    filtered = filtered.filter(c => c.status === statusFilter.value)
  }
  return filtered
})

const getContainerStatusClass = (status: string) => {
  const classMap: Record<string, string> = {
    running: 'badge-success',
    stopped: 'badge-error',
    paused: 'badge-warning',
    created: 'badge-info',
    restarting: 'badge-warning',
    exited: 'badge-error'
  }
  return classMap[status] || ''
}

const goBack = () => {
  router.push('/machines')
}

const refreshAll = async () => {
  loading.value = true
  connectionStatus.value = 'testing'
  try {
    const [m, c] = await Promise.all([
      remoteMachineService.get(machineId),
      remoteMachineService.listContainers(machineId)
    ])
    machine.value = m
    containers.value = c

    // Test connection and update status
    try {
      await remoteMachineService.testConnection(machineId)
      connectionStatus.value = 'online'
      machine.value = { ...m, status: 'online' }
    } catch {
      connectionStatus.value = 'offline'
      machine.value = { ...m, status: 'offline' }
    }
  } catch (e) {
    console.error('Failed to refresh:', e)
    connectionStatus.value = 'offline'
  } finally {
    loading.value = false
  }
}

const getConnectionStatusClass = (status: string) => {
  const classMap: Record<string, string> = {
    online: 'badge-success',
    offline: 'badge-error',
    testing: 'badge-info',
    unknown: 'badge-warning'
  }
  return classMap[status] || ''
}

const getConnectionStatusText = (status: string) => {
  const textMap: Record<string, string> = {
    online: t('machines.status.online'),
    offline: t('machines.status.offline'),
    testing: t('machines.testing'),
    unknown: t('machines.status.unknown')
  }
  return textMap[status] || status
}

const selectContainerForLogs = async (container: RemoteContainer) => {
  selectedContainerId.value = container.id
  selectedContainerIdForExec.value = container.id
  await loadLogs()
}

const selectContainerForExec = (container: RemoteContainer) => {
  selectedContainerIdForExec.value = container.id
  execCmd.value = ''
  execOutput.value = ''
}

const loadLogs = async () => {
  if (!selectedContainerId.value) return
  logsLoading.value = true
  try {
    logs.value = await remoteMachineService.getContainerLogs(machineId, selectedContainerId.value, logsTail.value)
  } catch (e) {
    logs.value = `Error loading logs: ${e}`
  } finally {
    logsLoading.value = false
  }
}

const startContainer = async (id: string) => {
  try {
    await remoteMachineService.startContainer(machineId, id)
    await refreshAll()
  } catch (e) {
    console.error('Failed to start container:', e)
  }
}

const stopContainer = async (id: string) => {
  try {
    await remoteMachineService.stopContainer(machineId, id)
    await refreshAll()
  } catch (e) {
    console.error('Failed to stop container:', e)
  }
}

const restartContainer = async (id: string) => {
  try {
    await remoteMachineService.restartContainer(machineId, id)
    await refreshAll()
  } catch (e) {
    console.error('Failed to restart container:', e)
  }
}

const removeContainer = async (id: string) => {
  if (!confirm(t('containers.confirmDelete'))) return
  try {
    await remoteMachineService.removeContainer(machineId, id)
    await refreshAll()
  } catch (e) {
    console.error('Failed to remove container:', e)
  }
}

const runExec = async () => {
  if (!selectedContainerIdForExec.value || !execCmd.value) return
  execLoading.value = true
  execOutput.value = ''
  try {
    const cmd = execCmd.value.trim().split(/\s+/)
    execOutput.value = await remoteMachineService.execContainer(machineId, selectedContainerIdForExec.value, cmd)
  } catch (e: any) {
    execOutput.value = `Error: ${e?.response?.data?.error || e?.message || e}`
  } finally {
    execLoading.value = false
  }
}

watch(autoRefreshLogs, (val) => {
  if (val && selectedContainerId.value) {
    logsInterval = setInterval(loadLogs, 5000)
  } else if (logsInterval) {
    clearInterval(logsInterval)
    logsInterval = null
  }
})

onMounted(() => {
  refreshAll()
})

onUnmounted(() => {
  if (logsInterval) {
    clearInterval(logsInterval)
  }
})
</script>

<style scoped>
.machine-detail-page {
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-4);
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}

.page-header h1 {
  margin: 0;
  color: var(--color-text-strong);
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
}

.machine-info-bar {
  display: flex;
  gap: var(--space-2);
  margin-bottom: var(--space-6);
  flex-wrap: wrap;
}

.info-chip {
  padding: var(--space-1) var(--space-3);
  background: var(--color-background-weak);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-sm);
  font-size: var(--font-size-xs);
  color: var(--color-text);
}

.info-chip.mono {
  font-family: var(--font-mono);
}

.containers-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(380px, 100%), 1fr));
  gap: var(--space-4);
  margin-bottom: var(--space-6);
}

.container-card {
  background: var(--color-background);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-sm);
  padding: var(--space-4);
  border-left: 3px solid var(--color-border);
  transition: border-color var(--transition-fast);
}

.container-card.status-running {
  border-left-color: var(--color-success);
}

.container-card:hover {
  border-color: var(--color-text-weaker);
}

.container-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-4);
}

.container-name {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
}

.container-info {
  margin-bottom: var(--space-4);
}

.info-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: var(--space-2);
  font-size: var(--font-size-sm);
}

.info-item .label {
  color: var(--color-text);
}

.info-item .value {
  color: var(--color-text-strong);
  font-weight: var(--font-weight-medium);
}

.info-item .value.mono,
.container-name.mono {
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
}

.container-actions {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
  padding-top: var(--space-3);
  border-top: 1px solid var(--color-border-weak);
}

.bottom-panels {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-4);
}

.panel {
  background: var(--color-background);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-md);
  display: flex;
  flex-direction: column;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--color-border-weak);
}

.panel-header h3 {
  margin: 0;
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
}

.panel-controls {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.panel-controls .input-sm {
  padding: var(--space-1) var(--space-2);
  font-size: var(--font-size-xs);
  width: 70px;
}

.auto-refresh {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  font-size: var(--font-size-xs);
  color: var(--color-text);
  cursor: pointer;
}

.logs-display {
  flex: 1;
  min-height: 300px;
  max-height: 400px;
  overflow: auto;
  background: var(--color-terminal-bg);
  padding: var(--space-3);
}

.logs-display pre {
  margin: 0;
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  color: var(--color-terminal-text);
  white-space: pre-wrap;
  word-break: break-all;
}

.exec-form {
  display: flex;
  gap: var(--space-2);
  padding: var(--space-3) var(--space-4);
  border-bottom: 1px solid var(--color-border-weak);
}

.exec-form .input {
  flex: 1;
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
}

.exec-output {
  flex: 1;
  min-height: 150px;
  max-height: 200px;
  overflow: auto;
  background: var(--color-terminal-bg);
  padding: var(--space-3);
}

.exec-output pre {
  margin: 0;
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  color: var(--color-terminal-text);
  white-space: pre-wrap;
  word-break: break-all;
}

.panel-loading,
.panel-empty {
  padding: var(--space-4);
  text-align: center;
  color: var(--color-text-weak);
  font-size: var(--font-size-sm);
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--space-16);
  color: var(--color-text-weak);
  grid-column: 1 / -1;
  gap: var(--space-4);
}

.empty-icon {
  opacity: 0.4;
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 1024px) {
  .bottom-panels {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .containers-grid {
    grid-template-columns: 1fr;
  }

  .container-actions {
    justify-content: center;
  }
}
</style>
