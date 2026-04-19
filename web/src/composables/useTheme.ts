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
