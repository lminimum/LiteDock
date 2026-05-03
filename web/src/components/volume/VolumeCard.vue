<template>
  <div class="card card-hover">
    <div class="card-header">
      <div class="card-title">{{ volume.name }}</div>
      <div class="flex items-center gap-2 flex-shrink-0">
        <span class="badge">{{ volume.driver }}</span>
        <span class="badge badge-info">{{ volume.scope }}</span>
      </div>
    </div>

    <div class="card-body">
      <div class="card-info-row">
        <span class="label">Mountpoint</span>
        <span class="value truncate" :title="volume.mountpoint">{{ volume.mountpoint }}</span>
      </div>
      <div class="card-info-row">
        <span class="label">Size</span>
        <span class="value">{{ formattedSize }}</span>
      </div>
      <div class="card-info-row">
        <span class="label">Created</span>
        <span class="value">{{ volume.createdAt }}</span>
      </div>
    </div>

    <div class="card-actions">
      <button @click="emit('inspect', volume.name)" class="btn btn-sm btn-ghost">
        <Info :size="14" />
        Details
      </button>
      <button @click="emit('delete', volume.name)" class="btn btn-sm btn-danger">
        <Trash2 :size="14" />
        Delete
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Info, Trash2 } from 'lucide-vue-next'
import type { Volume } from '@/types'

const props = defineProps<{
  volume: Volume
}>()

const emit = defineEmits<{
  delete: [volumeName: string]
  inspect: [volumeName: string]
}>()

function formatSize(bytes?: number): string {
  if (bytes === undefined || bytes === null) return '-'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

const formattedSize = computed(() => formatSize(props.volume.size))
</script>
