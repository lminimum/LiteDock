import { ref, watch } from 'vue'

export type ThemeId = 'light' | 'dark'

export interface ThemeInfo {
  id: ThemeId
  name: string
}

const THEME_STORAGE_KEY = 'litedock-theme'

const availableThemes: ThemeInfo[] = [
  { id: 'light', name: 'Light' },
  { id: 'dark', name: 'Dark' },
]

const currentTheme = ref<ThemeId>(loadTheme())

function loadTheme(): ThemeId {
  try {
    const stored = localStorage.getItem(THEME_STORAGE_KEY)
    if (stored === 'dark' || stored === 'light') {
      return stored
    }
  } catch {
    // localStorage unavailable (SSR / privacy mode)
  }
  return 'light'
}

function applyTheme(theme: ThemeId): void {
  document.documentElement.dataset.theme = theme
  try {
    localStorage.setItem(THEME_STORAGE_KEY, theme)
  } catch {
    // localStorage unavailable
  }
}

applyTheme(currentTheme.value)

watch(currentTheme, (theme) => {
  applyTheme(theme)
})

export function useTheme() {
  function toggleTheme(): void {
    currentTheme.value = currentTheme.value === 'light' ? 'dark' : 'light'
  }

  function setTheme(theme: ThemeId): void {
    currentTheme.value = theme
  }

  return {
    currentTheme,
    toggleTheme,
    setTheme,
    availableThemes,
  }
}
