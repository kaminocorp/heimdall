<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { usePipelineStore } from '@/stores/pipeline'
import type { PipelineEvent, PipelineStage } from '@/types/pipeline'

// Phase 4.2: replay view. Upgraded from the Phase 2 timeline-only stub.
//
// The modal now renders a scaled Sankey funnel that echoes the live page's
// five-stage geometry, plus a traced particle that animates through the
// stages this specific log visited. Each stage node lights up as the
// particle reaches it; stages the log never reached (e.g. assessment
// when the gate dropped it) stay dimmed so you can read the journey
// visually at a glance.
//
// Why a self-contained Sankey instead of reusing PipelineFunnel: the live
// funnel reads ratios off the store, which keep changing as events stream
// in. The replay backdrop should be a static visual that represents the
// *log's* journey, not the window's rolling flagged/safe split.

const props = defineProps<{
  appId: string
  logId: string | null
}>()

const emit = defineEmits<{
  close: []
}>()

const store = usePipelineStore()
const events = ref<PipelineEvent[]>([])
const loading = ref(false)
const error = ref<string | null>(null)

// Stage reach computed from the journey. 'activity' is present iff the log
// has an assessment event (activity is the downstream view).
const STAGE_ORDER: (PipelineStage | 'activity')[] = [
  'ingestion',
  'classified',
  'gate',
  'assessment',
  'activity',
]

const stageReached = computed<Record<string, boolean>>(() => {
  const seen: Record<string, boolean> = {}
  for (const evt of events.value) {
    seen[evt.stage] = true
  }
  // Activity mirrors assessment — if the log got an assessment, it's in
  // the activity feed downstream.
  if (seen.assessment) seen.activity = true
  return seen
})

const gateDecision = computed<'flagged' | 'safe' | null>(() => {
  const gate = events.value.find((e) => e.stage === 'gate')
  if (!gate || gate.escalated === undefined) return null
  return gate.escalated ? 'flagged' : 'safe'
})

// ── Replay animation ──
// Kept intentionally simple: advance `activeStage` through STAGE_ORDER on a
// fixed cadence. No interpolation between stages; the particle jumps from
// anchor to anchor. The visual becomes a guided reading of the journey
// rather than a physics sim, which is what the auditing surface needs.
const REPLAY_STAGE_MS = 650
const activeStage = ref<number>(-1)
const isPlaying = ref(false)
let replayTimer: ReturnType<typeof setTimeout> | null = null

function stopReplay() {
  if (replayTimer) {
    clearTimeout(replayTimer)
    replayTimer = null
  }
  isPlaying.value = false
}

function playReplay() {
  stopReplay()
  activeStage.value = -1
  isPlaying.value = true

  // Find the furthest stage this log reached — we stop the animation there.
  // A gate-safe log never reached assessment or activity, for example.
  let maxIndex = -1
  for (let i = 0; i < STAGE_ORDER.length; i++) {
    if (stageReached.value[STAGE_ORDER[i]]) maxIndex = i
  }
  if (maxIndex < 0) {
    isPlaying.value = false
    return
  }

  const step = (i: number) => {
    activeStage.value = i
    if (i < maxIndex) {
      replayTimer = setTimeout(() => step(i + 1), REPLAY_STAGE_MS)
    } else {
      replayTimer = setTimeout(() => {
        isPlaying.value = false
      }, REPLAY_STAGE_MS)
    }
  }
  step(0)
}

watch(
  () => props.logId,
  async (id) => {
    stopReplay()
    activeStage.value = -1
    if (!id) {
      events.value = []
      return
    }
    loading.value = true
    error.value = null
    events.value = []
    try {
      const j = await store.loadJourney(props.appId, id)
      events.value = j.events
      // Auto-play on load so the user sees the animation without clicking.
      playReplay()
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'failed to load journey'
    } finally {
      loading.value = false
    }
  },
  { immediate: true },
)

