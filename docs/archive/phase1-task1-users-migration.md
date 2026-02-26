# Phase 1 Task 1 — `public.users` Table + Sync Trigger

**Date:** 2026-02-22
**Reference:** [Phase 1 Auth Plan](../executing/phase1-auth.md) · Task 1

---

## What was done

Created and applied migration `006_create_users` against the Supabase Postgres instance.

### Files created / modified

| File | Purpose |
|------|---------|
| `backend/migrations/006_create_users.up.sql` | Creates `public.users` table, `handle_new_user()` trigger function, and `on_auth_user_created` trigger |
| `backend/migrations/006_create_users.down.sql` | Drops trigger, function, and table in correct order |
| `backend/internal/db/queries/users.sql` | sqlc query: `GetUser` (by UUID) |
| `backend/internal/db/users.sql.go` | **Generated** — `GetUser` query method |
| `backend/internal/db/models.go` | **Generated** — added `User` struct (id, email, created_at) |

### Schema

```sql
CREATE TABLE users (
    id         UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    email      TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

- FK to `auth.users(id)` with `ON DELETE CASCADE` — if a Supabase auth user is deleted, the public row is cleaned up.
- `UNIQUE` on `email` prevents duplicates.
- Intentionally minimal — only `id`, `email`, `created_at`. Additional fields (`display_name`, `avatar_url`) can be added later.

### Trigger

- **Function:** `public.handle_new_user()` — `SECURITY DEFINER` so it can write to `public.users` regardless of the inserting role.
- **Trigger:** `on_auth_user_created` — fires `AFTER INSERT ON auth.users`, copies `id` and `email` into `public.users`.
- This means any Supabase Auth sign-up automatically creates a corresponding application-level user row.

## How it was applied

```bash
source .env && migrate -path migrations -database "${DATABASE_URL}?sslmode=require" up 1
# Output: 6/u create_users (10.9324406s)
```

Migration version 6 is now recorded in the `schema_migrations` table on the Supabase instance.

### sqlc

Added `GetUser` query and ran `sqlc generate`, which updated:
- `internal/db/models.go` — new `User` struct at the end of the file.
- `internal/db/users.sql.go` — generated `GetUser` method on `*Queries`.

This pulls Task 4's sqlc work forward so the generated code is ready for the `/api/auth/me` handler.

## What's next

- Task 2: Backend JWT verification middleware (`internal/api/middleware/auth.go`)
- Task 4: `/api/auth/me` handler (sqlc query already done here)
