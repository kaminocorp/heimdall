# Database Connection Blueprint

How Heimdall's Go backend talks to its primary Postgres database (Supabase-hosted) — the shape of `DATABASE_URL`, what port 5432 actually means on Supabase, how the in-process pool is built and threaded, how Row-Level Security is wired through per-request transactions, how the WebSocket chat handler keeps connection pressure low, and what to tune when we hit real scale.

This document is **descriptive, not prescriptive** — it captures the connection topology that is actually in place as of `master`. Planned migrations away from this setup (e.g. moving to Supavisor for horizontal scale) live as follow-up docs, not edits to this file.

> **Post-rollout reading note (Phase 8 of the RLS role split, 2026-05).** The document below was originally written when the backend authenticated as a single `postgres` superuser via one `DATABASE_URL`. The eight-phase RLS rollout (see `docs/completions/rls-enforcement-phase-{1..8}.md` and the archived [`rls-enforcement-roadmap.md`](../archive/rls-enforcement-roadmap.md)) replaced that with a **three-role topology** under three env vars. Read §0 first; the rest of this doc still applies, but every inline mention of "the pool" / "the URL" / "the password" should be read as "the *app* pool / app URL / app password" plus an analogous *cron* pool, with `DIRECT_URL` carrying the migration-only superuser path.

---

## 0. Three URLs, three roles, two runtime pools

The runtime authenticates as one of two non-superuser roles — never `postgres`. Migrations run as `postgres` via a separate URL that the application process never reads.

| Env var | Role (post-rollout) | `BYPASSRLS` | Used by | Pool size |
|---|---|---|---|---|
| `DATABASE_URL` | `app_user` | **no** (`FORCE ROW LEVEL SECURITY` is the boundary) | JWT/WebSocket request path, `Pools.UserQueries` / `Pools.WithUserQueries` | ~30 |
| `CRON_DATABASE_URL` | `cron_user` | yes | Background loops with no user identity at entry — monitor, scheduler, pollers, log-buffer pruner. Cross-tenant enumerate, then hand off to the app pool via `WithUserQueries` for the actual write. | ~5 |
| `DIRECT_URL` | `postgres` | yes (superuser) | `make migrate-up` / `migrate-down` only — DDL needs ownership and can't be served by the runtime pools. | n/a (one-shot) |

`HEIMDALL_ENV=production` enables a startup invariant (`backend/cmd/heimdall/main.go:79–84`) that refuses to launch when `DATABASE_URL` and `CRON_DATABASE_URL` authenticate as the same role. The three URLs are independent secrets in the production environment; rotate them independently.

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

The server does not open one connection per request. `backend/cmd/heimdall/main.go` constructs **two pools** at startup — the *app pool* (size ~30) and the *cron pool* (size ~5) — and threads both into a `*db.Pools` struct that every subsystem receives. The single-pool snippet below is the original construction shape; today it runs twice with different URLs:

```go
pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
if err != nil {
    slog.Error("failed to connect to database", "err", err)
    os.Exit(1)
}
defer pool.Close()
```

This is **`pgx/v5`'s own client-side pool** (`github.com/jackc/pgx/v5/pgxpool`). It is entirely separate from any server-side pooler — pgxpool just keeps a set of live `*pgx.Conn` values, hands them out on `Acquire`, and returns them on `Release`.

### How a request uses the pool

```
HTTP Request
  │
  ▼
pgxpool.Pool               (bounded set of reusable connections)
  │
  ├─ Pool.Begin(ctx)       (acquires a connection, starts a transaction)
  │
  ├─ SET LOCAL app.current_user_id = '<uuid>'   (sets RLS identity for this tx)
  │
  ├─ Execute queries via Queries.WithTx(tx)
  │
  └─ tx.Commit(ctx)        (commits, releases connection back to pool)
```

That per-request transaction wrapper is `UserQueries` — see §4.

### Where the pool goes

The pool is threaded explicitly through the dependency graph; nothing ever reaches out to a package-level global:

