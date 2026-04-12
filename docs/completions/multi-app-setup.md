# Multi-App Setup & Settings

## Original brief

Currently, top left of the side nav bar is the APPLICATION selector in a dropdown.

I'd like to be able to add multiple apps there.

When clicking to add a new app, there should be an onboarding/set-up wizard that should be ultra-intuitive to navigate and help me set-up / configure a new application on Heimdall.

Certain steps should be skippable such as adding connectors (which we can add in the onboarding/set-up wizard, but also skip if we just want to create an empty App first).

At any point in time, I should be able to discard the app onboarding, which will discard the app creation.

Also, existing apps must be deletable somewhere intuitive — perhaps in a Settings page found somewhere, perhaps bottom of the left main side nav bar? Here I should be able to see my profile info (e.g. email and all that), as well as the Apps I've got, the linkages I've added for each app (e.g. to DB connections etc etc).

---

## Context: what already exists

A quick map of existing scaffolding so the plan below can build on it rather than reinvent:

| Area | Status | Location |
|---|---|---|
| Multi-app data model | ✅ Shipped (v0.10.0) | `backend/migrations/014_organizations_applications.up.sql` |
| `GET|POST /api/apps` | ✅ Shipped (v0.12.0) | `backend/internal/api/handlers/applications.go` |
| `DeleteApplication` SQL query | ✅ Exists | `backend/internal/db/queries/applications.sql:26` |
| `DELETE /api/apps/{id}` HTTP handler | ❌ **Missing** | — |
| App store `createApp` / `selectApp` | ✅ Shipped | `frontend/src/stores/app.ts:40-60` |
| Sidebar app selector (`BaseSelect`) | ✅ Shipped | `frontend/src/components/common/AppSidebar.vue:96-106` |
| "Add app" affordance in sidebar | ❌ **Missing** | — |
| Multi-step wizard pattern | ✅ Exists (reusable) | `frontend/src/components/connections/wizard/ConnectionWizard.vue` |
| First-time onboarding | ✅ Single-form only | `frontend/src/pages/OnboardingPage.vue` |
| Settings / Profile page | ❌ **Missing** — greenfield | — |
| `GET /api/auth/me` endpoint | ✅ Exists, unused by FE | `backend/internal/api/handlers/auth.go:10` |
| CASCADE deletes on app removal | ✅ All dependents cascade | `connections`, `app_agent_config`, `monitoring_state`, `notification_*`, `investigation_schedules` |

**Big takeaway:** the backend is ~90% ready. The bulk of this work is frontend — a new wizard component, a new Settings page, and wiring an "add app" entry point into the sidebar selector. The single backend gap is adding a `DELETE /api/apps/{id}` handler that wraps the already-existing SQL query.

## Design overview

Three user-visible surfaces:

1. **App Selector "+ New App"** — clicking it opens the new-app wizard as a fullscreen modal overlay (not a route). This keeps the user's current app/route context intact if they discard.
2. **New-App Wizard** — multi-step, skippable connector step, always-visible "Discard" button that rolls back the created app on cancel.
3. **Settings Page** — new `/settings` route, linked from the sidebar footer, containing Profile, Organisation, and Apps sections. Delete-app lives here.

### Wizard structure

Reusing the state-machine pattern from `ConnectionWizard.vue`:

```
Step 1: Name & description  (required)
  ↓
Step 2: Add a connector?    (skippable — "Add later" button)
  ↓  (if not skipped)
Step 2a: Pick platform      (reuses PlatformGrid from connection wizard)
Step 2b: Configure + test   (reuses connection-wizard steps entirely)
  ↓
Step 3: Confirmation / redirect to new app's dashboard
```

**Why create the app eagerly at step 1?** Two reasons. First, if the user adds a connector in step 2, it needs a valid `app_id` FK. Second, it lets us mirror the connection wizard's "orphan cleanup on discard" pattern — the backend stores the partial state, the frontend deletes it on cancel. The alternative (holding the whole wizard state in memory and POSTing atomically at the end) would mean forking the connection wizard instead of reusing it.

