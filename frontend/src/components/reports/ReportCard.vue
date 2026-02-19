<script setup lang="ts">
import type { Report } from '@/types/report'

defineProps<{
  report: Report
}>()
</script>

<template>
  <div class="border rounded-lg p-4">
    <div class="flex items-center justify-between">
      <h3 class="font-medium">{{ report.summary }}</h3>
      <span class="text-xs font-medium px-2 py-0.5 rounded" :class="{
        'bg-red-100 text-red-800': report.severity === 'critical',
        'bg-yellow-100 text-yellow-800': report.severity === 'warning',
        'bg-blue-100 text-blue-800': report.severity === 'info',
      }">{{ report.severity }}</span>
    </div>
    <p class="text-sm text-gray-500 mt-1">{{ report.status }}</p>
  </div>
</template>