- `db.New(pool)` at `main.go:91` wraps the pool in sqlc's generated `*db.Queries` façade, which is what most handler and agent code calls.
- `api.NewRouter(cfg, pool, ag, jwks, ...)` at `main.go:111` hands the pool to the HTTP layer, which stores it on `handlers.Server.Pool` (`internal/api/handlers/server.go:17`) for the small set of handlers that need a raw tx (RLS session-variable setup — see §4).
- `notifications.NewDispatcher(queries, cfg)` and `agent.New(queries, cfg, ...)` take the `queries` façade rather than the raw pool — most of the codebase works through sqlc, not pgx directly.

On shutdown, `defer pool.Close()` at `main.go:50` waits for all in-flight queries to finish before returning. This is chained after `srv.Shutdown(ctx)` and the connector/agent stop sequence so the pool is the last thing to go.

### One-shot connections (not pooled)

Two call sites deliberately skip the pool:

1. **`backend/cmd/dbping/main.go:23`** uses `pgx.Connect(ctx, url)` — a single connection with no pool. Correct for a CLI probe that opens, pings, and exits.
2. **`backend/internal/connectors/database/postgres.go:88`** uses `pgx.Connect` per user-configured Postgres connector. These connections reach *user-owned* databases (the DB-activity data source), typically long-lived and configured per-connection, so they don't belong in the application pool.

---

## 4. RLS via per-request transactions — the `UserQueries` helper

Every authenticated handler that needs user-scoped data calls `UserQueries(ctx, userID)` on `*handlers.Server` (`internal/api/handlers/userqueries.go`). It does four things:

