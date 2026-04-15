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
  // activeTools is the running list of tool names the agent is currently
  // executing. The chat page renders these as a small progress indicator so
  // the user can see what the agent is doing during the wait. Populated by
  // tool_start events from the backend; entries are removed on tool_result.
  const activeTools = ref<string[]>([])
  const error = ref<string | null>(null)

  const wsUrl = `${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/ws/chat`

  const { data, status, send, updateOptions } = useWebSocket(wsUrl, {
    token: auth.token ?? undefined,
    conversationId: existingConversationId,
    appId: appStore.currentAppId ?? undefined,
  })

  // Reactively update WebSocket options when token or appId changes.
  watch(
    () => auth.token,
    (newToken) => updateOptions({ token: newToken ?? undefined }),
  )
  watch(
    () => appStore.currentAppId,
    (newAppId) => updateOptions({ appId: newAppId ?? undefined }),
  )

  // Reset thinking/tool state when the WebSocket connection drops.
  watch(status, (newStatus) => {
    if (newStatus === 'closed') {
      isThinking.value = false
      activeTools.value = []
    }
  })

  watch(data, (msg) => {
    if (msg === null) return
    const raw = msg.payload
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
        activeTools.value = []
        error.value = parsed.content ?? 'Unknown error'
        return
      }

      // Tool progress events from the streaming agent loop. The backend
      // emits tool_start immediately before dispatch and tool_result
      // immediately after. We swap "thinking" for the active-tools indicator
      // on the first tool_start so the UI doesn't show both at once.
      if (parsed.type === 'tool_start') {
        isThinking.value = false
        if (parsed.tool && !activeTools.value.includes(parsed.tool)) {
          activeTools.value = [...activeTools.value, parsed.tool]
        }
        return
      }
      if (parsed.type === 'tool_result') {
        if (parsed.tool) {
          activeTools.value = activeTools.value.filter((t) => t !== parsed.tool)
        }
        // When the last tool finishes the agent is back to "thinking" while
        // it composes its synthesis. Re-arm the indicator only if no other
        // tools are still running.
        if (activeTools.value.length === 0) {
          isThinking.value = true
        }
        return
      }

      // Chat message from agent.
      isThinking.value = false
      activeTools.value = []
      error.value = null
      const msg = parsed as ChatMessage
      messages.value.push({
        id: msg.id ?? crypto.randomUUID(),
        role: msg.role ?? 'agent',
        content: msg.content ?? '',
        tool_calls: msg.tool_calls,
        timestamp: msg.timestamp ?? new Date().toISOString(),
      })
    } catch (e) {
      console.warn('WebSocket: failed to parse message:', e)
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

  return { messages, status, conversationId, isThinking, activeTools, error, sendMessage, loadMessages }
}
