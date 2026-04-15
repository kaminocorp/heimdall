import client from './client'
import type { Connection, CreateConnectionPayload, UpdateConnectionPayload } from '@/types/connection'

export async function listConnections(): Promise<Connection[]> {
  const { data } = await client.get<Connection[]>('/connections')
  return data
}

export async function createConnection(data: CreateConnectionPayload): Promise<Connection> {
  const { data: result } = await client.post<Connection>('/connections', data)
  return result
}

export async function updateConnection(id: string, payload: UpdateConnectionPayload): Promise<Connection> {
  const { data } = await client.put<Connection>(`/connections/${id}`, payload)
  return data
}

export async function deleteConnection(id: string): Promise<void> {
  await client.delete(`/connections/${id}`)
}

export async function testConnection(id: string): Promise<{ success: boolean; message: string }> {
  const { data } = await client.post<{ success: boolean; message: string }>(`/connections/${id}/test`)
  return data
}
