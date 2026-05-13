<!-- web/src/layouts/MainLayout.vue -->
<template>
  <div class="main-layout">
    <!-- Mobile backdrop overlay -->
    <div
      class="sidebar-backdrop"
      :class="{ visible: isMobile && mobileSidebarOpen, closing: isClosing }"
      @click="closeMobileSidebar"
    ></div>

    <!-- Sidebar - always rendered -->
    <div class="sidebar-wrapper">
      <Sidebar :mobile-open="mobileSidebarOpen" @toggle="handleSidebarToggle" />
    </div>

    <!-- Main content -->
    <div class="main-content">
      <AppHeader @toggle-sidebar="toggleMobileSidebar" />
      <main class="page-content">
        <router-view />
      </main>
    </div>

    <!-- AI Assistant (floating, position: fixed) -->
    <AssistantChat />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import Sidebar from '@/components/layout/Sidebar.vue'
import AppHeader from '@/components/layout/AppHeader.vue'
import AssistantChat from '@/components/ui/AssistantChat.vue'
import { useViewport } from '@/composables/useViewport'

// Mobile detection
const { isMobile } = useViewport()
const mobileSidebarOpen = ref(false)
const isClosing = ref(false)
let closeTimeout: ReturnType<typeof setTimeout> | null = null

onUnmounted(() => {
  if (closeTimeout) clearTimeout(closeTimeout)
})

const toggleMobileSidebar = () => {
  if (mobileSidebarOpen.value) {
    closeMobileSidebar()
  } else {
    mobileSidebarOpen.value = true
  }
}

const closeMobileSidebar = () => {
  if (!mobileSidebarOpen.value) return
  isClosing.value = true
  if (closeTimeout) clearTimeout(closeTimeout)
  closeTimeout = setTimeout(() => {
    mobileSidebarOpen.value = false
    isClosing.value = false
    closeTimeout = null
  }, 350)
}

const handleSidebarToggle = () => {
  // Sidebar handles its own animation, just sync closing state
  if (isMobile.value) {
    closeMobileSidebar()
  }
}

// Watch for rapid open/close to reset closing state
watch(mobileSidebarOpen, (open) => {
  if (open && isClosing.value) {
    isClosing.value = false
    if (closeTimeout) {
      clearTimeout(closeTimeout)
      closeTimeout = null
    }
  }
})
</script>

<style scoped>
.main-layout {
  display: flex;
  height: 100vh;
  background: var(--color-background);
  position: relative;
}

.sidebar-wrapper {
  flex-shrink: 0;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.page-content {
  flex: 1;
  padding: var(--space-6);
  overflow-y: auto;
}

/* Mobile backdrop */
.sidebar-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(32, 29, 29, 0.4);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  z-index: 40;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.35s cubic-bezier(0.32, 0.72, 0, 1);
}

.sidebar-backdrop.visible {
  opacity: 1;
  pointer-events: auto;
}

.sidebar-backdrop.closing {
  opacity: 0;
}

@media (max-width: 767px) {
  .page-content {
    padding: var(--space-4);
  }
}
</style>
