<template>
  <div class="card card-hover network-card">
    <div class="network-header">
      <div class="network-name">{{ network.name }}</div>
      <div class="flex items-center gap-2">
        <span class="badge" :class="driverBadgeClass">{{ network.driver }}</span>
        <span class="badge badge-info">{{ network.scope }}</span>
        <span v-if="network.internal" class="badge badge-warning">Internal</span>
      </div>
    </div>

    <div class="network-info">
      <div class="info-item">
        <span class="label">ID</span>
        <span class="value">{{ network.id.substring(0, 12) }}</span>
      </div>
      <div class="info-item">
        <span class="label">Containers</span>
        <span class="value">{{ network.containers?.length ?? 0 }}</span>
      </div>
    </div>

    <div class="network-actions">
      <button @click="emit('delete', network.id)" class="btn btn-sm btn-danger">
        <Trash2 :size="14" />
        Delete
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Trash2 } from 'lucide-vue-next'
import type { Network } from '@/types'

const props = defineProps<{
  network: Network
}>()

const emit = defineEmits<{
  delete: [networkId: string]
}>()

const driverBadgeClass = computed(() => {
  const map: Record<string, string> = {
    bridge: 'badge-info',
    host: 'badge-success',
    overlay: 'driver-overlay',
    macvlan: 'badge-warning',
    ipvlan: 'driver-ipvlan',
  }
  return map[props.network.driver] ?? ''
})
</script>

<style scoped>
.network-card {
  transition: box-shadow var(--transition-fast), border-color var(--transition-fast);
}

.network-card:hover {
  box-shadow: var(--shadow-md);
  border-color: var(--color-border);
}

.network-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--space-4);
  gap: var(--space-3);
}

.network-name {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
  flex-shrink: 0;
}

.network-info {
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

.network-actions {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
  padding-top: var(--space-3);
  border-top: 1px solid var(--color-border-weak);
}

/* Driver-specific badge colors for types not covered by global badge-* classes */
.driver-overlay {
  background: rgba(139, 92, 246, 0.12);
  color: #8b5cf6;
}

.driver-ipvlan {
  background: rgba(20, 184, 166, 0.12);
  color: #14b8a6;
}
</style>
