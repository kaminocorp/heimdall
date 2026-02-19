export interface Connection {
  id: string
  name: string
  type: string
  direction: 'one_way' | 'two_way'
  config: Record<string, unknown>
  status: 'active' | 'inactive' | 'error'
  last_seen: string | null
  created_at: string
  updated_at: string
}
