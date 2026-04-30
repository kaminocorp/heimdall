# RLS Enforcement — Mental Model

Companion to [`rls-enforcement-role-split.md`](./rls-enforcement-role-split.md). That doc is the execution plan; this one is the conceptual map to read alongside it.

The question this doc answers: *where, in which layers/files, does each of the three roles show up, and how do workflows that don't have a JWT (cron jobs, background agent work, webhook ingestion) authenticate, authorise, and route through the right DB connection?*

---

## Three things to anchor on first

- **The current codebase has exactly one pool** (`backend/cmd/heimdall/main.go:45`) feeding everything: HTTP handlers, the agent monitor loop, connector pollers, and syslog listeners. Everything shares `cfg.DatabaseURL` → `postgres` superuser. The role split is fundamentally "split that one pool into two runtime pools and repoint migrations at a third URL."
- **`UserQueries` already exists and is load-bearing** (`backend/internal/api/handlers/userqueries.go:24`). It begins a transaction, runs `SELECT set_config('app.current_user_id', $1, true)`, and returns a `*db.Queries` scoped to that tx. The `SET LOCAL` plumbing — the thing RLS policies read via `current_setting('app.current_user_id', true)` — is in place *today* and fires on every JWT-authenticated handler. It's just cosmetic because `postgres` bypasses RLS.
- **Two known gaps from a quick grep.** `s.Pool.Begin` is called directly (no `SET LOCAL`) from `webhooks.go:337` and `otlp.go:117` — those are the non-JWT ingestion paths and exactly the "won't go through `UserQueries`" cases §5.3 of the plan warns about.

---

## The three roles, in one table

| Role | Env var | Port | Privileges | Used by | Where it shows up in code |
|---|---|---|---|---|---|
| `app_user` | `DATABASE_URL` | 5432 | Non-superuser, **RLS-enforced**, CRUD on `public.*` | HTTP request handlers, WebSocket chat, JWT-authenticated writes | `main.go:45` → `router.go:16` → every handler via `s.UserQueries(...)` |
| `cron_user` | `CRON_DATABASE_URL` (new) | 5432 | Non-superuser, `BYPASSRLS`, blanket SELECT + narrow writes | Monitor loop, connector pollers, scheduler, notifications dispatcher | `main.go` (new second `pgxpool.New`) → `agent.Agent`, `connectors.Poller`, `ListenerManager` |
| `postgres` | `DIRECT_URL` | 5432 | Superuser | `golang-migrate` only | `Makefile` targets `migrate-up` / `migrate-down` / `migrate-create` |

---

## Layer-by-layer: where each connects

### Today (single pool)

```
cmd/heimdall/main.go:45    pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
                      :91    queries := db.New(pool)
                      :100   ag := agent.New(queries, …)              ← superuser → monitor loop
                      :111   router := api.NewRouter(cfg, pool, ag…)  ← superuser → handlers
```

Everything downstream — handlers, the agent, pollers, listeners — receives either `pool` or `queries` constructed from that one URL.

### After the split

```
cmd/heimdall/main.go
  appPool   := pgxpool.New(ctx, cfg.DatabaseURL)      ← app_user   (RLS-enforced)
  cronPool  := pgxpool.New(ctx, cfg.CronDatabaseURL)  ← cron_user  (BYPASSRLS, narrow)
  appQ      := db.New(appPool)
  cronQ     := db.New(cronPool)

  ag := agent.New(appPool, cronPool, …)   ← needs BOTH: cron for enumerate, app for writes
  router := api.NewRouter(cfg, appPool, ag, …)  ← handlers only ever see appPool
```

`Server.Pool` (consumed by every handler) keeps being `app_user`. The new `cronPool` is threaded into the three non-JWT subsystems: the agent, `connectors.Poller`, `connectors.ListenerManager`.

### Per-layer mapping

| Layer | Files | Role it uses | How |
|---|---|---|---|
| **HTTP handlers (JWT)** | `handlers/connections.go`, `logs.go`, `conversations.go`, `applications.go`, `chat.go`, etc. | `app_user` | Middleware extracts `user_id` from JWT → `s.UserQueries(ctx, userID)` → `SET LOCAL` → RLS policies fire |
| **Webhook ingestion (bearer token)** | `handlers/webhooks.go:337`, `handlers/otlp.go:117` | `app_user` *after refactor* | Currently `s.Pool.Begin` without `SET LOCAL`. Must become: resolve `connection.UserID` → `s.UserQueries(ctx, connection.UserID)` |
| **Syslog TLS listener** | `connectors/logs/syslog.go`, resume path at `main.go:200` | `app_user` *after refactor* | `NewSyslog(…, conn.UserID, *conn.AppID, queries)` — already carries `UserID`; just needs to switch from `queries` writes to `UserQueries(ctx, conn.UserID)` writes |
| **Connector pollers** (Fly.io, Supabase, Vercel, Railway, MongoDB) | `connectors/poller.go`, `connectors/logs/flyio.go`, etc., resume at `main.go:232` | Both | Enumeration in `resumePollers` uses `cron_user`; per-poll writes switch to `UserQueries(ctx, conn.UserID)` on `app_user` |
| **Monitor loop** | `agent/monitor.go:52` (enumeration), `:115`–`:200` (per-app work) | Both | `ListActiveApplications` on `cron_user`; everything after `resolveOrgUser` on `app_user` via `UserQueries(ctx, userID)` |
| **Pipeline writer** | `agent/pipeline_writer.go` | `app_user` | Called from inside `monitorApp`, which already has `userID` — writes go through the scoped transaction |
| **Scheduled investigations** | `agent/scheduler.go` / investigation handlers | Both | Cron fires → enumerate due schedules on `cron_user` → for each, run investigation via `UserQueries(ctx, schedule.UserID)` on `app_user` |
| **Notifications dispatcher** | `notifications/dispatcher.go` | `app_user` | User context is already present in the event being dispatched — tenant reads/writes scope to that user |
| **Migrations** | `backend/migrations/*.sql`, `Makefile` | `postgres` | `golang-migrate` binary reads `DIRECT_URL`; never touched by the running backend |

