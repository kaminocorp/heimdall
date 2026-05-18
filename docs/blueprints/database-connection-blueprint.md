# Database Connection Blueprint

How Heimdall's Go backend talks to its primary Postgres database (Supabase-hosted) — the shape of `DATABASE_URL`, what port 5432 actually means on Supabase, how the in-process pool is built and threaded, how Row-Level Security is wired through per-request transactions, how the WebSocket chat handler keeps connection pressure low, and what to tune when we hit real scale.

This document is **descriptive, not prescriptive** — it captures the connection topology that is actually in place as of `master`. Planned migrations away from this setup (e.g. moving to Supavisor for horizontal scale) live as follow-up docs, not edits to this file.

> **Post-rollout note (Phase 10 of the RLS role split, 2026-05).** This document originated as a single-pool, single-superuser description. The eight-phase RLS rollout plus the 0.48.1 polish pass and 0.48.2 review-pass fixes (see `docs/completions/rls-enforcement-phase-{1..10}.md`, the archived [`rls-enforcement-roadmap.md`](../archive/rls-enforcement-roadmap.md), and `docs/changelog.md` entries 0.48.0–0.48.2) replaced that with the **three-URL / two-runtime-pool** topology described below. The post-Phase-2 chat handler's "Option A" long-transaction pattern (§6) and the `UserQueriesForLoop` helper (§4) are also rollout artefacts. Read §0 first.

---

## 0. Three URLs, three roles, two runtime pools

The runtime authenticates as one of two non-superuser roles — never `postgres`. Migrations run as `postgres` via a separate URL that the application process never reads.

| Env var | Role (post-rollout) | `BYPASSRLS` | Used by | Pool size |
|---|---|---|---|---|
| `DATABASE_URL` | `app_user` | **no** (`FORCE ROW LEVEL SECURITY` is the boundary) | JWT/WebSocket request path, `Pools.UserQueries` / `Pools.UserQueriesForLoop` / `Pools.WithUserQueries` | 30 |
| `CRON_DATABASE_URL` | `cron_user` | yes | Background loops with no user identity at entry — monitor, scheduler, pollers, log-buffer pruner. Cross-tenant enumerate, then hand off to the app pool via `WithUserQueries` for the actual write. | 5 |
| `DIRECT_URL` | `postgres` | yes (superuser) | `make migrate-up` / `migrate-down` only — DDL needs ownership and can't be served by the runtime pools. | n/a (one-shot) |

Pool sizes are not defaults — `backend/cmd/heimdall/main.go::buildPool` sets `MaxConns` explicitly on both pools (`appPool, … := buildPool(ctx, cfg.DatabaseURL, 30, "app")`, `cronPool, … := buildPool(ctx, cronURL, 5, "cron")`).

Two startup checks gate the role split:

1. **Always-on role probe** (`main.go::logPoolRoles`). Runs unconditionally on every boot. Issues `SELECT current_user` against each pool and logs `app_role` / `cron_role` at INFO. If both pools resolve to the same role, emits a WARN with actionable text — the breadcrumb that catches a production deploy that forgot to set `HEIMDALL_ENV`.
2. **Production hard-fail** (`main.go::assertRoleSplit`). Gated on `HEIMDALL_ENV=production`. Same `SELECT current_user` probe, but `os.Exit(1)`s when the roles collide. Authoritative check — URL string comparison would miss the case where two distinct URLs both resolve to `postgres`.

The three URLs are independent secrets in the production environment; rotate them independently.

The "bypass-then-scope" pattern is the load-bearing handoff: `cron_user` does the smallest possible cross-tenant lookup (`SELECT owner_user_id FROM applications WHERE id = $1`), then the per-tenant work continues on the `app_user` pool inside a `WithUserQueries(ctx, ownerUserID, ...)` scope where `SET LOCAL app.current_user_id` is set. Audit trail and policy evaluation are identical to a normal JWT request.

For the *why* (threat model, design alternatives considered and rejected), see the archived [`rls-enforcement-role-split.md`](../archive/rls-enforcement-role-split.md). For the *what shipped where*, walk the eight phase completion docs in `docs/completions/`. For the conceptual map of which roles surface in which layers, see [`rls-enforcement-mental-model.md`](./rls-enforcement-mental-model.md). For the precedent codebase that pioneered this shape, see [`../refs/trajan-db-roles.md`](../refs/trajan-db-roles.md).

