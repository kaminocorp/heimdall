# Schedules Enrichment — Connector Shapes & Scheduled Fetches

## Context

With v0.31.0 (2026-04-11), Heimdall shipped **Scheduled Investigations** — a third agent operating mode where stored prompts fire on a cron schedule and run through Claude's tool-use loop, writing observations to the Activity feed. This is agent-centric scheduling: every schedule's payload is a prompt.

A follow-up question came up: *what about connections that don't continuously push logs into Heimdall?* Should users be able to create schedules that simply **fetch logs from a connector** on a cron cadence, distinct from schedules that run the agent?

The short answer is that the push/pull distinction the question is gesturing at already exists in the connector layer — but it's implicit in code rather than a first-class user concept. This doc captures the current shape of things so we can make an informed call when we revisit.

---

## Current Architecture — Four Connector Shapes

Heimdall's connector layer exposes four interface shapes in `backend/internal/connectors/connector.go`:

```go
type Connector interface {                        // base
    Connect(ctx) error
    Health(ctx) error
    Close() error
}

type StreamConnector interface { Connector; Stream(ctx, out) error }  // one-way streaming in
type QueryConnector  interface { Connector; Query(ctx, q) (any, error) }  // on-demand query out
type PollConnector   interface { Connector; Poll(ctx, queries) error }   // timer-driven pull
```

Plus a fourth, defined in `listener.go`:

```go
type Listener interface { Connector; Listen(ctx) error }  // long-lived network socket
```

At runtime, data flows into or out of Heimdall via **four distinct mechanisms**, one per shape:

| Shape | Runtime mechanism | Managed by | Examples |
|---|---|---|---|
| **Push (HTTP)** | External system POSTs to `/api/webhooks/logs` or `/api/otlp`; handler authenticates via bearer token in config | (no manager — just HTTP handler) | `webhook_logs`, `otlp` |
| **Listener** | Long-lived goroutine binding a port, one per connection | `ListenerManager` (`listener.go`) | `syslog` |
| **Poll** | Background goroutine with `time.Ticker`, one per connection, runs forever at `poll_interval_secs` | `Poller` (`poller.go`) | `supabase`, `flyio`, `vercel`, `railway`, `mongodb` |
| **Query** | Connector instantiated **on-demand** per agent tool call — connect, query, close. No persistent state. | Agent tool loop (`tools_db.go`, `tools_codebase.go`) | `postgres`, `github` |