---

## The non-JWT authentication question

The confusing piece: *if there's no JWT, how does the code know which user to `SET LOCAL` for?* The answer depends on **how the request entered the system**, and it splits into three distinct cases.

### Case 1 — Bearer-token ingestion (webhooks, OTLP, syslog, Fly.io drain)

These *do* have a per-request identity, just not a JWT. Every ingestion connection row in `connections` has a `user_id` column (the org owner who created the connection) and a bearer token. The lookup sequence:

```
POST /api/webhooks/logs
  Authorization: Bearer <webhook_token>
         ↓
  ResolveConnectionByToken(token)  → returns (conn.ID, conn.UserID, conn.AppID)
         ↓
  s.UserQueries(ctx, conn.UserID)  ← this is what's MISSING at webhooks.go:337 today
         ↓
  INSERT INTO log_buffer … through the RLS-scoped tx
```

Today `webhooks.go` calls `s.Pool.Begin` directly and inserts without `SET LOCAL` — it works because `postgres` bypasses RLS. Under `app_user`, RLS denies the insert unless `app.current_user_id` matches. The fix is routing the insert through `UserQueries(ctx, conn.UserID)`. The `conn.UserID` is already resolved — the refactor is mechanical.

**Identity source**: the bearer token resolves to a `connection`, which owns a `user_id`. That's the authenticated party.

### Case 2 — Background goroutines with no per-request identity (monitor loop, scheduler)

This is the case `cron_user` exists for. Look at `monitor.go:52`:

```go
apps, err := a.queries.ListActiveApplications(ctx)
```

There is no user this query "belongs to" — it's by design cross-tenant. Under `app_user` with RLS enforced, this query returns zero rows (no `app.current_user_id` is set, the policy rejects). That's why we need `cron_user` with `BYPASSRLS` — it's the only way to answer "give me all apps across all tenants that need monitoring."

But — critically — the moment the loop has *an app*, it resolves an owning user at `monitor.go:115`:

```go
userID, err := a.resolveOrgUser(ctx, app.OrgID)
```

That's the handoff point. Everything below that line — fetching logs, classifying, escalating to Claude, writing to `agent_log` and `log_pipeline_events` — should run on the `app_user` pool inside a `UserQueries(ctx, userID)` transaction. The cron pool did the enumeration; the app pool does the tenant work.

The pattern in pseudo-code:

```go
// On cron_user pool — cross-tenant enumeration is fine:
apps := cronQ.ListActiveApplications(ctx)

for _, app := range apps {
    userID := resolveOrgUser(app.OrgID)         // still cron_user, still BYPASSRLS

    // HANDOFF: switch to app_user pool for all tenant work
    q, commit, done, _ := s.UserQueries(ctx, userID)
    defer done()

    logs   := q.ListLogsSinceForApp(…)          // RLS-scoped
    result := classify(logs)
    q.InsertAgentLog(…)                         // RLS-scoped
    q.InsertPipelineEvent(…)                    // RLS-scoped
    commit()
}
```

**Identity source**: discovered *from the row being processed*. The "authenticated user" is derived from the data, not from an incoming request.

### Case 3 — Infra-shape writes (cursor updates, advisory locks)

Things like `UPDATE monitoring_state SET last_monitored_at = NOW() WHERE app_id = $1` or `pg_try_advisory_lock(app_id)` aren't tenant writes per se — they're coordination state. These stay on `cron_user`, which is why §4.2 of the plan explicitly grants `INSERT, UPDATE ON monitoring_state TO cron_user` as a narrow exception. The invariant: `cron_user` can update *infrastructure* rows but not *tenant data* rows.

---

## Putting it together

The mental model:

1. **JWT present** → middleware sets `ctx` with `user_id` → handler calls `UserQueries(ctx, userID)` → `app_user` pool → RLS enforced. Standard case, already works.
2. **Bearer token present** → handler resolves token to `connection.user_id` → `UserQueries(ctx, conn.UserID)` → `app_user` pool → RLS enforced. Currently broken at two files, small fix.
3. **No request at all (background)** → enumerate via `cron_user` → for each row, resolve owning user → hand off to `UserQueries(ctx, ownerUserID)` → `app_user` pool → RLS enforced for the actual writes. The "no user" is only true for the 1–2 queries that have to be cross-tenant; the rest is user-scoped.

The single most important invariant: **`cron_user` never writes tenant data directly.** Its job is "find the user, then get out of the way." Every tenant `INSERT`/`UPDATE`/`DELETE` goes through `UserQueries` on `app_user`, regardless of whether the work started from a JWT, a webhook, a cron tick, or a scheduled investigation. That's what makes RLS enforcement reach every write path in the system.
