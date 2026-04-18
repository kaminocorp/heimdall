// SourceFilterItem is the merged (discovery, filter) row returned by
// GET /api/connections/{id}/sources.
//
// Server-side this is produced by merging `connection_sources` and
// `app_source_filters`. Timestamps are null for manually-added sources that
// have never received traffic ("never seen" tier in the UI).
export interface SourceFilterItem {
  source_name: string
  enabled: boolean
  first_seen_at: string | null
  last_seen_at: string | null
}

export type StalenessTier = 'active' | 'quiet' | 'stale' | 'never'

// classifyStaleness maps last_seen_at to the four visual tiers used by the
// source selector dots. Thresholds come from the plan (docs/executing/
// source-filtering-and-org-connections.md): 1h / 24h.
export function classifyStaleness(lastSeenAt: string | null): StalenessTier {
  if (!lastSeenAt) return 'never'
  const ageMs = Date.now() - new Date(lastSeenAt).getTime()
  const hour = 60 * 60 * 1000
  if (ageMs < hour) return 'active'
  if (ageMs < 24 * hour) return 'quiet'
  return 'stale'
}
