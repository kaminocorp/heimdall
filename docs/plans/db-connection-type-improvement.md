# Database Connection Type Improvement — 5432 → 6543

**Status:** not started — planning doc.
**Owner:** TBD.
**Scheduling:** triggered by scale, not security. This is the follow-up to `rls-enforcement-role-split.md` but explicitly **does not block** it. The RLS role split should land first and bake; this migration is a separate project with separate triggers, separate failure modes, and separate rollback paths. Bundling them roughly triples the blast radius.

---

## 1. The question in one sentence

Heimdall's runtime connects to Supabase Postgres on **port 5432 (direct session mode)**. Every serious multi-tenant Supabase deployment eventually moves its primary runtime traffic to **port 6543 (Supavisor transaction mode)** so that thousands of clients can multiplex onto a small number of real Postgres backends. This doc captures *why* Heimdall is on 5432 today, *when* that stops being the right answer, and *how* to migrate when the trigger fires.

The honest framing: the 5432 decision is defensible but not "world-class end-state." It's "correct for Heimdall's current scale." A Series C AI-native SaaS with a serious infra function would almost certainly run its primary runtime pool on 6543 with prepared statements disabled. That transition is cheap-to-moderate in complexity but large in blast radius — exactly the kind of thing you plan before you need it rather than scramble to do during an incident.

---

## 2. Current state — what port 5432 actually gives us

Documented authoritatively in `docs/blueprints/database-connection-blueprint.md`. The short version:

```
DATABASE_URL  → postgresql://<user>:<pw>@db.<project-ref>.supabase.co:5432/postgres
                 └─ Direct Postgres on the IANA-registered port.
                 └─ One TCP socket = one Postgres backend process, for the lifetime of that socket.
                 └─ Used by pgxpool in backend/cmd/heimdall/main.go:45.
                 └─ pgxpool defaults: MaxConns = max(4, NumCPU), MinConns = 0, no prepared-stmt tuning.
```

Supabase exposes **three** endpoints per project; we use the first:

| Endpoint | Port | Server-side mode | What's usable |
|---|---|---|---|
| Direct database (**current**) | 5432 | Raw Postgres (1 client ↔ 1 backend process) | Everything: session state, `SET LOCAL`, named prepared statements, `LISTEN/NOTIFY`, advisory locks |
| Session pooler | 5432 via `*.pooler.supabase.com` | Supavisor / PgBouncer in session mode | Same as direct; PgBouncer holds one backend per client for the session lifetime. Useful from IPv4-only envs. |
| Transaction pooler (**target**) | 6543 via `*.pooler.supabase.com` | Supavisor in transaction mode | Per-transaction state only. Backends returned to pool at `COMMIT`/`ROLLBACK`. Compatible with `SET LOCAL` inside explicit tx (our RLS pattern); incompatible with session-scoped GUCs, persistent prepared statements, and `LISTEN/NOTIFY`. |

Capacity today:

```
effective_connections = pool.MaxConns × replicas + sidecar / migration tooling
```

Default pool sizing (4 or NumCPU) times one Fly.io backend replica = ~4–8 connections against Heimdall's own DB, plus migration tooling. That's nowhere near any Supabase plan ceiling (Free: 60 direct; Pro: 200+). **The scale trigger is horizontal replica count, not RPS per replica.** Once `MaxConns × replicas` starts approaching `max_connections`, the cost of staying on 5432 becomes real — and one of four things happens:

1. Starve request concurrency by capping `MaxConns` lower (kicks the can).
2. Move to Supabase's session pooler on 5432 (same semantics as direct, adds a network hop, preserves session state — one-config-line change).
3. **This document: move to Supavisor transaction mode on 6543.** Requires an audit of every query for transaction-mode safety, plus code-level disablement of named prepared statements.
4. Bump the Supabase plan / add a read replica.

Options (1), (2), and (4) are configuration changes. Option (3) is a project.

---

## 3. Why we're on 5432 today — the three real reasons

