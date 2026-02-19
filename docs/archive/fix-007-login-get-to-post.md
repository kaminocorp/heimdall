# Fix #7 — Login endpoint GET → POST

**Review ref:** scaffolding-review.md, Should Fix #7
**Date:** 2026-02-19

## Problem

The login route was registered as `r.Get("/api/auth/login", ...)`. Login accepts credentials and returns a token — this should be POST, not GET. Even as a stub, GET is misleading and would need changing later.

## Fix

Changed `r.Get` to `r.Post` for the `/api/auth/login` route.

No frontend changes needed — the login page currently calls `auth.login('placeholder-token')` directly without hitting the API.

## Files modified

| File | Line | Change |
|------|------|--------|
| `backend/internal/api/router.go` | 39 | `r.Get("/auth/login", ...)` → `r.Post("/auth/login", ...)` |

## Verification

- `go build ./...` — clean
- `go test ./...` — all pass
