<template>
  <div class="tasks-page">
    <PageHeader :title="t('tasks.title')">
      <button @click="fetchTasks" class="btn btn-secondary" :disabled="loading">
        <RefreshCw :size="16" :class="{ 'spinning': loading }" />
        {{ t('common.refresh') }}
      </button>
    </PageHeader>

    <CollapsibleFilters
      v-model="searchQuery"
      :search-placeholder="t('tasks.searchPlaceholder')"
      search-label="Search"
      :has-filters="true"
    >
      <template #filters>
        <select v-model="statusFilter" class="input">
          <option value="">{{ t('tasks.allStatuses') }}</option>
          <option value="pending">{{ t('tasks.status.pending') }}</option>
          <option value="running">{{ t('tasks.status.running') }}</option>
          <option value="completed">{{ t('tasks.status.completed') }}</option>
          <option value="failed">{{ t('tasks.status.failed') }}</option>
        </select>
      </template>
      <template #right>
        <ViewToggle v-model="viewMode" />
      </template>
    </CollapsibleFilters>

    <div v-if="loading && tasks.length === 0" class="loading-state">
      <RefreshCw :size="24" class="spinning" />
      <span>{{ t('common.loading') }}...</span>
    </div>

    <div v-else-if="error && tasks.length === 0" class="error-state card text-center">
      <p class="mb-4 text-error">{{ error }}</p>
      <button @click="fetchTasks" class="btn btn-secondary">{{ t('common.refresh') }}</button>
    </div>

    <div v-else class="tasks-sections">
      <!-- In-Progress Tasks Section -->
      <div class="tasks-section mb-8">
        <div class="section-header">
          <div class="section-title">
            <PlayCircle :size="18" class="section-icon text-info" />
            <h2>{{ t('tasks.inProgress') }} ({{ inProgressTasks.length }})</h2>
          </div>
        </div>

        <div v-if="inProgressTasks.length === 0" class="empty-state card text-center py-8">
          <Inbox :size="36" class="empty-icon mb-2 text-weak" />
          <p class="text-weak">{{ t('tasks.noTasksInProgress') }}</p>
        </div>

        <div v-else>
          <Transition name="view-fade" mode="out-in">
            <div v-if="viewMode === 'card'" class="tasks-grid" key="card">
              <div
                v-for="task in inProgressTasks"
                :key="task.id"
                class="task-card card"
                :class="`status-${task.status}`"
                @click="viewTaskDetail(task.id)"
              >
                <div class="task-header">
                  <div class="task-type">
                    <Activity :size="16" class="icon" />
                    {{ task.type }}
                  </div>
                  <div class="badge" :class="getStatusBadgeClass(task.status)">
                    {{ t(`tasks.status.${task.status}`) }}
                  </div>
                </div>

                <div class="task-body">
                  <div class="task-message text-truncate" :title="task.message">
                    {{ task.message }}
                  </div>
                  <div class="task-meta">
                    <div class="meta-item">
                      <Server :size="14" />
                      <span>{{ getMachineName(task.machine_id) }}</span>
                    </div>
                    <div class="meta-item">
                      <Clock :size="14" />
                      <span>{{ formatTime(task.created_at) }}</span>
                    </div>
                  </div>
                </div>

                <div class="task-footer">
                  <button class="btn btn-sm btn-ghost w-full">
                    {{ t('common.details') }}
                  </button>
                </div>
              </div>
            </div>

            <div v-else class="item-list" key="list">
              <div
                v-for="task in inProgressTasks"
                :key="task.id"
                class="item-list-row clickable"
                @click="viewTaskDetail(task.id)"
              >
                <div class="item-list-info">
                  <div class="item-list-title">
                    {{ task.type }}
                    <span class="badge badge-sm ml-2" :class="getStatusBadgeClass(task.status)">
                      {{ t(`tasks.status.${task.status}`) }}
                    </span>
                  </div>
                  <div class="item-list-meta">
                    <span class="text-truncate max-w-md">{{ task.message }}</span>
                    <span><Server :size="12" class="inline mb-1" /> {{ getMachineName(task.machine_id) }}</span>
                    <span><Clock :size="12" class="inline mb-1" /> {{ formatTime(task.created_at) }}</span>
                  </div>
                </div>
                <div class="item-list-actions">
                  <button class="btn btn-sm btn-ghost">
                    <ChevronRight :size="16" />
                  </button>
                </div>
              </div>
            </div>
          </Transition>
        </div>
      </div>

      <!-- History Tasks Section -->
      <div class="tasks-section">
        <div class="section-header">
          <div class="section-title">
            <History :size="18" class="section-icon text-success" />
            <h2>{{ t('tasks.history') }} ({{ historyTasks.length }})</h2>
          </div>
        </div>

        <div v-if="historyTasks.length === 0" class="empty-state card text-center py-8">
          <ClipboardList :size="36" class="empty-icon mb-2 text-weak" />
          <p class="text-weak">{{ t('tasks.noTasks') }}</p>
        </div>

        <div v-else>
          <Transition name="view-fade" mode="out-in">
            <div v-if="viewMode === 'card'" class="tasks-grid" key="card">
              <div
                v-for="task in historyTasks"
                :key="task.id"
                class="task-card card"
                :class="`status-${task.status}`"
                @click="viewTaskDetail(task.id)"
              >
                <div class="task-header">
                  <div class="task-type">
                    <Activity :size="16" class="icon" />
                    {{ task.type }}
                  </div>
                  <div class="badge" :class="getStatusBadgeClass(task.status)">
                    {{ t(`tasks.status.${task.status}`) }}
                  </div>
                </div>

                <div class="task-body">
                  <div class="task-message text-truncate" :title="task.message">
                    {{ task.message }}
                  </div>
                  <div class="task-meta">
                    <div class="meta-item">
                      <Server :size="14" />
                      <span>{{ getMachineName(task.machine_id) }}</span>
                    </div>
                    <div class="meta-item">
                      <Clock :size="14" />
                      <span>{{ formatTime(task.created_at) }}</span>
                    </div>
                  </div>
                </div>

                <div class="task-footer">
                  <button class="btn btn-sm btn-ghost w-full">
                    {{ t('common.details') }}
                  </button>
                </div>
              </div>
            </div>

            <div v-else class="item-list" key="list">
              <div
                v-for="task in historyTasks"
                :key="task.id"
                class="item-list-row clickable"
                @click="viewTaskDetail(task.id)"
              >
                <div class="item-list-info">
                  <div class="item-list-title">
                    {{ task.type }}
                    <span class="badge badge-sm ml-2" :class="getStatusBadgeClass(task.status)">
                      {{ t(`tasks.status.${task.status}`) }}
                    </span>
                  </div>
                  <div class="item-list-meta">
                    <span class="text-truncate max-w-md">{{ task.message }}</span>
                    <span><Server :size="12" class="inline mb-1" /> {{ getMachineName(task.machine_id) }}</span>
                    <span><Clock :size="12" class="inline mb-1" /> {{ formatTime(task.created_at) }}</span>
                  </div>
                </div>
                <div class="item-list-actions">
                  <button class="btn btn-sm btn-ghost">
                    <ChevronRight :size="16" />
                  </button>
                </div>
              </div>
            </div>
          </Transition>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  RefreshCw,
  ClipboardList,
  Activity,
  Server,
  Clock,
  ChevronRight,
  PlayCircle,
  History,
  Inbox
} from 'lucide-vue-next'
import { t } from '@/i18n'
import { taskService } from '@/services/taskService'
import { remoteMachineService } from '@/services/remoteMachineService'
import type { Task, RemoteMachine } from '@/types'
import PageHeader from '@/components/ui/PageHeader.vue'
import { useRefreshBus } from '@/composables/useRefreshBus'
import ViewToggle from '@/components/ui/ViewToggle.vue'
import CollapsibleFilters from '@/components/ui/CollapsibleFilters.vue'
import { useViewMode } from '@/composables/useViewMode'

