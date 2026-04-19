<template>
  <router-link :to="to" class="nav-item" :class="{ active, collapsed, 'mobile-mode': isMobile }">
    <component :is="icon" :size="20" class="nav-icon" />
    <span class="nav-text">{{ label }}</span>
    <span class="nav-badge" v-if="badge && !isMobile">{{ badge }}</span>
  </router-link>
</template>

<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  to: string
  icon: Component
  label: string
  active: boolean
  collapsed: boolean
  isMobile: boolean
  badge?: string
}>()
</script>

<style scoped>
.nav-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-4);
  color: var(--color-text-weak);
  text-decoration: none;
  font-size: var(--font-size-sm);
  transition: color 0.15s ease, background-color 0.15s ease;
  border-left: 2px solid transparent;
  margin: 1px 0;
  box-sizing: border-box;
}

.nav-item:hover {
  color: var(--color-text-strong);
  background: var(--color-background-hover);
}

.nav-item.active {
  color: var(--color-text-strong);
  background: var(--color-background-hover);
  border-left-color: var(--color-text-strong);
}

.nav-icon {
  flex-shrink: 0;
}

.nav-text {
  white-space: nowrap;
  opacity: 1;
  transition: opacity 0.2s ease;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}

.nav-badge {
  margin-left: auto;
  padding: 0 var(--space-2);
  font-size: var(--font-size-xs);
  background: var(--color-background-strong);
  color: var(--color-background);
  border-radius: var(--radius-full);
  line-height: 1.6;
  opacity: 1;
  transition: opacity 0.2s ease;
  flex-shrink: 0;
}

/* Collapsed state - text fades, icon centers */
.nav-item.collapsed {
  justify-content: center;
  padding: var(--space-2);
}

.nav-item.collapsed .nav-text {
  display: none;
}

.nav-item.collapsed .nav-badge {
  display: none;
}

/* Mobile mode - center items */
@media (max-width: 767px) {
  .nav-item.mobile-mode {
    justify-content: center;
    padding: var(--space-3) var(--space-6);
    margin: var(--space-1) var(--space-4);
    border-radius: var(--radius-md);
    border-left: none;
  }

  .nav-item.mobile-mode:hover {
    background: var(--color-background-hover);
  }

  .nav-item.mobile-mode.active {
    background: var(--color-background-hover);
    border-left: none;
  }

  .nav-item.mobile-mode .nav-text {
    flex: none;
    font-size: var(--font-size-base);
  }

  .nav-item.mobile-mode .nav-badge {
    display: none;
  }
}
</style>