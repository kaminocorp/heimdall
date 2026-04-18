-- Phase 1 of "Source Filtering & Org-Level Connections".
--
-- Two new tables:
--
--   connection_sources  — auto-discovery ledger. Populated by ingestion whenever
--                         a new source_name is seen (e.g. fly.app.name). Tracks
--                         first_seen_at / last_seen_at for staleness detection.
--
--   app_source_filters  — per-app toggle list. Enabled sources are accepted at
--                         ingestion time; everything else is dropped. Default
--                         is `enabled = false` (drop-by-default semantics).
--
-- Connections stay app-scoped in Phase 1 — the app_id in app_source_filters is
-- the owning connection's app_id. Phase 2 will introduce org-scoped connections
-- where a single (connection_id, source_name) pair can be enabled independently
-- across multiple apps; the schema already accommodates that, no change needed.

-- ── connection_sources: discovery ledger ─────────────────────────────────────

CREATE TABLE connection_sources (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id UUID NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    source_name   TEXT NOT NULL,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (connection_id, source_name)
);

CREATE INDEX idx_connection_sources_conn ON connection_sources(connection_id);

ALTER TABLE connection_sources ENABLE ROW LEVEL SECURITY;

-- Visible to any session whose user owns the parent connection. Matches the
-- user_id = app_current_user_id() shape of the existing connections_owner
-- policy. Phase 2 (org-level connections) will generalise this to org
-- membership; until then the connection is always directly user-owned.
CREATE POLICY connection_sources_owner ON connection_sources
  FOR ALL
  USING (
    connection_id IN (
      SELECT id FROM connections WHERE user_id = app_current_user_id()
    )
  );

-- ── app_source_filters: per-app enable/disable ───────────────────────────────

CREATE TABLE app_source_filters (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id        UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    connection_id UUID NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    source_name   TEXT NOT NULL,
    enabled       BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (app_id, connection_id, source_name)
);

-- Index supporting the (app_id, connection_id) list used by the "Manage
-- sources" UI — matches the ON CONFLICT target of UpsertAppSourceFilter.
CREATE INDEX idx_app_source_filters_app_conn
    ON app_source_filters(app_id, connection_id);

-- Partial index optimised for the ingestion-hot-path lookup: "which source
-- names are enabled for (connection_id, app_id)?" Only indexes enabled rows,
-- keeping the index small and the lookup O(matches).
CREATE INDEX idx_app_source_filters_lookup
    ON app_source_filters(connection_id, source_name)
    WHERE enabled = true;

ALTER TABLE app_source_filters ENABLE ROW LEVEL SECURITY;

-- Visible to any member of the app's org — matches the applications_user_policy
-- from migration 026 (post-org-members world).
CREATE POLICY app_source_filters_owner ON app_source_filters
  FOR ALL
  USING (
    app_id IN (
      SELECT a.id FROM applications a
      JOIN org_members om ON om.org_id = a.org_id
      WHERE om.user_id = app_current_user_id()
    )
  );
