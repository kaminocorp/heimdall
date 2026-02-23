-- name: InsertLogEntry :one
INSERT INTO log_buffer (connection_id, source_type, severity, payload)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListLogsByUser :many
SELECT lb.* FROM log_buffer lb
JOIN connections c ON c.id = lb.connection_id
WHERE c.user_id = $1
ORDER BY lb.ingested_at DESC
LIMIT $2 OFFSET $3;

-- name: ListLogsByUserAndSeverity :many
SELECT lb.* FROM log_buffer lb
JOIN connections c ON c.id = lb.connection_id
WHERE c.user_id = $1 AND lb.severity = $2
ORDER BY lb.ingested_at DESC
LIMIT $3 OFFSET $4;

-- name: ListLogsByUserAndConnection :many
SELECT lb.* FROM log_buffer lb
JOIN connections c ON c.id = lb.connection_id
WHERE c.user_id = $1 AND lb.connection_id = $2
ORDER BY lb.ingested_at DESC
LIMIT $3 OFFSET $4;

-- name: CountLogsByUser :one
SELECT count(*) FROM log_buffer lb
JOIN connections c ON c.id = lb.connection_id
WHERE c.user_id = $1;

-- name: GetConnectionByWebhookToken :one
SELECT * FROM connections
WHERE config->>'webhook_token' = @webhook_token::text AND type = 'webhook_logs' AND status = 'active';

-- name: PruneExpiredLogs :execrows
DELETE FROM log_buffer WHERE ingested_at < now() - interval '24 hours';
