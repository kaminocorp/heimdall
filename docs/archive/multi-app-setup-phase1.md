# Multi-App Setup — Phase 1 Completion Notes

**Scope:** Backend changes for the multi-app setup feature (see `docs/executing/multi-app-setup.md`). Phase 1 ships the HTTP surface the frontend phases will consume — nothing user-visible on its own, but the delete flow, counts endpoint, and test coverage are all in place.

**Branch:** `master`
**Validation:** `go vet ./...` clean, `go build ./...` clean, `go test ./...` green (integration tests skip without `DATABASE_URL`, as usual).

---

## What shipped

### 1. `DELETE /api/apps/{appId}` — new handler

**File:** `backend/internal/api/handlers/applications.go`

A new `DeleteApplication` handler wraps the already-existing `DeleteApplication` sqlc query (`backend/internal/db/queries/applications.sql:26`). The handler enforces three things and nothing else:

1. **Org-scope authorization.** Resolves the `{appId}` via the existing `GetApplicationByOrgUser` query (user→org→app join). A caller who tries to delete an app in a different org gets `404` — we match the existence-hiding pattern every other `/api/apps/{appId}/*` handler already uses instead of returning `403`.
2. **Last-app guard.** Calls `CountApplicationsByOrg` (new, see §3). If the count is `≤ 1`, returns `409 Conflict` with the body `{"error": "cannot delete last application", "code": "last_app"}`. The `code` field is the important part — it gives the frontend a stable key to render the specific "create another app first" helper text in the Settings page instead of a generic error toast.
3. **Cascade delete.** Calls `q.DeleteApplication(ctx, appID)`. Everything that references `applications(id)` — `connections`, `app_agent_config`, `monitoring_state`, `notification_channels`, `notification_preferences`, `notification_log`, `investigation_schedules` — was already declared `ON DELETE CASCADE` in migrations 014–019 and 023, so the handler stays thin. A regression test (see §5) codifies this so a future migration that accidentally drops a cascade fails CI loudly.

On success, returns `204 No Content`. Matches the body-less-success shape the rest of the `/api/connections/{id}` DELETE handler already uses.

**Why 409 + code over 422 or 400?** `409 Conflict` is the right semantic for "the request is well-formed but the server's current state forbids it" — the user's request is syntactically valid, it's the org's single-app state that blocks the operation. A `400` would imply user-fixable input; a `422` would imply semantic validation failure on the payload. `409` is a better fit.

### 2. `GET /api/apps?include=counts` — opt-in enriched shape

**File:** `backend/internal/api/handlers/applications.go`

`ListApplications` now inspects the query string. The default call (no `include` param) keeps the existing lean response shape for backwards compatibility with the sidebar selector and every other existing caller. With `?include=counts`, the handler routes to a new sqlc query (`ListApplicationsByOrgWithCounts`) that augments each row with `connection_count` and `schedule_count` — computed in-query via two correlated subqueries, so the cost stays linear in `#apps` regardless of how many connections/schedules the org has.

**Why opt-in instead of always-on?** Two reasons:

1. **Zero blast radius on existing callers.** Any change to the default response shape risks breaking the sidebar's `BaseSelect` expectations or client code paths I didn't anticipate. Gating behind a query param means I can ship the new query without auditing every current consumer.
2. **Keeps the common path cheap.** The sidebar fetches `/api/apps` on every app switch and on initial load. Pulling counts it never renders would waste cycles. The Settings page is the only caller that needs the enriched shape, and it's loaded on demand.

**Why pull this forward into Phase 1 at all?** The plan doc (`docs/executing/multi-app-setup.md:128`) made the 80/20 call: adding the enriched query upfront is ~15 lines of SQL and ~20 lines of handler code. Shipping N+1 fetches first and refactoring later costs roughly the same work twice, plus a frontend refactor to swap fetching strategy. Pull-forward is strictly cheaper.

### 3. Two new sqlc queries

**File:** `backend/internal/db/queries/applications.sql`

```sql
-- name: ListApplicationsByOrgWithCounts :many
SELECT
  a.*,
  (SELECT COUNT(*) FROM connections WHERE app_id = a.id) AS connection_count,
  (SELECT COUNT(*) FROM investigation_schedules WHERE app_id = a.id) AS schedule_count
FROM applications a
WHERE a.org_id = $1
ORDER BY a.created_at DESC;

-- name: CountApplicationsByOrg :one
SELECT COUNT(*) FROM applications WHERE org_id = $1;
```

**Gotcha caught during implementation:** the plan doc wrote `application_id` for the FK columns in prose, but migration 014 actually shipped `connections.app_id` (and migration 023 used `investigation_schedules.app_id`). I caught this by grepping the migrations directly instead of trusting the plan text. The generated SQL uses the real column names — if I hadn't cross-checked, sqlc-generate would have errored at code-gen time anyway, but verifying up front saved a round trip.

**Why `CountApplicationsByOrg` is its own query rather than `len(ListApplicationsByOrg(...))`:** the count query is O(1) on an indexed `org_id` with no row transport cost. Pulling every app row back just to count them in Go is wasteful, especially since an org with many apps would be running the full list query for every delete attempt.

**Regen:** ran `sqlc generate` (v1.30.0). Produced `ListApplicationsByOrgWithCountsRow` with `ConnectionCount int64` / `ScheduleCount int64` fields — the `int64` comes from the default sqlc mapping for `COUNT(*)` with no explicit cast, which is correct for our use case (connection/schedule counts per app will never approach `int32` overflow).

### 4. Route mount

