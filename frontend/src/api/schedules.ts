import client from './client'
import type { InvestigationSchedule, ScheduleInput } from '@/types/schedule'

// All schedule endpoints are app-scoped: schedules belong to an application,
// and the currentAppId from the app store is embedded in the URL. This
// mirrors the connections/agent-config routing pattern.

export function listSchedules(appId: string) {
  return client.get<InvestigationSchedule[]>(`/apps/${appId}/schedules`)
}

export function createSchedule(appId: string, input: ScheduleInput) {
  return client.post<InvestigationSchedule>(`/apps/${appId}/schedules`, input)
}

export function updateSchedule(appId: string, id: string, input: ScheduleInput) {
  return client.patch<InvestigationSchedule>(`/apps/${appId}/schedules/${id}`, input)
}

export function deleteSchedule(appId: string, id: string) {
  return client.delete(`/apps/${appId}/schedules/${id}`)
}

// runScheduleNow fires a schedule synchronously and returns the updated row
// (with fresh last_run_at / last_status / last_summary). The backend caps
// the run at 5 minutes internally, so the worst-case latency here is bounded
// by the same scheduler timeout — but callers should still show a spinner.
export function runScheduleNow(appId: string, id: string) {
  return client.post<InvestigationSchedule>(`/apps/${appId}/schedules/${id}/run`)
}
