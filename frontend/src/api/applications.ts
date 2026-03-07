import client from './client'
import type { Application, AppAgentConfig, MonitoringStatus } from '@/types/organization'
import type { Connection } from '@/types/connection'

export async function listApplications(): Promise<Application[]> {
  const { data } = await client.get<Application[]>('/apps')
  return data
}

export async function createApplication(name: string): Promise<Application> {
  const { data } = await client.post<Application>('/apps', { name })
  return data
}

export async function getApplication(appId: string): Promise<Application> {
  const { data } = await client.get<Application>(`/apps/${appId}`)
  return data
}

export async function getAppAgentConfig(appId: string): Promise<AppAgentConfig> {
  const { data } = await client.get<AppAgentConfig>(`/apps/${appId}/agent/config`)
  return data
}

export async function updateAppAgentConfig(appId: string, config: Partial<AppAgentConfig>): Promise<AppAgentConfig> {
  const { data } = await client.put<AppAgentConfig>(`/apps/${appId}/agent/config`, config)
  return data
}

export async function getMonitoringStatus(appId: string): Promise<MonitoringStatus> {
  const { data } = await client.get<MonitoringStatus>(`/apps/${appId}/monitoring/status`)
  return data
}

export async function getAppStats(appId: string): Promise<{ log_count_24h: number; connection_count: number; active_connections: number }> {
  const { data } = await client.get(`/apps/${appId}/stats`)
  return data
}

export async function listConnectionsByApp(appId: string): Promise<Connection[]> {
  const { data } = await client.get<Connection[]>(`/apps/${appId}/connections`)
  return data
}
