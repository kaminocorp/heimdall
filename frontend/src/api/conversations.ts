import client from './client'
import type { Conversation } from '@/types/agent'

export async function getConversation(id: string): Promise<Conversation> {
  const { data } = await client.get<Conversation>(`/conversations/${id}`)
  return data
}
