# Fix #3 — Incomplete tool registry

**Review ref:** scaffolding-review.md, Must Fix #3
**Date:** 2026-02-19

## Problem

`ToolRegistry()` only declared 2 of 5 tools (`query_database`, `search_logs`). The other 3 — `search_codebase`, `recall_similar_incidents`, `recall_lessons` — existed in the `Dispatch` switch and had stub implementations, but were never declared in the registry. The LLM would never call undeclared tools.

## Fix

Added the 3 missing tool definitions to `ToolRegistry()` with appropriate descriptions and parameters, matching the stub implementations in `tools_codebase.go` and `tools_memory.go`.

## Files modified

| File | Change |
|------|--------|
| `backend/internal/agent/tools.go` | Added `search_codebase`, `recall_similar_incidents`, `recall_lessons` to `ToolRegistry()` |

## Verification

- `go build ./...` — clean
- `go test ./...` — all pass
