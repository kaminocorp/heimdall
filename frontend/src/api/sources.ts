import client from './client'
import type { SourceFilterItem } from '@/types/source'

// SourceRequestOpts — the `appId` is required for org-scoped connections and
// is forwarded to the backend via `?app_id=`. For app-scoped connections the
// backend ignores the param (the connection itself carries the target app).
export interface SourceRequestOpts {
  appId?: string
}

function params(opts?: SourceRequestOpts): Record<string, string> | undefined {
  if (!opts?.appId) return undefined
  return { app_id: opts.appId }
}

export async function listSources(
  connectionId: string,
  opts?: SourceRequestOpts,
): Promise<SourceFilterItem[]> {
  const { data } = await client.get<SourceFilterItem[]>(
    `/connections/${connectionId}/sources`,
    { params: params(opts) },
  )
  return data
}

export async function updateSources(
  connectionId: string,
  sources: { source_name: string; enabled: boolean }[],
  opts?: SourceRequestOpts,
): Promise<void> {
  await client.put(`/connections/${connectionId}/sources`, { sources }, { params: params(opts) })
}

export async function addSource(
  connectionId: string,
  sourceName: string,
  opts?: SourceRequestOpts,
): Promise<void> {
  await client.post(
    `/connections/${connectionId}/sources`,
    { source_name: sourceName },
    { params: params(opts) },
  )
}

export async function deleteSource(
  connectionId: string,
  sourceName: string,
  opts?: SourceRequestOpts,
): Promise<void> {
  // Source name + optional app_id are both query params. Source name in the
  // path would break on embedded slashes (e.g. "vercel/lambda").
  await client.delete(`/connections/${connectionId}/sources`, {
    params: { name: sourceName, ...(params(opts) ?? {}) },
  })
}

// discoverSources triggers the connector-specific discovery hook that
// populates `connection_sources` from upstream (today: GitHub's API). Only
// meaningful for connector types with an explicit list-what-we-see-from-here
// API — webhook ingestion types discover sources passively.
//
// Accepts the same optional `{ appId }` as the CRUD functions. For GitHub
// the backend ignores it today (discovery is per-connection, not per-app),
// but future connector types that gain discovery (e.g. a Vercel projects
// endpoint) may need it for authz resolution — forwarding it here keeps
// the call-site shape consistent and future-proofs the selector.
//
// Returns the count of upstream entries encountered; the caller should
// refresh its source list via `listSources` afterwards to pick up the
// newly inserted rows.
export async function discoverSources(
  connectionId: string,
  opts?: SourceRequestOpts,
): Promise<{ discovered: number; connection_type: string }> {
  const { data } = await client.post<{ discovered: number; connection_type: string }>(
    `/connections/${connectionId}/sources/discover`,
    undefined,
    { params: params(opts) },
  )
  return data
}

