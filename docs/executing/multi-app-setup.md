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

Four phases, each independently shippable. Phase 1 is the smallest backend change; phases 2 and 3 are the bulk of the work; phase 4 is polish.

### Phase 1 — Backend: delete-app handler

**Goal:** expose the already-existing `DeleteApplication` query over HTTP, scoped to the caller's org.

**Tasks:**

1. **`backend/internal/api/handlers/applications.go`** — add `DeleteApplication(w, r)` handler:
   - Extract `{appId}` from URL, parse as `uuid.UUID`.
   - Call `authorizeApp(ctx, q, userID, appID)` (existing helper) — returns 404 if the app isn't in the caller's org. This matches how every other `/api/apps/{appId}/*` handler guards access.
   - Call `q.DeleteApplication(ctx, appID)`.
   - Handle the edge case: if this is the user's **last** app, return 409 with `{"error": "cannot delete last application"}`. This prevents users from orphaning themselves — they'd have no current-app and the sidebar selector would break. Alternative: allow it and redirect to onboarding. **I lean toward 409** because it's safer and easier to undo.
   - Return 204 on success.

2. **`backend/internal/api/router.go`** — mount `router.Delete("/{appId}", h.DeleteApplication)` alongside the existing `Get("/{appId}", …)` inside the `/apps` sub-router.

3. **`backend/internal/api/handlers/applications_test.go`** — add three test cases:
   - Happy path: create two apps, delete one, assert 204 + list returns only the remaining one.
   - Cross-org guard: user A cannot delete user B's app — 404.
   - Last-app guard: deleting the only app returns 409.

4. **Smoke test cascade behaviour**: in a test, create an app → create a connection → delete the app → assert the connection row is gone. This codifies the CASCADE assumption so a future migration that accidentally drops it will fail a test rather than silently leak rows.

**Files changed:** 3 (`applications.go`, `router.go`, `applications_test.go`). No migration, no sqlc regen.

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

6. **`frontend/src/components/common/AppSidebar.vue`** — modify the `BaseSelect` block to add a "+ New App" row. Two implementation options:
   - **(A)** Add a synthetic `{ value: '__new__', label: '+ New application' }` option to `appOptions`. Intercept in `handleSelectApp`: if value === `'__new__'`, open the wizard instead of calling `selectApp`. Fast but abuses the select component.
   - **(B)** Add a small `+` icon button next to the select. Cleaner UX, more markup.
   - **Recommendation: (B)** — the dropdown is already narrow and a "+" icon button is the industry-standard affordance. Similar to how most IDEs separate "switch workspace" from "new workspace".

7. **`frontend/src/stores/app.ts`** — add a `deleteApp(appId: string)` action that calls `DELETE /api/apps/{appId}`, removes from `applications`, and if `currentAppId === appId`, calls `selectApp(applications[0].id)`. Used by both the wizard's discard path and the Settings page.

8. **Discard confirmation modal** — if the user has progressed past step 1 (i.e. `draftAppId` is set), show a "Are you sure? This will discard the new app." confirmation. If still on step 1, close immediately (nothing to roll back).

**Files changed:** ~6 new, 2 edited (`AppSidebar.vue`, `stores/app.ts`).

### Phase 3 — Frontend: Settings page

**Goal:** new `/settings` route, accessible from the sidebar footer, containing Profile / Organisation / Applications sections.

**Tasks:**

1. **`frontend/src/pages/SettingsPage.vue`** *(new)* — page shell with three section components stacked. No tabs — vertical scroll is simpler and the sections are small.

2. **`frontend/src/components/settings/ProfileSection.vue`** *(new)* — renders `auth.user.email` and `auth.user.created_at`. Includes "Sign out" button (mirrors the footer one, gives users an obvious place to find account actions).

3. **`frontend/src/components/settings/OrganisationSection.vue`** *(new)* — renders `app.organization.name`, `app.organization.slug`, and the member count. Member management is out of scope for this feature — just show the data.

4. **`frontend/src/components/settings/ApplicationsSection.vue`** *(new)* — lists `app.applications`. Per-row:
   - Name, status badge, created-at
   - Connection count — fetched from existing `GET /api/apps/{appId}/connections` (already wired, just call it per row)
   - Schedule count — fetched from existing `GET /api/apps/{appId}/schedules`
   - "Delete" button → opens confirmation modal showing **exactly what will be deleted** ("This will permanently delete *My App*, its 3 connections, 2 schedules, and all associated logs and reports. This cannot be undone.")
   - A "+ New Application" button at the bottom that opens the same `AppWizard` component from Phase 2.

5. **`frontend/src/components/settings/DeleteAppModal.vue`** *(new)* — typed-to-confirm pattern (user must type the app name to enable the delete button). Common UX for destructive org-level actions. Calls `app.deleteApp(appId)` on confirm.

