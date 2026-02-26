# Phase 1 Task 2 — JWT Verification Middleware

**Date:** 2026-02-22
**Reference:** [Phase 1 Auth Plan](../executing/phase1-auth.md) · Task 2

---

## What was done

Replaced the stub auth middleware with real Supabase JWT validation.

### Files modified

| File | Change |
|------|--------|
| `backend/internal/config/config.go` | Added `SupabaseJWTSecret` field, loaded from `SUPABASE_JWT_SECRET` env var, required in `Validate()` |
| `backend/internal/api/middleware/auth.go` | Full rewrite — JWT parsing, HMAC-SHA256 validation, user ID extraction, context helper |
| `backend/go.mod` / `go.sum` | Added `github.com/golang-jwt/jwt/v5` dependency |

### Middleware behaviour

1. Extracts `Authorization: Bearer <token>` from the request header.
2. Parses the JWT using `golang-jwt/jwt/v5`, validating HMAC-SHA256 signature against the Supabase JWT secret.
3. Checks expiry (`exp` claim) — handled automatically by the library.
4. Extracts the user UUID from the `sub` claim.
5. Stores the user ID in the request context via `context.WithValue`.
6. Returns `401 Unauthorized` (JSON) on any validation failure.

### Exported API

- `Auth(jwtSecret string) func(http.Handler) http.Handler` — middleware constructor (takes the secret, returns chi-compatible middleware).
- `UserIDFromContext(ctx context.Context) (uuid.UUID, bool)` — context helper for handlers to retrieve the authenticated user's ID.

### Signature change

The old middleware was a plain `func(http.Handler) http.Handler` (no config). The new one is a **constructor** `Auth(secret) → middleware`, so the router will need updating (Task 3) to pass the JWT secret.

## What's next

- Task 3: Apply auth middleware to protected routes in `router.go`
- Task 4: `/api/auth/me` handler (uses `UserIDFromContext`)
