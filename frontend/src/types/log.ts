export interface LogEntry {
  id: string
  connection_id: string
  source_type: string
  severity: string | null
  payload: Record<string, unknown>
  ingested_at: string
}
