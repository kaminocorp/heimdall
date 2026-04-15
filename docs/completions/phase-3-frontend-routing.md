# Phase 3 — Frontend Routing & Layout Context Switching

**Status:** Complete
**Date:** 2026-04-13
**Plan:** `docs/executing/org-overview-implementation.md`

---

## What changed

The frontend now supports two navigation contexts: **org-level** (Projects,
Team, Billing, Settings) and **app-level** (Dashboard, Activity, Connections,
etc.). Clicking the organisation name in the header navigates to `/org`, where
the sidebar switches to org-level navigation. Clicking into an app returns to
app-level navigation.

---

## New files

### Page shells (`pages/org/`)

| Page | Route | Purpose |
|------|-------|---------|
| `OrgOverviewPage.vue` | `/org` | Projects grid (Phase 4 content) |
| `OrgTeamPage.vue` | `/org/team` | Team management (Phase 5 content) |
| `OrgSettingsPage.vue` | `/org/settings` | Org settings (Phase 6 content) |
| `OrgBillingPage.vue` | `/org/billing` | Billing placeholder |

All pages use the standard page header pattern (`pb-6 mb-8 border-b border-border`)
for visual consistency. Content is placeholder text pointing to the phase that
will fill them in.

### `OrgSidebar.vue` (`components/common/`)

Org-level sidebar with two sections:

```
Organisation
  - Projects      → /org
  - Team          → /org/team

Management
  - Billing       → /org/billing
  - Settings      → /org/settings
```

Identical visual structure to `AppSidebar.vue` — same section labels, same
`RouterLink` pattern, same active indicator bar with accent glow. Accepts the
same `mobile` prop and emits `close` for the mobile overlay drawer.

---

## Modified files

### `router/index.ts`

**Added 4 org routes** before the app routes block:

```typescript
{ path: '/org', name: 'org-overview', component: OrgOverviewPage },
{ path: '/org/team', name: 'org-team', component: OrgTeamPage },
{ path: '/org/settings', name: 'org-settings', component: OrgSettingsPage },
{ path: '/org/billing', name: 'org-billing', component: OrgBillingPage },
```

**Redirected `/settings`** to `/org/settings` (preserves bookmarks):

```typescript
{ path: '/settings', redirect: '/org/settings' }
```

**Route guard unchanged.** The existing guard logic correctly handles org routes:
- Auth check applies to all non-public routes (org routes require login)
- Onboarding redirect applies (no org = can't view org pages)
- No `currentAppId` check exists in the guard (it's page-level), so org pages
  work without an app selected

### `layouts/DefaultLayout.vue`

**Context-aware sidebar** using a computed `isOrgContext`:

```typescript
const isOrgContext = computed(() => route.path.startsWith('/org'))
```

Both the desktop sidebar and mobile overlay drawer now conditionally render
`OrgSidebar` or `AppSidebar` based on this computed.

### `components/common/AppHeader.vue`

**Org breadcrumb is now a `RouterLink` to `/org`** instead of a dropdown toggle.
The old dropdown (which just showed the current org name) is removed. The org
name is always visible and clickable — it's the primary "zoom out" action.

**App breadcrumb is conditionally rendered.** When on `/org/*` routes, the app
selector (separator + dropdown) is hidden. When on app routes, it shows as
before.

```
Org context:  [Heimdall] / Crimson Sun Technologies
App context:  [Heimdall] / Crimson Sun Technologies / heimdall-prod ▾
```

**Removed state:** `orgOpen`, `orgRef`, `toggleOrg()` — no longer needed since
the org breadcrumb is a link, not a dropdown.

**Settings link updated:** Profile dropdown "Settings" now navigates to
`org-settings` instead of `settings`.

---

## Navigation flow

```
User lands on /dashboard (app context)
  → AppSidebar shows: Dashboard, Activity, Connections, etc.
  → Header shows: [Heimdall] / OrgName / AppName ▾

User clicks "OrgName" in header
  → Navigates to /org (org context)
  → OrgSidebar shows: Projects, Team, Billing, Settings
  → Header shows: [Heimdall] / OrgName

User clicks "Projects" card (Phase 4)
  → appStore.selectApp(appId)
  → Navigates to /dashboard (app context)
  → Sidebar switches back to AppSidebar
```

---

## Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean (1.54s) |
| `vitest run` | 52/52 tests pass |
| `go build ./...` | Clean |
| `go vet ./...` | Clean |

---

## What's next

Phase 4 fills in `OrgOverviewPage.vue` with the actual projects grid — app
cards with status, health indicators, search, and grid/list toggle.
