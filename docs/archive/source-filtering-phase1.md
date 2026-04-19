# Source Filtering — Phase 1 Completion

**Scope:** Auto-discover source names from webhook payloads, let users per-app opt sources in or out, drop everything else at ingestion. Connections stay app-scoped — Phase 2 adds the org-scoped variant.
**Plan:** `docs/executing/source-filtering-and-org-connections.md`

---

## The Problem This Solves

Fly.io's Log Shipper is org-wide: it drains logs from every Fly app into one HTTP endpoint. Before Phase 1, every log arriving at a Heimdall webhook connection was stored verbatim — so a user with five Fly apps saw all five mixed together in whichever Heimdall Application the webhook was attached to. Any multi-source integration (GitHub repos aside, which uses a bespoke table) hits the same structural mismatch.

Phase 1 generalises the "one integration, many sources, filter at the edge" pattern GitHub already uses. The user configures one Fly drain, Heimdall discovers each Fly app automatically as its logs arrive, and the user flips the ones they care about on for each Heimdall Application independently. Everything else is dropped.

---

## What Changed

### 1. Two new tables — discovery + per-app toggles

**File:** `backend/migrations/034_source_filtering.up.sql` (+ `.down.sql`)

Two tables with RLS:

- **`connection_sources`** — the auto-discovery ledger. One row per `(connection_id, source_name)`. Populated by ingestion; `last_seen_at` refreshes on every hit so the UI can show staleness tiers. Indefinite retention — Fly apps that go silent aren't auto-pruned; staleness is surfaced in the UI instead.
- **`app_source_filters`** — the user-facing toggle list. One row per `(app_id, connection_id, source_name)` with `enabled` boolean. Missing rows mean "not configured" and are dropped by ingestion. A partial index `WHERE enabled = true` backs the hot-path ingestion lookup.

**RLS policies match each parent table's idiom**: `connection_sources` uses the single-owner `user_id = app_current_user_id()` pattern inherited from the connections table (migration 013). `app_source_filters` uses the org-member join from applications (migration 026). Two idioms on purpose — the right one for the parent's audience.

**Migration numbering:** 033 was the last slot (RLS on idempotency table, v0.45.1), so source filtering takes **034**. Phase 2's org-scoping migration will be 035.

### 2. sqlc queries

**File:** `backend/internal/db/queries/source_filters.sql`

Six queries: `UpsertConnectionSource`, `ListConnectionSources`, `ListAppSourceFilters`, `UpsertAppSourceFilter`, `DeleteAppSourceFilter`, `ListEnabledSourceNames`, plus `CountAppSourceFilters`. `ListEnabledSourceNames` is the ingestion hot-path query — one indexed scan per webhook batch, returning just the string names.

Regenerated via `make sqlc-generate` — new `ConnectionSource` and `AppSourceFilter` models land in `internal/db/models.go` automatically.

### 3. Parser populates `SourceName` from `fly.app.name`

**File:** `backend/internal/api/handlers/webhooks.go` (struct), `webhook_parsers.go` (population)

Added a `SourceName` field to `webhookLogRequest`. Only the Fly.io parser populates it directly: `SourceName = entry.Fly.App.Name` (bare app name like `"trajan"`, distinct from the prefixed `SourceType = "flyio/trajan"`). Other parsers leave it empty; a new helper `sourceNameOf(e webhookLogRequest)` falls back to `SourceType` when `SourceName` is blank.

**Why two fields:** users filter by `"trajan"` in the UI, not `"flyio/trajan"`. But `SourceType` is still useful as the "what kind of payload was this" label on the stored log entry, so we kept it unchanged.

**Non-Fly parsers were left alone** — source name extraction for OTLP (service.name), syslog (hostname), and generic webhooks is deferred to Phase 4. For Phase 1 those paths fall back to `SourceType`, which reproduces today's behaviour for callers that pre-configure a single source type.

### 4. Ingestion filtering

**File:** `backend/internal/api/handlers/webhooks.go` (`ingestEntries`)

