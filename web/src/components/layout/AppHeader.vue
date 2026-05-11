<!-- web/src/components/layout/AppHeader.vue -->
<template>
  <header class="top-header">
    <!-- Mobile menu button - only visible on mobile -->
    <button @click="$emit('toggle-sidebar')" class="menu-toggle" :title="t('common.menu')">
      <Menu :size="20" />
    </button>

    <div class="header-left">
      <h1 class="page-title">{{ currentPageTitle }}</h1>
      <AppBreadcrumb class="header-breadcrumb" :crumbs="breadcrumbs" />
    </div>

    <div class="header-right">
      <div class="header-actions">
        <button @click="refreshData" class="btn btn-ghost" :title="t('common.refresh')">
          <RefreshCw :size="18" :class="{ spinning: refreshing }" />
        </button>
        <LanguageSwitcher />
        <div class="notification-dropdown" ref="notifRef">
          <button @click="toggleNotifications" class="btn btn-ghost" :title="t('common.notifications')">
            <Bell :size="18" />
            <span class="notification-badge" v-if="unreadCount > 0">{{ unreadCount }}</span>
          </button>
        </div>
      </div>

      <UserMenu />
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { t } from '@/i18n'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'
import AppBreadcrumb from '@/components/layout/AppBreadcrumb.vue'
import UserMenu from '@/components/layout/UserMenu.vue'
import { RefreshCw, Bell, Menu } from 'lucide-vue-next'

defineEmits<{
  (e: 'toggle-sidebar'): void
}>()

const route = useRoute()

const refreshing = ref(false)
const notificationsOpen = ref(false)
const unreadCount = ref(3)
const notifRef = ref<HTMLElement | null>(null)

// Simplified nav items — only name + label needed for title/breadcrumb resolution
const allNavItems = computed(() => [
  { name: 'Dashboard', label: t('nav.overview') },
  { name: 'Containers', label: t('nav.containers') },
  { name: 'Orchestration', label: t('nav.orchestration') },
  { name: 'Images', label: t('nav.images') },
  { name: 'Networks', label: t('nav.networks') },
  { name: 'Volumes', label: t('nav.volumes') },
  { name: 'Settings', label: t('nav.settings') },
])

const currentPageTitle = computed(() => {
  const item = allNavItems.value.find(item => item.name === route.name)
  return item?.label || 'LiteDock'
})

const breadcrumbs = computed(() => {
  const item = allNavItems.value.find(item => item.name === route.name)
  return item ? ['LiteDock', item.label] : ['LiteDock']
})

const refreshData = async () => {
  refreshing.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 1000))
  } finally {
    refreshing.value = false
  }
}

const toggleNotifications = () => {
  notificationsOpen.value = !notificationsOpen.value
}

const handleClickOutside = (event: Event) => {
  if (notifRef.value && !notifRef.value.contains(event.target as Node)) {
    notificationsOpen.value = false
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onUnmounted(() => document.removeEventListener('click', handleClickOutside))
</script>

<style scoped>
.top-header {
  height: var(--header-height);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--space-8);
  background: transparent;
  z-index: 10;
}

.menu-toggle {
  display: none;
  align-items: center;
  justify-content: center;
  background: none;
  border: none;
  color: var(--color-text-weak);
  cursor: pointer;
  padding: var(--space-2);
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}

.menu-toggle:hover {
  background: var(--color-background-weak);
  color: var(--color-text-strong);
}

.page-title {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
  margin: 0;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--space-4);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: var(--space-1);
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  cursor: pointer;
  font-size: var(--font-size-sm);
  transition: all var(--transition-fast);
}

.btn-ghost {
  background: none;
  color: var(--color-text-weak);
  padding: var(--space-2);
  border-radius: var(--radius-sm);
}

.btn-ghost:hover {
  background: var(--color-background-weak);
  color: var(--color-text-strong);
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.notification-dropdown {
  position: relative;
}

.notification-badge {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 16px;
  height: 16px;
  background: var(--color-error);
  color: #fdfcfc;
  border-radius: var(--radius-full);
  font-size: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

@media (max-width: 767px) {
  .top-header {
    padding: 0 var(--space-4);
  }

  .menu-toggle {
    display: flex;
  }

  .page-title {
    font-size: var(--font-size-sm);
  }

  .header-breadcrumb {
    display: none;
  }

  .header-actions .notification-dropdown {
    display: none;
  }
}

@media (min-width: 768px) and (max-width: 1023px) {
  .top-header {
    padding: 0 var(--space-5);
  }
}
</style>
