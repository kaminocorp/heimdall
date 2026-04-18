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
//
// Accepts undefined in addition to null — TypeScript's structural typing
// means partial-response shapes, test mocks, or future callers might pass
// undefined. Without this widening, `new Date(undefined).getTime()` would
// yield NaN, all comparisons against NaN are false, and the function would
// silently return 'stale' (red) for every such case — misleading.
export function classifyStaleness(
  lastSeenAt: string | Date | null | undefined,
): StalenessTier {
  if (lastSeenAt == null) return 'never'
  const timestamp =
    lastSeenAt instanceof Date ? lastSeenAt.getTime() : new Date(lastSeenAt).getTime()
  if (!Number.isFinite(timestamp)) return 'never'
  const ageMs = Date.now() - timestamp
  const hour = 60 * 60 * 1000
  if (ageMs < hour) return 'active'
  if (ageMs < 24 * hour) return 'quiet'
  return 'stale'
}
