# MainLayout.vue Component Extraction

**Goal:** Decompose the 649-line `MainLayout.vue` into focused, reusable components with shared composables, preserving all existing functionality and styles.

**Architecture:** Extract state management into composables (`useSidebar`, `useTheme`), create reusable UI primitives (`UserAvatar`, `NavItem`), compose them into layout components (`Sidebar`, `AppHeader`), and slim `MainLayout` to a ~50-line orchestrator. Props-down data flow: `Sidebar` calls `useSidebar()` and passes `collapsed` as prop to all children.

**Tech Stack:** Vue 3.5 `<script setup lang="ts">`, Pinia 3, vue-router 4, vue-i18n 11, lucide-vue-next 1.0, CSS variables from `style.css`

---

## Dependency Graph

```
Wave 1 (no deps — fully parallel)
├── Task 1:  composables/useSidebar.ts      ← standalone composable
├── Task 2:  composables/useTheme.ts        ← standalone composable
├── Task 3:  components/ui/UserAvatar.vue   ← standalone UI primitive
└── Task 4:  components/nav/NavItem.vue     ← standalone UI primitive

Wave 2 (depends on Wave 1 — fully parallel)
├── Task 5:  components/layout/SidebarLogo.vue
├── Task 6:  components/nav/SidebarNav.vue
├── Task 7:  components/layout/SidebarUserInfo.vue
├── Task 8:  components/layout/AppBreadcrumb.vue
└── Task 9:  components/layout/UserMenu.vue

Wave 3 (depends on Waves 1+2 — parallel)
├── Task 10: components/layout/Sidebar.vue
└── Task 11: components/layout/AppHeader.vue

Wave 4 (depends on Wave 3 — sequential)
├── Task 12: layouts/MainLayout.vue (rewrite)
└── Task 13: Visual verification + build check
```

---

## Shared State Decisions

| State | Location | Access Pattern |
|-------|----------|---------------|
| `sidebarCollapsed` | `useSidebar()` composable | Module-level singleton ref + `toggle()`, localStorage persist |
| `isDarkMode` | `useTheme()` composable | Module-level singleton ref + `toggle()`, localStorage persist, syncs `<html>` class |
| `refreshing` | `AppHeader.vue` local | Header-owned |
| `userMenuOpen` | `UserMenu.vue` local | Self-contained with click-outside |
| `notificationsOpen` | `AppHeader.vue` local | Header-local |
| `unreadCount` | `AppHeader.vue` local | Hardcoded 3 (future: from API) |
| `mainNavItems` / `systemNavItems` | `SidebarNav.vue` local computed | Nav definitions live in nav component |
| `currentPageTitle` / `breadcrumbs` | `AppHeader.vue` local computed | Derived from route |

---

## Codebase Conventions Reference

- All components: `<script setup lang="ts">` with `<style scoped>`
- Props: `defineProps<{ prop: Type }>()` or `withDefaults(defineProps<T>(), {})`
- Emits: `defineEmits<{ (e: 'name'): void }>()`
- Icons: Named imports from `lucide-vue-next`, used as `<Icon :size="18" />`
- i18n: `import { t } from '@/i18n'` then `t('key.subkey')`
- Path alias: `@/` maps to `./src/`
- Auth store: `useAuthStore()` → `user: { id, username, email?, role }`
- CSS variables: `--color-*`, `--space-*`, `--font-*`, `--radius-*`, `--transition-*`

---

## Tasks

### Task 1: Create `useSidebar` composable

**Wave:** 1 (no dependencies)

**Files:**
- Create: `web/src/composables/useSidebar.ts`

**Step 1: Create directory and file**

```typescript
// web/src/composables/useSidebar.ts
import { ref } from 'vue'

const STORAGE_KEY = 'litedock-sidebar-collapsed'

const collapsed = ref(false)

// Restore from localStorage on first import
const saved = localStorage.getItem(STORAGE_KEY)
if (saved !== null) {
  collapsed.value = saved === 'true'
}

export function useSidebar() {
  const toggle = () => {
    collapsed.value = !collapsed.value
    localStorage.setItem(STORAGE_KEY, collapsed.value.toString())
  }

  const collapse = () => {
    collapsed.value = true
    localStorage.setItem(STORAGE_KEY, 'true')
  }

  const expand = () => {
    collapsed.value = false
    localStorage.setItem(STORAGE_KEY, 'false')
  }

  return {
    collapsed,
    toggle,
    collapse,
    expand,
  }
}
```

**Design notes:**
- Module-level `ref` = singleton. Every call to `useSidebar()` shares the same reactive state.
- localStorage restoration at module load (not `onMounted`) — this is a SPA, no SSR concerns.
- `collapse()` / `expand()` are convenience methods for programmatic control.

