export interface Connection {
  id: string
  name: string
  type: string
  direction: 'one_way' | 'two_way'
  config: Record<string, unknown>
  status: 'active' | 'inactive' | 'error' | 'paused'
  last_seen: string | null
  created_at: string
  updated_at: string
  user_id: string
  // Nullable for org-scoped connections (Phase 2). UI distinguishes the two
  // modes via `app_id === null` — never via a separate flag.
  app_id: string | null
  org_id: string
}

// `app_id` is omitted to create an org-scoped connection (shared across
// every app in the current org). Only connection types that support
// org-scoping accept this — see backend `supportsOrgScope` (webhook_logs,
// otlp). App-scoped creates are identical to Phase 1.
export interface CreateConnectionPayload {
  app_id?: string
  name: string
  type: string
  direction: 'one_way' | 'two_way'
  config: Record<string, unknown>
}

export interface UpdateConnectionPayload {
  name: string
  type: string
  direction: 'one_way' | 'two_way'
  config: Record<string, unknown>
  status: 'active' | 'inactive' | 'error' | 'paused'
}
