<script setup lang="ts">
import { computed, ref, onMounted, onBeforeUnmount, nextTick } from 'vue'
import type { Connection } from '@/types/connection'
import { flows } from './wizard/flows'
import BlueprintZone from './BlueprintZone.vue'

const props = defineProps<{
  connections: Connection[]
  testingId?: string | null
}>()

const emit = defineEmits<{
  delete: [id: string]
  edit: [connection: Connection]
  test: [id: string]
  'manage-repos': [id: string]
  add: []
}>()

// Build type → category lookup from the single source of truth
const typeToCategory = Object.fromEntries(
  flows.map(f => [f.connectorType, f.category])
) as Record<string, string>

// Icon map: connector type → 2-letter abbreviation
const iconMap = Object.fromEntries(
  flows.map(f => [f.connectorType, f.icon])
) as Record<string, string>

const grouped = computed(() => ({
  log_source: props.connections.filter(c => typeToCategory[c.type] === 'log_source'),
  database: props.connections.filter(c => typeToCategory[c.type] === 'database'),
  generic: props.connections.filter(c => (typeToCategory[c.type] ?? 'generic') === 'generic'),
}))

// — SVG line drawing —
const containerRef = ref<HTMLElement | null>(null)
const hubRef = ref<HTMLElement | null>(null)
const zoneLogRef = ref<HTMLElement | null>(null)
const zoneDbRef = ref<HTMLElement | null>(null)
const zoneIntRef = ref<HTMLElement | null>(null)

interface LineCoords {
  x1: number; y1: number
  x2: number; y2: number
  cx: number; cy: number
}

const lines = ref<LineCoords[]>([])
const svgWidth = ref(0)
const svgHeight = ref(0)

function computeLines() {
  if (!containerRef.value || !hubRef.value) return

  const containerRect = containerRef.value.getBoundingClientRect()
  svgWidth.value = containerRect.width
  svgHeight.value = containerRect.height

  const hubRect = hubRef.value.getBoundingClientRect()
  const hubCx = hubRect.left + hubRect.width / 2 - containerRect.left
  const hubCy = hubRect.top + hubRect.height / 2 - containerRect.top

  const zones = [zoneLogRef.value, zoneDbRef.value, zoneIntRef.value]
  const newLines: LineCoords[] = []

  for (const zone of zones) {
    if (!zone) continue
    const zoneRect = zone.getBoundingClientRect()
    const zoneCx = zoneRect.left + zoneRect.width / 2 - containerRect.left
    const zoneCy = zoneRect.top + zoneRect.height / 2 - containerRect.top

    // Control point for quadratic bezier — offset toward center
    const cx = (hubCx + zoneCx) / 2
    const cy = (hubCy + zoneCy) / 2

    newLines.push({ x1: hubCx, y1: hubCy, x2: zoneCx, y2: zoneCy, cx, cy })
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
})
</script>

<template>
  <div ref="containerRef" class="relative">
    <!-- SVG connection lines (desktop only) -->
    <svg
      class="absolute inset-0 pointer-events-none hidden md:block"
      :width="svgWidth"
      :height="svgHeight"
      :viewBox="`0 0 ${svgWidth} ${svgHeight}`"
    >
      <defs>
        <filter id="line-glow">
          <feGaussianBlur stdDeviation="2" result="blur" />
          <feMerge>
            <feMergeNode in="blur" />
            <feMergeNode in="SourceGraphic" />
          </feMerge>
        </filter>
      </defs>
      <path
        v-for="(line, i) in lines"
        :key="i"
        :d="`M ${line.x1} ${line.y1} Q ${line.cx} ${line.cy} ${line.x2} ${line.y2}`"
        fill="none"
        stroke="var(--accent)"
        stroke-opacity="0.25"
        stroke-width="1"
        stroke-dasharray="6 4"
        filter="url(#line-glow)"
        class="blueprint-line"
        :style="{ '--line-index': i }"
      />
    </svg>

    <!-- Grid layout -->
    <div class="grid grid-cols-1 md:grid-cols-[1fr_auto_1fr] gap-6 md:gap-8 items-start">

      <!-- Left column: Log Sources (row 1) + Integrations (row 2) -->
      <div class="space-y-6 md:row-span-2 order-2 md:order-1">
        <div ref="zoneLogRef">
          <BlueprintZone
            label="Log Sources"
            icon-type="server"
            :connections="grouped.log_source"
            :testing-id="testingId"
            :icon-map="iconMap"
            empty-prompt="+ Add a log source"
            @delete="emit('delete', $event)"
            @edit="emit('edit', $event)"
            @test="emit('test', $event)"
            @manage-repos="emit('manage-repos', $event)"
            @add="emit('add')"
          />
        </div>
        <div ref="zoneIntRef">
          <BlueprintZone
            label="Integrations"
            icon-type="code"
            :connections="grouped.generic"
            :testing-id="testingId"
            :icon-map="iconMap"
            empty-prompt="+ Add an integration"
            @delete="emit('delete', $event)"
            @edit="emit('edit', $event)"
            @test="emit('test', $event)"
            @manage-repos="emit('manage-repos', $event)"
            @add="emit('add')"
          />
        </div>
      </div>

      <!-- Center column: Heimdall hub -->
      <div ref="hubRef" class="flex flex-col items-center justify-center order-1 md:order-2 md:row-span-2 md:self-center">
        <div class="relative flex flex-col items-center">
          <!-- Glow layer -->
          <div class="absolute inset-0 rounded-full bg-accent/10 blur-xl animate-pulse" />

          <!-- Hub circle -->
          <div class="relative w-20 h-20 rounded-full border-2 border-accent/40 bg-bg-elevated flex items-center justify-center shadow-lg">
            <!-- Eye / H monogram -->
            <svg class="w-9 h-9 text-accent" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
              <circle cx="12" cy="12" r="3" />
            </svg>
          </div>

          <!-- Label -->
          <p class="mt-2 font-mono text-xs font-bold uppercase tracking-[0.2em] text-text-muted">
            Heimdall
          </p>
        </div>
      </div>

      <!-- Right column: Databases -->
      <div class="order-3 md:row-span-2">
        <div ref="zoneDbRef">
          <BlueprintZone
            label="Databases"
            icon-type="database"
            :connections="grouped.database"
            :testing-id="testingId"
            :icon-map="iconMap"
            empty-prompt="+ Connect a database"
            @delete="emit('delete', $event)"
            @edit="emit('edit', $event)"
            @test="emit('test', $event)"
            @manage-repos="emit('manage-repos', $event)"
            @add="emit('add')"
          />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.blueprint-line {
  stroke-dashoffset: 100;
  animation: draw-line 0.8s ease-out forwards;
  animation-delay: calc(var(--line-index, 0) * 150ms);
}

@keyframes draw-line {
  to {
    stroke-dashoffset: 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .blueprint-line {
    animation: none;
    stroke-dashoffset: 0;
  }
}
</style>
