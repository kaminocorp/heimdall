-- Enable RLS on system tables that are not user-scoped.
--
-- These tables have no CREATE POLICY statements — intentionally.
-- With RLS enabled and no policies, non-owner roles (anon, authenticated)
-- see zero rows. The backend connects as the postgres (owner) role which
-- bypasses RLS, so its access is unaffected.
--
-- This is defence-in-depth: even if PostgREST or a leaked key is used
-- to query these tables, nothing is returned.

-- agent_config: global singleton (legacy fallback for per-app config)
ALTER TABLE agent_config ENABLE ROW LEVEL SECURITY;

-- schema_migrations: golang-migrate bookkeeping
ALTER TABLE schema_migrations ENABLE ROW LEVEL SECURITY;
