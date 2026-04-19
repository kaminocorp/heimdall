<script setup lang="ts">
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { usePipelineStore } from '@/stores/pipeline'
import type { PipelineEvent, PipelineStage } from '@/types/pipeline'

const props = defineProps<{
  stage: PipelineStage | 'activity'
}>()

defineEmits<{
  inspect: [logId: string]
}>()

const store = usePipelineStore()
const { events, stats } = storeToRefs(store)

// Mirror events into the local computed slice. The Activity stage sources its
// content from assessment events plus a deep link, per Phase 2.6.
const sourceStage = computed<PipelineStage>(() =>
  props.stage === 'activity' ? 'assessment' : props.stage,
)

const recent = computed(() => {
  const out: PipelineEvent[] = []
  for (const evt of events.value) {
    if (evt.stage === sourceStage.value) {
      out.push(evt)
      if (out.length >= 20) break
    }
  }
  return out
})

// Per-source counts for ingestion's "where is data coming from" view.
const sourceBreakdown = computed(() => {
  if (props.stage !== 'ingestion') return []
  const counts = new Map<string, number>()
  for (const evt of events.value) {
    if (evt.stage !== 'ingestion') continue
    const k = evt.source_type || '—'
    counts.set(k, (counts.get(k) ?? 0) + 1)
  }
  return [...counts.entries()].sort((a, b) => b[1] - a[1]).slice(0, 6)
})

// Lumber type/category histogram (top 6).
const classificationDist = computed(() => {
  if (props.stage !== 'classified') return []
  const counts = new Map<string, number>()
  for (const evt of events.value) {
    if (evt.stage !== 'classified') continue
    const k = `${evt.type ?? '?'}.${evt.category ?? '?'}`
    counts.set(k, (counts.get(k) ?? 0) + 1)
  }
  return [...counts.entries()].sort((a, b) => b[1] - a[1]).slice(0, 6)
})

// Per-rule hit counts for the Gate panel.
const ruleHits = computed(() => {
  if (props.stage !== 'gate') return []
  const counts = new Map<string, number>()
  for (const evt of events.value) {
    if (evt.stage !== 'gate' || !evt.escalated) continue
    const k = evt.rule_hit || 'unknown'
    counts.set(k, (counts.get(k) ?? 0) + 1)
  }
  return [...counts.entries()].sort((a, b) => b[1] - a[1])
})

// Severity breakdown for the Activity panel — drawn from assessment events.
const severityDist = computed(() => {
  if (props.stage !== 'activity') return []
  const counts = new Map<string, number>()
  for (const evt of events.value) {
    if (evt.stage !== 'assessment') continue
    const k = evt.severity || '—'
    counts.set(k, (counts.get(k) ?? 0) + 1)
  }
  return [...counts.entries()]
})

const flaggedPct = computed(() => {
  const total = stats.value.flagged_count + stats.value.safe_count
  if (total === 0) return 0
  return (stats.value.flagged_count / total) * 100
})

function formatTime(iso: string): string {
  return iso.slice(11, 19)
}
</script>

