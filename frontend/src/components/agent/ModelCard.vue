<script setup lang="ts">
import type { ModelOption } from '@/types/models'

defineProps<{
  model: ModelOption
  selected: boolean
  focused: boolean
}>()

defineEmits<{
  select: [id: string]
}>()

function formatContext(tokens: number): string {
  if (tokens >= 1_000_000) return `${tokens / 1_000_000}M`
  return `${Math.round(tokens / 1_000)}k`
}

const tierColors: Record<string, string> = {
  flagship: 'border-status-warn/40 text-status-warn bg-status-warn/10',
  balanced: 'border-accent-bright/40 text-accent-bright bg-accent-bright/10',
  economy: 'border-status-ok/40 text-status-ok bg-status-ok/10',
  specialist: 'border-status-info/40 text-status-info bg-status-info/10',
}
</script>

<template>
  <button
    type="button"
    @click="$emit('select', model.id)"
    class="w-full text-left px-3 py-2.5 transition-colors cursor-pointer"
    :class="[
      selected
        ? 'bg-accent-subtle'
        : focused
          ? 'bg-bg-surface-hover'
          : 'hover:bg-bg-surface',
    ]"
    role="option"
    :aria-selected="selected"
  >
    <!-- Line 1: Name + tier badge + provider badge -->
    <div class="flex items-center gap-2 min-w-0">
      <span
        class="font-mono text-sm truncate"
        :class="selected ? 'text-accent-bright' : 'text-text-primary'"
      >{{ model.name }}</span>
      <span
        class="flex-shrink-0 font-mono text-[10px] uppercase tracking-wider px-1.5 py-0.5 rounded-full border"
        :class="tierColors[model.tier] || 'border-border text-text-muted'"
      >{{ model.tier }}</span>
      <span class="flex-shrink-0 font-mono text-[10px] text-text-muted">
        {{ model.provider === 'anthropic' ? 'direct' : 'via OR' }}
      </span>
    </div>

    <!-- Line 2: Description -->
    <div class="font-sans text-xs text-text-secondary mt-0.5 leading-relaxed">
      {{ model.description }}
    </div>

    <!-- Line 3: Context · pricing · strength tags -->
    <div class="flex items-center gap-2 mt-1 flex-wrap">
      <span class="font-mono text-[11px] text-text-muted">
        {{ formatContext(model.context_length) }} ctx
      </span>
      <span class="font-mono text-[11px] text-text-muted">·</span>
      <span class="font-mono text-[11px] text-text-muted">
        ${{ model.pricing.prompt }}/${{ model.pricing.completion }}
      </span>
      <span
        v-for="tag in model.strengths"
        :key="tag"
        class="font-mono text-[10px] text-text-muted/80 bg-bg-elevated rounded px-1.5 py-0.5"
      >{{ tag }}</span>
    </div>
  </button>
</template>
