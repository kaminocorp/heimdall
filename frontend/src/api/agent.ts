import client from './client'
import type { AgentConfig } from '@/types/agent'

export function getAgentConfig() {
  return client.get<AgentConfig>('/agent/config')
}

export function updateAgentConfig(data: Partial<AgentConfig>) {
  return client.put<AgentConfig>('/agent/config', data)
}
