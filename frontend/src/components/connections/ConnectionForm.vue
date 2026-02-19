<script setup lang="ts">
import { ref } from 'vue'

const name = ref('')
const type = ref('postgres')

const emit = defineEmits<{
  submit: [data: { name: string; type: string }]
}>()

function handleSubmit() {
  emit('submit', { name: name.value, type: type.value })
}
</script>

<template>
  <form @submit.prevent="handleSubmit" class="space-y-4">
    <div>
      <label class="block text-sm font-medium">Name</label>
      <input v-model="name" type="text" class="mt-1 block w-full border rounded px-3 py-2" />
    </div>
    <div>
      <label class="block text-sm font-medium">Type</label>
      <select v-model="type" class="mt-1 block w-full border rounded px-3 py-2">
        <option value="postgres">PostgreSQL</option>
        <option value="webhook_logs">Webhook Logs</option>
        <option value="syslog">Syslog</option>
        <option value="github">GitHub</option>
      </select>
    </div>
    <button type="submit" class="px-4 py-2 bg-gray-900 text-white rounded">Add Connection</button>
  </form>
</template>
