<template>
  <Teleport to="body">
    <Transition name="confirm-fade">
      <div v-if="show" class="confirm-overlay" @click.self="emit('cancel')">
        <div class="card confirm-modal">
          <div class="flex items-center gap-3 mb-4">
            <AlertTriangle :size="20" class="confirm-icon" />
            <span class="text-lg font-semibold" style="color: var(--color-text-strong)">
              {{ t('assistant.confirmation.title') }}
            </span>
          </div>

          <div class="confirm-body">
            <p class="confirm-message">{{ message }}</p>

            <div class="confirm-action-box">
              <div class="confirm-action-row">
                <span class="confirm-action-label">Action</span>
                <span class="confirm-action-value">{{ actionName }}</span>
              </div>
              <div
                v-for="(val, key) in actionParams"
                :key="key"
                class="confirm-action-row"
              >
                <span class="confirm-action-label">{{ key }}</span>
                <span class="confirm-action-value">{{ val }}</span>
              </div>
              <div v-if="riskLevel !== 'safe'" class="confirm-action-row">
                <span class="confirm-action-label">Risk</span>
                <span
                  class="badge"
                  :class="riskLevel === 'dangerous' ? 'badge-error' : 'badge-warning'"
                >
                  {{ riskLevel }}
                </span>
              </div>
            </div>

            <div v-if="typedRequired" class="confirm-typed-section">
              <p class="confirm-typed-hint">
                Type <strong>{{ expectedText }}</strong> to confirm this action.
              </p>
              <input
                :value="modelValue"
                class="input confirm-typed-input"
                :placeholder="expectedText"
                @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
                @keyup.enter="emit('confirm')"
              />
            </div>
          </div>

          <div class="flex justify-end gap-2">
            <button class="btn btn-ghost" :disabled="executing" @click="emit('cancel')">
              {{ t('assistant.confirmation.cancel') }}
            </button>
            <button
              class="btn btn-danger"
              :disabled="confirmDisabled"
              @click="emit('confirm')"
            >
              <Loader2 v-if="executing" :size="14" class="spin" />
              <span>{{ executing ? t('assistant.confirmation.waiting') : t('assistant.confirmation.confirm') }}</span>
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { AlertTriangle, Loader2 } from 'lucide-vue-next'
import type { RiskLevel } from '@/composables/useActionConfirmation'

const props = defineProps<{
  show: boolean
  executing: boolean
  message: string
  actionName: string
  actionParams: Record<string, string>
  riskLevel: RiskLevel
  typedRequired: boolean
  expectedText: string
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  confirm: []
  cancel: []
}>()

const { t } = useI18n()

const confirmDisabled = computed(() => {
  if (props.executing) return true
  if (props.typedRequired && props.modelValue !== props.expectedText) return true
  return false
})
</script>

<style scoped>
.confirm-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-background-overlay);
}

.confirm-modal {
  width: 440px;
  max-width: 90vw;
  padding: var(--space-6);
}

.confirm-icon {
  color: var(--color-warning);
  flex-shrink: 0;
}

.confirm-body {
  margin-bottom: var(--space-6);
}

.confirm-message {
  font-size: var(--font-size-sm);
  color: var(--color-text);
  line-height: var(--line-height-relaxed);
  margin: 0 0 var(--space-4);
}

.confirm-action-box {
  background: var(--color-background);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-sm);
  padding: var(--space-3) var(--space-4);
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
}

.confirm-action-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-1) 0;
}

.confirm-action-row + .confirm-action-row {
  border-top: 1px solid var(--color-border-weak);
}

.confirm-action-label {
  color: var(--color-text-weak);
  flex-shrink: 0;
  margin-right: var(--space-4);
}

.confirm-action-value {
  color: var(--color-text-strong);
  font-weight: var(--font-weight-medium);
  text-align: right;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.confirm-typed-section {
  margin-top: var(--space-4);
}

.confirm-typed-hint {
  font-size: var(--font-size-xs);
  color: var(--color-text-weak);
  margin: 0 0 var(--space-2);
}

.confirm-typed-hint strong {
  color: var(--color-error);
  font-weight: var(--font-weight-semibold);
}

.confirm-typed-input {
  font-family: var(--font-mono);
}

.confirm-fade-enter-active,
.confirm-fade-leave-active {
  transition: opacity var(--transition-base);
}

.confirm-fade-enter-from,
.confirm-fade-leave-to {
  opacity: 0;
}

.spin {
  animation: confirm-spin 1s linear infinite;
}

@keyframes confirm-spin {
  from { transform: rotate(0deg); }
  to   { transform: rotate(360deg); }
}
</style>
