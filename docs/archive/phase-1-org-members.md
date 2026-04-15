# Phase 1 — org_members Join Table

**Status:** Complete
**Date:** 2026-04-13
**Plan:** `docs/executing/org-overview-implementation.md`

---

## What changed

The `users.org_id` column (1:1 user-to-org relationship) has been replaced with
an `org_members` join table supporting many-to-many membership with roles.

---

## Migration: `026_org_members`

### Up migration

1. **Created `org_member_role` enum** — `'owner'`, `'admin'`, `'member'`
2. **Created `org_members` table** — composite PK on `(user_id, org_id)`, FK
   cascades to both `users` and `organizations`, with `role` and `created_at`
3. **Migrated existing data** — all users with a non-null `org_id` get an
   `org_members` row with role `'owner'`
4. **Rewrote 8 RLS policies** across tables `organizations`, `applications`,
   `app_agent_config`, `monitoring_state`, `notification_channels`,
   `notification_preferences`, `notification_log`, `investigation_schedules`
   — all changed from `JOIN users u ON u.org_id = a.org_id` to
   `JOIN org_members om ON om.org_id = a.org_id WHERE om.user_id = app_current_user_id()`
5. **Dropped `users.org_id`** column and its index

### Down migration

Fully reversible: re-adds `users.org_id`, copies data back from `org_members`
(picks earliest membership per user), reverts all 8 RLS policies to the old
pattern, then drops `org_members` table and `org_member_role` enum.

---

## sqlc query changes

### `users.sql`

| Query | Before | After |
|-------|--------|-------|
| `GetUser` | Selected `id, email, org_id, created_at` | Selects `id, email, created_at` (no org_id) |
| `SetUserOrg` | `UPDATE users SET org_id = $1` | **Removed** — replaced by `CreateOrgMember` |
| `GetFirstUserInOrg` | `WHERE org_id = $1` on `users` table | `SELECT om.user_id FROM org_members om WHERE om.org_id = $1` |
| `HasOrgMembership` | N/A | **New** — `SELECT EXISTS(... FROM org_members WHERE user_id = $1)` |

### `organizations.sql`

| Query | Before | After |
|-------|--------|-------|
| `GetOrganizationByUser` | `JOIN users u ON u.org_id = o.id` | `JOIN org_members om ON om.org_id = o.id` (returns earliest membership) |
| `ListOrganizationsByUser` | N/A | **New** — returns all orgs with role |
| `CreateOrgMember` | N/A | **New** — inserts `org_members` row |
| `GetOrgMembership` | N/A | **New** — single user+org lookup |
| `ListOrgMembers` | N/A | **New** — org members with email |
| `UpdateOrgMemberRole` | N/A | **New** — change member role |
| `DeleteOrgMember` | N/A | **New** — remove membership |
| `CountOrgMembers` | N/A | **New** — count members in org |

### `applications.sql`

| Query | Change |
|-------|--------|
| `GetApplicationByOrgUser` | `JOIN users u ON u.org_id = a.org_id` → `JOIN org_members om ON om.org_id = a.org_id` |

### `monitoring.sql`

No changes — `ListActiveApplications` selects `a.org_id` from `applications`
table (unchanged column), not from `users`.

---

## Go code changes

### `internal/api/handlers/organizations.go`

- **Onboard idempotency guard**: `user.OrgID.Valid` → `HasOrgMembership(userID)`
- **Org linking**: `SetUserOrg(SetUserOrgParams{...})` → `CreateOrgMember(CreateOrgMemberParams{UserID, OrgID, Role: OrgMemberRoleOwner})`
- **Removed import**: `pgtype` no longer needed

### `internal/agent/monitor.go`

- **`resolveOrgUser`**: Removed `pgtype.UUID` wrapping — `GetFirstUserInOrg`
  now accepts `uuid.UUID` directly (since `org_members.org_id` is non-nullable)
- **Removed import**: `pgtype` no longer needed

### `internal/api/handlers/testhelpers_test.go`

- Test data setup: `UPDATE users SET org_id = $1` → `INSERT INTO org_members (user_id, org_id, role) VALUES ($1, $2, 'owner')`

---

## Generated code changes (sqlc)

- **`models.go`**: `User` struct lost `OrgID` field; new `OrgMember` struct and
  `OrgMemberRole` enum type added
- **`users.sql.go`**: `GetUser` returns `User` directly (no more `GetUserRow`);
  `SetUserOrg` removed; `GetFirstUserInOrg` parameter changed from
  `pgtype.UUID` to `uuid.UUID`; new `HasOrgMembership` function
- **`organizations.sql.go`**: New functions for membership CRUD; `GetOrganizationByUser`
  now joins via `org_members`

---

## Verification

| Check | Result |
|-------|--------|
| `go build ./...` | Clean |
| `go vet ./...` | Clean |
| `vue-tsc --noEmit` | Clean |
| Frontend unaffected | `applications.org_id` column unchanged; no frontend types reference `users.org_id` |

---

## What's next

Phase 2 adds the backend API endpoints that expose the new `org_members` data:
`GET /api/org/members`, `POST /api/org/members/invite`, role management, and
`GET /api/orgs` for multi-org listing.
