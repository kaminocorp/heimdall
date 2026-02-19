# Fix 026 — Frontend Dockerfile production build & Postgres healthcheck

## Problems

### 1. Frontend Dockerfile ran Vite dev server
`frontend/Dockerfile` used `npm run dev` as its entrypoint. Vite's dev server is not
suitable for production — it's unoptimised, serves unbundled modules, and exposes HMR
internals.

### 2. Docker Compose missing Postgres healthcheck
`docker-compose.yml` declared `depends_on: postgres` for the backend service, but
`depends_on` only waits for the container to **start**, not for Postgres to be **ready**.
On slower machines or cold starts the backend would attempt to connect before Postgres
was accepting connections, causing a crash.

## Fixes

### Frontend Dockerfile → multi-stage production build
Replaced the single-stage dev image with a two-stage build:

1. **Build stage** (`node:22-alpine`) — installs deps, runs `npm run build` to produce
   static assets in `dist/`.
2. **Serve stage** (`nginx:alpine`) — copies the built assets into nginx's html root and
   serves them on port 80.

### Postgres healthcheck + conditional depends_on
- Added a `healthcheck` to the `postgres` service using `pg_isready -U heimdall`
  (5 s interval, 3 s timeout, 5 retries).
- Changed backend's `depends_on` from a bare list to the map form with
  `condition: service_healthy`, so the backend only starts once Postgres is confirmed
  ready.

## Files changed

| File | Change |
|------|--------|
| `frontend/Dockerfile` | Replaced dev-server image with multi-stage nginx build |
| `docker-compose.yml` | Added Postgres healthcheck; backend depends on `service_healthy` |

## Not addressed

**`go.mod` indirect deps (`pgx/v5`, `uuid`)** — these are indirect because the
sqlc-generated `internal/db` package isn't imported by the main binary yet. They will
automatically promote to direct dependencies when the DB layer is wired into handlers.
Running `go mod tidy` at that point is sufficient; forcing them direct now would be
reverted by the next tidy.
