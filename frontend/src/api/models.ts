import client from './client'
import type { ModelOption } from '@/types/models'

export async function getAvailableModels(): Promise<ModelOption[]> {
  const { data } = await client.get<ModelOption[]>('/models')
  return data
}