**Step 2: Verify**

Run: `cd /data/project/LiteDock/web && npx vue-tsc --noEmit`
Expected: No type errors.

**Step 3: Commit**

```bash
git add web/src/composables/useSidebar.ts
git commit -m "feat(frontend): add useSidebar composable for shared sidebar state"
```

---

### Task 2: Create `useTheme` composable

**Wave:** 1 (no dependencies)

**Files:**
- Create: `web/src/composables/useTheme.ts`

**Step 1: Create file**

```typescript
// web/src/composables/useTheme.ts
import { ref } from 'vue'

const STORAGE_KEY = 'litedock-theme'

function getInitialTheme(): boolean {
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved) return saved === 'dark'
  return window.matchMedia('(prefers-color-scheme: dark)').matches
}

const isDarkMode = ref(getInitialTheme())

// Apply theme on first import — prevents flash of wrong theme
if (isDarkMode.value) {
  document.documentElement.classList.add('dark')
}

export function useTheme() {
  const toggle = () => {
    isDarkMode.value = !isDarkMode.value
    document.documentElement.classList.toggle('dark', isDarkMode.value)
    localStorage.setItem(STORAGE_KEY, isDarkMode.value ? 'dark' : 'light')
  }

  return {
    isDarkMode,
    toggle,
  }
}
```

**Step 2: Verify**

Run: `cd /data/project/LiteDock/web && npx vue-tsc --noEmit`
Expected: No type errors.

**Step 3: Commit**

```bash
git add web/src/composables/useTheme.ts
git commit -m "feat(frontend): add useTheme composable for dark mode toggle"
```

---

### Task 3: Create `UserAvatar` component

**Wave:** 1 (no dependencies)

**Files:**
- Create: `web/src/components/ui/UserAvatar.vue`

**Step 1: Create file**

```vue
<!-- web/src/components/ui/UserAvatar.vue -->
<template>
  <div class="user-avatar" :class="[sizeClass]">
    {{ initial }}
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  username?: string
  size?: 'sm' | 'md'
}>(), {
  username: '',
  size: 'md',
})

const initial = computed(() =>
  props.username ? props.username.charAt(0).toUpperCase() : '?'
)

const sizeClass = computed(() => `avatar-${props.size}`)
</script>

<style scoped>
.user-avatar {
  border-radius: var(--radius-full);
  background: var(--color-background-strong);
  color: var(--color-background);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: var(--font-weight-semibold);
  flex-shrink: 0;
}

.avatar-md {
  width: 32px;
  height: 32px;
  font-size: var(--font-size-sm);
}

.avatar-sm {
  width: 28px;
  height: 28px;
  font-size: var(--font-size-xs);
}
</style>
```

**Step 2: Verify**

Run: `cd /data/project/LiteDock/web && npx vue-tsc --noEmit`
Expected: No type errors.

**Step 3: Commit**

```bash
git add web/src/components/ui/UserAvatar.vue
git commit -m "feat(frontend): add UserAvatar reusable UI component"
```

---

### Task 4: Create `NavItem` component

**Wave:** 1 (no dependencies)

**Files:**
- Create: `web/src/components/nav/NavItem.vue`

**Step 1: Create file**

```vue
<!-- web/src/components/nav/NavItem.vue -->
<template>
  <router-link :to="to" class="nav-item" :class="{ active }">
    <component :is="icon" :size="20" class="nav-icon" />
    <span class="nav-text" v-if="!collapsed">{{ label }}</span>
    <span class="nav-badge" v-if="badge && !collapsed">{{ badge }}</span>
  </router-link>
</template>

<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  to: string
  icon: Component
  label: string
  active: boolean
  collapsed: boolean
  badge?: string
}>()
</script>

<style scoped>
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
</style>
```

**Step 2: Verify**

Run: `cd /data/project/LiteDock/web && npx vue-tsc --noEmit`
Expected: No type errors.

**Step 3: Commit**

```bash
git add web/src/components/nav/NavItem.vue
git commit -m "feat(frontend): add NavItem reusable navigation component"
```

---

### Task 5: Create `SidebarLogo` component

**Wave:** 2 (depends on Wave 1 completing)

**Files:**
- Create: `web/src/components/layout/SidebarLogo.vue`

