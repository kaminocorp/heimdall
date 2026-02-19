import { ref, watch } from 'vue'
import type { ChatMessage } from '@/types/agent'
import { useWebSocket } from './useWebSocket'

export function useAgent() {
  const messages = ref<ChatMessage[]>([])
  const { data, status, send } = useWebSocket(`${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/ws/chat`)

  watch(data, (raw) => {
    if (raw === null) return
    try {
      const msg = JSON.parse(raw) as ChatMessage
      messages.value.push({
        id: msg.id ?? crypto.randomUUID(),
        role: msg.role ?? 'agent',
        content: msg.content ?? '',
        tool_calls: msg.tool_calls,
        timestamp: msg.timestamp ?? new Date().toISOString(),
      })
    } catch {
      // Ignore malformed messages
    }
  })

  function sendMessage(content: string) {
    messages.value.push({ id: crypto.randomUUID(), role: 'user', content, timestamp: new Date().toISOString() })
    send(JSON.stringify({ content }))
  }

  return { messages, status, sendMessage }
}
