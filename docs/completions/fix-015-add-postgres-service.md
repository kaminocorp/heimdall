# Fix 015 — Add Postgres service to docker-compose.yml

**Issue:** #15 from scaffolding-review.md
**Date:** 2026-02-19

## Problem

The blueprint describes docker-compose as providing "backend + frontend + postgres", but the actual file had no Postgres service. The backend requires a database to function.

## Fix

Added a `postgres` service using `postgres:17-alpine` with:
- Default dev credentials (`heimdall`/`heimdall`/`heimdall`)
- Port 5432 exposed to host
- Named volume `pgdata` for data persistence across restarts
- `depends_on: postgres` added to the backend service
