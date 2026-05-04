<template>
  <div class="card card-hover" :class="{ 'status-running': container.status === 'running' }">
    <div class="card-header">
      <div class="card-title">{{ container.name }}</div>
      <div class="badge" :class="statusClass">
        {{ statusText }}
      </div>
    </div>

    <div class="card-body">
      <div class="card-info-row">
        <span class="label">Image</span>
        <span class="value">{{ container.image }}</span>
      </div>
      <div class="card-info-row">
        <span class="label">Ports</span>
        <span class="value">{{ (container.ports || []).join(', ') || '-' }}</span>
      </div>
      <div v-if="container.machine" class="card-info-row">
        <span class="label">Machine</span>
        <span class="value">{{ container.machine }}</span>
      </div>
      <div class="card-info-row">
        <span class="label">Created</span>
        <span class="value">{{ formatDate(container.createdAt) }}</span>
      </div>
    </div>

    <div class="card-actions">
      <button @click="emit('inspect', container.id)" class="btn btn-sm btn-ghost">
        <Info :size="14" />
        Details
      </button>
      <button
        v-if="container.status === 'stopped'"
        @click="emit('start', container.id)"
        class="btn btn-sm btn-secondary"
      >
        <Play :size="14" />
        Start
      </button>
      <button
        v-if="container.status === 'running'"
        @click="emit('stop', container.id)"
        class="btn btn-sm btn-secondary"
      >
        <Square :size="14" />
        Stop
      </button>
      <button
        v-if="container.status === 'running'"
        @click="emit('restart', container.id)"
        class="btn btn-sm btn-secondary"
      >
        <RotateCcw :size="14" />
        Restart
      </button>
      <button @click="emit('logs', container.id)" class="btn btn-sm btn-ghost">
        <FileText :size="14" />
        Logs
      </button>
      <button @click="emit('delete', container.id)" class="btn btn-sm btn-ghost btn-danger-text">
        <Trash2 :size="14" />
        Delete
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Info, Play, Square, RotateCcw, FileText, Trash2 } from 'lucide-vue-next'

interface ContainerCardProps {
  id: string
  name: string
  image: string
  status: 'running' | 'stopped' | 'paused' | 'restarting' | 'exited' | 'created'
  ports: string[]
  createdAt: string
  machine: string
  machineId: string
}

const props = defineProps<{
  container: ContainerCardProps
}>()

const emit = defineEmits<{
  inspect: [id: string]
  start: [id: string]
  stop: [id: string]
  restart: [id: string]
  logs: [id: string]
  delete: [id: string]
}>()

const statusText = computed(() => {
  const map: Record<string, string> = {
    running: 'Running',
    stopped: 'Stopped',
    paused: 'Paused',
    restarting: 'Restarting',
    exited: 'Exited',
    created: 'Created',
  }
  return map[props.container.status] || props.container.status
})

const statusClass = computed(() => {
  const map: Record<string, string> = {
    running: 'badge-success',
    stopped: 'badge-error',
    paused: 'badge-warning',
    restarting: 'badge-warning',
    exited: 'badge-error',
    created: 'badge-info',
  }
  return map[props.container.status] || ''
})

function formatDate(date: string): string {
  if (!date) return '-'
  const d = new Date(date)
  if (isNaN(d.getTime())) return '-'
  return new Intl.DateTimeFormat('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(d)
}
</script>

<style scoped>
.card {
  border-left: 3px solid var(--color-border);
}

.status-running {
  border-left-color: var(--color-success);
}
</style>
