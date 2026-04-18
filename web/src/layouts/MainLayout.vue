<template>
  <div class="main-layout">
    <!-- 侧边栏 -->
    <aside class="sidebar" :class="{ collapsed: sidebarCollapsed }">
      <div class="sidebar-header">
        <div class="logo" v-if="!sidebarCollapsed">
          <img src="/src/assets/logo.svg" alt="LiteDock" v-if="hasLogo" />
          <span class="logo-text" v-else>LiteDock</span>
        </div>
        <div class="logo-collapsed" v-else>
          <span>LD</span>
        </div>
      </div>
      
      <nav class="sidebar-nav">
        <div class="nav-section">
          <div class="nav-section-title" v-if="!sidebarCollapsed">主要功能</div>
          <router-link
            v-for="item in mainNavItems"
            :key="item.name"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.name === item.name }"
          >
            <component :is="item.icon" class="nav-icon" />
            <span class="nav-text" v-if="!sidebarCollapsed">{{ item.label }}</span>
            <span class="nav-badge" v-if="item.badge && !sidebarCollapsed">{{ item.badge }}</span>
          </router-link>
        </div>
        
        <div class="nav-section">
          <div class="nav-section-title" v-if="!sidebarCollapsed">系统</div>
          <router-link
            v-for="item in systemNavItems"
            :key="item.name"
            :to="item.path"
            class="nav-item"
            :class="{ active: $route.name === item.name }"
          >
            <component :is="item.icon" class="nav-icon" />
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
          <ChevronLeft v-if="!sidebarCollapsed" />
          <ChevronRight v-if="sidebarCollapsed" />
        </button>
      </div>
    </aside>
    
    <!-- 主内容区 -->
    <div class="main-content">
      <!-- 顶部导航栏 -->
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
            <button @click="refreshData" class="action-btn" title="刷新">
              <RefreshCw :class="{ 'spinning': refreshing }" />
            </button>
            <button @click="toggleTheme" class="action-btn" title="切换主题">
              <Sun v-if="isDarkMode" />
              <Moon v-else />
            </button>
            <div class="notification-dropdown">
              <button @click="toggleNotifications" class="action-btn" title="通知">
                <Bell />
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
              <ChevronDown />
            </button>
            
            <div v-if="userMenuOpen" class="user-menu-dropdown">
              <a href="#" @click.prevent="goToSettings">
                <Settings />
                设置
              </a>
              <a href="#" @click.prevent="handleLogout">
                <LogOut />
                退出登录
              </a>
            </div>
          </div>
        </div>
      </header>
      
      <!-- 页面内容 -->
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

// 图标组件 (这里使用简单的SVG，实际项目中可以使用图标库)
const DashboardIcon = '📊'
const ContainerIcon = '📦'
const OrchestrationIcon = '🔀'
const ImageIcon = '🖼️'
const NetworkIcon = '🌐'
const VolumeIcon = '💾'
const SettingsIcon = '⚙️'
const ChevronLeft = '◀'
const ChevronRight = '▶'
const ChevronDown = '▼'
const RefreshCw = '🔄'
const Bell = '🔔'
const Sun = '☀️'
const Moon = '🌙'
const LogOut = '🚪'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const sidebarCollapsed = ref(false)
const refreshing = ref(false)
const isDarkMode = ref(false)
const userMenuOpen = ref(false)
const notificationsOpen = ref(false)
const unreadCount = ref(3)
const hasLogo = ref(true) // Logo file now exists

const mainNavItems = [
  { name: 'Dashboard', path: '/', label: '概览', icon: DashboardIcon },
  { name: 'Containers', path: '/containers', label: '容器', icon: ContainerIcon, badge: '12' },
  { name: 'Orchestration', path: '/orchestration', label: '编排', icon: OrchestrationIcon },
  { name: 'Images', path: '/images', label: '镜像', icon: ImageIcon },
  { name: 'Networks', path: '/networks', label: '网络', icon: NetworkIcon },
  { name: 'Volumes', path: '/volumes', label: '存储卷', icon: VolumeIcon }
]

const systemNavItems = [
  { name: 'Settings', path: '/settings', label: '设置', icon: SettingsIcon }
]