The blueprint's §8 names three constraints that would break under naïve transaction-mode pooling. Each deserves scrutiny because at least one of them is weaker than it sounds.

### 3.1 Prepared-statement cache (real, tractable)

pgx prepares queries on first use and caches plans **on the backend connection**. Under transaction-mode pooling, PgBouncer returns the backend to the pool at `COMMIT`; a different client gets it on the next transaction; the cache is invalid. pgx's default `QueryExecMode = QueryExecModeCacheStatement` then hits "prepared statement already exists" errors when the same backend comes back to a client that prepared a differently-named statement on it previously.

**How real is the performance hit?** Plan caching buys roughly 20–40% on hot small queries in benchmarks, but the effect is noisier in production — PostgreSQL's plan cache is invalidated by `ANALYZE`, and the sub-millisecond difference on `SELECT ... WHERE id = $1` matters less than most engineers assume. Heimdall runs a lot of small queries (monitoring loop fetches, pipeline writer inserts, source-filter lookups) and the aggregate cost is *something*, but it's not the difference between "runs fine" and "falls over."

**How tractable is the mitigation?** Two-line pgx config change. Switching to `QueryExecModeExec` (simple query protocol) or `QueryExecModeDescribeExec` (extended protocol without named statements) disables the cache entirely. Parameterised queries still work — you retain SQL-injection safety — you just pay the small cost of per-call planning. This is what Elephantasm does with psycopg (`prepare_threshold=None`) and SQLAlchemy (`postgresql_prepared_statement_cache_size=0`). Supavisor also has evolving server-side prepared-statement support that partially mitigates, but it's not as clean as just turning the client-side cache off.

**Net:** prepared statements are the most-cited blocker and the least load-bearing. Worth disabling in exchange for Supavisor's multiplexing at any meaningful scale.

### 3.2 Future `LISTEN/NOTIFY` (speculative, probably wrong tool anyway)

`pipeline_bus.go` is in-process today — it broadcasts pipeline events to SSE subscribers inside a single Go process. If Heimdall goes multi-replica and we want cross-replica fan-out of pipeline events without standing up new infrastructure, `LISTEN/NOTIFY` is the cheapest answer. It requires a persistent session; transaction-mode pooling returns the backend to the pool at `COMMIT`, breaking the subscription.

**How real is this constraint?** Speculative. It's a future concern, not a current one.

**How load-bearing is `LISTEN/NOTIFY` as the answer?** Questionable. `LISTEN/NOTIFY` has well-known throughput ceilings (thousands of notifies/sec, not millions), no replay, no consumer groups, no durability beyond the subscribing session, and no dead-letter handling. At Series C scale the right answer is **NATS**, **Redis Streams**, or (at higher scale) **Kafka** — proper message buses with the primitives you actually need for a production fan-out path. The argument "we need 5432 so we can use `LISTEN/NOTIFY`" is really "we haven't picked a real message bus yet and we're keeping the option open."

**Net:** this constraint should be retired as a blocker. The right forcing function is: if we go multi-replica, pick a real bus; don't let a hypothetical `LISTEN/NOTIFY` keep the primary pool on a pooling topology that doesn't scale. A small carve-out on 5432 for a specific workload that genuinely needs session persistence (e.g. `pg_try_advisory_lock` on the cron role for scheduler dedupe) is fine — that's not the same as keeping the whole runtime pool on 5432.

### 3.3 Session-scoped GUCs on user-owned Postgres connectors (red herring for this decision)

`internal/connectors/database/postgres.go:88` connects to **customer-supplied** Postgres databases (the DB-activity data source) and enforces `default_transaction_read_only=on` as a session parameter — set once on connect, applies to every transaction on that connection. Under transaction-mode pooling this GUC would reset at every `COMMIT`, and the connector would silently start allowing writes against customer databases. That would be bad.

**But this is about a different pool.** The user-connector code already uses `pgx.Connect` per customer database — it has never been part of Heimdall's own `DATABASE_URL` pool. Moving Heimdall's runtime to 6543 does not touch that code path. The blueprint's §8 notes this explicitly: "This applies to the *downstream* connector, not to Heimdall's own DB."

