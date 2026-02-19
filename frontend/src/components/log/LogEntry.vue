<script setup lang="ts">
import type { LogEntry } from '@/types/log'

defineProps<{
  entry: LogEntry
}>()
</script>

<template>
  <div class="border rounded p-3 text-sm font-mono">
    <span class="text-gray-400">{{ entry.ingested_at }}</span>
    <span v-if="entry.severity" class="ml-2 font-semibold" :class="{
      'text-red-600': entry.severity === 'critical',
      'text-yellow-600': entry.severity === 'warning',
      'text-gray-600': entry.severity === 'info',
    }">{{ entry.severity }}</span>
    <span class="ml-2">{{ entry.source_type }}</span>
  </div>
</template>
