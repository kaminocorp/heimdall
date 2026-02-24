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
  id: string
  role: 'user' | 'agent' | 'assistant'
  content: string
  tool_calls?: string[]
  timestamp: string
}

export interface ConversationSummary {
  id: string
  title: string | null
  created_at: string
  updated_at: string
}

export interface Conversation extends ConversationSummary {
  messages: ChatMessage[]
}

export interface WSSystemMessage {
  type: 'system'
  conversation_id: string
}

export interface WSStatusMessage {
  type: 'status'
  content: string
}

export interface WSErrorMessage {
  type: 'error'
  content: string
}

export type WSMessage = WSSystemMessage | WSStatusMessage | WSErrorMessage | ChatMessage
