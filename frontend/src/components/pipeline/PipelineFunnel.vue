<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, ref } from 'vue'
import { usePipelineStore } from '@/stores/pipeline'

// Five visual stages — Activity is downstream of assessment but visualised as
// the rightmost terminal so the funnel reads end-to-end. Particles for the
// Activity stage are sourced from assessment events plus a deep-link out to
// the Activity page.
const STAGES = [
  { key: 'ingestion', label: 'Ingestion', x: 0 },
  { key: 'classified', label: 'Lumber', x: 0.25 },
  { key: 'gate', label: 'Gate', x: 0.5 },
  { key: 'assessment', label: 'Agent', x: 0.75 },
  { key: 'activity', label: 'Activity', x: 1 },
] as const

const store = usePipelineStore()
const containerEl = ref<HTMLDivElement | null>(null)
const width = ref(960)
const height = ref(280)
let observer: ResizeObserver | null = null

onMounted(() => {
  if (!containerEl.value) return
  observer = new ResizeObserver((entries) => {
    const rect = entries[0].contentRect
    width.value = Math.max(420, Math.floor(rect.width))
    // Funnel is roughly 16:5 — tall enough for stage labels to breathe but
    // not so tall that it crowds the ticker beneath.
    height.value = Math.max(220, Math.min(360, Math.floor(rect.width * 0.32)))
  })
  observer.observe(containerEl.value)
})

onBeforeUnmount(() => {
  observer?.disconnect()
  observer = null
})

// Stage x positions in pixel space, derived once per resize. The trailing
// ratio leaves a 12% margin on each side so the labels and node cards fit
// without clipping.
const stagePoints = computed(() => {
  const pad = 0.06
  return STAGES.map((s) => ({
    ...s,
    px: pad * width.value + s.x * (1 - 2 * pad) * width.value,
  }))
})

// Vertical band geometry: pre-Gate the funnel is one slab; post-Gate it
// splits into a flagged stream (top) and a safe stream (bottom). Widths come
// from the store's flagged/safe ratios. Cap min thickness at 6px so a 0%
// stream still has *some* visible body — otherwise the gradient collapses.
const cy = computed(() => height.value / 2)
const trunkHalf = computed(() => Math.max(20, height.value * 0.18))

const flaggedHalf = computed(() =>
  Math.max(3, store.flaggedRatio * trunkHalf.value * 1.1),
)
const safeHalf = computed(() =>
  Math.max(3, store.safeRatio * trunkHalf.value * 1.1),
)

// Build the SVG paths. Pre-Gate trunk: rectangle from Ingestion → Gate.
// Post-Gate split: two narrowing tapers. Sankey aesthetic without the formal
// d3-sankey dependency — five stages is too few to justify the layout pass.
const pathTrunk = computed(() => {
  const x0 = stagePoints.value[0].px
  const x1 = stagePoints.value[2].px
  const y = cy.value
  const t = trunkHalf.value
  return `M ${x0} ${y - t} L ${x1} ${y - t} L ${x1} ${y + t} L ${x0} ${y + t} Z`
})

const pathFlagged = computed(() => {
  const x0 = stagePoints.value[2].px
  const x1 = stagePoints.value[3].px
  const x2 = stagePoints.value[4].px
  const y = cy.value
  const f = flaggedHalf.value
  // Top stream — thins from full flaggedHalf at the gate to a slimmer band
  // entering the Activity stage. Bezier control points midway for the curve.
  const cx1 = (x0 + x1) / 2
  const cx2 = (x1 + x2) / 2
  const fEnd = Math.max(2, f * 0.55)
  return [
    `M ${x0} ${y - f}`,
    `C ${cx1} ${y - f}, ${cx1} ${y - f * 0.6}, ${x1} ${y - f * 0.6}`,
    `C ${cx2} ${y - f * 0.6}, ${cx2} ${y - fEnd}, ${x2} ${y - fEnd}`,
    `L ${x2} ${y}`,
    `C ${cx2} ${y}, ${cx2} ${y}, ${x1} ${y}`,
    `C ${cx1} ${y}, ${cx1} ${y}, ${x0} ${y}`,
    `Z`,
  ].join(' ')
})

