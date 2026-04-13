<script setup lang="ts">
defineProps<{
  status: string
}>()
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full font-mono text-xs font-medium uppercase tracking-wider border"
    :class="{
      'border-status-ok/30 bg-status-ok/10 text-status-ok': status === 'active',
      'border-white/10 bg-white/5 text-text-muted': status === 'inactive',
      'border-status-critical/30 bg-status-critical/10 text-status-critical': status === 'error',
      'border-status-warn/30 bg-status-warn/10 text-status-warn': status === 'warning',
    }"
    :style="status === 'active' ? 'box-shadow: 0 0 8px var(--status-ok-glow)'
      : status === 'error' ? 'box-shadow: 0 0 8px var(--status-critical-glow)'
      : status === 'warning' ? 'box-shadow: 0 0 8px var(--status-warn-glow)'
      : ''"
  >
    <span
      class="w-1.5 h-1.5 rounded-full"
      :class="{
        'bg-status-ok glow-pulse': status === 'active',
        'bg-text-muted': status === 'inactive',
        'bg-status-critical glow-pulse-critical': status === 'error',
        'bg-status-warn glow-pulse-warn': status === 'warning',
      }"
    />
    {{ status }}
  </span>
</template>
