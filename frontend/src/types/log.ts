export interface LogEntry {
  id: string
  source: 'raw' | 'agent'
  timestamp: string
  severity: string | null
  source_type: string
  connection_id: string | null
  summary: string
  detail: Record<string, unknown> | null
}
