-- Adds a second partial index on app_source_filters to cover the
-- (connection_id, app_id) query shape used by the ingestion hot-path
-- query `ListEnabledSourceNames`.
--
-- Migration 034 introduced one partial index:
--   idx_app_source_filters_lookup (connection_id, source_name) WHERE enabled = true
-- which serves `ListAppsEnabledForSource` (filters by connection_id + source_name)
-- perfectly — used for org-scoped fan-out, one lookup per distinct source name.
--
-- But `ListEnabledSourceNames` filters by (connection_id, app_id), and the
-- existing index doesn't include app_id. Postgres can scan by connection_id
-- and then heap-filter by app_id, which is fine at small scale but linear in
-- the number of enabled filters for the connection — wasteful as org-scoped
-- connections start accumulating apps.
--
-- Keeping both indexes: each is small (partial, enabled=true only), and each
-- targets one of the two query shapes with no heap re-check.
CREATE INDEX IF NOT EXISTS idx_app_source_filters_app_lookup
    ON app_source_filters(connection_id, app_id)
    WHERE enabled = true;
