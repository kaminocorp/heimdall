<script setup lang="ts">
import type { ChatMessage } from '@/types/agent'

defineProps<{
  message: ChatMessage
}>()
</script>

<template>
  <div class="flex" :class="message.role === 'user' ? 'justify-end' : 'justify-start'">
    <div class="max-w-md rounded-lg px-4 py-2" :class="message.role === 'user' ? 'bg-gray-900 text-white' : 'bg-gray-100'">
      {{ message.content }}
    </div>
  </div>
</template>
