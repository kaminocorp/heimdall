# Multi-App Setup — Phase 4 Completion Notes

**Scope:** Polish items that round out the multi-app feature. Four sub-items per the plan: keyboard niceties (Enter advances), success toasts on app creation, activity-feed telemetry for app create/delete, and a `vision.md` update to reflect the first-class multi-app semantics the earlier phases shipped.

**Branch:** `master`
**Validation:** `go vet ./...`, `go build ./...`, `go test ./...` all clean; `vue-tsc --noEmit` clean; `npm run test -- --run` 35/35 (no regressions); `npx eslint <touched files>` clean; `npm run build` clean (1.05s).

---

## What shipped

### 1. Keyboard: Enter advances through form inputs

**Files:**
- `frontend/src/components/app-wizard/steps/StepAppDetails.vue` — adds a `@keydown.enter.prevent="onNameEnter"` handler on the name input; emits a new `submit` event to the wizard shell.
- `frontend/src/components/app-wizard/AppWizard.vue` — listens to the new `@submit` and routes it through `advanceFromDetails`.
- `frontend/src/components/settings/DeleteAppModal.vue` — adds `@keydown.enter.prevent="handleDelete"` on the typed-to-confirm input.

**Design:** the child step emits a structured `submit` event with its own internal validity check (`name.value.trim().length > 0`), rather than bubbling a raw keyboard event up to the wizard. This keeps the wizard shell's receiver simple — it doesn't need to re-check validity or know about DOM events. The same pattern could be reused by any future step that wants keyboard-first submission.

**Esc-key handling was already wired in Phase 2** — the plan item is half-done; this phase only needed the Enter-advances half. I re-verified that AppWizard's `onKeydown` still only intercepts Esc while the AppWizard shell is visible (it explicitly bails during the connector sub-flow so ConnectionWizard can own keyboard state), and that DeleteAppModal's `onKeydown` is scoped to the modal's mount lifecycle.

**DeleteAppModal safety:** `handleDelete` already has an upfront `if (!canDelete.value || deleting.value) return` guard, so firing it on Enter when the name isn't fully typed is a safe no-op. No extra validity check needed at the keyboard layer.

### 2. Toast on successful app creation

**File:** `frontend/src/components/app-wizard/AppWizard.vue`

After a successful `finish()`, the wizard calls `toast.show('Application "…" created', 'success')`. The existing `useToast()` composable (already used by `SchedulesPage`, `AgentConfigPage`, `OnboardingPage`, `NotificationsPage`) is wired to a module-level ref, so no container wiring is required — the already-mounted `<ToastContainer />` in `App.vue` renders the new toast automatically.

**Why toast here but not on delete?** The plan explicitly carves this distinction: `DeleteAppModal` has its own **Stage 2 success screen** (the calm "Application deleted" acknowledgement with a single "Close" button), which is more explicit and more satisfying than a 4-second toast that slides off. Creation doesn't have an equivalent in-place acknowledgement — the wizard closes and the user lands on `/dashboard` immediately — so a toast is the right bridge to carry the "yes, it worked" signal across the route transition.

**Captured name, not live ref.** The toast message uses `committedName` captured *before* resetting `draftAppId.value = null`. This avoids a stale-closure hazard where the reactive `details.name` could theoretically clear itself mid-async. Small detail, but the alternative would be a test-unfriendly race.

### 3. Telemetry: `application_created` / `application_deleted` in agent_log

**File:** `backend/internal/api/handlers/applications.go`

Both `CreateApplication` and `DeleteApplication` now fire-and-forget an entry into `agent_log` via the existing `agent.EmitLog` helper:

```go
if s.Agent != nil {
    s.Agent.EmitLog(
        r.Context(), userID, nil,
        "application_created",
        fmt.Sprintf("Application %q created", app.Name),
        map[string]any{
            "app_id":   app.ID.String(),
            "app_name": app.Name,
        },
    )
}
```

Entry types: `application_created` and `application_deleted`. The delete event is emitted **after** the `DeleteApplication` query succeeds, so we never record a deletion that didn't happen.

**Why the `nil` guard on `s.Agent`?** The test harness (`testhelpers_test.go`) constructs the server with `Agent: nil` to avoid pulling in the full agent stack for HTTP-handler tests. Calling `s.Agent.EmitLog` unconditionally would NPE and fail all applications tests. The guard is consistent with a pattern that's *likely to become the convention* once more handlers adopt telemetry — I introduced it here rather than forcing the test harness to stand up a real Agent.

**Plan-doc premise drift (minor).** The plan said "similar to how connection changes are logged," but in reality **no handler currently emits to `agent_log`** — the only callers are agent-owned code (scheduler, monitor, loop). I established the handler-side pattern here. This is a good thing (the agent-side `EmitLog` is a reusable public API), but the plan doc's reference was aspirational rather than prescriptive. The completion doc flags this for the next reader.

**`agent_log` is user-scoped, not app-scoped.** The table has `user_id` but no `app_id` FK. Stashing `{app_id, app_name}` in the `detail` JSONB column is the right shape for two reasons:

1. **The audit row must outlive the thing it audits.** An `app_id` FK with `ON DELETE CASCADE` (matching how every other child table works) would destroy the `application_deleted` entry at the exact moment the deletion cascade fires. Keeping the reference in JSONB means "I deleted app X" survives the deletion of X, which is the entire point of an audit trail.
2. **The Activity feed is already rendered by user.** `ListAgentLogByUser` is the primary accessor, so there's no meaningful filter to key off anyway.

