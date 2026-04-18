# Connection Pause / Resume

## Problem

Users want to temporarily disable a connection without deleting it. Use cases:

- **Maintenance windows** — pause ingestion from a noisy source during a deploy, resume after.
- **Cost control** — stop polling a third-party API you're not actively investigating.
- **Debugging** — isolate whether a connection is the source of bad data without losing its config.
- **Seasonal use** — keep a staging connection configured but dormant until needed.

Today the only way to "stop" a connection is to delete it and re-create it later, losing the webhook token, config, and history linkage.

## Current State

### Status field

The `connections` table already has a `status TEXT DEFAULT 'inactive'` column with three valid values enforced in application code:

| Value | Meaning |
|-------|---------|
| `active` | Connection is live — pollers run, webhooks accepted, agent tools available |
| `inactive` | Freshly created, not yet tested/activated |
| `error` | Last health check or poller init failed |

### Where status is checked today

| Location | File | What it does |
|----------|------|-------------|
| Webhook ingestion | `handlers/webhooks.go:156` | `GetConnectionByWebhookToken` SQL filters `status = 'active'` — rejects non-active connections with 401 |
| OTLP ingestion | `handlers/otlp.go:75` | Same query, same filter |
| Poller resume (startup) | `cmd/heimdall/main.go:202-219` | `ListActiveConnectionsByType` only returns `status = 'active'` rows |
| Syslog listener resume | `cmd/heimdall/main.go:159-200` | Same query |
| Update handler | `handlers/connections.go:350-373` | Always restarts poller/listener after update, regardless of status |
| Agent `query_database` | `agent/tools_db.go:36` | Fetches connection by user — **no status check** |
| Agent `search_logs` | `agent/tools_logs.go` | Queries `log_buffer` — no connection-level filter |
| Monitor loop | `agent/monitor.go` | Processes all logs for an app — no connection-level filter |
| Frontend | `ConnectionBubble.vue`, `ConnectionDetailModal.vue` | Visual treatment for `active` / `inactive` / `error` only |

### Key observation

The existing `status` field + the `active`-only SQL filters already do most of the work. A paused connection that has `status = 'paused'` would automatically:
- Be rejected by webhook/OTLP ingestion (the SQL `WHERE status = 'active'` excludes it)
- Not have its poller resumed on server restart
- Not have its syslog listener resumed on server restart

The gaps are:
1. **Validation** — `isValidConnectionStatus()` rejects `'paused'` as invalid.
2. **Update handler** — blindly restarts pollers/listeners even when status is paused.
3. **Agent tools** — `query_database` has no status gate; a paused DB connection can still be queried.
4. **Frontend** — no visual treatment for paused, no pause/resume button in UI.
5. **Semantic clarity** — need to distinguish "paused by user" from "inactive (never activated)" and "error (broken)".

---

## Design Decision: New Status Value vs. Boolean Column

### Option A — Add `'paused'` to the status enum (recommended)

Add `paused` as a fourth valid status value. The existing `WHERE status = 'active'` filters already exclude it everywhere that matters.

**Pros:**
- Zero SQL migration needed — `status` is already `TEXT`, not a Postgres `ENUM`.
- Minimal code changes — just update validation and add UI treatment.
- Status remains a single source of truth.
- The word "paused" is semantically clear and distinct from "inactive" (never activated) and "error" (broken).

**Cons:**
- A paused connection that was previously `active` loses the "was active" signal. But this is fine — resuming always sets status back to `active`, and `updated_at` tracks when.

### Option B — Add a `paused BOOLEAN DEFAULT false` column

A separate column orthogonal to status, so a connection can be `active + paused` or `error + paused`.

**Pros:**
- Preserves the underlying status (you can see it was `active` before pausing).

**Cons:**
- Requires a migration to add the column.
- Every `WHERE status = 'active'` query also needs `AND NOT paused` — higher surface area for bugs.
- Frontend and API need to handle a 2D state matrix (status x paused).
- Over-engineered for the actual use case.

