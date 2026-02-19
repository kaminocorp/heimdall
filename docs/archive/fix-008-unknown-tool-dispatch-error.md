# Fix #8 — Unknown tool dispatch silently succeeds

**Review ref:** scaffolding-review.md, Should Fix #8
**Date:** 2026-02-19

## Problem

`Dispatch` returned `"", nil` for unknown tool names. This silently swallowed dispatch errors — if the LLM hallucinated a tool name, the agent loop would receive an empty result with no error, making it impossible to diagnose.

## Fix

Changed the default case to return `fmt.Errorf("unknown tool: %s", name)`, surfacing the bad tool name in the error.

## Files modified

| File | Line | Change |
|------|------|--------|
| `backend/internal/agent/tools.go` | 1–2 | Added `import "fmt"` |
| `backend/internal/agent/tools.go` | 74 | `return "", nil` → `return "", fmt.Errorf("unknown tool: %s", name)` |

## Verification

- `go build ./...` — clean
- `go test ./...` — all pass
