<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { initNebula, type NebulaInstance } from './useAgentNebula'

const props = withDefaults(defineProps<{
  /** Dormant mode dims the nebula when no connections exist */
  dormant?: boolean
}>(), { dormant: false })

const containerRef = ref<HTMLElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)

let nebula: NebulaInstance | null = null

onMounted(() => {
  if (!canvasRef.value || !containerRef.value) return
  nebula = initNebula(
    { container: containerRef.value, canvas: canvasRef.value },
    props.dormant,
  )
})

onBeforeUnmount(() => {
  nebula?.dispose()
  nebula = null
})
</script>

<template>
  <div ref="containerRef" class="agent-nebula">
    <canvas ref="canvasRef" class="block w-full h-full" />
  </div>
</template>

<style scoped>
.agent-nebula {
  position: relative;
  width: 100%;
  height: 560px;
}

@media (prefers-reduced-motion: reduce) {
  .agent-nebula canvas {
    animation: none !important;
  }
}
</style>
