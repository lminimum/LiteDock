<template>
  <div class="minimal-progress">
    <div class="minimal-progress-head">
      <span class="minimal-progress-label">{{ label }}</span>
      <span class="minimal-progress-reading">
        <span class="minimal-progress-number">{{ clampedValue }}</span>
        <span class="minimal-progress-unit">{{ unit }}</span>
      </span>
    </div>
    <div class="minimal-progress-track" role="progressbar" :aria-valuenow="clampedValue" aria-valuemin="0" aria-valuemax="100">
      <div
        class="minimal-progress-fill"
        :style="{ width: `${clampedValue}%` }"
        :class="`fill--${colorKey}`"
      ></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

type ProgressColor = 'info' | 'warning' | 'success' | 'error' | 'accent'

interface Props {
  label: string
  value: number
  color?: ProgressColor
  unit?: string
}

const props = withDefaults(defineProps<Props>(), {
  value: 0,
  color: 'info',
  unit: '%'
})

const clampedValue = computed(() => Math.min(100, Math.max(0, Math.round(props.value))))

const colorKey = computed(() => props.color)
</script>

<style scoped>
/* ============================================
   MinimalProgress — thin line with label
   Monospace, flat, OpenCode aesthetic
   ============================================ */

.minimal-progress {
  display: flex;
  flex-direction: column;
  gap: var(--space-2);
}

/* Header row: label on left, value (number + unit) on right */
.minimal-progress-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}

.minimal-progress-label {
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-weak);
  letter-spacing: 0.02em;
  text-transform: uppercase;
}

.minimal-progress-reading {
  font-family: var(--font-mono);
  display: flex;
  align-items: baseline;
  gap: var(--space-1);
}

.minimal-progress-number {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
  line-height: var(--line-height-tight);
}

.minimal-progress-unit {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-normal);
  color: var(--color-text-weaker);
}

/* Track — dark, thin, subtle */
.minimal-progress-track {
  height: 4px;
  background: var(--color-background);
  border-radius: var(--radius-full);
  overflow: hidden;
}

/* Fill bar — animated width */
.minimal-progress-fill {
  height: 100%;
  border-radius: var(--radius-full);
  transition: width var(--transition-base);
}

/* Color variants — map to semantic color tokens */
.fill--info {
  background: var(--color-info);
}

.fill--warning {
  background: var(--color-warning);
}

.fill--success {
  background: var(--color-success);
}

.fill--error {
  background: var(--color-error);
}

.fill--accent {
  background: var(--color-accent);
}
</style>
