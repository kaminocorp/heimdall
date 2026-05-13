-- Migration B of the RLS role split (rls-enforcement-roadmap.md Phase 7).
--
-- Flips RLS from cosmetic to enforced. Two halves:
--
--   1. Policy patches — four tables Phase 1 §5.2 surfaced as either
--      RLS-enabled-without-policies (agent_config, webhook_idempotency,
--      log_pipeline_events) plus a users SELECT widening so the org-member
--      list view can read a sibling member's email under FORCE.
--
--      The original Phase 1 audit also flagged a stale github_repos_owner
--      policy carrying the pre-035 single-user shape, but migration 036
--      retired the github_repos table entirely (replaced by
--      connection_sources + app_source_filters), so that gap closed
--      itself before this migration runs. Phase 10 polish removed the
--      now-dangling references — see rls-enforcement-phase-10.md.
--
--   2. FORCE ROW LEVEL SECURITY — every public.* table with rowsecurity=true.
--      Once FORCE is set, the table owner (postgres in production) is also
--      subject to the policies unless the connection authenticates as a
--      SUPERUSER role. Migration runs through DIRECT_URL=postgres (which IS
--      SUPERUSER), so the migration itself is unaffected; runtime app_user
--      is the role that newly observes enforcement.
--
-- Ordering inside the migration: policies first, FORCE second. If FORCE went
-- first, every existing row read on agent_config/webhook_idempotency/
-- log_pipeline_events between the FORCE statement and its policy would fail
-- mid-transaction. Single transaction means it doesn't matter for atomicity
-- but writing it as policies-then-FORCE keeps the diff readable: "here's
-- what each table can do, then enforce it."
--
-- Reversal: 040.down.sql drops the new policies, restores the old
-- github_repos_owner body, and reverses FORCE on every table. Rehearsed
-- locally before deploy per roadmap Phase 7 acceptance check.
--
-- Lock-wait safety: the §2 ALTER TABLE … FORCE statements take ACCESS
-- EXCLUSIVE on ~22 tables. The ALTER itself is metadata-only (no row
-- rewrite), but a long-running query holding even a weak lock will queue
-- writers behind us for the duration of the wait. SET LOCAL lock_timeout
-- so a stalled deploy aborts the transaction in seconds instead of
-- jamming production traffic. golang-migrate runs each migration in its
-- own transaction, so SET LOCAL is the right scope — it expires with
-- the txn and doesn't bleed into subsequent statements.

SET LOCAL lock_timeout = '5s';

-- =========================================================================
-- 1. Policy patches.
-- =========================================================================

