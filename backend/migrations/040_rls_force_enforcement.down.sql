-- Reverse Migration B of the RLS role split (Phase 7).
--
-- Reverses FORCE on every table 040.up touched and drops the four new
-- policies. github_repos was retired by migration 036 before this
-- migration's window, so neither the FORCE flip nor the policy restore
-- references it (see Phase 10 polish notes). Rehearsed locally before
-- the prod deploy per roadmap Phase 7 acceptance check.
--
-- NOTE: this is a clean structural reverse. It does NOT undo the side
-- effects of running under FORCE — e.g. any data written through a
-- writer that was silently failing because of a missing GRANT will not
-- be retroactively recovered. The 48h Phase 7 bake is the watch window
-- for surfacing those.

-- =========================================================================
-- 2. Reverse FORCE (mirror order of 040.up §2).
-- =========================================================================

-- 038
ALTER TABLE log_pipeline_events NO FORCE ROW LEVEL SECURITY;

-- 034
ALTER TABLE app_source_filters  NO FORCE ROW LEVEL SECURITY;
ALTER TABLE connection_sources  NO FORCE ROW LEVEL SECURITY;

-- 033
ALTER TABLE webhook_idempotency NO FORCE ROW LEVEL SECURITY;

-- 030
ALTER TABLE schema_migrations  NO FORCE ROW LEVEL SECURITY;
ALTER TABLE agent_config       NO FORCE ROW LEVEL SECURITY;

-- 027
ALTER TABLE org_members NO FORCE ROW LEVEL SECURITY;

-- 023
ALTER TABLE investigation_schedules NO FORCE ROW LEVEL SECURITY;

-- 021
ALTER TABLE notification_log         NO FORCE ROW LEVEL SECURITY;
ALTER TABLE notification_preferences NO FORCE ROW LEVEL SECURITY;
ALTER TABLE notification_channels    NO FORCE ROW LEVEL SECURITY;
ALTER TABLE monitoring_state         NO FORCE ROW LEVEL SECURITY;
ALTER TABLE applications             NO FORCE ROW LEVEL SECURITY;
ALTER TABLE organizations            NO FORCE ROW LEVEL SECURITY;

-- 020 — github_repos retired by migration 036; nothing to reverse

-- 015
ALTER TABLE app_agent_config NO FORCE ROW LEVEL SECURITY;

-- 013
ALTER TABLE investigations NO FORCE ROW LEVEL SECURITY;
ALTER TABLE log_buffer     NO FORCE ROW LEVEL SECURITY;
ALTER TABLE agent_log      NO FORCE ROW LEVEL SECURITY;
ALTER TABLE conversations  NO FORCE ROW LEVEL SECURITY;
ALTER TABLE connections    NO FORCE ROW LEVEL SECURITY;
ALTER TABLE users          NO FORCE ROW LEVEL SECURITY;

-- =========================================================================
-- 1. Reverse policy patches (mirror order of 040.up §1).
-- =========================================================================

-- users — drop the SELECT widening
DROP POLICY IF EXISTS users_org_visible ON users;

-- github_repos — table retired by migration 036; no policy to restore

-- log_pipeline_events
DROP POLICY IF EXISTS log_pipeline_events_via_app ON log_pipeline_events;

-- webhook_idempotency
DROP POLICY IF EXISTS webhook_idempotency_via_connection ON webhook_idempotency;

-- agent_config
DROP POLICY IF EXISTS agent_config_global ON agent_config;
