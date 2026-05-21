import { onMounted, onUnmounted } from 'vue'

export const REFRESH_EVENT = 'litdock-refresh'

export const triggerRefresh = () => {
  window.dispatchEvent(new CustomEvent(REFRESH_EVENT))
}

export const useRefreshBus = (handler: () => void | Promise<void>) => {
  const onRefresh = () => {
    void handler()
  }

  onMounted(() => {
    window.addEventListener(REFRESH_EVENT, onRefresh)
  })

  onUnmounted(() => {
    window.removeEventListener(REFRESH_EVENT, onRefresh)
  })
}
