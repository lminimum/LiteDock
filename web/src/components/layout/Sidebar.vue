<!-- web/src/components/layout/Sidebar.vue -->
<template>
  <aside class="sidebar" :class="{ collapsed: effectiveCollapsed, 'mobile-open': mobileOpen, 'mobile-mode': isMobile }">
    <SidebarLogo :collapsed="effectiveCollapsed" :is-mobile="isMobile" @close="handleClose" />
    <SidebarNav :collapsed="effectiveCollapsed" :is-mobile="isMobile" />
    <SidebarUserInfo v-if="!isMobile" :collapsed="effectiveCollapsed" @toggle="handleToggle" />
  </aside>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useSidebar } from '@/composables/useSidebar'
import SidebarLogo from '@/components/layout/SidebarLogo.vue'
import SidebarNav from '@/components/nav/SidebarNav.vue'
import SidebarUserInfo from '@/components/layout/SidebarUserInfo.vue'

const props = defineProps<{
  mobileOpen?: boolean
}>()

const emit = defineEmits<{
  (e: 'toggle'): void
}>()

const route = useRoute()
const { collapsed: userCollapsed, toggle } = useSidebar()

// Mobile breakpoint: viewport < 768px
const isMobile = ref(false)
// Tablet breakpoint: viewport 768px - 1023px
const isTablet = ref(false)

const updateViewport = () => {
  isMobile.value = window.innerWidth < 768
  isTablet.value = window.innerWidth >= 768 && window.innerWidth <= 1023
}

onMounted(() => {
  updateViewport()
  window.addEventListener('resize', updateViewport)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateViewport)
})

// On mobile: never collapsed (show full nav)
// On tablet: always collapsed (icon only)
// On desktop: follows user preference
const effectiveCollapsed = computed(() => {
  if (isMobile.value) return false
  return userCollapsed.value || isTablet.value
})

// Toggle sidebar - call useSidebar toggle directly for PC/Tablet
const handleToggle = () => {
  if (!isMobile.value) {
    toggle()
  }
}

// Close sidebar on mobile
const handleClose = () => {
  emit('toggle')
}

// Watch route changes to close sidebar on mobile
watch(() => route.path, () => {
  if (isMobile.value && props.mobileOpen) {
    handleClose()
  }
})
</script>

<style scoped>
.sidebar {
  width: var(--sidebar-width);
  height: 100%;
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

/* Mobile: sidebar becomes fullscreen overlay drawer */
@media (max-width: 767px) {
  .sidebar {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    width: 100vw;
    height: 100vh;
    transform: translateX(-100%);
    z-index: 100;
  }

  .sidebar.mobile-open {
    transform: translateX(0);
  }

  /* On mobile, always show full width (not collapsed) */
  .sidebar.collapsed {
    width: 100vw;
  }
}

/* Tablet: 768px - 1023px */
@media (min-width: 768px) and (max-width: 1023px) {
  .sidebar {
    width: var(--sidebar-collapsed-width);
  }
}
</style>