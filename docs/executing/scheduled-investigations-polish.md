# Scheduled Investigations — Polish

## Problem

When a user creates a scheduled investigation, they write a prompt and set a timer — but the UI gives no indication of what the agent will actually have access to when it runs. The user has to mentally map "I have a Postgres connection and a GitHub repo connected" to "the agent can query my database and search my code." This is opaque, especially for new users or apps with many connections.

Two related gaps:

1. **No visibility** — The schedule modal doesn't show which tools or connections the agent will use.
2. **No scoping** — Every schedule gets access to everything. You can't say "only query Production, not Staging."

## Proposal

Two tiers, built sequentially. Tier 1 is the priority; Tier 2 is a follow-up if users need fine-grained control.

---

### Tier 1 — Tool Visibility (Connection Context Panel)

Add a read-only **"Agent capabilities"** section to the schedule create/edit modal. It lists what the agent can do based on the app's current connections.

#### UI

Below the prompt textarea, render a compact panel:

```
┌─ Agent capabilities ──────────────────────────────┐
│                                                    │
│  search_logs     3 log sources connected           │
│  query_database  Production DB, Staging DB         │
│  search_codebase heimdall (GitHub)                 │
│                                                    │
│  The agent can use all of the above when running   │
│  this investigation.                               │
└────────────────────────────────────────────────────┘
```

Rules:
- Tools with zero backing connections show as greyed out / unavailable (e.g. `search_codebase — no repos connected`).
- Connection names are pulled from the app's connection list (already available via `connectionsStore.fetchConnectionsByApp`).
- Tool-to-connection mapping:
  - `search_logs` — count of connections with direction containing log ingestion (`webhook_logs`, `flyio_logs`, `flyio_poller`, `otlp`, `syslog`)
  - `query_database` — list names of `postgres` / `supabase` type connections
  - `search_codebase` — list repo names from `github` type connections

#### Implementation

- **Frontend only** — no backend changes needed.
- Add a small component (e.g. `AgentCapabilities.vue`) that takes an `appId`, reads from `connectionsStore`, and renders the panel.
- Reusable: could later appear in the agent chat sidebar or monitoring config page too.

#### Effort

Small. One new presentational component, a few lines of connection-type mapping logic, drop it into `ScheduleModal.vue`.

---

### Tier 2 — Tool Scoping (Per-Schedule Tool Config)

Let users toggle which tools are enabled per schedule, and for `query_database`, select which specific connections the agent may use.

#### UI

Extend the capabilities panel from Tier 1 with interactive controls:

```
┌─ Agent capabilities ──────────────────────────────┐
│                                                    │
│  [x] search_logs     3 log sources                 │
│  [x] query_database                                │
│       [x] Production DB                            │
│       [ ] Staging DB                                │
│  [ ] search_codebase  heimdall (GitHub)            │
│                                                    │
└────────────────────────────────────────────────────┘
```

- Top-level toggles enable/disable entire tools.
- `query_database` expands to show individual database connections (because it requires an explicit `connection_id` — scoping here is most meaningful).
- `search_logs` and `search_codebase` are all-or-nothing (their implementations don't take a connection_id).
- Default: everything enabled (backwards compatible, matches current behaviour).

#### Data Model

Add a `tool_config` JSONB column to `investigation_schedules`:

```sql
ALTER TABLE investigation_schedules
  ADD COLUMN tool_config JSONB NOT NULL DEFAULT '{}';
```

Shape:

```jsonc
{
  "disabled_tools": ["search_codebase"],        // tools to exclude entirely
  "database_connection_ids": ["uuid-1"]          // if set, only these connections are available to query_database
}
```

Empty object = all tools enabled, all connections available (backwards compatible default).

#### Backend Changes

- **Validation** (`investigation_schedules.go`): Validate `tool_config` on create/update — `disabled_tools` entries must be known tool names, `database_connection_ids` must be valid UUIDs belonging to the app.
- **Tool filtering** (`scheduler.go` → `RunMonitoring`): Pass `tool_config` into the agent loop. Before dispatching a tool call, check whether the tool is disabled. For `query_database`, intercept the `connection_id` parameter and reject it if not in the allowed list.
- **System prompt hint**: When tools are restricted, append a line to the monitoring prompt: "You only have access to: search_logs, query_database (Production DB)." This prevents the LLM from attempting calls that will fail.

#### Effort

Medium. Schema migration, validation logic, tool filtering in the agent loop, and interactive UI in the modal. The filtering itself is straightforward — the main work is wiring the config through the `RunMonitoring` call path.

---

## Recommendation

**Build Tier 1 first.** It solves the immediate UX gap (users don't know what the agent can do) with minimal effort and no backend changes. Tier 2 adds real power but only matters once users have enough connections that "everything enabled" becomes a problem — at that point, the Tier 1 panel becomes the natural place to add toggles.

## Open Questions

- Should the capabilities panel also appear on the **Agent Chat** page? Same gap exists there — users don't know what tools the agent has until they ask it something.
- For Tier 2, should disabled tools be hidden from the LLM's tool list entirely, or kept visible with a "this tool is disabled" error? Hiding is cleaner (no wasted tokens), but erroring is simpler to implement.
- Should `search_logs` eventually support connection-level scoping? Currently it searches all ingested logs globally — but if an app has 5 log sources, a user might want to scope a schedule to "only logs from the payment service."
