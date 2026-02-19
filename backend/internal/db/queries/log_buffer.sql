-- name: InsertLogEntry :exec
INSERT INTO log_buffer (connection_id, source_type, severity, payload)
VALUES ($1, $2, $3, $4);

-- name: ListRecentLogs :many
SELECT * FROM log_buffer
WHERE ingested_at > now() - interval '1 hour'
ORDER BY ingested_at DESC;

-- name: ListLogsByConnection :many
SELECT * FROM log_buffer
WHERE connection_id = $1 AND ingested_at > now() - interval '1 hour'
ORDER BY ingested_at DESC;

-- name: ListLogsBySeverity :many
SELECT * FROM log_buffer
WHERE severity = $1 AND ingested_at > now() - interval '1 hour'
ORDER BY ingested_at DESC;

-- name: PruneExpiredLogs :execrows
DELETE FROM log_buffer WHERE ingested_at < now() - interval '24 hours';