**Discard semantics:** "Discard" on any step calls `DELETE /api/apps/{appId}` (which cascades to any connection that step 2 may have created). Closing the browser tab is equivalent to "Discard" — the app is only committed when the user hits "Finish" on the last step. This makes browser-crash recovery impossible but matches the connection wizard's behaviour (simpler than introducing a `draft` state column).

### Settings page structure

```
/settings
  ├─ Profile         — email, name, created-at, (logout)
  ├─ Organisation    — org name, slug, member list (future)
  └─ Applications    — list of all apps, per-app:
                        • name, status, created-at
                        • connection count (linked to /connections page)
                        • schedule count (linked to /schedules page)
                        • "Delete app" button (confirmation modal)
```

Sidebar footer gets a "Settings" link above the existing org-name/email/logout block. The existing footer info stays where it is — we're adding a link, not moving the footer.

---

## Implementation phases

Four phases, each independently shippable. Phase 1 bundles all backend changes; phases 2 and 3 are the bulk of the work (frontend); phase 4 is polish.

### Phase 1 — Backend: delete-app handler + counts endpoint

**Goal:** expose the already-existing `DeleteApplication` query over HTTP, and add an opt-in `?include=counts` variant of `GET /api/apps` so the Settings page can fetch connection/schedule counts in a single request instead of N+1.

**Tasks:**

1. **`backend/internal/api/handlers/applications.go`** — add `DeleteApplication(w, r)` handler:
   - Extract `{appId}` from URL, parse as `uuid.UUID`.
   - Call `authorizeApp(ctx, q, userID, appID)` (existing helper) — returns 404 if the app isn't in the caller's org. This matches how every other `/api/apps/{appId}/*` handler guards access.
   - Enforce the **last-app guard**: count apps in the caller's org; if count == 1, return 409 with `{"error": "cannot delete last application", "code": "last_app"}`. The `code` field is important — the frontend uses it to render the exact reason (vs. a generic error toast).
   - Call `q.DeleteApplication(ctx, appID)`.
   - Return 204 on success.

2. **`backend/internal/api/router.go`** — mount `router.Delete("/{appId}", h.DeleteApplication)` alongside the existing `Get("/{appId}", …)` inside the `/apps` sub-router.

3. **`backend/internal/db/queries/applications.sql`** — add `ListApplicationsByOrgWithCounts`:
   ```sql
   -- name: ListApplicationsByOrgWithCounts :many
   SELECT
     a.*,
     (SELECT COUNT(*) FROM connections WHERE application_id = a.id) AS connection_count,
     (SELECT COUNT(*) FROM investigation_schedules WHERE application_id = a.id) AS schedule_count
   FROM applications a
   WHERE a.org_id = $1
   ORDER BY a.created_at DESC;
   ```
   Run `make sqlc-generate` to regenerate the Go wrapper. The subquery approach keeps the query cost linear in `#apps` (not `#apps × #connections`), which matters once users have many connections per app.

4. **`backend/internal/api/handlers/applications.go`** — modify `ListApplications` to inspect `?include=counts` and route to the new query when present. Response shape gains `connection_count` and `schedule_count` fields on each app *only* when the query param is set — keeps the default sidebar fetch cheap and avoids breaking existing callers.

5. **`backend/internal/api/handlers/applications_test.go`** — add test cases:
   - Delete happy path: create two apps, delete one, assert 204 + list returns only the remaining one.
   - Delete cross-org guard: user A cannot delete user B's app — 404.
   - Delete last-app guard: deleting the only app returns 409 with `code: "last_app"`.
   - Delete cascade: create an app → create a connection → delete the app → assert the connection row is gone. This codifies the CASCADE assumption so a future migration that accidentally drops it will fail a test rather than silently leak rows.
   - List with counts: create an app with 2 connections and 1 schedule, `GET /api/apps?include=counts`, assert the response includes `connection_count: 2` and `schedule_count: 1`.
   - List without counts (default): assert the response does **not** include count fields (default behaviour unchanged).

