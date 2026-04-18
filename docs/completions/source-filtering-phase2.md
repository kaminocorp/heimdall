# Source Filtering — Phase 2 Completion

**Scope:** Connections can now be created at the organisation level, shared across every app in the org. Each app independently enables the sources it cares about via the Phase 1 filter machinery; ingestion fans out per source. Phase 3 (GitHub repo migration) and Phase 4 (extended source extraction) remain.
**Plan:** `docs/executing/source-filtering-and-org-connections.md`
**Builds on:** `docs/completions/source-filtering-phase1.md`

---

## The Problem This Solves

Phase 1 attached filters to a connection's owning app — one webhook, one Heimdall Application. But Fly's Log Shipper is org-wide by design: a single drain carries every Fly app's logs. With Phase 1 alone, you had to pick one Heimdall app to host that drain, and the filter UI for all *other* apps was empty.

Phase 2 breaks that mapping. One org-wide drain, N Heimdall apps, each app independently chooses which Fly apps' logs it stores. The filter infrastructure is unchanged — it was already keyed by `(app_id, connection_id, source_name)` — but connections themselves can now live above any specific app.

---

## What Changed

### 1. Connections table — add `org_id`, relax `app_id`

**File:** `backend/migrations/035_org_connections.up.sql` (+ `.down.sql`)

Three moves:

- `connections.org_id UUID NOT NULL` — backfilled from `applications.org_id` before the NOT NULL constraint is set. Every existing row becomes app-scoped with `org_id` inherited from its parent app.
- `connections.app_id DROP NOT NULL` — org-scoped connections leave it NULL. The sqlc Go type becomes `*uuid.UUID`.
- **RLS rewrite.** `connections_owner` (`user_id = app_current_user_id()`) is replaced with `connections_org_member` — a subquery through `org_members`. `connection_sources` gets the same treatment. Single-owner access gives way to shared org access; a connection your colleague created is now visible to you if you share an org.

`connections.user_id` stays NOT NULL and unchanged. It's the creator-audit field (who originally wired this integration) — the access gate is org membership, not creator identity.

**Down migration is destructive for org-scoped rows.** `app_id` can't be flipped back to NOT NULL without something occupying every NULL slot; the down migration `DELETE`s org-scoped connections before restoring the old constraint. This is documented in the file and tested by applying-then-reverting locally.

### 2. sqlc — pointer override for nullable UUIDs

**File:** `backend/sqlc.yaml`

Added a global nullable-uuid override:

```yaml
- db_type: "uuid"
  nullable: true
  go_type:
    type: "UUID"
    import: "github.com/google/uuid"
    pointer: true
```

Nullable UUID columns now generate `*uuid.UUID` instead of `pgtype.UUID`. This matches how nullable `timestamptz` is already handled, and JSON serialises as UUID-string-or-null — which is what the frontend expects. It's a schema-wide change; four existing columns switched: `agent_log.conversation_id`, `conversations.investigation_id`, `notification_log.agent_log_id`, and of course `connections.app_id`. The two non-generated callers (`notifications/notifier.go`, `agent/emit.go`) were updated to pass `&uuid` instead of `pgtype.UUID{Bytes: uuid, Valid: true}`.

### 3. Connection queries — org-member access everywhere

**File:** `backend/internal/db/queries/connections.sql`

All access queries moved from user-owner filters to org-member joins. Names were preserved for minimal caller churn:

- `GetConnectionByUser(id, user_id)` — now joins `connections → org_members` and returns the connection if the user is a member of the connection's org.
- `ListConnectionsByUser(user_id)` — every connection in every org the user belongs to.
- `ListConnectionsByApp(app_id)` — **the key Phase 2 query**. Returns app-scoped connections for the app *plus* org-scoped connections in the app's org. The nested subquery resolves org in one call; no join needed because we only want `org_id`.
- `UpdateConnection` / `DeleteConnectionByUser` — same org-member predicate, so any member can edit/delete.
- `CreateConnection` — new signature: takes `org_id` (always) and `app_id` (optional, via `sqlc.narg`). The backend resolves org before calling.
- `ListAppsEnabledForSource(connection_id, source_name)` — **new**. Returns every app with `enabled = true` for that pair. This is the fan-out hot-path query, backed by the same partial index `idx_app_source_filters_lookup` from Phase 1.
- `GetApplicationByOrgUser` is unchanged — it was already org-member-based.

One cross-phase touch: `ListEnabledGitHubReposByApp` had to grow an explicit `::UUID` cast on its `$1` parameter. With `connections.app_id` now nullable, sqlc inferred the query param as `*uuid.UUID`, which rippled unhelpfully into agent code. The cast pins the parameter type to non-null UUID — GitHub connections are always app-scoped, so the column value being nullable doesn't matter.

### 4. `CreateConnection` handler — two creation modes

