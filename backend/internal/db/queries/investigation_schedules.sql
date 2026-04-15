-- name: ListEnabledSchedules :many
-- Used by the scheduler loop on each tick. Ordered so never-run schedules
-- (NULL last_run_at) fire first, then oldest-last-run first.
SELECT * FROM investigation_schedules
WHERE enabled = true
ORDER BY last_run_at ASC NULLS FIRST;

-- name: ListSchedulesByApp :many
SELECT * FROM investigation_schedules
WHERE app_id = $1
ORDER BY created_at DESC;

-- name: GetSchedule :one
SELECT * FROM investigation_schedules
WHERE id = $1;

-- name: CreateSchedule :one
-- cron_expr is nullable: schedules created via the simple interval preset
-- store NULL here and rely on interval_secs. Schedules created via cron_expr
-- store 0 for interval_secs (unused; shouldFire prefers cron when set).
INSERT INTO investigation_schedules (app_id, name, prompt, interval_secs, cron_expr, enabled)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateSchedule :one
-- PATCH-style update that overwrites all four scheduling fields; the handler
-- is responsible for enforcing "exactly one of interval_secs / cron_expr".
UPDATE investigation_schedules
SET name = $2,
    prompt = $3,
    interval_secs = $4,
    cron_expr = $5,
    enabled = $6,
    updated_at = now()
WHERE id = $1 AND app_id = $7
RETURNING *;

-- name: DeleteSchedule :exec
DELETE FROM investigation_schedules WHERE id = $1 AND app_id = $2;

-- name: MarkScheduleRun :exec
-- Called by the scheduler after every run (success or error) to advance
-- last_run_at and record the outcome. Takes nullable status/error/summary
-- so both the success and error paths share a single query.
UPDATE investigation_schedules
SET last_run_at  = now(),
    last_status  = $2,
    last_error   = $3,
    last_summary = $4,
    updated_at   = now()
WHERE id = $1;
