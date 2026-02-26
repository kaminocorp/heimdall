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
  <form @submit.prevent="handleSubmit" class="space-y-4 border border-border rounded-lg p-5 bg-bg-surface">
    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Name</label>
      <input v-model="name" type="text" required placeholder="e.g. Production DB"
        class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm placeholder:text-text-muted focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors" />
    </div>
    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Type</label>
      <select v-model="type" class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors">
        <option value="postgres">PostgreSQL</option>
        <option value="webhook_logs">Webhook Logs</option>
        <option value="syslog">Syslog</option>
        <option value="github">GitHub</option>
      </select>
    </div>
    <div>
      <label class="block font-mono text-xs font-medium uppercase tracking-wider text-text-secondary mb-1.5">Direction</label>
      <select v-model="direction" class="block w-full bg-bg-elevated/80 border border-border rounded px-3 py-2 text-text-primary font-mono text-sm focus:border-accent/50 focus:ring-1 focus:ring-accent/20 focus:outline-none transition-colors">
        <option value="one_way">One-way (ingest only)</option>
        <option value="two_way">Two-way (ingest + query)</option>
      </select>
    </div>
    <div class="flex gap-2 pt-2">
      <button type="submit" class="px-4 py-2 bg-accent text-bg-primary font-mono text-sm font-medium uppercase tracking-wider rounded hover:bg-accent-hover transition-colors cursor-pointer">
        Add Connection
      </button>
      <button type="button" @click="emit('cancel')" class="px-4 py-2 border border-accent-border/50 text-text-secondary font-mono text-sm uppercase tracking-wider rounded hover:border-accent/50 hover:text-text-primary transition-colors cursor-pointer">
        Cancel
      </button>
    </div>
  </form>
</template>
