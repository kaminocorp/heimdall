<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import type { WizardState } from '../flows'

const props = defineProps<{ modelValue: WizardState }>()
const emit = defineEmits<{
  'update:modelValue': [state: WizardState]
  valid: [isValid: boolean]
}>()

const mode = ref<'drain' | 'polling'>((props.modelValue.config.flyio_mode as 'drain' | 'polling') ?? 'drain')

watch(() => props.modelValue.config, (cfg) => {
  const newMode = (cfg.flyio_mode as 'drain' | 'polling') ?? 'drain'
  if (newMode !== mode.value) mode.value = newMode
}, { deep: true })

function sync() {
  emit('update:modelValue', {
    ...props.modelValue,
    config: { ...props.modelValue.config, flyio_mode: mode.value },
  })
  emit('valid', true)
}

watch(mode, sync)
onMounted(sync)
</script>

<template>
  <div class="space-y-4">
    <p class="font-mono text-xs text-text-secondary leading-relaxed">
      Choose how Heimdall receives logs from your Fly.io app.
    </p>

    <!-- Drain option -->
    <button
      type="button"
      @click="mode = 'drain'"
      class="w-full text-left border rounded px-4 py-3 transition-colors cursor-pointer"
      :class="mode === 'drain'
        ? 'border-accent bg-accent/5'
        : 'border-border bg-bg-elevated/40 hover:border-border-hover'"
    >
      <div class="flex items-center justify-between mb-1">
        <span class="font-mono text-sm font-medium text-text-primary">Log Drain</span>
        <span class="font-mono text-[10px] uppercase tracking-widest px-2 py-0.5 rounded bg-accent/15 text-accent">Recommended</span>
      </div>
      <p class="font-mono text-xs text-text-tertiary leading-relaxed">
        Deploy the Fly Log Shipper to push logs to Heimdall in near-real-time via webhook. Sub-second latency, NATS-backed.
      </p>
    </button>

    <!-- Polling option -->
    <button
      type="button"
      @click="mode = 'polling'"
      class="w-full text-left border rounded px-4 py-3 transition-colors cursor-pointer"
      :class="mode === 'polling'
        ? 'border-accent bg-accent/5'
        : 'border-border bg-bg-elevated/40 hover:border-border-hover'"
    >
      <div class="flex items-center justify-between mb-1">
        <span class="font-mono text-sm font-medium text-text-primary">API Polling</span>
        <span class="font-mono text-[10px] uppercase tracking-widest px-2 py-0.5 rounded bg-bg-elevated text-text-muted">Zero setup</span>
      </div>
      <p class="font-mono text-xs text-text-tertiary leading-relaxed">
        Heimdall polls the Fly.io Logs API every 30 seconds. No extra infrastructure needed — just your API token.
      </p>
    </button>
  </div>
</template>
