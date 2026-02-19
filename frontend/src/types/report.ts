export interface Report {
  id: string
  trigger_type: string
  trigger_source: string | null
  summary: string
  severity: 'info' | 'warning' | 'critical'
  status: 'open' | 'investigating' | 'resolved' | 'dismissed'
  context: Record<string, unknown>
  findings: Record<string, unknown> | null
  tool_trace: Record<string, unknown>[] | null
  resolution: string | null
  started_at: string
  resolved_at: string | null
}