**Files changed:** 4 (`applications.go`, `router.go`, `applications.sql`, `applications_test.go`) + 1 regen (`applications.sql.go`). No migration.

**Why add counts now rather than in Phase 4?** You asked for the 80-20 call. Adding a second sqlc query with two subqueries is ~15 lines of SQL and ~20 lines of handler code. The alternative (ship N+1 first, refactor later) costs roughly the same work twice plus a frontend refactor to swap fetching strategy. Doing it upfront is strictly cheaper.

### Phase 2 — Frontend: "New App" wizard

**Goal:** clicking "+ New App" in the sidebar selector opens a fullscreen modal wizard that walks the user through naming + optional connector setup.

**Tasks:**

1. **`frontend/src/components/app-wizard/AppWizard.vue`** *(new)* — root wizard component, modelled on `ConnectionWizard.vue`. Holds `currentStep`, `draftAppId` (set after step 1), and a `discard()` method that calls `deleteApp(draftAppId)` if set, then closes the modal.

2. **`frontend/src/components/app-wizard/steps/StepAppDetails.vue`** *(new)* — first step. Fields: name (required), description (optional). On "Next", calls `app.createApp(name)` → captures returned `appId` into `draftAppId`. Validates name is non-empty via `@valid` emit to the wizard shell.

3. **`frontend/src/components/app-wizard/steps/StepConnectorChoice.vue`** *(new)* — second step, two large buttons: "Add a connector" and "Skip for now". Skip advances to step 3; Add advances into the reused connection-wizard flow.

4. **Reuse of `ConnectionWizard`** — the connector sub-flow is the hard part. Two options:
   - **(A) Embed `ConnectionWizard` directly** by passing `draftAppId` as a prop and a `@finished` callback. Requires `ConnectionWizard.vue` to accept an injected `appId` instead of reading it from `useAppStore()`. Clean but touches existing code.
   - **(B) Extract the step-runner portion of `ConnectionWizard` into a `<ConnectionWizardCore :appId=…>` subcomponent** and use it from both the existing connection page and the new app wizard. Cleaner long-term, larger diff.
   - **Recommendation: (A)** — smaller blast radius, and the `appId` injection is a reasonable refactor regardless. If embedding reveals deeper coupling, escalate to (B).

5. **`frontend/src/components/app-wizard/steps/StepConfirm.vue`** *(new)* — "You're all set. Go to dashboard." Calls `app.selectApp(draftAppId)` + router.push('/dashboard') on finish. Clears `draftAppId` so `discard()` becomes a no-op.

