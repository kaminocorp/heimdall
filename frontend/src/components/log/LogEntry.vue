<script setup lang="ts">
import { computed } from 'vue'
import type { LogEntry } from '@/types/log'

const props = defineProps<{
  entry: LogEntry
}>()

const timestamp = computed(() => {
  const d = new Date(props.entry.timestamp)
  return d.toLocaleString()
})

const isAgent = computed(() => props.entry.source === 'agent')

const entryTypeLabel = computed(() => {
  if (!isAgent.value) return props.entry.source_type
  switch (props.entry.source_type) {
    case 'tool_call': return 'Tool Call'
    case 'tool_result': return 'Tool Result'
    case 'observation': return 'Observation'
    default: return props.entry.source_type
  }
})
</script>

<template>
  <div class="border rounded p-3 text-sm font-mono" :class="{
    'border-purple-200 bg-purple-50/50': isAgent,
  }">
    <div class="flex items-baseline gap-2">
      <span class="text-gray-400 shrink-0">{{ timestamp }}</span>
      <span v-if="isAgent" class="text-purple-600 text-xs font-semibold shrink-0">AGENT</span>
      <span v-if="entry.severity" class="font-semibold shrink-0" :class="{
        'text-red-600': entry.severity === 'critical',
        'text-yellow-600': entry.severity === 'warning',
        'text-gray-600': entry.severity === 'info',
      }">{{ entry.severity }}</span>
      <span class="text-blue-600 shrink-0">{{ entryTypeLabel }}</span>
    </div>
    <div class="mt-1 text-gray-800 break-words">{{ entry.summary }}</div>
  </div>
</template>
