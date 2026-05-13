# Three-Role Database Architecture

> **Canonical reference for Trajan's `postgres` / `trajan_app` / `trajan_cron` PostgreSQL role separation.**
> Use this document as a blueprint when porting the same setup to a sibling codebase. It supersedes the scattered
> phase-by-phase docs in `docs/archive/cron-role-phase-*.md`, `docs/archive/post-commit-rls-context-rehydration*.md`,
> and the completion notes in `docs/completions/cron-role-and-bypass-then-scope.md` and
> `docs/completions/known-user-background-task-rls-audit.md`.
>
> **Heimdall analogue (shipped 2026-05-04):** Heimdall's eight-phase port of this topology landed as `app_user` / `cron_user` / `postgres`. See `docs/completions/rls-enforcement-phase-{1..8}.md` for the as-shipped record and the archived [`docs/archive/rls-enforcement-roadmap.md`](../archive/rls-enforcement-roadmap.md) / [`docs/archive/rls-enforcement-role-split.md`](../archive/rls-enforcement-role-split.md) for the execution plan and threat model. The rollout deviated from this reference in three places worth noting for future ports: (a) Heimdall's `cron_user` carries a narrow `DELETE on log_buffer` grant beyond the Trajan baseline (the audit-discovered pruner widening — Phase 1 §5.3 row "Log buffer pruner"); (b) the catalog drift tests (Phase 5 §5.6a) are Heimdall-specific extensions not present in Trajan's allowlist-only approach; (c) the `lookup_user_for_invite` SECURITY DEFINER helper (Heimdall migration 041) is a Phase 7 follow-up the Trajan precedent did not need because Trajan has no email-based invite flow.

---

## 1. The Three Roles

| Role | `BYPASSRLS` | Connection | SQLAlchemy engine | Purpose |
|------|-------------|------------|-------------------|---------|
| `postgres` | yes (superuser) | Direct, port 5432 | `direct_engine` | DDL, Alembic migrations, occasional long-running maintenance tasks. Never serves user traffic. |
| `trajan_app` | **no** (`NOBYPASSRLS`) | PgBouncer pooler, port 6543 | `engine` | All app request handlers. RLS is `FORCE`d on every protected table, so every query must run with `app.current_user_id` set. |
| `trajan_cron` | **yes** (`BYPASSRLS`) | PgBouncer pooler, port 6543 | `cron_engine` | Bootstrap-only: the small set of code paths with **no user identity at entry** — cron jobs, external webhooks, API-key authentication. Used to resolve "which user owns this work?" then immediately hands off to a `trajan_app` session for the actual work. |

### Why three roles, not one?

The single-role status quo (everything as `postgres`) means RLS policies are written but never actually evaluated — `BYPASSRLS` short-circuits them. That's a silent vulnerability: a missing or wrong policy looks fine in code review and passes every test, because no query ever invokes it. Adding `trajan_app` (NOBYPASSRLS) plus `ALTER TABLE … FORCE ROW LEVEL SECURITY` on every protected table makes RLS load-bearing — it is the actual access control, not a defense-in-depth layer.

But that creates a new problem: code paths that genuinely don't have a user identity at entry (a cron tick, an inbound Stripe webhook, an API-key request that hasn't been validated yet) cannot satisfy `app.current_user_id IS NOT NULL` policies. They need a narrow escape hatch. That's `trajan_cron` — it can read across tenants long enough to figure out *whose* work this is, then close that session and reopen one as `trajan_app` with the right context.

`postgres` stays around for migrations because DDL statements (`CREATE TABLE`, `ALTER POLICY`, etc.) require ownership, and because Alembic needs a long-running connection that the transaction pooler can't provide.

