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

const paginationBtnClasses = 'px-3 py-1 border border-border rounded font-mono text-xs uppercase tracking-wider text-text-secondary hover:border-border-hover hover:text-text-primary disabled:opacity-30 disabled:cursor-not-allowed transition-colors cursor-pointer'
</script>

<template>
  <div>
    <LogFilters :connections="connections" @filter="emit('filter', $event)" />

    <div v-if="entries.length === 0" class="mt-6 text-text-muted text-sm font-mono">
      No log entries found.
    </div>

    <div v-else>
      <div class="mt-4 space-y-2">
        <LogEntry
          v-for="(entry, i) in entries"
          :key="entry.id"
          :entry="entry"
          class="animate-fade-in"
          :style="{ '--stagger-index': i }"
        />
      </div>

      <div class="mt-4 flex items-center justify-between text-sm">
        <span class="font-mono text-xs text-text-muted">{{ showing }}</span>
        <div class="flex gap-2">
          <button
            :disabled="!hasPrev"
            :class="paginationBtnClasses"
            @click="emit('prev')"
          >Prev</button>
          <button
            :disabled="!hasNext"
            :class="paginationBtnClasses"
            @click="emit('next')"
          >Next</button>
        </div>
      </div>
    </div>
  </div>
</template>