**Net:** this is not a blocker for the Supavisor migration. It's worth noting only because anyone reading §8 for the first time might conflate the two pools. Call it out, move on.

### 3.4 The unstated real reason: inertia

The blueprint is honest but doesn't say this plainly: **we're on 5432 because we've never been forced off it.** Default pgxpool settings handle current traffic. One backend replica handles current users. The scale trigger hasn't fired. Migrating has a cost; not migrating has a cost (this doc); nobody's had to pay the migration cost yet.

That's a reasonable stance at Heimdall's current scale. It becomes unreasonable the moment any of §4's triggers fires.

---

## 4. When to move — trigger conditions

Any **one** of the following should start the migration:

### 4.1 Connection saturation signal

`pgxpool.Stat().EmptyAcquireCount()` grows steadily — i.e. handlers are waiting for a connection before they can run their query. Today we don't sample this; the blueprint's §11 flags wiring it up as a cheap one-hour change. **Do that first**, independently of this migration, so we have the data. The number to watch is "growing monotonically over a rolling hour," not "nonzero" — transient contention is normal; sustained contention is the trigger.

### 4.2 Replica count × `MaxConns` approaches `max_connections`

Supabase Pro gives 200+ direct connections. If we add a second Fly.io backend and raise `MaxConns` to a proper production value (say 25), we're at 50 connections for the app alone. Add a third replica and we're at 75 — still fine. At 5+ replicas × 25 `MaxConns` = 125+ against a 200-cap database, we're in the "one bad query hangs backends and the pool fills" zone. That's the architectural trigger.

### 4.3 Horizontal scale event

Any decision to go multi-replica for reasons *other* than Postgres capacity (e.g. geographic distribution, HA requirement, zero-downtime deploy) is a natural moment to reassess. The second replica is the one that changes pooling economics — the first replica's overhead is noise.

### 4.4 A customer ask that requires `LISTEN/NOTIFY`-grade fan-out

If a customer needs cross-replica real-time pipeline events *and* we haven't chosen a proper message bus yet, two things happen at once: we need to either pick a bus or carve out a 5432 connection for `LISTEN/NOTIFY`, *and* we're probably at a scale where 6543 makes sense anyway. Treat as a combined trigger.

### 4.5 An explicit enterprise/SOC2-shape commitment

Not a capacity trigger, but worth naming. Supavisor in front of the DB is a standard posture expectation for production SaaS ("connection pooling tier present"). Absence of one isn't a hard fail, but having it in front of an enterprise security review is cleaner than explaining why we don't need it.

### 4.6 Non-triggers (explicitly)

These should **not** motivate the migration on their own:

- **Latency anxiety.** 5432 direct has slightly lower per-query latency than 6543 (one fewer hop). Don't solve a problem you don't have.
- **"Best practices" pressure.** The best practice is "use a pooler when your connection topology justifies it." "Everyone else uses Supavisor" is not sufficient.
- **RLS role split finishing.** These are independent. Finishing the role split and then sitting on 5432 indefinitely is a legitimate end-state; see `rls-enforcement-role-split.md` §8.

---

## 5. Target topology — what 6543 looks like

```
DATABASE_URL       → app_user:<pw>@<project-ref>.pooler.supabase.com:6543/postgres
                       └─ Supavisor, transaction mode.
                       └─ pgxpool.MaxConns = 50+ (Supavisor allows much higher client-side;
                          it multiplexes onto fewer real Postgres backends).
                       └─ Named prepared statements DISABLED in pgx config.
                       └─ Still RLS-enforced as app_user (no change from role split).

CRON_DATABASE_URL  → cron_user:<pw>@<project-ref>.pooler.supabase.com:6543/postgres
                       └─ Supavisor, transaction mode.
                       └─ Same prepared-statement disablement.
                       └─ Still BYPASSRLS as cron_user (no change from role split).

LOCK_DATABASE_URL  → cron_user:<pw>@db.<project-ref>.supabase.co:5432/postgres   (OPTIONAL)
                       └─ Direct 5432, single connection, NOT pooled.
                       └─ Used ONLY for pg_try_advisory_lock in the scheduler's critical sections.
                       └─ Needed only once we go multi-replica. Can be added later.

DIRECT_URL         → postgres:<pw>@db.<project-ref>.supabase.co:5432/postgres
                       └─ Unchanged from the role-split plan. Migrations only.
```