```
┌──────────────┐   DDL only       ┌─────────┐
│   postgres   │ ───────────────▶ │   DB    │
│   (super)    │   port 5432      └─────────┘
└──────────────┘                       ▲
                                       │
┌──────────────┐   bootstrap reads     │
│ trajan_cron  │ ────────────────▶─────┤
│  (BYPASSRLS) │   port 6543           │
└──────┬───────┘                       │
       │ resolve user_id               │
       ▼                               │
┌──────────────┐   scoped writes       │
│  trajan_app  │ ────────────────▶─────┘
│ (NOBYPASSRLS)│   port 6543
└──────────────┘   SET LOCAL app.current_user_id
```

`★ Insight ─────────────────────────────────────`
- **"Bypass-then-scope" is the single mental model that makes this whole thing tractable.** Every `cron_engine` session does the smallest possible cross-tenant lookup (usually a single `SELECT owner_id FROM organizations WHERE …`), then closes. The actual mutation reopens as `trajan_app` with `SET LOCAL app.current_user_id` set — so the audit trail and policy enforcement are identical to a normal user request.
- **`SET LOCAL` is transaction-scoped**, which is what makes it safe under PgBouncer transaction pooling — the setting dies with the transaction, so a recycled connection can never leak context to another request. The flip side: every `commit()` drops it, which is what the `after_begin` listener exists to fix (Section 5).
- **`FORCE ROW LEVEL SECURITY` is the load-bearing word.** Plain `ENABLE ROW LEVEL SECURITY` exempts the table owner; `FORCE` does not. Without `FORCE`, `trajan_app` would still bypass RLS on tables it owned, defeating the entire point of the cutover.
`─────────────────────────────────────────────────`

---

## 2. The Layers Touched by the Refactor

The three-role model is not a single change — it ripples through ~7 distinct layers. When porting to another codebase, expect to touch all of them. The order below is the recommended implementation order; each layer assumes the previous one is in place.

| # | Layer | Files (in this repo) | Purpose |
|---|-------|---------------------|---------|
| 1 | **Database roles & grants** | `backend/alembic/versions/b9263ae03d26_*.py`, `b347e95c9611_*.py`, `fbb57f3f107a_*.py` | Create the roles, grant DML, grant `WITH SET TRUE`, FORCE RLS. |
| 2 | **Engine & session setup** | `backend/app/core/database.py` | Three engines, three session makers, `after_begin` listener. |
| 3 | **RLS context helpers** | `backend/app/core/rls.py` | `set_rls_user_context`, `clear_rls_context`, `get_current_rls_user_id`, `RLS_INFO_KEY`. |
| 4 | **Configuration** | `backend/app/config/settings.py`, env vars, Fly secrets | `DATABASE_URL`, `DATABASE_URL_DIRECT`, `DATABASE_URL_CRON`. |
| 5 | **Bypass-then-scope rewrites** | `backend/app/services/scheduler.py`, `api/v1/webhooks.py`, `api/v1/billing.py` (Stripe), `api/deps/api_key_auth.py`, `api/deps/org_api_key_auth.py`, `api/v1/internal.py`, plus all handler files that call those deps | Every code path with no user identity at entry. |
| 6 | **Known-user background-task audit** | Code Map indexing, docs-generation orchestrators, custom doc jobs, AI analysis pipeline, team summary cache, progress callbacks, ~14 other sites | Background tasks that *do* have a user identity but were opening fresh sessions without re-asserting it. |
| 7 | **Tests & tripwires** | `backend/tests/conftest.py`, `backend/tests/integration/conftest.py`, `tests/integration/test_rls_enforcement.py`, `tests/integration/test_rls_policy_coverage.py`, `backend/app/core/rls_policy_allowlist.py` | Opt-in `rls_api_client` fixture, `as_trajan_app` context manager, role-flag assertions, `pg_policies` coverage tripwire, post-commit listener tripwire, `async_session_maker` pairing tripwire. |

---

## 3. Layer 1 — Database Roles & Grants

### 3.1 Create the roles (manual, per environment)

The roles themselves are **not** created by Alembic — they're created by hand in the Supabase SQL editor (or `psql`) per environment. This is intentional: role creation requires superuser, and migrations should be idempotent across envs that may have the role pre-existing under different names or password policies.

