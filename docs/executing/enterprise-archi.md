# Enterprise Architecture: BYOK + BYODB

**Status**: Thinking / proposal. No code yet.
**Date**: 2026-04-18
**Author**: Agent pass, for review.

---

## Original prompt (preserved)

> I'm considering adding the ability to:
> 1. BYOK for your Agent
> 2. Persist anything/everything to your own db instead of ours
>
> This would mean you use and pay for Heimdall only as a service/harness. Full data persistence and processing ultimately happens on your end. It technically still runs through us, but everything is ephemeral.
>
> Thought this could be an interesting Enterprise option without being full on-prem deployed.

---

## Strategic take

The idea is good and it slots naturally between two existing points on the spectrum:

| Offering            | Data lives where     | Compute runs where | Ops burden on customer |
|---------------------|----------------------|--------------------|------------------------|
| Hosted SaaS (today) | Our Postgres         | Our Fly.io         | None                   |
| **Enterprise (this proposal)** | **Customer Postgres** | **Our Fly.io**     | **DB only**            |
| Full on-prem (future?) | Customer Postgres    | Customer K8s       | Everything             |

This "mid tier" is exactly what unlocks regulated buyers (HIPAA, FINRA, PCI, EU data-residency, gov) who can't let logs and AI-generated incident reports sit in a third-party DB — but don't want to run the whole stack themselves. It's also where a lot of contractual friction lives: "your logs never leave our VPC" is a clean answer for security review, even though in practice they *do* transit our process.

**One framing to be honest about up front**: this offering is *not* a privacy guarantee. Logs still pass through our Go process, the agent still runs on our infra, and the Claude API call still originates from our IP. What the customer gets is "nothing is persisted on our side." That's a real, sellable property, but we should position it as **data residency**, not **zero-trust isolation**.

The two asks are very differently sized:

- **BYOK** is roughly a day of work. It's additive, touches `config.go` and the agent provider dispatch, and has no migration story.
- **BYODB** is structural. Every query in the system currently runs through a single `*pgxpool.Pool` opened at startup. Enterprise mode makes that pool plural and keyed by org. This is the hard problem.

I'd ship them in separate releases.

---

## BYOK — proposal

### Current state

- `ANTHROPIC_API_KEY` is a single deployment-wide env var, validated as required at boot (`backend/internal/config/config.go:62-64`).
- `OPENROUTER_API_KEY` is an optional deployment-wide env var; the OpenRouter provider is only registered if it's set (`backend/internal/agent/agent.go:48-50`).
- `appConfig.Provider` (string on `AppAgentConfig`) already selects which provider the agent uses per-app; the *key* is global.

### Delta

1. Add `enterprise_config` columns to `organizations` (or a side table if it grows):
   - `anthropic_api_key_encrypted bytea NULL`
   - `openrouter_api_key_encrypted bytea NULL`
   - `enterprise_mode boolean NOT NULL DEFAULT false`
2. Introduce an `OrgKeyResolver` interface in `internal/agent/` with order: **per-org override → deployment default → error**.
3. Encrypt at rest using a KMS-backed key (AWS KMS, GCP KMS, or libsodium secret-box with a deployment-level root). Never write plaintext to logs, never return it on GET.
4. Plumb `orgID` into the code path that instantiates the provider. Today `agent.Agent` is constructed with provider objects; we'd refactor to lazy provider construction at the point we know the org.
5. Settings UI: org-scoped "API Keys" tab with write-only + rotate + test-connection actions.

### Blast radius
- ~1 migration, ~1 new handler, ~1 UI page, a refactor to `agent.New(...)` / `loop.Run(...)` to pass `orgID` where it isn't already passed.
- Zero impact on SaaS customers; new columns default to NULL and fall through to env.

`★ Insight ─────────────────────────────────────`
Because `CLASSIFIER_MODE` and the Lumber ONNX model are bound at process start (`main.go:52-75`), they're naturally deployment-wide and don't need BYOK equivalents — classification is stateless and doesn't bill per-call. Only the Claude/OpenRouter APIs have a per-call cost worth passing through.
`─────────────────────────────────────────────────`

---

## BYODB — the real problem

### What "BYODB" has to preserve

Whatever we build must still give us:

1. **Per-user RLS**. `s.UserQueries(ctx, userID)` begins a tx and runs `SELECT set_config('app.current_user_id', $1, true)` before any query (`backend/internal/handlers/userqueries.go:24-30`). Every RLS policy in `migrations/013_enable_rls.up.sql` keys off that GUC. This pattern has to work on the customer's DB too — which means we need to install our policies and our `app_current_user_id()` function there.
2. **Migration evolution**. Today `make migrate-up` runs `golang-migrate` against one DB. With BYODB we need to apply migrations to N customer DBs on every release, ideally automatically and safely.
3. **Connector writes**. The webhook ingestion path writes to `log_buffer` on the hot path. This has to hit the customer's DB with low latency. Pool config, connection reuse, and customer DB outages all become production concerns.
4. **Control plane integrity**. Supabase auth, billing, feature flags, per-org enterprise config — these can't move to the customer DB (chicken-and-egg: we need them *before* we know which DB to connect to).

### Three design options

#### Option A — **Control Plane + Data Plane Split** (recommended)

Heimdall operates two classes of Postgres:

- **Control plane** (our DB, always): the minimum to route a request.
  - `users` (mirror of Supabase `auth.users`)
  - `organizations` (+ `enterprise_mode`, `data_plane_dsn_encrypted`, `anthropic_api_key_encrypted`, ...)
  - `billing_*`, `audit_log`, feature flags