Every webhook batch now runs a two-step filter inside its insert transaction:

1. **Discover first.** Upsert each unique source name in the batch into `connection_sources`. Runs regardless of filter state — new sources must appear in the selector UI before the user can opt them in.
2. **Filter second.** Load `ListEnabledSourceNames(connection_id, app_id)` once per batch into a Go map, then drop any entry whose source name isn't in the map.

Everything runs in the same transaction, so discovery and filtered inserts are atomic — either all succeed or the batch is rolled back.

**Drop-by-default is the core semantic.** When `app_source_filters` has no enabled rows for the connection, every entry falls through to `filtered++`. The response is still 201 with the same shape — `{ accepted, filtered, format }`. Added a `filtered` field to `webhookLogResponse` (omitempty, so existing callers reading only `accepted` keep working).

**Idempotency interacts correctly.** The idempotency cache stores the final response body. A replay returns the cached body — which includes the `accepted`/`filtered` numbers from the first call, even if filters changed between request and replay. This is semantically correct: idempotency means "identical response for identical key".

**OTLP path unchanged.** The OTLP handler (`otlp.go`) has its own ingestion path that doesn't go through `ingestEntries`. Source extraction for OTLP service.name is Phase 4.

### 5. API — four endpoints under `/api/connections/{id}/sources`

**File:** `backend/internal/api/handlers/source_filters.go` (new), `backend/internal/api/router.go` (registration)

