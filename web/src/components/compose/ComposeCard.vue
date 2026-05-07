<template>
  <div class="card card-hover" :class="`status-${project.status}`">
    <div class="card-header">
      <div class="flex items-center gap-2 flex-shrink-1" style="min-width: 0">
        <FolderOpen :size="18" class="text-muted flex-shrink-0" />
        <div class="card-title">{{ project.name }}</div>
      </div>
      <div class="flex items-center gap-2 flex-shrink-0">
        <span :class="['badge', statusBadgeClass]">{{ project.status }}</span>
      </div>
    </div>

    <div class="card-body">
      <div class="card-info-row">
        <span class="label">File</span>
        <span class="value truncate" :title="project.filePath">{{ displayPath }}</span>
      </div>
      <div class="card-info-row">
        <span class="label">Services</span>
        <span class="value">{{ serviceCount }}</span>
      </div>
      <div class="card-info-row">
        <span class="label">Created</span>
        <span class="value">{{ formattedDate }}</span>
      </div>
    </div>

    <div class="card-actions">
      <button @click="emit('up', project.id)" class="btn btn-sm btn-primary">
        <Play :size="14" />
        Up
      </button>
      <button @click="emit('down', project.id)" class="btn btn-sm btn-warning">
        <Square :size="14" />
        Down
      </button>
      <button @click="emit('logs', project.id)" class="btn btn-sm btn-ghost">
        <ScrollText :size="14" />
        Logs
      </button>
      <button @click="emit('inspect', project.id)" class="btn btn-sm btn-ghost">
        <Eye :size="14" />
        Inspect
      </button>
      <button @click="emit('delete', project.id)" class="btn btn-sm btn-danger">
        <Trash2 :size="14" />
        Delete
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { FolderOpen, Play, Square, ScrollText, Eye, Trash2 } from 'lucide-vue-next'
import type { ComposeProject } from '@/types'

const props = defineProps<{
  project: ComposeProject
}>()

const emit = defineEmits<{
  inspect: [projectId: string]
  delete: [projectId: string]
  up: [projectId: string]
  down: [projectId: string]
  logs: [projectId: string]
}>()

const statusBadgeClass = computed(() => {
  switch (props.project.status) {
    case 'running':
      return 'badge-success'
    case 'paused':
      return 'badge-warning'
    default:
      return 'badge-error'
  }
})

const serviceCount = computed(() => props.project.services?.length ?? 0)

const displayPath = computed(() => {
  const path = props.project.filePath
  if (!path) return '-'
  const parts = path.split('/')
  return parts[parts.length - 1] || path
})

const formattedDate = computed(() => {
  if (!props.project.createdAt) return '-'
  try {
    return new Date(props.project.createdAt).toLocaleDateString()
  } catch {
    return props.project.createdAt
  }
})
</script>

<style scoped>
.status-running {
  border-color: var(--color-success);
}

.status-stopped,
.status-failed,
.status-exited {
  border-color: var(--color-error);
}

.status-paused {
  border-color: var(--color-warning);
}
</style>
