# Database Connection Architecture

## Overview

Heimdall uses **pgxpool** (pgx's built-in connection pool) with **per-request transactions** to serve all database queries. This document captures the architecture, the reasoning behind it, and production tuning guidance.

## Connection Type: Pool with Per-Request Transactions

### How It Works

```
HTTP Request
  |
  v
pgxpool.Pool  (bounded set of reusable connections)
  |
  Pool.Begin(ctx)  -->  acquires a connection, starts a transaction
  |
  SET LOCAL app.current_user_id = '<uuid>'  -->  sets RLS identity for this transaction
  |
  Execute queries via Queries.WithTx(tx)
  |
  tx.Commit(ctx)  -->  commits transaction, releases connection back to pool
```

**Key files:**
- `cmd/heimdall/main.go` — pool creation
- `internal/api/handlers/server.go` — pool held on `Server` struct
- `internal/api/handlers/userqueries.go` — per-request transaction helper
- `internal/db/db.go` — sqlc-generated `Queries` with `WithTx()` support

### The `UserQueries` Helper

Every authenticated handler calls `UserQueries(ctx, userID)`, which:

1. Acquires a connection from the pool and begins a transaction (`Pool.Begin`)
2. Runs `SET LOCAL app.current_user_id = $1` to set the RLS session variable (scoped to the transaction — cannot leak to other requests)
3. Returns a `*db.Queries` bound to that transaction, plus a `done()` cleanup function
4. The caller defers `done()`, which commits the transaction and releases the connection

```go
func (s *Server) UserQueries(ctx context.Context, userID uuid.UUID) (*db.Queries, func(), error) {
    tx, err := s.Pool.Begin(ctx)
    // ...
    tx.Exec(ctx, "SET LOCAL app.current_user_id = $1", userID.String())
    return s.Queries.WithTx(tx), func() { tx.Commit(ctx) }, nil
}
```

### Why This Pattern

| Concern | How it's addressed |
|---------|--------------------|
| **Connection efficiency** | Pool multiplexes many concurrent requests over a small number of DB connections. We don't need one connection per user — only one per concurrent query. |
| **RLS isolation** | `SET LOCAL` is scoped to the transaction. Concurrent requests from different users never share session state. This is the canonical way to do RLS with connection pooling. |
| **No connection leaks** | Connections are acquired and released within a single handler call. No long-lived connection holds. |
| **WebSocket safety** | The chat handler opens short-lived transactions per DB operation (not one for the whole WebSocket session). Connections aren't held during slow Claude API calls. |

### What We Considered

- **Session pooling** (one connection per user session): Wasteful — holds connections during idle time. Doesn't scale past a few hundred concurrent users.
- **Direct connections** (no pool): Each request opens/closes a TCP connection + TLS handshake. Extremely slow and resource-heavy.
- **Single shared connection**: No concurrency at all. Non-starter.

The pool-with-transactions approach is the standard for Go services and matches how PgBouncer's "transaction mode" works.

## Current Defaults (pgxpool)

The pool is created with **default settings** via `pgxpool.New(ctx, databaseURL)`:

| Setting | Default | Notes |
|---------|---------|-------|
| `MaxConns` | `max(4, runtime.NumCPU())` | Typically 4-8 on most machines |
| `MinConns` | 0 | No connections kept warm at idle |
| `MaxConnLifetime` | 1 hour | Connections recycled after 1h |
| `MaxConnIdleTime` | 30 minutes | Idle connections closed after 30m |
| `HealthCheckPeriod` | 1 minute | Background liveness check |

These defaults are fine for development and low traffic but must be tuned for production.

## Production Tuning Guide

### When to Tune

When the system serves more than ~50 concurrent users, or when you observe connection acquisition latency in logs/metrics.

### Recommended Configuration

```go
config, err := pgxpool.ParseConfig(cfg.DatabaseURL)
if err != nil {
    // handle error
}

config.MaxConns = 25                          // match your Supabase/Postgres plan limit
config.MinConns = 5                           // keep connections warm for fast acquisition
config.MaxConnLifetime = 30 * time.Minute     // recycle before server-side timeout
config.MaxConnIdleTime = 5 * time.Minute      // release idle connections sooner
config.HealthCheckPeriod = 30 * time.Second   // catch dead connections faster

pool, err := pgxpool.NewWithConfig(context.Background(), config)
```

### Choosing `MaxConns`

The pool's `MaxConns` should be set based on the **database's connection limit**, not the number of users:

- **Supabase Free**: 60 direct connections (leave headroom for migrations, admin)
- **Supabase Pro**: 200+ direct connections
- **Self-hosted Postgres**: Check `SHOW max_connections`

A good starting point is **50-75% of the database limit** for the application pool, reserving the rest for admin, migrations, and monitoring connections.

If you deploy multiple backend instances, divide the pool limit across them (e.g., 3 instances with `MaxConns=15` each against a 60-connection database).

### Supabase Connection Pooler (Supavisor)

For high-concurrency scenarios (thousands of concurrent users), consider using Supabase's built-in connection pooler (Supavisor) instead of direct connections:

- Use port `6543` instead of `5432` in your `DATABASE_URL`
- Supavisor supports up to 1500 concurrent client connections on Pro plans
- Operates in **transaction mode** by default — compatible with our `SET LOCAL` + per-request transaction pattern
- Allows a much higher `MaxConns` in pgxpool since Supavisor multiplexes onto fewer real Postgres connections

### Scaling Characteristics

| Concurrent users | Expected concurrent DB ops | Pool size needed | Notes |
|-------------------|---------------------------|------------------|-------|
| < 100 | ~10-20 | 4-8 (defaults) | Fine as-is |
| 100-500 | ~20-50 | 15-25 | Tune `MaxConns` |
| 500-2000 | ~50-200 | 25-50 | Use Supavisor |
| 2000+ | ~200+ | 50+ client-side | Supavisor + multiple backend instances |

These are rough estimates. Actual concurrency depends on query duration, request patterns, and how many users are actively chatting vs. idle.

## WebSocket Connection Pattern

The chat handler deserves special mention. It does **not** hold a database connection for the lifetime of the WebSocket:

```
WebSocket connected
  |
  [UserQueries → load conversation → done()]     <-- connection borrowed & returned
  |
  for each message:
    [UserQueries → persist user message → done()] <-- brief borrow
    [UserQueries → update title → done()]         <-- brief borrow
    [Claude API call... seconds pass]             <-- NO connection held
    [UserQueries → persist agent response → done()] <-- brief borrow
```

This means 1000 open WebSocket sessions might only need 10-20 pool connections, since most sessions are idle or waiting on Claude at any given moment.

## Monitoring

To observe pool health at runtime, `pgxpool.Pool` exposes a `Stat()` method:

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
