<template>
  <div class="main-layout">
    <aside class="sidebar" :class="{ collapsed: sidebarCollapsed }">
      <div class="sidebar-header">
        <div class="logo" v-if="!sidebarCollapsed">
          <img :src="logoUrl" alt="LiteDock" class="logo-img" />
        </div>
        <div class="logo-collapsed" v-else>
          <img :src="logoUrl" alt="LiteDock" class="logo-img-collapsed" />
        </div>
      </div>

      <nav class="sidebar-nav">
        <div class="nav-section">
          <div class="nav-section-title" v-if="!sidebarCollapsed">{{ t('nav.main') }}</div>
          <router-link
            v-for="item in mainNavItems"
            :key="item.name"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.name === item.name }"
          >
            <component :is="item.icon" :size="20" class="nav-icon" />
            <span class="nav-text" v-if="!sidebarCollapsed">{{ item.label }}</span>
            <span class="nav-badge" v-if="item.badge && !sidebarCollapsed">{{ item.badge }}</span>
          </router-link>
        </div>

        <div class="nav-section">
          <div class="nav-section-title" v-if="!sidebarCollapsed">{{ t('nav.system') }}</div>
          <router-link
            v-for="item in systemNavItems"
            :key="item.name"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.name === item.name }"
          >
            <component :is="item.icon" :size="20" class="nav-icon" />
            <span class="nav-text" v-if="!sidebarCollapsed">{{ item.label }}</span>
          </router-link>
        </div>
      </nav>

      <div class="sidebar-footer">
        <div class="user-info" v-if="!sidebarCollapsed">
          <div class="user-avatar">
            {{ authStore.user?.username?.charAt(0).toUpperCase() }}
          </div>
          <div class="user-details">
            <div class="user-name">{{ authStore.user?.username }}</div>
            <div class="user-role">{{ authStore.user?.role }}</div>
          </div>
        </div>

        <button @click="toggleSidebar" class="sidebar-toggle">
          <ChevronLeft v-if="!sidebarCollapsed" :size="18" />
          <ChevronRight v-if="sidebarCollapsed" :size="18" />
        </button>
      </div>
    </aside>

    <div class="main-content">
      <header class="top-header">
        <div class="header-left">
          <h1 class="page-title">{{ currentPageTitle }}</h1>
          <div class="breadcrumb">
            <span v-for="(crumb, index) in breadcrumbs" :key="index">
              {{ crumb }}
              <span v-if="index < breadcrumbs.length - 1" class="breadcrumb-separator">/</span>
            </span>
          </div>
        </div>

        <div class="header-right">
          <div class="header-actions">
            <button @click="refreshData" class="btn btn-ghost" :title="t('common.refresh')">
              <RefreshCw :size="18" :class="{ 'spinning': refreshing }" />
            </button>
            <button @click="toggleTheme" class="btn btn-ghost" :title="t('common.switchTheme')">
              <Sun v-if="isDarkMode" :size="18" />
              <Moon v-else :size="18" />
            </button>
            <LanguageSwitcher />
            <div class="notification-dropdown">
              <button @click="toggleNotifications" class="btn btn-ghost" :title="t('common.notifications')">
                <Bell :size="18" />
                <span class="notification-badge" v-if="unreadCount > 0">{{ unreadCount }}</span>
              </button>
            </div>
          </div>

          <div class="user-menu">
            <button @click="toggleUserMenu" class="user-menu-button">
              <div class="user-avatar small">
                {{ authStore.user?.username?.charAt(0).toUpperCase() }}
              </div>
              <span class="user-name">{{ authStore.user?.username }}</span>
              <ChevronDown :size="16" />
            </button>

            <div v-if="userMenuOpen" class="user-menu-dropdown">
              <a href="#" @click.prevent="goToSettings">
                <Settings :size="16" />
                {{ t('common.settings') }}
              </a>
              <a href="#" @click.prevent="handleLogout">
                <LogOut :size="16" />
                {{ t('common.logout') }}
              </a>
            </div>
          </div>
        </div>
      </header>

      <main class="page-content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { t } from '@/i18n'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'
