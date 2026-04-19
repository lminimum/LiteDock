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