---

## 1. The connection string

`DATABASE_URL` is loaded from `.env` by `make dev-backend`, or from the process environment in production. It is read once at startup in `backend/internal/config/config.go:36` and validated in `Validate()` at `config.go:59` — an empty value is a hard startup failure.

Its shape is the standard libpq URI format:

```
postgresql://<user>:<password>@<host>:<port>/<dbname>?<params>
```

For Heimdall the host resolves to a Supabase project's primary Postgres endpoint (`db.<project-ref>.supabase.co`) and the port is `5432`. The database name is `postgres` (Supabase's default — we don't create a per-app database; we namespace by schema and table ownership).

### What `postgresql://` means

`postgresql://` (and its alias `postgres://`) is a URI serialisation of libpq's connection-parameter set. It is **not** a separate protocol. Any driver that speaks the Postgres wire protocol (pgx, psycopg, node-postgres, etc.) accepts this form. The scheme tells the parser how to split the string; the parser hands the components to the driver; the driver opens a **raw TCP connection** to `<host>:<port>` and speaks the Postgres v3 frontend/backend protocol over it, upgraded to TLS (see §10).

---

## 2. Port 5432 — what kind of connection is this?

Port **5432** is the IANA-registered default for Postgres. Connecting to it on a Supabase project host gives you a **direct session connection** to the underlying Postgres primary: one TCP socket corresponds to one Postgres backend process, for the lifetime of that socket.

Supabase exposes **three distinct connection endpoints** per project; which one you use determines how the connection behaves on the server side, not just at the client.

| Endpoint | Port | Server-side mode | What's usable |
|---|---|---|---|
| Direct database | **5432** | Raw Postgres (1 client ↔ 1 backend process) | Everything: session state, `SET LOCAL`, prepared statements, `LISTEN/NOTIFY`, advisory locks, long transactions |
| Session pooler | 5432 (via `*.pooler.supabase.com`) | Supavisor / PgBouncer in **session mode** | Same as direct — PgBouncer holds one backend per client for the session's lifetime. Useful from IPv4-only environments (direct 5432 on Supabase is IPv6 in some regions). |
| Transaction pooler (Supavisor) | 6543 | PgBouncer in **transaction mode** | Per-transaction state only. Backends are returned to the pool at `COMMIT`/`ROLLBACK`. Compatible with `SET LOCAL` inside an explicit transaction (which is exactly our RLS pattern), incompatible with session-scoped GUCs, prepared statement caches, and `LISTEN/NOTIFY`. |

Heimdall uses the **direct session connection on 5432** today. §8 explains why that's load-bearing and what migrating to Supavisor would actually cost.

---

## 3. The application-side pool

The server does not open one connection per request. `backend/cmd/heimdall/main.go` constructs **two pools** at startup — the *app pool* (size 30) and the *cron pool* (size 5) — via `buildPool`, then wraps them in a `*db.Pools` struct that every subsystem receives:

```go
appPool, err := buildPool(ctx, cfg.DatabaseURL, 30, "app")
// ... error handling, defer appPool.Close()

cronURL := cfg.CronDatabaseURL
if cronURL == "" {
    slog.Warn("CRON_DATABASE_URL unset; falling back to DATABASE_URL until Phase 6 of the RLS role split")
    cronURL = cfg.DatabaseURL
}
cronPool, err := buildPool(ctx, cronURL, 5, "cron")
// ... error handling, defer cronPool.Close()

pools := db.NewPools(appPool, cronPool)
```

`buildPool` is a thin wrapper around `pgxpool.NewWithConfig` that applies `MaxConns` and logs the result; everything else (`MinConns`, lifetimes, healthcheck) stays at pgx defaults today.

This is **`pgx/v5`'s own client-side pool** (`github.com/jackc/pgx/v5/pgxpool`). It is entirely separate from any server-side pooler — pgxpool just keeps a set of live `*pgx.Conn` values, hands them out on `Acquire`, and returns them on `Release`.

### How a request uses the pool