<template>
  <div class="border border-border rounded bg-bg-surface px-4 py-4 mt-3">
    <!-- ── Ingestion ── -->
    <template v-if="stage === 'ingestion'">
      <div class="flex flex-wrap gap-6 text-sm">
        <div>
          <div class="font-mono text-xs uppercase tracking-widest text-text-muted mb-2">Sources (last buffered)</div>
          <ul class="space-y-1 font-mono text-xs">
            <li v-for="[name, count] in sourceBreakdown" :key="name" class="flex justify-between gap-6 text-text-secondary">
              <span>{{ name }}</span>
              <span class="text-text-primary">{{ count }}</span>
            </li>
            <li v-if="sourceBreakdown.length === 0" class="text-text-muted">No ingestion events yet.</li>
          </ul>
        </div>
        <div class="flex-1 min-w-[280px]">
          <div class="font-mono text-xs uppercase tracking-widest text-text-muted mb-2">Last 20 ingestions</div>
          <ul class="space-y-1 font-mono text-xs max-h-[180px] overflow-y-auto">
            <li
              v-for="evt in recent"
              :key="evt.id"
              class="flex items-baseline gap-3 text-text-secondary cursor-pointer hover:text-accent-bright"
              @click="$emit('inspect', evt.log_id)"
            >
              <span class="text-text-muted shrink-0">{{ formatTime(evt.occurred_at) }}</span>
              <span class="shrink-0 uppercase text-[10px] tracking-widest" :class="evt.severity === 'error' ? 'text-status-critical' : evt.severity === 'warn' ? 'text-status-warn' : 'text-text-muted'">
                {{ evt.severity || '—' }}
              </span>
              <span class="truncate">{{ evt.source_type || '—' }}</span>
            </li>
          </ul>
        </div>
      </div>
    </template>

    <!-- ── Lumber ── -->
    <template v-else-if="stage === 'classified'">
      <div class="flex flex-wrap gap-6 text-sm">
        <div class="min-w-[220px]">
          <div class="font-mono text-xs uppercase tracking-widest text-text-muted mb-2">Type.Category</div>
          <ul class="space-y-1 font-mono text-xs">
            <li v-for="[k, v] in classificationDist" :key="k" class="flex justify-between gap-6 text-text-secondary">
              <span>{{ k }}</span>
              <span class="text-text-primary">{{ v }}</span>
            </li>
            <li v-if="classificationDist.length === 0" class="text-text-muted">No classifications yet.</li>
          </ul>
        </div>
        <div class="flex-1 min-w-[280px]">
          <div class="font-mono text-xs uppercase tracking-widest text-text-muted mb-2">Last 20 classifications</div>
          <ul class="space-y-1 font-mono text-xs max-h-[180px] overflow-y-auto">
            <li
              v-for="evt in recent"
              :key="evt.id"
              class="flex items-baseline gap-3 text-text-secondary cursor-pointer hover:text-accent-bright"
              @click="$emit('inspect', evt.log_id)"
            >
              <span class="text-text-muted shrink-0">{{ formatTime(evt.occurred_at) }}</span>
              <span class="text-accent-bright shrink-0">{{ evt.type || '?' }}.{{ evt.category || '?' }}</span>
              <span class="text-text-muted shrink-0">{{ evt.confidence ? evt.confidence.toFixed(2) : '—' }}</span>
              <span class="truncate">{{ evt.summary || '' }}</span>
            </li>
          </ul>
        </div>
      </div>
    </template>

    <!-- ── Gate ── -->
    <template v-else-if="stage === 'gate'">
      <div class="flex flex-wrap gap-6 text-sm">
        <div class="min-w-[220px]">
          <div class="font-mono text-xs uppercase tracking-widest text-text-muted mb-2">Escalation rules fired</div>
          <ul class="space-y-1 font-mono text-xs">
            <li v-for="[name, count] in ruleHits" :key="name" class="flex justify-between gap-6 text-text-secondary">
              <span>{{ name }}</span>
              <span class="text-status-warn">{{ count }}</span>
            </li>
            <li v-if="ruleHits.length === 0" class="text-text-muted">No escalations in buffer.</li>
          </ul>
          <div class="mt-3 font-mono text-[11px] text-text-muted">
            Flagged share: <span class="text-status-warn">{{ flaggedPct.toFixed(1) }}%</span>
          </div>
        </div>
        <div class="flex-1 min-w-[280px]">
          <div class="font-mono text-xs uppercase tracking-widest text-text-muted mb-2">Last 20 gate decisions</div>
          <ul class="space-y-1 font-mono text-xs max-h-[180px] overflow-y-auto">
            <li
              v-for="evt in recent"
              :key="evt.id"
              class="flex items-baseline gap-3 text-text-secondary cursor-pointer hover:text-accent-bright"
              @click="$emit('inspect', evt.log_id)"
            >
              <span class="text-text-muted shrink-0">{{ formatTime(evt.occurred_at) }}</span>
              <span class="shrink-0 uppercase text-[10px] tracking-widest" :class="evt.escalated ? 'text-status-warn' : 'text-text-muted'">
                {{ evt.escalated ? 'flagged' : 'safe' }}
              </span>
              <span class="truncate">{{ evt.rule_hit || '—' }}</span>
            </li>
          </ul>
        </div>
      </div>
    </template>

    <!-- ── Agent (assessment) ── -->
    <template v-else-if="stage === 'assessment'">
      <div>
        <div class="font-mono text-xs uppercase tracking-widest text-text-muted mb-2">Recent assessments</div>
        <ul class="space-y-1 font-mono text-xs max-h-[220px] overflow-y-auto">
          <li
            v-for="evt in recent"
            :key="evt.id"
            class="flex items-baseline gap-3 text-text-secondary cursor-pointer hover:text-accent-bright"
            @click="$emit('inspect', evt.log_id)"
          >
            <span class="text-text-muted shrink-0">{{ formatTime(evt.occurred_at) }}</span>
            <span class="text-accent-bright shrink-0">{{ evt.severity || '—' }}</span>
            <span class="truncate">{{ evt.summary || evt.assessment_id || '—' }}</span>
          </li>
          <li v-if="recent.length === 0" class="text-text-muted">No assessments in buffer.</li>
        </ul>
      </div>
    </template>

    <!-- ── Activity ── -->
    <template v-else>
      <div class="flex flex-wrap gap-6 text-sm">
        <div class="min-w-[200px]">
          <div class="font-mono text-xs uppercase tracking-widest text-text-muted mb-2">Severity (recent)</div>
          <ul class="space-y-1 font-mono text-xs">
            <li v-for="[name, count] in severityDist" :key="name" class="flex justify-between gap-6 text-text-secondary">
              <span>{{ name }}</span>
              <span class="text-text-primary">{{ count }}</span>
            </li>
            <li v-if="severityDist.length === 0" class="text-text-muted">No assessments yet.</li>
          </ul>
        </div>
        <div class="flex items-center font-mono text-xs">
          <RouterLink to="/activity" class="text-accent-bright hover:underline">
            Open Activity →
          </RouterLink>
        </div>
      </div>
    </template>
  </div>
</template>
