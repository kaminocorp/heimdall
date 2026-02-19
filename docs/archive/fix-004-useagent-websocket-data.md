# Fix #4 — `useAgent` composable never consumes WebSocket data

**Review ref:** scaffolding-review.md, Must Fix #4
**Date:** 2026-02-19

## Problem

`useAgent` destructured `data` from `useWebSocket` but never watched or consumed it. When the backend sent agent messages over the WebSocket, the `data` ref would update but nothing would parse the incoming JSON or push it into the `messages` array. Agent replies were silently dropped.

## Fix

Added a `watch(data, ...)` that:
1. Parses the incoming JSON string as a `ChatMessage`
2. Pushes it into the `messages` array with sensible defaults (`role: 'agent'`, current timestamp if missing)
3. Silently ignores malformed messages

## Files modified

| File | Change |
|------|--------|
| `frontend/src/composables/useAgent.ts` | Added `watch` import, added watcher on `data` ref |

## Verification

- `vue-tsc --noEmit` — zero errors