```sql
-- App role: NO BYPASSRLS, INHERIT so test fixtures can SET ROLE into it.
CREATE ROLE trajan_app WITH LOGIN NOBYPASSRLS NOSUPERUSER INHERIT
  PASSWORD '<secret>';

-- Cron/bootstrap role: BYPASSRLS, INHERIT.
CREATE ROLE trajan_cron WITH LOGIN BYPASSRLS NOSUPERUSER INHERIT
  PASSWORD '<secret>';
```

The migrations below check for the role and fail loudly if it's absent, so you'll know immediately if you forget this step in a new environment.

### 3.2 Migrations

Three migrations land the role-related schema. They are independent files because each was a separate phase of the rollout, but they could be collapsed in a fresh install.

#### `b9263ae03d26_create_trajan_app_role_and_force_rls.py`
- Grants `trajan_app`: `USAGE` on `public` schema, all CRUD verbs on every table, `USAGE`/`SELECT` on every sequence, `EXECUTE` on the 10 RLS helper functions (`app_user_id()`, `is_org_member(org_id)`, `is_org_admin(org_id)`, `is_org_owner(org_id)`, `has_product_access(product_id, level)`, `can_view_product(product_id)`, `can_edit_product(product_id)`, `can_admin_product(product_id)`, `can_view_repo(repo_id)`, `can_edit_repo(repo_id)`).
- `ALTER DEFAULT PRIVILEGES FOR ROLE postgres … GRANT … TO trajan_app` so future migrations auto-grant.
- PL/pgSQL loop: `ALTER TABLE … FORCE ROW LEVEL SECURITY` on every table where `relrowsecurity = true`.

#### `b347e95c9611_grant_privileges_to_trajan_cron.py`
- Mirror of the above for `trajan_cron`. No `FORCE RLS` call (already done; `BYPASSRLS` makes it moot anyway, but the grants are needed so the role can read at all).

#### `fbb57f3f107a_grant_set_role_to_postgres_for_trajan_app_and_cron.py`
- PG17-only: `GRANT trajan_app TO postgres WITH SET TRUE` and same for `trajan_cron`.
- Lets test infrastructure run `SET LOCAL ROLE trajan_app` from a `postgres`-authenticated session without re-connecting. Production code never uses `SET ROLE` — apps connect directly as the target role via `DATABASE_URL`.
- Downgrade uses `REVOKE SET OPTION FOR` so `INHERIT` is preserved.

### 3.3 The 10 RLS helper functions

These predate the three-role refactor but are load-bearing for it. They're declared `SECURITY DEFINER` so they run with the function owner's privileges (`postgres`) — which lets them read auth tables (`organization_members`, `product_access`) without recursing into the same RLS policies that called them.

If you're porting to a fresh codebase, port these helpers first; the per-table RLS policies are written in terms of them.

### 3.4 The policy sweep (v0.31.9)

Once `trajan_app` was live, the existing RLS policies — many of which were `SELECT`-only because nothing previously *enforced* them — left writes silently failing. Migration `36e4127be7d3_rls_write_policies_post_cutover_sweep.py` added 11 missing `INSERT`/`UPDATE`/`DELETE` policies across 6 tables. The lesson: **if you're cutting over an existing codebase, audit `pg_policies` against your write paths before flipping `DATABASE_URL`** — the `pg_policies` coverage tripwire (Section 8.4) institutionalises this check.

---

## 4. Layer 2 — Engine & Session Setup

### 4.1 Three engines

In `backend/app/core/database.py`:

```python
# App traffic — pooler port 6543, connects as trajan_app.
engine = create_async_engine(
    settings.database_url,
    pool_size=10, max_overflow=20,
    connect_args={"statement_cache_size": 0,    # PgBouncer requirement
                  "prepared_statement_cache_size": 0,
                  "command_timeout": 60},
)

# Migrations & long-running ops — direct port 5432, connects as postgres.
direct_engine = create_async_engine(
    settings.database_url_direct,
    pool_size=3, max_overflow=5,
    connect_args={"command_timeout": 300},
)

# Bootstrap-only — pooler port 6543, connects as trajan_cron.
cron_engine = create_async_engine(
    settings.database_url_cron,
    pool_size=2, max_overflow=3,
    connect_args={"statement_cache_size": 0,
                  "prepared_statement_cache_size": 0,
                  "command_timeout": 30},
)
```

