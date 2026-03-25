# Supabase Connector — Implementation Plan

End-to-end implementation of the Supabase Management API polling connector, from backend ingestion through to frontend creation UX.

**Goal:** A user can create a Supabase connection in Heimdall, and logs from their Supabase project flow into `log_buffer` automatically — where the existing monitoring loop classifies and escalates them.

---

## Phase 1: Backend Connector

Build the core polling connector that calls the Supabase Management API and inserts logs into `log_buffer`.

### Step 1.1 — `PollConnector` interface

**File:** `backend/internal/connectors/connector.go`

Add a new interface alongside the existing `StreamConnector` and `QueryConnector`:

```go
type PollConnector interface {
    Connector
    Poll(ctx context.Context, queries *db.Queries) error
}
```

**Why a new interface:** The existing `StreamConnector.Stream()` assumes push-based data flowing into a channel. Polling is fundamentally pull-based — the connector initiates HTTP requests on a timer. `Poll()` takes `*db.Queries` so it can call `InsertLogEntry` directly, keeping the insertion logic inside the connector rather than requiring an external orchestrator to shuttle data.

### Step 1.2 — Supabase connector implementation

**File:** `backend/internal/connectors/logs/supabase.go`

Struct and constructor:

```go
type SupabaseConfig struct {
    ProjectRef       string   `json:"project_ref"`
    AccessToken      string   `json:"access_token"`
    PollTables       []string `json:"poll_tables"`
    PollIntervalSecs int      `json:"poll_interval_secs"`
}

type Supabase struct {
    config       SupabaseConfig
    connectionID uuid.UUID
    userID       uuid.UUID
    httpClient   *http.Client
    cursors      map[string]time.Time  // per-table cursor
}
```

Methods to implement:

| Method | Behaviour |
|--------|-----------|
| `New(configJSON, connectionID, userID)` | Parse config, validate required fields, set defaults (`poll_tables` → `["postgres_logs"]`, `poll_interval_secs` → `30`, min `15`) |
| `Connect(ctx)` | Validate PAT by calling `GET /v1/projects/{ref}/analytics/endpoints/logs.all?sql=SELECT 1`. Return error if 401/403/404. |
| `Health(ctx)` | Same as `Connect` — a lightweight API ping. |
| `Poll(ctx, queries)` | For each table in `poll_tables`: query logs since cursor, parse response, insert into `log_buffer`, advance cursor. |
| `Close()` | No-op (stateless HTTP client). |

**API call details for `Poll`:**

```
GET https://api.supabase.com/v1/projects/{ref}/analytics/endpoints/logs.all
  ?sql=SELECT timestamp, event_message, metadata FROM {table} WHERE timestamp > '{cursor}' ORDER BY timestamp ASC LIMIT 500
```

Headers: `Authorization: Bearer {access_token}`

Response shape (JSON):
```json
{
  "result": [
    { "timestamp": 1711036800000000, "event_message": "...", "metadata": {...} }
  ]
}
```

**Insertion mapping:**

| `log_buffer` column | Value |
|---------------------|-------|
| `connection_id` | The Supabase connection's UUID |
| `source_type` | `"supabase/{table}"` e.g. `"supabase/postgres_logs"` |
| `severity` | Derive from metadata if available, otherwise `null` |
| `payload` | Full row as JSONB: `{ "timestamp": ..., "event_message": ..., "metadata": ... }` |
| `user_id` | The connection owner's UUID |

**Cursor management:**

- In-memory map of `table_name → last_timestamp` (microsecond Unix).
- Initialise cursor to `now() - 5 minutes` on first poll (avoids pulling the full 24h window).
- After each successful poll, advance to the max timestamp seen.
- Cursor is ephemeral (in-memory only for Phase 1). If the process restarts, it re-polls from `now() - 5 minutes`, which may cause some duplication but won't miss logs.

**Rate limit handling:**

