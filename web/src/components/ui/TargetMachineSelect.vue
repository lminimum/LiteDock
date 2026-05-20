<template>
  <div class="form-field">
    <label class="form-label" :for="selectId">Target Machine <span class="required">*</span></label>
    <select
      :id="selectId"
      class="input"
      :value="modelValue"
      :disabled="disabled || options.length === 0"
      required
      @change="handleChange"
    >
      <option value="" disabled>Select target machine</option>
      <option v-for="option in options" :key="option.value" :value="option.value">
        {{ option.label }}{{ option.status ? ` (${option.status})` : '' }}
      </option>
    </select>
    <p v-if="hint" class="machine-hint">{{ hint }}</p>
  </div>
</template>

<script setup lang="ts">
export interface TargetMachineOption {
  value: string
  label: string
  status?: string
}

defineProps<{
  modelValue: string
  options: TargetMachineOption[]
  disabled?: boolean
  hint?: string
  selectId?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

function handleChange(event: Event) {
  const target = event.target as HTMLSelectElement
  emit('update:modelValue', target.value)
}
</script>

<style scoped>
.form-field {
  margin-bottom: var(--space-4);
}

.form-label {
  display: block;
  margin-bottom: var(--space-2);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text);
}

.required {
  color: var(--color-error);
}

.machine-hint {
  margin-top: var(--space-2);
  font-size: var(--font-size-xs);
  color: var(--color-text-weak);
}
</style>