6. **`frontend/src/router/index.ts`** — add `{ path: '/settings', component: () => import('@/pages/SettingsPage.vue'), meta: { requiresAuth: true } }`. Lazy-loaded, gated by the same guard as other app routes.

7. **`frontend/src/components/common/AppSidebar.vue`** — add a "Settings" link in the footer block, above the org/email display. Uses the existing nav-link styling.

**Open question baked into this phase:** connection & schedule counts on every page render means N+1 API calls where N = number of apps. With a typical user on 1-5 apps this is fine. If we expect more, add a `GET /api/apps?include=counts` query param later. **Not blocking — flag for Phase 4.**

**Files changed:** ~6 new, 2 edited (`router/index.ts`, `AppSidebar.vue`).

### Phase 4 — Polish & hardening

These are nice-to-haves that don't block the core feature.

1. **Keyboard support** — Esc closes the wizard (with discard confirmation), Enter advances. Matches existing modal conventions.
2. **Empty state** — if Applications section has 0 apps (only possible if we relax the "cannot delete last app" rule), show a prominent "Create your first app" CTA. If we keep the last-app guard, this is dead code.
3. **Toast notifications** — success toasts on create/delete. Find the existing toast system first; don't add one if none exists.
4. **Batch `?include=counts`** for the Applications section if the N+1 becomes real.
5. **Telemetry** — log app-create and app-delete events to the Activity feed (similar to how connection changes are logged). Lets users see "app created" entries in their own audit trail.
6. **Update vision.md** — the Platform Sections list currently implies a single-app experience. Clarify that apps are first-class and that Heimdall is multi-tenant at the app level within an org.

---

## Risks & trade-offs

**Eager app creation vs deferred.** We create the app at step 1 and clean up on discard. Risk: a user closing their browser mid-wizard leaves an orphaned empty app. Mitigation: the app is empty (no connections, no logs), and the Settings page makes it trivial to delete. **Acceptable.** Alternative — a `draft` column with a TTL sweep — is over-engineered for the symptom.

**The "last app" guard.** Blocking deletion of the only remaining app is paternalistic but prevents the sidebar selector from being empty. The alternative (allowing deletion + redirecting to onboarding) is more flexible but means the user can delete themselves into a first-run state, which is confusing. **Erring toward safety; revisit if anyone complains.**

**Reusing `ConnectionWizard` (Phase 2, task 4).** If the connection wizard reads from `useAppStore()` deep inside its steps, option (A) might require more plumbing than expected. I'd like to confirm the coupling depth before committing. See Question 3 below.

**No undo for app deletion.** Deleting an app cascade-deletes everything beneath it. There's no soft-delete / trash / recover flow. The typed-to-confirm modal is the only guardrail. If this becomes a problem we can add a `deleted_at` column and a sweep-after-N-days pattern — but this is a meaningful scope expansion.

---

## Questions for review

1. **Last-app deletion.** Do you want to block deletion of the user's only app (my recommendation, returns 409), or allow it and redirect to the onboarding flow? The former is safer; the latter is more flexible.

2. **App deletion confirmation UX.** Is "type the app name to confirm" the right level of friction, or is a simple "Yes, delete" button fine? I lean toward typed-to-confirm given cascade behaviour.

3. **Connection wizard coupling.** Are you OK with me refactoring `ConnectionWizard.vue` to accept an `appId` prop (defaulting to `useAppStore().currentAppId` for backwards compatibility) so the new app wizard can embed it? This is the crux of Phase 2 reuse — the alternative is duplicating the connection-wizard flow, which I want to avoid.

4. **"+ New App" affordance placement.** Dropdown-inline option (A) vs separate "+" icon button (B)? I recommended (B) but the sidebar is tight on horizontal space — worth a gut check from you.

5. **Settings route granularity.** Single `/settings` page with sections (my plan), or split into `/settings/profile`, `/settings/apps`, `/settings/org` sub-routes? Sub-routes are more structured but overkill for the current data volume.

6. **Connection/schedule counts on the Settings page.** Is N+1 fetching acceptable for v1 (1-5 apps per user typical), or should I add `?include=counts` to `GET /api/apps` upfront?

7. **Onboarding vs. new-app wizard — one component or two?** The existing `OnboardingPage.vue` is a single-form flow creating org + first app. The new wizard creates apps within an existing org. These have *mostly* different shapes but overlapping steps. Do you want me to:
   - (a) Leave `OnboardingPage.vue` alone and build a separate wizard for adding subsequent apps (my current plan), or
   - (b) Replace `OnboardingPage.vue` with the new wizard, conditionally showing an "org setup" step when the user has no org yet?
   - (a) is safer; (b) is more unified but risks destabilising the first-run flow.

8. **Wizard as modal vs. route.** I'm proposing a fullscreen modal overlay rather than a `/apps/new` route so the user's current app context isn't lost if they discard. Do you prefer a route-based wizard instead (cleaner URLs, bookmarkable, but loses context on cancel)?
