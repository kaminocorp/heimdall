# Connections Handler Refactor — Phase 1

**Date:** 2026-04-12
**Assessment reference:** `docs/plans/code-assessment-2026-04-12.md`, issue #1

## Problem

`backend/internal/api/handlers/connections.go` was **693 lines** — the only backend file exceeding the 500-line threshold. The root cause was duplicated connection validation logic:

- **CreateConnection** (lines 141–198) and **UpdateConnection** (lines 330–383) contained identical validation blocks for 6 connector types (supabase, flyio, vercel, railway, mongodb, syslog), including identical syslog TLS cert injection logic.
- The `// --- Config validation (mirrors CreateConnection) ---` comment in UpdateConnection explicitly acknowledged the duplication.
- Any new connector type or validation change had to be applied in both places — a reliable source of future bugs.

Additionally, the file mixed CRUD handlers, a connectivity-testing handler (TestConnection, ~170 lines with its own connector switch), and validation helpers — three distinct concerns in one file.

## Solution

Split `connections.go` into three cohesive files:

### 1. `connections.go` — CRUD handlers (395 lines)

Contains `ListConnections`, `GetConnection`, `CreateConnection`, `UpdateConnection`, `DeleteConnection`. Each handler now calls `s.validateConnectorConfig()` in a single line instead of inlining 60+ lines of validation:

```go
// Before (60 lines of duplicated if/else per handler):
if req.Type == "supabase" { ... }
if req.Type == "flyio" { ... }
if req.Type == "vercel" { ... }
// ...plus syslog TLS injection...

// After (3 lines, both handlers):
config, err = s.validateConnectorConfig(req.Type, config)
if err != nil {
    jsonError(w, err.Error(), http.StatusBadRequest)
    return
}
```

### 2. `connections_validate.go` — validation logic (80 lines)

Contains:
- `validConnectionTypes` map and the three `isValid*` helper functions
- `validateConnectorConfig(connType, config) (json.RawMessage, error)` — a method on `*Server` (needs access to `s.Config` for syslog TLS certs). Validates by instantiating the appropriate connector constructor with sentinel UUIDs, and injects server-level TLS certs for syslog when the client config omits them. Returns the (possibly modified) config or a client-safe error.

### 3. `connections_test_handler.go` — TestConnection handler (190 lines)

The connectivity-testing handler is functionally distinct from CRUD — it instantiates connectors and performs real Connect/Close operations with timeouts. Moving it to its own file improves discoverability without altering behaviour.

## What changed

| File | Before | After | Delta |
|------|--------|-------|-------|
| `connections.go` | 693 lines | 395 lines | -298 |
| `connections_validate.go` | — | 80 lines | new |
| `connections_test_handler.go` | — | 190 lines | new |
| **Total** | 693 | 665 | -28 net (duplication removed) |

## What didn't change

- **No behavioural changes.** All HTTP responses, status codes, and error messages are identical.
- **No interface changes.** The `*Server` method set is unchanged — routes wire to the same methods.
- **TestConnection's switch** was left as-is. It looks structurally similar to validation, but it's doing a different job (testing real connectivity with timeouts vs. validating config shape). Extracting it further would be premature abstraction.
- **Webhook token logic** (generate in Create, preserve in Update) was not extracted — the two paths are legitimately different operations, not duplication.

## Verification

- `go build ./...` — passes
- `go vet ./...` — clean
- `go test ./internal/api/handlers/ -count=1 -short` — all tests pass
