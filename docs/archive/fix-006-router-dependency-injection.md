# Fix #6 — Router dependency injection

**Review ref:** scaffolding-review.md, Should Fix #6
**Date:** 2026-02-19

## Problem

`NewRouter` accepted `*config.Config` but never used it. All handlers were standalone functions with no access to shared dependencies (DB pool, agent instance, config). As handlers get real implementations, they will all need these dependencies.

## Fix

Introduced a `Server` struct in the `handlers` package that holds shared dependencies. Converted all 12 handler functions to methods on `*Server`. The router creates a `Server` via `handlers.NewServer(cfg)` and wires method references.

The struct currently holds only `Config` — DB pool, agent instance, and other dependencies will be added as those features are implemented.

## Files modified

| File | Change |
|------|--------|
| `backend/internal/api/handlers/server.go` | **New** — `Server` struct with `Config` field + `NewServer` constructor |
| `backend/internal/api/handlers/connections.go` | 5 functions → methods on `*Server` |
| `backend/internal/api/handlers/agent.go` | 2 functions → methods on `*Server` |
| `backend/internal/api/handlers/reports.go` | 2 functions → methods on `*Server` |
| `backend/internal/api/handlers/logs.go` | 1 function → method on `*Server` |
| `backend/internal/api/handlers/auth.go` | 1 function → method on `*Server` |
| `backend/internal/api/handlers/chat.go` | 1 function → method on `*Server` |
| `backend/internal/api/router.go` | Creates `handlers.NewServer(cfg)`, routes wired to `s.Method` |

## Not modified

- `backend/internal/api/handlers/helpers.go` — `jsonError` stays as a package-level helper (no dependency access needed)
- `backend/cmd/heimdall/main.go` — no changes needed, still calls `api.NewRouter(cfg)`

## Verification

- `go build ./...` — clean
- `go test ./...` — all pass
