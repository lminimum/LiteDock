<template>
  <div class="card card-hover volume-card">
    <div class="volume-header">
      <div class="volume-name">{{ volume.name }}</div>
      <div class="flex items-center gap-2">
        <span class="badge">{{ volume.driver }}</span>
        <span class="badge badge-info">{{ volume.scope }}</span>
      </div>
    </div>

    <div class="volume-info">
      <div class="info-item">
        <span class="label">Mountpoint</span>
        <span class="value truncate" :title="volume.mountpoint">{{ volume.mountpoint }}</span>
      </div>
      <div class="info-item">
        <span class="label">Size</span>
        <span class="value">{{ formattedSize }}</span>
      </div>
      <div class="info-item">
        <span class="label">Created</span>
        <span class="value">{{ volume.createdAt }}</span>
      </div>
    </div>

    <div class="volume-actions">
      <button @click="emit('delete', volume.name)" class="btn btn-sm btn-danger">
        <Trash2 :size="14" />
        Delete
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Trash2 } from 'lucide-vue-next'
import type { Volume } from '@/types'

const props = defineProps<{
  volume: Volume
}>()

const emit = defineEmits<{
  delete: [volumeName: string]
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

<style scoped>
.volume-card {
  transition: box-shadow var(--transition-fast), border-color var(--transition-fast);
}

.volume-card:hover {
  box-shadow: var(--shadow-md);
  border-color: var(--color-border);
}

.volume-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--space-4);
  gap: var(--space-3);
}

.volume-name {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
  flex-shrink: 0;
}

.volume-info {
  margin-bottom: var(--space-4);
}

.info-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: var(--space-2);
  font-size: var(--font-size-sm);
}

.info-item:last-child {
  margin-bottom: 0;
}

.info-item .label {
  color: var(--color-text-weak);
}

.info-item .value {
  color: var(--color-text-strong);
  font-weight: var(--font-weight-medium);
}

.volume-actions {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
  padding-top: var(--space-3);
  border-top: 1px solid var(--color-border-weak);
}
</style>
