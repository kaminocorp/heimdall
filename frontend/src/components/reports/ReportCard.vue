<script setup lang="ts">
import type { Report } from '@/types/report'

defineProps<{
  report: Report
}>()
</script>

<template>
  <div
    class="border border-border rounded-lg p-6 bg-bg-surface hover:border-border-hover transition-colors border-l-2 glow-active"
    :class="{
      'border-l-status-critical': report.severity === 'critical',
      'border-l-status-warn': report.severity === 'warning',
      'border-l-status-info': report.severity === 'info',
    }"
  >
    <div class="flex items-start justify-between gap-3">
      <h3 class="text-sm text-text-primary font-medium">{{ report.summary }}</h3>
      <span
        class="shrink-0 font-mono text-xs font-bold uppercase tracking-wider px-2 py-0.5 rounded border"
        :class="{
          'border-status-critical/30 bg-status-critical/10 text-status-critical': report.severity === 'critical',
          'border-status-warn/30 bg-status-warn/10 text-status-warn': report.severity === 'warning',
          'border-status-info/30 bg-status-info/10 text-status-info': report.severity === 'info',
        }"
      >{{ report.severity }}</span>
    </div>
    <p class="text-xs text-text-muted font-mono mt-2">{{ report.status }}</p>
  </div>
</template>
