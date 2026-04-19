<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { fetchPipelineLogs } from '@/api/pipeline'
import type { PipelineLogSummary } from '@/types/pipeline'

// Phase 4.1: wired picker. User picks a time window, we query
// /api/apps/{appId}/pipeline/logs, and render a clickable list of log
// journeys. Clicking a row emits `inspect` so the parent page opens the
// replay modal with the log's full journey animated through the funnel.
//
// Retention note (Phase 4.4): log_pipeline_events inherits log_buffer's
// 48h cascade-delete, so anything older than 48h is structurally
// unreachable — surfaced in the footnote under the list.

const props = defineProps<{
  appId: string
  // Optional initial logId — the /pipeline/logs/:logId deep-link route passes
  // this so we can auto-open replay on first render. The block itself doesn't
  // use it for picker state; it just surfaces it as a visible "inspecting" pill.
  initialLogId?: string | null
}>()

const emit = defineEmits<{
  inspect: [logId: string]
}>()

// Defaults: last 24h, with datetime-local values in local time (the input
// widget requires that format). We convert to UTC ISO strings when hitting
// the backend. The display/round-trip is lossy across DST boundaries but
// acceptable for a debugging UI.
function localInputValue(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

const now = new Date()
const dayAgo = new Date(now.getTime() - 24 * 60 * 60 * 1000)
const since = ref(localInputValue(dayAgo))
const until = ref(localInputValue(now))
const limit = ref(50)

const logs = ref<PipelineLogSummary[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const truncated = ref(false)
const hasSearched = ref(false)

const hasResults = computed(() => hasSearched.value && logs.value.length > 0)
const emptyState = computed(() => hasSearched.value && logs.value.length === 0)

async function run() {
  if (!props.appId) return
  loading.value = true
  error.value = null
  try {
    const res = await fetchPipelineLogs(props.appId, {
      since: new Date(since.value).toISOString(),
      until: new Date(until.value).toISOString(),
      limit: limit.value,
    })
    logs.value = res.logs
    truncated.value = res.truncated
    hasSearched.value = true
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'failed to load journeys'
  } finally {
    loading.value = false
  }
}

// Clear results when the user switches apps — otherwise the previous app's
// logs linger until the user re-runs the query.
watch(
  () => props.appId,
  () => {
    logs.value = []
    hasSearched.value = false
    error.value = null
  },
)

function inspect(logId: string) {
  emit('inspect', logId)
}

// Format helpers kept local — the codebase doesn't have a shared date-utils
// module and the choices here (compact HH:MM:SS, relative-age) are picker-
// specific anyway.
function fmtTime(iso: string): string {
  return new Date(iso).toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  })
}

function fmtDuration(firstISO: string, lastISO: string): string {
  const ms = new Date(lastISO).getTime() - new Date(firstISO).getTime()
  if (ms < 1000) return `${ms}ms`
  if (ms < 60_000) return `${(ms / 1000).toFixed(1)}s`
  return `${(ms / 60_000).toFixed(1)}m`
}

function stageIcon(n: number): string {
  // Up-to-four stages: ingestion → classified → gate → assessment. Each
  // tick shows as a solid bullet; missing ones are muted outlines so the
  // picker row communicates how far the log made it at a glance.
  return ['●', '●', '●', '●']
    .map((dot, i) => (i < n ? dot : '○'))
    .join(' ')
}
</script>

<template>
  <section
    class="relative border border-border rounded bg-bg-surface backdrop-blur-sm px-4 py-4 mb-4"
  >
    <div class="flex flex-wrap items-start gap-4">
      <div>
        <div class="flex items-center gap-2">
          <span class="font-mono text-sm font-semibold text-text-primary">Time Machine</span>
          <span
            v-if="initialLogId"
            class="font-mono text-[10px] uppercase tracking-widest text-status-info border border-status-info/40 rounded px-1.5 py-px"
          >
            replaying
          </span>
        </div>
        <p class="font-sans text-xs text-text-muted mt-1 max-w-md">
          Replay any log's pipeline journey. Retention is 48 hours — older logs are pruned with their parent row.
        </p>
      </div>

      <div class="flex flex-wrap items-end gap-3 ml-auto">
        <label class="flex flex-col gap-1">
          <span class="font-mono text-[10px] uppercase tracking-widest text-text-muted">Since</span>
          <input
            v-model="since"
            type="datetime-local"
            class="font-mono text-xs bg-bg-elevated border border-border rounded px-2 py-1 text-text-secondary focus:outline-none focus:border-accent"
          />
        </label>
        <label class="flex flex-col gap-1">
          <span class="font-mono text-[10px] uppercase tracking-widest text-text-muted">Until</span>
          <input
            v-model="until"
            type="datetime-local"
            class="font-mono text-xs bg-bg-elevated border border-border rounded px-2 py-1 text-text-secondary focus:outline-none focus:border-accent"
          />
        </label>
        <button
          type="button"
          :disabled="loading"
          class="font-mono text-xs uppercase tracking-widest border border-accent-border text-accent rounded px-3 py-1.5 hover:bg-accent/10 disabled:opacity-60 disabled:cursor-wait"
          @click="run"
        >
          {{ loading ? 'Loading…' : 'Replay' }}
        </button>
      </div>
    </div>

    <div v-if="error" class="mt-3 font-mono text-xs text-status-critical">{{ error }}</div>

    <div v-if="hasResults" class="mt-4 border-t border-border pt-3">
      <div class="font-mono text-[10px] uppercase tracking-widest text-text-muted mb-2 flex items-center gap-3">
        <span>{{ logs.length }} {{ logs.length === 1 ? 'journey' : 'journeys' }}</span>
        <span v-if="truncated" class="text-status-warn">— result truncated, narrow the window or raise the limit</span>
      </div>
      <ul class="space-y-1 max-h-72 overflow-auto font-mono text-xs">
        <li
          v-for="log in logs"
          :key="log.log_id"
          class="flex items-center gap-3 border border-border/60 rounded px-3 py-1.5 bg-bg-elevated hover:border-accent-border cursor-pointer transition-colors"
          @click="inspect(log.log_id)"
        >
          <span class="text-text-muted tabular-nums">{{ fmtTime(log.last_seen_at) }}</span>
          <span class="text-accent" :title="`${log.stage_count}/4 stages reached`">{{ stageIcon(log.stage_count) }}</span>
          <span v-if="log.source_type" class="text-text-secondary">{{ log.source_type }}</span>
          <span
            v-if="log.severity"
            class="uppercase tracking-widest text-[10px]"
            :class="log.escalated ? 'text-status-warn' : 'text-text-muted'"
          >{{ log.severity }}</span>
          <span v-if="log.type" class="text-text-secondary">{{ log.type }}<span v-if="log.category" class="text-text-muted">.{{ log.category }}</span></span>
          <span v-if="log.summary" class="text-text-muted truncate flex-1">{{ log.summary }}</span>
          <span v-else class="flex-1" />
          <span class="text-text-muted">{{ fmtDuration(log.first_seen_at, log.last_seen_at) }}</span>
          <span
            v-if="log.escalated"
            class="text-[10px] uppercase tracking-widest text-status-warn border border-status-warn/40 rounded px-1.5 py-px"
          >flagged</span>
        </li>
      </ul>
    </div>

    <div v-if="emptyState" class="mt-4 font-mono text-xs text-text-muted">
      No journeys in this window. Try a wider range, or check source filters on your connections.
    </div>
  </section>
</template>
