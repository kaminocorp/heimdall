<script setup lang="ts">
import { ref } from 'vue'
import type { Connection } from '@/types/connection'

defineProps<{
  connections: Connection[]
}>()

const severity = ref('')
const connectionId = ref('')

const emit = defineEmits<{
  filter: [filters: { severity?: string; connection_id?: string }]
}>()

function handleFilter() {
  emit('filter', {
    ...(severity.value ? { severity: severity.value } : {}),
    ...(connectionId.value ? { connection_id: connectionId.value } : {}),
  })
}
</script>

<template>
  <div class="flex gap-2">
    <select v-model="severity" @change="handleFilter" class="border rounded px-3 py-1 text-sm">
      <option value="">All severities</option>
      <option value="info">Info</option>
      <option value="warning">Warning</option>
      <option value="critical">Critical</option>
    </select>
    <select v-model="connectionId" @change="handleFilter" class="border rounded px-3 py-1 text-sm">
      <option value="">All connections</option>
      <option v-for="conn in connections" :key="conn.id" :value="conn.id">
        {{ conn.name }}
      </option>
    </select>
  </div>
</template>
