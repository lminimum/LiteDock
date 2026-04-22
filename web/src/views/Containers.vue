<template>
  <div class="containers-page">
    <PageHeader :title="t('containers.title')">
      <button @click="refreshContainers" class="btn btn-secondary" :disabled="loading">
        <RefreshCw :size="16" :class="{ 'spinning': loading }" />
        {{ t('containers.refresh') }}
      </button>
      <button @click="showCreateModal = true" class="btn btn-primary">
        <Plus :size="16" />
        {{ t('containers.create') }}
      </button>
    </PageHeader>

    <div class="filters">
      <div class="search-box">
        <input
          v-model="searchQuery"
          :placeholder="t('containers.searchPlaceholder')"
          type="text"
          class="input"
        />
      </div>
      <div class="filter-options">
        <select v-model="statusFilter" class="input">
          <option value="">{{ t('containers.allStatuses') }}</option>
          <option value="running">{{ t('containers.running') }}</option>
          <option value="stopped">{{ t('containers.stopped') }}</option>
          <option value="paused">{{ t('containers.paused') }}</option>
          <option value="restarting">{{ t('containers.restarting') }}</option>
          <option value="exited">{{ t('containers.exited') }}</option>
          <option value="created">{{ t('containers.created') }}</option>
        </select>
      </div>
    </div>

    <div class="containers-grid">
      <div
        v-for="container in filteredContainers"
        :key="container.id"
        class="container-card"
        :class="{ 'status-running': container.status === 'running' }"
      >
        <div class="container-header">
          <div class="container-name">{{ container.name }}</div>
          <div class="badge" :class="getStatusClass(container.status)">
            {{ getStatusText(container.status) }}
          </div>
        </div>

        <div class="container-info">
          <div class="info-item">
            <span class="label">{{ t('containers.image') }}</span>
            <span class="value">{{ container.image }}</span>
          </div>
          <div class="info-item">
            <span class="label">{{ t('containers.ports') }}</span>
            <span class="value">{{ (container.ports || []).join(', ') || '-' }}</span>
          </div>
          <div v-if="container.machine" class="info-item">
            <span class="label">{{ t('containers.machine') }}</span>
            <span class="value">{{ container.machine }}</span>
          </div>
          <div class="info-item">
            <span class="label">{{ t('containers.createdAt') }}</span>
            <span class="value">{{ formatDate(container.createdAt) }}</span>
          </div>
        </div>

        <div class="container-actions">
          <button
            v-if="container.status === 'stopped'"
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
          <button @click="showLogs(container.id)" class="btn btn-sm btn-ghost">
            <FileText :size="14" />
            {{ t('containers.logs') }}
          </button>
          <button @click="deleteContainer(container.id)" class="btn btn-sm btn-ghost btn-danger-text">
            <Trash2 :size="14" />
            {{ t('containers.delete') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  RefreshCw,
  Plus,
  Play,
  Square,
  RotateCcw,
  FileText,
  Trash2
} from 'lucide-vue-next'
import { t } from '@/i18n'
import { remoteMachineService } from '@/services/remoteMachineService'
import type { RemoteMachine } from '@/types'
import PageHeader from '@/components/ui/PageHeader.vue'

interface Container {
  id: string
  name: string
  image: string
  status: 'running' | 'stopped' | 'paused' | 'restarting' | 'exited' | 'created'
  ports: string[]
  createdAt: string
  machine: string
  machineId: string
}

const loading = ref(false)
const searchQuery = ref('')
const statusFilter = ref('')
const showCreateModal = ref(false)
const machines = ref<RemoteMachine[]>([])

const containers = ref<Container[]>([])

const refreshContainers = async () => {
  loading.value = true
  try {
    const allMachines = await remoteMachineService.list()
    machines.value = allMachines

    const allContainers: Container[] = []
    await Promise.all(
      allMachines.map(async (m) => {
        try {
          const remoteContainers = await remoteMachineService.listContainers(m.id)
          for (const c of remoteContainers) {
            allContainers.push({ ...c, machine: m.name, machineId: m.id })
          }
        } catch (e) {
          // machine offline or unreachable
        }
      })
    )

    // Sort by machine name, then by container name
    allContainers.sort((a, b) => {
      if (a.machine !== b.machine) return a.machine.localeCompare(b.machine)
      return a.name.localeCompare(b.name)
    })

    containers.value = allContainers
  } catch (e) {
    console.error('Failed to load containers:', e)
  } finally {
    loading.value = false
  }
}

const filteredContainers = computed(() => {
  let filtered = containers.value

  if (searchQuery.value) {
    filtered = filtered.filter(container =>
      container.name.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      container.image.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      container.machine.toLowerCase().includes(searchQuery.value.toLowerCase())
    )
  }

  if (statusFilter.value) {
    filtered = filtered.filter(container => container.status === statusFilter.value)
  }

  return filtered
})

const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    running: t('containers.running'),
    stopped: t('containers.stopped'),
    paused: t('containers.paused'),
    restarting: t('containers.restarting'),
    exited: t('containers.exited'),
    created: t('containers.created')
  }
  return statusMap[status] || status
}

const getStatusClass = (status: string) => {
  const classMap: Record<string, string> = {
    running: 'badge-success',
    stopped: 'badge-error',
    paused: 'badge-warning',
    restarting: 'badge-warning',
    exited: 'badge-error',
    created: 'badge-info'
  }
  return classMap[status] || ''
}

const formatDate = (date: string) => {
  if (!date) return '-'
  const d = new Date(date)
  if (isNaN(d.getTime())) return '-'
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(d)
}

const startContainer = async (id: string) => {
  const container = containers.value.find(c => c.id === id)
  if (container) container.status = 'running'
}

const stopContainer = async (id: string) => {
  const container = containers.value.find(c => c.id === id)
  if (container) container.status = 'stopped'
}

const restartContainer = async (id: string) => {
  console.log('restart container:', id)
}

const showLogs = (id: string) => {
  console.log('show logs:', id)
}

const deleteContainer = async (id: string) => {
  if (confirm(t('containers.confirmDelete'))) {
    containers.value = containers.value.filter(c => c.id !== id)
  }
}

onMounted(() => refreshContainers())
</script>

<style scoped>
.containers-page {
  max-width: 1400px;
  margin: 0 auto;
}

.filters {
  display: flex;
  gap: var(--space-4);
  margin-bottom: var(--space-6);
  padding: var(--space-4);
  background: var(--color-background);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-md);
}

.search-box {
  flex: 1;
}

.filter-options select {
  min-width: 140px;
}

.containers-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: var(--space-4);
}

.container-card {
  background: var(--color-background);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-md);
  padding: var(--space-4);
  border-left: 3px solid var(--color-border);
  transition: box-shadow var(--transition-fast), border-color var(--transition-fast);
}

.container-card.status-running {
  border-left-color: var(--color-success);
}

.container-card:hover {
  box-shadow: var(--shadow-md);
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

.container-actions {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
  padding-top: var(--space-3);
  border-top: 1px solid var(--color-border-weak);
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
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
