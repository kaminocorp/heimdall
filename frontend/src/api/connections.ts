import client from './client'
import type { Connection } from '@/types/connection'

export function listConnections() {
  return client.get<Connection[]>('/connections')
}

export function getConnection(id: string) {
  return client.get<Connection>(`/connections/${id}`)
}

export function createConnection(data: Partial<Connection>) {
  return client.post<Connection>('/connections', data)
}

export function updateConnection(id: string, data: Partial<Connection>) {
  return client.put<Connection>(`/connections/${id}`, data)
}

export function deleteConnection(id: string) {
  return client.delete(`/connections/${id}`)
}
