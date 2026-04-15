import { ref, onUnmounted } from 'vue'

export interface WebSocketOptions {
  token?: string
  conversationId?: string
  appId?: string
}

const RECONNECT_BASE_MS = 1000
const RECONNECT_MAX_MS = 30000

export function useWebSocket(url: string, options?: WebSocketOptions) {
  // Wrapped in an object so Vue's watch fires even when consecutive
  // messages have identical string content (shallow equality bypass).
  const data = ref<{ payload: string; ts: number } | null>(null)
  const status = ref<'connecting' | 'open' | 'closed'>('connecting')
  const error = ref<Event | null>(null)
  let ws: WebSocket | null = null
  let reconnectAttempts = 0
  let reconnectTimer: ReturnType<typeof setTimeout> | null = null
  let intentionalClose = false

  // Allow options to be updated externally for reactive token/appId.
  let currentOptions = { ...options }

  function updateOptions(newOptions: Partial<WebSocketOptions>) {
    const changed =
      newOptions.token !== currentOptions.token ||
      newOptions.appId !== currentOptions.appId
    currentOptions = { ...currentOptions, ...newOptions }
    // Reconnect if the socket is open and critical params changed.
    if (changed && ws?.readyState === WebSocket.OPEN) {
      intentionalClose = true
      ws.close()
      connect()
    }
  }

  function buildUrl(): string {
    const u = new URL(url)
    if (currentOptions.token) {
      u.searchParams.set('token', currentOptions.token)
    }
    if (currentOptions.conversationId) {
      u.searchParams.set('conversation_id', currentOptions.conversationId)
    }
    if (currentOptions.appId) {
      u.searchParams.set('app_id', currentOptions.appId)
    }
    return u.toString()
  }

  function connect() {
    intentionalClose = false
    status.value = 'connecting'
    ws = new WebSocket(buildUrl())

    ws.onopen = () => {
      status.value = 'open'
      reconnectAttempts = 0
    }

    ws.onmessage = (event) => {
      data.value = { payload: event.data, ts: Date.now() }
    }

    ws.onclose = () => {
      // When the close was triggered by updateOptions (intentional reconnect),
      // skip setting status to 'closed' — the new socket is already connecting
      // and the brief 'closed' flash would reset useAgent's thinking state.
      if (!intentionalClose) {
        status.value = 'closed'
        scheduleReconnect()
      }
    }

    ws.onerror = (event) => {
      error.value = event
      console.warn('WebSocket error:', event)
      // onclose will fire after onerror, which triggers reconnect.
    }
  }

  function scheduleReconnect() {
    if (reconnectTimer) return
    const delay = Math.min(
      RECONNECT_BASE_MS * Math.pow(2, reconnectAttempts),
      RECONNECT_MAX_MS,
    )
    reconnectAttempts++
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      connect()
    }, delay)
  }

  function send(message: string) {
    if (ws?.readyState === WebSocket.OPEN) {
      ws.send(message)
    }
  }

  function close() {
    intentionalClose = true
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    ws?.close()
  }

  connect()

  onUnmounted(() => {
    close()
  })

  return { data, status, error, send, close, updateOptions }
}
