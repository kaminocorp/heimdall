<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  label: string
  count: number
  rate: number       // events/sec rolling
  tone?: 'default' | 'warn' | 'safe'
  active?: boolean
  expanded?: boolean
}>()

defineEmits<{
  toggle: []
}>()

const tone = computed(() => props.tone ?? 'default')

const accentClass = computed(() => {
  switch (tone.value) {
    case 'warn': return 'text-status-warn border-status-warn/30'
    case 'safe': return 'text-text-secondary border-border'
    default: return 'text-accent-bright border-accent-border'
  }
})

const formattedRate = computed(() => {
  if (props.rate === 0) return '0 / s'
  if (props.rate < 1) return props.rate.toFixed(2) + ' / s'
  return props.rate.toFixed(1) + ' / s'
})
</script>

<template>
  <button
    type="button"
    class="group flex flex-col items-stretch px-3 py-2 rounded border bg-bg-surface backdrop-blur-sm transition-colors text-left min-w-[110px]"
    :class="[
      accentClass,
      expanded ? 'border-accent bg-bg-surface-hover' : 'hover:border-border-hover',
    ]"
    @click="$emit('toggle')"
  >
    <div class="flex items-center justify-between gap-3">
      <span class="font-mono text-xs uppercase tracking-[0.18em]">{{ label }}</span>
      <span
        v-if="active"
        class="w-1.5 h-1.5 rounded-full"
        :class="tone === 'warn' ? 'bg-status-warn' : 'bg-accent-bright'"
        style="box-shadow: 0 0 6px currentColor;"
      />
    </div>
    <div class="font-mono text-xl font-semibold text-text-primary mt-1">
      {{ count.toLocaleString() }}
    </div>
    <div class="font-mono text-[10px] text-text-muted">
      {{ formattedRate }}
    </div>
  </button>
</template>
