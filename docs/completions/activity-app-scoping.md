# Activity Feed App-Scoping — Implementation Plan

**Status:** Not started
**Owner:** TBD
**Prereqs:** None — all tables, queries, and UI already exist

---

## Why this matters

Heimdall treats applications as first-class monitoring targets — every other
surface (Connections, Agent Config, Scheduled Investigations, Notifications)
is scoped to the currently selected app. The Activity feed is the exception.

Both underlying tables (`log_buffer` and `agent_log`) are queried by
`user_id` only. When a user switches apps in the sidebar, the Activity feed
doesn't change — it shows a flat, org-wide merge of all logs across every
application. For a user monitoring three services (e.g. `api`, `worker`,
`dashboard`), the feed becomes a noisy interleave that defeats the purpose
of per-app isolation.

### Current state

| Table | Scoping column | How app is known |
|-------|---------------|-----------------|
| `log_buffer` | `connection_id` (FK → `connections`) | Join path: `log_buffer.connection_id → connections.app_id` |
| `agent_log` | none on the row | `app_id` is embedded in the `detail` JSONB for monitoring/scheduled entries, but absent for interactive chat entries |

The data to scope logs exists — it's just not used at query time.

---

## Design decisions

### Direct column vs. join

**`log_buffer`** already has `connection_id`, and `connections` already has
`app_id`. We *could* join through `connections` at query time, but this adds
complexity to every query and makes the `log_buffer` table's app affinity
implicit. A denormalised `app_id` column on `log_buffer` is simpler: one
`WHERE` clause, one index, no join. The trade-off (data duplication) is
minimal — `app_id` never changes for a given connection.

**`agent_log`** has no connection to apps at the schema level. The `detail`
JSONB sometimes contains `app_id`, but only for monitoring and scheduled
investigation entries. Interactive chat entries (`observation`, `tool_call`,
`tool_result`) have no app reference at all. A new `app_id` column is
required.

**Decision:** Add a nullable `app_id UUID REFERENCES applications(id)` column
to both `log_buffer` and `agent_log`. Nullable because:
- Historical rows predate the column and cannot be backfilled reliably.
- Some agent_log entry types (e.g. system-level heartbeats) may genuinely
  have no app context.

### Query behaviour

When `app_id` is provided in the API request:
- Filter both tables by `app_id` (in addition to `user_id`).
- Rows with `NULL` app_id are **excluded** — they belong to no app and
  shouldn't appear in a per-app view.

When `app_id` is omitted (or a future "All apps" toggle):
- Current behaviour — return all rows for the user, regardless of app.

### Frontend contract

The Activity page will send `app_id` as a query parameter, sourced from
`useAppStore().currentAppId`. The logs store already passes filter params
through to the API — `app_id` is just another one.

---

## Part 1 — Database migration

### Migration 025: Add `app_id` to `log_buffer` and `agent_log`

**Up:**
```sql
-- log_buffer: denormalised app_id for direct filtering.
ALTER TABLE log_buffer
  ADD COLUMN app_id UUID REFERENCES applications(id) ON DELETE CASCADE;

-- Backfill from the connection's app_id for existing rows.
UPDATE log_buffer lb
  SET app_id = c.app_id
  FROM connections c
  WHERE lb.connection_id = c.id;

-- agent_log: new app_id column.
ALTER TABLE agent_log
  ADD COLUMN app_id UUID REFERENCES applications(id) ON DELETE CASCADE;

-- Backfill agent_log where detail contains app_id.
UPDATE agent_log
  SET app_id = (detail->>'app_id')::uuid
  WHERE detail IS NOT NULL
    AND detail->>'app_id' IS NOT NULL;

-- Indexes for the new filter path.
CREATE INDEX idx_log_buffer_app_id ON log_buffer (app_id, ingested_at DESC)
  WHERE app_id IS NOT NULL;
CREATE INDEX idx_agent_log_app_id ON agent_log (app_id, created_at DESC)
  WHERE app_id IS NOT NULL;
```

**Down:**
```sql
DROP INDEX IF EXISTS idx_agent_log_app_id;
DROP INDEX IF EXISTS idx_log_buffer_app_id;
ALTER TABLE agent_log DROP COLUMN IF EXISTS app_id;
ALTER TABLE log_buffer DROP COLUMN IF EXISTS app_id;
```