-- ── agent_config — global singleton (legacy, still read as a fallback by
--    agent/loop.go when an app's per-app config is unset). RLS-enable in
--    030 was defence-in-depth against PostgREST, not tenant isolation; the
--    table holds no tenant-scoped data, so the policy needs to permit
--    blanket reads from any runtime role.
--
--    Reads only. The legacy UpsertAgentConfig sqlc query is generated but
--    has no runtime caller (verified pre-deploy: only loop.go:104's
--    GetAgentConfig hits this table). Issuing the singleton's body via
--    app_user under FORCE could let any tenant session globally rewrite
--    default_model / mode / system_prompt_override, so the policy is
--    SELECT-only — writes go through DIRECT_URL (postgres) via migrations
--    or an out-of-band admin path. If a future feature needs runtime
--    writes, the right shape is a SECURITY DEFINER helper (mirroring
--    041's lookup_user_for_invite), not a widening of this policy.

CREATE POLICY agent_config_global ON agent_config
    FOR SELECT
    USING (true);

-- ── webhook_idempotency — written by handlers/webhooks.go (and read by
--    the same path) under app_user post-Phase-2. The dedup row's parent
--    connection is the natural ownership anchor. The inner SELECT runs
--    under RLS too, so it transitively inherits connections_org_member's
--    org-membership filter without re-spelling the join here.

CREATE POLICY webhook_idempotency_via_connection ON webhook_idempotency
    FOR ALL
    USING (
        connection_id IN (SELECT id FROM connections)
    )
    WITH CHECK (
        connection_id IN (SELECT id FROM connections)
    );

-- ── log_pipeline_events — written by agent/pipeline_writer.go from every
--    Write* method (monitor loop and ingestion handlers) and read by the
--    pipeline.go handlers (Time Machine + SSE + per-log replay modal).
--    Scope by app_id; the inner SELECT runs under applications_user_policy
--    so visibility transitively follows the caller's org membership.

CREATE POLICY log_pipeline_events_via_app ON log_pipeline_events
    FOR ALL
    USING (
        app_id IN (SELECT id FROM applications)
    )
    WITH CHECK (
        app_id IN (SELECT id FROM applications)
    );

-- ── users — widening for the org-member LIST view (organizations.sql
--    ListOrgMembers joins users on user_id to surface the email column).
--    The existing users_self policy (013) keeps INSERT/UPDATE/DELETE
--    locked to the caller's own row; this new SELECT-only policy lets
--    members of shared orgs read each other's id/email/created_at so the
--    member-list UI works under FORCE.
--
--    The invite-by-email flow (org_members.go AddOrgMember →
--    GetUserByEmail) is NOT covered by this policy: a target user not yet
--    in any of the caller's orgs is hidden, so GetUserByEmail returns
--    ErrNoRows and the handler reports "user not found — they must have
--    a Heimdall account first." That's a known regression of the FORCE
--    flip; the long-term fix is a SECURITY DEFINER helper exposing a
--    scoped lookup-by-email, tracked in the Phase 7 completion doc as a
--    follow-up rather than bundled here.

CREATE POLICY users_org_visible ON users
    FOR SELECT
    USING (
        id IN (
            SELECT om.user_id FROM org_members om
            WHERE om.org_id IN (SELECT public.app_user_org_ids())
        )
    );

-- =========================================================================
-- 2. FORCE ROW LEVEL SECURITY.
--
-- One ALTER TABLE per rowsecurity=true table from migrations
-- 013/015/020/021/023/027/030/033/034/038. Listed in
-- migration-introduction order so a future drift can be eyeballed against
-- the file tree.
-- =========================================================================

-- 013 — original rls_enable_rls cohort
ALTER TABLE users          FORCE ROW LEVEL SECURITY;
ALTER TABLE connections    FORCE ROW LEVEL SECURITY;
ALTER TABLE conversations  FORCE ROW LEVEL SECURITY;
ALTER TABLE agent_log      FORCE ROW LEVEL SECURITY;
ALTER TABLE log_buffer     FORCE ROW LEVEL SECURITY;
ALTER TABLE investigations FORCE ROW LEVEL SECURITY;

-- 015 — per-app agent config
ALTER TABLE app_agent_config FORCE ROW LEVEL SECURITY;

-- 020 — github repos ledger (table retired by migration 036; intentionally absent)

-- 021 — org/app/monitoring/notifications cohort
ALTER TABLE organizations            FORCE ROW LEVEL SECURITY;
ALTER TABLE applications             FORCE ROW LEVEL SECURITY;
ALTER TABLE monitoring_state         FORCE ROW LEVEL SECURITY;
ALTER TABLE notification_channels    FORCE ROW LEVEL SECURITY;
ALTER TABLE notification_preferences FORCE ROW LEVEL SECURITY;
ALTER TABLE notification_log         FORCE ROW LEVEL SECURITY;

-- 023 — investigation schedules
ALTER TABLE investigation_schedules FORCE ROW LEVEL SECURITY;

-- 027 — org members (the recursion-fix table)
ALTER TABLE org_members FORCE ROW LEVEL SECURITY;

-- 030 — defence-in-depth on system tables
ALTER TABLE agent_config       FORCE ROW LEVEL SECURITY;
ALTER TABLE schema_migrations  FORCE ROW LEVEL SECURITY;

-- 033 — webhook idempotency cache
ALTER TABLE webhook_idempotency FORCE ROW LEVEL SECURITY;

-- 034 — source filtering tables
ALTER TABLE connection_sources  FORCE ROW LEVEL SECURITY;
ALTER TABLE app_source_filters  FORCE ROW LEVEL SECURITY;

-- 038 — pipeline events spine
ALTER TABLE log_pipeline_events FORCE ROW LEVEL SECURITY;
