<template>
  <div class="task-detail-page">
    <PageHeader :title="t('tasks.detailTitle')">
      <template #left>
        <button @click="router.back()" class="btn btn-ghost btn-sm mr-2">
          <ArrowLeft :size="16" />
        </button>
      </template>
      <div class="header-actions">
        <div v-if="task" class="badge mr-3" :class="getStatusBadgeClass(task.status)">
          {{ t(`tasks.status.${task.status}`) }}
        </div>
        <button @click="fetchTask" class="btn btn-secondary" :disabled="loading">
          <RefreshCw :size="16" :class="{ 'spinning': loading }" />
        </button>
      </div>
    </PageHeader>

    <div v-if="loading && !task" class="loading-state">
      <RefreshCw :size="24" class="spinning" />
      <span>{{ t('common.loading') }}...</span>
    </div>

    <div v-else-if="error" class="error-state card">
      <p class="text-error">{{ error }}</p>
      <button @click="fetchTask" class="btn btn-primary mt-4">{{ t('common.retry') }}</button>
    </div>

    <div v-else-if="task" class="task-content">
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <!-- Task Info Card -->
        <div class="lg:col-span-1">
          <div class="card info-card">
            <h3 class="section-title">{{ t('tasks.info') }}</h3>
            <div class="info-list">
              <div class="info-item">
                <span class="label">ID</span>
                <span class="value mono text-xs">{{ task.id }}</span>
              </div>
              <div class="info-item">
                <span class="label">{{ t('tasks.type') }}</span>
                <span class="value">{{ task.type }}</span>
              </div>
              <div class="info-item">
                <span class="label">{{ t('tasks.targetMachine') }}</span>
                <span class="value">{{ getMachineName(task.machine_id) }}</span>
              </div>
              <div class="info-item">
                <span class="label">{{ t('tasks.startTime') }}</span>
                <span class="value">{{ formatTime(task.created_at) }}</span>
              </div>
              <div v-if="task.finished_at" class="info-item">
                <span class="label">{{ t('tasks.endTime') }}</span>
                <span class="value">{{ formatTime(task.finished_at) }}</span>
              </div>
              <div class="info-item vertical">
                <span class="label">{{ t('tasks.message') }}</span>
                <span class="value message-box">{{ task.message }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Logs Card -->
        <div class="lg:col-span-2">
          <div class="card log-card">
            <div class="log-header">
              <h3 class="section-title m-0">{{ t('tasks.logs') }}</h3>
              <div class="log-actions">
                <label class="auto-scroll-toggle">
                  <input type="checkbox" v-model="autoScroll" />
                  <span>{{ t('tasks.autoScroll') }}</span>
                </label>
                <button @click="copyLogs" class="btn btn-sm btn-ghost" :title="t('common.copy')">
                  <Copy :size="14" />
                </button>
              </div>
            </div>
            <div class="log-container" ref="logContainer">
              <pre v-if="task.logs" class="log-content"><code>{{ task.logs }}</code></pre>
              <div v-else class="log-empty">
                <span v-if="task.status === 'pending'">{{ t('tasks.waitingLogs') }}...</span>
                <span v-else>{{ t('tasks.noLogs') }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ArrowLeft,
  RefreshCw,
  Copy
} from 'lucide-vue-next'
import { t } from '@/i18n'
import { taskService } from '@/services/taskService'
import { remoteMachineService } from '@/services/remoteMachineService'
import type { Task, RemoteMachine } from '@/types'
import PageHeader from '@/components/ui/PageHeader.vue'

const route = useRoute()
const router = useRouter()
const taskId = route.params.id as string

const loading = ref(false)
const error = ref('')
const task = ref<Task | null>(null)
const machines = ref<RemoteMachine[]>([])
const logContainer = ref<HTMLElement | null>(null)
const autoScroll = ref(true)
let pollTimer: number | null = null

const fetchTask = async () => {
  if (!task.value) loading.value = true
  try {
    const data = await taskService.get(taskId)
    task.value = data
    
    // If task is still running or pending, continue polling
    if (data.status === 'running' || data.status === 'pending') {
      startPolling()
    } else {
      stopPolling()
    }

    if (autoScroll.value) {
      scrollToBottom()
    }
  } catch (e: unknown) {
    error.value = e instanceof Error ? e.message : t('common.error')
    stopPolling()
  } finally {
    loading.value = false
  }
}

const fetchMachines = async () => {
  try {
    machines.value = await remoteMachineService.list()
  } catch (e) {
    console.error('Failed to fetch machines:', e)
  }
}

const getStatusBadgeClass = (status: Task['status']) => {
  const map = {
    pending: 'badge-warning',
    running: 'badge-info',
    completed: 'badge-success',
    failed: 'badge-error'
  }
  return map[status] || ''
}

const getMachineName = (id: string) => {
  if (id === 'local') return 'Local'
  const machine = machines.value.find(m => m.id === id)
  return machine ? machine.name : id
}

const formatTime = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString()
}

