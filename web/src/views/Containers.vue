<template>
  <div class="containers-page">
    <div class="page-header">
      <h1>{{ t('containers.title') }}</h1>
      <div class="header-actions">
        <button @click="refreshContainers" class="btn btn-secondary" :disabled="loading">
          <RefreshCw :size="16" :class="{ 'spinning': loading }" />
          {{ t('containers.refresh') }}
        </button>
        <button @click="showCreateModal = true" class="btn btn-primary">
          <Plus :size="16" />
          {{ t('containers.create') }}
        </button>
      </div>
    </div>

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
            <span class="value">{{ container.ports || '-' }}</span>
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
  const statusMap: Record<string, string> = {
    running: t('containers.running'),
    stopped: t('containers.stopped'),
    paused: t('containers.paused')
  }
  return statusMap[status] || status
}

const getStatusClass = (status: string) => {
  const classMap: Record<string, string> = {
    running: 'badge-success',
    stopped: 'badge-error',
    paused: 'badge-warning'
  }
  return classMap[status] || ''
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
  } finally {
    loading.value = false
  }
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

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-6);
}

.page-header h1 {
  margin: 0;
  color: var(--color-text-strong);
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
}

.header-actions {
  display: flex;
  gap: var(--space-3);
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

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-1);
  padding: var(--space-2) var(--space-3);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  cursor: pointer;
  transition: all var(--transition-fast);
  font-family: var(--font-mono);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-sm {
  padding: var(--space-1) var(--space-2);
  font-size: var(--font-size-xs);
}

.btn-primary {
  background: var(--color-background-strong);
  color: var(--color-background);
  border-color: var(--color-background-strong);
}

.btn-primary:hover:not(:disabled) {
  background: var(--color-text-weak);
  border-color: var(--color-text-weak);
}

.btn-secondary {
  background: transparent;
  color: var(--color-text-strong);
  border-color: var(--color-border);
}

.btn-secondary:hover:not(:disabled) {
  background: var(--color-background-weak);
  border-color: var(--color-border);
}

.btn-ghost {
  background: transparent;
  color: var(--color-text);
  border-color: transparent;
}

.btn-ghost:hover:not(:disabled) {
  background: var(--color-background-weak);
  color: var(--color-text-strong);
}

.btn-danger-text {
  color: var(--color-error);
}

.btn-danger-text:hover:not(:disabled) {
  background: var(--color-error-bg);
  color: var(--color-error);
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