Pool sizes reflect intended use: the cron pool is deliberately tiny (2+3) because every `cron_engine` session should be doing a single small lookup — if you find yourself wanting more, you're probably using the wrong engine.

### 4.2 Three session makers

Parallel structure: `async_session_maker` (RLS-enforced), `direct_session_maker` (DDL), `cron_session_maker` (bypass bootstrap). The cron maker is **not** exposed as a FastAPI dependency on purpose — handlers must `from app.core.database import cron_session_maker` and use it explicitly. This makes `grep cron_session_maker backend/app` an exhaustive list of every bypass site.

### 4.3 The `after_begin` listener

```python
@event.listens_for(Session, "after_begin")
def _rls_after_begin(session, transaction, connection) -> None:
    from app.core.rls import RLS_INFO_KEY
    rls_user_id = session.info.get(RLS_INFO_KEY)
    if rls_user_id is None:
        return
    connection.execute(text(f"SET LOCAL app.current_user_id = '{rls_user_id}'"))
```

This is the structural fix for the post-commit RLS gap. `SET LOCAL` is dropped by every `commit()`; without this listener, any service that committed mid-flight would lose RLS context for every subsequent query on the same session, leading to silent zero-row reads under `trajan_app`. The listener re-issues `SET LOCAL` automatically at the start of every new transaction, reading the user_id from `session.info` (where `set_rls_user_context` stashed it).

No-op for cron and test sessions, which never set the info key.

---

## 5. Layer 3 — RLS Context Helpers

In `backend/app/core/rls.py`:

```python
RLS_INFO_KEY = "rls_user_id"

async def set_rls_user_context(session: AsyncSession, user_id: UUID) -> None:
    session.sync_session.info[RLS_INFO_KEY] = user_id            # for the listener
    await session.execute(
        text(f"SET LOCAL app.current_user_id = '{user_id}'"))    # for *this* txn

async def clear_rls_context(session: AsyncSession) -> None:
    session.sync_session.info.pop(RLS_INFO_KEY, None)
    await session.execute(text("RESET app.current_user_id"))

async def get_current_rls_user_id(session: AsyncSession) -> UUID | None:
    result = await session.execute(text(
        "SELECT NULLIF(current_setting('app.current_user_id', true), '')::uuid"))
    return result.scalar_one_or_none()
```

The two-write design (info key + `SET LOCAL`) is the entire handshake: the immediate `SET LOCAL` covers the current transaction, the info key covers every subsequent transaction on the same session.

**Invariant:** every site that opens an `async_session_maker()` and runs RLS-protected queries must call `set_rls_user_context(db, user_id)` before the first such query. The listener will then keep that context alive across commits for the rest of the session's life.

---

## 6. Layer 4 — Configuration

### 6.1 Settings

In `backend/app/config/settings.py`:

```python
database_url: str         # → DATABASE_URL,        trajan_app   @ pooler:6543
database_url_direct: str  # → DATABASE_URL_DIRECT, postgres     @ direct:5432
database_url_cron: str    # → DATABASE_URL_CRON,   trajan_cron  @ pooler:6543
```

### 6.2 Environment variables / secrets

| Variable | Local (`backend/.env`) | Production (Fly secret) |
|----------|------------------------|--------------------------|
| `DATABASE_URL` | `…@aws-0-…pooler.supabase.com:6543/postgres` as `trajan_app` | same, prod credentials |
| `DATABASE_URL_DIRECT` | `…@db.<ref>.supabase.co:5432/postgres` as `postgres` | same, prod credentials |
| `DATABASE_URL_CRON` | `…@aws-0-…pooler.supabase.com:6543/postgres` as `trajan_cron` | same, prod credentials |

