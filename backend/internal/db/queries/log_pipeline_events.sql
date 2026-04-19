-- log_pipeline_events: per-log stage rows for the Pipeline page.
-- Backed by migration 038. RLS enabled with no policies; owner-role pool only.

-- InsertPipelineEvent writes a single stage event. Used by every stage
-- call site — ingestion, classified, gate, assessment. `metadata` is NOT
-- NULL with a '{}' default in the schema; callers that don't need extra
-- fields can pass []byte("{}") or json.RawMessage(`{}`).
-- name: InsertPipelineEvent :one
INSERT INTO log_pipeline_events (
    log_id, app_id, stage,
    source_type, severity,
    type, category, confidence, summary,
    escalated, rule_hit,
    assessment_id,
    metadata
) VALUES (
    $1, $2, $3,
    $4, $5,
    $6, $7, $8, $9,
    $10, $11,
    $12,
    $13
)
RETURNING *;

-- GetPipelineEventsByLog returns every stage event for a single log_id,
-- ordered chronologically. Used by /pipeline/logs/{logId}/journey.
-- name: GetPipelineEventsByLog :many
SELECT * FROM log_pipeline_events
WHERE log_id = $1
ORDER BY occurred_at ASC, stage ASC;

-- GetRecentPipelineEventsByApp returns the most recent events for an app,
-- newest-first. Backs /pipeline/bootstrap's recentEvents payload and the
-- Time Machine list view.
-- name: GetRecentPipelineEventsByApp :many
SELECT * FROM log_pipeline_events
WHERE app_id = $1
ORDER BY occurred_at DESC
LIMIT $2;

-- GetPipelineEventsSince drains events newer than @since for an app, oldest
-- first. Used by the SSE /stream endpoint's catch-up replay to close the
-- hydration-to-stream gap. Callers cap the result (e.g. 500) — if the cap
-- is hit the stream emits a `resync` frame asking the client to refetch
-- /bootstrap.
-- name: GetPipelineEventsSince :many
SELECT * FROM log_pipeline_events
WHERE app_id = $1 AND occurred_at > $2
ORDER BY occurred_at ASC
LIMIT $3;

-- ListPipelineLogsByApp returns one summary row per log within a time window,
-- newest-last-seen first. Backs the Time Machine picker's list view.
--
-- The table stores one event per stage per log, so this GROUP BY collapses
-- a log's ≤4 stage rows into a single summary with:
--   * first_seen_at / last_seen_at — the journey's temporal span
--   * stage_count — how far the log made it (1 = only ingestion, 4 = full)
--   * source_type / severity — from the ingestion row
--   * type / category / confidence / summary — from the classified row
--   * escalated / rule_hit — from the gate row
--   * assessment_id — from the assessment row (nullable)
--
-- MAX(...) FILTER (WHERE stage = X) is safe here because there's exactly one
-- row per (log_id, stage) in practice — the writer emits each stage at most
-- once per log. BOOL_OR is the same story for the boolean column.
--
-- Uses idx_lpe_app_occurred for the WHERE + ORDER BY. The HAVING clause is
-- applied after GROUP BY so the ORDER BY's alias is accepted in HAVING by
-- Postgres (8.5+).
-- name: ListPipelineLogsByApp :many
SELECT
    log_id,
    MIN(occurred_at)::timestamptz                                       AS first_seen_at,
    MAX(occurred_at)::timestamptz                                       AS last_seen_at,
    COUNT(*)::bigint                                                    AS stage_count,
    MAX(source_type)   FILTER (WHERE stage = 'ingestion')               AS source_type,
    MAX(severity)      FILTER (WHERE stage = 'ingestion')               AS severity,
    MAX(type)          FILTER (WHERE stage = 'classified')              AS type,
    MAX(category)      FILTER (WHERE stage = 'classified')              AS category,
    MAX(confidence)    FILTER (WHERE stage = 'classified')              AS confidence,
    MAX(summary)       FILTER (WHERE stage = 'classified')              AS summary,
    BOOL_OR(escalated) FILTER (WHERE stage = 'gate')                    AS escalated,
    MAX(rule_hit)      FILTER (WHERE stage = 'gate')                    AS rule_hit,
    MAX(assessment_id) FILTER (WHERE stage = 'assessment')              AS assessment_id
FROM log_pipeline_events
WHERE app_id = $1
  AND occurred_at >= $2
  AND occurred_at <= $3
GROUP BY log_id
ORDER BY MAX(occurred_at) DESC
LIMIT $4 OFFSET $5;

-- PipelineStatsByApp returns per-stage event counts in a time window plus
-- a few derived aggregates used by the Sankey funnel widths. The single
-- scan groups by stage so the planner can use idx_lpe_stage.
--
-- Stages counted:
--   ingestion       — one per log that survived source filtering
--   classified      — one per log after Lumber returned
--   gate_flagged    — gate events with escalated = true
--   gate_safe       — gate events with escalated = false
--   assessment      — one per flagged log that landed in an agent_log batch
--
-- avg_confidence and flagged_ratio are computed over the window's
-- classified/gate rows respectively.
-- name: PipelineStatsByApp :one
SELECT
    COUNT(*) FILTER (WHERE stage = 'ingestion')::bigint                           AS ingestion_count,
    COUNT(*) FILTER (WHERE stage = 'classified')::bigint                          AS classified_count,
    COUNT(*) FILTER (WHERE stage = 'gate' AND escalated = true)::bigint           AS flagged_count,
    COUNT(*) FILTER (WHERE stage = 'gate' AND escalated = false)::bigint          AS safe_count,
    COUNT(*) FILTER (WHERE stage = 'assessment')::bigint                          AS assessment_count,
    COALESCE(AVG(confidence) FILTER (WHERE stage = 'classified'), 0)::float8      AS avg_confidence
FROM log_pipeline_events
WHERE app_id = $1 AND occurred_at >= $2;
