<script setup lang="ts">
import type { Report } from '@/types/report'

defineProps<{
  report: Report
}>()
</script>

<template>
  <div class="space-y-4">
    <h2 class="font-mono text-lg font-bold uppercase tracking-wider text-text-primary">{{ report.summary }}</h2>
    <div class="grid grid-cols-2 gap-4 text-sm">
      <div><span class="font-mono text-xs uppercase tracking-wider text-text-muted">Severity:</span> <span class="text-text-secondary">{{ report.severity }}</span></div>
      <div><span class="font-mono text-xs uppercase tracking-wider text-text-muted">Status:</span> <span class="text-text-secondary">{{ report.status }}</span></div>
      <div><span class="font-mono text-xs uppercase tracking-wider text-text-muted">Trigger:</span> <span class="text-text-secondary">{{ report.trigger_type }}</span></div>
      <div><span class="font-mono text-xs uppercase tracking-wider text-text-muted">Started:</span> <span class="text-text-secondary">{{ report.started_at }}</span></div>
    </div>
  </div>
</template>
