# Multi-App Setup — Phase 3 Completion Notes

**Scope:** New `/settings` page with three sections (Profile / Organisation / Applications), a two-stage `DeleteAppModal`, and a Settings link in the sidebar footer. Closes the loop on the multi-app feature: users can now manage their entire application list — view, create, and delete — from a single dedicated page instead of relying solely on the sidebar selector.

**Branch:** `master`
**Validation:** `vue-tsc --noEmit` clean, `npm run test -- --run` 35/35 (no regressions), `npx eslint <touched files>` clean, `npm run build` clean (1.10s). `SettingsPage` lazy-splits into its own 13.31 kB chunk (3.96 kB gzipped).

---

## What shipped

### 1. Types & API client

**Files:** `frontend/src/types/organization.ts`, `frontend/src/api/applications.ts`

- **`ApplicationWithCounts`** — new interface extending `Application` with `connection_count: number` and `schedule_count: number`, mirroring the backend's `ListApplicationsByOrgWithCountsRow` struct from Phase 1.
- **`listApplicationsWithCounts()`** — thin axios wrapper that hits `GET /api/apps?include=counts`. The existing `listApplications()` is untouched so lean consumers (sidebar selector, initial store hydration) keep their O(1) payload. Only the Settings page opts into the enriched shape.

**Why a separate function instead of a query-param overload?** A separate function gives each call site a discriminated return type. Hiding the `include=counts` branch behind a single `listApplications(opts?)` signature would force every caller to type-narrow the response, which is a worse ergonomic trade-off than just shipping two named functions.

### 2. `ProfileSection.vue` — read-only account identity

**File:** `frontend/src/components/settings/ProfileSection.vue` *(new)*

Renders `auth.user.email` and `auth.user.created_at` from the existing Supabase session — **no separate `/auth/me` fetch is required** because the session already carries both fields. Adding a roundtrip here would waste latency for data we already have in memory.

Includes a **duplicated "Sign out" button**. The sidebar footer already has one, but users searching for "account actions" naturally visit Settings first, and the sidebar footer version is easy to miss inside the mobile drawer. The teardown logic mirrors the sidebar (`app.reset()` before `auth.logout()`) so the next login lands on a clean state.

### 3. `OrganisationSection.vue` — read-only org summary

**File:** `frontend/src/components/settings/OrganisationSection.vue` *(new)*

Renders `app.organization.name`, `slug`, and `created_at`.

**Intentional plan deviation:** the plan mentions a member count. **I explicitly skipped it.** Rationale: there's no members table in the data model, no backend endpoint for org members, and the current onboarding flow enforces a one-user-per-org invariant — a member count would always display `1`. Surfacing that would be misleading, and adding a real members endpoint is a meaningfully larger feature that belongs in its own PR. An inline comment in the component explains the decision so the next reader doesn't wonder where the missing field went.

When multi-user orgs eventually land, this section is the right place to surface an invite button, a member list, and role management — keeping everything org-related in one visual unit.

### 4. `ApplicationsSection.vue` — the interaction core

**File:** `frontend/src/components/settings/ApplicationsSection.vue` *(new)*

Fetches `GET /api/apps?include=counts` on mount and stores the result in a **local ref**, not in the app store. This is a deliberate boundary:

- The app store's `applications` list is the lean variant that every other consumer (sidebar selector, dashboard header, etc.) uses.
- Pushing the enriched shape into the store would risk accidentally coupling the sidebar selector to a heavier payload, *and* would force me to reconcile two shapes at every mutation site.
- The Settings page is the only surface that needs counts, and it's refetched on every mutation anyway, so local state is strictly simpler.

**Per-row UI:**

- **Identity:** name, status badge (reused `<StatusBadge>`), formatted `created_at`.
- **Click-through counts:** "N connections" and "M schedules" are buttons that select the target app *and* route to the corresponding per-app page. Selecting first ensures `/connections` and `/schedules` render data for the row the user actually clicked, not whatever was selected before. This is the smallest possible "deep-link with context" affordance — no URL-level filtering required.
- **Delete button with last-app disabled state:** rendered disabled when `rows.length <= 1`, with **explicit helper text beneath the button**: *"Your organisation must have at least one application. Create another first."* The helper line renders *only* when the button is disabled for this specific reason, so users get a targeted explanation rather than hunting for a tooltip.

