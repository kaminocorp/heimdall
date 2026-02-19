<script setup lang="ts">
import type { LogEntry as LogEntryType } from '@/types/log'
import LogEntry from './LogEntry.vue'
import LogFilters from './LogFilters.vue'

defineProps<{
  entries: LogEntryType[]
}>()

const emit = defineEmits<{
  filter: [filters: Record<string, string>]
}>()
</script>

<template>
  <div>
    <LogFilters @filter="emit('filter', $event)" />
    <div class="mt-4 space-y-2">
      <LogEntry v-for="entry in entries" :key="entry.id" :entry="entry" />
    </div>
  </div>
</template>
