<template>
  <div class="lang-switcher" ref="switcherRef">
    <button @click="toggleDropdown" class="lang-btn" title="Switch Language">
      <Globe :size="18" />
      <span class="lang-abbrev">{{ locale === 'en' ? 'EN' : '中文' }}</span>
    </button>
    <div v-if="dropdownOpen" class="lang-dropdown">
      <button @click="setLocale('en')" :class="{ active: locale === 'en' }">
        English
      </button>
      <button @click="setLocale('zh-CN')" :class="{ active: locale === 'zh-CN' }">
        中文
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Globe } from 'lucide-vue-next'
import { locale } from '@/i18n'

const dropdownOpen = ref(false)
const switcherRef = ref<HTMLElement | null>(null)

const toggleDropdown = () => {
  dropdownOpen.value = !dropdownOpen.value
}

const setLocale = (l: string) => {
  locale.value = l
  dropdownOpen.value = false
}

const handleClickOutside = (e: MouseEvent) => {
  if (switcherRef.value && !switcherRef.value.contains(e.target as Node)) {
    dropdownOpen.value = false
  }
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onUnmounted(() => document.removeEventListener('click', handleClickOutside))
</script>

<style scoped>
.lang-switcher {
  position: relative;
}

.lang-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  background: transparent;
  border: 1px solid var(--color-border, rgba(255, 255, 255, 0.1));
  border-radius: 6px;
  color: var(--color-text-secondary, #a0a0a0);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.lang-btn:hover {
  color: var(--color-text-primary, #ffffff);
  border-color: var(--color-border-hover, rgba(255, 255, 255, 0.2));
  background: var(--color-background-hover, rgba(255, 255, 255, 0.05));
}

.lang-abbrev {
  font-weight: 500;
  letter-spacing: 0.5px;
}

.lang-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  min-width: 120px;
  background: var(--color-background-elevated, #1a1a2e);
  border: 1px solid var(--color-border, rgba(255, 255, 255, 0.1));
  border-radius: 8px;
  padding: 4px;
  z-index: 1000;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.lang-dropdown button {
  display: block;
  width: 100%;
  padding: 8px 12px;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: var(--color-text-secondary, #a0a0a0);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  text-align: left;
  cursor: pointer;
  transition: all 0.15s ease;
}

.lang-dropdown button:hover {
  background: var(--color-background-hover, rgba(255, 255, 255, 0.05));
  color: var(--color-text-primary, #ffffff);
}

.lang-dropdown button.active {
  background: var(--color-background-interactive, rgba(99, 102, 241, 0.15));
  color: var(--color-text-primary, #ffffff);
}
</style>
