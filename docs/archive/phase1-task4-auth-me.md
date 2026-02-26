# Phase 1 Task 4 — `/api/auth/me` Endpoint

**Date:** 2026-02-22
**Reference:** [Phase 1 Auth Plan](../executing/phase1-auth.md) · Task 4

---

## What was done

Added the `GET /api/auth/me` endpoint and wired the database pool into the server.

### Files created / modified

| File | Change |
|------|--------|
| `backend/cmd/heimdall/main.go` | Creates `pgxpool.Pool` from `DATABASE_URL` on startup, passes to `NewRouter` |
| `backend/internal/api/router.go` | Accepts `*pgxpool.Pool`, passes to `NewServer`; added `GET /api/auth/me` route |
| `backend/internal/api/handlers/server.go` | Added `*db.Queries` field to `Server`; `NewServer` now takes pool and wraps with `db.New(pool)` |
| `backend/internal/api/handlers/auth.go` | New `Me` handler — extracts user ID from JWT context, queries `public.users`, returns JSON |
| `backend/go.mod` / `go.sum` | Promoted `golang-jwt`, `google/uuid`, `pgx/v5` to direct; added `pgxpool` transitive deps |

### Handler behaviour

1. Extracts user UUID from request context (set by auth middleware in Task 2).
2. Queries `public.users` via the `GetUser` sqlc method (generated in Task 1).
3. Returns `200` with `{ "id": "...", "email": "...", "created_at": "..." }`.
4. Returns `401` if the context has no user ID (shouldn't happen behind auth middleware, but defensive).
5. Returns `404` if the user exists in `auth.users` but not yet in `public.users`.

### Database pool

This is the first real database connection in the server. Prior to this change, `Server` held only `*config.Config` — no DB access. The pool is created in `main.go` using `pgxpool.New` and closed on shutdown via `defer pool.Close()`.

## What's next

- Task 5: Frontend Supabase client setup (`@supabase/supabase-js`)
- Task 6: Rewrite auth store to use Supabase Auth
