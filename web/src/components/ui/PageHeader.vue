<template>
  <div class="page-header">
    <div class="page-header-content">
      <h1>{{ title }}</h1>
      <p v-if="description" class="page-description">{{ description }}</p>
    </div>
    <div v-if="$slots.default" class="page-header-actions">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description?: string
}>()
</script>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-6);
  flex-wrap: wrap;
  gap: var(--space-4);
}

.page-header-content h1 {
  margin: 0;
  color: var(--color-text-strong);
  font-size: var(--font-size-2xl);
  font-weight: var(--font-weight-semibold);
}

.page-description {
  margin: var(--space-1) 0 0;
  color: var(--color-text);
  font-size: var(--font-size-sm);
}

.page-header-actions {
  display: flex;
  gap: var(--space-3);
}
</style>