const currentPageTitle = computed(() => {
  const routeItem = [...mainNavItems, ...systemNavItems].find(item => item.name === route.name)
  return routeItem?.label || 'LiteDock'
})

const breadcrumbs = computed(() => {
  const routeItem = [...mainNavItems, ...systemNavItems].find(item => item.name === route.name)
  return routeItem ? ['LiteDock', routeItem.label] : ['LiteDock']
})

const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
  localStorage.setItem('litedock-sidebar-collapsed', sidebarCollapsed.value.toString())
}

const refreshData = async () => {
  refreshing.value = true
  try {
    // 这里可以添加刷新数据的逻辑
    await new Promise(resolve => setTimeout(resolve, 1000))
  } finally {
    refreshing.value = false
  }
}

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

// 点击外部关闭菜单
const handleClickOutside = (event: Event) => {
  const target = event.target as HTMLElement
  if (!target.closest('.user-menu') && !target.closest('.notification-dropdown')) {
    userMenuOpen.value = false
    notificationsOpen.value = false
  }
}

onMounted(() => {
  // 恢复侧边栏状态
  const savedCollapsed = localStorage.getItem('litedock-sidebar-collapsed')
  if (savedCollapsed) {
    sidebarCollapsed.value = savedCollapsed === 'true'
  }
  
  // 恢复主题状态
  const savedTheme = localStorage.getItem('litedock-theme')
  if (savedTheme) {
    isDarkMode.value = savedTheme === 'dark'
    document.documentElement.classList.toggle('dark', isDarkMode.value)
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
  background: #f5f5f5;
}

.sidebar {
  width: 260px;
  background: white;
  border-right: 1px solid #d4d4d4;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
}

.sidebar.collapsed {
  width: 60px;
}

.sidebar-header {
  padding: 20px;
  border-bottom: 1px solid #d4d4d4;
}

.logo {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logo img {
  width: 32px;
  height: 32px;
}

.logo-text {
  font-size: 1.25rem;
  font-weight: 600;
  color: #000000;
}

.logo-collapsed {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #000000 0%, #333333 100%);
  color: white;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.875rem;
}

.sidebar-nav {
  flex: 1;
  padding: 16px 0;
  overflow-y: auto;
}

.nav-section {
  margin-bottom: 24px;
}

.nav-section-title {
  padding: 0 20px 8px 20px;
  font-size: 0.75rem;
  font-weight: 600;
  color: #404040;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 20px;
  color: #404040;
  text-decoration: none;
  transition: all 0.2s;
  position: relative;
}

.nav-item:hover {
  background: #f5f5f5;
  color: #000000;
}

.nav-item.active {
  background: #e5e5e5;
  color: #000000;
}

.nav-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: #000000;
}

.nav-icon {
  font-size: 1.25rem;
  width: 20px;
  text-align: center;
}

.nav-text {
  flex: 1;
  font-weight: 500;
}

.nav-badge {
  background: #000000;
  color: white;
  font-size: 0.75rem;
  padding: 2px 6px;
  border-radius: 10px;
  font-weight: 600;
}

.sidebar-footer {
  padding: 16px;
  border-top: 1px solid #d4d4d4;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.user-avatar {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #000000 0%, #333333 100%);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 0.875rem;
}

.user-avatar.small {
  width: 32px;
  height: 32px;
  font-size: 0.75rem;
}

.user-details {
  flex: 1;
  min-width: 0;
}

.user-name {
  font-weight: 600;
  color: #000000;
  font-size: 0.875rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-role {
  color: #404040;
  font-size: 0.75rem;
}

.sidebar-toggle {
  width: 100%;
  padding: 8px;
  background: #f5f5f5;
  border: 1px solid #d4d4d4;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.sidebar-toggle:hover {
  background: #e5e5e5;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.top-header {
  height: 70px;
  background: white;
  border-bottom: 1px solid #d4d4d4;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.page-title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
  color: #000000;
}

.breadcrumb {
  font-size: 0.875rem;
  color: #404040;
}

.breadcrumb-separator {
  margin: 0 8px;
  color: #a3a3a3;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-btn {
  width: 40px;
  height: 40px;
  background: #f5f5f5;
  border: 1px solid #d4d4d4;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.125rem;
  position: relative;
}

.action-btn:hover {
  background: #e5e5e5;
  border-color: #a3a3a3;
}

.spinning {
  animation: spin 1s linear infinite;
}

.notification-badge {
  position: absolute;
  top: -4px;
  right: -4px;
  background: #000000;
  color: white;
  font-size: 0.625rem;
  padding: 2px 4px;
  border-radius: 8px;
  font-weight: 600;
  min-width: 16px;
  text-align: center;
}

.user-menu {
  position: relative;
}

.user-menu-button {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #f5f5f5;
  border: 1px solid #d4d4d4;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.user-menu-button:hover {
  background: #e5e5e5;
  border-color: #a3a3a3;
}

.user-menu-dropdown {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 8px;
  background: white;
  border: 1px solid #d4d4d4;
  border-radius: 8px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
  min-width: 180px;
  z-index: 1000;
}

.user-menu-dropdown a {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  color: #000000;
  text-decoration: none;
  transition: background 0.2s;
}

.user-menu-dropdown a:hover {
  background: #f5f5f5;
}

.user-menu-dropdown a:first-child {
  border-radius: 8px 8px 0 0;
}

.user-menu-dropdown a:last-child {
  border-radius: 0 0 8px 8px;
  border-top: 1px solid #e5e5e5;
}

.page-content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* 暗色主题 */
:global(.dark) .main-layout {
  background: #000000;
}

:global(.dark) .sidebar {
  background: #1a1a1a;
  border-color: #404040;
}

:global(.dark) .sidebar-header {
  border-color: #404040;
}

:global(.dark) .logo-text {
  color: #ffffff;
}

:global(.dark) .nav-section-title {
  color: #a3a3a3;
}

:global(.dark) .nav-item {
  color: #d4d4d4;
}

:global(.dark) .nav-item:hover {
  background: #2e2e2e;
  color: #ffffff;
}

:global(.dark) .nav-item.active {
  background: #000000;
  color: #ffffff;
}

:global(.dark) .sidebar-footer {
  border-color: #404040;
}

:global(.dark) .user-name {
  color: #ffffff;
}

:global(.dark) .user-role {
  color: #a3a3a3;
}

:global(.dark) .sidebar-toggle {
  background: #2e2e2e;
  border-color: #404040;
}

:global(.dark) .sidebar-toggle:hover {
  background: #404040;
}

:global(.dark) .top-header {
  background: #1a1a1a;
  border-color: #404040;
}

:global(.dark) .page-title {
  color: #ffffff;
}

:global(.dark) .breadcrumb {
  color: #a3a3a3;
}

:global(.dark) .action-btn {
  background: #2e2e2e;
  border-color: #404040;
}

:global(.dark) .action-btn:hover {
  background: #404040;
  border-color: #737373;
}

:global(.dark) .user-menu-button {
  background: #2e2e2e;
  border-color: #404040;
}

:global(.dark) .user-menu-button:hover {
  background: #404040;
  border-color: #737373;
}

:global(.dark) .user-menu-dropdown {
  background: #1a1a1a;
  border-color: #404040;
}

:global(.dark) .user-menu-dropdown a {
  color: #ffffff;
}

:global(.dark) .user-menu-dropdown a:hover {
  background: #2e2e2e;
}

:global(.dark) .user-menu-dropdown a:last-child {
  border-color: #404040;
}

/* Mobile Responsive (< 768px) */
@media (max-width: 767px) {
  .sidebar {
    display: none;
  }
  
  .main-content {
    padding-bottom: 70px;
  }
  
  .top-header {
    padding: 0 16px;
  }
  
  .page-title {
    font-size: 1.25rem;
  }
  
  .header-right {
    gap: 8px;
  }
  
  .user-name {
    display: none;
  }
  
  .page-content {
    padding: 16px;
  }
}

/* Tablet Responsive (768px - 1023px) */
@media (min-width: 768px) and (max-width: 1023px) {
  .sidebar {
    width: 60px !important;
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
    padding: 4px;
  }
  
  .nav-item {
    padding: 12px 16px;
    justify-content: center;
  }
  
  .top-header {
    padding: 0 20px;
  }
}

/* Desktop Responsive (>= 1024px) */
@media (min-width: 1024px) {
  .sidebar {
    width: 260px;
  }
  
  .sidebar.collapsed {
    width: 60px;
  }
}
</style>