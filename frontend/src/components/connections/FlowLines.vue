<script setup lang="ts">
/**
 * FlowLines — SVG overlay rendering animated curved arrow paths from each
 * connection bubble down to the agent nebula canvas.
 *
 * Paths are cubic beziers computed from DOM positions via
 * getBoundingClientRect(). A ResizeObserver recomputes on layout changes.
 *
 * Each category gets a distinct visual treatment:
 * - Ingestion: arrowhead at nebula end, dashes flow downward
 * - Enrichment: arrowheads on both ends, dashes pulse bidirectionally
 * - Outbound: arrowhead at bubble end, dashes flow upward from nebula
 */

import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import type { ConnectionCategory } from './wizard/flows'

const props = defineProps<{
  /** Refs to each bubble element's root DOM node */
  bubbleEls: (HTMLElement | null)[]
  /** Ref to the nebula container element */
  nebulaEl: HTMLElement | null
  /** Total connection count — used to trigger recompute on add/remove */
  count: number
  /** Category per bubble index — drives line style */
  categories?: ConnectionCategory[]
}>()

interface LinePath {
  d: string
  index: number
  category: ConnectionCategory
}

const containerRef = ref<HTMLElement | null>(null)
const lines = ref<LinePath[]>([])
const svgWidth = ref(0)
const svgHeight = ref(0)

function computeLines() {
  if (!containerRef.value || !props.nebulaEl) return

  const containerRect = containerRef.value.getBoundingClientRect()
  svgWidth.value = containerRect.width
  svgHeight.value = containerRect.height

  // Target: top-center of the nebula
  const nebulaRect = props.nebulaEl.getBoundingClientRect()
  const targetX = nebulaRect.left + nebulaRect.width / 2 - containerRect.left
  const targetY = nebulaRect.top + 20 - containerRect.top // slightly into the nebula

  const newLines: LinePath[] = []

  for (let i = 0; i < props.bubbleEls.length; i++) {
    const el = props.bubbleEls[i]
    if (!el) continue

    const bubbleRect = el.getBoundingClientRect()
    // Source: bottom-center of each bubble
    const sourceX = bubbleRect.left + bubbleRect.width / 2 - containerRect.left
    const sourceY = bubbleRect.bottom - containerRect.top

    const category = props.categories?.[i] ?? 'ingestion'

    // Cubic bezier with two control points for a smooth S-curve:
    // CP1 drops straight down from the bubble (vertical departure),
    // CP2 approaches the nebula horizontally (smooth arrival).
    const dy = targetY - sourceY
    const dx = targetX - sourceX
    const curvature = Math.min(Math.abs(dx) * 0.5, dy * 0.4)

    // CP1: straight below the bubble, then curves outward slightly
    const cp1x = sourceX
    const cp1y = sourceY + dy * 0.45

    // CP2: approaches nebula center, arriving more horizontally
    const cp2x = targetX + (sourceX < targetX ? -curvature : curvature)
    const cp2y = targetY - dy * 0.15

    newLines.push({
      d: `M ${sourceX} ${sourceY} C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${targetX} ${targetY}`,
      index: i,
      category,
    })
  }

  lines.value = newLines
}

let resizeObserver: ResizeObserver | null = null

onMounted(async () => {
  await nextTick()
  computeLines()
  if (containerRef.value) {
    resizeObserver = new ResizeObserver(() => computeLines())
    resizeObserver.observe(containerRef.value)
  }
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
})

// Recompute when connections change or nebula mounts
watch(() => props.count, async () => {
  await nextTick()
  computeLines()
})

watch(() => props.nebulaEl, async () => {
  await nextTick()
  computeLines()
})

// Expose recompute for parent to call after layout settles
defineExpose({ recompute: computeLines })
</script>

<template>
  <div ref="containerRef" class="flow-lines-container">
    <svg
      class="flow-lines-svg"
      :width="svgWidth"
      :height="svgHeight"
      :viewBox="`0 0 ${svgWidth} ${svgHeight}`"
    >
      <defs>
        <filter id="flow-glow">
          <feGaussianBlur stdDeviation="2.5" result="blur" />
          <feMerge>
            <feMergeNode in="blur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>

        <!-- Arrowhead pointing forward (end of path) -->
        <marker
          id="arrow-end"
          viewBox="0 0 10 8"
          refX="9"
          refY="4"
          markerWidth="8"
          markerHeight="6"
          orient="auto-start-reverse"
        >
          <path d="M 0 0 L 10 4 L 0 8 z" fill="var(--accent)" fill-opacity="0.4" />
        </marker>

        <!-- Arrowhead pointing backward (start of path) -->
        <marker
          id="arrow-start"
          viewBox="0 0 10 8"
          refX="1"
          refY="4"
          markerWidth="8"
          markerHeight="6"
          orient="auto-start-reverse"
        >
          <path d="M 10 0 L 0 4 L 10 8 z" fill="var(--accent)" fill-opacity="0.4" />
        </marker>
      </defs>

      <!-- Faint static base curve with arrowheads -->
      <path
        v-for="line in lines"
        :key="'base-' + line.index"
        :d="line.d"
        fill="none"
        stroke="var(--accent)"
        :stroke-opacity="line.category === 'enrichment' ? '0.06' : '0.08'"
        stroke-width="1"
        :stroke-dasharray="line.category === 'enrichment' ? '2 6' : 'none'"
        :marker-end="line.category !== 'outbound' ? 'url(#arrow-end)' : undefined"
        :marker-start="line.category === 'outbound' || line.category === 'enrichment' ? 'url(#arrow-start)' : undefined"
      />

      <!-- Animated flowing dashes -->
      <path
        v-for="line in lines"
        :key="'flow-' + line.index"
        :d="line.d"
        fill="none"
        stroke="var(--accent)"
        :stroke-opacity="line.category === 'outbound' ? '0.25' : '0.3'"
        stroke-width="1"
        :stroke-dasharray="line.category === 'enrichment' ? '2 6' : '4 8'"
        stroke-linecap="round"
        filter="url(#flow-glow)"
        :class="[
          'flow-line',
          line.category === 'outbound' ? 'flow-line--outbound' : '',
          line.category === 'enrichment' ? 'flow-line--enrichment' : '',
        ]"
        :style="{ '--flow-index': line.index }"
      />
    </svg>
  </div>
</template>

<style scoped>
.flow-lines-container {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 1;
}

.flow-lines-svg {
  position: absolute;
  inset: 0;
}

.flow-line {
  animation: flow-down 2.5s linear infinite;
  animation-delay: calc(var(--flow-index, 0) * 300ms);
}

/* Outbound: dashes flow upward (from nebula to bubble) */
.flow-line--outbound {
  animation-name: flow-up;
}

/* Enrichment: dashes pulse back and forth */
.flow-line--enrichment {
  animation-name: flow-bidir;
  animation-duration: 3s;
  animation-timing-function: ease-in-out;
}

@keyframes flow-down {
  from { stroke-dashoffset: 24; }
  to { stroke-dashoffset: 0; }
}

@keyframes flow-up {
  from { stroke-dashoffset: 0; }
  to { stroke-dashoffset: 24; }
}

@keyframes flow-bidir {
  0%, 100% { stroke-dashoffset: 0; }
  50% { stroke-dashoffset: 16; }
}

/* Hidden on mobile — spatial metaphor doesn't work in stacked layout */
@media (max-width: 767px) {
  .flow-lines-container {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .flow-line,
  .flow-line--outbound,
  .flow-line--enrichment {
    animation: none;
    stroke-dashoffset: 0;
  }
}
</style>