- `GET /api/connections/{id}/sources` — merged view: union of `connection_sources` (with timestamps) and `app_source_filters` (with enabled state). Same merge shape as `ListGitHubRepos`. Empty timestamps encode the "never seen" staleness tier for manually-added sources.
- `PUT /api/connections/{id}/sources` — bulk-upsert from `{ sources: [{ source_name, enabled }] }`. Single transaction; cap 500 items (matches GitHub's own upper bound).
- `POST /api/connections/{id}/sources` — manual add. Creates both a placeholder `connection_sources` row and an `app_source_filters` row with `enabled=true` (manual adds are always on — users who type a name in the UI intend to accept those logs).
- `DELETE /api/connections/{id}/sources?name=...` — query-parameter instead of path-segment because source names can contain slashes (`"vercel/lambda"`) and some proxies normalise `%2F` back to `/`. Deletes only the filter row; the discovery row stays so the source reappears if traffic returns.

Auth is the standard pattern used elsewhere: JWT → `GetConnectionByUser` validates ownership → `UserQueries` opens the RLS transaction → writes are committed explicitly on the success path.

### 6. Frontend — `SourceSelector.vue` + wizard integration

**Files:** `frontend/src/types/source.ts`, `frontend/src/api/sources.ts`, `frontend/src/components/connections/SourceSelector.vue`, `frontend/src/components/connections/ConnectionDetailModal.vue`, `frontend/src/pages/ConnectionsPage.vue`, `frontend/src/components/connections/wizard/steps/StepFlyioDrainSetup.vue`

**`SourceSelector.vue`** is the core component. It:

- Fetches the merged list on mount and polls every 10s while open (catches new sources discovered during initial drain setup — the user doesn't have to refresh).
- Renders each source with a staleness dot (green < 1h, amber 1–24h, red > 24h, muted "never seen") using a ticking `nowTick` ref so the dots update as time passes without re-fetching.
- Shows a warning banner when any enabled source is stale — the signal "your Fly app was deleted or stopped".
- Manual-add input → calls `POST` → refreshes the list.
- Per-row Remove button → calls `DELETE` → refreshes.
- Bulk Save → calls `PUT` with the current enabled states.
- **Embedded prop**: hides the Close/Cancel buttons when the component is mounted inside another container (the Fly.io wizard step). In embedded mode, Save keeps the panel in place rather than emitting `close`.

**Integration points:**

- `ConnectionDetailModal` shows a `Sources` action button for `webhook_logs` connections, emitting `manage-sources`. A `SOURCE_FILTERED_TYPES` set inside the component gates the button; expand in Phase 4 as new parsers populate `SourceName`.
- `ConnectionsPage` hosts the opened selector inline (same pattern as `GitHubRepoSelector`) and wires the `manage-sources` event.
- `StepFlyioDrainSetup` embeds the selector as a "Step 6: Select Sources" panel after the Fly Log Shipper deploy instructions, with `embedded` set. New users thus see the drop-by-default behaviour explained in context and can pre-add their app names before logs start arriving.

**Design token alignment:** staleness dots use the existing techno-brutalist `status-ok` / `status-warn` / `status-critical` / `text-muted` tokens — same set used by `ConnectionBubble` and other connection UI.

### 7. Test coverage

**Backend:**

- `TestParseVectorFly_Single` — extended to assert `SourceName == "my-fly-app"` (bare app name, not prefixed).
- `TestParseVectorFly_SourceNameFallback` — when `fly.app.name` is missing, `SourceName` is empty and `sourceNameOf()` falls back to `SourceType`.
- `TestIngestWebhookLogs` — updated to pre-enable the `"webhook"` source before posting (otherwise drop-by-default would make the existing assertion fail).
- `TestIngestWebhookLogs_DropByDefault` — new. Verifies that ingestion with no filters returns `accepted=0, filtered=1` and that the source is nonetheless surfaced in the selector.
- `TestIngestWebhookLogs_FlyioSourceName` — new. Batch with two Fly apps; only `trajan` is enabled; asserts `accepted=1, filtered=1` and both sources in the selector.
- `TestSourceFilters_CRUD` — new. Full lifecycle: empty list → manual add → toggle off via PUT → delete via query param. Asserts discovery row survives filter delete.
- `TestSourceFilters_AuthorizationScoping` — new. Invalid/random connection IDs return 404; missing `?name=` and empty `source_name` return 400.

**Frontend:**

- `src/types/__tests__/source.test.ts` — 6 pure unit tests for `classifyStaleness` covering the "never seen" null case, the three temporal bands, and the exact 1h and 24h boundary transitions.

### 8. Test-harness fixes (pre-existing issues surfaced by new tests)

**File:** `backend/internal/api/handlers/testhelpers_test.go`

Two pre-existing failures blocked me from running the new handler tests. I made the smallest viable fixes so the Phase 1 suite is green:

- **`INSERT INTO users` → `ON CONFLICT DO UPDATE`**: Hosted Supabase ships an `on_auth_user_created` trigger that auto-populates `public.users` when `auth.users` gets a row. The old test helper unconditionally tried a plain INSERT right after, which duplicate-keyed. The upsert works whether the trigger is present or not.
- **`Listener: connectors.NewListenerManager()`** on the test server. v0.45.2 added `s.Listener.Stop(connID)` to `UpdateConnection` unconditionally, but the test harness had been leaving `Listener` nil — so `TestUpdateConnection` was panicking on a nil-pointer dereference. Wiring in a real listener manager restores the green path.

Both fixes touch the test harness only; neither changes production code.

---

## What Did NOT Change (and why)

### No changes to OTLP ingestion

OTLP has its own ingestion handler (`otlp.go`). Extracting `service.name` as a source name is Phase 4. Phase 1 webhook filtering doesn't apply to `/v1/logs` — OTLP callers continue to get today's unfiltered behaviour until Phase 4.

### No changes to the Postgres or Supabase pollers

Pollers pull from external databases and insert through a different path (`connectors.Poller`, not the webhook handler). They're out of scope for Phase 1. The filter spec generalises naturally to pollers in a future phase, but we didn't expand into that territory here.

### No changes to the GitHub repo selector

Phase 3 of the plan migrates `github_repos` into this generic `connection_sources` / `app_source_filters` system, with a data migration. Phase 1 ships the infrastructure without touching GitHub.

### No new columns on `connections`

Connections remain strictly app-scoped. `org_id` and nullable `app_id` are Phase 2 — the plan explicitly carves Phase 2 as "depends on Phase 1's source filtering infrastructure."

### No filtered-traffic counter in the UI

The plan mentions "Showing 142 filtered entries/hr from disabled sources" in the mock. Producing that number needs either a metrics pipeline or a per-connection counter table; both are disproportionate to Phase 1's scope. The `filtered` field is in the webhook response for operator debugging, and `slog.Info` logs it in the server log. The UI counter can land later.

### No migration of existing callers

Existing webhook payloads that only set `source_type` (no `fly.app.name`) now fall back to filtering by the `source_type` value. Because drop-by-default applies, those callers will start seeing `accepted=0, filtered=N` after this migration — **existing integrations must enable their source names explicitly** or they'll stop storing logs. This is the deliberate, documented behaviour of drop-by-default; the wizard's new "Step 6: Select Sources" panel is the new-user mitigation path.

---

## Verification

- `go build ./...` — clean
- `go vet ./...` — clean
- `npx vue-tsc --noEmit` — clean
- `migrate up` — applied cleanly on the local Supabase instance (`34/u source_filtering`)
- Phase 1 backend tests — all 13 pass (`TestIngestWebhookLogs*`, `TestSourceFilters_*`, `TestParseVectorFly_*`, `TestUpdateConnection`)
- Phase 1 frontend tests — all 6 pass (`classifyStaleness`)

Eleven pre-existing test failures remain in unrelated handler tests (`TestListLogs_*`, `TestInviteMember*`, `TestOnboard*`, etc.). These were already broken before Phase 1: `log_buffer` inserts without `app_id` (migration 029), and `public.users` collisions in other test helpers that don't use the same upsert pattern. They're tracked separately.

---

## File Map

```
backend/
  migrations/034_source_filtering.up.sql              NEW
  migrations/034_source_filtering.down.sql            NEW
  internal/db/queries/source_filters.sql              NEW
  internal/db/source_filters.sql.go                   GENERATED (sqlc)
  internal/db/models.go                               REGENERATED (sqlc)
  internal/api/handlers/webhooks.go                   CHANGED — SourceName field, sourceNameOf, filter pipeline, webhookLogResponse.Filtered
  internal/api/handlers/webhook_parsers.go            CHANGED — Fly parser populates SourceName
  internal/api/handlers/source_filters.go             NEW — four endpoints
  internal/api/handlers/source_filters_test.go        NEW
  internal/api/handlers/webhooks_test.go              CHANGED — drop-by-default + Fly filtering tests
  internal/api/handlers/webhook_parsers_test.go       CHANGED — SourceName assertions
  internal/api/handlers/testhelpers_test.go           CHANGED — users upsert + Listener wiring
  internal/api/router.go                              CHANGED — four new routes

frontend/
  src/types/source.ts                                 NEW
  src/types/__tests__/source.test.ts                  NEW
  src/api/sources.ts                                  NEW
  src/components/connections/SourceSelector.vue       NEW
  src/components/connections/ConnectionDetailModal.vue CHANGED — Sources button + manage-sources emit
  src/components/connections/wizard/steps/StepFlyioDrainSetup.vue CHANGED — embedded SourceSelector
  src/pages/ConnectionsPage.vue                       CHANGED — open/close source selector
```

---

## Handoff to Phase 2

Phase 2 adds the `org_id` column to `connections` and makes `app_id` nullable so a single connection can feed multiple apps. The data-model piece of Phase 1 anticipates this: `app_source_filters` is already keyed by `(app_id, connection_id, source_name)` — one source name can be enabled independently for many apps without schema changes. The ingestion hot-path query (`ListEnabledSourceNames`) becomes a fan-out (iterate over all apps with an enabled filter for the source), but its inputs don't need to change.

The main surface area Phase 2 will touch:

- Connection create/update endpoints accept optional `app_id` (omit → org-scoped).
- Webhook handler routes to one or more apps depending on whether `conn.app_id` is NULL.
- Wizard gains a "Connection Scope" step before the platform-specific flow.
- Connections page adds an "Org-wide" badge and renders org-scoped connections in every app's list.
