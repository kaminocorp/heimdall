# Phase 2 — Org-Level API Endpoints

**Status:** Complete
**Date:** 2026-04-13
**Plan:** `docs/executing/org-overview-implementation.md`

---

## What changed

New API endpoints for organization membership management, multi-org listing,
org updates, and role-based authorization. The frontend can now power an org
overview page, team management, and an org switcher.

---

## New endpoints

### Organization management

| Method | Path | Handler | Auth | Description |
|--------|------|---------|------|-------------|
| `GET` | `/api/org` | `GetOrganization` | any member | Returns org **with caller's role** (new `role` field) |
| `PUT` | `/api/org` | `UpdateOrganization` | admin+ | Update org name and/or slug |
| `GET` | `/api/orgs` | `ListUserOrganizations` | any | List all orgs the user belongs to (multi-org switcher) |

### Membership management

| Method | Path | Handler | Auth | Description |
|--------|------|---------|------|-------------|
| `GET` | `/api/org/members` | `ListOrgMembers` | any member | List all members with email and role |
| `POST` | `/api/org/members/invite` | `InviteMember` | admin+ | Add user by email (must have existing account) |
| `PUT` | `/api/org/members/{userId}/role` | `UpdateMemberRole` | owner only | Change a member's role |
| `DELETE` | `/api/org/members/{userId}` | `RemoveMember` | admin+ | Remove member (sole-owner guard) |

---

## New files

### `internal/api/handlers/org_members.go`

Contains all membership-related handlers plus two shared helpers:

- **`resolveOrgAndRole(w, r)`** — fetches the caller's org and their
  `org_member_role` in a single call. Returns `(org, role, ok)`. Used by
  every handler that needs role-based authorization.

- **`hasMinRole(actual, required)`** — compares roles against the hierarchy
  `owner (3) > admin (2) > member (1)`. Used instead of middleware so each
  handler can apply its own minimum role requirement inline.

**Design decision: helpers instead of middleware.** The original plan proposed a
`requireOrgRole(minRole)` middleware, but that approach requires the org to be
resolved in middleware context (before the handler runs) and stored in the
request context. Since `resolveOrgAndRole` is only ~15 lines and the role check
is a one-liner, keeping it as handler-level helpers is simpler, avoids context
key management, and makes the authorization logic visible in each handler
instead of hidden in middleware configuration.

---

## Modified files

### `internal/api/handlers/organizations.go`

- **`GetOrganization`**: Now includes `role` field in response by querying
  `GetOrgMembership` after fetching the org. Falls back to role-less response
  if membership lookup fails (shouldn't happen, but defensive).

- **`ListUserOrganizations`** (new): Delegates to `ListOrganizationsByUser` sqlc
  query which returns orgs with role via the `org_members` join.

- **`UpdateOrganization`** (new): Uses `resolveOrgAndRole` to enforce admin+
  access, then calls `UpdateOrganization` query. Preserves existing values for
  omitted fields (partial update semantics).

### `internal/api/router.go`

New route group under the existing `/api` protected block:

```go
r.Put("/org", s.UpdateOrganization)
r.Get("/orgs", s.ListUserOrganizations)
r.Route("/org/members", func(r chi.Router) {
    r.Get("/", s.ListOrgMembers)
    r.Post("/invite", s.InviteMember)
    r.Put("/{userId}/role", s.UpdateMemberRole)
    r.Delete("/{userId}", s.RemoveMember)
})
```

### `internal/api/handlers/testhelpers_test.go`

Test router mirrors all new routes so integration tests can exercise them.

---

## New sqlc query

### `users.sql`

| Query | Purpose |
|-------|---------|
| `GetUserByEmail` | Look up user by email for the invite flow |

---

## Role-based authorization rules

| Action | Required role | Guard logic |
|--------|--------------|-------------|
| List members | any member | `resolveOrgAndRole` succeeds |
| Invite member | admin+ | `hasMinRole(role, OrgMemberRoleAdmin)` |
| Update role | owner | `role == OrgMemberRoleOwner` |
| Remove member | admin+ | `hasMinRole(role, OrgMemberRoleAdmin)` + cannot remove owner unless caller is owner |
| Update org | admin+ | `hasMinRole(role, OrgMemberRoleAdmin)` |
| Sole-owner guard | — | `CountOrgMembers <= 1` prevents sole member from removing themselves |

---

## Invite flow (V1 limitations)

The invite endpoint currently requires the target user to **already have a
Heimdall account** (`GetUserByEmail` must return a row). If the email isn't
found, the endpoint returns 404 with a clear message. Future phases can add:

- Pending invites table with email-based lookup on signup
- Email notification via Supabase edge function
- Invite link generation

---

## Verification

| Check | Result |
|-------|--------|
| `go build ./...` | Clean |
| `go vet ./...` | Clean |
| `vue-tsc --noEmit` | Clean |
| Route structure | Mirrors production router in test helpers |

---

## API response shapes

### `GET /api/org` (updated)

```json
{
  "id": "uuid",
  "name": "Crimson Sun Technologies",
  "slug": "crimson-sun",
  "created_at": "2026-...",
  "updated_at": "2026-...",
  "role": "owner"
}
```

### `GET /api/orgs`

```json
[
  {
    "id": "uuid",
    "name": "Crimson Sun Technologies",
    "slug": "crimson-sun",
    "created_at": "2026-...",
    "updated_at": "2026-...",
    "role": "owner"
  }
]
```

### `GET /api/org/members`

```json
[
  {
    "user_id": "uuid",
    "email": "user@example.com",
    "role": "owner",
    "created_at": "2026-..."
  }
]
```

---

## What's next

Phase 3 adds the frontend routing and layout context switching — `/org/*` routes
get an org sidebar, app routes keep the current sidebar, and the header
breadcrumb becomes the navigation switch mechanism.