The Fly cutover for production was the final irreversible step of the rollout — see `docs/archive/cron-role-phase-4-tests-and-verification.md` for the cutover playbook. The revert lever is "flip `DATABASE_URL` back to `postgres`"; takes <60 seconds.

---

## 7. Layer 5 — Bypass-then-Scope Rewrites

This is the largest layer by line count and the one most likely to bite you in a sibling codebase. Every code path with no user identity at entry needs the same shape:

```python
# 1. Bootstrap: cross-tenant lookup under BYPASSRLS.
async with cron_session_maker() as cron_db:
    owner_user_id = await _resolve_owner(cron_db, …)

# 2. Scope: scoped session with RLS context.
async with async_session_maker() as db:
    await set_rls_user_context(db, owner_user_id)
    # … real work …
```

### 7.1 Cron jobs — `backend/app/services/scheduler.py`

5 jobs, all rewritten to bypass-then-scope:
- `run_auto_progress()`
- `run_plan_prompt_emails()`
- `run_weekly_digest()`
- `run_daily_digest()`
- `run_overage_reconciliation()`

Helper `_resolve_org_owners()` does a single batch `SELECT id, owner_id FROM organizations` under cron, then each unit of work opens its own `async_session_maker()` scoped to that owner. The advisory lock that prevents duplicate runs is held on the `cron_session_maker` connection (non-RLS-protected; safe to bypass).

### 7.2 Internal cron-trigger endpoints — `backend/app/api/v1/internal.py`

4 endpoints: `/internal/auto-progress`, `/internal/send-plan-prompt-emails`, `/internal/send-weekly-digest`, `/internal/send-daily-digest`. Authenticated by `X-Cron-Secret` header only. Each delegates to `scheduler.trigger_now(job_id)` rather than opening its own session — so the bypass-then-scope discipline lives in exactly one place (the scheduler).

### 7.3 GitHub webhook — `backend/app/api/v1/webhooks.py`

1. Verify HMAC-SHA256 signature.
2. `_resolve_installation_owner()` on `cron_session_maker` → `(org_id, owner_user_id)`. Returns `None` if installation row doesn't exist (legitimate during install race).
3. `async_session_maker()` + `set_rls_user_context(db, owner_user_id)` → mutate `github_app_installations`, `github_app_installation_repos`.

### 7.4 Stripe webhook — `backend/app/api/v1/billing.py`

1. Verify Stripe signature, parse event.
2. `cron_session_maker()` → resolve `customer_id → organization → owner_user_id` (single JOIN on `subscriptions JOIN organizations`).
3. `async_session_maker()` + `set_rls_user_context(db, owner_user_id)` → idempotency check (must be inside the scoped session — `billing_events` is RLS-protected) → handler.

### 7.5 API-key authentication

#### Product API keys — `backend/app/api/deps/api_key_auth.py`
- `get_api_key()` validates `Bearer trj_pk_…` on `cron_session_maker()` (BYPASSRLS — `product_api_keys` has a SELECT policy requiring `app_user_id()`, which is null at this point).
- Returns the `ProductApiKey` record. Caller is responsible for opening the scoped session.
- Callers: `public_tickets.py` (4 endpoints), `mcp.py` (~15 endpoints).

