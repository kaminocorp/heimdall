<script setup lang="ts">
import { ref } from 'vue'
import type { Connection } from '@/types/connection'

defineProps<{
  connections: Connection[]
}>()

const severity = ref('')
const connectionId = ref('')
const source = ref('all')

const emit = defineEmits<{
  filter: [filters: { severity?: string; connection_id?: string; source?: string }]
}>()

function handleFilter() {
  emit('filter', {
    ...(severity.value ? { severity: severity.value } : {}),
    ...(connectionId.value ? { connection_id: connectionId.value } : {}),
    ...(source.value !== 'all' ? { source: source.value } : {}),
  })
}
</script>

<template>
  <div class="flex gap-2">
    <select v-model="source" @change="handleFilter" class="border rounded px-3 py-1 text-sm">
      <option value="all">All sources</option>
      <option value="raw">Raw logs</option>
      <option value="agent">Agent activity</option>
    </select>
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
