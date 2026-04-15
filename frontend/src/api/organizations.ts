import client from './client'
import type { Organization, OrganizationWithRole, OrgMember, OnboardingPayload, OnboardingResponse } from '@/types/organization'

export async function getOrganization(): Promise<Organization> {
  const { data } = await client.get<Organization>('/org')
  return data
}

export async function updateOrganization(payload: { name?: string; slug?: string }): Promise<Organization> {
  const { data } = await client.put<Organization>('/org', payload)
  return data
}

export async function listUserOrganizations(): Promise<OrganizationWithRole[]> {
  const { data } = await client.get<OrganizationWithRole[]>('/orgs')
  return data
}

export async function createOrganization(name: string, slug: string): Promise<Organization> {
  const { data } = await client.post<Organization>('/orgs', { name, slug })
  return data
}

export async function listOrgMembers(): Promise<OrgMember[]> {
  const { data } = await client.get<OrgMember[]>('/org/members')
  return data
}

export async function inviteOrgMember(email: string, role: string = 'member'): Promise<OrgMember> {
  const { data } = await client.post<OrgMember>('/org/members/invite', { email, role })
  return data
}

export async function updateMemberRole(userId: string, role: string): Promise<void> {
  await client.put(`/org/members/${userId}/role`, { role })
}

export async function removeOrgMember(userId: string): Promise<void> {
  await client.delete(`/org/members/${userId}`)
}

export async function deleteOrganization(confirmSlug: string): Promise<void> {
  await client.delete('/org', { data: { confirm: confirmSlug } })
}

export async function onboard(payload: OnboardingPayload): Promise<OnboardingResponse> {
  const { data } = await client.post<OnboardingResponse>('/onboard', payload)
  return data
}
