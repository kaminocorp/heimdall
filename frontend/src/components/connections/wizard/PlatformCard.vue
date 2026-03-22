<script setup lang="ts">
import type { PlatformFlow } from './flows'

defineProps<{
  flow: PlatformFlow
}>()

const emit = defineEmits<{
  select: [flowId: string]
}>()
</script>

<template>
  <button
    type="button"
    :disabled="!flow.available"
    @click="flow.available && emit('select', flow.id)"
    class="relative text-left w-full border rounded-lg p-4 transition-all duration-150"
    :class="flow.available
      ? 'border-border bg-bg-surface hover:border-accent/50 hover:bg-bg-surface-hover cursor-pointer'
      : 'border-border/50 bg-bg-surface/50 opacity-40 cursor-not-allowed'"
  >
    <!-- Coming soon badge -->
    <span
      v-if="!flow.available"
      class="absolute top-2 right-2 font-mono text-[9px] uppercase tracking-wider text-text-muted border border-border rounded px-1.5 py-0.5"
    >
      Soon
    </span>

    <div class="flex items-start gap-3">
      <!-- Icon badge -->
      <div
        class="shrink-0 w-9 h-9 rounded flex items-center justify-center font-mono text-xs font-bold uppercase tracking-wider"
        :class="flow.available
          ? 'bg-accent/10 text-accent border border-accent/20'
          : 'bg-border/20 text-text-muted border border-border'"
      >
        {{ flow.icon }}
      </div>

      <div class="min-w-0">
        <h4 class="font-mono text-sm font-medium uppercase tracking-wider text-text-primary">
          {{ flow.name }}
        </h4>
        <p class="font-mono text-xs text-text-muted mt-0.5 leading-relaxed">
          {{ flow.description }}
        </p>
      </div>
    </div>
  </button>
</template>
