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
