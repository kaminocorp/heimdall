-- --------------------------------------------------------------------------
-- connection_sources — auto-discovery ledger
-- --------------------------------------------------------------------------

-- UpsertConnectionSource records that a source name was seen for a
-- connection. Called from ingestion after the payload is parsed; `last_seen_at`
-- refreshes on every call so staleness ticks down while traffic flows.
-- name: UpsertConnectionSource :one
INSERT INTO connection_sources (connection_id, source_name)
VALUES ($1, $2)
ON CONFLICT (connection_id, source_name)
DO UPDATE SET last_seen_at = now()
RETURNING *;

-- ListConnectionSources returns every source ever seen on a connection,
-- ordered by recency. Used to render the discovery half of the source
-- selector UI.
-- name: ListConnectionSources :many
SELECT * FROM connection_sources
WHERE connection_id = $1
ORDER BY last_seen_at DESC;

-- --------------------------------------------------------------------------
-- app_source_filters — per-app enable/disable
-- --------------------------------------------------------------------------

-- ListAppSourceFilters returns the filter rows for a given (app, connection)
-- pair. Only these rows are authoritative — missing rows mean "not configured"
-- and are treated as disabled by the ingestion handler (drop-by-default).
-- name: ListAppSourceFilters :many
SELECT * FROM app_source_filters
WHERE app_id = $1 AND connection_id = $2
ORDER BY source_name;

-- UpsertAppSourceFilter toggles a single source on/off for a given app.
-- Used by both bulk PUT (one call per item) and the manual POST add path.
-- name: UpsertAppSourceFilter :one
INSERT INTO app_source_filters (app_id, connection_id, source_name, enabled)
VALUES ($1, $2, $3, $4)
ON CONFLICT (app_id, connection_id, source_name)
DO UPDATE SET enabled = EXCLUDED.enabled
RETURNING *;

-- DeleteAppSourceFilter removes a filter row entirely. Distinct from toggling
-- enabled=false: this is only used for manually-added sources that the user
-- wants to forget. Auto-discovered sources should be toggled off, not deleted.
-- name: DeleteAppSourceFilter :exec
DELETE FROM app_source_filters
WHERE app_id = $1 AND connection_id = $2 AND source_name = $3;

-- ListEnabledSourceNames is the ingestion-hot-path query. Called once per
-- webhook batch to load the set of enabled source names for a connection's
-- app. Backed by idx_app_source_filters_lookup (partial index on
-- enabled = true) so the planner never touches disabled rows.
-- name: ListEnabledSourceNames :many
SELECT source_name FROM app_source_filters
WHERE connection_id = $1 AND app_id = $2 AND enabled = true;

-- CountAppSourceFilters counts all filter rows for an (app, connection)
-- pair. Used by the ingestion handler to distinguish "no filters exist yet"
-- (drop-by-default) from "filters exist but none enabled" (also drops, but
-- the UI can surface a different message).
-- name: CountAppSourceFilters :one
SELECT count(*) FROM app_source_filters
WHERE app_id = $1 AND connection_id = $2;

-- --------------------------------------------------------------------------
-- Connector-specific projections over the generic tables
-- --------------------------------------------------------------------------

-- ListEnabledGitHubReposByApp replaces the pre-Phase-3 query of the same
-- name that read from the bespoke github_repos table. GitHub repos are now
-- just rows in app_source_filters whose parent connection has type='github'
-- — the agent's search_codebase tool joins through here to resolve repo
-- full names plus the connection's installation config.
--
-- Returns the connection id (so callers can group by installation), the
-- repo full name, and the connection's config (for installation_id lookup).
-- name: ListEnabledGitHubReposByApp :many
SELECT
    c.id               AS connection_id,
    asf.source_name    AS repo_full_name,
    c.config           AS connection_config
FROM app_source_filters asf
JOIN connections c ON c.id = asf.connection_id
WHERE asf.app_id = $1
  AND asf.enabled = true
  AND c.type = 'github'
  AND c.status = 'active'
ORDER BY asf.source_name;