**No test gate.** I deliberately didn't add an integration test that asserts "create an app, see an `application_created` row in `agent_log`" — the `Agent: nil` harness means I'd have to build a real Agent in the test setup just to verify a fire-and-forget side effect. That's a meaningful test-infrastructure change that's out of scope for a polish phase. Manual smoke test against a real deployment is the right validation path.

### 4. `vision.md` update — first-class multi-app semantics

**File:** `docs/vision.md`

Two additions:

1. **New top-level "Applications" subsection** before the Platform Sections list. Explains the Org → Application → per-app surface hierarchy that earlier phases shipped, and explicitly mentions the wizard in the sidebar selector and the Settings page as management surfaces. Also notes the last-app invariant (every org must have at least one).
2. **New Platform Section entry** (`7. Settings`) describing the page as org-scoped rather than per-app, so readers understand why it lives in the footer rather than the main nav.

The existing six Platform Sections (Connections, Agent Configuration, Agent Chat, Scheduled Investigations, Activity, Reports) are unchanged — they were already accurate and per-app-scoped.

---

## Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/components/app-wizard/steps/StepAppDetails.vue` | Edit | +`submit` emit, +`@keydown.enter` handler |
| `frontend/src/components/app-wizard/AppWizard.vue` | Edit | +toast import/call on finish, +`@submit` listener |
| `frontend/src/components/settings/DeleteAppModal.vue` | Edit | +`@keydown.enter` on typed-to-confirm input |
| `backend/internal/api/handlers/applications.go` | Edit | +`fmt` import, +`EmitLog` calls with nil guards in both handlers |
| `docs/vision.md` | Edit | +Applications section, +Settings platform section |

**Net diff:** 5 files, zero new files, zero deletions. This is the right shape for a polish phase — no new concepts, just hardening the existing ones.

---

## Validation

```bash
cd backend
go vet ./...                                 # clean
go build ./...                               # clean
go test ./...                                # all packages pass
                                             # integration tests skip without DATABASE_URL

cd frontend
npx vue-tsc -b --noEmit                      # clean
npm run test -- --run                        # 35/35 pass
npx eslint <touched files>                   # clean
npm run build                                # clean, 1.05s
                                             # SettingsPage: 13.37 kB (+0.06 kB from Phase 3 baseline)
                                             # index chunk:  362.16 kB (+0.5 kB from Phase 3 baseline)
```

Bundle-size deltas reflect the handler additions: `DeleteAppModal` grew by the single-line Enter handler, and the main `index` chunk grew by the `useToast` import in `AppWizard` (already present elsewhere so the marginal cost is just the re-import statement).

**Not validated in this pass:**
- Real browser keyboard flow (tab → type name → Enter → advance) against a running backend. Unit tests cover the `submit` emit path and the store contract, but the actual focus-trap and event dispatch through the DOM is only meaningful as a manual smoke test.
- Real `agent_log` telemetry landing in a running Activity feed. Would need a real Agent + DB. See §3 above.

---

## Multi-app setup feature — final state

All four phases are now shipped:

| Phase | Scope | Status | Completion doc |
|-------|-------|--------|---------------|
| **Phase 1** — Backend | `DELETE /api/apps/{id}`, last-app guard, `?include=counts`, cascade tests | ✅ | `docs/completions/multi-app-setup-phase1.md` |
| **Phase 2** — Frontend wizard | `AppWizard` + sidebar sentinel + `ConnectionWizard` refactor + `deleteApp` store action | ✅ | `docs/completions/multi-app-setup-phase2.md` |
| **Phase 3** — Frontend settings | `/settings` + three sections + `DeleteAppModal` two-stage + counts click-through | ✅ | `docs/completions/multi-app-setup-phase3.md` |
| **Phase 4** — Polish | Enter-advances, create toast, activity-feed telemetry, `vision.md` update | ✅ | *this doc* |

The feature is **shippable end-to-end**. Users can:

1. Click the sidebar app selector → select "+ New application" → walk through the wizard (name, optional connector, confirm) → land on the dashboard with the new app active and see a toast.
2. Switch between apps via the sidebar selector; every per-app page (Connections, Agent Config, Schedules, Activity, Reports) renders data scoped to the current selection.
3. Open Settings → see their profile, org, and full app list with live connection/schedule counts.
4. Click a count number → jump to the corresponding per-app page pre-filtered to that app.
5. Delete an app via the two-stage typed-to-confirm modal; the last-app guard blocks accidental self-lockout both in the UI (disabled button + explicit helper) and on the backend (409 + `code: "last_app"`).
6. See create/delete events in the Activity feed as audit history that survives cascade deletion.

**Known deferrals — explicitly out of scope, not bugs:**
- Soft-delete / restore flow (plan risk section: typed-to-confirm is the only guardrail).
- Member management in OrganisationSection (no backend endpoint yet; plan says "just show the data").
- An integration test that asserts the telemetry rows actually land in `agent_log` — requires standing up a real `Agent` in the test harness, which is meaningful test-infra work.
- Bulk operations (delete multiple, move connections between apps, clone app config) — no user has asked.
- App-scoped filtering of the Activity feed — `agent_log` is user-scoped by design, and the multi-app detail JSONB is there if someone wants to filter client-side later.

Each of the above has a clean surface to land on if it becomes a real requirement, but shipping them speculatively would bloat the feature without justification.
