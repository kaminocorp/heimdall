import client from './client'
import type { InvestigationSchedule, ScheduleInput } from '@/types/schedule'

// All schedule endpoints are app-scoped: schedules belong to an application,
// and the currentAppId from the app store is embedded in the URL. This
// mirrors the connections/agent-config routing pattern.

export async function listSchedules(appId: string): Promise<InvestigationSchedule[]> {
  const { data } = await client.get<InvestigationSchedule[]>(`/apps/${appId}/schedules`)
  return data
}

export async function createSchedule(appId: string, input: ScheduleInput): Promise<InvestigationSchedule> {
  const { data } = await client.post<InvestigationSchedule>(`/apps/${appId}/schedules`, input)
  return data
}

export async function updateSchedule(appId: string, id: string, input: ScheduleInput): Promise<InvestigationSchedule> {
  const { data } = await client.patch<InvestigationSchedule>(`/apps/${appId}/schedules/${id}`, input)
  return data
}

export async function deleteSchedule(appId: string, id: string): Promise<void> {
  await client.delete(`/apps/${appId}/schedules/${id}`)
}

// runScheduleNow fires a schedule synchronously and returns the updated row
// (with fresh last_run_at / last_status / last_summary). The backend caps
// the run at 5 minutes internally, so the worst-case latency here is bounded
// by the same scheduler timeout — but callers should still show a spinner.
export async function runScheduleNow(appId: string, id: string): Promise<InvestigationSchedule> {
  const { data } = await client.post<InvestigationSchedule>(`/apps/${appId}/schedules/${id}/run`)
  return data
}
