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

const setLocale = (l: 'en' | 'zh-CN') => {
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
  gap: var(--space-2);
  padding: var(--space-1) var(--space-3);
  background: transparent;
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-sm);
  color: var(--color-text-weak);
  font-family: var(--font-mono);
  font-size: var(--font-size-xs);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.lang-btn:hover {
  color: var(--color-text-strong);
  border-color: var(--color-text-weaker);
  background: var(--color-background-weak);
}

.lang-abbrev {
  font-weight: var(--font-weight-medium);
  letter-spacing: 0.5px;
}

.lang-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  right: 0;
  min-width: 120px;
  background: var(--color-background-weak);
  border: 1px solid var(--color-border-weak);
  border-radius: var(--radius-sm);
  padding: var(--space-1);
  z-index: 1000;
}

.lang-dropdown button {
  display: block;
  width: 100%;
  padding: var(--space-2) var(--space-3);
  background: transparent;
  border: none;
  border-radius: var(--radius-sm);
  color: var(--color-text-weak);
  font-family: var(--font-mono);
  font-size: var(--font-size-sm);
  text-align: left;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.lang-dropdown button:hover {
  background: var(--color-background-weak);
  color: var(--color-text-strong);
}

.lang-dropdown button.active {
  background: var(--color-background-interactive-weaker);
  color: var(--color-text-strong);
}
</style>
