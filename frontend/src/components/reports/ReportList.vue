<script setup lang="ts">
import type { Report } from '@/types/report'
import ReportCard from './ReportCard.vue'

defineProps<{
  reports: Report[]
}>()
</script>

<template>
  <div class="space-y-3">
    <ReportCard
      v-for="(report, i) in reports"
      :key="report.id"
      :report="report"
      class="animate-fade-in"
      :style="{ '--stagger-index': i }"
    />
  </div>
</template>
