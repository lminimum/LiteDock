<!-- web/src/components/nav/SidebarNav.vue -->
<template>
  <nav class="sidebar-nav" :class="{ collapsed, 'mobile-mode': isMobile }">
    <div class="nav-section">
      <div class="nav-section-header" v-if="!isMobile && !collapsed">
        <div class="nav-section-title">{{ t('nav.main') }}</div>
        <div class="nav-section-spacer"></div>
      </div>
      <NavItem
        v-for="item in mainNavItems"
        :key="item.name"
        :to="item.path"
        :icon="item.icon"
        :label="item.label"
        :active="currentRouteName === item.name"
        :collapsed="collapsed"
        :is-mobile="isMobile"
        :children="item.children"
        :expanded="expandedMenus[item.name]"
        @toggle-children="toggleMenu(item.name)"
      />
    </div>

    <div class="nav-section">
      <div class="nav-section-header" v-if="!isMobile && !collapsed">
        <div class="nav-section-title">{{ t('nav.system') }}</div>
        <div class="nav-section-spacer"></div>
      </div>
      <NavItem
        v-for="item in systemNavItems"
        :key="item.name"
        :to="item.path"
        :icon="item.icon"
        :label="item.label"
        :active="currentRouteName === item.name"
        :collapsed="collapsed"
        :is-mobile="isMobile"
        :children="item.children"
        :expanded="expandedMenus[item.name]"
        @toggle-children="toggleMenu(item.name)"
      />
    </div>

    <div class="nav-section theme-section">
      <button
        class="nav-item theme-toggle"
        :class="{ collapsed, 'mobile-mode': isMobile }"
        @click="toggleTheme"
        :title="t('theme.toggle')"
      >
        <component :is="themeIcon" :size="20" class="nav-icon" />
        <span class="nav-text">{{ themeLabel }}</span>
      </button>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { t } from '@/i18n'
import NavItem from '@/components/nav/NavItem.vue'
import {
  LayoutDashboard,
  Box,
  Settings,
  Server,
  Container,
  GitBranch,
  Image as ImageIcon,
  HardDrive,
  Globe,
  Bot,
  Sun,
  Moon,
} from 'lucide-vue-next'
import { useTheme } from '@/themes'

import type { Component } from 'vue'

interface NavChildItem {
  name: string
  path: string
  label: string
  icon?: Component
}

interface NavItemDef {
  name: string
  path?: string
  label: string
  icon: Component
  badge?: string
  children?: NavChildItem[]
}

defineProps<{
  collapsed: boolean
  isMobile: boolean
}>()

defineEmits<{
  (e: 'navigate'): void
}>()

const route = useRoute()
const containerCount = ref<string | undefined>(undefined)
const { currentTheme, toggleTheme: toggleThemeFn } = useTheme()

const currentRouteName = computed(() => route.name)

const themeIcon = computed(() => (currentTheme.value === 'light' ? Moon : Sun))
const themeLabel = computed(() =>
  currentTheme.value === 'light' ? t('theme.light') : t('theme.dark'),
)

const toggleTheme = () => {
  toggleThemeFn()
}

// Track which menus are expanded
const expandedMenus = ref<Record<string, boolean>>({})
const parentMenuNames = ['Docker', 'Infrastructure']

// Default expand all menus
onMounted(() => {
  parentMenuNames.forEach(name => {
    expandedMenus.value[name] = true
  })
})

const toggleMenu = (name: string) => {
  expandedMenus.value[name] = !expandedMenus.value[name]
}

const mainNavItems = computed<NavItemDef[]>(() => [
  { name: 'Dashboard', path: '/', label: t('nav.overview'), icon: LayoutDashboard },
  {
    name: 'Docker',
    label: t('nav.docker'),
    icon: Box,
    badge: containerCount.value,
    children: [
      { name: 'Containers', path: '/containers', label: t('nav.containers'), icon: Container },
      { name: 'Orchestration', path: '/orchestration', label: t('nav.orchestration'), icon: GitBranch },
      { name: 'Images', path: '/images', label: t('nav.images'), icon: ImageIcon },
      { name: 'Volumes', path: '/volumes', label: t('nav.volumes'), icon: HardDrive },
      { name: 'Networks', path: '/networks', label: t('nav.networks'), icon: Globe },
    ],
  },
  {
    name: 'Infrastructure',
    label: t('nav.infrastructure'),
    icon: Server,
    children: [
      { name: 'Machines', path: '/machines', label: t('nav.machines'), icon: Server },
    ],
  },
])

const systemNavItems = computed<NavItemDef[]>(() => [
  { name: 'Settings', path: '/settings', label: t('nav.settings'), icon: Settings },
  { name: 'AI', path: '/ai', label: t('nav.ai'), icon: Bot },
])

onMounted(async () => {
  try {
    const { default: api } = await import('@/utils/api')
    const data: any = await api.get('/dashboard/stats')
    if (data?.containers) {
      containerCount.value = String(data.containers.total || 0)
    }
  } catch (e) {
    console.error('Failed to fetch container count:', e)
  }
})
</script>

<style scoped>
.sidebar-nav {
  flex: 1;
  padding: var(--space-3) 0;
  overflow-y: auto;
  overflow-x: hidden;
}

.nav-section {
  margin-bottom: var(--space-4);
}

.nav-section-header {
  display: flex;
  align-items: center;
  height: 24px;
  margin: 1px 0;
  padding: 0 var(--space-4);
  box-sizing: border-box;
}

.nav-section-title {
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-weaker);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  white-space: nowrap;
  opacity: 1;
  transition: opacity 0.2s ease;
  line-height: 1;
}

.nav-section-spacer {
  flex: 1;
}

/* Mobile mode - center items */
@media (max-width: 767px) {
  .sidebar-nav.mobile-mode {
    padding: var(--space-4) 0;
  }

  .sidebar-nav.mobile-mode .nav-section {
    margin-bottom: var(--space-6);
  }

  .sidebar-nav.mobile-mode .nav-section:first-child {
    margin-top: var(--space-4);
  }
}

/* Theme toggle */
.theme-section {
  margin-top: auto;
  padding-top: var(--space-3);
  border-top: 1px solid var(--color-border-weak);
}

.theme-toggle {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-4);
  color: var(--color-text-weak);
  font-size: var(--font-size-sm);
  cursor: pointer;
  user-select: none;
  width: 100%;
  text-align: left;
  background: none;
  border: 2px solid transparent;
  border-left-width: 0;
  border-right-width: 0;
  font-family: inherit;
  transition: color 0.15s ease, background-color 0.15s ease;
  margin: 1px 0;
  box-sizing: border-box;
}

.theme-toggle:hover {
  color: var(--color-text-strong);
  background: var(--color-background-weak);
}

.theme-toggle.collapsed {
  justify-content: center;
  padding: var(--space-2);
  border-left: 2px solid transparent;
  border-right: 2px solid transparent;
}

.theme-toggle.collapsed .nav-text {
  display: none;
}

.theme-toggle.mobile-mode {
  justify-content: flex-start;
  padding: var(--space-3) var(--space-6);
  margin: var(--space-1) var(--space-4);
  border-radius: var(--radius-sm);
  border: none;
}

.theme-toggle.mobile-mode:hover {
  background: var(--color-background-weak);
}
</style>