const router = useRouter()
const loading = ref(false)
const error = ref('')
const searchQuery = ref('')
const statusFilter = ref('')
const viewMode = useViewMode('tasks')
const tasks = ref<Task[]>([])
const machines = ref<RemoteMachine[]>([])
let pollTimer: number | null = null

const fetchTasks = async () => {
  loading.value = true
  error.value = ''
  const hasExistingTasks = tasks.value.length > 0 || machines.value.length > 0
  try {
    const [tasksResult, machinesResult] = await Promise.allSettled([
      taskService.list(),
      remoteMachineService.list()
    ])

    if (tasksResult.status === 'fulfilled') {
      tasks.value = tasksResult.value.sort((a, b) =>
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
      )
    }

    if (machinesResult.status === 'fulfilled') {
      machines.value = machinesResult.value
    }

    const failure = tasksResult.status === 'rejected' ? tasksResult.reason : machinesResult.status === 'rejected' ? machinesResult.reason : null
    if (failure) {
      if (!hasExistingTasks) {
        error.value = failure instanceof Error ? failure.message : t('common.error')
      } else {
        console.error('Failed to refresh tasks:', failure)
      }
    }
  } catch (e: unknown) {
    if (!hasExistingTasks) {
      error.value = e instanceof Error ? e.message : t('common.error')
    } else {
      console.error('Failed to refresh tasks:', e)
    }
  } finally {
    loading.value = false
  }
}

