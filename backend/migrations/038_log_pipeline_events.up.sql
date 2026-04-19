-- Per-log stage events for the Pipeline page.
--
-- Every log flowing through the pipeline (ingestion → classify → gate →
-- assessment) writes one row per stage here. The table is the replay spine
-- for the Time Machine view and the persisted twin of the in-memory
-- PipelineBus used by the live SSE stream.
--
-- Retention is bounded by ON DELETE CASCADE on log_id: when PruneExpiredLogs
-- (agent/pruner.go) sheds a log_buffer row past the 48h window, every one of
-- its pipeline events goes with it. No separate pruner.
--
-- RLS follows the system-table pattern from migrations 030 and 033 — enabled
-- with no policies. Non-owner roles (anon, authenticated) see zero rows; the
-- backend connects as the postgres owner role which bypasses RLS. Only the
-- owner-role pool (via s.Queries) writes/reads this table.

CREATE TABLE log_pipeline_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    log_id          UUID NOT NULL REFERENCES log_buffer(id) ON DELETE CASCADE,
    app_id          UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    stage           TEXT NOT NULL,      -- 'ingestion' | 'classified' | 'gate' | 'assessment'
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- Stage-specific columns (nullable; filled per stage):
    source_type     TEXT,               -- ingestion
    severity        TEXT,               -- ingestion
    type            TEXT,               -- classified (Lumber Type)
    category        TEXT,               -- classified (Lumber Category)
    confidence      DOUBLE PRECISION,   -- classified
    summary         TEXT,               -- classified (Lumber summary text)
    escalated       BOOLEAN,            -- gate
    rule_hit        TEXT,               -- gate (which escalation rule fired)
    -- ON DELETE SET NULL on assessment_id so that deleting an assessment
    -- entry doesn't orphan the earlier stage events — the log's journey
    -- still reconstructs up to the gate decision.
    assessment_id   UUID REFERENCES agent_log(id) ON DELETE SET NULL,
    -- Forward-compat bag for rare extra fields without a column migration.
    metadata        JSONB NOT NULL DEFAULT '{}'::jsonb
);

-- Journey lookup: one log → ordered stage history.
CREATE INDEX idx_lpe_log_id ON log_pipeline_events (log_id, occurred_at);

-- Ticker + /bootstrap recent feed: newest N events for an app.
CREATE INDEX idx_lpe_app_occurred ON log_pipeline_events (app_id, occurred_at DESC);

-- Per-stage aggregates for PipelineStatsByApp (counts by stage in a window).
CREATE INDEX idx_lpe_stage ON log_pipeline_events (app_id, stage, occurred_at DESC);

ALTER TABLE log_pipeline_events ENABLE ROW LEVEL SECURITY;
