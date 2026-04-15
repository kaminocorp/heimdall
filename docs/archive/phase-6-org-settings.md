# Phase 6 — Organisation Settings Page

**Status:** Complete
**Date:** 2026-04-13
**Plan:** `docs/executing/org-overview-implementation.md`

---

## What changed

The `OrgSettingsPage.vue` shell from Phase 3 is now a fully functional settings
page with editable org name/slug, a copyable org ID, and a danger zone for
deleting the entire organisation. A new backend endpoint (`DELETE /api/org`)
supports the deletion flow.

---

## New backend endpoint

### `DELETE /api/org` — Delete organisation

| Aspect | Detail |
|--------|--------|
| Handler | `DeleteOrganization` in `organizations.go` |
| Auth | Owner only (`callerRole == OrgMemberRoleOwner`) |
| Confirmation | Request body `{ "confirm": "org-slug" }` must match the org's slug |
| Cascade | FK `ON DELETE CASCADE` handles all child data: apps, connections, configs, logs, schedules, notifications, org_members |
| Response | `204 No Content` |

### New sqlc query

`DeleteOrganization` — `DELETE FROM organizations WHERE id = $1`

### Route wiring

Added `r.Delete("/org", s.DeleteOrganization)` in both `router.go` and
`testhelpers_test.go`.

---

## Modified files

### `pages/org/OrgSettingsPage.vue`

**General section:**
- **Organisation ID** — read-only, displayed via `CopyableField` component
  (click-to-copy with success feedback)
- **Name** — editable text input for admin+, read-only div for members
- **Slug** — editable text input for admin+, read-only div for members
- **Created** — read-only date
- **Save button** — calls `PUT /api/org`, updates store on success so the
  header org name reflects the change immediately without a page reload.
  Shows "Saved" confirmation for 3 seconds.

**Danger zone** (owner only):
- Red-bordered section with description of consequences
- "Delete organisation" button opens confirmation modal
- Modal requires typing the org slug to enable the delete button
- On success: `appStore.reset()` → redirect to `/onboarding`
- Error feedback inline in the modal

**Permission gating:**

| Caller role | Can edit name/slug | Can see danger zone | Can delete |
|-------------|:---:|:---:|:---:|
| Member | No | No | No |
| Admin | Yes | No | No |
| Owner | Yes | Yes | Yes (with slug confirmation) |

### `api/organizations.ts`

Added `deleteOrganization(confirmSlug: string)` — sends `DELETE /api/org`
with `{ confirm: slug }` in the request body.

### `internal/api/handlers/organizations.go`

Added `DeleteOrganization` handler with:
- Owner-only role check via `resolveOrgAndRole`
- Slug confirmation validation
- Delegates to `DeleteOrganization` sqlc query

### `internal/api/router.go` + `testhelpers_test.go`

Wired `DELETE /api/org` route.

### `internal/db/queries/organizations.sql`

Added `DeleteOrganization` query.

---

## Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean (1.58s) |
| `vitest run` | 52/52 tests pass |
| `go build ./...` | Clean |
| `go vet ./...` | Clean |

---

## What's next

Phase 7 enables multi-org switching — the header org dropdown lists all orgs
the user belongs to, with the ability to switch between them and create new
organisations.
