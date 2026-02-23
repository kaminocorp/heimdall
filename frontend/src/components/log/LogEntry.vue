<script setup lang="ts">
import { computed } from 'vue'
import type { LogEntry } from '@/types/log'

const props = defineProps<{
  entry: LogEntry
}>()

const timestamp = computed(() => {
  const d = new Date(props.entry.ingested_at)
  return d.toLocaleString()
})

const message = computed(() => {
  const p = props.entry.payload
  if (p && typeof p === 'object' && 'message' in p) {
    return String(p.message)
  }
  return JSON.stringify(p)
})
</script>

<template>
  <div class="border rounded p-3 text-sm font-mono">
    <div class="flex items-baseline gap-2">
      <span class="text-gray-400 shrink-0">{{ timestamp }}</span>
      <span v-if="entry.severity" class="font-semibold shrink-0" :class="{
        'text-red-600': entry.severity === 'critical',
        'text-yellow-600': entry.severity === 'warning',
        'text-gray-600': entry.severity === 'info',
      }">{{ entry.severity }}</span>
      <span class="text-blue-600 shrink-0">{{ entry.source_type }}</span>
    </div>
    <div class="mt-1 text-gray-800 break-words">{{ message }}</div>
  </div>
</template>