1. Acquires a connection from the pool and begins a transaction: `Pool.Begin(ctx)`.
2. Runs `SET LOCAL app.current_user_id = $1` to set the RLS session variable. `SET LOCAL` is scoped to the transaction — it cannot leak to other requests, even if the same backend is reused.
3. Returns a `*db.Queries` bound to that transaction (via sqlc's generated `WithTx`), plus a `done()` cleanup function.
4. The caller defers `done()`, which commits the transaction and releases the connection.

```go
func (s *Server) UserQueries(ctx context.Context, userID uuid.UUID) (*db.Queries, func(), error) {
    tx, err := s.Pool.Begin(ctx)
    // ...
    tx.Exec(ctx, "SET LOCAL app.current_user_id = $1", userID.String())
    return s.Queries.WithTx(tx), func() { tx.Commit(ctx) }, nil
}
```

### Why this pattern

| Concern | How it's addressed |
|---|---|
| **Connection efficiency** | Pool multiplexes many concurrent requests over a small number of DB connections. We don't need one connection per user — only one per concurrent query. |
| **RLS isolation** | `SET LOCAL` is scoped to the transaction. Concurrent requests from different users never share session state. This is the canonical way to do RLS with connection pooling. |
| **No connection leaks** | Connections are acquired and released within a single handler call. No long-lived connection holds. |
| **WebSocket safety** | The chat handler opens short-lived transactions per DB op, not one for the whole WebSocket session. Connections aren't held during slow Claude API calls (see §6). |

### Alternatives considered (and rejected)

- **Session pooling** (one connection per user session): wasteful — holds connections during idle time. Doesn't scale past a few hundred concurrent users.
- **Direct connections with no pool**: every request pays TCP + TLS handshake. Slow and resource-heavy.
- **Single shared connection**: no concurrency. Non-starter.

The pool-with-per-request-transactions approach is the standard for Go services and composes cleanly with PgBouncer transaction mode if we ever move to Supavisor.

---

## 5. The sqlc layer on top of pgxpool

`db.New(pool)` (`backend/internal/db/`) returns a `*Queries` value generated by sqlc from `backend/internal/db/queries/*.sql`. sqlc-generated methods call into the pool by type-asserting the `DBTX` interface, which is satisfied by both `*pgxpool.Pool` and `pgx.Tx` — that's how the same generated code runs pooled or inside a manual transaction.

Key config from `backend/sqlc.yaml`: uuid → `google/uuid.UUID`, jsonb → `json.RawMessage`, timestamptz → `time.Time`. Migrations live in `backend/migrations/` (latest visible: `038_log_pipeline_events.up.sql`) and are applied with golang-migrate (`make migrate-up`). **Migrations are immutable once applied** — 0.46.4's schema-drift audit (see `docs/changelog.md`) covers why drift-auditing is a recurring concern.

---

## 6. The WebSocket chat connection pattern

The chat handler deserves special mention: it does **not** hold a database connection for the lifetime of the WebSocket.

```
WebSocket connected
  │
  [UserQueries → load conversation → done()]      ← connection borrowed & returned
  │
  for each message:
    [UserQueries → persist user message → done()] ← brief borrow
    [UserQueries → update title → done()]         ← brief borrow
    [Claude API call... seconds pass]             ← NO connection held
    [UserQueries → persist agent response → done()] ← brief borrow
```

This means 1000 open WebSocket sessions might only need 10–20 pool connections, since most sessions are idle or waiting on Claude at any given moment. Pool sizing (§7) is driven by concurrent *queries*, not concurrent *users*.

---

## 7. Current pgxpool defaults and production tuning

Today the pool is created with **default settings** via `pgxpool.New(ctx, databaseURL)`:

| Setting | Default | Notes |
|---|---|---|
| `MaxConns` | `max(4, runtime.NumCPU())` | Typically 4–8 on most machines |
| `MinConns` | 0 | No connections kept warm at idle |
| `MaxConnLifetime` | 1 hour | Connections recycled after 1h |
| `MaxConnIdleTime` | 30 minutes | Idle connections closed after 30m |
| `HealthCheckPeriod` | 1 minute | Background liveness check |

These are fine for development and low traffic but should be tuned for production.

### When to tune

When the system serves more than ~50 concurrent users, or when `pgxpool.Stat().EmptyAcquireCount()` (see §11) grows steadily in logs — that's the unambiguous signal that acquirers are waiting.

### Recommended production configuration

```go
config, err := pgxpool.ParseConfig(cfg.DatabaseURL)
if err != nil {
    // handle error
}

config.MaxConns = 25                        // match your Supabase/Postgres plan limit
config.MinConns = 5                         // keep connections warm for fast acquisition
config.MaxConnLifetime = 30 * time.Minute   // recycle before server-side timeout
config.MaxConnIdleTime = 5 * time.Minute    // release idle connections sooner
config.HealthCheckPeriod = 30 * time.Second // catch dead connections faster

pool, err := pgxpool.NewWithConfig(context.Background(), config)
```

### Choosing `MaxConns`

`MaxConns` should be set based on the **database's connection limit**, not the number of users:

- **Supabase Free**: 60 direct connections (leave headroom for migrations, admin)
- **Supabase Pro**: 200+ direct connections
- **Self-hosted Postgres**: check `SHOW max_connections`

A good starting point is **50–75% of the database limit** for the application pool, reserving the rest for admin, migrations, and monitoring connections.

If you deploy multiple backend replicas, divide the pool limit across them (e.g. 3 replicas with `MaxConns=15` each against a 60-connection database). The scale trigger is **replica count × `MaxConns`**, not RPS per replica — once that product approaches `max_connections`, you either shrink the pool, move to Supavisor, or raise the Supabase plan.

### Scaling characteristics

| Concurrent users | Concurrent DB ops (est.) | Pool size | Notes |
|---|---|---|---|
| < 100 | ~10–20 | 4–8 (defaults) | Fine as-is |
| 100–500 | ~20–50 | 15–25 | Tune `MaxConns` explicitly |
| 500–2000 | ~50–200 | 25–50 client-side | Move to Supavisor (§8) |
| 2000+ | ~200+ | 50+ client-side | Supavisor + multiple backend replicas |

These are rough estimates. Actual concurrency depends on query duration, request patterns, and how many users are chatting vs. idle (§6 keeps the idle ones cheap).

---

## 8. Why we're on direct (5432) and not the transaction pooler today

Our RLS pattern (`UserQueries`: `BEGIN` → `SET LOCAL` → queries → `COMMIT`) is **fully compatible with Supavisor transaction mode** — `SET LOCAL` is tx-scoped, and the transaction is the atomic unit PgBouncer hands to a backend. If RLS were the only constraint, we could move to port 6543 with no code change.

The reasons we stay on direct today are about *other* things that would break under transaction-mode pooling:

1. **Session GUCs on user-owned Postgres connectors.** `internal/connectors/database/postgres.go` connects to *user-supplied* Postgres databases (the DB-activity data source) and enforces `default_transaction_read_only=on` as a session parameter — set once on connect, applies to every transaction on that connection. That's a session-scoped GUC; it would be reset at every `COMMIT` under PgBouncer transaction mode. This applies to the *downstream* connector, not to Heimdall's own DB, but it illustrates the invariant, and if we ever share connection handling patterns we'd need to rework it.
2. **Prepared statement cache.** pgx prepares queries on first use and caches plans on the connection. Under transaction-mode pooling the cache is invalidated on every `COMMIT`, killing the performance benefit and risking "prepared statement already exists" errors when the same backend comes back around. Supavisor has partial mitigations, but no complete fix.
3. **Future `LISTEN/NOTIFY`.** Pub/sub-style event fan-out (currently TODO-tier — `pipeline_bus.go` runs in-process today) *requires* session persistence. A future decision to use Postgres-backed pub/sub across replicas would force a carve-out for those connections regardless of which port handles HTTP traffic.

**Takeaway:** moving the primary endpoint to Supavisor is a project, not a config change, but it's a smaller project than one might assume — the RLS pattern is safe. A future migration doc should enumerate the call sites above and their replacement patterns before anyone flips the URL.

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
effective_connections = pool.MaxConns × replicas + sidecar / migration tooling
```

User-owned Postgres connectors (`internal/connectors/database/postgres.go`) don't count against this — they go to the user's database, not ours.

Today the default pool sizing (4 or `NumCPU`) times a small number of backend replicas sits comfortably under any Supabase plan ceiling. **The scale trigger is horizontal replica count, not traffic per replica.** In ascending cost, the options when we approach the ceiling:

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

We don't currently log or export these. When pool saturation becomes a question ("are 502s from pool exhaustion or upstream slowness?"), the fast answer is to wire a periodic `slog.Info("pool", "acquired", stat.AcquiredConns(), ...)` ticker into `main.go` alongside the existing connector heartbeats. This is a one-hour change that pays off the first time there's an incident.

Supabase also surfaces per-project connection and query metrics in its dashboard — those are authoritative for server-side backend count; pgxpool's numbers are authoritative for client-side pressure. Mismatches between the two (e.g. pgxpool idle but Supabase reports connections held) almost always mean long-running queries rather than a pooling bug.

---

## 12. Summary

- `DATABASE_URL` is a libpq URI pointing at `db.<project-ref>.supabase.co:5432` — Supabase's **direct Postgres** endpoint on the Postgres-native port.
- Port 5432 is session-mode: one TCP socket = one backend process. Alternative endpoints (5432-via-session-pooler, 6543 transaction pooler) trade session semantics for scale and are **not** what we use today.
- In-process, `pgxpool.New` at `backend/cmd/heimdall/main.go:45` builds one client-side pool; it is passed explicitly into the router, agent, and notifications dispatcher, and wrapped by sqlc's `db.Queries` for typed access.
- Every authenticated request funnels through `UserQueries`: `Pool.Begin` → `SET LOCAL app.current_user_id` → sqlc queries → `tx.Commit`. This composes cleanly with Supavisor transaction mode, so the RLS pattern is not a blocker to eventual migration.
- What *is* a blocker: session GUCs on user-owned Postgres connectors, the prepared-statement cache, and any future `LISTEN/NOTIFY` use. Moving to Supavisor is a project, not a config change.
- WebSocket chat handlers borrow and release connections per DB op, not per session — 1000 open sockets cost roughly the query concurrency of an active few.
- Scale ceiling is `pool.MaxConns × replicas` vs. Supabase's `max_connections`. The trigger is replica count, not RPS.
- Defaults are acceptable to ~50 concurrent users; beyond that, set `MaxConns` to 50–75% of the DB's `max_connections` divided across replicas, set `MinConns` > 0 to keep connections warm, and shorten `MaxConnIdleTime`.
- Observability: `pgxpool.Stat()` exists, we don't sample it yet, wiring it up is cheap and pays off at the first incident.
