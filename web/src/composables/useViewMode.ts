import { ref, watch } from 'vue'

const STORAGE_PREFIX = 'litedock-view-mode-'

export function useViewMode(pageKey: string, defaultMode: 'card' | 'list' = 'list') {
  const storageKey = STORAGE_PREFIX + pageKey

  const getInitial = (): 'card' | 'list' => {
    const saved = localStorage.getItem(storageKey)
    if (saved === 'card' || saved === 'list') return saved
    return defaultMode
  }

  const viewMode = ref<'card' | 'list'>(getInitial())

  watch(viewMode, (val) => {
    localStorage.setItem(storageKey, val)
  })

  return viewMode
}
