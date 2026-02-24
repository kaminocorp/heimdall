import client from './client'
import type { ConversationSummary, Conversation } from '@/types/agent'

export function listConversations() {
  return client.get<ConversationSummary[]>('/conversations')
}

export function getConversation(id: string) {
  return client.get<Conversation>(`/conversations/${id}`)
}
