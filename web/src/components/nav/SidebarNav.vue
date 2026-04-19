<!-- web/src/components/nav/SidebarNav.vue -->
<template>
  <nav class="sidebar-nav" :class="{ collapsed }">
    <div class="nav-section">
      <div class="nav-section-header">
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
        :badge="item.badge"
      />
    </div>

    <div class="nav-section">
      <div class="nav-section-header">
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
      />
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { t } from '@/i18n'
import NavItem from '@/components/nav/NavItem.vue'
import {
  LayoutDashboard,
  Box,
  GitBranch,
  Image,
  Network,
  HardDrive,
  Settings,
} from 'lucide-vue-next'

defineProps<{
  collapsed: boolean
}>()

const route = useRoute()

const currentRouteName = computed(() => route.name)

const mainNavItems = computed(() => [
  { name: 'Dashboard', path: '/', label: t('nav.overview'), icon: LayoutDashboard },
  { name: 'Containers', path: '/containers', label: t('nav.containers'), icon: Box, badge: '12' },
  { name: 'Orchestration', path: '/orchestration', label: t('nav.orchestration'), icon: GitBranch },
  { name: 'Images', path: '/images', label: t('nav.images'), icon: Image },
  { name: 'Networks', path: '/networks', label: t('nav.networks'), icon: Network },
  { name: 'Volumes', path: '/volumes', label: t('nav.volumes'), icon: HardDrive },
])

const systemNavItems = computed(() => [
  { name: 'Settings', path: '/settings', label: t('nav.settings'), icon: Settings },
])
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

/* Collapsed state - title fades out but space remains */
.sidebar-nav.collapsed .nav-section-title {
  opacity: 0;
}
</style>
