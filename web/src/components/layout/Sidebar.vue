<!-- web/src/components/layout/Sidebar.vue -->
<template>
  <aside class="sidebar" :class="{ collapsed: effectiveCollapsed }">
    <SidebarLogo :collapsed="effectiveCollapsed" />
    <SidebarNav :collapsed="effectiveCollapsed" />
    <SidebarUserInfo :collapsed="effectiveCollapsed" @toggle="toggle" />
  </aside>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useSidebar } from '@/composables/useSidebar'
import SidebarLogo from '@/components/layout/SidebarLogo.vue'
import SidebarNav from '@/components/nav/SidebarNav.vue'
import SidebarUserInfo from '@/components/layout/SidebarUserInfo.vue'

const { collapsed: userCollapsed, toggle } = useSidebar()

// Tablet breakpoint: force collapsed when viewport is 768-1023px
const isTablet = ref(false)

const updateTablet = () => {
  isTablet.value = window.innerWidth >= 768 && window.innerWidth <= 1023
}

onMounted(() => {
  updateTablet()
  window.addEventListener('resize', updateTablet)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateTablet)
})

// Children see collapsed=true when user collapsed OR tablet viewport
const effectiveCollapsed = computed(() => userCollapsed.value || isTablet.value)
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
}

.sidebar.collapsed {
  width: var(--sidebar-collapsed-width);
}

@media (max-width: 767px) {
  .sidebar {
    display: none;
  }
}

@media (min-width: 768px) and (max-width: 1023px) {
  .sidebar {
    width: var(--sidebar-collapsed-width) !important;
  }
}
</style>
