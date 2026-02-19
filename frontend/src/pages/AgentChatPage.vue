<script setup lang="ts">
import { useAgent } from '@/composables/useAgent'
import ChatWindow from '@/components/agent/ChatWindow.vue'

const { messages, sendMessage } = useAgent()
</script>

<template>
  <div class="h-full">
    <h2 class="text-2xl font-semibold mb-4">Agent Chat</h2>
    <ChatWindow :messages="messages" @send="sendMessage" />
  </div>
</template>
