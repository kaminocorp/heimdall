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
  <form @submit.prevent="handleSend" class="border-t p-4 flex gap-2">
    <input
      v-model="input"
      type="text"
      :placeholder="disabled ? 'Agent is thinking...' : 'Ask the agent...'"
      :disabled="disabled"
      class="flex-1 border rounded px-3 py-2 disabled:opacity-50 disabled:cursor-not-allowed"
    />
    <button
      type="submit"
      :disabled="disabled || !input.trim()"
      class="px-4 py-2 bg-gray-900 text-white rounded disabled:opacity-50 disabled:cursor-not-allowed"
    >
      Send
    </button>
  </form>
</template>
