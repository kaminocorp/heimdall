-- Migration A of the RLS role split (rls-enforcement-roadmap.md Phase 4).
--
-- Privilege topology for the two runtime roles (app_user, cron_user). Roles
-- themselves are NOT created here — they are bootstrapped manually per
-- environment via the SQL editor / psql, so passwords never live in version
-- control. The bootstrap SQL lives verbatim in
-- docs/completions/rls-enforcement-phase-4.md.
--
-- This migration is purely additive: no REVOKE, no FORCE RLS, no policy
-- changes. Migration B (Phase 7, planned 040_rls_force_enforcement) flips
-- enforcement on. Today both DATABASE_URL and CRON_DATABASE_URL still
-- authenticate as postgres; the env-var flip is Phase 6's job.

-- =========================================================================
-- 1. Bootstrap assertion. The migration refuses to apply unless both roles
--    exist. The error message tells the operator exactly where to find the
--    bootstrap procedure.
-- =========================================================================

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'app_user') THEN
        RAISE EXCEPTION
            'app_user role missing; run the bootstrap from docs/completions/rls-enforcement-phase-4.md before applying migration 039';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'cron_user') THEN
        RAISE EXCEPTION
            'cron_user role missing; run the bootstrap from docs/completions/rls-enforcement-phase-4.md before applying migration 039';
    END IF;
END $$;

-- =========================================================================
-- 2. app_user — the RLS-enforced runtime role (parent plan §4.1).
--    No SUPERUSER, no BYPASSRLS, no CREATE on schema. Full CRUD on
--    public.* (RLS becomes the boundary), with default privileges so future
--    migrations inherit the same shape automatically.
-- =========================================================================

GRANT USAGE ON SCHEMA public TO app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO app_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO app_user;

ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_user;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO app_user;

-- =========================================================================
-- 3. cron_user — the BYPASSRLS enumeration role (parent plan §4.2).
--    Blanket SELECT for cross-tenant lookups; deliberately narrow writes.
--    Every write here is a documented exception with a single justification.
-- =========================================================================

GRANT USAGE ON SCHEMA public TO cron_user;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO cron_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO cron_user;

-- Narrow write grant 1: monitoring_state cursor advancement.
-- Owner of advance: agent/monitor.go's monitorTick → UpsertMonitoringState.
-- Justified in parent plan §4.2 verbatim — infra cursor, not tenant data.
GRANT INSERT, UPDATE ON monitoring_state TO cron_user;

-- Narrow write grant 2 (audit-discovered, Phase 1 / Phase 2 surprise):
-- log_buffer retention pruning. agent/pruner.go runs DELETE FROM log_buffer
-- across all tenants on a 1h schedule. Pure enumeration shape — no per-row
-- decisions, no cross-tenant data exfiltration risk — so it stays inside
-- the cron-pool philosophy rather than fanning out to per-tenant app_user
-- handoffs (which would 30x the connection cost for a janitorial sweep).
-- Captured in rls-enforcement-phase-1.md §5.3 row "Log buffer pruner" and
-- reaffirmed in rls-enforcement-phase-2.md "What Phase 4 needs from this".
GRANT DELETE ON log_buffer TO cron_user;

-- Future-table defaults: SELECT only. Forces every new cron write to be
-- an explicit grant (= an explicit security review).
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public
    GRANT SELECT ON TABLES TO cron_user;
ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO cron_user;

-- =========================================================================
-- 4. PG17 conditional grant block (parent plan §5.6d).
--    The Phase 5 asAppUser test fixture issues SET LOCAL ROLE app_user
--    inside a postgres-authenticated transaction. SET ROLE requires
--    membership; PG17 also requires WITH SET TRUE for the membership to
--    permit role-flipping inside an active session.
--    Production roles connect as app_user / cron_user directly and never
--    issue SET ROLE, so these grants are dormant in prod paths.
-- =========================================================================

DO $$
BEGIN
    IF current_setting('server_version_num')::int >= 170000 THEN
        EXECUTE 'GRANT app_user TO postgres WITH SET TRUE';
        EXECUTE 'GRANT cron_user TO postgres WITH SET TRUE';
    ELSE
        EXECUTE 'GRANT app_user TO postgres';
        EXECUTE 'GRANT cron_user TO postgres';
    END IF;
END $$;

-- =========================================================================
-- 5. SECURITY DEFINER helper retrofit.
--    rls-enforcement-phase-1.md §4.5 confirmed all three user-defined
--    functions are already correctly classified:
--      - handle_new_user        → SECURITY DEFINER (006)
--      - app_user_org_ids       → SECURITY DEFINER (028)
--      - app_current_user_id    → SECURITY INVOKER, but reads no RLS table
--    No retrofit is needed. EXECUTE grants on the two DEFINER helpers are
--    issued here so app_user / cron_user can call them without inheriting
--    PUBLIC's reach (defence-in-depth — PUBLIC already has EXECUTE today).
-- =========================================================================

GRANT EXECUTE ON FUNCTION public.app_current_user_id() TO app_user, cron_user;
GRANT EXECUTE ON FUNCTION public.app_user_org_ids() TO app_user, cron_user;
-- handle_new_user is a trigger function fired by the auth.users insert
-- trigger; runtime roles never EXECUTE it directly, so no grant is added.