Key configuration changes on the Go side:

```go
// backend/cmd/heimdall/main.go
cfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
if err != nil { /* handle */ }

// Disable named prepared statements — required for Supavisor transaction mode.
cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

// Size the pool for Supavisor multiplexing, not direct-5432 economics.
cfg.MaxConns        = 50                     // was 4-8 default
cfg.MinConns        = 5                      // keep warm for p99 acquisition
cfg.MaxConnLifetime = 30 * time.Minute
cfg.MaxConnIdleTime = 5  * time.Minute

pool, err := pgxpool.NewWithConfig(ctx, cfg)
```

Parallel configuration for the cron pool. Same `DefaultQueryExecMode` change applies.

Optional — but strongly recommended post-`ENABLE_BACKGROUND_JOBS` refactor — a **third, single-connection direct-5432 handle** for advisory locks. Holding `pg_try_advisory_lock` across transactions requires session persistence; under transaction mode the backend returns to the pool at every `COMMIT` and the lock is released. When we go multi-replica, the scheduler's "only one replica runs this investigation" guard breaks if we try to do it through 6543.

---

## 6. Execution plan

Order matters. Each step is independently reversible until the env-var flip.

### 6.1 Wire observability first (no-regret)

Sample `pgxpool.Stat()` every minute in both pools and log/export:
- `TotalConns`, `IdleConns`, `AcquiredConns`, `MaxConns`
- `AcquireCount`, `EmptyAcquireCount` (the important one)

Export to Prometheus with labels `pool="app"` / `pool="cron"`. One-hour change. This happens independently of any migration and pays for itself the first time someone asks "are 502s from pool exhaustion?"

**Gate:** only start §6.2 after you have at least a week of data showing either (a) EmptyAcquireCount rising or (b) a triggered move per §4.

### 6.2 Audit every query for transaction-mode safety

For each query in `backend/internal/db/queries/*.sql`, ask:

- Does it depend on session state surviving `COMMIT`? (Should be no — our RLS pattern uses `SET LOCAL` inside a single transaction.)
- Does it use session-level `SET` (not `SET LOCAL`)? Greppable; should be zero.
- Does it declare a server-side cursor that's intended to outlive the transaction? (Heimdall today uses `LIMIT`/`OFFSET` pagination, so no.)
- Does it use `LISTEN` / `NOTIFY`? (Not today; if it does at migration time, that code needs to move to a different connection handle — see §5 optional lock pool or, preferably, a real message bus.)
- Does it use temporary tables that outlive the transaction? (Should be zero.)

Expected outcome: no query-level changes needed. Confirm, don't assume.

### 6.3 Audit the cron handoff for transaction-mode compatibility

The role-split plan's §5.3 establishes the handoff pattern: cron pool for enumeration, app pool for per-user writes. Under 6543 this pattern **still works** — both pools go through Supavisor, both get per-transaction backends, `SET LOCAL` still scopes identity to the transaction. But confirm:

- No scheduler code calls `pg_try_advisory_lock` against the Supavisor-pooled cron connection (the lock would release at `COMMIT`). If/when advisory locks enter production, they need the optional §5 direct-5432 lock handle.
- No scheduler code uses `cursor`-based pagination that needs to span transactions.
- Monitor loop iterations complete within a single transaction or structure themselves to re-`SET LOCAL app.current_user_id` at the top of each transaction. Today they already do — the per-app iteration opens its own `UserQueries` — so this should be a no-op.

