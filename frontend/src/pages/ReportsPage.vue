<script setup lang="ts">
import { onMounted } from 'vue'
import { useReportsStore } from '@/stores/reports'
import ReportList from '@/components/reports/ReportList.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'

const store = useReportsStore()

onMounted(() => {
  store.fetchReports()
})
</script>

<template>
  <div>
    <h2 class="text-2xl font-semibold mb-4">Reports</h2>
    <LoadingSpinner v-if="store.loading" />
    <ReportList v-else :reports="store.reports" />
  </div>
</template>
