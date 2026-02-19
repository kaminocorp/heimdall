# Fix #1 — `go.mod` all deps marked `// indirect`

**Review ref:** scaffolding-review.md, Must Fix #1
**Date:** 2026-02-19

## Problem

Every dependency in `backend/go.mod` was tagged `// indirect`, including packages directly imported by the codebase (e.g. `chi/v5`, `coder/websocket`). This is incorrect metadata — `go mod tidy` would reclassify or remove them.

## Fix

Ran `go mod tidy` in `backend/`.

## Result

Most dependencies were removed entirely because the stub implementations don't yet import them (pgx, anthropic-sdk-go, google/uuid, golang-migrate, etc.). Two remain as direct dependencies:

- `github.com/coder/websocket v1.8.14`
- `github.com/go-chi/chi/v5 v5.2.5`

`go.sum` was also cleaned up accordingly.

## Files modified

| File | Change |
|------|--------|
| `backend/go.mod` | Reclassified deps; removed 11 unused entries |
| `backend/go.sum` | Reduced to match actual dependency graph |

## Verification

- `go build ./...` — clean
- `go test ./...` — all pass

## Note

Removed dependencies (pgx, anthropic-sdk-go, uuid, golang-migrate, etc.) will be re-added as direct deps naturally when their stub implementations are filled in with real import statements.
