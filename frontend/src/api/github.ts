import client from './client'
import type { GitHubRepo } from '@/types/github'

export async function getGitHubInstallURL(appId: string): Promise<{ url: string }> {
  const { data } = await client.get<{ url: string }>(`/github/install`, { params: { app_id: appId } })
  return data
}

export async function listGitHubRepos(connectionId: string): Promise<GitHubRepo[]> {
  const { data } = await client.get<GitHubRepo[]>(`/connections/${connectionId}/github/repos`)
  return data
}

export async function updateGitHubRepos(connectionId: string, repos: GitHubRepo[]): Promise<void> {
  await client.put(`/connections/${connectionId}/github/repos`, repos)
}