import logoUrl from '@/assets/logo.svg?url'
import {
  LayoutDashboard,
  Box,
  GitBranch,
  Image,
  Network,
  HardDrive,
  Settings,
  ChevronLeft,
  ChevronRight,
  ChevronDown,
  RefreshCw,
  Bell,
  Sun,
  Moon,
  LogOut
} from 'lucide-vue-next'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const sidebarCollapsed = ref(false)
const refreshing = ref(false)
const userMenuOpen = ref(false)
const notificationsOpen = ref(false)
const unreadCount = ref(3)

const mainNavItems = computed(() => [
  { name: 'Dashboard', path: '/', label: t('nav.overview'), icon: LayoutDashboard },
  { name: 'Containers', path: '/containers', label: t('nav.containers'), icon: Box, badge: '12' },
  { name: 'Orchestration', path: '/orchestration', label: t('nav.orchestration'), icon: GitBranch },
  { name: 'Images', path: '/images', label: t('nav.images'), icon: Image },
  { name: 'Networks', path: '/networks', label: t('nav.networks'), icon: Network },
  { name: 'Volumes', path: '/volumes', label: t('nav.volumes'), icon: HardDrive }
])

const systemNavItems = computed(() => [
  { name: 'Settings', path: '/settings', label: t('nav.settings'), icon: Settings }
])

const currentPageTitle = computed(() => {
  const routeItem = [...mainNavItems.value, ...systemNavItems.value].find(item => item.name === route.name)
  return routeItem?.label || 'LiteDock'
})

const breadcrumbs = computed(() => {
  const routeItem = [...mainNavItems.value, ...systemNavItems.value].find(item => item.name === route.name)
  return routeItem ? ['LiteDock', routeItem.label] : ['LiteDock']
})

const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
  localStorage.setItem('litedock-sidebar-collapsed', sidebarCollapsed.value.toString())
}

const refreshData = async () => {
  refreshing.value = true
  try {
    await new Promise(resolve => setTimeout(resolve, 1000))
  } finally {
    refreshing.value = false
  }
}

const getInitialTheme = () => {
  const saved = localStorage.getItem('litedock-theme')
  if (saved) return saved === 'dark'
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

const isDarkMode = ref(getInitialTheme())

const toggleTheme = () => {
  isDarkMode.value = !isDarkMode.value
  document.documentElement.classList.toggle('dark', isDarkMode.value)
  localStorage.setItem('litedock-theme', isDarkMode.value ? 'dark' : 'light')
}

const toggleNotifications = () => {
  notificationsOpen.value = !notificationsOpen.value
  userMenuOpen.value = false
}

const toggleUserMenu = () => {
  userMenuOpen.value = !userMenuOpen.value
  notificationsOpen.value = false
}

const goToSettings = () => {
  userMenuOpen.value = false
  router.push('/settings')
}

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}

const handleClickOutside = (event: Event) => {
  const target = event.target as HTMLElement
  if (!target.closest('.user-menu') && !target.closest('.notification-dropdown')) {
    userMenuOpen.value = false
    notificationsOpen.value = false
  }
}

