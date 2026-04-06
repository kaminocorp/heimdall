# Phase 4 Completion — Ingestion Hardening

**Date:** 2026-04-06
**Roadmap ref:** Post-implementation review of Phases 1–3
**Status:** Complete

---

## What was done

Code review of the full Phase 1–3 ingestion implementation surfaced 5 issues. All fixed in a single pass.

---

## Fixes

### 1. Python `__pycache__` files tracked in git

`.pyc` bytecode files from the Python SDK were committed to the repo. Added `**/__pycache__/` and `*.pyc` to `.gitignore` and removed the 4 tracked files from the index via `git rm --cached`.

| File | Change |
|------|--------|
| `.gitignore` | Added `**/__pycache__/` and `*.pyc` entries |

### 2. GraphQL injection in Railway `Connect()`

`railway.go:76` concatenated `ProjectID` directly into the GraphQL query string. A malformed `project_id` value could break out of the string literal. The `Poll` method already used parameterized variables correctly — `Connect` now does the same.

| Before | After |
|--------|-------|
| `` query { project(id: "` + r.config.ProjectID + `") { id name } } `` | `query($id: String!) { project(id: $id) { id name } }` with variables map |

| File | Change |
|------|--------|
| `backend/internal/connectors/logs/railway.go` | `Connect()` uses parameterized `$id` variable instead of string interpolation |

### 3. Dead code removal — `parseOTLPTimestamp`

`parseOTLPTimestamp` in `otlp.go` was defined but never called. Removed the function and the now-unused `strconv` import.

| File | Change |
|------|--------|
| `backend/internal/api/handlers/otlp.go` | Removed `parseOTLPTimestamp`, removed `strconv` import |

### 4. MongoDB missing `ClusterName` validation

`NewMongoDB` validated `PublicKey`, `PrivateKey`, and `GroupID` as required but not `ClusterName`, despite using it for `source_type` (`"mongodb/<cluster>"`) and hostname discovery. Added the required check.

| File | Change |
|------|--------|
| `backend/internal/connectors/logs/mongodb.go` | Added `cluster_name is required` validation in `NewMongoDB` |

### 5. `resolveAnyValue` `BoolValue` documentation

`resolveAnyValue` in the OTLP handler checks `if v.BoolValue`, which only matches `true`. A `boolValue: false` field is indistinguishable from an unset field due to Go's zero-value semantics. Fixing this properly would require `*bool`, which would complicate the OTLP type definitions for a rare edge case. Added a comment documenting the trade-off.

| File | Change |
|------|--------|
| `backend/internal/api/handlers/otlp.go` | Added explanatory comment on `BoolValue` limitation |

---

## Phase ordering review

Confirmed that Phases 1–3 were implemented in the correct order with no do/undo conflicts:

| Phase | Files touched | Overlap with prior phases |
|-------|--------------|--------------------------|
| 1 (Syslog) | `listener.go`, `syslog.go`, `connections.go`, `main.go`, `server.go`, `router.go`, `config.go` | None (new files + additive changes) |
| 2 (OTLP + JS SDK) | `otlp.go`, `connections.go`, `router.go`, `flows.ts` | `connections.go` — different section (OTLP token gen) |
| 3 (Parsers + Pollers + SDKs) | `webhook_parsers.go`, `webhooks.go`, 4 poller files, `connections.go`, `main.go` | `connections.go` — different section (poller validation). `main.go` — `resumePollers` refactored from Supabase-only to generic loop (additive, no revert) |

Each phase built on the last without reverting prior work.

---

## Test results

All 56 tests from Phases 1–3 continue to pass after these fixes.

| Component | Tests | Status |
|-----------|-------|--------|
| Syslog connector | 9 | Pass |
| OTLP handler | 7 | Pass |
| Webhook parsers | 12 | Pass |
| JS SDK | 10 | Pass |
| Python SDK | 9 | Pass |
| Go SDK | 9 | Pass |
