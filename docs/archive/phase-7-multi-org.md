# Phase 7 — Multi-Org Switching

**Status:** Complete
**Date:** 2026-04-13
**Plan:** `docs/executing/org-overview-implementation.md`

---

## What changed

Users who belong to multiple organisations can now switch between them via a
dropdown in the header breadcrumb. A new "Create organisation" flow allows
creating additional orgs from the header. The app store has been restructured
to track multiple organisations and a `currentOrgId`.

---

## New backend endpoint

### `POST /api/orgs` — Create organisation

| Aspect | Detail |
|--------|--------|
| Handler | `CreateNewOrganization` in `organizations.go` |
| Auth | Any authenticated user |
| Body | `{ "name": "...", "slug": "..." }` |
| Behaviour | Creates org, adds caller as owner (transactional) |
| Response | `201 Created` with the new `Organization` object |

Unlike `POST /api/onboard` (which creates org + default app + agent config),
this endpoint creates only the org and its membership row. The user can then
create apps within the org at their own pace from the overview page.

---

## Store changes (`stores/app.ts`)

### New state

| Field | Type | Persistence |
|-------|------|-------------|
| `organizations` | `OrganizationWithRole[]` | In-memory (fetched on init) |

### New methods

| Method | Purpose |
|--------|---------|
| `selectOrg(orgId)` | Switch to a different org. Sets `organization`, persists to localStorage, refetches apps, clears current app, selects first app. |
| `addOrg(org)` | Add a newly created org to the list and select it. Called by CreateOrgModal after `POST /api/orgs`. |

### Modified `init()`

**Before:** Fetched single org via `GET /api/org`, then apps.

**After:**
1. Fetches all orgs via `GET /api/orgs` (returns list with roles)
2. Restores `currentOrgId` from `localStorage` key `heimdall_current_org`
3. Falls back to first org if stored ID not found
4. Sets `organization` to the active org (includes role)
5. Fetches apps for that org
6. Restores `currentAppId` from localStorage

### Modified `onboard()`

Now also populates `organizations` list with the new org (as owner).

### Modified `reset()`

Also clears `organizations` list.

---

## New frontend files

### `components/org/CreateOrgModal.vue`

Modal dialog for creating a new organisation:
- Name input with auto-generated slug (slug updates as you type name)
- Slug input editable independently (auto-slug disables on manual edit)
- Slug sanitisation: lowercase, hyphens only, max 48 chars
- Creates via `POST /api/orgs`, then calls `appStore.addOrg()` to add
  to the local list and switch to the new org
- Error feedback inline

---

## Modified files

### `components/common/AppHeader.vue`

The org breadcrumb is now a **dropdown** (reverted from Phase 3's plain
`RouterLink`) that lists all organisations:

**Dropdown contents:**
- All orgs the user belongs to, each showing name + role badge
- Current org highlighted with accent background
- Clicking a different org: `app.selectOrg(orgId)` → navigates to `/org`
- Clicking the current org: navigates to `/org` (quick access to overview)
- "+ New organisation" option at bottom → opens `CreateOrgModal`

**Role badges in dropdown:**
- Owner: green accent
- Admin: amber/warn
- Member: neutral

### `api/organizations.ts`

Added `createOrganization(name, slug)` — `POST /api/orgs`.

### `internal/api/router.go` + `testhelpers_test.go`

Wired `POST /api/orgs` route.

---

## Navigation flow with multi-org

```
User has 2 orgs: "Crimson Sun" (owner) + "Acme Corp" (member)

Header shows: [Heimdall] / Crimson Sun ▾ / my-app ▾

User clicks org chevron → dropdown shows:
  ● Crimson Sun     OWNER   ← highlighted
    Acme Corp       MEMBER
    ─────────────────────
    + New organisation

User clicks "Acme Corp"
  → app.selectOrg('acme-id')
     → organization = acme
     → apps = [...acme's apps]
     → currentAppId = acme's first app
  → router.push('/org')
  → Header shows: [Heimdall] / Acme Corp ▾
  → OrgSidebar, org overview with Acme's projects

User clicks a project card
  → app.selectApp(appId)
  → router.push('/dashboard')
  → Header shows: [Heimdall] / Acme Corp ▾ / acme-prod ▾
  → AppSidebar, dashboard for that app
```

---

## Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean (1.55s) |
| `vitest run` | 52/52 tests pass |
| `go build ./...` | Clean |
| `go vet ./...` | Clean |

---

## What's next

Phase 8 is the final polish pass — remove the old `/settings` page, smooth
sidebar transitions, breadcrumb refinement, mobile nav, and empty states.
