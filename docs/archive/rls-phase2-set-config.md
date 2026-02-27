# RLS Phase 2 — `set_config` Per-Request User Context (Completed)

**Date:** 2026-02-26
**Scope:** Wire up `SET LOCAL app.current_user_id` on every user-scoped database request so Postgres knows which user is acting. This is the plumbing that Phase 3's RLS policies will rely on.

---

## Why

Phase 1 gave every user-scoped table a `user_id` column. Phase 3 will enable RLS policies that check `current_setting('app.current_user_id')`. But the backend connects to Postgres via a shared `pgxpool.Pool` with no per-request identity — Supabase's `auth.uid()` isn't available because we use direct `pgx` connections, not PostgREST. Without this phase, RLS policies would see no user identity and deny all rows.

---

## What Changed

### New: `UserQueries` helper

**`backend/internal/api/handlers/userqueries.go`** (new file)

```go
func (s *Server) UserQueries(ctx, userID) (*db.Queries, func(), error)
```

1. Begins a transaction on the pool
2. Runs `SET LOCAL app.current_user_id = $1` (scoped to the transaction)
3. Returns `*db.Queries` wrapping that transaction + a `done()` cleanup function
4. Caller defers `done()` which commits the transaction and releases the connection

Uses sqlc's existing `Queries.WithTx(tx)` method — no changes to generated code.

### Server struct

**`backend/internal/api/handlers/server.go`** — added `Pool *pgxpool.Pool` field alongside the existing `Queries`. The pool was already passed into `NewServer` but wasn't stored.

### Updated handlers

Every handler that extracts a `userID` from the JWT context now calls `s.UserQueries(ctx, userID)` instead of using `s.Queries` directly.

| File | Handlers updated |
|------|-----------------|
| `connections.go` | `ListConnections`, `GetConnection`, `CreateConnection`, `UpdateConnection`, `TestConnection`, `DeleteConnection` |
| `conversations.go` | `ListConversations`, `GetConversation` |
| `chat.go` | `HandleChat` (conversation load/create), `persistMessages`, title update |
| `logs.go` | `ListLogs` |
| `auth.go` | `Me` |

### Chat handler (WebSocket) — special treatment

The WebSocket handler is long-lived, so a single transaction for the entire connection isn't practical. Instead, each discrete DB operation gets its own `UserQueries` call:
- Conversation load/create at connection start
- `persistMessages` on each message save
- Title update on first message

### NOT updated (intentionally)

| File | Reason |
|------|--------|
| `agent.go` | System-wide agent config — no `user_id` scoping |
| `webhooks.go` | Auth via webhook token, not JWT — no user context in middleware |
| `reports.go` | Stub handlers, no DB calls |

The agent's own `*db.Queries` (in `internal/agent/`) is unchanged — it will be addressed in Phase 4 when RLS is actually enabled.

---

## Verification

- `go build ./...` — compiles cleanly.
- Grep for `s.Queries.` in handler files — only `userqueries.go` (the helper itself), `agent.go`, and `webhooks.go` remain, all intentionally non-user-scoped.