```
HTTP Request
  │
  ▼
pgxpool.Pool (App)         (bounded set of reusable connections, MaxConns=30)
  │
  ├─ App.Begin(ctx)        (acquires a connection, starts a transaction)
  │
  ├─ set_config('app.current_user_id', $1, true)   (RLS identity, tx-local)
  │
  ├─ Execute queries via Queries.WithTx(tx)
  │
  ├─ commit() → tx.Commit(WithoutCancel(ctx))      (caller-controlled,
  │                                                 success path only)
  │
  └─ done()  → tx.Rollback if commit wasn't called (always deferred)
```

That per-request transaction wrapper is `UserQueries`; the long-tx variant for the chat path is `UserQueriesForLoop`; both live on `*db.Pools` — see §4.

### Where the pools go

The `*db.Pools` value is threaded explicitly through the dependency graph; nothing reaches out to a package-level global. The struct is the single chokepoint for "where does a query run?" decisions (see `backend/internal/db/pools.go::Pools`):

- `api.NewRouter(cfg, pools, ag, jwks, ...)` hands the pair to the HTTP layer, which stores it on `handlers.Server.Pools` (`internal/api/handlers/server.go::Server`). Handlers reach the per-tenant pool via `s.Pools.UserQueries(...)` (or the shim `s.UserQueries(...)`, a one-line forwarder kept for ergonomics).
- `notifications.NewDispatcher(pools, cfg)` and `agent.New(pools, cfg, ...)` take the full `Pools` value rather than a bare `*Queries` — background subsystems need both the cron path (cross-tenant enumeration) and the per-tenant handoff in the same package.
- `connectors.NewPoller(pools)` and the syslog listener manager receive `Pools` for the same reason: enumerate active connections on `Cron`, then per-message ingestion runs against `App` under `WithUserQueries`.

On shutdown, both `defer appPool.Close()` and `defer cronPool.Close()` in `main.go` wait for in-flight queries to drain before returning. These are chained after `srv.Shutdown(ctx)` and the connector/agent stop sequence so the pools are the last things to go.

### One-shot connections (not pooled)

Two call sites deliberately skip the pool:

1. **`backend/cmd/dbping/main.go:23`** uses `pgx.Connect(ctx, url)` — a single connection with no pool. Correct for a CLI probe that opens, pings, and exits.
2. **`backend/internal/connectors/database/postgres.go:88`** uses `pgx.Connect` per user-configured Postgres connector. These connections reach *user-owned* databases (the DB-activity data source), typically long-lived and configured per-connection, so they don't belong in the application pool.

---

## 4. RLS via per-request transactions — the `Pools` helpers

The canonical helpers live on `*db.Pools` (`backend/internal/db/pools.go`); handlers reach them via a one-line `s.UserQueries(...)` shim. The two per-tenant variants share a single implementation (`userTx`) with one bool flag that toggles agent-loop-friendly timeout disables.

### `Pools.UserQueries` — short HTTP / ingestion transactions

For HTTP handlers and webhook ingestion. Acquires a connection from the **app pool**, begins a transaction, sets `app.current_user_id` for RLS, and returns four values:

```go
func (p *Pools) UserQueries(ctx context.Context, userID uuid.UUID) (
    queries *Queries,     // bound to the tx via sqlc's WithTx
    commit  func() error, // call on success path before encoding the response
    done    func(),       // always defer — rolls back if commit was not called
    err     error,
)
```

The cleanup pattern is split into `commit` + `done` rather than a single `done()` that auto-commits, because the original auto-commit shape silently persisted partial writes when a handler returned early on error after a successful first write. The current contract: read-only callers assign `_` to `commit` and just defer `done()`; write callers call `commit()` explicitly on the success path before encoding their response, and `done()` rolls back if `commit()` never ran.

The session-variable write uses `SELECT set_config('app.current_user_id', $1, true)` — the `true` third arg makes it transaction-local, equivalent to `SET LOCAL` but parameterisable (`SET LOCAL` does not accept bind parameters). It cannot leak to other requests, even if the same backend is reused.

