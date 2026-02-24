<script setup lang="ts">
import { onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAgent } from '@/composables/useAgent'
import { getConversation } from '@/api/conversations'
import ChatWindow from '@/components/agent/ChatWindow.vue'

const route = useRoute()
const existingId = route.query.conversation_id as string | undefined

const { messages, status, conversationId, isThinking, error, sendMessage, loadMessages } =
  useAgent(existingId)

onMounted(async () => {
  if (existingId) {
    try {
      const { data } = await getConversation(existingId)
      if (data.messages?.length) {
        loadMessages(data.messages)
      }
    } catch {
      // Conversation not found — start fresh.
    }
  }
})
</script>

<template>
  <div class="h-full flex flex-col">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-2xl font-semibold">Agent Chat</h2>
      <div class="flex items-center gap-2 text-sm">
        <span
          class="inline-block w-2 h-2 rounded-full"
          :class="{
            'bg-yellow-400': status === 'connecting',
            'bg-green-500': status === 'open',
            'bg-red-500': status === 'closed',
          }"
        />
        <span class="text-gray-500">{{ status }}</span>
      </div>
    </div>

    <div v-if="error" class="mb-2 rounded bg-red-50 border border-red-200 px-4 py-2 text-sm text-red-700">
      {{ error }}
    </div>

    <ChatWindow
      :messages="messages"
      :is-thinking="isThinking"
      :disabled="status !== 'open'"
      @send="sendMessage"
    />
  </div>
</template>
