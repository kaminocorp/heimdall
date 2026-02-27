Our db tables should have RLS policies in place. We should not just add these directly to the DB via SQL commands, but through something akin to a migration script (if possible), so we keep track of it for future new db deployments.

---

# RLS Policy Implementation Plan

## Current State

- **10 migrations** exist (`001`–`010`), all as `.up.sql`/`.down.sql` pairs in `backend/migrations/`.
- **Zero RLS policies** are enabled on any table.
- The backend connects via a **direct `pgx` pool** to `DATABASE_URL` (not the Supabase client SDK). Auth is enforced at the **application layer** — the middleware extracts `user_id` from a Supabase JWT and every query manually filters with `WHERE user_id = $1`.
- Application-level filtering is good practice but not sufficient on its own: a single missed WHERE clause, a new query, or a raw SQL debug session could leak data across users. RLS provides defence-in-depth at the database level.

### Table Inventory

| Table | Has `user_id` | Current Scoping | RLS-Ready? |
|-------|--------------|-----------------|------------|
| `users` | N/A (is the identity table) | Auth-synced | Needs policy (users should only read their own row) |
| `connections` | Yes | Direct `WHERE user_id` | Yes |
| `conversations` | Yes | Direct `WHERE user_id` | Yes |
| `agent_log` | Yes | Direct `WHERE user_id` | Yes |
| `log_buffer` | No | Via `JOIN connections` | Needs `user_id` column first |
| `investigations` | No | None (queries are global) | Needs `user_id` column first |
| `agent_config` | N/A | Single-row system table | No RLS needed (service-only access) |

### Auth Architecture Consideration

