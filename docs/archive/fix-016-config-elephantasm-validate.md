# Fix 016 — Config missing Elephantasm fields and validation

**Issue:** #16 from scaffolding-review.md
**Date:** 2026-02-19

## Problem

`Config` struct had no `ElephantasmURL` or `ElephantasmKey` fields despite `internal/memory/client.go` needing both. No `Validate()` method existed to catch missing required config at startup.

## Fix

- Added `ElephantasmURL` and `ElephantasmKey` fields to `Config`, loaded from `ELEPHANTASM_URL` and `ELEPHANTASM_API_KEY` env vars (empty defaults — Elephantasm is optional).
- Added `Validate() error` method that requires `DATABASE_URL` and `ANTHROPIC_API_KEY`.
- Called `cfg.Validate()` in `main.go` — exits immediately with a clear error if required config is missing.

## Verification

- `go build ./...` passes clean.
