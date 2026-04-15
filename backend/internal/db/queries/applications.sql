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

-- name: ListApplicationsByOrgWithCounts :many
-- Same as ListApplicationsByOrg but augments each row with connection/schedule
-- counts so the Settings page can render per-app summaries in a single request
-- instead of N+1. Subqueries keep the cost linear in #apps — the counts never
-- join across connections/schedules rows, so adding more data to either table
-- has no multiplicative effect on the query plan.
SELECT
  a.*,
  (SELECT COUNT(*) FROM connections WHERE app_id = a.id) AS connection_count,
  (SELECT COUNT(*) FROM investigation_schedules WHERE app_id = a.id) AS schedule_count
FROM applications a
WHERE a.org_id = $1
ORDER BY a.created_at DESC;

-- name: CountApplicationsByOrg :one
-- Used by the delete-app handler to enforce the "can't delete the last app"
-- guard without pulling back every row.
SELECT COUNT(*) FROM applications WHERE org_id = $1;

-- name: UpdateApplication :one
UPDATE applications
SET name = $2, status = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetApplicationByOrgUser :one
-- Returns the application only if it belongs to an org the given user is a member of.
SELECT a.* FROM applications a
JOIN org_members om ON om.org_id = a.org_id
WHERE a.id = sqlc.arg(app_id) AND om.user_id = sqlc.arg(user_id);

-- name: DeleteApplication :exec
DELETE FROM applications WHERE id = $1;
