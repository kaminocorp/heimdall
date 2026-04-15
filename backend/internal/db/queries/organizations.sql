-- name: CreateOrganization :one
INSERT INTO organizations (name, slug)
VALUES ($1, $2)
RETURNING *;

-- name: GetOrganization :one
SELECT * FROM organizations WHERE id = $1;

-- name: GetOrganizationBySlug :one
SELECT * FROM organizations WHERE slug = $1;

-- name: GetOrganizationByUser :one
-- Returns the user's primary org (earliest membership).
-- For multi-org users, use ListOrganizationsByUser instead.
SELECT o.* FROM organizations o
JOIN org_members om ON om.org_id = o.id
WHERE om.user_id = $1
ORDER BY om.created_at ASC
LIMIT 1;

-- name: ListOrganizationsByUser :many
-- Returns all organizations the user belongs to.
SELECT o.*, om.role FROM organizations o
JOIN org_members om ON om.org_id = o.id
WHERE om.user_id = $1
ORDER BY om.created_at ASC;

-- name: UpdateOrganization :one
UPDATE organizations
SET name = $2, slug = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CreateOrgMember :exec
INSERT INTO org_members (user_id, org_id, role)
VALUES ($1, $2, $3);

-- name: GetOrgMembership :one
-- Returns the membership row for a specific user+org pair.
SELECT * FROM org_members WHERE user_id = $1 AND org_id = $2;

-- name: ListOrgMembers :many
-- Returns all members of an organization with their email.
SELECT om.user_id, u.email, om.role, om.created_at
FROM org_members om
JOIN users u ON u.id = om.user_id
WHERE om.org_id = $1
ORDER BY om.created_at ASC;

-- name: UpdateOrgMemberRole :exec
UPDATE org_members SET role = $3
WHERE user_id = $1 AND org_id = $2;

-- name: DeleteOrgMember :exec
DELETE FROM org_members WHERE user_id = $1 AND org_id = $2;

-- name: CountOrgMembers :one
SELECT COUNT(*) FROM org_members WHERE org_id = $1;

-- name: CountOrgOwners :one
-- Counts how many owners an org has. Used by the sole-owner removal guard.
SELECT COUNT(*) FROM org_members WHERE org_id = $1 AND role = 'owner';

-- name: DeleteOrganization :exec
-- Cascade-deletes all applications, connections, configs, logs, schedules, etc.
-- via FK ON DELETE CASCADE constraints.
DELETE FROM organizations WHERE id = $1;