- Read `X-RateLimit-Remaining` from response headers.
- If `0`, read `X-RateLimit-Reset` (Unix timestamp) and sleep until then.
- On HTTP 429, back off for the reset window.
- Log rate limit events at `slog.Warn` level.

### Step 1.3 — Polling loop manager

**File:** `backend/internal/connectors/poller.go`

A manager that runs one goroutine per active poll-based connection:

```go
type Poller struct {
    queries    *db.Queries
    mu         sync.Mutex
    pollers    map[uuid.UUID]context.CancelFunc  // connectionID → cancel
}
```

Methods:

| Method | Behaviour |
|--------|-----------|
| `NewPoller(queries)` | Constructor |
| `Start(conn PollConnector, connectionID, interval)` | Launch a goroutine that calls `conn.Poll()` on a ticker. Store cancel func. |
| `Stop(connectionID)` | Cancel the goroutine for this connection. |
| `StopAll()` | Cancel all pollers (called on server shutdown). |

The goroutine loop:

```go
func (p *Poller) run(ctx context.Context, conn PollConnector, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    // Run once immediately, then on each tick.
    conn.Poll(ctx, p.queries)
    for {
        select {
        case <-ticker.C:
            if err := conn.Poll(ctx, p.queries); err != nil {
                slog.Error("poll failed", "err", err)
            }
        case <-ctx.Done():
            return
        }
    }
}
```

### Step 1.4 — Wire into server startup

**File:** `backend/cmd/heimdall/main.go`

- Create `Poller` instance after DB pool init.
- Pass `Poller` to `Server` (add field to `handlers.Server`).
- On shutdown, call `poller.StopAll()` before `pool.Close()`.

**File:** `backend/internal/api/handlers/server.go`

- Add `Poller *connectors.Poller` field to `Server` struct.
- Update `NewServer()` signature to accept `*connectors.Poller`.

### Step 1.5 — Connection handler updates

**File:** `backend/internal/api/handlers/connections.go`

**`TestConnection`** — add `case "supabase"`:

```go
case "supabase":
    sb, err := logs.NewSupabase(conn.Config, conn.ID, userID)
    if err != nil {
        result = testResult{Success: false, Message: fmt.Sprintf("Invalid config: %v", err)}
        break
    }
    ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
    defer cancel()
    if err := sb.Connect(ctx); err != nil {
        result = testResult{Success: false, Message: fmt.Sprintf("Failed to connect to Supabase API: %v", err)}
    } else {
        result = testResult{Success: true, Message: "Connected to Supabase Management API"}
    }
```

**`CreateConnection`** — after successful creation, if type is `"supabase"` and a test passes, start the poller:

```go
if req.Type == "supabase" {
    sb, err := logs.NewSupabase(config, conn.ID, userID)
    if err == nil {
        interval := 30 // default
        if cfg, _ := sb.ParsedConfig(); cfg.PollIntervalSecs > 0 {
            interval = cfg.PollIntervalSecs
        }
        s.Poller.Start(sb, conn.ID, time.Duration(interval)*time.Second)
    }
}
```

**`DeleteConnection`** — stop the poller if the connection is `supabase`:

```go
// Before deleting, stop any active poller.
s.Poller.Stop(connID)
```

### Step 1.6 — Resume pollers on server restart

**File:** `backend/cmd/heimdall/main.go` (or a new `boot.go` in handlers)

On startup, query all active Supabase connections and start their pollers:

```go
func (s *Server) ResumePollers(ctx context.Context) {
    // Query all connections where type = 'supabase' AND status = 'active'
    // For each, create a Supabase connector and call s.Poller.Start(...)
}
```

**New sqlc query needed** in `connections.sql`:

```sql
-- name: ListActiveConnectionsByType :many
SELECT * FROM connections WHERE type = $1 AND status = 'active';
```

Run `make sqlc-generate` after adding the query.

---

## Phase 2: Frontend — Connection Creation

Allow users to create and manage Supabase connections through the existing UI patterns.

### Step 2.1 — Add `supabase` type to ConnectionForm

**File:** `frontend/src/components/connections/ConnectionForm.vue`

