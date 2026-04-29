<!-- web/src/components/layout/Sidebar.vue -->
<template>
  <aside class="sidebar" :class="{ collapsed: effectiveCollapsed, 'mobile-open': mobileOpen && !isClosing, 'mobile-mode': isMobile, 'mobile-closing': isClosing }">
    <SidebarLogo :collapsed="effectiveCollapsed" :is-mobile="isMobile" />
    <SidebarNav :collapsed="effectiveCollapsed" :is-mobile="isMobile" />
    <SidebarUserInfo v-if="!isMobile" :collapsed="effectiveCollapsed" @toggle="handleToggle" />
  </aside>
</template>

<script setup lang="ts">
import { computed, ref, onUnmounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useSidebarStore } from '@/stores/sidebar'
import { useViewport } from '@/composables/useViewport'
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
const sidebarStore = useSidebarStore()
const { isMobile, isTablet } = useViewport()

// Track closing state for slide-out animation
const isClosing = ref(false)

let closeTimeout: ReturnType<typeof setTimeout> | null = null

onUnmounted(() => {
  if (closeTimeout) clearTimeout(closeTimeout)
})

// On mobile: never collapsed (show full nav)
// On tablet: always collapsed (icon only)
// On desktop: follows user preference
const effectiveCollapsed = computed(() => {
  if (isMobile.value) return false
  return sidebarStore.collapsed || isTablet.value
})

// Toggle sidebar - call store toggle directly for PC/Tablet
const handleToggle = () => {
  if (!isMobile.value) {
    sidebarStore.toggle()
  }
}

// Watch route changes to close sidebar on mobile
watch(() => route.path, () => {
  if (isMobile.value && props.mobileOpen) {
    emit('toggle')
  }
})

// Animate out when mobileOpen becomes false
watch(() => props.mobileOpen, (open) => {
  if (isMobile.value) {
    if (!open && !isClosing.value) {
      // Starting close animation
      isClosing.value = true
      if (closeTimeout) clearTimeout(closeTimeout)
      closeTimeout = setTimeout(() => {
        isClosing.value = false
        emit('toggle')
      }, 350)
    } else if (open) {
      // Opening - cancel any closing animation
      if (closeTimeout) {
        clearTimeout(closeTimeout)
        closeTimeout = null
      }
      isClosing.value = false
    }
  }
})
</script>

<style scoped>
.sidebar {
  width: var(--sidebar-width);
  height: 100%;
  background: var(--color-background);
  border-right: 1px solid rgba(255, 255, 255, 0.12);
  display: flex;
  flex-direction: column;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  z-index: 50;
}

.sidebar.collapsed {
  width: var(--sidebar-collapsed-width);
}

/* Mobile: sidebar becomes drawer overlay */
@media (max-width: 767px) {
  .sidebar {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    width: min(calc(100vw - var(--space-6)), 260px);
    height: 100vh;
    transform: translateX(-100%);
    z-index: 100;
    box-shadow: none;
    transition: transform 0.35s cubic-bezier(0.32, 0.72, 0, 1), box-shadow 0.35s ease;
    border-radius: 0 var(--radius-lg) var(--radius-lg) 0;
  }

  .sidebar.mobile-open {
    transform: translateX(0);
    box-shadow: -8px 0 48px rgba(0, 0, 0, 0.5);
  }

  .sidebar.mobile-closing {
    transform: translateX(-100%);
    box-shadow: none;
  }

  .sidebar.collapsed {
    width: min(calc(100vw - var(--space-6)), 260px);
  }
}

/* Tablet: 768px - 1023px */
@media (min-width: 768px) and (max-width: 1023px) {
  .sidebar {
    width: var(--sidebar-collapsed-width);
  }
}
</style>