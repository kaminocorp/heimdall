# Fix 018 — Logging middleware doesn't capture status code

**Issue:** #18 from scaffolding-review.md
**Date:** 2026-02-19

## Problem

The logging middleware passed `http.ResponseWriter` through without wrapping, so logged requests had no status code.

## Fix

Added a `statusWriter` wrapper that captures the `WriteHeader` call (defaulting to 200). The logged `"status"` field now reflects the actual response status code.

## Verification

- `go build ./...` passes clean.
