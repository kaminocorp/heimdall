<script setup lang="ts">
import { ref } from 'vue'

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
    <input v-model="input" type="text" placeholder="Ask the agent..." class="flex-1 border rounded px-3 py-2" />
    <button type="submit" class="px-4 py-2 bg-gray-900 text-white rounded">Send</button>
  </form>
</template>
