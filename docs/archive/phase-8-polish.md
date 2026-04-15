# Phase 8 — Polish & Migration from Old Settings

**Status:** Complete
**Date:** 2026-04-13
**Plan:** `docs/executing/org-overview-implementation.md`

---

## What changed

Final cleanup pass: removed the old `/settings` page and its sub-components,
polished the sidebar context-switch transition, refined header breadcrumb
behaviour, and ensured mobile navigation works correctly across contexts.

---

## Deleted files

| File | Reason |
|------|--------|
| `pages/SettingsPage.vue` | Replaced by `/org/settings`, `/org/team`, and `/org` |
| `components/settings/ProfileSection.vue` | Profile info available in header dropdown |
| `components/settings/OrganisationSection.vue` | Replaced by `OrgSettingsPage.vue` |
| `components/settings/ApplicationsSection.vue` | Replaced by `OrgOverviewPage.vue` |

**Kept:** `components/settings/DeleteAppModal.vue` — reusable component for
app deletion, will be used by the org overview page's three-dot menu in the future.

---

## Modified files

### `layouts/DefaultLayout.vue`

**Sidebar crossfade transition:**
Desktop sidebar swap wrapped in `<Transition name="sidebar" mode="out-in">`
with `key="org"` / `key="app"` to trigger the transition. CSS crossfade at
150ms ease — fast enough to feel responsive, slow enough to be visible.

**Mobile menu auto-close:**
Added `watch(isOrgContext, () => { mobileMenuOpen.value = false })` so the
mobile overlay closes automatically when context switches (e.g. user taps an
app card on the overview page).

### `components/common/AppHeader.vue`

**Context-aware logo link:**
Logo now links to `/org` when in org context, `/dashboard` when in app context.
Previously it always linked to `/dashboard`, which felt wrong when you're on
the org overview — you'd expect the logo to take you "home" for the current
context.

**Profile dropdown expanded:**
Added "Organisation" link that navigates to `/org` (overview page). Dropdown
now shows: Organisation, Settings, Sign out.

### `router/index.ts`

Cleaned up the `/settings` redirect comment (removed reference to preserved
`SettingsPage.vue` since it's now deleted).

---

## Sidebar transition CSS

```css
.sidebar-enter-active,
.sidebar-leave-active {
  transition: opacity 0.15s ease;
}
.sidebar-enter-from,
.sidebar-leave-to {
  opacity: 0;
}
```

The `mode="out-in"` ensures the old sidebar fades out before the new one fades
in, preventing a flash of both sidebars overlapping. The 150ms duration matches
the dropdown transition timing used throughout the app.

---

## Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean (1.61s) |
| `vitest run` | 52/52 tests pass |
| `go build ./...` | Clean |
| `go vet ./...` | Clean |

---

## Full feature summary (Phases 1-8)

The organisation overview and context-switching feature is complete. Here's
what was built across all 8 phases:

### Database
- `org_members` join table with roles (owner/admin/member)
- Migrated from `users.org_id` (1:1) to M:N membership
- 8 RLS policies rewritten to use `org_members`
- `DeleteOrganization` cascade query

### Backend API (9 new endpoints)
| Endpoint | Purpose |
|----------|---------|
| `GET /api/orgs` | List all user's orgs with roles |
| `POST /api/orgs` | Create new org |
| `PUT /api/org` | Update org name/slug |
| `DELETE /api/org` | Delete org (with slug confirmation) |
| `GET /api/org/members` | List org members |
| `POST /api/org/members/invite` | Invite member by email |
| `PUT /api/org/members/{userId}/role` | Change member role |
| `DELETE /api/org/members/{userId}` | Remove member |
| `GET /api/org` (updated) | Now includes caller's role |

### Frontend pages (4 new)
| Page | Route | Content |
|------|-------|---------|
| Org Overview | `/org` | Projects grid with search, sort, grid/list toggle |
| Team | `/org/team` | Member list, invite, role management |
| Org Settings | `/org/settings` | Editable name/slug, danger zone |
| Billing | `/org/billing` | Coming soon placeholder |

### Navigation
- Two-context sidebar (OrgSidebar + AppSidebar) with crossfade transition
- Org switcher dropdown in header with role badges
- Create org modal with auto-slug
- Context-aware logo link and profile dropdown
- `/settings` redirects to `/org/settings`
- Mobile nav auto-closes on context switch