### Recommendation: Option A

A new `'paused'` status value is surgical. It leverages the existing `WHERE status = 'active'` filters that already gate ingestion, pollers, and listeners. The only changes needed are validation + UI + a few handler tweaks.

---

## Implementation Plan

### Phase 1 — Backend: Accept and Enforce `paused` Status

**Goal:** The backend accepts `paused` as a valid status, stops pollers/listeners for paused connections, and blocks agent tool access to paused DB connections.

#### Step 1.1 — Validation

**File:** `backend/internal/api/handlers/connections_validate.go:29`

Update `isValidConnectionStatus` to accept `paused`:

```go
func isValidConnectionStatus(s string) bool {
    return s == "active" || s == "inactive" || s == "error" || s == "paused"
}
```

#### Step 1.2 — Update handler: skip poller/listener restart when paused

**File:** `backend/internal/api/handlers/connections.go:350-373`

The update handler currently always stops then restarts the poller/listener. When the new status is `paused`, it should stop and not restart:

```go
// Stop existing poller/listener.
s.Poller.Stop(connID)
s.Listener.Stop(connID)

// Only restart if the connection is not paused.
if status != "paused" {
    if err := connectors.StartPoller(...); err != nil { ... }
    if req.Type == "syslog" { ... }
}
```

#### Step 1.3 — Agent tool: block `query_database` on paused connections

**File:** `backend/internal/agent/tools_db.go:43-47`

After fetching the connection and before creating the database connector, add a status check:

```go
if conn.Status == "paused" {
    return "", fmt.Errorf("query_database: connection %q is paused", conn.Name)
}
```

This returns the error as an `isError` tool result to Claude, so the agent can inform the user rather than silently failing.

#### Step 1.4 — Webhook/OTLP: return 403 instead of 401 for paused connections (optional, low priority)

Currently, paused webhook connections return 401 (indistinguishable from "bad token"). For better DX, we could add a separate query or adjust the response. However, this is low priority — the 401 is functional and doesn't leak information. Defer unless users report confusion.

**No migration required.** The `status` column is `TEXT`, not a Postgres enum.

---

### Phase 2 — Frontend: Visual Treatment + Pause/Resume Button

**Goal:** Users can see when a connection is paused and toggle it with one click.

#### Step 2.1 — TypeScript types

**File:** `frontend/src/types/connection.ts:7`

Add `'paused'` to the status union:

```typescript
status: 'active' | 'inactive' | 'error' | 'paused'
```

Same for `UpdateConnectionPayload:28`.

#### Step 2.2 — ConnectionBubble visual treatment

**File:** `frontend/src/components/connections/ConnectionBubble.vue`

Add a paused state to `statusColor` and `statusGlow`:

- **Status dot colour:** A muted amber/yellow (`bg-status-warning` or a custom `bg-amber-500/60`) — distinct from green (active), red (error), and grey (inactive).
- **Logo colour:** `text-text-muted` (same as inactive) — dimmed but not errored.
- **No pulse animation** — paused is deliberately still.
- **Reduced opacity** on the entire bubble (e.g. `opacity-60`) to visually communicate "dormant".

#### Step 2.3 — StatusBadge

**File:** `frontend/src/components/common/StatusBadge.vue`

Add a `paused` case with amber/yellow styling and "PAUSED" label text.

#### Step 2.4 — ConnectionDetailModal: pause/resume button

**File:** `frontend/src/components/connections/ConnectionDetailModal.vue:256-287`

Add a Pause/Resume toggle button in the actions bar, next to Ping/Edit/Delete:

```
[Ping] [Edit] [Pause]                    [Delete]
```

When connection is paused, the button reads "Resume":

```
[Ping] [Edit] [Resume]                   [Delete]
```

The button emits a new `pause` or `resume` event. The handler calls `updateConnection` with the current connection's fields but flips the status:
- If current status is `active` → set to `paused`
- If current status is `paused` → set to `active`

