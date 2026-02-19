import client from './client'
import type { LogEntry } from '@/types/log'

export function listRecentLogs(params?: { severity?: string; connection_id?: string }) {
  return client.get<LogEntry[]>('/logs', { params })
}