// ── Layout geometry for the scaled Sankey ──
const SVG_W = 640
const SVG_H = 140
const PAD = 40
const stageX = (i: number) => PAD + (i * (SVG_W - 2 * PAD)) / (STAGE_ORDER.length - 1)
const CY = SVG_H / 2

const particleX = computed(() => {
  if (activeStage.value < 0) return stageX(0)
  return stageX(activeStage.value)
})

// Gate fork: if flagged, particle rides the top (amber) lane after gate;
// if safe, particle rides the bottom (muted grey) lane.
const particleY = computed(() => {
  if (activeStage.value < 2 || !gateDecision.value) return CY
  return gateDecision.value === 'flagged' ? CY - 18 : CY + 18
})

function stageLabel(key: string): string {
  return key === 'classified' ? 'Lumber'
    : key === 'assessment' ? 'Agent'
    : key.charAt(0).toUpperCase() + key.slice(1)
}

function fmt(iso: string): string {
  return iso.slice(0, 19).replace('T', ' ')
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="logId"
      class="fixed inset-0 z-50 flex items-center justify-center p-6 bg-black/60 backdrop-blur-sm"
      @click.self="emit('close')"
    >
      <div class="w-full max-w-3xl bg-bg-elevated border border-border rounded-lg shadow-xl max-h-[90vh] flex flex-col">
        <div class="flex items-center justify-between px-5 py-3 border-b border-border">
          <div>
            <div class="font-mono text-xs uppercase tracking-widest text-text-muted">Pipeline replay</div>
            <div class="font-mono text-sm text-text-primary truncate" style="max-width: 28rem;">{{ logId }}</div>
          </div>
          <div class="flex items-center gap-3">
            <button
              type="button"
              :disabled="isPlaying || loading || events.length === 0"
              class="font-mono text-xs uppercase tracking-widest border border-accent-border text-accent rounded px-3 py-1 hover:bg-accent/10 disabled:opacity-60 disabled:cursor-not-allowed"
              @click="playReplay"
            >
              {{ isPlaying ? 'Replaying…' : 'Replay' }}
            </button>
            <button
              type="button"
              class="font-mono text-xs text-text-secondary hover:text-text-primary"
              @click="emit('close')"
            >
              close
            </button>
          </div>
        </div>

        <div class="flex-1 overflow-auto">
          <!-- Sankey + traced particle -->
          <div v-if="!loading && events.length > 0" class="px-5 py-4 border-b border-border">
            <svg
              :viewBox="`0 0 ${SVG_W} ${SVG_H}`"
              class="w-full"
              role="img"
              aria-label="Log journey through the pipeline"
            >
              <defs>
                <linearGradient id="grad-replay-pre" x1="0" x2="1" y1="0" y2="0">
                  <stop offset="0%" stop-color="rgba(74, 122, 92, 0.15)" />
                  <stop offset="100%" stop-color="rgba(74, 122, 92, 0.3)" />
                </linearGradient>
                <linearGradient id="grad-replay-flagged" x1="0" x2="1" y1="0" y2="0">
                  <stop offset="0%" stop-color="rgba(212, 168, 50, 0.3)" />
                  <stop offset="100%" stop-color="rgba(212, 168, 50, 0.15)" />
                </linearGradient>
                <linearGradient id="grad-replay-safe" x1="0" x2="1" y1="0" y2="0">
                  <stop offset="0%" stop-color="rgba(140, 145, 155, 0.18)" />
                  <stop offset="100%" stop-color="rgba(140, 145, 155, 0.05)" />
                </linearGradient>
              </defs>

              <!-- Pre-gate trunk -->
              <rect
                :x="stageX(0)"
                :y="CY - 12"
                :width="stageX(2) - stageX(0)"
                height="24"
                fill="url(#grad-replay-pre)"
              />

              <!-- Post-gate lanes -->
              <path
                :d="`M ${stageX(2)} ${CY - 12} L ${stageX(4)} ${CY - 24} L ${stageX(4)} ${CY - 12} L ${stageX(2)} ${CY} Z`"
                fill="url(#grad-replay-flagged)"
              />
              <path
                :d="`M ${stageX(2)} ${CY} L ${stageX(4)} ${CY + 12} L ${stageX(4)} ${CY + 24} L ${stageX(2)} ${CY + 12} Z`"
                fill="url(#grad-replay-safe)"
              />

              <!-- Stage nodes. Filled when the particle has passed; outlined otherwise. -->
              <g v-for="(key, i) in STAGE_ORDER" :key="key">
                <circle
                  :cx="stageX(i)"
                  :cy="CY"
                  r="9"
                  :class="[
                    'transition-colors duration-200',
                    activeStage >= i ? 'fill-accent' : 'fill-bg-surface',
                  ]"
                  stroke="rgba(140, 145, 155, 0.6)"
                  stroke-width="1"
                />
                <text
                  :x="stageX(i)"
                  :y="CY + 32"
                  text-anchor="middle"
                  class="font-mono"
                  :class="activeStage >= i ? 'fill-text-primary' : 'fill-text-muted'"
                  style="font-size: 10px; letter-spacing: 0.14em; text-transform: uppercase;"
                >{{ stageLabel(key) }}</text>
                <text
                  v-if="!stageReached[key]"
                  :x="stageX(i)"
                  :y="CY - 22"
                  text-anchor="middle"
                  class="fill-text-muted"
                  style="font-size: 9px;"
                >—</text>
              </g>

              <!-- Traced particle. Uses CSS transition to glide between anchors. -->
              <circle
                :cx="particleX"
                :cy="particleY"
                r="5"
                class="fill-accent-bright transition-all duration-500 ease-out"
                style="filter: drop-shadow(0 0 6px rgba(90, 158, 106, 0.85));"
              />
            </svg>
            <div v-if="gateDecision" class="mt-2 font-mono text-[10px] uppercase tracking-widest text-text-muted">
              gate:
              <span :class="gateDecision === 'flagged' ? 'text-status-warn' : 'text-text-muted'">
                {{ gateDecision }}
              </span>
            </div>
          </div>

          <!-- Timeline detail -->
          <div class="px-5 py-4">
            <div v-if="loading" class="font-mono text-xs text-text-muted">Loading…</div>
            <div v-else-if="error" class="font-mono text-xs text-status-critical">{{ error }}</div>
            <div v-else-if="events.length === 0" class="font-mono text-xs text-text-muted">
              No events recorded for this log (yet).
            </div>

            <ul v-else class="space-y-2">
              <li
                v-for="evt in events"
                :key="evt.id"
                class="font-mono text-xs border border-border rounded px-3 py-2 bg-bg-surface"
              >
                <div class="flex items-baseline gap-3">
                  <span class="text-text-muted">{{ fmt(evt.occurred_at) }}</span>
                  <span class="uppercase tracking-widest text-accent-bright">{{ evt.stage }}</span>
                </div>
                <div class="mt-1 text-text-secondary space-y-0.5">
                  <div v-if="evt.source_type">source: {{ evt.source_type }}</div>
                  <div v-if="evt.severity">severity: {{ evt.severity }}</div>
                  <div v-if="evt.type || evt.category">
                    classification: {{ evt.type || '?' }}.{{ evt.category || '?' }}
                    <span v-if="evt.confidence !== undefined" class="text-text-muted">
                      ({{ evt.confidence.toFixed(2) }})
                    </span>
                  </div>
                  <div v-if="evt.summary" class="truncate">summary: {{ evt.summary }}</div>
                  <div v-if="evt.escalated !== undefined">
                    decision: <span :class="evt.escalated ? 'text-status-warn' : 'text-text-muted'">
                      {{ evt.escalated ? 'flagged' : 'safe' }}
                    </span>
                    <span v-if="evt.rule_hit" class="text-text-muted"> · {{ evt.rule_hit }}</span>
                  </div>
                  <div v-if="evt.assessment_id">
                    assessment: <span class="text-status-info">{{ evt.assessment_id }}</span>
                  </div>
                </div>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>
