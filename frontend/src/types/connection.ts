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
  app_id: string
}

export interface CreateConnectionPayload {
  app_id: string
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