### 6.4 Disable prepared statements and benchmark

**Before** changing any URL, land the `QueryExecMode = QueryExecModeExec` change against 5432 in staging and measure. The expected hit is small (single-digit ms added to p99 on the hottest queries, often lost in network noise). If it's not small — if a specific query surfaces as plan-cache-dependent — we have a problem that's easier to solve on 5432 (where we can reason about the cache) than on 6543 (where Supavisor is also in the picture).

This step is de-risking: it separates "we lost the plan cache" from "we added a pooler hop." If both ship together and something regresses, we don't know which half caused it.

### 6.5 Flip `DATABASE_URL` to 6543 in staging

Order: staging first, then prod. Per pool:

- Flip `DATABASE_URL` to the Supavisor endpoint on 6543.
- Raise `MaxConns` to production value (25–50).
- Set `MinConns` > 0 to keep connections warm.
- Watch for: `prepared statement "..." does not exist` errors (means §6.4 wasn't applied to all code paths); `server_version` advertisement issues (a pgx quirk with Supavisor — usually harmless, worth an extension string in logs to confirm it's benign); and any previously-transient deadlock pattern becoming non-transient (transaction mode serialises slightly differently than direct, though this is rare).
- Bake for 48h in staging before prod.

### 6.6 Flip `CRON_DATABASE_URL` to 6543 in staging