Both error-path rollbacks and the deferred `done()` use `context.WithoutCancel(ctx)` so a client disconnect can't strand a half-applied transaction. Per-statement `Exec` calls still use the caller's context, so an in-flight query can still cancel on disconnect — only cleanup is shielded.

### `Pools.UserQueriesForLoop` — long agent transactions

For the WebSocket chat handler's per-turn transaction (see §6). Identical to `UserQueries` plus two `SET LOCAL` guards inside the tx:

- `SET LOCAL idle_in_transaction_session_timeout = 0` — the gap between tool calls while Claude is thinking would otherwise trip Postgres' idle-in-txn limit and abort the transaction mid-loop.
- `SET LOCAL statement_timeout = 0` — Phase 10 fix. The first guard alone wasn't enough on Supabase: Supabase enforces a per-role 8s `statement_timeout` by default, which would abort the *next tool query* inside the same txn after a long ChatCompletion, not the LLM call itself.

The trade-off is documented inline: one app-pool connection is pinned for the loop's lifetime, typically tens of seconds, occasionally minutes. App-pool sizing (30) accounts for this.

### `Pools.WithUserQueries` — the closure-shaped sibling

Cleaner shape when the caller has a single linear block of work and doesn't need an explicit commit point. Used heavily by background subsystems' per-tenant work:

```go
err := pools.WithUserQueries(ctx, ownerUserID, func(q *db.Queries) error {
    // ... writes via q
    return nil
})
```

Opens a transaction, sets the user-id GUC, runs `fn`, commits on `nil` error, rolls back otherwise. Same underlying `userTx` path as `UserQueries`.

### `Pools.CronQueries` — the BYPASSRLS path

Returns a `*Queries` bound *directly* to the cron pool — no transaction, no GUC. Used by cross-tenant enumeration paths (monitor's `ListActiveApplications`, scheduler's `ListEnabledSchedules`, the log-buffer pruner, the connector resume helpers in `main.go`). The `cron_user` role's narrow grants are the safety net: SELECT-only on tenant tables; targeted INSERT/UPDATE on `monitoring_state`; DELETE on `log_buffer`. An accidental tenant write from this path fails loudly on a privilege violation.

The "bypass-then-scope" handoff: enumerate on `CronQueries`, then for each tenant call `WithUserQueries(ctx, ownerUserID, ...)` for the actual work. Audit trail and RLS policy evaluation are identical to a normal JWT request.

### Why this pattern

| Concern | How it's addressed |
|---|---|
| **Connection efficiency** | App pool multiplexes many concurrent requests over 30 backend connections. We don't need one connection per user — only one per concurrent query (plus one per active chat turn — see §6). |
| **RLS isolation** | `app.current_user_id` is set transaction-locally via `set_config(..., true)`. Concurrent transactions from different users never share session state. This is the canonical way to do RLS with connection pooling. |
| **No partial commits** | Explicit `commit` / `done` split. Early returns after the first write roll back instead of silently persisting. |
| **Disconnect safety** | Cleanup paths use `context.WithoutCancel`, so a client hang-up doesn't strand a half-applied tx. |
| **Long agent loops** | `UserQueriesForLoop` disables both timeouts via `SET LOCAL`; the loop survives Claude's think time without bleeding into other paths. |

### Alternatives considered (and rejected)

- **One DB connection per user session** (HTTP or WebSocket scope): wasteful — pins connections during idle time. Doesn't scale past a few hundred concurrent users, and is the opposite of the pgxpool model.
- **Direct connections with no pool**: every request pays TCP + TLS handshake. Slow and resource-heavy.
- **Single shared connection**: no concurrency. Non-starter.
- **Per-tool-call short transactions inside the agent loop** (the pre-Phase-2 shape): would have made the loop compatible with Supavisor transaction mode, but lost the snapshot guarantee — a tool query mid-loop could see writes from a concurrent turn that landed between iterations. Option A (one tx per turn) was chosen for consistency; the cost is the pinned connection covered above.

The pool-with-per-request-transactions approach is the standard for Go services and composes cleanly with Supavisor transaction mode for everything except the agent-loop path (see §8).

---

## 5. The sqlc layer on top of pgxpool

