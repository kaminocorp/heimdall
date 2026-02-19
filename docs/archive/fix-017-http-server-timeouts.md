# Fix 017 — Add HTTP server timeouts

**Issue:** #17 from scaffolding-review.md
**Date:** 2026-02-19

## Problem

`http.Server` in `main.go` had no `ReadTimeout`, `WriteTimeout`, or `IdleTimeout`. Without timeouts, slow or idle connections can exhaust server resources.

## Fix

Added timeouts to `http.Server`:
- `ReadTimeout: 30s`
- `WriteTimeout: 30s`
- `IdleTimeout: 120s`

## Verification

- `go build ./...` passes clean.
