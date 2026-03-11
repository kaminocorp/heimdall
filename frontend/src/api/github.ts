import client from './client'
import type { GitHubRepo } from '@/types/github'

export function getGitHubInstallURL(appId: string) {
  return client.get<{ url: string }>(`/github/install`, { params: { app_id: appId } }).then(r => r.data)
}

export function listGitHubRepos(connectionId: string) {
  return client.get<GitHubRepo[]>(`/connections/${connectionId}/github/repos`).then(r => r.data)
}

export function updateGitHubRepos(connectionId: string, repos: GitHubRepo[]) {
  return client.put(`/connections/${connectionId}/github/repos`, repos)
}
