-- name: ListConnectionsByUser :many
-- Every connection in any org the authenticated user is a member of.
-- The org-member join replaces the single-owner user_id check that existed
-- pre-035 — an org-scoped connection created by a colleague is now visible
-- to every member of the org.
SELECT c.*
FROM connections c
JOIN org_members om ON om.org_id = c.org_id
WHERE om.user_id = $1
ORDER BY c.created_at DESC;

-- name: ListConnectionsByApp :many
-- Returns app-scoped connections AND org-scoped connections that belong to
-- the same org as the requested app. The nested subquery resolves the app's
-- org once per call — cheaper than joining because we only need org_id.
SELECT c.*
FROM connections c
WHERE c.app_id = sqlc.arg(app_id)::UUID
   OR (
     c.app_id IS NULL
     AND c.org_id = (SELECT org_id FROM applications WHERE id = sqlc.arg(app_id)::UUID)
   )
ORDER BY c.created_at DESC;

-- name: GetConnectionByUser :one
-- Access is granted via org membership rather than direct ownership. Name
-- preserved for back-compat with Phase 1 callers — the param list is
-- unchanged (id, user_id), only the predicate shifts to "user is a member
-- of connection's org".
SELECT c.*
FROM connections c
JOIN org_members om ON om.org_id = c.org_id
WHERE c.id = sqlc.arg(id)::UUID AND om.user_id = sqlc.arg(user_id)::UUID;

-- name: CreateConnection :one
-- org_id is always required; app_id is optional — pass NULL to create an
-- org-scoped connection shared across every app in the org. sqlc.narg yields
-- a *uuid.UUID thanks to the nullable-uuid override in sqlc.yaml.
INSERT INTO connections (user_id, org_id, app_id, name, type, direction, config, status)
VALUES (
    sqlc.arg(user_id)::UUID,
    sqlc.arg(org_id)::UUID,
    sqlc.narg(app_id),
    sqlc.arg(name)::TEXT,
    sqlc.arg(type)::TEXT,
    sqlc.arg(direction)::TEXT,
    sqlc.arg(config)::JSONB,
    sqlc.arg(status)::TEXT
)
RETURNING *;

-- name: UpdateConnection :one
-- Authorization is via org membership, matching the Get/Delete queries.
-- The transition away from user_id-based filtering is transparent to
-- callers — the user_id param is still the caller's user id, just
-- interpreted as "must be a member of connection's org".
UPDATE connections c
SET name = $2, type = $3, direction = $4, config = $5, status = $6, updated_at = now()
WHERE c.id = $1
  AND c.org_id IN (
    SELECT om.org_id FROM org_members om WHERE om.user_id = $7
  )
RETURNING *;

-- name: UpdateConnectionStatus :exec
UPDATE connections SET status = $2, last_seen = now(), updated_at = now() WHERE id = $1;

-- name: DeleteConnectionByUser :exec
DELETE FROM connections c
WHERE c.id = $1
  AND c.org_id IN (
    SELECT om.org_id FROM org_members om WHERE om.user_id = $2
  );

-- name: ListActiveConnectionsByType :many
-- Startup rehydration for pollers/listeners. No access filter — the server
-- connects as the postgres (owner) role which bypasses RLS, and polling
-- connections are always app-scoped in practice.
SELECT * FROM connections WHERE type = $1 AND status = 'active';

-- name: ListAppsEnabledForSource :many
-- Fan-out helper for org-scoped webhook ingestion. Returns every app that
-- has enabled=true for the given (connection, source_name). Backed by the
-- partial index idx_app_source_filters_lookup from migration 034.
SELECT app_id FROM app_source_filters
WHERE connection_id = $1 AND source_name = $2 AND enabled = true;
