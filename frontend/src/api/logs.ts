import client from './client'
import type { LogEntry } from '@/types/log'

export interface PaginatedLogs {
  data: LogEntry[]
  total: number
  limit: number
  offset: number
}

export function listLogs(params?: {
  severity?: string
  connection_id?: string
  limit?: number
  offset?: number
}) {
  return client.get<PaginatedLogs>('/logs', { params })
}
