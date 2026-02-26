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
  <div
    class="border rounded p-3 text-sm font-mono transition-colors"
    :class="isAgent
      ? 'border-l-2 border-l-accent border-t-border border-r-border border-b-border bg-accent-subtle'
      : 'border-border bg-bg-surface hover:border-border-hover'
    "
  >
    <div class="flex items-baseline gap-2 flex-wrap">
      <span class="text-text-muted text-xs shrink-0">{{ timestamp }}</span>
      <span
        v-if="isAgent"
        class="text-[10px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded border border-accent-border/50 bg-accent-subtle text-accent shrink-0"
      >Agent</span>
      <span
        v-if="entry.severity"
        class="text-[10px] font-bold uppercase tracking-wider px-1.5 py-0.5 rounded border shrink-0"
        :class="{
          'border-status-critical/30 bg-status-critical/10 text-status-critical': entry.severity === 'critical',
          'border-status-warn/30 bg-status-warn/10 text-status-warn': entry.severity === 'warning',
          'border-status-info/30 bg-status-info/10 text-status-info': entry.severity === 'info',
        }"
      >{{ entry.severity }}</span>
      <span class="text-accent-bright text-xs shrink-0">{{ entryTypeLabel }}</span>
    </div>
    <div class="mt-1.5 text-text-secondary break-words">{{ entry.summary }}</div>
  </div>
</template>