`db.New(pool)` (`backend/internal/db/`) returns a `*Queries` value generated by sqlc from `backend/internal/db/queries/*.sql`. sqlc-generated methods call into the pool by type-asserting the `DBTX` interface, which is satisfied by both `*pgxpool.Pool` and `pgx.Tx` — that's how the same generated code runs pooled or inside a manual transaction.

Key config from `backend/sqlc.yaml`: uuid → `google/uuid.UUID`, jsonb → `json.RawMessage`, timestamptz → `time.Time`. Migrations live in `backend/migrations/` (latest visible: `041_lookup_user_for_invite.up.sql`; the 039–041 trio is the role-split landing: `runtime_role_grants`, `rls_force_enforcement`, and the sqlc-stable `LookupUserIDForInvite` rewrite) and are applied with golang-migrate (`make migrate-up`). **Migrations are immutable once applied** — 0.46.4's schema-drift audit (see `docs/changelog.md`) covers why drift-auditing is a recurring concern; the catalog-drift test added in Phase 7 (`internal/db/catalog_drift_test.go`) pins `forcerowsecurity = true` on every tenant-scoped table to prevent silent regressions.

---

## 6. The WebSocket chat connection pattern

The chat handler is the one place that deliberately pins a connection. Phase 2 of the RLS role-split rollout introduced "Option A": run each chat *turn* (one user message → one assistant response, with N tool dispatches in between) inside a single `UserQueriesForLoop` transaction. The earlier shape (short txn per DB op, no connection held during Claude calls) was traded away for transactional consistency — a tool query mid-loop must see the same snapshot as the persist that opened the turn.

```
WebSocket connected
  │
  [UserQueries → setupConversation → commit → done()]  ← brief borrow,
  │                                                       ~ms
  for each user message:
    [UserQueriesForLoop: BEGIN, set GUCs, disable timeouts]   ← borrow,
      persist user message                                       held for
      RunConversationStream (up to 10 LLM iterations):           the
        tool_start / tool_result emits via q                     entire
        Claude API call (seconds — sometimes 30s+)               turn
        tool query via q
        ...
      persist agent response
    commit() → done()                                         ← release
```

Per-turn DB connection cost is therefore: **one app-pool connection per active chat turn**, held for the wall-clock duration of the turn (LLM round-trips dominate). Between turns, the connection is back in the pool — idle WebSockets are still cheap.

**Concurrency invariant.** The producer goroutine in `agent.RunConversationStream` (`agent/loop_stream.go`) holds exclusive use of `q` while it's emitting events; `pgx.Tx` is not safe for concurrent use. The consumer in `chat.go` must drain the events channel to completion *before* `done()` runs, even on the error path. A `wsWriteErr` flag captures the failure, the loop keeps draining (still capturing terminal `message`/`error` events for symmetry), and the early `return` happens only after the producer's `defer close(ch)` has fired. The 0.48.2 review pass caught a regression here: returning on the first WS write error fired `defer done()` while the producer was still issuing `q.X` calls, racing `tx.Rollback` against concurrent statements. The fix is documented inline at `chat.go:194–222` and the contract is explained in detail in the `pools.go::UserQueriesForLoop` docstring.

**Sizing implication.** 30 app-pool connections supports ~25 concurrent active chat turns plus normal HTTP traffic. That's the headroom number, not "30 concurrent WebSocket sessions" — idle sockets don't count. If chat concurrency outgrows app-pool capacity during the operator bake, the planned response is to add a third dedicated "loop" pool rather than reverting to per-iteration short transactions; the snapshot guarantee the loop relies on is load-bearing for tool-result correctness.

---

## 7. Current pgxpool sizing and production tuning

`MaxConns` is set explicitly per pool in `buildPool`; everything else stays at pgx defaults:

| Setting | App pool | Cron pool | Notes |
|---|---|---|---|
| `MaxConns` | **30** | **5** | Set in `main.go`. App size accounts for §6's pinned-during-turn cost; cron is enumeration-only and rarely concurrent. |
| `MinConns` | 0 | 0 | pgx default. No connections kept warm at idle; first request pays the open cost. |
| `MaxConnLifetime` | 1 hour | 1 hour | Connections recycled after 1h |
| `MaxConnIdleTime` | 30 minutes | 30 minutes | Idle connections closed after 30m |
| `HealthCheckPeriod` | 1 minute | 1 minute | Background liveness check |

