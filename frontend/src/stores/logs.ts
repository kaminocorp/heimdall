import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { LogEntry } from '@/types/log'
import * as logsApi from '@/api/logs'

export const useLogsStore = defineStore('logs', () => {
  const entries = ref<LogEntry[]>([])
  const loading = ref(false)

  async function fetchLogs(params?: { severity?: string; connection_id?: string }) {
    loading.value = true
    try {
      const { data } = await logsApi.listRecentLogs(params)
      entries.value = data
    } finally {
      loading.value = false
    }
  }

  return { entries, loading, fetchLogs }
})
