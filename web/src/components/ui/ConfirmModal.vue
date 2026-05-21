<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('cancel')">
    <div class="card confirm-modal">
      <div class="modal-header">
        <h3>{{ title }}</h3>
        <button class="btn btn-ghost btn-sm" @click="emit('cancel')" :disabled="disabled" aria-label="Close">
          <X :size="16" />
        </button>
      </div>

      <div class="modal-body">
        <p class="confirm-message">{{ message }}</p>
      </div>

      <div class="modal-actions">
        <button class="btn btn-ghost" @click="emit('cancel')" :disabled="disabled">
          {{ cancelText }}
        </button>
        <button :class="confirmButtonClass" @click="emit('confirm')" :disabled="disabled">
          {{ confirmText }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { X } from 'lucide-vue-next'
import { t } from '@/i18n'

const props = withDefaults(defineProps<{
  visible: boolean
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  danger?: boolean
  disabled?: boolean
}>(), {
  confirmText: () => t('common.confirm'),
  cancelText: () => t('common.cancel'),
  danger: false,
  disabled: false,
})

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()

const confirmButtonClass = computed(() => (
  props.danger ? 'btn btn-danger' : 'btn btn-primary'
))
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  background: var(--color-background-overlay);
  padding: var(--space-4);
}

.confirm-modal {
  width: min(92vw, 480px);
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
}

.modal-header h3 {
  margin: 0;
  font-family: var(--font-mono);
  font-size: var(--font-size-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
}

.modal-body {
  padding: var(--space-6);
}

.confirm-message {
  margin: 0;
  color: var(--color-text-strong);
  line-height: 1.6;
  white-space: pre-line;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: var(--space-3);
  padding: 0 var(--space-6) var(--space-6);
}
</style>
