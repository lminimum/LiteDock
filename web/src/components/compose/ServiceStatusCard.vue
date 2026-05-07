<template>
  <div class="card">
    <div class="card-header">
      <span class="card-title">Services</span>
      <span class="badge badge-info">{{ services.length }}</span>
    </div>

    <div class="card-body" v-if="services.length > 0">
      <div
        v-for="service in services"
        :key="service.name"
        class="flex items-start gap-3 py-2"
        style="border-bottom: 1px solid var(--color-border-weak)"
      >
        <component :is="statusIcon(service.status)" :size="16" :class="statusColor(service.status)" class="flex-shrink-0 mt-0.5" />

        <div class="flex-1" style="min-width: 0">
          <div class="flex items-center gap-2 flex-wrap">
            <span class="font-semibold text-primary">{{ service.serviceName }}</span>
            <span class="text-xs text-muted">{{ service.image }}</span>
          </div>
          <div v-if="service.publishers && service.publishers.length > 0" class="mt-1">
            <span
              v-for="(pub, idx) in service.publishers"
              :key="idx"
              class="text-xs text-muted"
            >
              {{ formatPublisher(pub) }}<span v-if="idx < service.publishers.length - 1">, </span>
            </span>
          </div>
        </div>

        <span
          v-if="service.health"
          :class="['badge', healthBadgeClass(service.health)]"
          class="flex-shrink-0"
        >
          {{ service.health }}
        </span>
      </div>
    </div>

    <div class="card-body" v-else>
      <p class="text-sm text-muted">No services configured.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { CheckCircle, XCircle, PauseCircle, HelpCircle } from 'lucide-vue-next'
import type { ComposeService, ComposePortPublish } from '@/types'

defineProps<{
  services: ComposeService[]
}>()

function statusIcon(status: string) {
  switch (status) {
    case 'running':
      return CheckCircle
    case 'exited':
      return XCircle
    case 'paused':
      return PauseCircle
    default:
      return HelpCircle
  }
}

function statusColor(status: string): string {
  switch (status) {
    case 'running':
      return 'text-success'
    case 'exited':
      return 'text-error'
    case 'paused':
      return 'text-warning'
    default:
      return 'text-muted'
  }
}

function healthBadgeClass(health: string): string {
  switch (health) {
    case 'healthy':
      return 'badge-success'
    case 'unhealthy':
      return 'badge-error'
    default:
      return 'badge-info'
  }
}

function formatPublisher(pub: ComposePortPublish): string {
  return `${pub.url}:${pub.publishedPort}->${pub.targetPort}`
}
</script>

<style scoped>
.text-success {
  color: var(--color-success);
}

.text-error {
  color: var(--color-error);
}

.text-warning {
  color: var(--color-warning);
}
</style>
