<!-- web/src/components/layout/Sidebar.vue -->
<template>
  <aside class="sidebar" :class="{ collapsed: effectiveCollapsed, 'mobile-open': mobileOpen }">
    <SidebarLogo :collapsed="effectiveCollapsed" />
    <SidebarNav :collapsed="effectiveCollapsed" />
    <SidebarUserInfo :collapsed="effectiveCollapsed" @toggle="$emit('toggle')" />
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useSidebar } from '@/composables/useSidebar'
import SidebarLogo from '@/components/layout/SidebarLogo.vue'
import SidebarNav from '@/components/nav/SidebarNav.vue'
import SidebarUserInfo from '@/components/layout/SidebarUserInfo.vue'

defineProps<{
  mobileOpen?: boolean
}>()

defineEmits<{
  (e: 'toggle'): void
}>()

const { collapsed: userCollapsed } = useSidebar()

// Tablet breakpoint: force collapsed when viewport is 768-1023px
const effectiveCollapsed = computed(() => userCollapsed.value)
</script>

<style scoped>
.sidebar {
  width: var(--sidebar-width);
  background: var(--color-background);
  border-right: 1px solid var(--color-border-weak);
  display: flex;
  flex-direction: column;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  z-index: 50;
}

.sidebar.collapsed {
  width: var(--sidebar-collapsed-width);
}

/* Mobile: sidebar becomes overlay drawer */
@media (max-width: 767px) {
  .sidebar {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    width: var(--sidebar-width);
    transform: translateX(-100%);
    box-shadow: 4px 0 20px rgba(0, 0, 0, 0.15);
  }

  .sidebar.mobile-open {
    transform: translateX(0);
  }

  .sidebar.collapsed {
    width: var(--sidebar-width);
  }
}

/* Tablet: 768px - 1023px */
@media (min-width: 768px) and (max-width: 1023px) {
  .sidebar {
    width: var(--sidebar-collapsed-width);
  }
}
</style>
