import { createI18n } from 'vue-i18n'
import { computed } from 'vue'
import en from './locales/en.json'
import zhCN from './locales/zh-CN.json'

const LOCALE_STORAGE_KEY = 'litedock-locale'

function detectLocale(): string {
  const stored = localStorage.getItem(LOCALE_STORAGE_KEY)
  if (stored) return stored
  return navigator.language.startsWith('zh') ? 'zh-CN' : 'en'
}

export const i18n = createI18n({
  legacy: false,
  locale: detectLocale(),
  fallbackLocale: 'en',
  messages: {
    en,
    'zh-CN': zhCN,
  },
})

export const locale = computed({
  get: () => i18n.global.locale.value,
  set: (val: string) => {
    i18n.global.locale.value = val
    localStorage.setItem(LOCALE_STORAGE_KEY, val)
  },
})

export const t = i18n.global.t.bind(i18n.global)