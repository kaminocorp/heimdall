-- name: ListConnectionsByUser :many
SELECT * FROM connections WHERE user_id = $1 ORDER BY created_at DESC;

-- name: ListConnectionsByApp :many
SELECT * FROM connections WHERE app_id = $1 ORDER BY created_at DESC;

-- name: GetConnectionByUser :one
SELECT * FROM connections WHERE id = $1 AND user_id = $2;

-- name: CreateConnection :one
INSERT INTO connections (user_id, app_id, name, type, direction, config, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateConnection :one
UPDATE connections
SET name = $2, type = $3, direction = $4, config = $5, status = $6, updated_at = now()
WHERE id = $1 AND user_id = $7
RETURNING *;

-- name: UpdateConnectionStatus :exec
UPDATE connections SET status = $2, last_seen = now(), updated_at = now() WHERE id = $1;

-- name: DeleteConnectionByUser :exec
DELETE FROM connections WHERE id = $1 AND user_id = $2;