Because the backend uses a **direct Postgres connection** (not Supabase's PostgREST/anon key), Supabase's `auth.uid()` function is not available in the session context. Two options exist for RLS enforcement:

- **Option A — `set_config` per request:** Before each query, run `SELECT set_config('app.current_user_id', '<uuid>', true)` to inject the user into the transaction, then write policies using `current_setting('app.current_user_id')`. This is the standard pattern for direct-connection apps with RLS.
- **Option B — Service role bypass:** Keep the backend connection as a privileged service role that bypasses RLS entirely, and rely on application-layer filtering only. RLS then only protects against direct DB access (Supabase dashboard, `psql`, PostgREST).

**Recommendation: Option A.** It provides true defence-in-depth — even a bug in application code cannot leak cross-user data. Option B would make the migration effort pointless for the backend.

---

## Phase 1 — Schema Gaps (migrations 011–012)

Fill in missing `user_id` columns so every user-scoped table can support RLS.

### Task 1.1 — Add `user_id` to `investigations`

**Migration `011_add_user_id_to_investigations.up.sql`:**

```sql
ALTER TABLE investigations
  ADD COLUMN user_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000'
  REFERENCES public.users(id) ON DELETE CASCADE;

-- Remove the default after backfill (no existing rows expected in prod yet)
ALTER TABLE investigations ALTER COLUMN user_id DROP DEFAULT;

CREATE INDEX idx_investigations_user_id ON investigations(user_id);
```

**Down migration** drops the column and index.

**Backend changes:**
- Update sqlc queries: `ListOpenInvestigations` → `ListOpenInvestigationsByUser`, add `user_id` param.
- Update `CreateInvestigation` to accept and insert `user_id`.
- Update handler/agent code that creates or lists investigations to pass `user_id` from context.

### Task 1.2 — Add `user_id` to `log_buffer`

**Migration `012_add_user_id_to_log_buffer.up.sql`:**

```sql
ALTER TABLE log_buffer
  ADD COLUMN user_id UUID REFERENCES public.users(id) ON DELETE CASCADE;

-- Backfill from the parent connection
UPDATE log_buffer lb
  SET user_id = c.user_id
  FROM connections c
  WHERE lb.connection_id = c.id;

ALTER TABLE log_buffer ALTER COLUMN user_id SET NOT NULL;

-- Replace the old connection-scoped index with a user-scoped one
CREATE INDEX idx_log_buffer_user_id_ingested ON log_buffer(user_id, ingested_at DESC);
```

**Backend changes:**
- Update the webhook ingestion handler to write `user_id` (looked up from the connection) when inserting log rows.
- Simplify `ListLogsByUserAndConnection` — can now filter directly on `user_id` instead of joining connections.

---

## Phase 2 — `set_config` Middleware (no migration)

Wire up the user identity so Postgres can see it during RLS evaluation.

### Task 2.1 — Add a DB helper to set the session user

In `backend/internal/db/`, add a function:

```go
func SetCurrentUser(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) error {
    _, err := pool.Exec(ctx,
        "SELECT set_config('app.current_user_id', $1, true)", userID.String())
    return err
}
```

The `true` argument scopes the setting to the **current transaction**.

### Task 2.2 — Call `set_config` in the auth middleware (or per-handler)

Two sub-options:

- **Per-transaction wrapper:** Create a helper that acquires a connection from the pool, calls `set_config`, runs the callback, and returns. All user-facing handlers use this wrapper.
- **Middleware approach:** Inject `set_config` in the auth middleware after extracting `user_id`. This is cleaner but requires passing a `*pgx.Conn` (not pool) through context so the config sticks for the duration of the request.

**Recommendation:** Per-transaction wrapper — simpler to reason about, no need to change the pool/context threading.

---

## Phase 3 — Enable RLS (migration 013)

A single migration that enables RLS and creates policies for all user-scoped tables.

### Task 3.1 — Migration `013_enable_rls.up.sql`

```sql
-- =====================================================
-- Row Level Security policies for all user-scoped tables
-- Uses app.current_user_id set via set_config() per transaction
-- =====================================================

-- Helper: safely read the session user (returns NULL if unset)
CREATE OR REPLACE FUNCTION app_current_user_id() RETURNS UUID AS $$
  SELECT nullif(current_setting('app.current_user_id', true), '')::UUID;
$$ LANGUAGE sql STABLE;

-- ── users ────────────────────────────────────────────
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

CREATE POLICY users_self ON users
  FOR ALL
  USING (id = app_current_user_id());

-- ── connections ──────────────────────────────────────
ALTER TABLE connections ENABLE ROW LEVEL SECURITY;

CREATE POLICY connections_owner ON connections
  FOR ALL
  USING (user_id = app_current_user_id());

-- ── conversations ────────────────────────────────────
ALTER TABLE conversations ENABLE ROW LEVEL SECURITY;

CREATE POLICY conversations_owner ON conversations
  FOR ALL
  USING (user_id = app_current_user_id());

-- ── agent_log ────────────────────────────────────────
ALTER TABLE agent_log ENABLE ROW LEVEL SECURITY;

CREATE POLICY agent_log_owner ON agent_log
  FOR ALL
  USING (user_id = app_current_user_id());

-- ── log_buffer ───────────────────────────────────────
ALTER TABLE log_buffer ENABLE ROW LEVEL SECURITY;

CREATE POLICY log_buffer_owner ON log_buffer
  FOR ALL
  USING (user_id = app_current_user_id());

-- ── investigations ───────────────────────────────────
ALTER TABLE investigations ENABLE ROW LEVEL SECURITY;

CREATE POLICY investigations_owner ON investigations
  FOR ALL
  USING (user_id = app_current_user_id());

-- ── agent_config ─────────────────────────────────────
-- No RLS — single-row system table, accessed only by the service role.
-- If the backend DB user is the table owner, RLS is bypassed automatically.
-- If not, grant explicit access:
--   GRANT ALL ON agent_config TO heimdall_service;
```

**Down migration** drops all policies, disables RLS on each table, and drops the helper function.

### Task 3.2 — Verify the backend DB role

RLS is **bypassed for table owners** and for roles with `BYPASSRLS`. The backend's `DATABASE_URL` must connect as a role that does **not** bypass RLS (otherwise the policies have no effect).

- Check the current role with `SELECT current_user, rolbypassrls FROM pg_roles WHERE rolname = current_user`.
- If the role bypasses RLS, either create a dedicated `heimdall_app` role without `BYPASSRLS`, or use `ALTER ROLE ... NOBYPASSRLS`.
- The Supabase `postgres` role is the superuser/owner — it bypasses RLS. The backend should use the `authenticated` role or a custom app role.

**This is a critical detail** — if skipped, all the policies will exist but do nothing.

---

## Phase 4 — Backend Integration & Testing

### Task 4.1 — Wrap all user-facing queries in the `set_config` transaction helper

Audit every handler in `backend/internal/api/handlers/` and ensure they use the transaction wrapper from Phase 2. Queries that run outside the wrapper will get zero rows back (because `app.current_user_id` will be unset and the policy will deny everything).

Key handlers to update:
- `connections.go` — CRUD + test
- `conversations.go` — list, get, create, update
- `logs.go` — list logs
- `investigations.go` — list, get (if handlers exist; currently agent-internal)
- `agent_log` queries in `logs.go`

### Task 4.2 — Agent/background process access

The agent loop and background processes (log ingestion, webhook handler) run **without a user HTTP request**. These need a different strategy:

- **Option A:** Run them as the table-owner role (bypasses RLS).
- **Option B:** Use `set_config` with the relevant user's ID before each agent operation.

**Recommendation:** Option B for the agent loop (it already knows which user it's acting on behalf of). Option A for system-level operations like webhook ingestion (which then must set `user_id` on the inserted row via application code, as it already does).

### Task 4.3 — Integration tests

Write tests that verify:
1. User A cannot read User B's connections, conversations, logs, or investigations.
2. Unset `app.current_user_id` returns zero rows (not an error).
3. The agent loop can still read/write when `set_config` is called with the correct user.
4. Webhook ingestion (service role) can still insert into `log_buffer`.

---

## Phase Summary

| Phase | Migrations | Backend Changes | Risk |
|-------|-----------|----------------|------|
| **1 — Schema Gaps** | `011`, `012` | Update queries + handlers for new `user_id` columns | Low — additive columns |
| **2 — set_config** | None | New DB helper + transaction wrapper | Medium — threading change |
| **3 — Enable RLS** | `013` | None (pure SQL) | High — wrong DB role = silent bypass; wrong policy = lockout |
| **4 — Integration** | None | Wrap all handlers, update agent loop | Medium — must not miss any query path |

**Recommended order:** Phases 1 → 2 → 4 (partial — wrap handlers) → 3 (enable RLS) → 4 (tests). This way, the `set_config` plumbing and handler changes are in place *before* RLS is turned on, avoiding a window where queries return empty results.
