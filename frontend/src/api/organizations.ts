import client from './client'
import type { Organization, OnboardingPayload, OnboardingResponse } from '@/types/organization'

export async function getOrganization(): Promise<Organization> {
  const { data } = await client.get<Organization>('/org')
  return data
}

export async function onboard(payload: OnboardingPayload): Promise<OnboardingResponse> {
  const { data } = await client.post<OnboardingResponse>('/onboard', payload)
  return data
}