6. **`frontend/src/components/common/AppSidebar.vue`** — modify the `BaseSelect` block to add a "+ New App" row as the **last entry** in the dropdown, beneath all existing apps. Implementation:
   - Append `{ value: '__new_app__', label: '+ New application' }` to the `appOptions` computed property after mapping the user's apps.
   - In `handleSelectApp`, intercept the sentinel: if `value === '__new_app__'`, open the `AppWizard` modal via a local `wizardOpen` ref **without** calling `selectApp` (so the user's current app context is preserved if they discard).
   - Check `BaseSelect`'s implementation to ensure the sentinel option can be visually distinguished (e.g. via a divider above it or custom styling). If `BaseSelect` doesn't support option groups, it's fine to rely on the `+` prefix — users parse that cue quickly.
   - **Note on `BaseSelect` coupling:** if the component resets its internal selection state to the chosen value after an option is clicked, intercept the event *before* it commits so `currentAppId` isn't momentarily set to `__new_app__`. Inspect `BaseSelect.vue` before writing the handler to confirm the event flow.

7. **`frontend/src/stores/app.ts`** — add a `deleteApp(appId: string)` action that calls `DELETE /api/apps/{appId}`, removes from `applications`, and if `currentAppId === appId`, calls `selectApp(applications[0].id)`. Used by both the wizard's discard path and the Settings page.

8. **Discard confirmation modal** — if the user has progressed past step 1 (i.e. `draftAppId` is set), show a "Are you sure? This will discard the new app." confirmation. If still on step 1, close immediately (nothing to roll back).

**Files changed:** ~6 new, 2 edited (`AppSidebar.vue`, `stores/app.ts`).

### Phase 3 — Frontend: Settings page

**Goal:** new `/settings` route, accessible from the sidebar footer, containing Profile / Organisation / Applications sections.

**Tasks:**

1. **`frontend/src/pages/SettingsPage.vue`** *(new)* — page shell with three section components stacked. No tabs — vertical scroll is simpler and the sections are small.

2. **`frontend/src/components/settings/ProfileSection.vue`** *(new)* — renders `auth.user.email` and `auth.user.created_at`. Includes "Sign out" button (mirrors the footer one, gives users an obvious place to find account actions).

3. **`frontend/src/components/settings/OrganisationSection.vue`** *(new)* — renders `app.organization.name`, `app.organization.slug`, and the member count. Member management is out of scope for this feature — just show the data.

4. **`frontend/src/components/settings/ApplicationsSection.vue`** *(new)* — fetches `GET /api/apps?include=counts` on mount (single request, no N+1) and lists the result. Per-row:
   - Name, status badge, created-at
   - `connection_count` with a click-through link to `/connections` (pre-filtered by appId via `selectApp`)
   - `schedule_count` with a click-through link to `/schedules` (same pattern)
   - **"Delete" button with last-app handling:** if `applications.length === 1`, the button is rendered **disabled** with an inline tooltip / helper text explaining *"Your organisation must have at least one application. Create another app before deleting this one."* This matches the backend 409 guard. The disabled state must be visually obvious, not just a greyed button — use the existing disabled styling plus an explicit helper line below the button.
   - A "+ New Application" button at the bottom that opens the same `AppWizard` component from Phase 2.

5. **`frontend/src/components/settings/DeleteAppModal.vue`** *(new)* — **two-stage UX**:
   - **Stage 1 — Confirmation:** Typed-to-confirm pattern. Modal shows an explicit, itemised preview of what will be deleted: *"This will permanently delete **{app name}** and all of its data: **{n} connections**, **{m} schedules**, **all logs**, **all reports**, and **all conversation history**. This cannot be undone."* Delete button stays disabled until the user types the exact app name. The counts come from the already-fetched `?include=counts` payload — no extra request.
   - **Stage 2 — Post-delete confirmation:** After `app.deleteApp(appId)` resolves successfully, the modal transitions to a success state showing *"**{app name}** has been deleted."* with a single "Close" button. This gives the user clear, explicit feedback that the destructive action completed. The modal then closes and the list refreshes.
   - **Error handling:** if the backend returns 409 with `code: "last_app"` (defensive — the button should already be disabled), surface the backend error text inline in the modal rather than closing it. Any other error shows a generic "Failed to delete — please try again" with the raw error message in a collapsible details block.

6. **`frontend/src/router/index.ts`** — add `{ path: '/settings', component: () => import('@/pages/SettingsPage.vue'), meta: { requiresAuth: true } }`. Lazy-loaded, gated by the same guard as other app routes.

7. **`frontend/src/components/common/AppSidebar.vue`** — add a "Settings" link in the footer block, above the org/email display. Uses the existing nav-link styling.

**Files changed:** ~6 new, 2 edited (`router/index.ts`, `AppSidebar.vue`).

### Phase 4 — Polish & hardening

These are nice-to-haves that don't block the core feature.

1. **Keyboard support** — Esc closes the wizard (with discard confirmation), Enter advances. Matches existing modal conventions.
2. **Toast notifications** — non-blocking success toasts on create (the DeleteAppModal already handles its own post-delete confirmation in-place, so it doesn't need a toast). Find the existing toast system first; don't add one if none exists.
3. **Telemetry** — log app-create and app-delete events to the Activity feed (similar to how connection changes are logged). Lets users see "app created" entries in their own audit trail.
4. **Update vision.md** — the Platform Sections list currently implies a single-app experience. Clarify that apps are first-class and that Heimdall is multi-tenant at the app level within an org.

---

## Risks & trade-offs

**Eager app creation vs deferred.** We create the app at step 1 and clean up on discard. Risk: a user closing their browser mid-wizard leaves an orphaned empty app. Mitigation: the app is empty (no connections, no logs), and the Settings page makes it trivial to delete. **Acceptable.** Alternative — a `draft` column with a TTL sweep — is over-engineered for the symptom.

**The "last app" guard.** Blocking deletion of the only remaining app is paternalistic but prevents the sidebar selector from being empty. The alternative (allowing deletion + redirecting to onboarding) is more flexible but means the user can delete themselves into a first-run state, which is confusing. **Erring toward safety; revisit if anyone complains.**

**Reusing `ConnectionWizard` (Phase 2, task 4).** If the connection wizard reads from `useAppStore()` deep inside its steps, option (A) might require more plumbing than expected. I'd like to confirm the coupling depth before committing. See Question 3 below.

**No undo for app deletion.** Deleting an app cascade-deletes everything beneath it. There's no soft-delete / trash / recover flow. The typed-to-confirm modal is the only guardrail. If this becomes a problem we can add a `deleted_at` column and a sweep-after-N-days pattern — but this is a meaningful scope expansion.

---

## Decisions (resolved)

Original open questions have been answered. Recorded here so implementation follows the agreed path.

1. **Last-app deletion → BLOCKED.** Backend returns 409 with `code: "last_app"`; frontend renders the Delete button as disabled with explicit helper text when it's the only app. No redirect-to-onboarding fallback.
2. **Delete confirmation UX → typed-to-confirm + explicit post-delete confirmation.** The `DeleteAppModal` is two-stage: typed-to-confirm stage, then a success-state stage after the API call resolves, before closing. Intuitive and explicit both before *and* after the destructive action.
3. **`ConnectionWizard.vue` refactor → APPROVED.** The component will accept an optional `appId` prop (defaulting to `useAppStore().currentAppId` for backwards compatibility) so the new `AppWizard` can embed it directly for the optional connector step. No need to extract a `ConnectionWizardCore` subcomponent.
4. **"+ New App" affordance → dropdown last option.** Appended as a sentinel `{ value: '__new_app__', label: '+ New application' }` row beneath all existing apps in the `BaseSelect`. Click is intercepted before `currentAppId` is mutated, opening the `AppWizard` modal.
5. **Settings route granularity → single `/settings` page with sections.** No sub-routes; vertical scroll. Sub-routes can be added later if the surface grows.
6. **Counts endpoint → shipped in Phase 1.** Adding `?include=counts` upfront costs roughly the same as shipping N+1 then refactoring later, so it's pulled forward into the backend phase. Applied the 80-20 rule: small additional effort for a meaningfully better architecture.
7. **Onboarding vs. new-app wizard → two separate components.** `OnboardingPage.vue` stays as-is for first-run; the new `AppWizard` handles all subsequent app creation. Unifying them is an option for a later refactor pass, not this one.
8. **Wizard surface → fullscreen modal overlay.** Not a route. Preserves the user's current app/route context if they discard mid-wizard.

---

## Ready-to-execute summary

The plan is fully specified. Implementation order:

1. **Phase 1 (backend)** — `DELETE /api/apps/{id}` with last-app guard, `?include=counts` on `GET /api/apps`, full test coverage including cascade verification. No migration, no sqlc schema changes (just a new query). Shippable standalone.
2. **Phase 2 (frontend — wizard)** — `AppWizard` modal, sidebar dropdown sentinel, `ConnectionWizard` appId-prop refactor, `app.deleteApp()` store action, discard-rollback flow.
3. **Phase 3 (frontend — settings)** — `/settings` route, sidebar footer link, three sections (Profile / Organisation / Applications), two-stage `DeleteAppModal`, last-app disabled state.
4. **Phase 4 (polish)** — keyboard handling, toasts for create, activity-feed telemetry, vision.md update.

All decisions locked. Awaiting green light to start Phase 1.
