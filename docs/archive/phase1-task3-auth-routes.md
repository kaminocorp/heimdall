# Phase 1 Task 3 — Apply Auth Middleware to Protected Routes

**Date:** 2026-02-22
**Reference:** [Phase 1 Auth Plan](../executing/phase1-auth.md) · Task 3

---

## What was done

Wired the JWT auth middleware (Task 2) into the router so all `/api/*` routes require a valid Supabase token. Removed the placeholder login endpoint.

### Files modified

| File | Change |
|------|--------|
| `backend/internal/api/router.go` | Added `middleware.Auth(cfg.SupabaseJWTSecret)` to the `/api` route group; removed `POST /api/auth/login` |
| `backend/internal/api/handlers/auth.go` | Removed dead `Login` handler (Supabase handles login directly). File kept for Task 4's `Me` handler |

### Route structure after this change

```
Global middleware: Logging, CORS

/api/*              — protected (JWT required)
  /connections/*    — CRUD
  /agent/config     — GET / PUT
  /logs             — GET
  /reports/*        — GET list / detail

/ws/chat            — unprotected (WebSocket auth deferred to Phase 5)
```

### Design decisions

- **Auth scoped to `/api` group, not global.** Applied via `r.Use()` inside the `/api` `r.Route()` block. This keeps `/ws/chat` and any future public endpoints outside the JWT gate without needing explicit exemptions.
- **No public `/api` routes for now.** Supabase Auth handles signup/login on the client side. The only auth-related API endpoint will be `GET /api/auth/me` (Task 4), which is protected.
- **`Login` handler removed, not repurposed.** It was a stub returning `{"token":"placeholder"}`. Since Supabase issues tokens, there's nothing to repurpose. Task 4 will add a `Me` handler in the same file.

## What's next

- Task 4: `GET /api/auth/me` endpoint (uses `UserIDFromContext` from the middleware + `GetUser` sqlc query)