const copyLogs = () => {
  if (task.value?.logs) {
    navigator.clipboard.writeText(task.value.logs)
    // Could add a toast here if available
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight
    }
  })
}

const startPolling = () => {
  if (!pollTimer) {
    pollTimer = window.setInterval(fetchTask, 2000)
  }
}

const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

watch(() => task.value?.logs, () => {
  if (autoScroll.value) {
    scrollToBottom()
  }
})

onMounted(() => {
  fetchTask()
  fetchMachines()
})

onUnmounted(() => {
  stopPolling()
})
</script>

<style scoped>
.task-detail-page {
  max-width: 1400px;
  margin: 0 auto;
}

.header-actions {
  display: flex;
  align-items: center;
}

.section-title {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-strong);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: var(--space-4);
  border-bottom: 1px solid var(--color-border-weak);
  padding-bottom: var(--space-2);
}

.info-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  font-size: var(--font-size-sm);
}

.info-item.vertical {
  flex-direction: column;
  gap: var(--space-2);
}

.info-item .label {
  color: var(--color-text-weak);
  white-space: nowrap;
}

.info-item .value {
  color: var(--color-text-strong);
  font-weight: var(--font-weight-medium);
  text-align: right;
}

.info-item.vertical .value {
  text-align: left;
  width: 100%;
}

.message-box {
  background: var(--color-background-weak);
  padding: var(--space-3);
  border-radius: var(--radius-sm);
  border: 1px solid var(--color-border-weak);
  line-height: 1.5;
}

.log-card {
  display: flex;
  flex-direction: column;
  height: 600px;
}

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-4);
}

.log-actions {
  display: flex;
  align-items: center;
  gap: var(--space-4);
}

.auto-scroll-toggle {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-size: var(--font-size-xs);
  color: var(--color-text-weak);
  cursor: pointer;
}

.log-container {
  flex: 1;
  background: #1e1e1e;
  border-radius: var(--radius-sm);
  overflow: auto;
  padding: var(--space-4);
  color: #d4d4d4;
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  line-height: 1.5;
}

.log-content {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

.log-empty {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #666;
}

.mono {
  font-family: var(--font-mono);
}

.loading-state, .error-state {
  padding: var(--space-16) 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.grid { display: grid; }
.grid-cols-1 { grid-template-columns: repeat(1, minmax(0, 1fr)); }
.gap-6 { gap: 1.5rem; }

@media (min-width: 1024px) {
  .lg\:grid-cols-3 { grid-template-columns: repeat(3, minmax(0, 1fr)); }
  .lg\:col-span-1 { grid-column: span 1 / span 1; }
  .lg\:col-span-2 { grid-column: span 2 / span 2; }
}

.m-0 { margin: 0; }
.mr-2 { margin-right: 0.5rem; }
.mr-3 { margin-right: 0.75rem; }
.mt-4 { margin-top: 1rem; }
.mb-4 { margin-bottom: 1rem; }
.ml-2 { margin-left: 0.5rem; }
.w-full { width: 100%; }
.text-xs { font-size: 0.75rem; }
.text-error { color: var(--color-error); }
</style>
