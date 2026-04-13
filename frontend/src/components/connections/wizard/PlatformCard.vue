<script setup lang="ts">
import type { PlatformFlow } from './flows'
import ConnectorLogo from '@/components/icons/ConnectorLogo.vue'

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
    class="relative flex flex-col items-center text-center w-full border rounded-lg px-3 py-4 transition-all duration-150"
    :class="flow.available
      ? 'border-border bg-bg-surface hover:border-accent/50 hover:bg-bg-surface-hover cursor-pointer'
      : 'border-border/50 bg-bg-surface/50 opacity-40 cursor-default'"
  >
    <!-- Coming soon badge -->
    <span
      v-if="!flow.available"
      class="absolute top-1.5 right-1.5 font-mono text-[8px] uppercase tracking-wider text-text-muted border border-border rounded px-1 py-0.5"
    >
      Soon
    </span>

    <!-- Logo -->
    <ConnectorLogo
      :type="flow.id"
      :size="26"
      class="mb-2 transition-colors duration-150"
      :class="flow.available ? 'text-accent' : 'text-text-muted'"
    />

    <!-- Name -->
    <h4 class="font-mono text-xs font-medium uppercase tracking-wider text-text-primary leading-tight">
      {{ flow.name }}
    </h4>

    <!-- Subtitle -->
    <p class="font-mono text-[10px] text-text-muted mt-0.5 leading-tight">
      {{ flow.subtitle }}
    </p>
  </button>
</template>
