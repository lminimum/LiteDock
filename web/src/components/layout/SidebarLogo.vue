<template>
  <div class="sidebar-header" :class="{ collapsed }">
    <div class="logo">
      <span class="logo-text">LiteDock</span>
    </div>
    <div class="logo-collapsed">
      <span class="logo-text">LD</span>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  collapsed: boolean
}>()
</script>

<style scoped>
.sidebar-header {
  padding: 0 var(--space-4);
  border-bottom: 1px solid var(--color-border-weak);
  height: var(--header-height);
  display: flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  position: relative;
}

.logo,
.logo-collapsed {
  position: absolute;
  display: flex;
  align-items: center;
  transition: opacity 0.2s ease;
}

.logo {
  opacity: 1;
}

.logo-collapsed {
  opacity: 0;
}

.logo-text {
  font-family: var(--font-mono);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-strong);
  letter-spacing: -0.02em;
  white-space: nowrap;
}

.logo .logo-text {
  font-size: var(--font-size-base);
}

.logo-collapsed .logo-text {
  font-size: var(--font-size-sm);
}

/* Collapsed state - show short form */
.sidebar-header.collapsed .logo {
  opacity: 0;
  pointer-events: none;
}

.sidebar-header.collapsed .logo-collapsed {
  opacity: 1;
}
</style>