```vue
<!-- web/src/components/layout/SidebarLogo.vue -->
<template>
  <div class="sidebar-header">
    <div class="logo" v-if="!collapsed">
      <span class="logo-text">LiteDock</span>
    </div>
    <div class="logo-collapsed" v-else>
      <span class="logo-text">LD</span>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  collapsed: boolean
}>()
</script>

<style scoped>
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

.logo-text {
  font-family: var(--font-mono);
  font-size: var(--font-size-base);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-strong);
  letter-spacing: -0.02em;
}

.logo-collapsed {
  display: flex;
  align-items: center;
  justify-content: center;
}

.logo-collapsed .logo-text {
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-strong);
}
</style>
```

---

### Task 6: Create `SidebarNav` component

**Wave:** 2 (depends on Task 4: NavItem)

**Files:**
- Create: `web/src/components/nav/SidebarNav.vue`

```vue
<!-- web/src/components/nav/SidebarNav.vue -->
<template>
  <nav class="sidebar-nav">
    <div class="nav-section">
      <div class="nav-section-title" v-if="!collapsed">{{ t('nav.main') }}</div>
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
      <div class="nav-section-title" v-if="!collapsed">{{ t('nav.system') }}</div>
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
</style>
```

---

### Task 7: Create `SidebarUserInfo` component

**Wave:** 2 (depends on Task 3: UserAvatar)

**Files:**
- Create: `web/src/components/layout/SidebarUserInfo.vue`

```vue
<!-- web/src/components/layout/SidebarUserInfo.vue -->
<template>
  <div class="sidebar-footer">
    <div v-if="!collapsed" class="user-info">
      <UserAvatar :username="authStore.user?.username" size="md" />
      <div class="user-details">
        <div class="user-name">{{ authStore.user?.username }}</div>
        <div class="user-role">{{ authStore.user?.role }}</div>
      </div>
    </div>

    <button @click="$emit('toggle')" class="sidebar-toggle">
      <ChevronLeft v-if="!collapsed" :size="18" />
      <ChevronRight v-if="collapsed" :size="18" />
    </button>
  </div>
</template>

<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import UserAvatar from '@/components/ui/UserAvatar.vue'
import { ChevronLeft, ChevronRight } from 'lucide-vue-next'

defineProps<{
  collapsed: boolean
}>()

defineEmits<{
  (e: 'toggle'): void
}>()

const authStore = useAuthStore()
</script>

<style scoped>
.sidebar-footer {
  padding: var(--space-3) 0;
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
  padding: var(--space-2) var(--space-4);
  color: var(--color-text-weak);
  font-size: var(--font-size-sm);
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
  display: flex;
  align-items: center;
  padding: var(--space-2) var(--space-4);
  background: none;
  border: none;
  color: var(--color-text-weak);
  cursor: pointer;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
}

.sidebar-toggle:hover {
  background: var(--color-background-hover);
  color: var(--color-text-strong);
}
</style>
```

---

### Task 8: Create `AppBreadcrumb` component

**Wave:** 2 (no dependencies)

**Files:**
- Create: `web/src/components/layout/AppBreadcrumb.vue`

```vue
<!-- web/src/components/layout/AppBreadcrumb.vue -->
<template>
  <div class="breadcrumb">
    <span v-for="(crumb, index) in crumbs" :key="index">
      {{ crumb }}
      <span v-if="index < crumbs.length - 1" class="breadcrumb-separator">/</span>
    </span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  crumbs: string[]
}>()
</script>

<style scoped>
.breadcrumb {
  font-size: var(--font-size-xs);
  color: var(--color-text-weaker);
  margin-top: 2px;
}

.breadcrumb-separator {
  margin: 0 var(--space-1);
  color: var(--color-border);
}
</style>
```

---

### Task 9: Create `UserMenu` component

**Wave:** 2 (depends on Task 3: UserAvatar)

**Files:**
- Create: `web/src/components/layout/UserMenu.vue`

```vue
<!-- web/src/components/layout/UserMenu.vue -->
<template>
  <div class="user-menu" ref="menuRef">
    <button @click="toggleMenu" class="user-menu-button">
      <UserAvatar :username="authStore.user?.username" size="sm" />
      <span class="user-name">{{ authStore.user?.username }}</span>
      <ChevronDown :size="16" />
    </button>

    <div v-if="menuOpen" class="user-menu-dropdown">
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
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { t } from '@/i18n'
import UserAvatar from '@/components/ui/UserAvatar.vue'
import { ChevronDown, Settings, LogOut } from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()

const menuOpen = ref(false)
const menuRef = ref<HTMLElement | null>(null)

const toggleMenu = () => {
  menuOpen.value = !menuOpen.value
}

const goToSettings = () => {
  menuOpen.value = false
  router.push('/settings')
}

const handleLogout = () => {
  menuOpen.value = false
  authStore.logout()
  router.push('/login')
}

const handleClickOutside = (event: Event) => {
  if (menuRef.value && !menuRef.value.contains(event.target as Node)) {
    menuOpen.value = false
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onUnmounted(() => document.removeEventListener('click', handleClickOutside))
</script>

<style scoped>
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

.user-name {
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-strong);
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
</style>
```