- **Data plane** (our DB for SaaS, customer's DB for Enterprise): everything tenant-scoped.
  - `applications`, `app_agent_config`, `app_monitoring_state`, `app_source_filters`
  - `connections`, `connection_sources`
  - `log_buffer`, `log_pipeline_events`
  - `conversations`, `agent_log`, `investigations`, `notification_log`

Query routing becomes: `userID → orgID → data-plane DSN → pgxpool.Pool (cached)`.

`UserQueries` changes from:
```go
func (s *Server) UserQueries(ctx, userID) (*db.Queries, Tx, error) {
    tx := s.pool.Begin(ctx)
    tx.Exec("SELECT set_config('app.current_user_id', $1, true)", userID)
    return db.New(tx), tx, nil
}
```
to:
```go
func (s *Server) UserQueries(ctx, userID) (*db.Queries, Tx, error) {
    pool := s.tenantPools.For(ctx, userID) // resolves via control plane
    tx := pool.Begin(ctx)
    tx.Exec("SELECT set_config('app.current_user_id', $1, true)", userID)
    return db.New(tx), tx, nil
}
```

A new `internal/tenantdb/` package owns:
- DSN lookup (reads `organizations.data_plane_dsn_encrypted` via the control plane, decrypts, memoises)
- An LRU of `*pgxpool.Pool` keyed by org ID (bounded — pools are expensive; idle eviction after N minutes)
- Health checks (surface "your DB is unreachable" in the UI instead of 500s)
- Migration version check on pool check-out (refuse to serve if customer DB is at a lower migration than the running binary expects)

**Pros**: clean separation, easy to reason about, migration logic is mechanical, Heimdall-as-harness story is clean.
**Cons**: two DBs to think about per deployment; split migrations (see below); webhook ingestion hot path now has a tenant lookup hop.

#### Option B — **Virtual Schema per Tenant** (one physical DB, N schemas)

Customer provides one Postgres URL. Heimdall creates `heimdall_<org_id>` schema per org. All queries are issued with `SET search_path = heimdall_<org>,public`.

This is basically the SaaS model with schema-per-tenant instead of row-per-tenant-via-RLS. It's a well-trodden pattern (PostgREST, Citus) but solves a problem we don't have: we already have RLS, and enterprise customers typically map 1:1 with a single org anyway. I don't think the flexibility is worth the rebuild.

**Not recommended** — extra complexity, no enterprise selling point over Option A.

#### Option C — **Pure Ephemeral** (no control plane, customer DB owns everything)

Every table moves to the customer DB including `users` and `organizations`. Heimdall is pure code; zero persistent state.

Philosophically appealing but practically broken: we still need to authenticate requests before we know which DB to talk to (Supabase JWTs need a user lookup), bill usage, track enterprise contracts, and hold the encrypted DSNs themselves. Those things *have* to live somewhere we own.

You could partially rescue this with "the DSN is derived from the JWT claims" (e.g., a `heimdall_db_url` custom claim set at provisioning time) — but you've just re-invented the control plane inside Supabase auth, which is worse ergonomics than owning a small control-plane DB.

**Not recommended**.

### Recommended architecture: Option A, fleshed out

```
 ┌──────────────────────────────┐
 │  Heimdall (our Fly.io)       │
 │                              │
 │  ┌────────────┐              │
 │  │ Handlers   │              │
 │  └─────┬──────┘              │
 │        ▼                     │
 │  ┌────────────────┐          │
 │  │ tenantdb.For() │──────────┼──> Control plane DB (ours)
 │  └─────┬──────────┘          │    · users, orgs, billing
 │        ▼                     │    · enterprise_config
 │  ┌──────────────┐             │      (DSN, BYOK, plan)
 │  │ pgxpool per  │              │
 │  │ org (LRU)    │              │
 │  └─────┬────────┘              │
 │        ▼                       │
 │  ┌────────────────┐ ─── SaaS ──┼──> Shared data plane DB (ours)
 │  │ UserQueries()  │            │    (tenant rows, RLS)
 │  │ + SET LOCAL    │ ─── Ent. ──┼──> Customer data plane DB
 │  │   app.current_ │             │    (tenant rows, RLS,
 │  │   user_id      │             │     schema installed by us)
 │  └────────────────┘             │
 └─────────────────────────────────┘
```

### Provisioning flow ("Connect your DB" wizard)

A new `/settings/org/enterprise/database` page runs the following server-side flow:

1. **Collect connection**: DSN, optional app-name, optional CA cert (for managed Postgres with custom root CAs).
2. **Probe**: open a throwaway connection, check `SELECT version()` (require Postgres 14+), check extensions (`pgcrypto`, `pg_trgm` if we use it), check that the connection role has `CREATE` on the target schema.
3. **Emptiness check**: refuse if any `heimdall_*` or expected-named table already exists, unless `?adopt=true` and we recognise the schema version from `schema_migrations`.
4. **Migrate**: run every data-plane migration in order (see below), capturing version in the customer's own `schema_migrations` table.
5. **Smoke test**: write-and-rollback into every RLS-protected table to confirm policies + `app.current_user_id` GUC work end-to-end.
6. **Commit**: encrypt DSN, store in `organizations.data_plane_dsn_encrypted`, flip `enterprise_mode = true`.
7. **Cutover**: subsequent `UserQueries(userID)` calls for this org route to the new pool. (Migration of *existing* data from SaaS → customer DB is a separate, optional step — see Open Questions.)

### Migration story

We need to split `backend/migrations/` into two sources:

```
backend/migrations/
  control/       # users, organizations, billing, audit_log, ...
    001_initial.up.sql
    ...
  data/          # applications, connections, log_buffer, agent_log, ...
    001_initial.up.sql
    ...
    038_log_pipeline_events.up.sql
```

`make migrate-up` continues to work for our own control plane and shared data plane; it just targets the appropriate directory per DB.

For **customer data planes**, ship an in-process migrator instead of relying on CLI:

- Use `golang-migrate/migrate/v4` with `source/iofs` pointing at `go:embed migrations/data/*.sql`.
- Expose two operations: `Migrate(dsn)` for wizard-triggered first run, and `ApplyPending(dsn)` called on app startup for every enterprise org.
- Enforce the rule that data-plane migrations are **append-only and backward-compatible with N-1** — so rolling-deploy Heimdall ≠ bricking customer DBs at a pinned older version. (We already had an incident around this: `0.46.4` is literally a schema-drift audit release. We should formalise what `0acf558` implicitly learned.)

### Hot path: webhook ingestion

The webhook ingestion endpoint already does `userID → org → app → connection → insert into log_buffer`. Under Option A it becomes `userID → org → **tenant pool** → connection → insert`. Two concerns:

1. **Latency**: tenant-pool lookup is a map read (O(1) hot, O(DSN decrypt + control-plane SELECT) cold). Cold lookups need to be rare — warm the pool on first ingestion of the day and keep it open.
2. **Backpressure**: if a customer DB goes down, ingestion requests pile up. Need a circuit-breaker per tenant pool: after N consecutive failures, return 503 from the webhook and buffer upstream (Fly.io Log Shipper retries, so dropping briefly is survivable).

### What stays in the control plane

| Table              | Why stays |
|--------------------|-----------|
| `users`            | Needed to authenticate before we know the tenant DB |
| `organizations`    | Owns the DSN itself (chicken-and-egg) |
| `enterprise_config`| API keys, entitlements, plan |
| `billing_*`        | Usage records must survive customer DB loss |
| `audit_log`        | Our own operational log; not the customer's |

### What moves to the data plane

Everything with a user-visible tenant payload: applications, connections, connection_sources, app_agent_config, app_monitoring_state, app_source_filters, log_buffer, log_pipeline_events, conversations, agent_log, investigations, notification_log.

---

## Rollout sketch

1. **Phase 0 — Refactor (no customer-visible change)**: introduce `tenantdb.Resolver`, make `UserQueries` route through it, but have it always return the single shared pool. Ship. Verify no perf regression.
2. **Phase 1 — Split migrations**: reorganise `backend/migrations/` into `control/` and `data/`. Verify the combined-apply ordering produces an identical schema to today (this is the kind of thing `0.46.4` would catch).
3. **Phase 2 — BYOK**: ship per-org API keys. Independent deliverable, no BYODB prerequisite.
4. **Phase 3 — In-process migrator**: embed `data/` migrations, expose `ApplyPending(dsn)`. No UI yet; drive via admin CLI for the first customer.
5. **Phase 4 — Provisioning wizard**: ship the UI, first design-partner enterprise org goes live.
6. **Phase 5 — Operational hardening**: circuit breaker, migration-version gate on pool check-out, DR runbook for "customer DB is down."

Each phase is independently shippable and reversible, which is how this codebase has been operating (cf. the .0→.1→.2→.3 assessment cadence visible in the changelog).

---

## Open questions — for review

1. **MVP scope**: should v1 be "BYOK only" to validate enterprise appetite, with BYODB deferred until we have at least one paying enterprise design partner asking for it? Or do we need the full story to *win* the first enterprise deal?
2. **Migration ownership**: who runs data-plane migrations on customer DBs — Heimdall auto-applies on startup (convenient, risky if a migration is slow or locks), or the customer triggers from the UI (safer, but means a stale Heimdall talking to a stale customer DB)? I lean auto-apply with a "dry-run" preview and a per-org "paused" flag.
3. **Migrating existing data**: when a SaaS org flips to Enterprise, do we export → import their historical logs/conversations/reports into the new DB, or start fresh? Export is a *lot* of code; "start fresh" is cleaner but a rough sell.
4. **Supported Postgres flavours**: strict vanilla Postgres ≥ 14, or also Aurora, CockroachDB, Yugabyte? Aurora is easy (it's Postgres). Cockroach/Yugabyte will break us in subtle ways (RLS semantics, `SET LOCAL`, specific extensions).
5. **Network model**: does Heimdall reach the customer DB over the public internet (TLS + IP allowlist), over a Tailscale tailnet, or via a customer-run outbound tunnel (PrivateLink / reverse-proxy)? Each has a different security-review narrative.
6. **What about the classifier**: Lumber ONNX is loaded from `LUMBER_MODEL_DIR` at boot. Do enterprise customers want per-org classifier weights (BYOM)? Probably not for v1, but worth confirming before the Settings page IA gets designed.
7. **Pricing**: is Enterprise a flat tier plus BYOK/BYODB toggles, or is BYODB a specific SKU? If BYOK means we're not paying Anthropic, does the base subscription fee drop accordingly, or do we hold the margin as "you're paying for the harness, not the inference"?
8. **Disaster posture**: if a customer DB is unreachable for 10 minutes, what does the product *look like* — a dashboard saying "your DB is offline," or generic 500s? I lean toward a banner + degraded mode; confirm?
9. **Audit surface**: when Heimdall reads/writes a customer DB, do we need to emit a per-query audit record into *their* DB for their SOC2 auditors, or is our own control-plane audit log enough?
10. **Test strategy**: do we run the full backend integration suite against a matrix of (shared DB, BYODB) permutations in CI, or only against one and trust the `tenantdb` abstraction? I'd argue for both — otherwise we won't catch RLS/GUC bugs that only appear under split-DB routing.
