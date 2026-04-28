<template>
  <div class="nav-item-wrapper">
    <!-- Parent nav item -->
    <router-link
      v-if="to && !children"
      :to="to"
      class="nav-item"
      :class="{ active, collapsed, 'mobile-mode': isMobile, 'has-children': !!children }"
    >
      <component :is="icon" :size="20" class="nav-icon" />
      <span class="nav-text">{{ label }}</span>
      <span class="nav-badge" v-if="badge && !isMobile">{{ badge }}</span>
    </router-link>

    <!-- Parent with children (expandable) -->
    <div
      v-else-if="children"
      class="nav-item nav-parent"
      :class="{ active, collapsed, 'mobile-mode': isMobile, expanded, 'no-toggle': collapsed && !isMobile }"
      @click="handleClick"
    >
      <component :is="icon" :size="20" class="nav-icon" />
      <span class="nav-text">{{ label }}</span>
      <ChevronDown :size="16" class="nav-chevron" :class="{ rotated: expanded }" />
    </div>

    <!-- Child items - always show when collapsed on PC (icons only) -->
    <div v-if="children && (expanded || (collapsed && !isMobile))" class="nav-children" :class="{ 'collapsed-children': collapsed && !isMobile }">
      <router-link
        v-for="child in children"
        :key="child.name"
        :to="child.path"
        class="nav-item nav-child"
        :class="{ active: childActive(child.name), 'mobile-mode': isMobile, 'collapsed-mode': collapsed && !isMobile }"
      >
        <component v-if="child.icon" :is="child.icon" :size="16" class="nav-icon nav-child-icon" />
        <span class="nav-text" v-if="!collapsed || isMobile">{{ child.label }}</span>
      </router-link>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Component } from 'vue'
import { useRoute } from 'vue-router'
import { ChevronDown } from 'lucide-vue-next'

interface NavChildItem {
  name: string
  path: string
  label: string
  icon?: Component
}

const props = defineProps<{
  to?: string
  icon?: Component
  label: string
  active: boolean
  collapsed: boolean
  isMobile: boolean
  badge?: string
  children?: NavChildItem[]
  expanded?: boolean
}>()

const emit = defineEmits<{
  (e: 'toggle-children'): void
}>()

const route = useRoute()
const childActive = (name: string) => route.name === name

const handleClick = () => {
  // Don't toggle when collapsed on PC (can't see children anyway)
  if (props.collapsed && !props.isMobile) {
    return
  }
  emit('toggle-children')
}
</script>

<style scoped>
.nav-item-wrapper {
  position: relative;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-2) var(--space-4);
  color: var(--color-text-weak);
  text-decoration: none;
  font-size: var(--font-size-sm);
  transition: color 0.15s ease, background-color 0.15s ease;
  border-left: 2px solid transparent;
  margin: 1px 0;
  box-sizing: border-box;
  cursor: pointer;
  user-select: none;
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
  opacity: 1;
  transition: opacity 0.2s ease;
  flex: 1;
  min-width: 0;
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
  opacity: 1;
  transition: opacity 0.2s ease;
  flex-shrink: 0;
}

.nav-chevron {
  flex-shrink: 0;
  transition: transform 0.2s ease;
  opacity: 0.5;
}

.nav-chevron.rotated {
  transform: rotate(180deg);
}

/* Collapsed state - text fades, icon centers */
.nav-item.collapsed {
  justify-content: center;
  padding: var(--space-2);
}

.nav-item.collapsed .nav-text {
  display: none;
}

.nav-item.collapsed .nav-badge {
  display: none;
}

.nav-item.collapsed .nav-chevron {
  display: none;
}

/* Parent with children - no toggle when collapsed on PC */
.nav-item.no-toggle {
  cursor: default;
}

.nav-item.no-toggle:hover {
  background: transparent;
}

/* Child items */
.nav-children {
  overflow: hidden;
}

.nav-child {
  padding-left: calc(var(--space-4) + 12px);
  margin: 1px 0;
  font-size: var(--font-size-sm);
  border-left: 2px solid transparent;
}

.nav-child.active {
  border-left-color: var(--color-text-strong);
  background: var(--color-background-hover);
}

.nav-child-icon {
  flex-shrink: 0;
  margin-right: var(--space-1);
}

/* Collapsed PC mode - children show as icon stack */
.nav-children.collapsed-children {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0;
}

.nav-child.collapsed-mode {
  padding: var(--space-2);
  margin: 1px 0;
  justify-content: center;
  width: 100%;
  box-sizing: border-box;
  border-left: 2px solid transparent;
  border-radius: 0;
}

.nav-child.collapsed-mode.active {
  border-left-color: var(--color-text-strong);
  background: var(--color-background-hover);
}

.nav-child.collapsed-mode .nav-icon {
  margin-right: 0;
}

.nav-child.collapsed-mode .nav-text {
  display: none;
}

/* Mobile mode */
@media (max-width: 767px) {
  .nav-item.mobile-mode {
    justify-content: flex-start;
    padding: var(--space-3) var(--space-6);
    margin: var(--space-1) var(--space-4);
    border-radius: var(--radius-md);
    border-left: none;
  }

  .nav-item.mobile-mode:hover {
    background: var(--color-background-hover);
  }

  .nav-item.mobile-mode.active {
    background: var(--color-background-hover);
    border-left: none;
  }

  .nav-item.mobile-mode .nav-text {
    flex: none;
    font-size: var(--font-size-base);
  }

  .nav-item.mobile-mode .nav-badge {
    display: none;
  }

  .nav-child.mobile-mode {
    padding-left: calc(var(--space-4) + 12px);
    margin: var(--space-1) var(--space-4);
    border-left: 2px solid transparent;
  }

  .nav-child.mobile-mode.active {
    border-left-color: var(--color-text-strong);
  }

  .nav-child.mobile-mode .nav-text {
    font-size: var(--font-size-base);
  }
}
</style>