const filteredTasks = computed(() => {
  let filtered = tasks.value

  if (searchQuery.value) {
    const q = searchQuery.value.toLowerCase()
    filtered = filtered.filter(t =>
      t.type.toLowerCase().includes(q) ||
      t.message.toLowerCase().includes(q) ||
      t.id.toLowerCase().includes(q)
    )
  }

  if (statusFilter.value) {
    filtered = filtered.filter(t => t.status === statusFilter.value)
  }

  return filtered
})

const inProgressTasks = computed(() => {
  return filteredTasks.value.filter(t => t.status === 'pending' || t.status === 'running')
})

const historyTasks = computed(() => {
  return filteredTasks.value.filter(t => t.status === 'completed' || t.status === 'failed')
})

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
  if (id === 'local') return t('common.local')
  const machine = machines.value.find(m => m.id === id)
  return machine ? machine.name : id
}

const formatTime = (dateStr: string) => {
  const date = new Date(dateStr)
  return date.toLocaleString()
}

const viewTaskDetail = (id: string) => {
  router.push(`/tasks/${id}`)
}

const startPolling = () => {
  pollTimer = window.setInterval(fetchTasks, 5000)
}

const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onMounted(() => {
  fetchTasks()
  startPolling()
})

useRefreshBus(fetchTasks)

onUnmounted(() => {
  stopPolling()
})
</script>

<style scoped>
.tasks-page {
  max-width: 1400px;
  margin: 0 auto;
}

.tasks-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: var(--space-4);
}

.task-card {
  cursor: pointer;
  transition: transform var(--transition-fast), border-color var(--transition-fast);
  border-left: 4px solid var(--color-border);
}

.task-card:hover {
  transform: translateY(-2px);
  border-color: var(--color-text-weaker);
}

.task-card.status-running { border-left-color: var(--color-info); }
.task-card.status-completed { border-left-color: var(--color-success); }
.task-card.status-failed { border-left-color: var(--color-error); }
.task-card.status-pending { border-left-color: var(--color-warning); }

.task-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-3);
}

.task-type {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
}

.task-message {
  font-size: var(--font-size-sm);
  color: var(--color-text);
  margin-bottom: var(--space-4);
  line-height: 1.5;
  height: 3em;
  overflow: hidden;
}

.task-meta {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
  font-size: var(--font-size-xs);
  color: var(--color-text-weak);
}

.meta-item {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.task-footer {
  margin-top: var(--space-4);
  padding-top: var(--space-3);
  border-top: 1px solid var(--color-border-weak);
}

.item-list-row.clickable {
  cursor: pointer;
}

.item-list-row.clickable:hover {
  background: var(--color-background-weak);
}

.loading-state, .error-state, .empty-state {
  padding: var(--space-16) 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-4);
}

.tasks-sections {
  display: flex;
  flex-direction: column;
  gap: var(--space-8);
}

.tasks-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--color-border-weak);
  padding-bottom: var(--space-2);
}

.section-title {
  display: flex;
  align-items: center;
  gap: var(--space-2);
}

.section-title h2 {
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  margin-bottom: 0;
}

.section-icon {
  flex-shrink: 0;
}

.mb-8 {
  margin-bottom: var(--space-8);
}

.py-8 {
  padding-top: var(--space-8);
  padding-bottom: var(--space-8);
}

.text-weak {
  color: var(--color-text-weak);
}

.empty-icon {
  opacity: 0.3;
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.text-truncate {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.max-w-md {
  max-width: 400px;
}
</style>
