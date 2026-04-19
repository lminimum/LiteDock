<template>
  <div class="breadcrumb">
    <span v-for="(crumb, index) in crumbs" :key="index">
      {{ crumb }}
      <span v-if="index < crumbs.length - 1" class="breadcrumb-separator">/</span>
    </span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  crumbs: string[]
}>()
</script>

<style scoped>
.breadcrumb {
  font-size: var(--font-size-xs);
  color: var(--color-text-weaker);
  margin-top: 2px;
}

.breadcrumb-separator {
  margin: 0 var(--space-1);
  color: var(--color-border);
}
</style>
