import client from './client'
import type { Report } from '@/types/report'

export function listReports(params?: { status?: string }) {
  return client.get<Report[]>('/reports', { params })
}

export function getReport(id: string) {
  return client.get<Report>(`/reports/${id}`)
}