### When to tune further

The current sizing is sufficient through the operator bake. Signals that say "increase":

- `pgxpool.Stat().EmptyAcquireCount()` (see §11) grows steadily — acquirers waiting.
- Tail latency on chat-turn start climbs without a corresponding LLM-side slowdown.
- Supabase dashboard shows the app pool sitting at 30 backends for sustained periods.

Signals that say "split" (introduce a third pool, e.g. a dedicated "loop" pool for `UserQueriesForLoop`):

- Short HTTP requests start blocking on chat-turn acquisition while plenty of cron capacity sits idle. Symptom is HTTP p95 climbing in lockstep with active chat count.

### Knobs worth tuning before changing pool count

```go
poolCfg.MinConns = 5                         // keep a warm baseline
poolCfg.MaxConnLifetime = 30 * time.Minute   // recycle before any server-side timeout
poolCfg.MaxConnIdleTime = 5 * time.Minute    // release idle conns sooner under bursty load
poolCfg.HealthCheckPeriod = 30 * time.Second // catch dead conns faster
```

None of these are wired today; doing so is a 10-line change to `buildPool`.

### Choosing `MaxConns` (when sizing changes)

`MaxConns` is bounded by the **database's connection limit**, not the number of users:

- **Supabase Free**: 60 direct connections (leave headroom for migrations, admin)
- **Supabase Pro**: 200+ direct connections
- **Self-hosted Postgres**: check `SHOW max_connections`

Today's `30 + 5 = 35` per replica leaves comfortable headroom on Free for single-replica deploys, and would fit ~5 replicas on Pro before approaching the ceiling. The scale trigger is **replica count × (app + cron `MaxConns`)**, not RPS per replica — once that product approaches `max_connections`, you either shrink the pool, move to Supavisor for the non-loop traffic (§8), or raise the Supabase plan.

### Scaling characteristics

| Active chats + HTTP RPS | Likely bottleneck | Pool action | Notes |
|---|---|---|---|
| < ~25 concurrent turns, < 100 RPS | None | Hold at 30/5 | Current state |
| 25–100 concurrent turns | App-pool saturation during turns | Bump app to 50, consider third "loop" pool | Splits the chat path from short HTTP |
| 500+ concurrent turns | App-pool × backend replicas vs Supabase ceiling | Multi-replica + dedicated loop pool + Supavisor for non-loop | The migration plan in §8 becomes load-bearing |
| 2000+ | Postgres throughput, not pool size | Supavisor + multi-replica + plan upgrade / read replica | Reshape the workload |

These are estimates. Actual concurrency depends on turn duration (LLM-dominated), HTTP request mix, and how many connectors are actively ingesting (each ingestion handler also takes an app-pool conn briefly).

---

## 8. Why we're on direct (5432) and not the transaction pooler today

Our RLS pattern (`UserQueries`: `BEGIN` → `set_config('app.current_user_id', ..., true)` → queries → `COMMIT`) is **fully compatible with Supavisor transaction mode** — the GUC is tx-scoped, and the transaction is the atomic unit PgBouncer hands to a backend. If RLS were the only constraint, we could move to port 6543 with no code change.

The reasons we stay on direct today are about *other* things that would break under transaction-mode pooling:

