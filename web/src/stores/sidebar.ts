import { defineStore } from 'pinia'
import { ref } from 'vue'

const STORAGE_KEY = 'litedock-sidebar-collapsed'

export const useSidebarStore = defineStore('sidebar', () => {
  const collapsed = ref(false)

  // Restore from localStorage on initialization
  const saved = localStorage.getItem(STORAGE_KEY)
  if (saved !== null) {
    collapsed.value = saved === 'true'
  }

  const toggle = () => {
    collapsed.value = !collapsed.value
    localStorage.setItem(STORAGE_KEY, collapsed.value.toString())
  }

  const collapse = () => {
    collapsed.value = true
    localStorage.setItem(STORAGE_KEY, 'true')
  }

  const expand = () => {
    collapsed.value = false
    localStorage.setItem(STORAGE_KEY, 'false')
  }

  return {
    collapsed,
    toggle,
    collapse,
    expand,
  }
})