### Notes

- The `UPDATE ... FROM connections` backfill is safe because `log_buffer`
  already has a NOT NULL `connection_id` FK. Every row will match.
- The `agent_log` backfill is best-effort — only monitoring and scheduled
  investigation entries have `app_id` in their detail. Interactive chat
  entries will remain `NULL` until Part 3 addresses write-time population.
- Partial indexes (`WHERE app_id IS NOT NULL`) avoid indexing historical
  NULLs that will never match a filtered query.

---

## Part 2 — Backend: queries and handler

### 2a. New sqlc queries

**`log_buffer.sql`** — add app-scoped variants:

```sql
-- name: ListLogsByApp :many
SELECT * FROM log_buffer
WHERE user_id = $1 AND app_id = $2
ORDER BY ingested_at DESC
LIMIT $3 OFFSET $4;

-- name: CountLogsByApp :one
SELECT count(*) FROM log_buffer
WHERE user_id = $1 AND app_id = $2;

-- name: ListLogsByAppAndSeverity :many
SELECT * FROM log_buffer
WHERE user_id = $1 AND app_id = $2 AND severity = $3
ORDER BY ingested_at DESC
LIMIT $4 OFFSET $5;

-- name: CountLogsByAppAndSeverity :one
SELECT count(*) FROM log_buffer
WHERE user_id = $1 AND app_id = $2 AND severity = $3;
```

**`agent_log.sql`** — add app-scoped variant:

```sql
-- name: ListAgentLogByApp :many
SELECT * FROM agent_log
WHERE user_id = $1 AND app_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountAgentLogByApp :one
SELECT count(*) FROM agent_log
WHERE user_id = $1 AND app_id = $2;
```

### 2b. Handler changes (`logs.go`)

1. Parse optional `app_id` query parameter (validated as UUID).
2. When present, use the `*ByApp` query variants instead of the `*ByUser`
   variants. The `authorizeApp` helper should validate ownership.
3. The merge logic (fetch both sources, sort, paginate) is unchanged — only
   the underlying queries change.

### 2c. Run `make sqlc-generate`

Regenerate Go code after adding the new queries.

---

## Part 3 — Backend: populate `app_id` at write time

Every code path that inserts into `log_buffer` or `agent_log` must set
`app_id` so new rows are always scoped.

### `log_buffer` inserts

**`InsertLogEntry`** — add `app_id` parameter. Callers:

| Caller | Where `app_id` comes from |
|--------|--------------------------|
| Supabase poller (`supabase.go`) | `connection_id` → look up `connections.app_id` at poller construction time, store on the struct |
| Fly.io / Vercel / Railway / MongoDB pollers | Same pattern — store `app_id` on the connector struct |
| Webhook ingestion handler | `GetConnectionByWebhookToken` already returns the full connection row, which has `app_id` |
| Syslog listener | Same — connection row is loaded at listener start |
| OTLP receiver | Same — connection row is loaded at receiver start |

Each poll-based connector already receives `connectionID` and `userID` at
construction. Add `appID uuid.UUID` as a third identity field, sourced from
the connection row when the poller is started (both in `StartPoller` and
`resumePollers`).

### `agent_log` inserts

**`InsertAgentLog`** — add `app_id` parameter. Callers:

