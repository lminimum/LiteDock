<!-- web/src/components/layout/SidebarLogo.vue -->
<template>
  <div class="sidebar-header" :class="{ collapsed, 'mobile-mode': isMobile }">
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
  isMobile: boolean
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

/* Mobile close button */
.close-btn {
  position: absolute;
  right: var(--space-4);
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  background: none;
  border: none;
  color: var(--color-text-weak);
  cursor: pointer;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}

.close-btn:hover {
  background: var(--color-background-weak);
  color: var(--color-text-strong);
}

.close-btn:active {
  transform: translateY(-50%) scale(0.95);
}

/* Mobile mode adjustments */
.sidebar-header.mobile-mode {
  justify-content: center;
}

.sidebar-header.mobile-mode .logo,
.sidebar-header.mobile-mode .logo-collapsed {
  position: static;
}
</style>