const pathSafe = computed(() => {
  const x0 = stagePoints.value[2].px
  const x2 = stagePoints.value[4].px
  const y = cy.value
  const s = safeHalf.value
  const cx = (x0 + x2) / 2
  const sEnd = Math.max(2, s * 0.4)
  return [
    `M ${x0} ${y}`,
    `C ${cx} ${y}, ${cx} ${y + s * 0.4}, ${x2} ${y + sEnd}`,
    `L ${x2} ${y + sEnd + 1.5}`,
    `C ${cx} ${y + s * 0.4 + 1.5}, ${cx} ${y + 1.5}, ${x0} ${y + s}`,
    `Z`,
  ].join(' ')
})

// Connecting beam between Ingestion and the trunk start — gives the funnel
// a sense of inflow rather than appearing out of thin air.
const ingestionBeam = computed(() => {
  const x0 = stagePoints.value[0].px - 24
  const x1 = stagePoints.value[0].px
  const y = cy.value
  const t = trunkHalf.value * 0.8
  return `M ${x0} ${y - t * 0.3} L ${x1} ${y - t} L ${x1} ${y + t} L ${x0} ${y + t * 0.3} Z`
})

// Re-export geometry so the particle layer can spawn along the same paths.
defineExpose({
  width,
  height,
  stagePoints,
  cy,
  trunkHalf,
  flaggedHalf,
  safeHalf,
})
</script>

<template>
  <div ref="containerEl" class="relative w-full">
    <svg
      :width="width"
      :height="height"
      :viewBox="`0 0 ${width} ${height}`"
      class="block"
      role="img"
      aria-label="Pipeline funnel — Ingestion to Activity"
    >
      <defs>
        <!-- Pre-Gate gradient — phosphor green flow -->
        <linearGradient id="grad-pre" x1="0" x2="1" y1="0" y2="0">
          <stop offset="0%" stop-color="rgba(74, 122, 92, 0.18)" />
          <stop offset="100%" stop-color="rgba(74, 122, 92, 0.36)" />
        </linearGradient>
        <!-- Post-Gate flagged stream — amber / warn -->
        <linearGradient id="grad-flagged" x1="0" x2="1" y1="0" y2="0">
          <stop offset="0%" stop-color="rgba(212, 168, 50, 0.36)" />
          <stop offset="100%" stop-color="rgba(212, 168, 50, 0.20)" />
        </linearGradient>
        <!-- Post-Gate safe stream — muted grey, drains away -->
        <linearGradient id="grad-safe" x1="0" x2="1" y1="0" y2="0">
          <stop offset="0%" stop-color="rgba(140, 145, 155, 0.18)" />
          <stop offset="100%" stop-color="rgba(140, 145, 155, 0.05)" />
        </linearGradient>
      </defs>

      <!-- Inflow beam -->
      <path :d="ingestionBeam" fill="url(#grad-pre)" opacity="0.7" />

      <!-- Pre-Gate trunk: Ingestion → Lumber → Gate -->
      <path :d="pathTrunk" fill="url(#grad-pre)" stroke="rgba(74, 122, 92, 0.4)" stroke-width="0.5" />

      <!-- Post-Gate split -->
      <path :d="pathFlagged" fill="url(#grad-flagged)" stroke="rgba(212, 168, 50, 0.45)" stroke-width="0.5" />
      <path :d="pathSafe" fill="url(#grad-safe)" />

      <!-- Stage anchor lines — vertical guides where node cards plant -->
      <line
        v-for="s in stagePoints"
        :key="s.key"
        :x1="s.px"
        :x2="s.px"
        :y1="cy - trunkHalf - 32"
        :y2="cy + trunkHalf + 32"
        stroke="rgba(140, 145, 155, 0.15)"
        stroke-dasharray="2 4"
      />

      <!-- Stage labels along the bottom -->
      <text
        v-for="s in stagePoints"
        :key="`label-${s.key}`"
        :x="s.px"
        :y="height - 6"
        text-anchor="middle"
        class="font-mono fill-text-muted"
        style="font-size: 10px; letter-spacing: 0.18em; text-transform: uppercase;"
      >
        {{ s.label }}
      </text>
    </svg>
  </div>
</template>
