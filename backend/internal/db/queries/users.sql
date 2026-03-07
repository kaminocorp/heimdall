-- name: GetUser :one
SELECT id, email, org_id, created_at FROM users WHERE id = $1;

-- name: SetUserOrg :exec
UPDATE users SET org_id = $1 WHERE id = $2;
