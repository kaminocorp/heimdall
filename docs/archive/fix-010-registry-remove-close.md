# Fix 010 — Connector Registry `Remove` Doesn't Close

**Review item:** #10 (Should Fix)
**Status:** Done

## Problem

`Registry.Remove()` deleted the connector from the map but never called `Close()` on it, risking resource leaks (open DB connections, network sockets, goroutines).

## Changes

### `backend/internal/connectors/registry.go`
- `Remove` now returns `error`.
- Before deleting from the map, it retrieves the connector and calls `Close()` after deletion, returning any error.
- Returns `nil` if the id doesn't exist (no-op).

## Verification
- `go build ./...` passes clean.
- No existing callers of `Remove` — signature change is non-breaking.
