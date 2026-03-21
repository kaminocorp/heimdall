<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import type { Connection } from '@/types/connection'
import BaseSelect from '@/components/common/BaseSelect.vue'

const props = defineProps<{
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

const sourceOptions = [
  { value: 'all', label: 'All sources' },
  { value: 'raw', label: 'Raw logs' },
  { value: 'agent', label: 'Agent activity' },
]

const severityOptions = [
  { value: '', label: 'All severities' },
  { value: 'info', label: 'Info' },
  { value: 'warning', label: 'Warning' },
  { value: 'critical', label: 'Critical' },
]

const connectionOptions = computed(() => [
  { value: '', label: 'All connections' },
  ...props.connections.map(c => ({ value: c.id, label: c.name })),
])

watch(source, handleFilter)
watch(severity, handleFilter)
watch(connectionId, handleFilter)
</script>

<template>
  <div class="flex flex-wrap gap-2">
    <BaseSelect v-model="source" :options="sourceOptions" size="sm" />
    <BaseSelect v-model="severity" :options="severityOptions" size="sm" />
    <BaseSelect v-model="connectionId" :options="connectionOptions" size="sm" />
  </div>
</template>
