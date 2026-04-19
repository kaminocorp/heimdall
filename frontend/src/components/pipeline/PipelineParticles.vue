<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { usePipelineStore } from '@/stores/pipeline'
import type { PipelineEvent, PipelineStage } from '@/types/pipeline'

const props = defineProps<{
  width: number
  height: number
  // Stage x positions (px) supplied by the funnel so particles ride the same
  // geometry as the SVG stream paths.
  stagePxs: number[]
  cy: number
  flaggedHalf: number
  safeHalf: number
}>()

const store = usePipelineStore()
const { events } = storeToRefs(store)

const canvasRef = ref<HTMLCanvasElement | null>(null)

// Particle struct kept allocation-free per frame: a single fixed-size pool
// that recycles slots as old particles fade. 600 slots is plenty for the
// expected throughput; sampling drops particles past that.
interface Particle {
  alive: boolean
  start: number      // ms
  duration: number   // ms
  fromX: number
  toX: number
  laneY: number
  flagged: boolean   // post-Gate-only colour split
  stage: PipelineStage
}

const POOL_SIZE = 600
const TRAVEL_MS = 1800
const SAMPLE_THRESHOLD_PER_S = 50 // beyond this we draw 1 of every N

const pool: Particle[] = Array.from({ length: POOL_SIZE }, () => ({
  alive: false,
  start: 0,
  duration: TRAVEL_MS,
  fromX: 0,
  toX: 0,
  laneY: 0,
  flagged: false,
  stage: 'ingestion',
}))

let raf: number | null = null
let lastSeenId = ''
let recentSpawns = 0
let sampleCounter = 0

function spawn(evt: PipelineEvent) {
  // Map stage → travel segment. Each particle traverses the gap between two
  // stage anchor xs so multi-stage events show as a relay race.
  const stageIndex: Record<PipelineStage, number> = {
    ingestion: 0,
    classified: 1,
    gate: 2,
    assessment: 3,
  }
  const i = stageIndex[evt.stage]
  if (i === undefined) return
  const fromX = props.stagePxs[i]
  const toX = props.stagePxs[i + 1] ?? fromX
  const flagged = evt.escalated === true || evt.stage === 'assessment'

  // Lane: pre-Gate ride the centre with slight jitter; post-Gate flagged
  // ride above centre, safe ride below — matches the SVG split.
  let laneY: number
  if (i < 2) {
    laneY = props.cy + (Math.random() * 12 - 6)
  } else if (flagged) {
    laneY = props.cy - Math.max(6, props.flaggedHalf * 0.5) - Math.random() * 4
  } else {
    laneY = props.cy + Math.max(6, props.safeHalf * 0.5) + Math.random() * 4
  }

  // Find a dead slot. Linear scan is fine at 600 — it's <1µs and beats
  // managing a free-list under reactivity.
  for (let k = 0; k < POOL_SIZE; k++) {
    if (!pool[k].alive) {
      pool[k].alive = true
      pool[k].start = performance.now()
      pool[k].fromX = fromX
      pool[k].toX = toX
      pool[k].laneY = laneY
      pool[k].flagged = flagged
      pool[k].stage = evt.stage
      return
    }
  }
  // Pool saturated — silent drop. We're already over the visual sampling
  // threshold so the user can't tell.
}

// React to new events. shallowRef on the store means we only fire when the
// reference changes — and we only spawn for the head delta since last tick.
watch(events, (next) => {
  if (next.length === 0) return
  const head = next[0]
  if (head.id === lastSeenId) return

  // Find how many new events we haven't yet drawn (newest is index 0).
  // Cap iteration to the buffer size to avoid pathological cases.
  const fresh: PipelineEvent[] = []
  for (const evt of next) {
    if (evt.id === lastSeenId) break
    fresh.push(evt)
  }
  lastSeenId = head.id
  recentSpawns += fresh.length

  // Reverse so oldest is spawned first → the visual order on screen matches
  // the ticker order.
  for (let i = fresh.length - 1; i >= 0; i--) {
    if (recentSpawns > SAMPLE_THRESHOLD_PER_S) {
      sampleCounter = (sampleCounter + 1) % 4
      if (sampleCounter !== 0) continue
    }
    spawn(fresh[i])
  }
})

// Colour palette tied to design tokens — read once, applied per particle.
const COLORS = {
  pre: 'rgba(106, 173, 122, %A)',     // accent-bright
  flagged: 'rgba(212, 168, 50, %A)',  // status-warn
  safe: 'rgba(140, 145, 155, %A)',    // text-muted
}

function colorFor(p: Particle, alpha: number): string {
  const stencil = (() => {
    if (p.stage === 'ingestion' || p.stage === 'classified') return COLORS.pre
    return p.flagged ? COLORS.flagged : COLORS.safe
  })()
  return stencil.replace('%A', alpha.toFixed(2))
}

function tick() {
  const ctx = canvasRef.value?.getContext('2d')
  if (!ctx) {
    raf = requestAnimationFrame(tick)
    return
  }
  const now = performance.now()
  ctx.clearRect(0, 0, props.width, props.height)

  // Drift the spawn counter back down so sampling auto-relaxes when traffic
  // calms — recompute once per second rather than per frame.
  if (Math.floor(now / 1000) !== Math.floor((now - 16) / 1000)) {
    recentSpawns = Math.max(0, Math.floor(recentSpawns * 0.9))
  }

  for (let k = 0; k < POOL_SIZE; k++) {
    const p = pool[k]
    if (!p.alive) continue
    const t = (now - p.start) / p.duration
    if (t >= 1) {
      p.alive = false
      continue
    }
    // Ease-out cubic — particle slows as it nears the next stage.
    const eased = 1 - Math.pow(1 - t, 3)
    const x = p.fromX + (p.toX - p.fromX) * eased
    const y = p.laneY
    const alpha = t < 0.15 ? t / 0.15 : 1 - (t - 0.85) / 0.15
    const a = Math.max(0, Math.min(1, alpha))
    ctx.fillStyle = colorFor(p, a * 0.9)
    ctx.beginPath()
    ctx.arc(x, y, 2.4, 0, Math.PI * 2)
    ctx.fill()
    // Outer glow — very faint to keep total cost low under load.
    ctx.fillStyle = colorFor(p, a * 0.18)
    ctx.beginPath()
    ctx.arc(x, y, 5.5, 0, Math.PI * 2)
    ctx.fill()
  }

  raf = requestAnimationFrame(tick)
}

const dpi = computed(() => (typeof window !== 'undefined' ? window.devicePixelRatio || 1 : 1))

onMounted(() => {
  const cv = canvasRef.value
  if (!cv) return
  // High-DPI: back the canvas at devicePixelRatio so particles stay crisp on
  // Retina displays. The CSS size stays = layout px so SVG and canvas align.
  cv.width = props.width * dpi.value
  cv.height = props.height * dpi.value
  cv.getContext('2d')?.scale(dpi.value, dpi.value)
  raf = requestAnimationFrame(tick)
})

watch(
  () => [props.width, props.height],
  ([w, h]) => {
    const cv = canvasRef.value
    if (!cv) return
    cv.width = w * dpi.value
    cv.height = h * dpi.value
    cv.getContext('2d')?.scale(dpi.value, dpi.value)
  },
)

onBeforeUnmount(() => {
  if (raf !== null) cancelAnimationFrame(raf)
  raf = null
})
</script>

<template>
  <canvas
    ref="canvasRef"
    :style="{ width: width + 'px', height: height + 'px' }"
    class="absolute inset-0 pointer-events-none"
  />
</template>
