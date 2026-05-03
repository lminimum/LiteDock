<template>
  <div class="card card-hover">
    <div class="card-header">
      <div class="card-title">{{ network.name }}</div>
      <div class="flex items-center gap-2 flex-shrink-0">
        <span class="badge" :class="driverBadgeClass">{{ network.driver }}</span>
        <span class="badge badge-info">{{ network.scope }}</span>
        <span v-if="network.internal" class="badge badge-warning">Internal</span>
      </div>
    </div>

    <div class="card-body">
      <div class="card-info-row">
        <span class="label">ID</span>
        <span class="value">{{ network.id.substring(0, 12) }}</span>
      </div>
      <div class="card-info-row">
        <span class="label">Containers</span>
        <span class="value">{{ network.containers?.length ?? 0 }}</span>
      </div>
    </div>

    <div class="card-actions">
      <button @click="emit('inspect', network.id)" class="btn btn-sm btn-ghost">
        <Info :size="14" />
        Details
      </button>
      <button @click="emit('delete', network.id)" class="btn btn-sm btn-danger">
        <Trash2 :size="14" />
        Delete
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Info, Trash2 } from 'lucide-vue-next'
import type { Network } from '@/types'

const props = defineProps<{
  network: Network
}>()

const emit = defineEmits<{
  delete: [networkId: string]
  inspect: [networkId: string]
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
.driver-overlay {
  background: rgba(139, 92, 246, 0.12);
  color: #8b5cf6;
}

.driver-ipvlan {
  background: rgba(20, 184, 166, 0.12);
  color: #14b8a6;
}
</style>
