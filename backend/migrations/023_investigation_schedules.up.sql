-- Phase 3 of the logs-feed-and-scheduled-investigations plan.
--
-- investigation_schedules: a third agent operating mode. Unlike monitoring
-- (reactive, classifier-driven) or interactive (user-initiated WebSocket chat),
-- scheduled investigations fire on a time trigger with a caller-provided prompt
-- and reuse the RunMonitoring tool-use loop under the hood.
--
-- Naming: we use "investigation_schedules" (not "investigations") because the
-- `investigations` table already exists (migration 003) and stores incident
-- reports. Each row here is "a schedule that produces an investigation".
--
-- Phase 3 MVP uses `interval_secs` for a dumb fixed-interval tick. Phase 4
-- will add `cron_expr`-based scheduling as an alternative — the column exists
-- here (nullable) so Phase 4 can populate it without a schema change. Both
-- columns coexist permanently; the handler enforces exactly-one semantics.

CREATE TABLE investigation_schedules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id          UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    prompt          TEXT NOT NULL,              -- the investigation prompt sent to the agent
    interval_secs   INTEGER NOT NULL,           -- Phase 3: plain integer interval (>=60)
    cron_expr       TEXT,                       -- Phase 4: cron expression (nullable, takes precedence)
    enabled         BOOLEAN NOT NULL DEFAULT true,
    last_run_at     TIMESTAMPTZ,
    last_status     TEXT,                       -- 'success' | 'error' | null
    last_error      TEXT,                       -- error message from last failed run
    last_summary    TEXT,                       -- short blurb for UI display
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Partial indexes: the scheduler only ever queries enabled schedules, so
-- narrowing the index keeps it small and improves the tick-time query.
CREATE INDEX idx_investigation_schedules_app_enabled
    ON investigation_schedules (app_id) WHERE enabled;

CREATE INDEX idx_investigation_schedules_next_run
    ON investigation_schedules (last_run_at) WHERE enabled;

-- RLS: users can access schedules for apps in their org. Mirrors the pattern
-- used by app_agent_config (015), monitoring_state (021), and others.
ALTER TABLE investigation_schedules ENABLE ROW LEVEL SECURITY;

CREATE POLICY investigation_schedules_org ON investigation_schedules
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN public.users u ON u.org_id = a.org_id
            WHERE u.id = app_current_user_id()
        )
    );
