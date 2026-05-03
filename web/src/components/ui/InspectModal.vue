<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('close')">
    <div class="card inspect-modal">
      <div class="modal-header">
        <h3>{{ title }}</h3>
        <button class="btn btn-ghost btn-sm" @click="emit('close')">
          <X :size="16" />
        </button>
      </div>
      <div class="modal-body">
        <div
          v-for="field in fields"
          :key="field.label"
          class="inspect-field"
        >
          <span class="inspect-label">{{ field.label }}</span>
          <span class="inspect-value">{{ field.value }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { X } from 'lucide-vue-next'

defineProps<{
  visible: boolean
  title: string
  fields: Array<{ label: string; value: string }>
  entityId?: string
}>()

const emit = defineEmits<{
  close: []
}>()
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 50;
  background: var(--color-background-overlay);
}

.inspect-modal {
  max-width: 480px;
  width: 90vw;
  max-height: 80vh;
  overflow-y: auto;
  padding: 0;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-4) var(--space-6);
  border-bottom: 1px solid var(--color-border);
  margin-bottom: var(--space-4);
}

.modal-header h3 {
  margin: 0;
  font-family: var(--font-mono);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
}

.modal-body {
  padding: 0 var(--space-6) var(--space-6);
}

.inspect-field {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-3) 0;
  border-bottom: 1px solid var(--color-border-weak);
  font-size: var(--font-size-sm);
}

.inspect-field:last-child {
  border-bottom: none;
}

.inspect-label {
  color: var(--color-text-weak);
  flex-shrink: 0;
  margin-right: var(--space-4);
}

.inspect-value {
  color: var(--color-text-strong);
  font-weight: var(--font-weight-medium);
  text-align: right;
  word-break: break-all;
  min-width: 0;
  overflow: hidden;
}
</style>