Same treatment, separate flip. The cron pool has lower throughput than the app pool, so any issue here is more likely to be correctness (a writer that was relying on session state we didn't audit) than performance.

### 6.7 (Conditional) Stand up the direct-5432 lock handle

Do this step only if/when advisory locks enter the hot path — today they're a nice-to-have; when Heimdall goes multi-replica they become load-bearing. Isolated in its own config, its own pool (single connection, no pooling), its own review surface. Document in the blueprint as "the one thing we keep on 5432."

### 6.8 Clean up

- Update `docs/blueprints/database-connection-blueprint.md` — §2 endpoint table, §7 pool config recommendations, §8 ("why 5432 today" becomes historical / prior-state discussion), §9 scaling table.
- Update `CLAUDE.md` with the `6543` endpoint in setup instructions.
- Mark this plan's status as `shipped` with the date.
- Cross-reference from `rls-enforcement-role-split.md` §3 port-5432-not-6543 note so future readers see the followup happened.

---

## 7. Risks and what could go wrong

### 7.1 Silent cache dependence

Some query turns out to depend on the prepared-statement cache for acceptable latency. Surfaces as a p99 regression on a specific handler. Mitigation: §6.4 catches this before the pooler flip, and the fix is to either rewrite the hot query or pin that specific call path to a direct-5432 sidecar pool (rare, should be last resort). Likelihood: low.

### 7.2 Supavisor version / behaviour drift

Supavisor is actively developed by Supabase; behaviour around prepared statements, max clients, and edge cases has changed across versions. Mitigation: pin the Supabase project to a known-good Supavisor release during the migration window; subscribe to Supabase's platform changelog; don't migrate in the same sprint as a Supavisor major version bump on their end. Likelihood: moderate over years, low in any given month.

### 7.3 Scheduler breakage we didn't predict

A background writer turns out to have been relying on some subtle session-scoped behaviour nobody grepped for. Mitigation: §6.5/6.6 flip separately; staging bake times; observability from §6.1. If the symptom is "monitor loop throughput drops" rather than an error log, the observability data is how we notice. Likelihood: low-to-moderate.

### 7.4 Advisory locks forgotten

We go multi-replica, forget that `pg_try_advisory_lock` was silently load-bearing, and the scheduler double-fires for a few hours before anyone notices. Mitigation: §6.7 is conditional *specifically because* it's easy to miss; make it an explicit precondition of any multi-replica deploy, written into the runbook for adding a second replica. Likelihood: moderate if unmanaged; low with the precondition.

### 7.5 Rollback not clean

Migration rollback for config-level changes is "flip the env var back" which is clean. But if `MaxConns` was raised to 50 and the 5432 Supabase plan can't handle `50 × replicas` backends, rolling back to 5432 runs into `too many clients` errors. Mitigation: lower `MaxConns` to the pre-migration number before flipping the URL back, not after. Likelihood: low but annoying.

---

## 8. Out of scope (don't bundle)

- **RLS role split.** Separate project, separate doc, must land first and bake. See `rls-enforcement-role-split.md`.
- **Read replicas / query routing.** Orthogonal scaling work. If/when we need read replicas, the read-replica URL is a fourth env var, not a modification of the three runtime URLs.
- **A separate message bus (NATS / Redis / Kafka).** If cross-replica fan-out becomes real, the message-bus selection is its own project with its own trade-offs. This plan should not force that decision.
- **Moving to a dedicated schema** (`CREATE SCHEMA heimdall`). Cleaner long-term, orthogonal to pooling.
- **Supabase plan upgrade.** Triggered by the same scale signal as this plan but is a separate commercial decision.
- **Rewriting RLS policies for performance.** Some multi-join subquery policies would be cheaper as cached `app.current_org_id` reads. Worth doing eventually, not in this PR.

---

## 9. Rough effort estimate

- §6.1 pool observability: **2 hours** (Prometheus metrics + Grafana panel).
- §6.2 query audit: **half a day** (every sqlc query and every handler that uses raw SQL).
- §6.3 cron-handoff audit: **2 hours** once §6.2 is done.
- §6.4 prepared-statement disable + benchmark: **half a day** (two-line code change, most time is measuring).
- §6.5 app pool flip: **1 day of work, 2 days staging bake, 1 day prod flip with observability** — elapsed ~4 days.
- §6.6 cron pool flip: **~2 days elapsed** (same pattern, lower volume).
- §6.7 (conditional) advisory-lock handle: **half a day** if needed; skip otherwise.
- §6.8 cleanup: **2 hours**.

Total active work: **3–4 days**. Total wall-clock with staging baking: **~2 weeks**. This assumes the RLS role split has already landed and the cron pool exists — if either is not true, this plan cannot ship.

---

## 10. Acceptance checklist

Ship-ready when all of these are true:

**Prerequisites**
- [ ] `rls-enforcement-role-split.md` is shipped. `DATABASE_URL` points at `app_user`, `CRON_DATABASE_URL` at `cron_user`, RLS is FORCE-enforced.
- [ ] Pool observability (§6.1) has been running in prod for at least a week with at least one data point triggering the migration (§4.1 EmptyAcquireCount rise, or §4.2 replica count, or an explicit §4.3–4.5 event).

**Code and config**
- [ ] `pgx.QueryExecModeExec` (or `QueryExecModeDescribeExec`) is set on both pools. No code path relies on named prepared statements persisting across transactions.
- [ ] `MaxConns` raised to production value (≥25) on both pools. `MinConns` > 0 on both.
- [ ] Both pools construct cleanly against the Supavisor endpoint in staging and prod.
- [ ] If multi-replica: a dedicated direct-5432 single-connection lock handle exists and is used by `pg_try_advisory_lock` call sites.

**Observational**
- [ ] No `prepared statement ... does not exist` errors in staging over 48h of realistic traffic.
- [ ] No net p99 latency regression on the top-10 handlers vs. the pre-flip baseline.
- [ ] Monitor loop throughput unchanged.
- [ ] Pipeline SSE, webhook/OTLP/syslog/GitHub ingestion all confirmed working under the new topology.

**Cleanup**
- [ ] `docs/blueprints/database-connection-blueprint.md` reflects the new topology (§2 endpoint table, §7 pool tuning, §8 historical framing, §9 scaling table).
- [ ] `CLAUDE.md` setup instructions reflect the `6543` URLs.
- [ ] This plan is marked `shipped` with the date.
- [ ] `rls-enforcement-role-split.md` §3 "Port 5432, not 6543" note is updated to reference this doc as completed.
