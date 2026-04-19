<script setup lang="ts">
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { usePipelineStore } from '@/stores/pipeline'
import type { PipelineEvent } from '@/types/pipeline'

defineEmits<{
  inspect: [logId: string]
}>()

const store = usePipelineStore()
const { events } = storeToRefs(store)

// Render the most recent 80 — the ring buffer holds 500 but the ticker only
// needs the readable head. v-for over the truncated slice keeps DOM size
// bounded under high-throughput streams.
const visible = computed(() => events.value.slice(0, 80))

function fmtTime(iso: string): string {
  return iso.slice(11, 19)
}

function stageColor(stage: PipelineEvent['stage']): string {
  switch (stage) {
    case 'ingestion': return 'text-text-muted'
    case 'classified': return 'text-accent-bright'
    case 'gate': return 'text-status-warn'
    case 'assessment': return 'text-status-info'
  }
}

function stageLabel(stage: PipelineEvent['stage']): string {
  switch (stage) {
    case 'ingestion': return 'INGEST'
    case 'classified': return 'LUMBER'
    case 'gate': return 'GATE'
    case 'assessment': return 'AGENT'
  }
}
</script>

<template>
  <div class="border border-border rounded bg-bg-elevated">
    <div class="flex items-center justify-between px-4 py-2 border-b border-border">
      <span class="font-mono text-xs uppercase tracking-widest text-text-muted">Live ticker</span>
      <span class="font-mono text-[10px] text-text-muted">{{ visible.length }} / {{ events.length }}</span>
    </div>
    <div class="max-h-[280px] overflow-y-auto font-mono text-xs">
      <div
        v-for="evt in visible"
        :key="evt.id"
        class="flex items-baseline gap-3 px-4 py-1 border-b border-border/40 hover:bg-bg-surface-hover cursor-pointer transition-colors"
        @click="$emit('inspect', evt.log_id)"
      >
        <span class="text-text-muted shrink-0 w-[68px]">{{ fmtTime(evt.occurred_at) }}</span>
        <span :class="stageColor(evt.stage)" class="shrink-0 w-[58px] uppercase tracking-widest">
          {{ stageLabel(evt.stage) }}
        </span>
        <span class="truncate text-text-secondary">
          <template v-if="evt.stage === 'ingestion'">
            {{ evt.source_type || '—' }} · {{ evt.severity || '—' }}
          </template>
          <template v-else-if="evt.stage === 'classified'">
            {{ evt.type || '?' }}.{{ evt.category || '?' }} · {{ evt.summary || '' }}
          </template>
          <template v-else-if="evt.stage === 'gate'">
            <span :class="evt.escalated ? 'text-status-warn' : 'text-text-muted'">
              {{ evt.escalated ? 'flagged' : 'safe' }}
            </span>
            · {{ evt.rule_hit || '—' }}
          </template>
          <template v-else>
            {{ evt.severity || '—' }} · {{ evt.summary || evt.assessment_id || '' }}
          </template>
        </span>
      </div>
      <div v-if="visible.length === 0" class="px-4 py-6 text-text-muted">
        Waiting for events…
      </div>
    </div>
  </div>
</template>
