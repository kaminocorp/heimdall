<script setup lang="ts">
/**
 * FlowLines — SVG overlay rendering animated dashed paths from each
 * connection bubble down to the agent nebula canvas.
 *
 * Paths are quadratic beziers computed from DOM positions via
 * getBoundingClientRect(). A ResizeObserver recomputes on layout changes.
 *
 * The dash animation runs continuously top-to-bottom to suggest
 * ongoing data flow into the agent — not a one-time draw-in.
 */

import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'

const props = defineProps<{
  /** Refs to each bubble element's root DOM node */
  bubbleEls: (HTMLElement | null)[]
  /** Ref to the nebula container element */
  nebulaEl: HTMLElement | null
  /** Total connection count — used to trigger recompute on add/remove */
  count: number
}>()

interface LinePath {
  d: string
  index: number
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

    // Control point: midpoint horizontally biased toward target, vertically between
    const cx = sourceX + (targetX - sourceX) * 0.5
    const cy = sourceY + (targetY - sourceY) * 0.55

    newLines.push({
      d: `M ${sourceX} ${sourceY} Q ${cx} ${cy} ${targetX} ${targetY}`,
      index: i,
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
      </defs>

      <!-- Faint static base line -->
      <path
        v-for="line in lines"
        :key="'base-' + line.index"
        :d="line.d"
        fill="none"
        stroke="var(--accent)"
        stroke-opacity="0.08"
        stroke-width="1"
      />

      <!-- Animated flowing dashes -->
      <path
        v-for="line in lines"
        :key="'flow-' + line.index"
        :d="line.d"
        fill="none"
        stroke="var(--accent)"
        stroke-opacity="0.3"
        stroke-width="1"
        stroke-dasharray="4 8"
        stroke-linecap="round"
        filter="url(#flow-glow)"
        class="flow-line"
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

@keyframes flow-down {
  from {
    stroke-dashoffset: 24;
  }
  to {
    stroke-dashoffset: 0;
  }
}

/* Hidden on mobile — spatial metaphor doesn't work in stacked layout */
@media (max-width: 767px) {
  .flow-lines-container {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .flow-line {
    animation: none;
    stroke-dashoffset: 0;
  }
}
</style>
