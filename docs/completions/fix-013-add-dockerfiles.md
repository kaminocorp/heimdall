# Fix 013 — Add missing Dockerfiles

**Issue:** #13 from scaffolding-review.md
**Date:** 2026-02-19

## Problem

`docker-compose.yml` references `backend/Dockerfile` and `frontend/Dockerfile`, but neither file existed. `docker compose up` would fail immediately.

## Fix

Created both Dockerfiles:

- **`backend/Dockerfile`** — Multi-stage build: compiles the Go binary in `golang:1.24-alpine`, copies it into a minimal `alpine:3.21` runtime image. Exposes port 8080.
- **`frontend/Dockerfile`** — Dev-oriented: `node:22-alpine`, installs deps via `npm ci`, runs Vite dev server with `--host` for container access. Exposes port 5173.

## Notes

The frontend Dockerfile is intentionally dev-mode (runs `vite dev`). A production variant with `npm run build` + nginx would be added when production builds are needed.
