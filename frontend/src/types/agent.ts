export interface AgentConfig {
  id: string
  model: string
  mode: 'continuous' | 'scheduled'
  schedule: string | null
  system_prompt_override: string | null
  created_at: string
  updated_at: string
}

export interface ChatMessage {
  role: 'user' | 'agent'
  content: string
  tool_calls?: string[]
  timestamp: string
}
