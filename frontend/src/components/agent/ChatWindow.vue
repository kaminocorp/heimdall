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
  <div class="flex flex-col flex-1 min-h-0 border border-border rounded-lg bg-bg-surface overflow-hidden">
    <div ref="scrollContainer" class="flex-1 overflow-y-auto space-y-4 p-4">
      <ChatMessage v-for="msg in messages" :key="msg.id" :message="msg" />

      <!-- Thinking indicator -->
      <div v-if="isThinking" class="flex justify-start">
        <div class="max-w-2xl w-full">
          <div class="font-mono text-[10px] font-medium uppercase tracking-widest mb-1 text-text-muted">
            Heimdall
          </div>
          <div class="border border-border rounded-lg px-4 py-3 bg-bg-surface overflow-hidden relative">
            <span class="font-mono text-xs text-text-muted uppercase tracking-wider">Processing</span>
            <div class="scanning-line" />
          </div>
        </div>
      </div>
    </div>
    <ChatInput :disabled="disabled || isThinking" @send="emit('send', $event)" />
  </div>
</template>

<style scoped>
.scanning-line {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, transparent 0%, var(--accent) 50%, transparent 100%);
  animation: scan 1.5s ease-in-out infinite;
}
@keyframes scan {
  0%, 100% { transform: translateX(-100%); }
  50% { transform: translateX(100%); }
}
</style>
