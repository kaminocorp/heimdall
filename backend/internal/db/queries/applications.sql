-- name: CreateApplication :one
INSERT INTO applications (org_id, name, status)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetApplication :one
SELECT * FROM applications WHERE id = $1;

-- name: ListApplicationsByOrg :many
SELECT * FROM applications
WHERE org_id = $1
ORDER BY created_at DESC;

-- name: UpdateApplication :one
UPDATE applications
SET name = $2, status = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteApplication :exec
DELETE FROM applications WHERE id = $1;
