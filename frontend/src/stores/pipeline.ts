import { defineStore } from 'pinia'
import { computed, ref, shallowRef, triggerRef } from 'vue'
import type {
  PipelineEvent,
  PipelineJourney,
  PipelineStats,
} from '@/types/pipeline'
import { fetchLogJourney } from '@/api/pipeline'

// Cap on the live ticker buffer. 500 lines is what the user can actually
// scroll through; past that we evict oldest-first. The matching Set<string>
// lets `insertEvent` do O(1) dedupe, which is what makes bootstrap→replay→live
// overlap safe to merge.
const RING_BUFFER_CAP = 500

// Per-stage rolling rates recompute every 1s. Anything faster churns reactivity
// without the user noticing; anything slower makes the Sankey widths feel stale.
const RATE_WINDOW_MS = 60_000
const RATE_RECOMPUTE_MS = 1000

// LRU for per-log journey modals. Re-clicking the same row in the ticker
// should not re-fetch.
const JOURNEY_CACHE_CAP = 100

const EMPTY_STATS: PipelineStats = {
  ingestion_count: 0,
  classified_count: 0,
  flagged_count: 0,
  safe_count: 0,
  assessment_count: 0,
  avg_confidence: 0,
  window_seconds: 3600,
}

export const usePipelineStore = defineStore('pipeline', () => {
  // Stats snapshot — replaced wholesale on bootstrap/refresh. Fine as a deep
  // ref since the object is tiny and replaced, not mutated in place.
  const stats = ref<PipelineStats>({ ...EMPTY_STATS })

  // Ring buffer held in a shallowRef so Vue doesn't deep-watch hundreds of
  // events. We trigger reactivity manually after each mutation; consumers
  // read .value, and template v-for over it picks up changes.
  const events = shallowRef<PipelineEvent[]>([])
  const eventIds = new Set<string>()

  // Per-stage event timestamps (ms) for rolling-rate calculation. Oldest
  // entries drain naturally when the recompute-tick window trims them.
  const stageTimestamps: Record<string, number[]> = {
    ingestion: [],
    classified: [],
    gate: [],
    assessment: [],
  }
  const rates = ref<Record<string, number>>({
    ingestion: 0,
    classified: 0,
    gate: 0,
    assessment: 0,
  })
  let rateTimer: ReturnType<typeof setInterval> | null = null

  // Per-log journey LRU — insertion-order Map gives us eviction for free.
  const journeyCache = new Map<string, PipelineJourney>()

  // ── Derived ratios for Sankey width calculation ──
  //
  // Use flagged/safe counts over the stats window. When the window is empty
  // default to 50/50 so the funnel still renders with *some* geometry and
  // doesn't collapse to a zero-width path on first paint.
  const flaggedRatio = computed(() => {
    const total = stats.value.flagged_count + stats.value.safe_count
    if (total === 0) return 0.5
    return stats.value.flagged_count / total
  })

  const safeRatio = computed(() => 1 - flaggedRatio.value)

  // Escalation ratio — what % of ingested logs end up in an assessment. Used
  // by the funnel's overall narrow-ing slope between Ingestion and Agent.
  const assessmentRatio = computed(() => {
    if (stats.value.ingestion_count === 0) return 0
    return stats.value.assessment_count / stats.value.ingestion_count
  })

  // ── Mutations ──

  function setStats(next: PipelineStats) {
    stats.value = next
  }

  function insertEvent(evt: PipelineEvent): boolean {
    if (eventIds.has(evt.id)) return false

    eventIds.add(evt.id)
    // Prepend so newest is at index 0 — the ticker renders top-down.
    const next = [evt, ...events.value]
    if (next.length > RING_BUFFER_CAP) {
      const evicted = next.pop() as PipelineEvent
      eventIds.delete(evicted.id)
    }
    events.value = next
    triggerRef(events)

    // Record timestamp for the rolling rate. Parse once at insert time so
    // the rate-recompute tick doesn't repeat the Date.parse for every event.
    const bucket = stageTimestamps[evt.stage]
    if (bucket) {
      bucket.push(Date.parse(evt.occurred_at))
    }
    return true
  }

  function mergeEvents(batch: PipelineEvent[]) {
    // Bootstrap delivers newest-first; iterate in reverse so each insert's
    // "prepend" keeps the overall order right.
    for (let i = batch.length - 1; i >= 0; i--) {
      insertEvent(batch[i])
    }
  }

  // ── Rolling per-stage rates ──

  function recomputeRates() {
    const cutoff = Date.now() - RATE_WINDOW_MS
    const next = { ...rates.value }
    for (const stage of Object.keys(stageTimestamps)) {
      const bucket = stageTimestamps[stage]
      // Trim in-place — bucket is naturally monotonic (we only append) so a
      // simple prefix cut is enough. No need to scan.
      let trimFrom = 0
      while (trimFrom < bucket.length && bucket[trimFrom] < cutoff) {
        trimFrom++
      }
      if (trimFrom > 0) bucket.splice(0, trimFrom)
      next[stage] = bucket.length / 60 // events/second over the window
    }
    rates.value = next
  }

  function startRates() {
    if (rateTimer) return
    rateTimer = setInterval(recomputeRates, RATE_RECOMPUTE_MS)
  }

  function stopRates() {
    if (rateTimer) {
      clearInterval(rateTimer)
      rateTimer = null
    }
  }

  // ── Journey cache (LRU) ──

  function cachedJourney(logId: string): PipelineJourney | undefined {
    const hit = journeyCache.get(logId)
    if (hit) {
      // Touch for LRU semantics — re-insertion moves it to the end.
      journeyCache.delete(logId)
      journeyCache.set(logId, hit)
    }
    return hit
  }

  async function loadJourney(appId: string, logId: string): Promise<PipelineJourney> {
    const hit = cachedJourney(logId)
    if (hit) return hit
    const fresh = await fetchLogJourney(appId, logId)
    if (journeyCache.size >= JOURNEY_CACHE_CAP) {
      // Evict least-recently-used (oldest insertion order).
      const first = journeyCache.keys().next().value
      if (first) journeyCache.delete(first)
    }
    journeyCache.set(logId, fresh)
    return fresh
  }

  // ── Reset on app switch ──

  function reset() {
    stats.value = { ...EMPTY_STATS }
    events.value = []
    eventIds.clear()
    for (const stage of Object.keys(stageTimestamps)) {
      stageTimestamps[stage] = []
    }
    rates.value = { ingestion: 0, classified: 0, gate: 0, assessment: 0 }
    journeyCache.clear()
  }

  return {
    stats,
    events,
    rates,
    flaggedRatio,
    safeRatio,
    assessmentRatio,
    setStats,
    insertEvent,
    mergeEvents,
    startRates,
    stopRates,
    loadJourney,
    reset,
  }
})
