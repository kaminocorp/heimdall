<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { Report } from '@/types/report'
import * as reportsApi from '@/api/reports'
import ReportList from '@/components/reports/ReportList.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const reports = ref<Report[]>([])
const loading = ref(false)

onMounted(async () => {
  loading.value = true
  try {
    const { data } = await reportsApi.listReports()
    reports.value = data
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <h2 class="text-2xl font-semibold mb-4">Reports</h2>
    <LoadingSpinner v-if="loading" />
    <ReportList v-else :reports="reports" />
  </div>
</template>
