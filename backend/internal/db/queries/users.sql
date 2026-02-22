-- name: GetUser :one
SELECT id, email, created_at FROM users WHERE id = $1;
