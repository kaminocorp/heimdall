# Multi-App Setup — Phase 2 Completion Notes

**Scope:** Frontend wizard for creating a new application, accessible from the sidebar app selector. Builds directly on the Phase 1 backend (`DELETE /api/apps/{appId}`, last-app guard, cascade deletes). This is the first phase that's user-visible — clicking the sidebar selector now offers a "+ New application" row that opens a multi-step wizard with a clean discard-and-rollback flow.

**Branch:** `master`
**Validation:** `vue-tsc --noEmit` clean, `npm run test -- --run` 35/35 green (6 new), `npx eslint <touched files>` clean, `npm run build` clean.

---

## What shipped

### 1. Store layer — `app.ts`

**File:** `frontend/src/stores/app.ts`

Two changes and one dead-code cleanup:

1. **`createApp(name, { select } = { select: true })` — new optional flag.** The default preserves every existing caller's behaviour (onboarding, "quick create" paths) by auto-selecting the new app. The AppWizard passes `{ select: false }` because it creates the app *eagerly at step 1* and can't afford to mutate `currentAppId` until the user actually commits at the final step — if we auto-selected the draft and the user then discarded, their previously-active app would be silently replaced.

2. **`deleteApp(appId)` — new action.** Calls `DELETE /api/apps/{appId}` via the new API client function, filters the row out of `applications`, and reconciles `currentAppId` by falling back to the first remaining app. Handles the "deleted everything" edge case by clearing `currentAppId` and the localStorage key entirely, so consumers see a defined (if empty) state instead of a dangling id.

3. **Dead-code cleanup.** The file was importing `AppAgentConfig` and `MonitoringStatus` types that were never referenced in the store body — a pre-existing eslint failure on master (verified with `git stash`). Since I was already editing imports in this file, I deleted the unused ones. Purely dead-code removal, not refactoring.

### 2. API client — `applications.ts`

**File:** `frontend/src/api/applications.ts`

