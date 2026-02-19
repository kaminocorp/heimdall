-- name: ListConnections :many
SELECT * FROM connections ORDER BY created_at DESC;

-- name: GetConnection :one
SELECT * FROM connections WHERE id = $1;

-- name: CreateConnection :one
INSERT INTO connections (name, type, direction, config, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateConnection :one
UPDATE connections
SET name = $2, type = $3, direction = $4, config = $5, status = $6, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateConnectionStatus :exec
UPDATE connections SET status = $2, last_seen = now(), updated_at = now() WHERE id = $1;

-- name: DeleteConnection :exec
DELETE FROM connections WHERE id = $1;
