<script setup lang="ts">
import { ref } from 'vue'

defineProps<{
  disabled?: boolean
}>()

const input = ref('')
const emit = defineEmits<{
  send: [message: string]
}>()

function handleSend() {
  if (!input.value.trim()) return
  emit('send', input.value.trim())
  input.value = ''
}
</script>

<template>
  <form @submit.prevent="handleSend" class="border-t border-border p-4 flex gap-2">
    <input
      v-model="input"
      type="text"
      :placeholder="disabled ? 'Agent is processing...' : 'Type your message...'"
      :disabled="disabled"
      class="flex-1 bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
    />
    <button
      type="submit"
      :disabled="disabled || !input.trim()"
      class="px-4 py-2 bg-accent text-bg-primary font-mono text-sm font-medium uppercase tracking-wider rounded hover:bg-accent-hover disabled:opacity-30 disabled:cursor-not-allowed transition-colors cursor-pointer"
    >
      Send
    </button>
  </form>
</template>
