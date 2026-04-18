-- Phase 2 of "Source Filtering & Org-Level Connections".
--
-- Makes a single connection feedable to multiple applications within the
-- same organisation. An org-scoped connection has app_id IS NULL; each app
-- then independently picks which sources (via app_source_filters from
-- Phase 1) to accept.
--
-- Three schema moves:
--   1. Add connections.org_id (NOT NULL, backfilled from applications.org_id)
--   2. Make connections.app_id nullable — org-scoped connections leave it NULL
--   3. Replace the single-owner (`user_id = app_current_user_id()`) RLS
--      policies on connections / connection_sources with org-member joins,
--      so all members of the owning org can see and manage the connection.
--
-- Non-moves: connections.user_id stays NOT NULL as the creator-audit field.
-- The RLS policy no longer reads it — org membership is the access gate —
-- but the column is still useful for activity-feed attribution.

-- ── 1. Add org_id to connections ────────────────────────────────────────────

ALTER TABLE connections
    ADD COLUMN org_id UUID REFERENCES organizations(id) ON DELETE CASCADE;

-- Backfill from the parent application. Every existing connection is
-- app-scoped, so applications.org_id is the correct source.
UPDATE connections c
SET org_id = a.org_id
FROM applications a
WHERE c.app_id = a.id;

ALTER TABLE connections ALTER COLUMN org_id SET NOT NULL;

CREATE INDEX idx_connections_org_id ON connections(org_id);

-- ── 2. Allow app_id NULL for org-scoped connections ─────────────────────────

ALTER TABLE connections ALTER COLUMN app_id DROP NOT NULL;

-- ── 3. Replace RLS with org-member access ───────────────────────────────────

-- connections: visible to any member of the owning org (matches the
-- applications_user_policy idiom from migration 026).
DROP POLICY IF EXISTS connections_owner ON connections;

CREATE POLICY connections_org_member ON connections
    FOR ALL
    USING (
        org_id IN (
            SELECT om.org_id FROM org_members om
            WHERE om.user_id = app_current_user_id()
        )
    );

-- connection_sources: inherit access from the parent connection — now via
-- the connection's org rather than a single user.
DROP POLICY IF EXISTS connection_sources_owner ON connection_sources;

CREATE POLICY connection_sources_org_member ON connection_sources
    FOR ALL
    USING (
        connection_id IN (
            SELECT c.id FROM connections c
            JOIN org_members om ON om.org_id = c.org_id
            WHERE om.user_id = app_current_user_id()
        )
    );
