-- name: InsertLogEntry :one
INSERT INTO log_buffer (connection_id, source_type, severity, payload, user_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListLogsByUser :many
SELECT * FROM log_buffer
WHERE user_id = $1
ORDER BY ingested_at DESC
LIMIT $2 OFFSET $3;

-- name: ListLogsByUserAndSeverity :many
SELECT * FROM log_buffer
WHERE user_id = $1 AND severity = $2
ORDER BY ingested_at DESC
LIMIT $3 OFFSET $4;

-- name: ListLogsByUserAndConnection :many
SELECT * FROM log_buffer
WHERE user_id = $1 AND connection_id = $2
ORDER BY ingested_at DESC
LIMIT $3 OFFSET $4;

-- name: CountLogsByUser :one
SELECT count(*) FROM log_buffer
WHERE user_id = $1;

-- name: CountLogsByUserAndSeverity :one
SELECT count(*) FROM log_buffer
WHERE user_id = $1 AND severity = $2;

-- name: CountLogsByUserAndConnection :one
SELECT count(*) FROM log_buffer
WHERE user_id = $1 AND connection_id = $2;

-- name: GetConnectionByWebhookToken :one
SELECT * FROM connections
WHERE config->>'webhook_token' = @webhook_token::text AND type = 'webhook_logs' AND status = 'active';

-- name: SearchLogsByUser :many
SELECT * FROM log_buffer
WHERE user_id = @user_id AND payload::text ILIKE '%' || @query::text || '%' ESCAPE '\'
ORDER BY ingested_at DESC
LIMIT @row_limit OFFSET @row_offset;

-- name: SearchLogsByUserAndSeverity :many
SELECT * FROM log_buffer
WHERE user_id = @user_id AND payload::text ILIKE '%' || @query::text || '%' ESCAPE '\' AND severity = @severity
ORDER BY ingested_at DESC
LIMIT @row_limit OFFSET @row_offset;

-- name: PruneExpiredLogs :execrows
DELETE FROM log_buffer WHERE ingested_at < now() - interval '48 hours';
