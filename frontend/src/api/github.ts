import client from './client'

// Repo-listing endpoints were retired in Phase 3 when github_repos migrated
// to the generic (connection_sources, app_source_filters) pair. GitHub repo
// selection now goes through api/sources.ts with the `discoverable` flag.

export async function getGitHubInstallURL(appId: string): Promise<{ url: string }> {
  const { data } = await client.get<{ url: string }>(`/github/install`, { params: { app_id: appId } })
  return data
}