**Key files:**
- `backend/internal/connectors/connector.go` — interface definitions
- `backend/internal/connectors/poller.go` — `Poller` manager, `minPollInterval = 5s` global floor
- `backend/internal/connectors/listener.go` — `ListenerManager`
- `backend/internal/connectors/factory.go:16-50` — `StartPoller` switch (the authoritative list of what's a poller)
- `backend/internal/connectors/registry.go` — shared `Connector` registry (used sparingly today)

---

## Where the Shape is Decided

There is **no single place** that decides what shape a connection has. The shape is implicit in the `type` string, and dispatch to the right runtime path is spread across four call sites that each do their own type check.

### The `connections` table does not encode shape

```
connections
├─ type       text  — e.g. "vercel", "syslog", "postgres"
├─ direction  text  — "one_way" | "two_way"  (semantic flag for the agent, NOT dispatch)
└─ config     jsonb — connector-specific, includes poll_interval_secs for pull types
```

The `direction` column is easy to misread as push/pull — it isn't. It's describing whether data *also* flows outward on demand (`two_way` = "the agent can also query this connection back"). For example, `postgres` is `two_way` because logs could flow in *and* the agent can run SQL against it. It's a semantic flag, not a dispatch mechanism.

### The four dispatch sites

1. **Webhook marker** — `connections.go:201`
   ```go
   if req.Type == "webhook_logs" || req.Type == "otlp" {
       // auto-generate webhook_token in config
   }
   ```
   Pure marker. No goroutine, no connector instance. The HTTP handler at `/api/webhooks/logs` gates on the `webhook_token` stamped into config.

2. **Poll dispatch** — `connections.go:248` → `factory.go:16-50`
   ```go
   connectors.StartPoller(s.Poller, req.Type, config, conn.ID, userID)
   ```
   `StartPoller` is called unconditionally for every connection, but internally switches on `connType` over `{supabase, flyio, vercel, railway, mongodb}`. Unknown types fall through the switch and return `nil` (silent no-op). This is why calling `StartPoller` on a webhook connection is safe.

3. **Listener dispatch** — `connections.go:253`
   ```go
   if req.Type == "syslog" {
       sl, _ := logs.NewSyslog(...)
       s.Listener.Start(r.Context(), sl, conn.ID)
   }
   ```
   Literal `if` statement in the handler — no factory. Today `syslog` is the only listener type.

4. **Query dispatch** — `backend/internal/agent/tools_db.go:40`
   ```go
   if conn.Type != "database" && conn.Type != "postgres" {
       return "", fmt.Errorf("query_database: connection type %q is not a database", conn.Type)
   }
   pg, _ := database.New(conn.Config)
   pg.Connect(ctx); defer pg.Close()
   pg.Query(ctx, sql)
   ```
   Not touched at connection creation at all. The agent's `query_database` tool instantiates a fresh connector per call, type-checks against its allow-list, and tears it down after the query. `search_codebase` (GitHub) follows the same pattern.

### Full shape matrix

| Type | Webhook handler? | `StartPoller` switch? | `syslog` branch? | Agent tool allow-list? | Shape |
|---|---|---|---|---|---|
| `webhook_logs` | ✓ | — | — | — | push |
| `otlp` | ✓ | — | — | — | push |
| `syslog` | — | — | ✓ | — | listener |
| `supabase` | — | ✓ | — | — | poll |
| `flyio` | — | ✓ | — | — | poll |
| `vercel` | — | ✓ | — | — | poll |
| `railway` | — | ✓ | — | — | poll |
| `mongodb` | — | ✓ | — | — | poll |
| `postgres` | — | — | — | ✓ (`query_database`) | query |
| `github` | — | — | — | ✓ (`search_codebase`) | query |

The canonical list of *valid* types lives at `connections.go:642-653` (`validConnectionTypes`), but that map only says "yes this is a known type" — it does not encode shape. Shape is whatever the dispatch sites happen to agree about.

---

## Current Polling Cadence Model

For pull connectors, `poll_interval_secs` is **user-supplied in the connection's config JSON** at creation time. Each connector has its own `DefaultInterval` and `MinInterval` constants (e.g. `vercel.go:23-25`), and the factory spins up a goroutine that ticks at that interval forever.

- **Lifecycle:** goroutine starts at `POST /api/connections`, is replaced on `PUT /api/connections/{id}` (the `Poller.Start` method cancels any existing goroutine for the same `connectionID` first), and is stopped on delete or server shutdown.
- **Boot recovery:** `backend/cmd/heimdall/main.go:204` (`resumePollers`) iterates every connection in the DB on startup and re-calls `StartPoller` for each pull type. Nothing about "what's supposed to be polling" lives in memory between runs — it's all derived from the `connections` table. This is a nice property to preserve in any future redesign.
- **Defaulting gotcha:** if `poll_interval_secs` is below the per-connector floor, the current code *silently replaces* it with the default rather than returning 400. Worth knowing about if we surface this as a user-facing field.

**What you cannot express today:** any cadence richer than "every N seconds forever." No business-hours windows, no pause without delete, no coordinated multi-connector fetches, no cron expressions.

---

## The Distinction the Question Is Asking About

The v0.31.0 Scheduled Investigations are **agent-centric**. Each schedule's payload is a prompt → Claude tool-use loop → observation row. Cost is LLM tokens. Useful for things like *"every hour check pg_stat_statements for slow queries and summarize the worst offenders."*

What the question proposes is a **second schedule type** where the payload is a connection reference instead of a prompt, and firing the schedule means running `conn.Poll()` directly — no agent in the loop. Cost is just the upstream API quota.

|  | Scheduled Investigation (exists) | Scheduled Fetch (proposed) |
|---|---|---|
| **Payload** | Prompt string | Connection reference |
| **Fire action** | Claude tool-use loop | `conn.Poll(ctx, queries)` |
| **Output** | Agent observation row | Raw log rows |
| **Cost** | LLM tokens | Upstream API quota |
| **Use case** | *"Summarize yesterday's error spikes"* | *"Fetch Vercel logs every weekday 09:00–17:00"* |

---

## Design Options

### Option A — Second schedule kind

Extend `investigation_schedules` with a `kind` column (`agent_prompt` | `connector_fetch`). Fork the scheduler's fire path: `agent_prompt` runs the existing agent loop, `connector_fetch` resolves the connection and calls `conn.Poll()`.

- **Pro:** unified UI — `/schedules` shows everything cron-driven in one place. Pause-without-delete. Multiple schedules per connector (cheap/aggressive modes).
- **Con:** goroutine ownership question — does the background `Poller` still exist alongside scheduled fetches, or does one replace the other? If both exist, you can double-poll.
- **Scope:** ~moderate. Schema change on `investigation_schedules`, fork in scheduler, new UI kind in the schedule modal, new REST shape.

### Option B — Move `poll_interval_secs` into schedules entirely

Deprecate the always-on `Poller` for pull connectors. When a pull connector is created, auto-provision a default schedule (e.g. `"*/2 * * * *"`). All cadence lives in schedules; `Poller` is retired.

- **Pro:** strictly simpler mental model. One cadence mechanism, not two. The push/pull distinction becomes literal (pull connectors have schedule entries, push connectors don't).
- **Con:** bigger refactor. Migrate every existing pull connection. Scheduler isn't goroutine-per-connection today (single ticker) — need to verify a fan-out tick doesn't serialize N connector fetches badly. Also loses the "goroutine boots from DB state" property unless we port it to the schedule table.
- **Scope:** ~large. Migration, scheduler redesign, `Poller` removal.

### Option C — Don't add the feature

The existing `Poller` already does on-schedule fetches. Just surface `poll_interval_secs` more prominently in the connection UI. A cron entry that says "fetch Vercel logs" is a more roundabout way of setting `poll_interval_secs` to the cron's period.

- **Pro:** zero new concepts. Conservative and probably fine for now.
- **Con:** doesn't enable any of the use cases that motivate the feature (business-hours polling, coordinated fetches, pause-without-delete).

---

## Architectural Prerequisite — Promote "shape" to a first-class concept

Independent of which option we pick, there's a standing issue worth addressing **before** adding a fifth dispatch site (which any of Option A or B would constitute):

**The taxonomy is implicit and duplicated.** The `StartPoller` switch, the `if req.Type == "syslog"` branch, the `validConnectionTypes` map, and the `query_database` allow-list all have to stay in sync manually. The first time someone adds a new connector type and forgets one of these, the connection gets created successfully but silently never polls. This is the same class of bug as `PruneExpiredLogs` being dead code since v0.4.0 (v0.30.3 discovery).

The clean fix is to extract a `ConnectorKind` concept onto the `Connector` interface or into a central registry:

```go
type ConnectorKind int
const (
    KindWebhook ConnectorKind = iota
    KindListener
    KindPoll
    KindQuery
)

// Option 1: method on the connector
type Connector interface {
    Kind() ConnectorKind
    Connect(ctx) error
    Health(ctx) error
    Close() error
}

// Option 2: static map next to validConnectionTypes
var connectorKinds = map[string]ConnectorKind{
    "webhook_logs": KindWebhook,
    "otlp":         KindWebhook,
    "syslog":       KindListener,
    "supabase":     KindPoll,
    // ...
}
```

With this in place:
- `StartPoller` asks `Kind()` instead of maintaining its own switch
- The `/schedules` UI can answer "is this connection schedulable?" without hardcoding a list
- Adding a new connector type becomes a single-file change (plus tests)
- Frontend can fetch shapes from a `GET /api/connector-types` endpoint and render UI affordances appropriately (e.g. hide "Add schedule" for webhook connectors)

**Recommendation:** treat the `Kind()` extraction as a prerequisite for any scheduled-fetch work. It's ~1 day of focused refactor and it removes a whole class of drift bugs that will otherwise get worse as more connectors land.

---

## Decision Points to Revisit

When we come back to this, the questions to answer in order:

1. **Is there a concrete user story** that the current `Poller` can't serve? Name one or two specific connections where *"every N seconds forever"* is wrong. Candidates:
   - Business-hours-only polling (rate-limited upstream APIs, cost-sensitive)
   - Coordinated multi-connector fetches ("at 03:00 pull yesterday's Postgres slow-query report AND yesterday's Vercel logs so the agent can cross-reference")
   - Pull connectors that are expensive to query and shouldn't run while the team is asleep

   If we can't name one: **Option C** (do nothing, defer).

2. **If we have a use case, do we want one cadence mechanism or two?**
   - Two → **Option A** (sibling schedule kind, lower refactor cost, mental overhead of "which thing owns the tick")
   - One → **Option B** (bigger refactor, cleaner long-term shape)

3. **Do we do the `Kind()` extraction first?** Recommended yes regardless of which option we pick. Also valuable on its own for any future frontend work that needs to reason about connector shapes (e.g. the connection wizard's "what can this connector do" affordances).

4. **Scheduler fan-out concern for Option B:** the current scheduler in `backend/internal/agent/scheduler.go` is a single ticker. Verify it can handle fan-out without serializing N connector fetches before committing to Option B.

5. **`direction` column cleanup:** while we're in this area, consider renaming or removing `direction` — it's currently confused with push/pull and only really exists as a semantic flag for the agent. Either promote it to something meaningful or drop it.

---

## Related

- `docs/executing/schedules-enrichment.md` — the original question that kicked this off
- `docs/changelog.md` §0.31.0 — Scheduled Investigations (the v0.31.0 agent-centric schedules feature)
- `docs/plans/scaling-assessment.md` — touches on the `Poller` goroutine lifecycle from a different angle (scaling to hundreds of customers)
- `docs/vision.md` — §Connections, §The Agent (describes the three agent modes including Scheduled)
