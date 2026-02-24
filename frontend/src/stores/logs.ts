import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { LogEntry } from '@/types/log'
import * as logsApi from '@/api/logs'

export const useLogsStore = defineStore('logs', () => {
  const entries = ref<LogEntry[]>([])
  const total = ref(0)
  const limit = ref(50)
  const offset = ref(0)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const source = ref<'raw' | 'agent' | 'all'>('all')

  async function fetchLogs(params?: { severity?: string; connection_id?: string }) {
    loading.value = true
    error.value = null
    try {
      const { data } = await logsApi.listLogs({
        ...params,
        source: source.value,
        limit: limit.value,
        offset: offset.value,
      })
      entries.value = data.data
      total.value = data.total
    } catch (e: any) {
      error.value = e.response?.data?.error ?? 'Failed to load logs'
    } finally {
      loading.value = false
    }
  }

  function nextPage(filters?: { severity?: string; connection_id?: string }) {
    if (offset.value + limit.value < total.value) {
      offset.value += limit.value
      fetchLogs(filters)
    }
  }

  function prevPage(filters?: { severity?: string; connection_id?: string }) {
    if (offset.value > 0) {
      offset.value = Math.max(0, offset.value - limit.value)
      fetchLogs(filters)
    }
  }

  function resetPagination() {
    offset.value = 0
  }

  function setSource(s: 'raw' | 'agent' | 'all') {
    source.value = s
    offset.value = 0
  }

  return { entries, total, limit, offset, loading, error, source, fetchLogs, nextPage, prevPage, resetPagination, setSource }
})
