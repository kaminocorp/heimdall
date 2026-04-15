import client from './client'
import type { LogEntry } from '@/types/log'

export interface PaginatedLogs {
  data: LogEntry[]
  total: number
  limit: number
  offset: number
}

export async function listLogs(params?: {
  app_id?: string
  severity?: string
  connection_id?: string
  source?: string
  limit?: number
  offset?: number
}): Promise<PaginatedLogs> {
  const { data } = await client.get<PaginatedLogs>('/logs', { params })
  return data
}