The Ping and Edit buttons should remain functional for paused connections (you might want to test or edit config before resuming).

#### Step 2.5 — Connections store: add `pauseConnection` / `resumeConnection` actions

**File:** `frontend/src/stores/connections.ts`

Add convenience actions that call `updateConnection` with only the status field changed. These construct the full `UpdateConnectionPayload` from the existing connection data, flipping only `status`.

#### Step 2.6 — ConnectionsPage: wire up pause/resume events

**File:** `frontend/src/pages/ConnectionsPage.vue:311-319`

Add `@pause` and `@resume` handlers on `ConnectionDetailModal` that call the new store actions.

---

### Phase 3 — Polish & Edge Cases

#### Step 3.1 — testConnection store action

**File:** `frontend/src/stores/connections.ts:51-63`

Currently, `testConnection` sets status to `active` on success. If the connection was paused, a successful ping should not auto-resume it. Update:

```typescript
if (idx !== -1 && connections.value[idx].status !== 'paused') {
    connections.value[idx].status = result.success ? 'active' : 'error'
}
```

#### Step 3.2 — Backend test endpoint

Check the `TestConnection` handler (if it exists) — ensure it doesn't change the persisted status of a paused connection. Testing should verify reachability without side effects on pause state.

#### Step 3.3 — Agent chat awareness

When the agent lists available connections (e.g. in tool descriptions or system prompts), paused connections should be indicated as such so the LLM doesn't attempt to use them and get confusing errors. Check how the agent's tool definitions reference available connections.

---

## What This Does NOT Change

- **Log retention** — Logs already ingested from a now-paused connection remain in `log_buffer` and are fully searchable. Pausing stops new data, not access to historical data.
- **Monitor loop** — The monitor processes logs from `log_buffer` regardless of connection status. Paused connections just stop producing new logs.
- **Scheduled investigations** — These query the database or logs on a schedule. If a paused connection is a Postgres tool target, the `query_database` gate (Step 1.3) will return an error to the agent.
- **Existing migrations** — No schema changes needed.

## Scope & Effort

| Phase | Scope | Files touched |
|-------|-------|---------------|
| Phase 1 | Backend validation + enforcement | 3 files (`connections_validate.go`, `connections.go`, `tools_db.go`) |
| Phase 2 | Frontend UI + store | 5 files (`connection.ts`, `ConnectionBubble.vue`, `StatusBadge.vue`, `ConnectionDetailModal.vue`, `connections.ts`, `ConnectionsPage.vue`) |
| Phase 3 | Edge cases + polish | 2-3 files (store, test handler, agent tool descriptions) |

Each phase is independently shippable. Phase 1 alone makes pause functional via the API (e.g. `curl -X PUT ... -d '{"status":"paused"}'`). Phase 2 makes it user-facing. Phase 3 hardens edge cases.

---

## Questions

1. **Should pausing a connection show a toast/confirmation?** Pausing is non-destructive and easily reversible, so a single click (no confirm) seems right. But deleting has a confirm step — should pause match that pattern for consistency, or is the low risk sufficient to skip it?

2. **Should we allow pausing `inactive` or `error` connections?** Currently only `active` → `paused` and `paused` → `active` are the intended transitions. Should the Pause button appear for connections in `inactive` or `error` state, or only for `active` ones? (Recommendation: only show for `active` and `paused` — pausing an inactive/errored connection has no practical effect.)

3. **Webhook response for paused connections:** Currently returns 401 (same as bad token). Should we return a distinct status like `403 Forbidden` with `{"error": "connection_paused"}` so automated senders know the difference? This is more informative but slightly leaks that the token is valid. (Recommendation: keep 401 for security, add an slog.Info line server-side for observability.)

4. **Should the agent proactively mention paused connections?** When a user asks the agent to investigate, should it mention "Note: connection X is paused, so recent logs may be incomplete"? This requires the agent's system prompt to include connection states — worth doing but adds complexity.
