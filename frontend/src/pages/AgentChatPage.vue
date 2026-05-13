<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAgent } from '@/composables/useAgent'
import { getConversation } from '@/api/conversations'
import ChatWindow from '@/components/agent/ChatWindow.vue'

const route = useRoute()
const existingId = route.query.conversation_id as string | undefined

const { messages, status, isThinking, activeTools, error, sendMessage, loadMessages } =
  useAgent(existingId)

function formatTool(name: string): string {
  return name.replace(/_/g, ' ')
}

onMounted(async () => {
  if (existingId) {
    try {
      const conversation = await getConversation(existingId)
      if (conversation.messages?.length) {
        loadMessages(conversation.messages)
      }
    } catch {
      // Conversation not found — start fresh.
    }
  }
})
</script>

<template>
  <div class="h-full flex flex-col">
    <!-- Page header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 mb-6 border-b border-border">
      <div>
        <h2 class="font-mono text-2xl font-bold uppercase tracking-wider text-text-primary">Agent Chat</h2>
        <p class="font-sans text-sm text-text-secondary mt-1">Direct terminal link to the monitoring agent</p>
      </div>
      <div class="flex items-center gap-2 font-mono text-xs uppercase tracking-wider">
        <span
          class="w-2 h-2 rounded-full"
          :class="{
            'bg-status-warn glow-pulse-warn': status === 'connecting',
            'bg-status-ok glow-pulse': status === 'open',
            'bg-status-critical glow-pulse-critical': status === 'closed',
          }"
        />
        <span class="text-text-muted">{{ status === 'open' ? 'Connected' : status === 'connecting' ? 'Connecting' : 'Disconnected' }}</span>
      </div>
    </div>

    <div v-if="error" class="mb-3 rounded border border-status-critical/30 bg-status-critical/10 px-4 py-2 text-sm font-mono text-status-critical">
      {{ error }}
    </div>

    <!-- Live tool-progress indicator. Populated from streaming tool_start /
         tool_result events on the WebSocket; empty when no tool is running. -->
    <div
      v-if="activeTools.length"
      class="mb-3 flex items-center gap-2 rounded border border-accent-border/30 bg-accent-subtle/40 px-4 py-2 font-mono text-xs uppercase tracking-wider text-text-secondary"
    >
      <span class="w-1.5 h-1.5 rounded-full bg-accent glow-pulse" />
      <span>{{ activeTools.map(formatTool).join(' · ') }}</span>
    </div>

    <ChatWindow
      :messages="messages"
      :is-thinking="isThinking"
      :disabled="status !== 'open'"
      @send="sendMessage"
    />
  </div>
</template>
