# GitHub Connection Flow Fixes — Phase 1 Completion

**Status:** Complete (code changes landed; `FRONTEND_URL` must be set in `.env` and Fly.io)
**Date:** 2026-04-10
**Plan:** [github-connection-fixes.md](../executing/github-connection-fixes.md), Phase 1
**Validation:** Manual review (Go 1.25 toolchain not available locally for `go build`)

---

## Goal

After the user installs the GitHub App and GitHub redirects to `/api/github/callback`, the browser should land back on the Heimdall frontend connections page — not a blank backend route.

---

## Root cause

`github.go:240` redirected to a bare relative path:

```go
http.Redirect(w, r, "/connections?github=installed", http.StatusFound)
```

The callback URL registered with GitHub is `https://heimdall-backend.fly.dev/api/github/callback` — the **backend** origin. A relative redirect from there resolves to `https://heimdall-backend.fly.dev/connections`, which is the backend server. The backend has no route for `/connections`, so the user sees nothing.

The frontend runs on a separate origin (its own nginx container), so the redirect must use an absolute URL pointing to the frontend.

---

## Changes

### `backend/internal/config/config.go`

Added `FrontendURL` field to the `Config` struct, loaded from `FRONTEND_URL` env var with a default of `http://localhost:5173` (matches the Vite dev server).

### `backend/internal/api/handlers/github.go:240`

Changed the redirect from a bare path to a fully qualified URL:

```go
http.Redirect(w, r, s.Config.FrontendURL+"/connections?github=installed", http.StatusFound)
```

---

## Deployment action required

`FRONTEND_URL` must be set in two places before this fix is live:

1. **Local `.env`** — already correct by default (`http://localhost:5173`)
2. **Fly.io production** — set to the frontend's production origin, e.g.:
   ```
   fly secrets set FRONTEND_URL=https://<frontend-domain> --app heimdall-backend
   ```

Without this secret, the default `http://localhost:5173` will be used in production, which will break the redirect.

---

## Why `FRONTEND_URL` rather than inferring from the request

Alternatives considered:

- **`Referer` / `Origin` header**: Not reliable — the callback comes from GitHub's redirect, not from the Heimdall frontend. These headers may be absent or point to `github.com`.
- **Hardcoded production URL**: Fragile, breaks in staging/local dev.
- **Derive from `CORS_ALLOWED_ORIGINS`**: Overloads the meaning of that config and picks arbitrarily if multiple origins are allowed.

An explicit `FRONTEND_URL` is the simplest correct approach, consistent with how the CORS config already separates backend/frontend concerns.
