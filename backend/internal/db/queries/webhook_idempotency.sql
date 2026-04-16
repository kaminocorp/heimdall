-- name: GetIdempotencyResult :one
SELECT response_status, response_body FROM webhook_idempotency
WHERE connection_id = $1 AND idempotency_key = $2 AND created_at > now() - interval '24 hours';

-- name: InsertIdempotencyResult :exec
INSERT INTO webhook_idempotency (connection_id, idempotency_key, response_status, response_body)
VALUES ($1, $2, $3, $4)
ON CONFLICT (connection_id, idempotency_key) DO NOTHING;

-- name: PruneExpiredIdempotencyKeys :execrows
DELETE FROM webhook_idempotency WHERE created_at < now() - interval '24 hours';
