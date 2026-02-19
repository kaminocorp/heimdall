<script setup lang="ts">
import type { Report } from '@/types/report'

defineProps<{
  report: Report
}>()
</script>

<template>
  <div class="space-y-4">
    <h2 class="text-xl font-semibold">{{ report.summary }}</h2>
    <div class="grid grid-cols-2 gap-4 text-sm">
      <div><span class="text-gray-500">Severity:</span> {{ report.severity }}</div>
      <div><span class="text-gray-500">Status:</span> {{ report.status }}</div>
      <div><span class="text-gray-500">Trigger:</span> {{ report.trigger_type }}</div>
      <div><span class="text-gray-500">Started:</span> {{ report.started_at }}</div>
    </div>
  </div>
</template>
