import { ref, onUnmounted } from 'vue'

interface UseWebSocketOptions {
  url: string
  onMessage?: (data: Record<string, unknown>) => void
  onError?: (error: Event) => void
  reconnectDelay?: number
}

export function useWebSocket(options: UseWebSocketOptions) {
  const { url, onMessage, onError, reconnectDelay = 5000 } = options
  
  const ws = ref<WebSocket | null>(null)
  const isConnected = ref(false)
  const reconnectTimer = ref<ReturnType<typeof setTimeout> | null>(null)

  const connect = () => {
    if (ws.value?.readyState === WebSocket.OPEN) return

    ws.value = new WebSocket(url)
    
    ws.value.onopen = () => {
      isConnected.value = true
      console.log('WebSocket connected')
    }
    
    ws.value.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        onMessage?.(data)
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e)
      }
    }
    
    ws.value.onerror = (e) => {
      console.error('WebSocket error:', e)
      onError?.(e)
    }
    
    ws.value.onclose = () => {
      isConnected.value = false
      console.log('WebSocket disconnected')
      reconnectTimer.value = setTimeout(connect, reconnectDelay)
    }
  }
  
  const disconnect = () => {
    if (reconnectTimer.value) {
      clearTimeout(reconnectTimer.value)
      reconnectTimer.value = null
    }
    if (ws.value) {
      ws.value.onclose = null
      ws.value.close()
      ws.value = null
    }
  }

  const send = (data: unknown): void => {
    if (ws.value?.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify(data))
    }
  }
  
  onUnmounted(disconnect)
  
  return {
    ws,
    isConnected,
    connect,
    disconnect,
    send
  }
}
