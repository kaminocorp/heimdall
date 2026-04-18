-- Phase 3 of "Source Filtering & Org-Level Connections".
--
-- Retires the bespoke github_repos table in favour of the generic
-- (connection_sources, app_source_filters) pair introduced in Phase 1 and
-- extended to the org-scope world in Phase 2. "Connect once, select what to
-- include" is now one pattern across GitHub, Fly.io, and every future
-- connector type.
--
-- Two-phase data move + schema drop in a single migration:
--   1. For every row in github_repos, insert a connection_sources row
--      (one per (connection_id, repo_full_name)). repo_id and default_branch
--      are discarded — the codebase connector only ever read the full name,
--      and GitHub renames will now manifest as "new source" the same way a
--      renamed Fly app would.
--   2. For every github_repos row, insert an app_source_filters row keyed on
--      (connection.app_id, connection.id, repo_full_name) preserving the
--      enabled bit. GitHub connections are strictly app-scoped today, so we
--      can trust connections.app_id to be non-null.
--   3. DROP TABLE github_repos. ON CONFLICT DO NOTHING on the inserts keeps
--      the migration idempotent if someone runs it twice against a partially
--      migrated database.

INSERT INTO connection_sources (connection_id, source_name, first_seen_at, last_seen_at)
SELECT gr.connection_id, gr.repo_full_name, gr.created_at, gr.created_at
FROM github_repos gr
ON CONFLICT (connection_id, source_name) DO NOTHING;

INSERT INTO app_source_filters (app_id, connection_id, source_name, enabled, created_at)
SELECT c.app_id, gr.connection_id, gr.repo_full_name, gr.enabled, gr.created_at
FROM github_repos gr
JOIN connections c ON c.id = gr.connection_id
WHERE c.app_id IS NOT NULL
ON CONFLICT (app_id, connection_id, source_name) DO NOTHING;

DROP TABLE github_repos;
