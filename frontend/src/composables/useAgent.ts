import { ref, watch } from 'vue'
import type { ChatMessage } from '@/types/agent'
import { useWebSocket } from './useWebSocket'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

export function useAgent(existingConversationId?: string) {
  const auth = useAuthStore()
  const appStore = useAppStore()
  const messages = ref<ChatMessage[]>([])
  const conversationId = ref<string | null>(existingConversationId ?? null)
  const isThinking = ref(false)
  const error = ref<string | null>(null)

  const wsUrl = `${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/ws/chat`

  const { data, status, send } = useWebSocket(wsUrl, {
    token: auth.token ?? undefined,
    conversationId: existingConversationId,
    appId: appStore.currentAppId ?? undefined,
  })

  watch(data, (raw) => {
    if (raw === null) return
    try {
      const parsed = JSON.parse(raw)

      // System message — contains conversation_id on connect.
      if (parsed.type === 'system') {
        if (parsed.conversation_id) {
          conversationId.value = parsed.conversation_id
        }
        return
      }

      // Status message — thinking indicator.
      if (parsed.type === 'status') {
        if (parsed.content === 'thinking') {
          isThinking.value = true
        }
        return
      }

      // Error message from agent.
      if (parsed.type === 'error') {
        isThinking.value = false
        error.value = parsed.content ?? 'Unknown error'
        return
      }

      // Chat message from agent.
      isThinking.value = false
      error.value = null
      const msg = parsed as ChatMessage
      messages.value.push({
        id: msg.id ?? crypto.randomUUID(),
        role: msg.role ?? 'agent',
        content: msg.content ?? '',
        tool_calls: msg.tool_calls,
        timestamp: msg.timestamp ?? new Date().toISOString(),
      })
    } catch {
      // Ignore malformed messages.
    }
  })

  function sendMessage(content: string) {
    error.value = null
    messages.value.push({
      id: crypto.randomUUID(),
      role: 'user',
      content,
      timestamp: new Date().toISOString(),
    })
    send(JSON.stringify({ content }))
  }

  // Load existing messages into the reactive array (used when resuming a conversation).
  function loadMessages(existingMessages: ChatMessage[]) {
    messages.value = existingMessages.map((m) => ({
      ...m,
      // Backend stores "assistant" but frontend uses "agent" for display.
      role: m.role === 'assistant' ? 'agent' : m.role,
    }))
  }

  return { messages, status, conversationId, isThinking, error, sendMessage, loadMessages }
}
