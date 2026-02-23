<script setup lang="ts">
import { ref } from 'vue'
import type { CreateConnectionPayload } from '@/types/connection'

const name = ref('')
const type = ref('postgres')
const direction = ref<'one_way' | 'two_way'>('one_way')

const emit = defineEmits<{
  submit: [data: CreateConnectionPayload]
  cancel: []
}>()

function handleSubmit() {
  emit('submit', {
    name: name.value,
    type: type.value,
    direction: direction.value,
    config: {},
  })
  name.value = ''
  type.value = 'postgres'
  direction.value = 'one_way'
}
</script>

<template>
  <form @submit.prevent="handleSubmit" class="space-y-4 border rounded-lg p-4 bg-gray-50">
    <div>
      <label class="block text-sm font-medium">Name</label>
      <input v-model="name" type="text" required placeholder="e.g. Production DB"
        class="mt-1 block w-full border rounded px-3 py-2" />
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
    <div>
      <label class="block text-sm font-medium">Direction</label>
      <select v-model="direction" class="mt-1 block w-full border rounded px-3 py-2">
        <option value="one_way">One-way (ingest only)</option>
        <option value="two_way">Two-way (ingest + query)</option>
      </select>
    </div>
    <div class="flex gap-2">
      <button type="submit" class="px-4 py-2 bg-gray-900 text-white rounded">Add Connection</button>
      <button type="button" @click="emit('cancel')" class="px-4 py-2 border rounded">Cancel</button>
    </div>
  </form>
</template>