1. **Long-lived agent-loop transactions.** `UserQueriesForLoop` holds one transaction across an entire chat turn — multiple Claude round-trips, multiple tool dispatches, often 10s+ wall-clock. Transaction-mode pooling's whole pitch is *"the backend is free between transactions, so it can serve other clients."* Long transactions pin a backend regardless of pooling mode — txn-mode multiplexing gives us zero benefit on this path, while we'd still pay the costs in items 2–4 below. The Phase 2 `SET LOCAL idle_in_transaction_session_timeout = 0` + Phase 10 `SET LOCAL statement_timeout = 0` guards are also tx-scoped, so they survive a hypothetical port-6543 move — the constraint is genuinely about lost performance, not correctness.
2. **Session GUCs on user-owned Postgres connectors.** `internal/connectors/database/postgres.go` connects to *user-supplied* Postgres databases (the DB-activity data source) and enforces `default_transaction_read_only=on` as a session parameter — set once on connect, applies to every transaction on that connection. That's a session-scoped GUC; it would be reset at every `COMMIT` under PgBouncer transaction mode. This applies to the *downstream* connector, not to Heimdall's own DB, but it illustrates the invariant, and if we ever share connection handling patterns we'd need to rework it.
3. **Prepared statement cache.** pgx prepares queries on first use and caches plans on the connection. Under transaction-mode pooling the cache is invalidated on every `COMMIT`, killing the performance benefit and risking "prepared statement already exists" errors when the same backend comes back around. Supavisor has partial mitigations, but no complete fix.
4. **Future `LISTEN/NOTIFY`.** Pub/sub-style event fan-out (currently TODO-tier — `pipeline_bus.go` runs in-process today) *requires* session persistence. A future decision to use Postgres-backed pub/sub across replicas would force a carve-out for those connections regardless of which port handles HTTP traffic.

**Takeaway:** moving the primary endpoint to Supavisor is a project, not a config change. The RLS pattern is safe, but item 1 means the chat path would need to stay on a session-mode endpoint (direct or session-pooler) regardless — a future migration doc should describe a split (chat / loop on session-mode, short HTTP on transaction-mode) rather than a single flip, and enumerate items 2–4 with their replacement patterns.

### Supavisor when you do move

- Use port `6543` instead of `5432` in the `DATABASE_URL`.
- Supavisor supports up to 1500 concurrent client connections on Pro plans.
- Operates in transaction mode by default — compatible with our `UserQueries` pattern.
- Allows a much higher `MaxConns` in pgxpool since Supavisor multiplexes onto fewer real Postgres backends.

---

## 9. Capacity and scale ceiling

A Supabase project on a given plan has a fixed `max_connections` ceiling on the underlying Postgres instance. Every `pgxpool.Pool` opens up to `MaxConns` TCP connections; every backend process counts against the same server-side limit.

Per deployment we therefore consume, against Heimdall's own DB:

```
effective_connections = (app.MaxConns + cron.MaxConns) × replicas
                      + migration / sidecar tooling on DIRECT_URL
                      = 35 × replicas (+ migrations)
```

User-owned Postgres connectors (`internal/connectors/database/postgres.go`) don't count against this — they go to the user's database, not ours.

Today's 30 + 5 = 35 per replica times a small number of backend replicas sits comfortably under any Supabase plan ceiling. **The scale trigger is horizontal replica count, not traffic per replica.** In ascending cost, the options when we approach the ceiling:

1. Cap `pool.MaxConns` lower (starves request concurrency first).
2. Move `DATABASE_URL` to Supabase's session pooler (same semantics as direct, PgBouncer-fronted — preserves session state, adds a network hop).
3. Bifurcate: keep direct for anything in §8's list, route short transactional workloads to Supavisor on 6543 (requires auditing every query for transaction-mode safety).
4. Bump the Supabase plan / add a read replica.

Step 3 is the one that would require the migration doc mentioned in §8; steps 1, 2, and 4 are configuration-only.

---

## 10. TLS, auth, and the password in `DATABASE_URL`

The password in `DATABASE_URL` is the **Supabase project's database password** — not a user JWT, not the Supabase service-role key. It's a static credential scoped to the Postgres user embedded in the URL (`postgres` by default, or a dedicated role if the project has been configured to use one).

Supabase enforces TLS on port 5432; pgx negotiates it automatically when the server advertises SSL support. Our `DATABASE_URL` doesn't pin an `sslmode`, which means pgx defaults to `prefer` — it will attempt TLS and fall back to plaintext if the server refuses. In practice Supabase never refuses, so the connection is always TLS. If we ever needed hard assurance, setting `sslmode=require` (or `verify-full` with a bundled CA) in the URL is a one-line change.

### Rotation

Rotating the DB password is a Supabase-side operation: regenerate in the Supabase dashboard, update `DATABASE_URL` in the deployment environment, redeploy. There is no in-app caching of the URL beyond process lifetime — the next process start picks up the new value.

---

## 11. Monitoring

### What's wired today

