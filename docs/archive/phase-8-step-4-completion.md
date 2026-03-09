# Phase 8, Step 4 — Lumber Integration: Completion Notes

Reference: [Phase 8 Step 4 Implementation Plan](../executing/phase-8-step-4-lumber-integration.md)

---

## Overview

Added deterministic log classification to Heimdall using [Lumber](https://github.com/kaminocorp/lumber) as an in-process Go library. Lumber classifies each log entry into one of 42 taxonomy labels (e.g. `ERROR.connection_failure`, `REQUEST.success`) and a severity gate decides which logs get escalated to the LLM agent.

**Key principle:** Lumber is a **router**, not a filter. All logs remain stored in `log_buffer`. Classification only determines which logs warrant LLM attention. A well-behaved production app in steady state generates zero LLM calls.

---

## Task 1 — Config: Classifier Mode & Model Directory

| Action | File |
|--------|------|
| Edited | `backend/internal/config/config.go` |

Added two fields to `Config` struct:

| Field | Env Var | Default | Description |
|-------|---------|---------|-------------|
| `ClassifierMode` | `CLASSIFIER_MODE` | `fallback` | `on` = required, `off` = skip classification, `fallback` = use if available |
| `LumberModelDir` | `LUMBER_MODEL_DIR` | `/opt/lumber/models` | Directory containing ONNX model files |

No validation added — both have sensible defaults. The three-way mode logic is handled at init time in `main.go` (Task 7).

---

## Task 2 — Classifier Interface & Types

| Action | File |
|--------|------|
| Created | `backend/internal/agent/classifier.go` |

Defines the `Classifier` interface consumed by the monitor loop:

```go
type Classifier interface {
    Classify(logs []db.LogBuffer) (flagged []ClassifiedLog, safeCount int)
    Close() error
}
```

`ClassifiedLog` struct pairs a `db.LogBuffer` with classification metadata: `Type`, `Category`, `Severity`, `Confidence`, `Summary`.

**Why an interface?** Enables mock injection in monitor loop tests (Step 5) without ONNX. Also cleanly handles the two concrete implementations: `LumberClassifier` (real ONNX) and `PassthroughClassifier` (escalate all).

---

## Task 3 — Text Extraction from LogBuffer Payloads

| Action | File |
|--------|------|
| Created | `backend/internal/agent/extract.go` |

### `ExtractText(entry db.LogBuffer) string`

Pulls classifiable text from arbitrary JSON payloads. Strategy:

1. **Priority-ordered field scan** — checks `message` → `msg` → `error` → `text` → `log` → `body`
2. **Level prepend** — if `level`, `severity`, or `error_severity` field exists, prepends it uppercased (e.g. `"ERROR: connection refused"`)
3. **Fallback** — returns raw JSON string if no message field found, or raw string if payload isn't JSON

**Why prepend the level?** A log like `"connection refused"` is ambiguous. `"ERROR: connection refused"` gives the classifier a strong signal. The level comes from the webhook payload's own severity metadata.

### `ExtractTexts(logs []db.LogBuffer) []string`

Batch wrapper — returns texts in same order as input slice.

---

## Task 4 — Severity Gate

| Action | File |
|--------|------|
| Created | `backend/internal/agent/severity_gate.go` |

### `ShouldEscalate(event lumber.Event) bool`

Pure function — hardcoded rules per the Phase 8 severity gate table:

| Type | Escalated Categories | Safe Categories |
|------|---------------------|-----------------|
| ERROR | All | — |
| PERFORMANCE | All | — |
| REQUEST | `server_error`, `slow_request` | `success`, `redirect`, `client_error` |
| DEPLOY | All | — |
| SYSTEM | `resource_alert`, `config_change` | `health_check`, `process_lifecycle`, `scaling_event` |
| ACCESS | `login_failure`, `auth_failure`, `permission_change`, `api_key_event` | `login_success`, `session_expired` |
| DATA | `migration` | `query_executed`, `replication` |
| SCHEDULED | `cron_failed` | `cron_started`, `cron_completed` |
| UNCLASSIFIED | Always | — |
| Unknown type | Always (fail-safe) | — |

**Separation of concerns:** The confidence backstop (`< 0.5 → escalate`) is handled in `LumberClassifier.Classify()`, not here. This function is pure domain logic; confidence is a classifier-quality concern.

---

## Task 5 — LumberClassifier Implementation

| Action | File |
|--------|------|
| Created | `backend/internal/agent/classifier_lumber.go` |

### `NewLumberClassifier(modelDir string) (*LumberClassifier, error)`

Wraps `lumber.New()` with:
- `WithModelDir(modelDir)` — model files location
- `WithConfidenceThreshold(0.5)` — below this, Lumber marks events as UNCLASSIFIED
- `WithVerbosity("minimal")` — short summaries to reduce LLM token usage

### `Classify(logs []db.LogBuffer) ([]ClassifiedLog, int)`

Pipeline:
1. `ExtractTexts(logs)` — pull message text from each payload
2. `engine.ClassifyBatch(texts)` — single batched ONNX forward pass (~0.5ms/line)
3. Apply severity gate + confidence backstop per event
4. Return `flagged` (sent to LLM) and `safeCount` (logged, never sent)

**Fail-open on error:** If `ClassifyBatch` returns an error, all logs in the batch are escalated as `UNCLASSIFIED` with severity `warning`. Never silently drops logs.

**Dual escalation triggers:** A log escalates if `event.Confidence < 0.5` OR `ShouldEscalate(event)` returns true.

---

## Task 6 — PassthroughClassifier

| Action | File |
|--------|------|
| Created (same file as Task 2) | `backend/internal/agent/classifier.go` |

Escalates all logs without classification. Used when `CLASSIFIER_MODE=off` or when Lumber fails to initialize in `fallback` mode. All logs get `Type: "UNCLASSIFIED"`, `Severity: "unknown"`, and `Summary` from `ExtractText`.

---

## Task 7 — Wire Classifier into Agent

| Action | File |
|--------|------|
| Edited | `backend/internal/agent/agent.go` |
| Edited | `backend/cmd/heimdall/main.go` |
| Edited | `backend/internal/agent/loop_test.go` |
| Edited | `backend/internal/agent/tools_test.go` |

### Agent struct

Added `classifier Classifier` field. Constructor updated to `New(queries, cfg, classifier)`.

### main.go initialization

Three-way switch on `cfg.ClassifierMode`:

| Mode | Behavior |
|------|----------|
| `"on"` | Init `LumberClassifier` — fatal exit if it fails |
| `"off"` | Use `PassthroughClassifier` — all logs escalated |
| `"fallback"` (default) | Try `LumberClassifier` — fall back to `PassthroughClassifier` on error, log warning |

### Shutdown

`classifier.Close()` called after HTTP server shutdown, before `pool.Close()`. Ensures no in-flight requests try to classify after ONNX runtime is released.

### Test updates

Both `loop_test.go` and `tools_test.go` updated to include `classifier: &PassthroughClassifier{}` in the `Agent` struct literal. Prevents nil pointer panics if any test path touches the classifier.

---

## Task 8 — Dockerfile: ONNX Runtime & Model Files

| Action | File |
|--------|------|
| Edited | `backend/Dockerfile` |

### Changes from original

| Before | After | Why |
|--------|-------|-----|
| `golang:1.24-alpine` build stage | `golang:1.24` (Debian) | Consistency with runtime stage |
| `alpine:3.21` runtime | `debian:bookworm-slim` | ONNX Runtime requires glibc (Alpine uses musl) |
| Single-stage | Three-stage: `build`, `models`, `runtime` | Model downloads cached independently of code changes |

### Three-stage build

1. **`build`** — compiles Go binary
2. **`models`** — downloads model files from HuggingFace + ONNX Runtime 1.24.1 shared library from GitHub releases
3. **`runtime`** — Debian slim with `ca-certificates`, binary, and model files

### Deviation from plan

The plan placed `libonnxruntime.so` in `/usr/local/lib/` with `ldconfig`. Reading Lumber's source (`embedder/onnx.go:41`) revealed it looks for the library **in the same directory as the model files** via `filepath.Join(modelDir, "libonnxruntime.so")`. So the library goes to `/opt/lumber/models/` alongside the model files — no `ldconfig` needed.

### Volume mount safety

`docker-compose.yml` mounts `./backend:/app`. Model files are placed at `/opt/lumber/models` (outside the mount) to avoid being shadowed.

---

## Task 9 — Unit & Integration Tests

| Action | File | Test Count |
|--------|------|------------|
| Created | `backend/internal/agent/extract_test.go` | 12 tests |
| Created | `backend/internal/agent/severity_gate_test.go` | 32 tests |
| Created | `backend/internal/agent/classifier_test.go` | 3 + 4 integration tests |

### Extract tests (`extract_test.go`)

| Test | Payload | Expected |
|------|---------|----------|
| Standard `message` | `{"level":"error","message":"connection refused"}` | `ERROR: connection refused` |
| Shorthand `msg` | `{"msg":"request completed","status":200}` | `request completed` |
| `error` field | `{"error":"timeout after 30s"}` | `timeout after 30s` |
| Level prepended | `{"level":"warn","message":"high latency"}` | `WARN: high latency` |
| Supabase-style | `{"error_severity":"ERROR","msg":"deadlock detected"}` | `ERROR: deadlock detected` |
| No message field | `{"path":"/api/users","status":500}` | raw JSON string |
| Non-JSON payload | `plain text log line` | `plain text log line` |
| Empty object | `{}` | `{}` |
| Priority ordering | `{"message":"primary","error":"secondary"}` | `primary` |
| Severity field | `{"severity":"info","message":"startup complete"}` | `INFO: startup complete` |
| Empty message skipped | `{"message":"","error":"real error"}` | `real error` |
| Batch extraction | 3 mixed entries | correct order preserved |

### Severity gate tests (`severity_gate_test.go`)

Tests every branch: 20 escalation cases + 12 safe cases covering all 8 taxonomy root types plus UNCLASSIFIED and unknown types.

### Classifier tests (`classifier_test.go`)

**Pure Go (always run):**
- `TestPassthroughClassifier` — all logs flagged, empty batch, close is no-op

**Integration (gated by `LUMBER_MODEL_DIR`):**
- Empty batch → `nil, 0`
- Known error log → flagged, `type=ERROR`
- Known healthy request → safe (`REQUEST.success`)
- Mixed batch (5 safe + 2 errors) → uses `GreaterOrEqual` assertions (ML isn't exact at boundaries)

---

## Verification

- `go mod tidy` — clean, Lumber promoted from indirect to direct dependency
- `go build ./...` — clean
- `go test ./internal/agent/ -v` — 53 tests passing (47 new + 6 existing)
- No regressions in existing `loop_test.go` or `tools_test.go`

---

## Files Summary

| Action | File | Task |
|--------|------|------|
| Edited | `backend/internal/config/config.go` | 1 |
| Created | `backend/internal/agent/classifier.go` | 2, 6 |
| Created | `backend/internal/agent/extract.go` | 3 |
| Created | `backend/internal/agent/severity_gate.go` | 4 |
| Created | `backend/internal/agent/classifier_lumber.go` | 5 |
| Edited | `backend/internal/agent/agent.go` | 7 |
| Edited | `backend/cmd/heimdall/main.go` | 7 |
| Edited | `backend/internal/agent/loop_test.go` | 7 |
| Edited | `backend/internal/agent/tools_test.go` | 7 |
| Edited | `backend/Dockerfile` | 8 |
| Edited | `backend/go.mod` | tidy |
| Edited | `backend/go.sum` | tidy |
| Created | `backend/internal/agent/extract_test.go` | 9 |
| Created | `backend/internal/agent/severity_gate_test.go` | 9 |
| Created | `backend/internal/agent/classifier_test.go` | 9 |
