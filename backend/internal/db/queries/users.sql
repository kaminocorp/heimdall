-- name: GetUser :one
SELECT id, email, created_at FROM users WHERE id = $1;

-- name: GetFirstUserInOrg :one
SELECT om.user_id AS id FROM org_members om
WHERE om.org_id = $1
ORDER BY om.created_at ASC LIMIT 1;

-- name: GetUserByEmail :one
SELECT id, email, created_at FROM users WHERE email = $1;

-- name: HasOrgMembership :one
-- Returns true if the user belongs to any organization.
-- Used by the onboarding idempotency guard (replaces old user.OrgID.Valid check).
SELECT EXISTS(SELECT 1 FROM org_members WHERE user_id = $1) AS has_org;
