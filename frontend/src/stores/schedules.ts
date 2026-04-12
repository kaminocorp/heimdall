import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { InvestigationSchedule, ScheduleInput } from '@/types/schedule'
import * as schedulesApi from '@/api/schedules'
import { extractApiError } from '@/utils/apiError'

// Pinia setup-style store for investigation schedules.
//
// Every action takes appId as its first argument — we don't cache a
// "current" app in this store because the source of truth lives in
// useAppStore(), and threading appId through explicitly makes the
// app-switch lifecycle in SchedulesPage easier to reason about (no
// hidden stale state when the user switches apps mid-session).
export const useSchedulesStore = defineStore('schedules', () => {
  const schedules = ref<InvestigationSchedule[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  // runningId tracks which schedule is mid-"Run now" call so the card can
  // disable its button individually (the scheduler also rate-limits, but
  // the UI protection is faster feedback).
  const runningId = ref<string | null>(null)

  async function fetchSchedules(appId: string) {
    loading.value = true
    error.value = null
    try {
      const { data } = await schedulesApi.listSchedules(appId)
      schedules.value = data
    } catch (e: unknown) {
      error.value = extractApiError(e, 'Failed to load schedules')
    } finally {
      loading.value = false
    }
  }

  async function createSchedule(appId: string, input: ScheduleInput) {
    const { data } = await schedulesApi.createSchedule(appId, input)
    // The backend returns rows in created_at DESC from the list endpoint,
    // so newest-first; unshift keeps the local view consistent without a refetch.
    schedules.value.unshift(data)
    return data
  }

  async function updateSchedule(appId: string, id: string, input: ScheduleInput) {
    const { data } = await schedulesApi.updateSchedule(appId, id, input)
    const idx = schedules.value.findIndex((s) => s.id === id)
    if (idx !== -1) schedules.value[idx] = data
    return data
  }

  async function deleteSchedule(appId: string, id: string) {
    await schedulesApi.deleteSchedule(appId, id)
    schedules.value = schedules.value.filter((s) => s.id !== id)
  }

  async function runNow(appId: string, id: string) {
    runningId.value = id
    try {
      const { data } = await schedulesApi.runScheduleNow(appId, id)
      const idx = schedules.value.findIndex((s) => s.id === id)
      if (idx !== -1) schedules.value[idx] = data
      return data
    } finally {
      runningId.value = null
    }
  }

  return {
    schedules,
    loading,
    error,
    runningId,
    fetchSchedules,
    createSchedule,
    updateSchedule,
    deleteSchedule,
    runNow,
  }
})