Add `supabase` to the type options array:

```typescript
{ value: 'supabase', label: 'Supabase' }
```

Add config field definitions:

```typescript
supabase: [
  { key: 'project_ref', label: 'Project Reference', type: 'text', required: true,
    placeholder: 'e.g. abcdefghijklmnopqrst' },
  { key: 'access_token', label: 'Personal Access Token', type: 'password', required: true,
    placeholder: 'sbp_...' },
  { key: 'poll_interval_secs', label: 'Poll Interval', type: 'select', required: false,
    default: 30,
    options: [
      { value: 15, label: '15 seconds' },
      { value: 30, label: '30 seconds' },
      { value: 60, label: '60 seconds' },
    ] },
]
```

Set direction to `one_way` automatically for `supabase` type (it's ingestion-only).

### Step 2.2 — Poll tables multi-select

**File:** `frontend/src/components/connections/ConnectionForm.vue`

Add a checkbox group for `poll_tables` when type is `supabase`. This is not a standard `BaseSelect` — it's a list of checkboxes since the user selects multiple tables.

Available tables with descriptions:

| Table | Label |
|-------|-------|
| `postgres_logs` | Database queries and errors |
| `auth_logs` | Authentication events |
| `edge_logs` | API gateway requests |
| `function_logs` | Edge Function output |
| `storage_logs` | Object storage operations |
| `realtime_logs` | WebSocket connections |

Default: `postgres_logs` and `auth_logs` checked.

Build the config payload:

```typescript
config: {
  project_ref: form.project_ref,
  access_token: form.access_token,
  poll_tables: selectedTables,        // string[]
  poll_interval_secs: form.poll_interval_secs,  // number
}
```

### Step 2.3 — ConnectionCard display

**File:** `frontend/src/components/connections/ConnectionCard.vue`

Ensure the card renders `supabase` type with:

- Label: "Supabase"
- Subtitle: show polled tables count, e.g. "Polling 2 tables"
- Status badge works as-is (active/inactive/error)
- Ping action triggers the existing `TestConnection` flow

### Step 2.4 — Helper text in form

Add a brief info note below the access token field:

> Uses the Supabase Management API — works on all plans including Free. Generate a Personal Access Token at **supabase.com/dashboard/account/tokens**.

Style as an info callout matching the existing design system (`text-text-tertiary text-xs`).

---

## Phase 3: End-to-End Validation

Verify the full pipeline works: creation → polling → ingestion → classification → escalation.

### Step 3.1 — Manual integration test

Test with a real Supabase project:

1. Create a Supabase connection in the UI with a valid PAT and project ref.
2. Verify the test passes (connection status → `active`).
3. Wait one polling interval (30s).
4. Check `log_buffer` for entries with `source_type = 'supabase/postgres_logs'`.
5. Check the Agent Log page for heartbeat entries that include the polled logs.
6. Trigger a database error in the Supabase project and verify it gets classified and escalated.

### Step 3.2 — Error scenarios

Test each failure mode:

| Scenario | Expected behaviour |
|----------|-------------------|
| Invalid PAT | Test fails with "401 Unauthorized" message |
| Invalid project ref | Test fails with "404 Not Found" message |
| PAT expires mid-polling | Connection status → `error`, poller continues retrying |
| Rate limit hit | Poller backs off, logs warning, resumes after reset window |
| Supabase API outage | Poller logs error, continues on next tick |
| Delete connection | Poller stops immediately, logs removed (cascade) |
| Server restart | Pollers resume for all active Supabase connections |

### Step 3.3 — Backend unit tests

**File:** `backend/internal/connectors/logs/supabase_test.go`

- `TestNewSupabase_ValidConfig` — parses all fields correctly, applies defaults
- `TestNewSupabase_MissingRequired` — errors on missing `project_ref` or `access_token`
- `TestNewSupabase_PollIntervalMinimum` — clamps interval to 15s minimum
- `TestSupabase_ParseResponse` — correctly maps API response rows to `InsertLogEntry` params
- `TestSupabase_CursorAdvancement` — cursor advances to max timestamp after poll

**File:** `backend/internal/connectors/poller_test.go`

- `TestPoller_StartStop` — goroutine starts and stops cleanly
- `TestPoller_StopAll` — all goroutines stop on shutdown

### Step 3.4 — Frontend unit tests

**File:** `frontend/src/components/connections/ConnectionForm.test.ts`

- Test that selecting `supabase` type shows the correct config fields
- Test that `poll_tables` defaults to `["postgres_logs", "auth_logs"]`
- Test that the form payload includes all Supabase-specific config

---

## Phase 4: Connection Wizard

Replace the dropdown-based form with a guided multi-step wizard for creating new connections. This is a pure frontend change — the backend and API are identical.

### Step 4.1 — Wizard shell

**File:** `frontend/src/components/connections/wizard/ConnectionWizard.vue`

Full-width modal with:

- Step indicator (dot-line progress bar)
- Step content area
- Back / Continue / Done buttons
- Local reactive `WizardState` (not Pinia — ephemeral, scoped to modal)
- Close button with confirmation if partially filled

### Step 4.2 — Step indicator

**File:** `frontend/src/components/connections/wizard/WizardStepIndicator.vue`

The `● ─── ○ ─── ○` progress bar. Props: `steps` (label array), `currentIndex`. Active dot: `bg-accent`, completed: `bg-accent/60`, future: `bg-border`.

### Step 4.3 — Platform grid

**Files:**
- `frontend/src/components/connections/wizard/PlatformGrid.vue`
- `frontend/src/components/connections/wizard/PlatformCard.vue`

Card grid grouped by category (Log Sources, Databases, Generic). Each card: icon/letter badge, name, one-line description. Unavailable cards show "Coming soon" badge with `opacity-40 cursor-not-allowed`.

Initially available cards: **Supabase**, **PostgreSQL**, **Webhook**, **GitHub**. All others marked "Coming soon".

### Step 4.4 — Flow definitions

**File:** `frontend/src/components/connections/wizard/flows.ts`

Declarative array of `PlatformFlow` objects. Supabase flow:

```typescript
{
  id: 'supabase',
  name: 'Supabase',
  icon: 'SB',
  description: 'Database logs, auth events, edge functions',
  category: 'log_source',
  connectorType: 'supabase',
  available: true,
  steps: [
    { id: 'name', label: 'Name', component: StepName },
    { id: 'auth', label: 'Auth', component: StepSupabaseAuth },
    { id: 'tables', label: 'Tables', component: StepSupabaseTables },
    { id: 'test', label: 'Test', component: StepTest },
  ],
}
```

Also define flows for the existing types (postgres, webhook_logs, github) so the wizard fully replaces the form for creation.

### Step 4.5 — Shared step components

| File | Purpose |
|------|---------|
| `wizard/steps/StepName.vue` | Connection name input (shared across all flows) |
| `wizard/steps/StepTest.vue` | Connection test — reuse logic from `ConnectionTestModal` |

### Step 4.6 — Supabase-specific step components

| File | Purpose |
|------|---------|
| `wizard/steps/StepSupabaseAuth.vue` | Project ref + PAT inputs, info callout about Management API |
| `wizard/steps/StepSupabaseTables.vue` | Table checkboxes + poll interval presets |

### Step 4.7 — Existing type step components

| File | Purpose |
|------|---------|
| `wizard/steps/StepPostgresConfig.vue` | Host/port/db/user/pass/ssl — extracted from ConnectionForm |
| `wizard/steps/StepWebhookSetup.vue` | Shows generated endpoint URL + bearer token + payload format |
| `wizard/steps/StepGitHubInstall.vue` | Wraps existing GitHub App OAuth redirect |

### Step 4.8 — CopyableField component

**File:** `frontend/src/components/common/CopyableField.vue`

Reusable component for values users need to copy (webhook URLs, tokens). Clipboard button with "Copied" feedback. Used by `StepWebhookSetup`.

### Step 4.9 — Wire wizard into ConnectionsPage

**File:** `frontend/src/pages/ConnectionsPage.vue`

- "New Connection" button opens `ConnectionWizard` modal instead of showing inline form.
- Existing inline `ConnectionForm` preserved for **editing** existing connections.
- Wizard emits `created` event → refresh connection list.

### Step 4.10 — Transition animations

Add slide-left/slide-right transitions between wizard steps via Vue `<Transition>`. Match the existing fade+slide pattern from `BaseSelect`.

---

## Dependency Graph

```
Phase 1                    Phase 2              Phase 3          Phase 4
─────────                  ─────────            ─────────        ─────────
1.1 PollConnector ─┐
                    ├→ 1.3 Poller ─→ 1.4 Wire ─→ 2.1 Form ─→ 3.1 E2E ─→ 4.1 Wizard shell
1.2 Supabase ──────┘       │                      2.2 Tables    3.2 Errors   4.2 Indicator
                           │                      2.3 Card      3.3 Tests    4.3 Grid
                           └→ 1.5 Handlers         2.4 Help     3.4 FE tests 4.4 Flows
                              1.6 Resume                                      4.5–4.10 Steps
```

Phases 1 and 2 are sequential (backend must exist before frontend can call it). Phase 3 runs after Phase 2. Phase 4 is independent of Phase 3 — it can start as soon as Phase 2 is done, or be deferred entirely.

---

## Files Created / Modified Summary

### New files

| File | Phase |
|------|-------|
| `backend/internal/connectors/logs/supabase.go` | 1.2 |
| `backend/internal/connectors/logs/supabase_test.go` | 3.3 |
| `backend/internal/connectors/poller.go` | 1.3 |
| `backend/internal/connectors/poller_test.go` | 3.3 |
| `frontend/src/components/connections/wizard/ConnectionWizard.vue` | 4.1 |
| `frontend/src/components/connections/wizard/WizardStepIndicator.vue` | 4.2 |
| `frontend/src/components/connections/wizard/PlatformGrid.vue` | 4.3 |
| `frontend/src/components/connections/wizard/PlatformCard.vue` | 4.3 |
| `frontend/src/components/connections/wizard/flows.ts` | 4.4 |
| `frontend/src/components/connections/wizard/steps/StepName.vue` | 4.5 |
| `frontend/src/components/connections/wizard/steps/StepTest.vue` | 4.5 |
| `frontend/src/components/connections/wizard/steps/StepSupabaseAuth.vue` | 4.6 |
| `frontend/src/components/connections/wizard/steps/StepSupabaseTables.vue` | 4.6 |
| `frontend/src/components/connections/wizard/steps/StepPostgresConfig.vue` | 4.7 |
| `frontend/src/components/connections/wizard/steps/StepWebhookSetup.vue` | 4.7 |
| `frontend/src/components/connections/wizard/steps/StepGitHubInstall.vue` | 4.7 |
| `frontend/src/components/common/CopyableField.vue` | 4.8 |

### Modified files

| File | Phase | Change |
|------|-------|--------|
| `backend/internal/connectors/connector.go` | 1.1 | Add `PollConnector` interface |
| `backend/internal/api/handlers/server.go` | 1.4 | Add `Poller` field |
| `backend/internal/api/handlers/connections.go` | 1.5 | Supabase test/create/delete handling |
| `backend/cmd/heimdall/main.go` | 1.4, 1.6 | Create poller, resume on startup, shutdown |
| `backend/internal/db/queries/connections.sql` | 1.6 | Add `ListActiveConnectionsByType` query |
| `frontend/src/components/connections/ConnectionForm.vue` | 2.1, 2.2 | Add supabase type + config fields |
| `frontend/src/components/connections/ConnectionCard.vue` | 2.3 | Supabase display |
| `frontend/src/pages/ConnectionsPage.vue` | 4.9 | Wire wizard modal |