Added `deleteApplication(appId: string): Promise<void>` as a thin wrapper around `client.delete('/apps/${appId}')`. No response body on success (matches the backend's 204), no client-side error handling — errors propagate to the caller so the store/UI layer can render them in context.

### 3. `ConnectionWizard` refactor — `appId` prop injection

**File:** `frontend/src/components/connections/wizard/ConnectionWizard.vue`

Added an **optional** `appId` prop and a `targetAppId` computed that resolves `props.appId ?? appStore.currentAppId`. All interior uses of `appStore.currentAppId` were swapped to `targetAppId.value`. Existing callers (`ConnectionsPage.vue` — the only one in the repo) don't pass the prop, so they inherit the historical "use the current app" behaviour with zero diff.

**Why this approach (Option A) over extracting a `ConnectionWizardCore` subcomponent (Option B)?** The plan approved Option A specifically because it's a smaller blast radius: a single prop + a single computed, with no structural changes to either file. Option B would mean carving the step-runner out of the modal shell — a meaningful refactor I'd have to own and regression-test.

**Dead-code cleanup.** Also deleted the unused `isFirstStep` computed that was declared but never referenced — another pre-existing eslint failure on master. Same rationale as the `app.ts` cleanup: I was already editing the file, the error would have polluted the validation output, and it's pure dead-code removal.

### 4. New `AppWizard` component tree

**Files:**
- `frontend/src/components/app-wizard/AppWizard.vue` *(new)*
- `frontend/src/components/app-wizard/steps/StepAppDetails.vue` *(new)*
- `frontend/src/components/app-wizard/steps/StepConnectorChoice.vue` *(new)*
- `frontend/src/components/app-wizard/steps/StepConfirm.vue` *(new)*

The wizard has **three visible steps** (`Details → Connector → Confirm`) and one sub-state (`connector`) that hands off to the embedded `ConnectionWizard`. Step state is modelled as a discriminated `StepId` union, and `currentStepIndex` is derived through a mapping so the step indicator doesn't advance when the user is inside the connector sub-flow — from the user's point of view, "add a connector" is still conceptually step 2.

**The eager-create pattern.** The wizard POSTs the new application at the end of step 1, not at the end of the wizard. Two reasons, both carried over from `ConnectionWizard`:

1. If the user adds a connector in step 2, it needs a valid `app_id` FK to attach to. Deferring creation until step 3 would force us to hold the whole wizard state in memory and POST atomically, which would require forking the connection wizard instead of reusing it.
2. The rollback-on-discard semantics are simpler: the backend is the source of truth, and the wizard just deletes the draft on cancel. No "pending draft" column, no TTL sweep.

**Trade-off.** A user who closes their browser mid-wizard leaves an orphan empty app. The plan's risk section explicitly accepts this — the app is empty (no connections, no logs), Settings will make it trivial to delete in Phase 3, and a `draft` column with a TTL sweep would be over-engineering for the symptom.

**The nested-modal problem.** `ConnectionWizard` is a full-screen modal with its own `fixed inset-0 z-50` backdrop. Embedding it *inside* `AppWizard` (which is also a full-screen modal) would stack two backdrops. The fix: when `stepId === 'connector'`, `AppWizard` hides its own shell (`v-else` on the outer `<div>`) and renders `ConnectionWizard` in its place. From the user's perspective, the wizard morphs from "New Application" into "New Connection" with a smooth visual handoff — no double-tap of darkening backdrops.

**Discard semantics.** A single `requestClose()` path handles both the X button, the Esc key, and the backdrop click. If no draft has been created yet (user on step 1 with nothing to roll back), it closes immediately. If a draft exists, it shows a confirmation overlay. On confirm, it calls `app.deleteApp(draftAppId)` to roll back the eager-created row. The `try/catch` around the delete is best-effort — if the API call fails (network, unexpected 409), the wizard still closes so the user isn't trapped; any orphan shows up in Settings → Applications in Phase 3.

**Keyboard handling.** Esc triggers `requestClose()`, but only when the AppWizard shell is actually visible. During the connector sub-flow, ConnectionWizard installs its own Esc handler and AppWizard's listener explicitly bails (`if (stepId === 'connector') return`) so the two don't fight for the same key.

**Step 1 validity.** `StepAppDetails` emits `@valid` based on `name.trim().length > 0`. The field is `maxlength="100"`, the description is `maxlength="500"` and purely cosmetic (the backend doesn't have a `description` column yet — the wizard passes the value through so if/when the column lands, the UI is already wired).

### 5. Sidebar integration — `AppSidebar.vue`

**File:** `frontend/src/components/common/AppSidebar.vue`

The app selector's `appOptions` computed now appends a **sentinel row** after the real apps:

```ts
const NEW_APP_SENTINEL = '__new_app__'
// ...
const appOptions = computed(() => [
  ...app.applications.map(a => ({ value: a.id, label: a.name })),
  { value: NEW_APP_SENTINEL, label: '+ New application' },
])
```

`handleSelectApp` intercepts the sentinel value **before** calling `app.selectApp`, so the user's current app context is never mutated. Instead it flips a local `wizardOpen` ref that conditionally mounts `<AppWizard>`.

**Why this works without a glitch.** `BaseSelect` is one-way: it reads `modelValue` from props rather than tracking its own internal state. When the user clicks the sentinel row, BaseSelect emits `update:modelValue` with the sentinel value, and our handler declines to propagate that into `app.currentAppId`. On the next render, BaseSelect re-reads `modelValue` (still pointing at the user's real current app) and visually snaps back. No intermediate state flicker.

### 6. Store tests — `app.test.ts`

**File:** `frontend/src/stores/__tests__/app.test.ts` *(new)*

Six new vitest cases using the same `client` mock pattern as the existing `schedules.test.ts`:

| Test | What it guards |
|------|----------------|
| `createApp selects the new app by default` | Lock the pre-wizard behaviour so a future "always opt-in" refactor can't silently break onboarding |
| `createApp with { select: false } leaves currentAppId untouched` | The AppWizard's eager-create contract — the draft must not mutate `currentAppId` |
| `deleteApp removes the row from the list` | Happy path; confirms local list reconciliation |
| `deleteApp falls back to the first remaining app when current is deleted` | Prevents the sidebar from pointing at a dead row after deletion of the active app |
| `deleteApp clears currentAppId when no apps remain` | Edge case — the store must leave itself in a defined state instead of a dangling id, even if the backend guard would normally block this |
| `deleteApp propagates backend errors without touching local state` | Ensures failures don't cause a half-applied local state (important for the Settings-page last-app case) |

All tests follow the repo's existing pattern: `beforeEach` resets the global client mock and localStorage; test bodies use typed `MockAxiosResponse<T>` helpers to avoid leaning on `any`.

---

## Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/stores/app.ts` | Edit | `createApp` gains `{ select }` option, `+deleteApp` action, -unused type imports |
| `frontend/src/api/applications.ts` | Edit | +`deleteApplication(appId)` |
| `frontend/src/components/connections/wizard/ConnectionWizard.vue` | Edit | +optional `appId` prop, `targetAppId` computed, -unused `isFirstStep` |
| `frontend/src/components/common/AppSidebar.vue` | Edit | +sentinel option, +sentinel interception in `handleSelectApp`, +`<AppWizard>` mount |
| `frontend/src/components/app-wizard/AppWizard.vue` | **New** | Root wizard component, step orchestration, discard-rollback flow |
| `frontend/src/components/app-wizard/steps/StepAppDetails.vue` | **New** | Step 1 — name + optional description |
| `frontend/src/components/app-wizard/steps/StepConnectorChoice.vue` | **New** | Step 2 — add connector / skip fork |
| `frontend/src/components/app-wizard/steps/StepConfirm.vue` | **New** | Step 3 — terminal success screen |
| `frontend/src/stores/__tests__/app.test.ts` | **New** | 6 store tests |

**No new Vite assets.** The wizard is small enough that Vite folded it into the main `index-*.js` chunk rather than splitting it out — correct behaviour for a sidebar-triggered modal that should load instantly on click.

---

## Validation

```bash
cd frontend
npx vue-tsc -b --noEmit                  # clean
npm run test -- --run                    # 35/35 pass (6 new)
npx eslint <touched files>               # clean
npm run build                            # clean, 1.08s
```

**Pre-existing lint errors that I also fixed (since the files were already in my diff):**

1. `frontend/src/stores/app.ts` — removed unused `AppAgentConfig` and `MonitoringStatus` type imports.
2. `frontend/src/components/connections/wizard/ConnectionWizard.vue` — removed unused `isFirstStep` computed.

Both were in master before Phase 2 began (verified with `git stash`). They're purely dead-code removal; no behaviour changes. Flagging them here explicitly so the review diff isn't confusing.

**Not validated in this pass:** end-to-end wizard flow against a live backend. The unit tests cover the store contract thoroughly, and the component shape is wired exactly as the plan specifies, but a real click-through against a running backend (and a browser) is the only way to catch modal z-index collisions, transition glitches, and the full Discard → 409 → error-text path under real network conditions. Manual smoke test recommended once a staging environment is available.

---

## What Phase 2 does *not* do

Intentional scope boundaries — deferred to later phases:

- **No Settings page.** That's Phase 3. For now, the only way to delete an existing app is via the wizard's discard flow (which only applies to *drafts*) or direct API call.
- **No last-app handling surface.** The backend's 409 + `code: "last_app"` guard is wired and tested (Phase 1), but there's no UI consumer yet — the wizard never hits it because it only deletes newly-created drafts.
- **No toast notifications.** The wizard's success path goes directly to the dashboard route; there's no existing toast system in the repo to hook into. Phase 4 may add one.
- **No telemetry to the activity feed.** Create/delete events aren't being written to `agent_log` yet. Phase 4.
- **No Esc-key support in the connector sub-flow itself.** `ConnectionWizard` manages its own keyboard handling, and AppWizard explicitly stands down while it's in view. This is correct, not a gap.

---

## Ready for Phase 3

The user-facing creation flow is complete. Phase 3 (the Settings page) can now:

- Call `app.deleteApp(appId)` directly — the store action and API client are in place.
- Fetch `GET /api/apps?include=counts` — Phase 1 shipped the endpoint.
- Open the same `<AppWizard />` from a "+ New Application" button in the Applications section — the component already takes no props beyond `@close`.
- Key its last-app disabled state off `app.applications.length === 1` — the store is already the source of truth.

No backend work is expected for Phase 3; it's pure frontend.