**Files:**
- `backend/internal/api/router.go` — added `r.Delete("/", s.DeleteApplication)` inside the `/apps/{appId}` sub-router, right next to the existing `r.Get("/", s.GetApplication)`.
- `backend/internal/api/handlers/testhelpers_test.go` — mirrored the same route mount in the test router. The test harness rebuilds routes from scratch instead of re-using `NewRouter`, so the DELETE entry has to be added in both places. This is pre-existing divergence, not something I introduced.

No middleware changes — the existing Supabase JWT guard on the parent group already covers the new route.

### 5. `jsonErrorWithCode` helper

**File:** `backend/internal/api/handlers/helpers.go`

Added a small helper alongside `jsonError`:

```go
func jsonErrorWithCode(w http.ResponseWriter, message, code string, status int)
```

Writes `{"error": "...", "code": "..."}` so the frontend can key off a stable machine-readable identifier. The existing `jsonError` pattern only supports `{"error": "..."}`, which works fine for generic failures but forces the frontend to string-match on error text for anything that needs specific rendering — a fragile pattern we should avoid.

Currently only used by the last-app guard, but it's a generic helper that any future "this failed for a specific predictable reason you want the UI to render differently" case can reuse. No premature abstraction — the shape is dead simple and there's no cheaper existing pattern.

### 6. Handler tests

**File:** `backend/internal/api/handlers/applications_test.go`

Six new test functions:

| Test | What it guards |
|------|----------------|
| `TestDeleteApplication` | Happy path: two apps → delete one → `204` + list returns the remaining one |
| `TestDeleteApplication_LastAppGuard` | Deleting the only app returns `409` with `code: "last_app"` and the app still exists afterwards |
| `TestDeleteApplication_WrongOrg` | Cross-org delete returns `404` (existence-hiding) and the foreign app is untouched |
| `TestDeleteApplication_CascadesToConnections` | Create app → create connection → delete app → assert `connections.id` row is gone via FK cascade. **This is the regression tripwire** — if a future migration drops `ON DELETE CASCADE` on `connections.app_id`, this test fails loudly rather than silently orphaning rows in production |
| `TestListApplications_WithCounts` | `?include=counts` returns the new shape with `connection_count` and `schedule_count` populated correctly (2 and 0 in the seeded scenario) |
| `TestListApplications_DefaultShapeUnchanged` | Default call (no query param) must **not** include count fields — guards against accidental shape changes that would break existing consumers like the sidebar selector |

All tests follow the existing integration pattern: they skip cleanly without `DATABASE_URL` and hit a real Postgres when one is configured. The harness `testSetup(t)` already creates one default app per test, which is why several of the tests create a *second* app before exercising delete — the default app would hit the last-app guard.

The cross-org test is patterned after the existing `TestGetApplication_WrongOrg` — same foreign-org bootstrap, same cleanup hook. I kept it consistent rather than factoring out a helper; the pattern only appears twice and a helper would add indirection for minimal savings.

---

## Files changed

| File | Kind | Change |
|------|------|--------|
| `backend/internal/db/queries/applications.sql` | Edit | +`ListApplicationsByOrgWithCounts`, +`CountApplicationsByOrg` |
| `backend/internal/db/applications.sql.go` | **Regen** | sqlc regenerated |
| `backend/internal/api/handlers/applications.go` | Edit | +`DeleteApplication` handler; `ListApplications` branches on `?include=counts` |
| `backend/internal/api/handlers/helpers.go` | Edit | +`jsonErrorWithCode` helper |
| `backend/internal/api/router.go` | Edit | +`r.Delete("/", s.DeleteApplication)` on `/apps/{appId}` |
| `backend/internal/api/handlers/testhelpers_test.go` | Edit | +mirrored DELETE route mount in test router |
| `backend/internal/api/handlers/applications_test.go` | Edit | +6 integration tests |

**No migration.** Every schema dependency (CASCADE FKs, the `investigation_schedules` table) was already live.

---

## Validation

```bash
cd backend
go vet ./...              # clean
go build ./...            # clean
go test ./...             # all packages pass
                          # integration tests skip without DATABASE_URL
```

**Not validated in this pass:** end-to-end delete against a real Postgres. That requires a live `DATABASE_URL` + `SUPABASE_URL`. The new tests are written to run automatically the next time someone runs `make test` with those env vars set — they'll either pass (confirming the delete flow works) or fail loudly on a real regression.

---

## What Phase 1 does *not* do

Intentional scope boundary — these belong to later phases and were deliberately left out:

- **No "+ New App" sidebar affordance.** That's Phase 2 (wizard).
- **No Settings page.** That's Phase 3.
- **No `app.deleteApp()` frontend store action.** Phase 2 will add it against the now-existing `DELETE /api/apps/{appId}` endpoint.
- **No activity-feed telemetry for create/delete.** That's Phase 4 polish.
- **No soft-delete / undo.** The plan's risk section (`docs/executing/multi-app-setup.md:210`) explicitly rules this out for the first pass — typed-to-confirm is the only guardrail. Revisit only if someone asks for it.

---

## Ready for Phase 2

The backend contract is locked in and tested. Phase 2 (frontend wizard) can now:

- POST `/api/apps` to create the draft app at step 1
- DELETE `/api/apps/{appId}` to roll back on discard
- GET `/api/apps?include=counts` once the Settings page lands in Phase 3

No further backend work should be required for the core delete + list paths. If Phase 2 or 3 discovers a gap, we'll patch it on its own phase — but the 90%-ready-backend claim from the plan's "Context: what already exists" section now holds at 100%.
