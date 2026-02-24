<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import type { ChatMessage as ChatMessageType } from '@/types/agent'
import ChatMessage from './ChatMessage.vue'
import ChatInput from './ChatInput.vue'

const props = defineProps<{
  messages: ChatMessageType[]
  isThinking?: boolean
  disabled?: boolean
}>()

const emit = defineEmits<{
  send: [message: string]
}>()

const scrollContainer = ref<HTMLElement | null>(null)

function scrollToBottom() {
  nextTick(() => {
    if (scrollContainer.value) {
      scrollContainer.value.scrollTop = scrollContainer.value.scrollHeight
    }
  })
}

watch(() => props.messages.length, scrollToBottom)
watch(() => props.isThinking, scrollToBottom)
</script>

<template>
  <div class="flex flex-col flex-1 min-h-0">
    <div ref="scrollContainer" class="flex-1 overflow-y-auto space-y-4 p-4">
      <ChatMessage v-for="msg in messages" :key="msg.id" :message="msg" />

      <div v-if="isThinking" class="flex justify-start">
        <div class="bg-gray-100 rounded-lg px-4 py-2 text-gray-500">
          <span class="inline-flex gap-1">
            <span class="animate-bounce" style="animation-delay: 0ms">.</span>
            <span class="animate-bounce" style="animation-delay: 150ms">.</span>
            <span class="animate-bounce" style="animation-delay: 300ms">.</span>
          </span>
        </div>
      </div>
    </div>
    <ChatInput :disabled="disabled || isThinking" @send="emit('send', $event)" />
  </div>
</template>