#### Organisation API keys — `backend/app/api/deps/org_api_key_auth.py`
- `get_org_api_key()` validates `Bearer trj_org_…` on `cron_session_maker()`.
- Returns `PartnerAuthContext(api_key, effective_user_id)`. The `effective_user_id` is `api_key.created_by_user_id` if set, else falls back to `organizations.owner_id` (so a key created by a user who has since been deleted still works under the org owner's identity).
- Callers: `partner.py` (6 endpoints), `partner_config.py` (1 endpoint).

Every caller follows:

```python
async with async_session_maker() as db:
    await set_rls_user_context(db, api_key.created_by_user_id)  # or ctx.effective_user_id
    # … actual handler logic …
```

### 7.6 Recap — every `cron_session_maker()` site

```bash
grep -rn "cron_session_maker()" backend/app --include="*.py"
```

Should return *only* the files in 7.1–7.5. If a new file appears in this list, it's a new bypass site and needs the bypass-then-scope review. The test suite has a tripwire (Section 8.5) that flags drift in the opposite direction.

---

## 8. Layer 6 — Known-User Background Tasks

Distinct from Layer 5: these code paths *do* have a user identity at entry (a request handler kicked off the work), but they were opening fresh `async_session_maker()` sessions in the background without re-asserting RLS context. Under the old `postgres` role, this didn't matter; under `trajan_app`, every fresh session starts with `app.current_user_id IS NULL` and silently returns zero rows.

The audit (v0.31.3) hit ~14 sites:
- Code Map indexing (sub-task workers).
- Docs-generation orchestrators and the changelog/blueprint/plans sub-agents.
- Custom doc generation jobs.
- AI analysis pipeline.
- Team summary cache regeneration.
- Progress callbacks in long-running orchestrators.
- Fingerprint storage, changelog progress tracking.
- Completion/failure status marking helpers.
- AI feedback interpretation.

**Fix pattern:** thread `user_id` into the task's parameters (don't try to read it from request context — the request is gone), and call `set_rls_user_context(db, user_id)` inside the fresh-session block before any RLS-protected query.

**Reference implementations** to mimic when porting:
- `backend/app/api/v1/products/analysis.py` (the `set_rls_user_context` call inside the background task body).
- `backend/app/api/v1/products/docs_generation.py` (same pattern, with sub-agent fan-out).

The `after_begin` listener (Section 4.3) means you only need to call `set_rls_user_context` *once* per session — every subsequent commit-and-continue pattern is automatically rehydrated.

---

## 9. Layer 7 — Tests & Tripwires

### 9.1 Default test posture: bypass

The root `conftest.py` `db_session` fixture connects via the direct (port 5432) engine as `postgres`, in a SAVEPOINT-rolled-back transaction. The `api_client` fixture overrides `get_db` / `get_db_with_rls` / `get_current_user` to use this session. **Default behaviour is RLS-bypassed** — almost all unit and integration tests run as `postgres` so they can seed fixtures without policy friction.

### 9.2 Opt-in `rls_api_client` fixture

For tests that need to verify RLS actually denies/allows under `trajan_app`:

```python
def test_user_cannot_see_other_orgs_products(rls_api_client, …):
    response = rls_api_client.get(f"/api/v1/products/{other_org_product_id}")
    assert response.status_code == 404
```

The fixture calls `_activate_trajan_app_role(db_session, test_user.id)`, which:
1. Sets `session.info[RLS_INFO_KEY] = user_id`.
2. `SET LOCAL app.current_user_id = '{user_id}'`.
3. `SET LOCAL ROLE trajan_app`.
4. Skips the test if the role doesn't exist (so CI on a fresh DB doesn't break).

### 9.3 `as_trajan_app` context manager

In `backend/tests/integration/conftest.py`. Async context manager that wraps a block in `SET LOCAL ROLE trajan_app` / `RESET ROLE`. Used inside individual integration tests that need to flip role mid-test.

### 9.4 RLS enforcement & coverage tripwires