**File:** `backend/internal/api/handlers/connections.go`

The create handler now branches on whether the request body carries `app_id`:

- **App-scoped** (Phase 1 behaviour): the app is looked up via `GetApplicationByOrgUser` and its `org_id` becomes the connection's `org_id`. TOCTOU-safe because the lookup and insert share one transaction.
- **Org-scoped**: the org is resolved via `resolveOrgForUser` (X-Org-ID header or user's primary org). `app_id` is NULL on the connection.

A validation gate — `supportsOrgScope(connType)` — permits org-scoping only for `webhook_logs` and `otlp`. Postgres, GitHub, syslog, etc. must still pass `app_id`. The frontend wizard hides the option for non-eligible types; the backend enforces defensively so a handcrafted API call can't bypass the constraint.

A new helper `appIDOrZero(*uuid.UUID) uuid.UUID` converts the nullable pointer into a concrete value for connector APIs (StartPoller, NewSyslog) that pre-date the nullable world. These APIs only ever run for pull-style connection types (postgres/supabase/flyio-polling/vercel/railway/mongodb/syslog) that are exclusively app-scoped today — for those, the pointer is always non-nil and the helper is a noisy but safe dereference. The startup rehydration paths in `cmd/heimdall/main.go` additionally skip any row that somehow has a NULL app_id for these types, logging a warning instead of crashing.

### 5. Webhook ingestion — fan-out for org-scoped connections

**File:** `backend/internal/api/handlers/webhooks.go`

`ingestEntries` got a new signature — it takes the whole `connResult` (which now includes `OrgID` and `*uuid.UUID` AppID) instead of three positional args. The filter resolution step was renamed from "load enabled names" to "build routes per source":

- **App-scoped** (`conn.AppID != nil`): one query (`ListEnabledSourceNames`) populates `routesBySource[name] = [app]`. Same behaviour as Phase 1, just reshaped.
- **Org-scoped** (`conn.AppID == nil`): one `ListAppsEnabledForSource` call per *distinct source name in the batch* populates `routesBySource[name] = [app1, app2, ...]`. The query cost is O(distinct sources), not O(entries) — source names typically repeat within a batch.

The insert loop then iterates `targets` per entry, inserting one row per (entry, app). The response body adds an `inserts` counter alongside `accepted`/`filtered` for operator visibility — a batch of 3 entries ingested into 2 apps each reports `accepted=3, inserts=6`. Fan-out on write; cheap reads (the activity feed's `WHERE app_id = ?` queries stay the same, no filter join).

**Atomicity:** the discovery upsert, route lookup, fan-out inserts, and idempotency cache write are all inside one Postgres transaction. Partial failure rolls the whole batch back — no half-ingested org-scoped batches.

**Logging** gained a `scope` field (`"app"` or `"org"`) so a single `slog` call distinguishes the two paths in production log aggregators.

### 6. OTLP — explicit rejection for org-scoped

**File:** `backend/internal/api/handlers/otlp.go`

OTLP ingestion has its own path that doesn't go through `ingestEntries`. Phase 4 adds `service.name` extraction (the OTLP equivalent of `fly.app.name`), which is what'd make org-scoped OTLP meaningful. Until then, the OTLP handler rejects NULL-app-id connections with a 400 — better than silently storing a log with no routing key.

### 7. Source-filter endpoints — `?app_id=` for org-scoped

**File:** `backend/internal/api/handlers/source_filters.go`

A new helper `resolveSourceFilterApp(ctx, queries, conn, userID, r)` enforces the policy:

- **App-scoped connections**: the connection's `app_id` is authoritative. `?app_id=` is optional; if provided, it must match (catches mistakes where the UI passes the wrong app).
- **Org-scoped connections**: `?app_id=` is required. The app must exist, the user must be a member of its org, *and* the app's org must match the connection's org. The third check is the key security boundary — a user in orgs A and B could otherwise pass an app from org B against a connection in org A.

The Phase 1 four endpoints (`GET/PUT/POST/DELETE /api/connections/{id}/sources`) each call this resolver once and then use the resolved `app_id` for their underlying query. Their signatures don't change.

### 8. Frontend — scope step, org-wide badge, app-scoped selector

**Files:**
- `frontend/src/types/connection.ts` — `app_id: string | null`, `org_id: string`, `app_id?: string` on the create payload
- `frontend/src/api/sources.ts` — all four functions accept optional `{ appId }` and forward as `?app_id=`
- `frontend/src/components/connections/SourceSelector.vue` — new `appId` prop threads through to every request
- `frontend/src/components/connections/wizard/steps/StepConnectionScope.vue` — **new**
- `frontend/src/components/connections/wizard/flows.ts` — scope step inserted into `webhook_logs` and the Fly.io drain flow; filtered out of Fly.io polling mode
- `frontend/src/components/connections/wizard/ConnectionWizard.vue` — `createConnection` reads `state.config.__scope` and omits `app_id` when scope is `'org'`
- `frontend/src/components/connections/ConnectionDetailModal.vue` — "Org-wide" pill next to the connection name when `app_id === null`
- `frontend/src/components/connections/ConnectionBubble.vue` — corner `ORG` tag on org-scoped bubbles; paired with a style rule
- `frontend/src/pages/ConnectionsPage.vue` — passes `appStore.currentAppId` into `SourceSelector`, so the selector edits filters for whichever app the sidebar is currently focused on
- `frontend/src/components/connections/wizard/steps/StepFlyioDrainSetup.vue` — the embedded selector there also threads `appStore.currentAppId`

**Behaviour:**

- **Wizard scope step** — a two-radio "This app only" / "Entire organisation" chooser with inline descriptions. Writes to `state.config.__scope` (a wizard-only key stripped before send). Default is `'app'` because it matches Phase 1 and is the tighter-scoped choice; the plan's "neither positioned as recommended" is still honoured — both options render identically except for which is preselected.
- **Step visibility** — only appears on flows where `supportsOrgScope` is true on the backend: webhook_logs and the Fly.io *drain* sub-mode. Fly.io polling skips it via `getFlyioSteps`, and all other platform flows (Postgres, Supabase, GitHub, Syslog, OTLP-direct) don't include the step at all.
- **Connection page** — org-scoped connections render with a small `ORG` corner tag on the bubble and an `Org-wide` pill in the detail modal header. They appear in every app's connection list because the backend query (`ListConnectionsByApp`) returns them for any app in the same org.
- **Source selector in org mode** — the same component, just driven by the currently active app's ID. Switching the sidebar's app while the selector is open effectively rescopes it on next mount; the selector doesn't itself watch `currentAppId` because the user would be surprised if their filter choices silently changed under them.

### 9. Test coverage

**Backend (`org_connections_test.go`):**

- `TestCreateConnection_OrgScoped` — POST without `app_id` creates a connection with `app_id: null`, `org_id` matching the caller's org. Confirms Postgres creates still 400 without `app_id`.
- `TestListConnectionsByApp_IncludesOrgScoped` — after creating one app-scoped and one org-scoped connection, `GET /api/apps/{appId}/connections` returns both. Org-scoped row has `app_id: null`; app-scoped row has the concrete app id.
- `TestIngestWebhookLogs_OrgFanOut` — two apps enable two different source names; a batch with both sources fans out correctly (one row per app). Response reports `accepted=2, filtered` absent (omitempty).
- `TestIngestWebhookLogs_OrgFanOutSharedSource` — both apps enable the *same* source name; one incoming entry becomes two rows in `log_buffer`, one per app — the deliberate write-time duplication.
- `TestSourceFilters_OrgScopedRequireAppID` — GET/POST on `/sources` for an org-scoped connection 400 without `?app_id=`, work with it.

All Phase 1 tests still pass unchanged.

### 10. Test-harness touches

Same two fixes I introduced in Phase 1 remain required (they weren't regressed): `users` upsert on insert (Supabase trigger coexistence) and `Listener: NewListenerManager()` (pause/resume panic guard). Nothing new for Phase 2.

---

## What Did NOT Change (and why)

### No new frontend tests

The scope step and org-wide badges are thin presentational logic on top of the existing store — covered by typecheck and the existing component tests for neighbouring components. Adding headless Vue component tests here would add more harness than signal. Manual smoke-testing in a dev browser is the right verification tier for these surfaces; the backend is where the meaningful invariants live.

### No migration of existing connections to org scope

Migration 035 backfills `org_id` from the parent app but leaves `app_id` populated. Existing connections remain app-scoped. Users opt into org-scoping per-connection by choosing it in the wizard for new drains; existing drains keep their current behaviour unless explicitly recreated.

### No `?app_id=` required for app-scoped connections

The frontend currently always passes `appId` for safety, but the backend accepts the parameter's absence on app-scoped connections (the connection's own `app_id` is authoritative). Third-party API callers with app-scoped connections don't need to change anything — they can keep calling `/sources` unadorned.

### No org-wide monitoring or agent

The monitor loop, scheduler, and agent chat are all per-application surfaces — a monitoring run targets one specific app. Org-scoped connections are an *ingestion-time* concept: they change where raw logs are routed, not what the monitoring agent reasons over. An app's monitor still sees only logs that landed in `log_buffer` with its `app_id`, which is exactly what fan-out guarantees.

### No OTLP org support

Deferred to Phase 4 when `service.name` extraction lands. OTLP ingestion explicitly rejects NULL-app-id connections today rather than silently routing logs to `uuid.Nil`.

### No Phase 3 (GitHub migration)

Phase 3 migrates `github_repos` onto the generic `connection_sources`/`app_source_filters` system. It's fully independent of Phase 2 and can land whenever; no Phase 2 decision constrained it.

---

## Risks & Follow-ups

- **Fan-out write amplification.** An org-scoped connection with 10 apps all enabling the same source produces 10× the log_buffer rows. Acceptable at current scale (single-digit apps per org) but should be monitored. The simplest mitigation if this becomes a hotspot: batch inserts (one INSERT with multiple rows per entry-app pair). Not implemented in Phase 2 — premature until there's a real signal.
- **Pre-existing RLS on `app_source_filters`** already used the org-member idiom (Phase 1 migration 034), so filter access for org-scoped connections inherits it for free. No RLS change needed for `app_source_filters` in this phase.
- **Idempotency + org-scoped**: the response cache is per `(connection_id, idempotency_key)`, so replays return the original fan-out counts even if filters change later. Same semantics as Phase 1; nothing new to think about.
- **Switching a connection between scopes** is not supported. Turning an app-scoped connection into an org-scoped one (or vice versa) would require migrating `app_source_filters` rows, invalidating the activity feed, and re-assessing monitoring state. Users can delete + recreate to achieve the effect — explicitly, since the two modes have different user mental models.

---

## Verification

- `go build ./...` — clean
- `go vet ./...` — clean
- `npx vue-tsc --noEmit` — clean
- `make migrate-up` — migration 035 applied cleanly on the local Supabase instance
- Backend tests: 13 Phase 1 + 5 Phase 2 = 18 relevant tests, all pass
- Frontend tests: 58 unit tests, all pass
- The 11 pre-existing failures (`TestListLogs_*`, `TestInviteMember*`, `TestOnboard*`) from Phase 1 are unchanged — still broken, still unrelated

---

## File Map

```
backend/
  migrations/035_org_connections.up.sql             NEW
  migrations/035_org_connections.down.sql           NEW
  sqlc.yaml                                         CHANGED — nullable uuid → *uuid.UUID
  internal/db/queries/connections.sql               CHANGED — org-member access everywhere, CreateConnection takes org_id + nullable app_id, ListAppsEnabledForSource added
  internal/db/queries/github_repos.sql              CHANGED — explicit ::UUID cast to keep the param non-null
  internal/db/**/*.sql.go                           REGENERATED (sqlc)
  internal/db/models.go                             REGENERATED — Connection.AppID is now *uuid.UUID, AgentLog/Conversation/NotificationLog nullable UUIDs switch from pgtype to pointer
  internal/notifications/notifier.go                CHANGED — pgtype.UUID → pointer
  internal/agent/emit.go                            CHANGED — pgtype.UUID → pointer
  internal/api/handlers/connections.go              CHANGED — org-scoped create path, supportsOrgScope gate, appIDOrZero helper
  internal/api/handlers/connections_test_handler.go CHANGED — appIDOrZero at every NewX call site
  internal/api/handlers/webhooks.go                 CHANGED — ingestEntries takes connResult, routesBySource fan-out, inserts counter, scope log field
  internal/api/handlers/otlp.go                     CHANGED — reject org-scoped, dereference conn.AppID
  internal/api/handlers/github_install.go           CHANGED — carry app.OrgID through CreateConnection
  internal/api/handlers/source_filters.go           CHANGED — resolveSourceFilterApp helper replaces conn.AppID usage
  internal/api/handlers/org_connections_test.go     NEW — Phase 2 test coverage
  cmd/heimdall/main.go                              CHANGED — resume paths skip NULL app_id (defensive)

frontend/
  src/types/connection.ts                                           CHANGED — Connection.app_id nullable, org_id added, payload app_id optional
  src/api/sources.ts                                                CHANGED — all four functions accept { appId }
  src/components/connections/SourceSelector.vue                     CHANGED — appId prop
  src/components/connections/wizard/steps/StepConnectionScope.vue   NEW
  src/components/connections/wizard/flows.ts                        CHANGED — scope step added to webhook_logs + Fly.io drain
  src/components/connections/wizard/ConnectionWizard.vue            CHANGED — conditionally omit app_id in createConnection
  src/components/connections/ConnectionDetailModal.vue              CHANGED — Org-wide pill
  src/components/connections/ConnectionBubble.vue                   CHANGED — ORG corner tag
  src/components/connections/wizard/steps/StepFlyioDrainSetup.vue   CHANGED — embedded selector uses appStore.currentAppId
  src/pages/ConnectionsPage.vue                                     CHANGED — selector receives appId
```

---

## Handoff

Phases 3 and 4 remain from the original plan. Phase 3 migrates `github_repos` to the generic filter system; Phase 4 extends source-name extraction to OTLP, syslog, and generic webhooks so org-scoping becomes useful for those connector types too. Neither depends on Phase 2 internals beyond the tables that shipped in Phase 1, so they can be picked up independently.