- **Role attribution at startup.** `logPoolRoles` (always on) emits an INFO with `app_role` / `cron_role` from a `SELECT current_user` on each pool. If both roles match, a WARN with actionable text fires regardless of environment. This is the breadcrumb that catches a production deploy that forgot to flip the URLs.
- **Production hard-fail.** `assertRoleSplit` (under `HEIMDALL_ENV=production`) runs the same probe and `os.Exit(1)`s on a role collision.
- **Pool-ready INFO.** `buildPool` logs `database pool ready` with `label` and `max_conns` for each pool so the sizing is auditable from log search.

### What's not wired

`pgxpool.Pool` exposes a `Stat()` snapshot we don't currently sample:

```go
stats := pool.Stat()
// stats.TotalConns()        — current pool size
// stats.IdleConns()         — connections sitting idle
// stats.AcquiredConns()     — connections in use right now
// stats.MaxConns()          — configured maximum
// stats.AcquireCount()      — total acquisitions since startup
// stats.EmptyAcquireCount() — acquisitions that had to wait (pool exhausted)
```

When `EmptyAcquireCount` grows steadily, the pool is too small.

We don't currently log or export these. When pool saturation becomes a question ("are 502s from pool exhaustion or upstream slowness?"), the fast answer is to wire a periodic `slog.Info("pool", "label", "app", "acquired", stat.AcquiredConns(), ...)` ticker into `main.go` alongside the existing connector heartbeats — one for each pool. This is a one-hour change that pays off the first time there's an incident.

Supabase also surfaces per-project connection and query metrics in its dashboard — those are authoritative for server-side backend count; pgxpool's numbers are authoritative for client-side pressure. Mismatches between the two (e.g. pgxpool idle but Supabase reports connections held) almost always mean long-running queries rather than a pooling bug — `UserQueriesForLoop` transactions are a known and intentional cause.

---

## 12. Summary

- **Three URLs, three roles, two runtime pools.** `DATABASE_URL` → `app_user` (per-tenant CRUD, FORCE RLS), `CRON_DATABASE_URL` → `cron_user` (BYPASSRLS, narrow grants), `DIRECT_URL` → `postgres` (migrations only). `HEIMDALL_ENV=production` hard-fails if app and cron resolve to the same role; an always-on WARN catches the same misconfig in any environment.
- All three URLs are libpq URIs pointing at `db.<project-ref>.supabase.co:5432` — Supabase's **direct Postgres** endpoint. Port 5432 is session-mode: one TCP socket = one backend process. Alternative endpoints (session-pooler, 6543 transaction pooler) trade session semantics for scale and are **not** what we use today.
- In-process, `buildPool` constructs two client-side pools with explicit `MaxConns` (`app = 30`, `cron = 5`); both are wrapped in `*db.Pools` and passed explicitly into the router, agent, connectors, and notifications dispatcher.
- Per-tenant work funnels through one of three Pools helpers: `UserQueries` (short HTTP / ingestion), `UserQueriesForLoop` (agent chat turn, with timeout disables for LLM round-trips), or `WithUserQueries` (closure shape for background subsystems). All three set `app.current_user_id` transaction-locally via `set_config(..., true)`. Cross-tenant enumeration uses `CronQueries` on the cron pool; the bypass-then-scope handoff is the load-bearing pattern for background work.
- The chat path pins one app-pool connection per active *turn* (Option A from the Phase 2 rollout). Idle WebSockets are still free; concurrent turns are not. Sizing accounts for this; if needed, the next move is a dedicated "loop" pool rather than reverting to per-iteration short transactions.
- RLS composes cleanly with Supavisor transaction mode, so it is not a blocker to eventual migration. What *is* a blocker: long-lived agent transactions (item 1 of §8), session GUCs on user-owned Postgres connectors, the prepared-statement cache, and any future `LISTEN/NOTIFY` use. Moving to Supavisor is a project — and probably a split, not a single flip — not a config change.
- Scale ceiling is `(app + cron) × replicas = 35 × replicas` vs. Supabase's `max_connections`. The trigger is replica count, not RPS.
- Observability: role probes log at startup (both pools), `pgxpool.Stat()` is not yet sampled — wiring it up per pool is cheap and pays off at the first incident.