onMounted(() => {
  const savedCollapsed = localStorage.getItem('litedock-sidebar-collapsed')
  if (savedCollapsed) {
    sidebarCollapsed.value = savedCollapsed === 'true'
  }

  if (isDarkMode.value) {
    document.documentElement.classList.add('dark')
  }

  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.main-layout {
  display: flex;
  height: 100vh;
  background: var(--color-background);
}

.sidebar {
  width: var(--sidebar-width);
  background: var(--color-background);
  border-right: 1px solid var(--color-border-weak);
  display: flex;
  flex-direction: column;
  transition: width var(--transition-base);
}

.sidebar.collapsed {
  width: var(--sidebar-collapsed-width);
}

.sidebar-header {
  padding: var(--space-4) var(--space-4);
  border-bottom: 1px solid var(--color-border-weak);
  height: var(--header-height);
  display: flex;
  align-items: center;
}

.logo {
  display: flex;
  align-items: center;
}

.logo-img {
  height: 28px;
  width: auto;
}

.logo-collapsed {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
}

.logo-img-collapsed {
  height: 24px;
  width: auto;
}

.sidebar-nav {
  flex: 1;
  padding: var(--space-3) 0;
  overflow-y: auto;
}

.nav-section {
  margin-bottom: var(--space-4);
}

.nav-section-title {
  padding: var(--space-2) var(--space-4);
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-weaker);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-4);
  color: var(--color-text-weak);
  text-decoration: none;
  font-size: var(--font-size-sm);
  transition: all var(--transition-fast);
  border-left: 2px solid transparent;
  margin: 1px 0;
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
}

.sidebar-footer {
  padding: var(--space-3) var(--space-4);
  border-top: 1px solid var(--color-border-weak);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.user-info {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  flex: 1;
  min-width: 0;
}

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: var(--radius-full);
  background: var(--color-background-strong);
  color: var(--color-background);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: var(--font-weight-semibold);
  font-size: var(--font-size-sm);
  flex-shrink: 0;
}

.user-avatar.small {
  width: 28px;
  height: 28px;
  font-size: var(--font-size-xs);
}

.user-details {
  min-width: 0;
}

.user-name {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-strong);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-role {
  font-size: var(--font-size-xs);
  color: var(--color-text-weaker);
}

.sidebar-toggle {
  background: none;
  border: none;
  color: var(--color-text-weak);
  cursor: pointer;
  padding: var(--space-2);
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}

.sidebar-toggle:hover {
  background: var(--color-background-hover);
  color: var(--color-text-strong);
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.top-header {
  height: var(--header-height);
  border-bottom: 1px solid var(--color-border-weak);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--space-6);
  background: var(--color-background);
}

.page-title {
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-strong);
  margin: 0;
}

.breadcrumb {
  font-size: var(--font-size-xs);
  color: var(--color-text-weaker);
  margin-top: 2px;
}

.breadcrumb-separator {
  margin: 0 var(--space-1);
  color: var(--color-border);
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
  background: var(--color-background-hover);
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
  color: white;
  border-radius: var(--radius-full);
  font-size: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-menu {
  position: relative;
}

.user-menu-button {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  background: none;
  border: none;
  cursor: pointer;
  padding: var(--space-1) var(--space-2);
  border-radius: var(--radius-sm);
  color: var(--color-text);
  transition: all var(--transition-fast);
}

.user-menu-button:hover {
  background: var(--color-background-hover);
}

.user-menu-dropdown {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: var(--space-2);
  background: var(--color-background);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  min-width: 160px;
  z-index: 50;
  padding: var(--space-1);
}

.user-menu-dropdown a {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  padding: var(--space-2) var(--space-3);
  color: var(--color-text);
  text-decoration: none;
  font-size: var(--font-size-sm);
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}

.user-menu-dropdown a:hover {
  background: var(--color-background-hover);
}

.page-content {
  flex: 1;
  padding: var(--space-6);
  overflow-y: auto;
}

@media (max-width: 767px) {
  .sidebar {
    display: none;
  }

  .top-header {
    padding: 0 var(--space-4);
  }

  .page-content {
    padding: var(--space-4);
  }
}

@media (min-width: 768px) and (max-width: 1023px) {
  .sidebar {
    width: var(--sidebar-collapsed-width) !important;
  }

  .sidebar-header,
  .nav-section-title,
  .nav-text,
  .nav-badge,
  .user-info,
  .user-name,
  .user-role {
    display: none !important;
  }

  .logo-collapsed {
    display: flex !important;
  }

  .sidebar-toggle {
    padding: var(--space-1);
  }

  .nav-item {
    padding: var(--space-3);
    justify-content: center;
  }

  .top-header {
    padding: 0 var(--space-5);
  }
}
</style>
