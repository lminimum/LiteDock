<template>
  <div class="view-toggle">
    <button
      class="view-toggle-btn"
      :class="{ active: modelValue === 'card' }"
      @click="$emit('update:modelValue', 'card')"
      :title="t('common.viewCard')"
    >
      <Grid :size="16" />
    </button>
    <button
      class="view-toggle-btn"
      :class="{ active: modelValue === 'list' }"
      @click="$emit('update:modelValue', 'list')"
      :title="t('common.viewList')"
    >
      <List :size="16" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { Grid, List } from 'lucide-vue-next'
import { t } from '@/i18n'

defineProps<{
  modelValue: 'card' | 'list'
}>()

defineEmits<{
  'update:modelValue': [value: 'card' | 'list']
}>()
</script>

<style scoped>
.view-toggle {
  display: inline-flex;
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-sm);
  overflow: hidden;
}

.view-toggle-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: transparent;
  border: none;
  color: var(--color-text-weaker);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.view-toggle-btn:hover {
  background: var(--color-background-weak);
  color: var(--color-text-weak);
}

.view-toggle-btn.active {
  background: var(--color-background-strong);
  color: var(--color-text-strong);
}

.view-toggle-btn + .view-toggle-btn {
  border-left: 1px solid var(--color-border-weak);
}
</style>
