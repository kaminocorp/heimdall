-- name: CreateOrganization :one
INSERT INTO organizations (name, slug)
VALUES ($1, $2)
RETURNING *;

-- name: GetOrganization :one
SELECT * FROM organizations WHERE id = $1;

-- name: GetOrganizationBySlug :one
SELECT * FROM organizations WHERE slug = $1;

-- name: GetOrganizationByUser :one
SELECT o.* FROM organizations o
JOIN users u ON u.org_id = o.id
WHERE u.id = $1;

-- name: UpdateOrganization :one
UPDATE organizations
SET name = $2, slug = $3, updated_at = now()
WHERE id = $1
RETURNING *;
