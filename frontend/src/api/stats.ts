import client from './client'

export interface DashboardStats {
  log_count_24h: number
  connection_count: number
  active_connections: number
}

export function getDashboardStats() {
  return client.get<DashboardStats>('/stats')
}
