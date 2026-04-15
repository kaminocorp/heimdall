# Phase 4 — Org Overview Page (Projects Grid)

**Status:** Complete
**Date:** 2026-04-13
**Plan:** `docs/executing/org-overview-implementation.md`

---

## What changed

The `OrgOverviewPage.vue` shell from Phase 3 is now a fully functional
projects grid showing all applications in the organisation with search,
sorting, grid/list toggle, and a "New project" creation flow.

---

## New files

### `components/org/AppCard.vue`

Grid-view card for a single application. Shows:
- App name (clickable, navigates to Dashboard in app context)
- Status badge (active/paused/archived) via `StatusBadge`
- Connection count with link icon
- Schedule count with clock icon
- Created date

Click handler emits `select` event with `appId`. Paused apps render at
reduced opacity.

### `components/org/AppListRow.vue`

List-view row for a single application. Compact table-row layout showing
the same data as AppCard in a horizontal format:
- Name + status badge (flex-1)
- Connection count (fixed width)
- Schedule count (fixed width)
- Created date (hidden on mobile)

Same click/emit pattern as AppCard.

---

## Modified files

### `pages/org/OrgOverviewPage.vue`

Replaced placeholder shell with full implementation:

**Toolbar:**
- Search input with magnifying glass icon — client-side filter by app name
- Sort dropdown — by name (default), date, or status
- Grid/List toggle — persists to `localStorage` key `heimdall_org_view`
- "+ New project" button — opens existing `AppWizard` component

**Content states:**
| State | Render |
|-------|--------|
| Loading | 3-column skeleton grid |
| Error | Red banner with retry button |
| Empty (no apps) | Centred CTA with database icon |
| No search results | "No projects matching ..." message |
| Grid view | Responsive 1/2/3-column card grid |
| List view | Table with header row and `AppListRow` items |

**Card/row click navigation:**
`selectApp(appId)` → `appStore.selectApp(appId)` → `router.push('/dashboard')`

This "zooms in" from org context to app context — the sidebar switches from
OrgSidebar to AppSidebar, the header shows the app breadcrumb, and the
dashboard loads data for the selected app.

**Wizard integration:**
On wizard close (success or cancel), the page refetches `listApplicationsWithCounts`
to ensure the grid stays in sync.

### `types/organization.ts`

Added types for the new backend endpoints:
- `OrgMemberRole` — `'owner' | 'admin' | 'member'` union type
- `OrgMember` — member list item (`user_id`, `email`, `role`, `created_at`)
- `OrganizationWithRole` — `Organization` extended with required `role` field
- `Organization.role` — made optional on base type (backwards compatible)

### `api/organizations.ts`

Added 5 new API client functions:

| Function | Method | Path | Purpose |
|----------|--------|------|---------|
| `updateOrganization` | PUT | `/org` | Update org name/slug |
| `listUserOrganizations` | GET | `/orgs` | Multi-org listing |
| `listOrgMembers` | GET | `/org/members` | Team member list |
| `inviteOrgMember` | POST | `/org/members/invite` | Add member by email |
| `updateMemberRole` | PUT | `/org/members/{userId}/role` | Change role |
| `removeOrgMember` | DELETE | `/org/members/{userId}` | Remove member |

These are wired to the Phase 2 backend endpoints but not yet consumed by
UI components (Teams page will use them in Phase 5).

---

## Data flow

```
OrgOverviewPage mounts
  → listApplicationsWithCounts() → GET /api/apps?include=counts
  → Response: ApplicationWithCounts[] (name, status, connection_count, schedule_count)
  → Rendered as AppCard grid or AppListRow table

User types in search
  → Client-side filter on name (computed, no API call)

User clicks card/row
  → appStore.selectApp(appId) — updates localStorage + reactive state
  → router.push('/dashboard') — enters app context
  → Sidebar switches, header updates, Dashboard fetches app-specific data
```

---

## Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean (1.68s) |
| `vitest run` | 52/52 tests pass |

---

## What's next

Phase 5 builds the Team management page (`/org/team`) — member list, invite
flow, role management — consuming the `listOrgMembers`, `inviteOrgMember`,
`updateMemberRole`, and `removeOrgMember` API functions added in this phase.
