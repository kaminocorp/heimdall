<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount } from 'vue'
import type { LogEntry } from '@/types/log'
import type { Connection } from '@/types/connection'

const props = defineProps<{
  entry: LogEntry
  connections: Connection[]
}>()

const emit = defineEmits<{
  close: []
}>()

const timestamp = computed(() => {
  const d = new Date(props.entry.timestamp)
  return d.toLocaleString()
})

const isAgent = computed(() => props.entry.source === 'agent')

const sourceLabel = computed(() => {
  if (!isAgent.value) return 'Raw Log'
  switch (props.entry.source_type) {
    case 'tool_call': return 'Tool Call'
    case 'tool_result': return 'Tool Result'
    case 'observation': return 'Observation'
    case 'monitoring': return 'Monitoring'
    case 'heartbeat': return 'Heartbeat'
    default: return props.entry.source_type
  }
})

const connectionName = computed(() => {
  if (!props.entry.connection_id) return null
  const conn = props.connections.find(c => c.id === props.entry.connection_id)
  return conn ? conn.name : props.entry.connection_id
})

const detailEntries = computed(() => {
  if (!props.entry.detail) return null
  const d = props.entry.detail
  // Flatten top-level keys into label/value pairs for display
  return Object.entries(d).map(([key, value]) => ({
    key,
    value: typeof value === 'object' && value !== null ? JSON.stringify(value, null, 2) : String(value ?? ''),
    isObject: typeof value === 'object' && value !== null,
  }))
})

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') emit('close')
}

onMounted(() => {
  document.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <!-- Backdrop -->
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
    @click.self="emit('close')"
  >
    <!-- Modal -->
    <div class="w-full max-w-2xl mx-4 border border-border rounded-lg bg-bg-surface shadow-2xl max-h-[85vh] flex flex-col">
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-border shrink-0">
        <div class="flex items-center gap-3">
          <h3 class="font-mono text-sm font-bold uppercase tracking-wider text-text-primary">
            Activity Detail
          </h3>
          <span
            v-if="entry.source_type === 'monitoring'"
            class="text-xs font-bold uppercase tracking-wider px-1.5 py-0.5 rounded border border-status-warn/30 bg-status-warn/10 text-status-warn"
          >Monitor</span>
          <span
            v-else-if="entry.source_type === 'heartbeat'"
            class="text-xs font-bold uppercase tracking-wider px-1.5 py-0.5 rounded border border-status-ok/30 bg-status-ok/10 text-status-ok"
          >Heartbeat</span>
          <span
            v-else-if="isAgent"
            class="text-xs font-bold uppercase tracking-wider px-1.5 py-0.5 rounded border border-accent-border/50 bg-accent-subtle text-accent"
          >Agent</span>
          <span
            v-else
            class="text-xs font-bold uppercase tracking-wider px-1.5 py-0.5 rounded border border-border text-text-muted"
          >Raw</span>
          <span
            v-if="entry.severity"
            class="text-xs font-bold uppercase tracking-wider px-1.5 py-0.5 rounded border"
            :class="{
              'border-status-critical/30 bg-status-critical/10 text-status-critical': entry.severity === 'critical',
              'border-status-warn/30 bg-status-warn/10 text-status-warn': entry.severity === 'warning',
              'border-status-info/30 bg-status-info/10 text-status-info': entry.severity === 'info',
            }"
          >{{ entry.severity }}</span>
        </div>
        <button
          @click="emit('close')"
          class="font-mono text-lg text-text-muted hover:text-text-primary transition-colors cursor-pointer leading-none"
        >&times;</button>
      </div>

      <!-- Scrollable body -->
      <div class="overflow-y-auto px-6 py-4 space-y-4 min-h-0">
        <!-- Metadata row -->
        <div class="grid grid-cols-2 gap-x-6 gap-y-2">
          <div>
            <span class="font-mono text-xs uppercase tracking-wider text-text-muted block">Timestamp</span>
            <span class="font-mono text-sm text-text-secondary">{{ timestamp }}</span>
          </div>
          <div>
            <span class="font-mono text-xs uppercase tracking-wider text-text-muted block">Source Type</span>
            <span class="font-mono text-sm text-accent-bright">{{ sourceLabel }}</span>
          </div>
          <div v-if="connectionName">
            <span class="font-mono text-xs uppercase tracking-wider text-text-muted block">Connection</span>
            <span class="font-mono text-sm text-text-secondary">{{ connectionName }}</span>
          </div>
          <div>
            <span class="font-mono text-xs uppercase tracking-wider text-text-muted block">Entry ID</span>
            <span class="font-mono text-xs text-text-muted">{{ entry.id }}</span>
          </div>
        </div>

        <!-- Summary -->
        <div>
          <span class="font-mono text-xs uppercase tracking-wider text-text-muted block mb-1">Summary</span>
          <div class="rounded border border-border bg-bg-base p-3">
            <p class="font-mono text-sm text-text-primary whitespace-pre-wrap break-words leading-relaxed">{{ entry.summary }}</p>
          </div>
        </div>

        <!-- Detail / Payload -->
        <div v-if="detailEntries && detailEntries.length > 0">
          <span class="font-mono text-xs uppercase tracking-wider text-text-muted block mb-1">
            {{ isAgent ? 'Detail' : 'Payload' }}
          </span>
          <div class="rounded border border-border bg-bg-base divide-y divide-border">
            <div
              v-for="item in detailEntries"
              :key="item.key"
              class="px-3 py-2"
            >
              <span class="font-mono text-xs uppercase tracking-wider text-accent-bright">{{ item.key }}</span>
              <pre
                v-if="item.isObject"
                class="font-mono text-xs text-text-secondary mt-1 whitespace-pre-wrap break-words overflow-x-auto max-h-60 overflow-y-auto"
              >{{ item.value }}</pre>
              <p
                v-else
                class="font-mono text-sm text-text-secondary mt-0.5 whitespace-pre-wrap break-words"
              >{{ item.value }}</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="flex justify-end px-6 py-3 border-t border-border shrink-0">
        <button
          @click="emit('close')"
          class="px-5 py-2 font-mono text-xs uppercase tracking-wider rounded border border-border text-text-primary hover:border-border-hover transition-colors cursor-pointer"
        >
          Close
        </button>
      </div>
    </div>
  </div>
</template>
