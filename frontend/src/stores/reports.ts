import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Report } from '@/types/report'
import * as reportsApi from '@/api/reports'

export const useReportsStore = defineStore('reports', () => {
  const reports = ref<Report[]>([])
  const loading = ref(false)

  async function fetchReports(params?: { status?: string }) {
    loading.value = true
    try {
      const { data } = await reportsApi.listReports(params)
      reports.value = data
    } finally {
      loading.value = false
    }
  }

  return { reports, loading, fetchReports }
})
