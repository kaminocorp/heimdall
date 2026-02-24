<script setup lang="ts">
import { computed } from 'vue'
import type { LogEntry as LogEntryType } from '@/types/log'
import type { Connection } from '@/types/connection'
import LogEntry from './LogEntry.vue'
import LogFilters from './LogFilters.vue'

const props = defineProps<{
  entries: LogEntryType[]
  connections: Connection[]
  total: number
  limit: number
  offset: number
}>()

const emit = defineEmits<{
  filter: [filters: { severity?: string; connection_id?: string; source?: string }]
  next: []
  prev: []
}>()

const showing = computed(() => {
  if (props.total === 0) return ''
  const start = props.offset + 1
  const end = Math.min(props.offset + props.limit, props.total)
  return `${start}–${end} of ${props.total}`
})

const hasNext = computed(() => props.offset + props.limit < props.total)
const hasPrev = computed(() => props.offset > 0)
</script>

<template>
  <div>
    <LogFilters :connections="connections" @filter="emit('filter', $event)" />

    <div v-if="entries.length === 0" class="mt-4 text-gray-500 text-sm">
      No log entries found.
    </div>

    <div v-else>
      <div class="mt-4 space-y-2">
        <LogEntry v-for="entry in entries" :key="entry.id" :entry="entry" />
      </div>

      <div class="mt-4 flex items-center justify-between text-sm text-gray-500">
        <span>{{ showing }}</span>
        <div class="flex gap-2">
          <button
            :disabled="!hasPrev"
            class="px-3 py-1 border rounded disabled:opacity-50"
            @click="emit('prev')"
          >Previous</button>
          <button
            :disabled="!hasNext"
            class="px-3 py-1 border rounded disabled:opacity-50"
            @click="emit('next')"
          >Next</button>
        </div>
      </div>
    </div>
  </div>
</template>
