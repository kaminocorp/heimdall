import { ref, onUnmounted } from 'vue'

export function useWebSocket(url: string) {
  const data = ref<string | null>(null)
  const status = ref<'connecting' | 'open' | 'closed'>('connecting')
  let ws: WebSocket | null = null

  function connect() {
    ws = new WebSocket(url)

    ws.onopen = () => {
      status.value = 'open'
    }

    ws.onmessage = (event) => {
      data.value = event.data
    }

    ws.onclose = () => {
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
