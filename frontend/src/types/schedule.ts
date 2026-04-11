// Matches the Go InvestigationSchedule struct serialized by the backend
// handlers at /api/apps/{appId}/schedules.
//
// Exactly one of interval_secs (>0) or cron_expr (non-empty) is in effect
// at any time. Both fields exist because Phase 3 shipped with interval-only
// schedules, and Phase 4 added cron_expr as the preferred mode without
// removing the legacy path. The backend enforces exactly-one-of at the
// validation boundary; this type models the possibility of either.
export interface InvestigationSchedule {
  id: string
  app_id: string
  name: string
  prompt: string
  // 0 when cron_expr is set, otherwise the active interval in seconds.
  interval_secs: number
  // null when interval_secs is set, otherwise a 5-field cron expression.
  cron_expr: string | null
  enabled: boolean
  last_run_at: string | null
  last_status: 'success' | 'error' | null
  last_error: string | null
  last_summary: string | null
  created_at: string
  updated_at: string
}

// Request shape for POST /schedules and PATCH /schedules/{id}. The backend
// accepts either interval_secs (>0) or cron_expr (non-empty) but not both,
// so the UI always sends exactly one via the modal's mode toggle. enabled
// is optional on create (defaults to true).
export interface ScheduleInput {
  name: string
  prompt: string
  interval_secs?: number
  cron_expr?: string
  enabled?: boolean
}
