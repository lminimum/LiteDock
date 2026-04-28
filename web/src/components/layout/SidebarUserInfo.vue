<!-- web/src/components/layout/SidebarUserInfo.vue -->
<template>
  <div class="sidebar-footer" :class="{ collapsed }">
    <div class="user-info-wrapper">
      <div class="user-info">
        <UserAvatar :username="authStore.user?.username" size="md" />
        <div class="user-details">
          <div class="user-name">{{ authStore.user?.username }}</div>
          <div class="user-role">{{ authStore.user?.role }}</div>
        </div>
      </div>
      <div class="user-info-spacer"></div>
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
  padding: var(--space-3) var(--space-4);
  border-top: 1px solid var(--color-border-weak);
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-sizing: border-box;
  height: 56px;
}

.user-info-wrapper {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.user-info {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  color: var(--color-text-weak);
  font-size: var(--font-size-sm);
  white-space: nowrap;
  opacity: 1;
  transition: opacity 0.2s ease;
}

.user-info-spacer {
  flex: 1;
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
  padding: var(--space-2);
  background: none;
  border: none;
  color: var(--color-text-weak);
  cursor: pointer;
  border-radius: var(--radius-sm);
  transition: all var(--transition-fast);
  flex-shrink: 0;
}

.sidebar-toggle:hover {
  background: var(--color-background-weak);
  color: var(--color-text-strong);
}

/* Collapsed state - user info fades out but space remains */
.sidebar-footer.collapsed .user-info {
  opacity: 0;
}
</style>