**Why local helper text instead of a `title` tooltip?** Tooltips are invisible on touch, easy to miss with the keyboard, and require hovering — the disabled state has to be visually obvious, not discoverable. The plan explicitly called this out (*"The disabled state must be visually obvious, not just a greyed button"*), so I went with a visible text line.

**"+ New application" button:** opens the same `<AppWizard>` component from Phase 2, no duplication. On wizard close the section refetches unconditionally — the wizard may or may not have created a new row (user can discard or finish), and refetching is cheap enough to stay source-of-truth-agnostic.

**List refresh strategy:** on successful delete, the section refetches the whole list rather than surgically splicing the deleted row out. The refetch is one request for < 10 rows; the splice approach would save ~50ms on average but risks UI drift if server-side state changed while the modal was open (e.g. another tab). Simpler is strictly better here.

### 5. `DeleteAppModal.vue` — two-stage destructive action

**File:** `frontend/src/components/settings/DeleteAppModal.vue` *(new)*

Two internal stages driven by a `stage: 'confirm' | 'deleted'` ref:

**Stage 1 — Confirmation.** Typed-to-confirm pattern. The modal:

1. Shows an **itemised list** of what will be cascade-deleted: `{n} connections`, `{m} schedules`, all logs, all reports, all conversation history. The connection/schedule counts come straight from the already-fetched `ApplicationWithCounts` row — no extra request.
2. Requires the user to **type the exact app name** (case-sensitive, no trimming) before the Delete button enables. The strictness is intentional: being lenient on whitespace or case defeats the purpose of the safeguard, which is to force the user to *read* the app name they're about to destroy.
3. **Auto-focuses** the input on mount via `nextTick + useTemplateRef`, so keyboard-first users can start typing immediately.
4. **Resets the input** via a `watch` on `props.application.id` in case the parent swaps to a different row mid-flight (unlikely but defensive — the input shouldn't hold stale content against the new name).

**Stage 2 — Post-delete acknowledgement.** After `app.deleteApp(appId)` resolves successfully, the stage flips and the modal renders a **calm success screen** with a single "Close" button. The user sees a second, explicit confirmation that the destructive action actually landed. On close the parent emits `@deleted`, which triggers the list refetch.

**Error handling:** the `try/catch` around the delete surfaces backend error text **inline in the modal** rather than closing it. Specifically:

- **409 `last_app`** (defensive — the parent button should already be disabled): shows the server's message directly (`"cannot delete last application"`) so the user sees the real reason.
- **Any other error**: falls through to `"Failed to delete application. Please try again."` as a last-resort fallback.

The stage stays at `confirm` on any error, so the user can cancel out without losing context. Closing is blocked during the `deleting` state — no backdrop-click escape mid-request — so we don't fire `@close` against a promise that's still in flight.

**Esc-key support:** `handleClose` is wired to Esc, which respects the `deleting` guard. Clicking the backdrop triggers the same path.

### 6. `SettingsPage.vue` — thin shell

**File:** `frontend/src/pages/SettingsPage.vue` *(new)*

A flat vertical stack of the three sections wrapped in `max-w-3xl`. No tabs, no sub-routes — the surface is small enough that a single scroll is the clearest UX. The header matches the existing page-header pattern used by `SchedulesPage` and the other inner pages for visual consistency.

If the Settings surface grows later (e.g. Billing, API keys, Webhooks), the natural next step is in-page anchor navigation rather than sub-routes. Sub-routes add URL churn and require a per-section `RouterView` wiring that isn't worth it at this scale.

### 7. Route & sidebar footer link

**Files:** `frontend/src/router/index.ts`, `frontend/src/components/common/AppSidebar.vue`

- **Route:** `{ path: '/settings', name: 'settings', component: () => import('@/pages/SettingsPage.vue') }` — lazy-loaded, covered by the existing `auth.isAuthenticated` guard. No new route-guard logic; the global `router.beforeEach` already handles the auth gate for every non-public route.
- **Sidebar footer link:** a new `<RouterLink>` placed **above** the org/email/sign-out block in the footer. Styled as a small uppercase text link matching the footer's existing visual weight, with an active-state check (`isActive('settings')`) so the link gets the same "I'm on this page" treatment as the rest of the nav.

**Why footer placement instead of a nav section?** The main nav above is per-app context (Dashboard, Connections, Agent, etc.) — it renders data scoped to `currentAppId`. Settings is org/account-scoped, not per-app. Grouping it visually with the org-name/email/sign-out block signals that distinction to the user at a glance.

---

## Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/types/organization.ts` | Edit | +`ApplicationWithCounts` interface |
| `frontend/src/api/applications.ts` | Edit | +`listApplicationsWithCounts()` |
| `frontend/src/components/settings/ProfileSection.vue` | **New** | Email, joined date, sign-out button |
| `frontend/src/components/settings/OrganisationSection.vue` | **New** | Org name, slug, created date |
| `frontend/src/components/settings/ApplicationsSection.vue` | **New** | App list with counts, click-throughs, delete, + new app |
| `frontend/src/components/settings/DeleteAppModal.vue` | **New** | Two-stage typed-to-confirm modal |
| `frontend/src/pages/SettingsPage.vue` | **New** | Page shell stacking the three sections |
| `frontend/src/router/index.ts` | Edit | +`/settings` route |
| `frontend/src/components/common/AppSidebar.vue` | Edit | +Settings RouterLink in footer |

---

## Validation

```bash
cd frontend
npx vue-tsc -b --noEmit                    # clean
npm run test -- --run                      # 35/35 pass (no regressions)
npx eslint <touched files>                 # clean
npm run build                              # clean, 1.10s
                                           # SettingsPage chunk: 13.31 kB (3.96 kB gzipped)
```

**Not validated in this pass:** real click-through of the full delete flow against a live backend. The unit test coverage from Phase 2 (`app.test.ts`) already locks down `deleteApp`'s store contract, and the DeleteAppModal wraps that contract in a thin UI layer — but the actual "type the name → click delete → watch the row vanish → see the success screen → close → list refreshes" sequence is only meaningful as a manual browser test. Recommended once a staging environment is available.

**Skipped plan item:** the org-member count. There's no members table or endpoint to read from, and showing a hardcoded `1` would be misleading. Documented inline in `OrganisationSection.vue` with a comment explaining the deferral rationale. When multi-user orgs ship, the count belongs next to an invite button and a member list, not as a lonely integer.

---

## What Phase 3 does *not* do

Intentional scope boundaries — deferred to Phase 4 polish or later:

- **No keyboard shortcut for opening Settings.** Users click the sidebar link; there's no `Cmd+,` binding. Matches the existing repo convention (no global shortcuts defined anywhere).
- **No toast feedback on delete.** The `DeleteAppModal` handles its own post-delete acknowledgement in-place, which is both more explicit *and* more satisfying than a toast that slides off after 3 seconds. The plan specifically noted this.
- **No audit log / activity feed entry for app creation or deletion.** That's Phase 4 telemetry work.
- **No soft-delete / restore flow.** Delete cascade-deletes everything. The two-stage typed-to-confirm is the only guardrail — see the plan's risk section for why we accept this.
- **No settings sub-routes.** Single page, vertical scroll. In-page anchor navigation can be added later if the surface grows.

---

## Multi-app setup feature status

With Phase 3 shipped, the end-to-end user journey is complete:

1. **Phase 1 (backend)** ✅ `DELETE /api/apps/{id}` with last-app guard, `?include=counts` variant of the list endpoint, full test coverage including cascade verification.
2. **Phase 2 (frontend wizard)** ✅ Sidebar "+ New application" sentinel → `AppWizard` modal → eager-create with discard-rollback → `ConnectionWizard` refactored to accept an `appId` prop so the optional connector step works with draft apps.
3. **Phase 3 (frontend settings)** ✅ `/settings` page → three sections → two-stage `DeleteAppModal` → full CRUD reach from one place.
4. **Phase 4 (polish)** — still open: keyboard niceties, activity-feed telemetry for create/delete, `vision.md` update to reflect first-class multi-app semantics.

The core feature is shippable as-is. Phase 4 is nice-to-have.
