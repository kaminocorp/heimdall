# Fix #2 — `http.Error` overwrites Content-Type

**Review ref:** scaffolding-review.md, Must Fix #2
**Date:** 2026-02-19

## Problem

Stub error responses set `Content-Type: application/json` then call `http.Error(...)`. Internally, `http.Error` calls `h.Set("Content-Type", "text/plain; charset=utf-8")`, overwriting the JSON header. Clients receive JSON-shaped bodies served as `text/plain`.

## Fix

Introduced a `jsonError` helper in a new `helpers.go` file that correctly sets the JSON content type, writes the status code, and encodes a `{"error": "..."}` JSON body. Replaced all `http.Error` calls in stub handlers with `jsonError`.

## Files modified

| File | Change |
|------|--------|
| `backend/internal/api/handlers/helpers.go` | **New** — `jsonError(w, message, status)` helper |
| `backend/internal/api/handlers/connections.go` | Replaced `http.Error` in `GetConnection`, `CreateConnection`, `UpdateConnection`, `DeleteConnection` |
| `backend/internal/api/handlers/agent.go` | Replaced `http.Error` in `UpdateAgentConfig` |
| `backend/internal/api/handlers/reports.go` | Replaced `http.Error` in `GetReport` |

## Not affected

- `handlers/auth.go` — no error responses
- `handlers/logs.go` — no error responses
- `handlers/chat.go` — WebSocket handler, no HTTP error responses

## Verification

- `go build ./...` — clean
- `go test ./...` — all pass
