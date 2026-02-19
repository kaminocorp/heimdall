import { ref } from 'vue'
import type { ChatMessage } from '@/types/agent'
import { useWebSocket } from './useWebSocket'

export function useAgent() {
  const messages = ref<ChatMessage[]>([])
  const { data, status, send } = useWebSocket(`${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/ws/chat`)

  function sendMessage(content: string) {
    messages.value.push({ role: 'user', content, timestamp: new Date().toISOString() })
    send(JSON.stringify({ content }))
  }

  return { messages, status, sendMessage }
}