| Caller | Where `app_id` comes from |
|--------|--------------------------|
| `EmitLog` / `EmitLogWithSeverity` | Add `appID *uuid.UUID` parameter |
| Monitoring mode (`monitor.go`) | `app.ID` is already in scope in the per-app loop |
| Scheduled investigations (`scheduler.go`) | `s.AppID` is already available |
| Interactive chat (`loop.go`) | The WebSocket handler receives `app_id` from the query param (the chat is already app-scoped via the frontend's `currentAppId`) |

### `EmitLog` signature change

```go
// Before
func (a *Agent) EmitLog(ctx context.Context, userID uuid.UUID,
    conversationID *uuid.UUID, entryType, summary string,
    detail map[string]any) uuid.UUID

// After
func (a *Agent) EmitLog(ctx context.Context, userID uuid.UUID,
    appID *uuid.UUID, conversationID *uuid.UUID,
    entryType, summary string, detail map[string]any) uuid.UUID
```

All existing call sites already have app context available (or can pass
`nil` for the rare cases where no app applies).

---

## Part 4 — Frontend: pass `app_id` to the API

### 4a. API client (`api/logs.ts`)

Add `app_id` to the `listLogs` params interface:

```ts
export function listLogs(params?: {
  app_id?: string       // ← new
  severity?: string
  connection_id?: string
  source?: string
  limit?: number
  offset?: number
})
```

### 4b. Logs store (`stores/logs.ts`)

Import `useAppStore` and include `currentAppId` in every fetch:

```ts
async function fetchLogs(params?: { severity?: string; connection_id?: string }) {
  const appStore = useAppStore()
  // ...
  const { data } = await logsApi.listLogs({
    ...params,
    app_id: appStore.currentAppId ?? undefined,
    source: source.value,
    limit: limit.value,
    offset: offset.value,
  })
}
```

### 4c. Activity page (`ActivityPage.vue`)

Watch `appStore.currentAppId` and re-fetch when it changes (same pattern
used by `ConnectionsPage` and `SchedulesPage`):

```ts
const appStore = useAppStore()

watch(() => appStore.currentAppId, () => {
  logsStore.resetPagination()
  logsStore.fetchLogs()
})
```

### 4d. LogDetailModal connection name resolution

The modal already receives the connections list. No change needed — the
connection names are resolved from the store which is already app-scoped.

---

## Part 5 — Verification

### Backend tests

- `ListLogs` handler test: verify that passing `app_id` returns only logs
  for that app, and omitting it returns all.
- `InsertLogEntry` / `InsertAgentLog`: verify `app_id` is persisted.

### Frontend tests

- Logs store test: verify `app_id` is included in the API call params.

### Manual verification

1. Create two apps, each with a Supabase connection polling different tables.
2. Wait for logs to ingest.
3. Switch between apps in the sidebar — Activity feed should show only logs
   from the selected app's connections.
4. Agent monitoring/scheduled investigation entries should also filter
   correctly.
5. Interactive chat entries (with `app_id` set via WebSocket) should appear
   under the correct app.

---

## Risk and rollback

**Risk:** The migration backfills existing rows, which involves a full table
scan of `log_buffer` and `agent_log`. On a small dataset (Heimdall's current
scale) this is instantaneous. On a larger dataset, consider batching the
`UPDATE`.

**Rollback:** The down migration drops the column entirely. No data loss
beyond the denormalised `app_id` values, which can be re-derived from
`connections.app_id` and `agent_log.detail->>'app_id'` at any time.

**Backwards compatibility:** The `app_id` query parameter is optional. If
omitted, the handler falls back to user-scoped queries — existing API
consumers are unaffected.

---

## Appendix — Org-level integrations and per-app scoping

### The problem

Some connections are inherently **org-level** on the provider's side. A
GitHub App installation is tied to a GitHub organisation — it grants access
to all (or selected) repos in that org. Supabase projects, similarly, are
a single entity that spans multiple Heimdall apps.

Heimdall's connection model is strictly per-app: `connections.app_id` is
`NOT NULL`. When a user creates App B and wants the same GitHub org, the
current flow redirects them to GitHub's OAuth screen. GitHub says "Heimdall
is already installed" — the user either re-authorises (confusing) or backs
out (stuck). Even when it works, the result is a second connection row with
the same `installation_id`, which is invisible duplication.

This is a pre-existing design tension — the app-scoping refactoring in
Parts 1–5 doesn't make it worse, but it does make it more visible because
users will interact with per-app views more deliberately.

### Affected connection types

| Type | Provider scope | Per-app what? |
|------|---------------|---------------|
| GitHub App | Org/account installation (one per GitHub org) | Which repos to enable |
| Supabase | Project (one project_ref, one access token) | Which tables to poll |
| Postgres | Database (one connection string) | Which schemas/queries to allow |
| Webhook | Per-endpoint token | Always 1:1 with app — no sharing issue |
| Fly.io / Vercel / Railway | Account/project API token | Which app/service to poll |

GitHub is the clearest case. Supabase is a grey area — the same project
could serve multiple Heimdall apps if they care about different tables.

### Recommended approach: "Link existing installation"

Keep the current per-app connection model (no schema changes), but add a
**shortcut flow** in the connection wizard that detects existing
installations within the same org and offers to reuse them.

#### How it works

**Backend — new query:**

```sql
-- name: ListInstallationsByOrgAndType :many
SELECT DISTINCT ON (config->>'installation_id')
  id, config, name, type
FROM connections
WHERE user_id = $1
  AND type = $2
  AND status = 'active'
ORDER BY config->>'installation_id', created_at ASC;
```

**Backend — new endpoint:**

```
GET /api/connections/available-installations?type=github
```

Returns installations already connected to *any* app in the user's org.
Response shape:

```json
[
  {
    "installation_id": 67890,
    "account_login": "acme-corp",
    "account_type": "Organization",
    "connected_apps": ["API Service", "Worker"]
  }
]
```

**Frontend — wizard branching:**

When the user selects "GitHub" in the connection wizard:

1. Call `GET /api/connections/available-installations?type=github`
2. **If installations exist**, show a choice screen:

```
┌─────────────────────────────────────────────────┐
│  GitHub Connection                              │
│                                                 │
│  Your organisation already has GitHub connected: │
│                                                 │
│  ┌─────────────────────────────────────────┐    │
│  │ ● Link existing: acme-corp             │    │
│  │   Already used by: API Service, Worker  │    │
│  └─────────────────────────────────────────┘    │
│                                                 │
│  ┌─────────────────────────────────────────┐    │
│  │ ○ Connect a different GitHub org        │    │
│  │   Opens GitHub App installation flow    │    │
│  └─────────────────────────────────────────┘    │
│                                                 │
│                              [Continue →]       │
└─────────────────────────────────────────────────┘
```

3. **"Link existing"** → skip OAuth entirely. Create a new connection row
   with the same `installation_id` config, scoped to the current app.
   Proceed directly to the repo picker (for GitHub) or table picker (for
   Supabase).

4. **"Connect a different org"** → normal OAuth flow. For cases where the
   user genuinely has a second GitHub org.

5. **If no installations exist**, skip the choice screen and go straight
   to OAuth. The user sees no difference from today.

#### What this creates in the database

```
Heimdall Org: acme-corp-team
├─ App A (api-service)
│   └─ Connection: GitHub (conn-1, installation_id=67890)
│       └─ Enabled repos: acme-corp/api, acme-corp/shared-lib
│
└─ App B (worker)
    └─ Connection: GitHub (conn-2, installation_id=67890)  ← same install
        └─ Enabled repos: acme-corp/worker, acme-corp/shared-lib
```

Two connection rows, same `installation_id`, different `app_id`, different
enabled repos. Each app sees only its own repos in codebase search.

#### Why not a shared connection table?

An `installations` table with a many-to-many join to `connections` would be
cleaner relationally, but:

1. Every handler that touches connections would need updating.
2. The `config` JSONB would need splitting (installation-level fields vs.
   app-level fields like `poll_tables`).
3. The poller/listener resume logic assumes one config per connection.
4. The benefit is marginal — `installation_id` duplication across 2–5 rows
   costs nothing at Heimdall's scale.

The "link existing" approach solves the UX problem (no confusing re-auth)
without touching the data model.

#### Uninstall handling

If the user uninstalls the GitHub App from GitHub's side, **all connections
sharing that `installation_id` break simultaneously**. The connection health
check (`Connect()`) will fail for each one. This is correct behaviour — the
installation is gone — but the UI should surface it clearly:

- The Connections page already shows connection status badges.
- Consider a banner: "This GitHub installation was removed. Reconnect to
  restore access." on each affected app's Connections page.

#### Applicability to other types

The same "link existing" pattern works for Supabase (same `project_ref` +
`access_token`, different `poll_tables`) and Postgres (same connection
string, different query scope). The wizard just needs type-specific
detection logic:

| Type | Shared key | Per-app config |
|------|-----------|----------------|
| GitHub | `installation_id` | Enabled repos |
| Supabase | `project_ref` | `poll_tables`, `poll_interval_secs` |
| Postgres | `host` + `port` + `database` | Read-only query scope |

#### Implementation priority

This is a **separate work item** from the Activity feed app-scoping (Parts
1–5). The Activity scoping is a correctness fix — logs must filter by app.
The "link existing installation" flow is a UX improvement that becomes more
valuable *after* app-scoping lands, because users will create more apps
once per-app views actually work properly.