In `backend/tests/integration/test_rls_enforcement.py`:
- **`test_trajan_app_has_no_bypass_flag`** — asserts `pg_roles.rolbypassrls = false` for `trajan_app`. Catches a DBA accidentally flipping the flag.
- **`test_trajan_cron_connects_with_bypassrls`** — connects via `cron_session_maker()` and asserts `current_user = 'trajan_cron'` and `rolbypassrls = true`.
- **`test_every_rls_enabled_table_has_force_on`** — scans `pg_class` for any RLS-enabled table where `relforcerowsecurity = false`. New tables that enable RLS but forget `FORCE` show up here.
- **`TestRlsContextSurvivesCommit`** — two tests. First sets context, does `begin_nested` + commit, asserts the next query still sees the user_id. Second clears context and confirms the listener no-ops (so a bypass session can't accidentally rehydrate). Together these guard the `after_begin` listener.
- **`TestBackgroundTaskRlsContext.test_async_session_maker_pairs_with_set_rls_user_context`** — file-level grep tripwire. Every file in `backend/app` that calls `async_session_maker()` must also reference `set_rls_user_context` (or be in the explicit allowlist for genuinely RLS-irrelevant infrastructure). Catches new background-task sites that forget the rehydration call.

In `backend/tests/integration/test_rls_policy_coverage.py`:
- **`test_every_rls_table_in_allowlist`** — every RLS-enabled table must be listed in `RLS_POLICY_ALLOWLIST` (Section 9.5).
- **`test_every_allowlist_table_is_rls_enabled`** — reverse check; catches stale allowlist entries.
- **`test_policy_cmds_cover_allowlist`** — for each table, scans `pg_policies` and asserts the recorded verbs cover at least the allowlist's required set. This is the tripwire that would have caught the v0.31.9 class of bug (write policies missing post-cutover, writes silently rejected).

### 9.5 The policy allowlist

`backend/app/core/rls_policy_allowlist.py` declares the source-of-truth mapping `table → required verbs`. Snapshot at the time of writing (~33 tables):

| Verb shape | Tables |
|------------|--------|
| **Full CRUD** (`{SELECT, INSERT, UPDATE, DELETE}`) | 28 tables — the bulk of user-owned data: `users`, `organizations`, `products`, `repositories`, `documents`, `work_items`, `subscriptions`, `org_api_keys`, `product_api_keys`, etc. |
| `{INSERT, SELECT}` | `commit_stats_cache`, `billing_events`, `referral_codes` (append-only) |
| `{SELECT, UPDATE}` | `discount_codes` (admin-managed, no user delete) |
| `{INSERT, SELECT, DELETE}` | `discount_redemptions` |
| `{INSERT, SELECT, UPDATE}` | `feedback`, `org_api_keys` (no end-user delete; admin-only via support) |
| `{SELECT}` | `announcement`, `usage_snapshots` (server-written only) |

When you add a new RLS-protected table, you **must** add it to this allowlist or `test_every_rls_table_in_allowlist` fails. When you change which verbs a table's policies cover, update the allowlist to match — that's the contract that future-you and reviewers can read at a glance.

### 9.6 PG17 `WITH SET TRUE` requirement

The opt-in `rls_api_client` fixture and `as_trajan_app` context manager both rely on `SET LOCAL ROLE trajan_app` working from a `postgres`-authenticated test connection. PG14–PG16 allowed this implicitly via `INHERIT`; PG17 requires explicit `GRANT trajan_app TO postgres WITH SET TRUE`. Migration `fbb57f3f107a` does this. If you're porting to a Postgres version older than 17, the `WITH SET TRUE` clause is harmless (older versions ignore it).

---

## 10. Porting Checklist

When applying this blueprint to a sibling codebase:

### 10.1 Database setup
- [ ] Identify every table that stores tenant data; enable RLS + `FORCE RLS` on each.
- [ ] Write or port the `app_user_id()` and access-check helpers as `SECURITY DEFINER` functions.
- [ ] Ensure every protected table has policies covering every verb the app uses (don't trust `SELECT`-only policies if you also write).
- [ ] Create `app_role` and `cron_role` (rename to taste) in each environment manually.
- [ ] Write three Alembic migrations: privileges-for-app, privileges-for-cron, `WITH SET TRUE` (if PG17+).

### 10.2 Engine layer
- [ ] In your DB module, declare three engines and three session makers.
- [ ] Add the `after_begin` listener and the `info`-key handshake.
- [ ] Add `set_rls_user_context`, `clear_rls_context`, `get_current_rls_user_id` helpers.

### 10.3 Configuration
- [ ] Add `DATABASE_URL_DIRECT` and `DATABASE_URL_CRON` env vars + settings fields.
- [ ] Set the corresponding secrets in every deployment environment.

### 10.4 Audit & rewrite
- [ ] `grep` every site that opens a fresh DB session in the background. For each, decide: known-user (call `set_rls_user_context`) or no-user-at-entry (bypass-then-scope).
- [ ] Rewrite cron jobs to enumerate work cross-tenant on the cron session, then per-unit reopen on the app session with context set.
- [ ] Rewrite webhooks (every external sender) to bootstrap-then-scope.
- [ ] Rewrite API-key auth deps to validate on cron session, return identity, let handlers reopen on app session.

### 10.5 Tests
- [ ] Default test session = bypass; add an opt-in `rls_*_client` fixture for RLS-asserting tests.
- [ ] Add the four enforcement tripwires: `BYPASSRLS` flag check, `FORCE RLS` table scan, post-commit listener pair, `async_session_maker` grep tripwire.
- [ ] Add the policy allowlist module and the three coverage tripwires.

### 10.6 Cutover
- [ ] Run the full test suite under the new role-aware setup before flipping any production secret.
- [ ] In production: flip `DATABASE_URL` to the app role *last*, after verifying every other layer is in place. Have the revert ready (`DATABASE_URL` back to `postgres`).
- [ ] Watch for silent-zero-row symptoms in the first 24 hours: empty cron output, webhook events dropped, API keys returning 401 — each maps to a layer above that you may have missed.

---

## 11. Useful greps

```bash
# Every bypass-then-scope site:
grep -rn "cron_session_maker()" backend/app --include="*.py"

# Every place that asserts RLS context (should pair with cron sites + background tasks):
grep -rn "set_rls_user_context" backend/app --include="*.py"

# Background-task pairing tripwire (manual variant of the test):
for f in $(grep -rl "async_session_maker()" backend/app --include="*.py"); do
  grep -q "set_rls_user_context\|get_db_with_rls" "$f" || echo "MISSING: $f"
done

# DB role inspection (run via psql, not the app):
SELECT rolname, rolbypassrls, rolsuper FROM pg_roles
  WHERE rolname IN ('postgres', 'trajan_app', 'trajan_cron');

# Verify FORCE RLS is on everywhere:
SELECT relname FROM pg_class
  WHERE relkind='r' AND relrowsecurity AND NOT relforcerowsecurity;

# Verify policy coverage for a table:
SELECT cmd, polname FROM pg_policies WHERE tablename = '<table>';
```

---

## 12. Source documents (now superseded by this blueprint)

These remain in the repo for historical context but should not be the starting point for new work:

- `docs/archive/cron-role-phase-1-plumbing.md` — engine/session/setting plumbing.
- `docs/archive/cron-role-phase-2-bypass-then-scope.md` — cron + scheduler rewrite.
- `docs/archive/cron-role-phase-3-webhooks-and-api-keys.md` — webhook + API-key rewrite.
- `docs/archive/cron-role-phase-4-tests-and-verification.md` — test infra + Fly cutover playbook.
- `docs/archive/cron-role-plan-review.md` — design review.
- `docs/archive/post-commit-rls-context-rehydration.md` and `…-plan.md` — `after_begin` listener.
- `docs/archive/known-user-background-tasks-rls-context.md` — Layer 6 audit.
- `docs/archive/background-task-rls-context-tripwire.md` — Section 9.4 tripwire design.
- `docs/archive/test-suite-rls-enforcement-coverage.md` — Layer 7 design.
- `docs/archive/billing-events-rls-followups.md` and the two `docs/completions/billing-events-rls-insert-policy*.md` — `billing_events` post-cutover fix.
- `docs/completions/cron-role-and-bypass-then-scope.md` and `docs/completions/known-user-background-task-rls-audit.md` — high-level completion writeups.
- `docs/completions/rls-write-policy-sweep-post-trajan-app-cutover.md` and `…-pickup-notes.md` — v0.31.9 policy sweep.
