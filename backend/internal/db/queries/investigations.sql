-- name: CreateInvestigation :one
INSERT INTO investigations (
    trigger_type, trigger_source, summary, severity, status, context, user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetInvestigation :one
SELECT * FROM investigations WHERE id = $1 AND user_id = $2;

-- name: ListOpenInvestigations :many
SELECT * FROM investigations
WHERE user_id = $1 AND status IN ('open', 'investigating')
ORDER BY started_at DESC;

-- name: ListInvestigationsByDateRange :many
SELECT * FROM investigations
WHERE user_id = $1 AND started_at >= $2 AND started_at <= $3
ORDER BY started_at DESC;

-- name: UpdateInvestigationFindings :exec
UPDATE investigations
SET findings = $2, tool_trace = $3, status = $4, resolved_at = $5
WHERE id = $1 AND user_id = $6;

-- name: ResolveInvestigation :exec
UPDATE investigations
SET status = 'resolved', resolution = $2, resolved_at = now()
WHERE id = $1 AND user_id = $3;

-- name: DismissInvestigation :exec
UPDATE investigations
SET status = 'dismissed', resolution = $2
WHERE id = $1 AND user_id = $3;
