import { ref, onUnmounted } from 'vue'

export interface WebSocketOptions {
  token?: string
  conversationId?: string
  appId?: string
}

export function useWebSocket(url: string, options?: WebSocketOptions) {
  const data = ref<string | null>(null)
  const status = ref<'connecting' | 'open' | 'closed'>('connecting')
  let ws: WebSocket | null = null

  function buildUrl(): string {
    const u = new URL(url)
    if (options?.token) {
      u.searchParams.set('token', options.token)
    }
    if (options?.conversationId) {
      u.searchParams.set('conversation_id', options.conversationId)
    }
    if (options?.appId) {
      u.searchParams.set('app_id', options.appId)
    }
    return u.toString()
  }

  function connect() {
    ws = new WebSocket(buildUrl())

    ws.onopen = () => {
      status.value = 'open'
    }

    ws.onmessage = (event) => {
      data.value = event.data
    }

    ws.onclose = () => {
      status.value = 'closed'
    }

    ws.onerror = () => {
      status.value = 'closed'
    }
  }

  function send(message: string) {
    ws?.send(message)
  }

  function close() {
    ws?.close()
  }

  connect()

  onUnmounted(() => {
    close()
  })

  return { data, status, send, close }
}
