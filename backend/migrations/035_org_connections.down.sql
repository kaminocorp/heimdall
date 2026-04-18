-- Reverse Phase 2 org-scoping.
--
-- Order of operations matters:
--   1. Drop the new RLS policies and restore the single-owner ones.
--   2. Delete org-scoped connections (app_id IS NULL) before flipping
--      app_id back to NOT NULL, otherwise the ALTER will fail.
--   3. Drop the org_id column and its index.
--
-- This is destructive for org-scoped connections by design — there's no way
-- to reverse them back into app-scoped without picking a target app, which
-- the migration can't do safely.

DROP POLICY IF EXISTS connection_sources_org_member ON connection_sources;

CREATE POLICY connection_sources_owner ON connection_sources
    FOR ALL
    USING (
        connection_id IN (
            SELECT id FROM connections WHERE user_id = app_current_user_id()
        )
    );

DROP POLICY IF EXISTS connections_org_member ON connections;

CREATE POLICY connections_owner ON connections
    FOR ALL
    USING (user_id = app_current_user_id());

DELETE FROM connections WHERE app_id IS NULL;

ALTER TABLE connections ALTER COLUMN app_id SET NOT NULL;

DROP INDEX IF EXISTS idx_connections_org_id;
ALTER TABLE connections DROP COLUMN org_id;
