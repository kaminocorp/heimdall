<script setup lang="ts">
import type { ChatMessage as ChatMessageType } from '@/types/agent'
import ChatMessage from './ChatMessage.vue'
import ChatInput from './ChatInput.vue'

defineProps<{
  messages: ChatMessageType[]
}>()

const emit = defineEmits<{
  send: [message: string]
}>()
</script>

<template>
  <div class="flex flex-col h-full">
    <div class="flex-1 overflow-y-auto space-y-4 p-4">
      <ChatMessage v-for="(msg, i) in messages" :key="i" :message="msg" />
    </div>
    <ChatInput @send="emit('send', $event)" />
  </div>
</template>