---

### Task 10: Create `Sidebar` composite component

**Wave:** 3 (depends on Tasks 1, 5, 6, 7)

**Files:**
- Create: `web/src/components/layout/Sidebar.vue`

```vue
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
  transition: width var(--transition-base);
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
```

---

### Task 11: Create `AppHeader` component

**Wave:** 3 (depends on Tasks 2, 8, 9)

**Files:**
- Create: `web/src/components/layout/AppHeader.vue`

```vue
<!-- web/src/components/layout/AppHeader.vue -->
<template>
  <header class="top-header">
    <div class="header-left">
      <h1 class="page-title">{{ currentPageTitle }}</h1>
      <AppBreadcrumb :crumbs="breadcrumbs" />
    </div>

    <div class="header-right">
      <div class="header-actions">
        <button @click="refreshData" class="btn btn-ghost" :title="t('common.refresh')">
          <RefreshCw :size="18" :class="{ spinning: refreshing }" />
        </button>
        <button @click="theme.toggle()" class="btn btn-ghost" :title="t('common.switchTheme')">
          <Sun v-if="theme.isDarkMode.value" :size="18" />
          <Moon v-else :size="18" />
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
import { useTheme } from '@/composables/useTheme'
import LanguageSwitcher from '@/components/LanguageSwitcher.vue'
import AppBreadcrumb from '@/components/layout/AppBreadcrumb.vue'
import UserMenu from '@/components/layout/UserMenu.vue'
import { RefreshCw, Bell, Sun, Moon } from 'lucide-vue-next'

const route = useRoute()
const theme = useTheme()

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

@media (max-width: 767px) {
  .top-header {
    padding: 0 var(--space-4);
  }
}

@media (min-width: 768px) and (max-width: 1023px) {
  .top-header {
    padding: 0 var(--space-5);
  }
}
</style>
```

---

### Task 12: Rewrite `MainLayout.vue`

**Wave:** 4 (depends on Tasks 10, 11)

**Files:**
- Modify: `web/src/layouts/MainLayout.vue` (649 lines → ~50 lines)

```vue
<!-- web/src/layouts/MainLayout.vue -->
<template>
  <div class="main-layout">
    <Sidebar />
    <div class="main-content">
      <AppHeader />
      <main class="page-content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import Sidebar from '@/components/layout/Sidebar.vue'
import AppHeader from '@/components/layout/AppHeader.vue'
</script>

<style scoped>
.main-layout {
  display: flex;
  height: 100vh;
  background: var(--color-background);
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

@media (max-width: 767px) {
  .page-content {
    padding: var(--space-4);
  }
}
</style>
```

---

### Task 13: Visual Verification & Build

**Wave:** 4 (depends on Task 12)

**Step 1: Run build**

```bash
cd /data/project/LiteDock/web && npm run build
```

**Step 2: Run type check**

```bash
cd /data/project/LiteDock/web && npx vue-tsc --noEmit
```

---

## Atomic Commit Strategy

| # | Commit | Files Changed | Wave |
|---|--------|---------------|------|
| 1 | `feat(frontend): add useSidebar composable` | `composables/useSidebar.ts` | 1 |
| 2 | `feat(frontend): add useTheme composable` | `composables/useTheme.ts` | 1 |
| 3 | `feat(frontend): add UserAvatar UI component` | `components/ui/UserAvatar.vue` | 1 |
| 4 | `feat(frontend): add NavItem navigation component` | `components/nav/NavItem.vue` | 1 |
| 5 | `feat(frontend): add SidebarLogo component` | `components/layout/SidebarLogo.vue` | 2 |
| 6 | `feat(frontend): add SidebarNav component` | `components/nav/SidebarNav.vue` | 2 |
| 7 | `feat(frontend): add SidebarUserInfo component` | `components/layout/SidebarUserInfo.vue` | 2 |
| 8 | `feat(frontend): add AppBreadcrumb component` | `components/layout/AppBreadcrumb.vue` | 2 |
| 9 | `feat(frontend): add UserMenu component` | `components/layout/UserMenu.vue` | 2 |
| 10 | `feat(frontend): add Sidebar composite component` | `components/layout/Sidebar.vue` | 3 |
| 11 | `feat(frontend): add AppHeader composite component` | `components/layout/AppHeader.vue` | 3 |
| 12 | `refactor(frontend): rewrite MainLayout to compose extracted components` | `layouts/MainLayout.vue` | 4 |
