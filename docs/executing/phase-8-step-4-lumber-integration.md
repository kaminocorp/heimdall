# Phase 8, Step 4 — Lumber Integration: Implementation Plan

Reference: [Phase 8 Master Plan](./phase-8-monitoring-mode.md) | [Lumber Integration Guide](./lumber-integration-guide.md)

---

## Goal

Add deterministic log classification to Heimdall using [Lumber](https://github.com/kaminocorp/lumber) as an in-process Go library. Lumber classifies each log entry into one of 42 taxonomy labels (e.g. `ERROR.connection_failure`, `REQUEST.success`) and the severity gate decides which logs get escalated to the LLM agent.

**Key principle:** Lumber is a **router**, not a filter. All logs remain stored in `log_buffer`. Lumber determines which logs warrant LLM attention and which are routine (BAU). A well-behaved production app in steady state generates zero LLM calls.

---

## Current State

| Component | Status | File |
|-----------|--------|------|
| Lumber Go module | In `go.mod` as indirect dep | `backend/go.mod:20` |
| `Agent` struct | No classifier field | `backend/internal/agent/agent.go:14-18` |
| `LogBuffer` model | Has `Payload json.RawMessage` | `backend/internal/db/models.go:95-103` |
| `ListLogsSinceForApp` | Returns `[]LogBuffer` | `backend/internal/db/monitoring.sql.go:86` |
| Config struct | No classifier fields | `backend/internal/config/config.go:8-15` |
| Dockerfile | Alpine-based, no ONNX runtime | `backend/Dockerfile` |
| Monitor loop | Stub, awaits implementation | `backend/internal/agent/monitor.go` |

---

## Architecture

```
ListLogsSinceForApp returns []LogBuffer
                    │
                    ▼
        ┌───────────────────────┐
        │  ExtractText()        │  ← Pull classifiable text from JSON payload
        │  []LogBuffer → []string│
        └───────────┬───────────┘
                    │
                    ▼
        ┌───────────────────────┐
        │  Lumber ClassifyBatch │  ← Single batched ONNX forward pass (~0.5ms/line)
        │  []string → []Event  │
        └───────────┬───────────┘
                    │
                    ▼
        ┌───────────────────────┐
        │  ShouldEscalate()     │  ← Hardcoded severity gate per Event type+category
        │  Event → bool         │
        └───────────┬───────────┘
                    │
            ┌───────┴────────┐
            │                │
         safe             flagged
            │                │
         count it       → escalate to LLM (Step 5+)
```

The `Classifier` interface wraps this pipeline so the monitor loop (Step 5) gets a clean `Classify([]LogBuffer) → ([]ClassifiedLog, safeCount)` call.

---

## Implementation Tasks

### Task 1 — Config: Classifier Mode & Model Directory

**File:** `backend/internal/config/config.go`

Add two new environment variables to the `Config` struct:

```go
type Config struct {
    // ... existing fields ...
    ClassifierMode     string // "on", "off", "fallback" (default: "fallback")
    LumberModelDir     string // path to ONNX model files (default: "/app/models")
}
```

| Env Var | Default | Description |
|---------|---------|-------------|
| `CLASSIFIER_MODE` | `fallback` | `on` = required (fail if unavailable), `off` = skip classification (escalate all), `fallback` = use if available, else escalate all |
| `LUMBER_MODEL_DIR` | `/opt/lumber/models` | Directory containing `model_quantized.onnx`, `vocab.txt`, `2_Dense/model.safetensors` |

**No validation change needed** — both have sensible defaults. The classifier init (Task 4) handles the `on` mode failure case.

**Changes:**
- Add fields to `Config` struct
- Add `getEnv` calls in `Load()`

---

### Task 2 — Classifier Interface & Types

**New file:** `backend/internal/agent/classifier.go`

Define the interface and supporting types that the monitor loop (Step 5) will consume.

```go
package agent

import (
    "github.com/hejijunhao/heimdall/backend/internal/db"
)

// Classifier classifies a batch of log entries and returns which ones
// should be escalated to the LLM agent.
type Classifier interface {
    // Classify takes a batch of log buffer entries, classifies each one,
    // and returns the flagged (escalation-worthy) logs plus the count of safe logs.
    Classify(logs []db.LogBuffer) (flagged []ClassifiedLog, safeCount int)

    // Close releases resources (ONNX runtime, model memory).
    Close() error
}

// ClassifiedLog pairs a log buffer entry with its Lumber classification metadata.
type ClassifiedLog struct {
    Log        db.LogBuffer
    Type       string  // Lumber root category: ERROR, REQUEST, DEPLOY, etc.
    Category   string  // Lumber leaf label: connection_failure, success, etc.
    Severity   string  // error, warning, info, debug
    Confidence float64 // Cosine similarity score (0.0–1.0)
    Summary    string  // Lumber's compacted summary of the log
}
```

**Why an interface?** The monitor loop tests (Step 5) need to inject a mock classifier without ONNX. The interface costs 4 lines and decouples classification from orchestration.

---

### Task 3 — Text Extraction from LogBuffer Payloads

**New file:** `backend/internal/agent/extract.go`

The `LogBuffer.Payload` field is `json.RawMessage` — arbitrary JSON from webhook ingestion. Different providers (Fly.io, Vercel, Supabase) use different field names for the log message. We need a generic extraction strategy.

**Priority-ordered field scan:**

```go
package agent

import (
    "encoding/json"
    "strings"

    "github.com/hejijunhao/heimdall/backend/internal/db"
)

// messageFields is the priority-ordered list of JSON fields to check
// for the primary log message text.
var messageFields = []string{"message", "msg", "error", "text", "log", "body"}

// ExtractText pulls classifiable text from a LogBuffer entry's JSON payload.
// Strategy:
//  1. Try known message fields in priority order
//  2. If a "level"/"severity" field exists, prepend it for classification context
//  3. Fall back to the full JSON string (Lumber handles raw JSON fine)
func ExtractText(entry db.LogBuffer) string {
    var obj map[string]any
    if err := json.Unmarshal(entry.Payload, &obj); err != nil {
        // Not valid JSON — return raw string
        return strings.TrimSpace(string(entry.Payload))
    }

    // Find the primary message
    var message string
    for _, field := range messageFields {
        if v, ok := obj[field]; ok {
            if s, ok := v.(string); ok && s != "" {
                message = s
                break
            }
        }
    }

    // Prepend level/severity for classification context
    var prefix string
    for _, field := range []string{"level", "severity", "error_severity"} {
        if v, ok := obj[field]; ok {
            if s, ok := v.(string); ok && s != "" {
                prefix = strings.ToUpper(s)
                break
            }
        }
    }

    if message != "" {
        if prefix != "" {
            return prefix + ": " + message
        }
        return message
    }

    // No known message field — stringify the whole payload
    return strings.TrimSpace(string(entry.Payload))
}

// ExtractTexts extracts classifiable text from a batch of log entries.
// Returns texts in the same order as the input slice.
func ExtractTexts(logs []db.LogBuffer) []string {
    texts := make([]string, len(logs))
    for i, log := range logs {
        texts[i] = ExtractText(log)
    }
    return texts
}
```

**Why prepend the level?** A log like `"connection refused"` is ambiguous. But `"ERROR: connection refused"` gives Lumber a strong signal. The level field comes from the webhook payload's own severity metadata (e.g., Fly.io sends `"level": "error"`).

---

### Task 4 — Severity Gate

**New file:** `backend/internal/agent/severity_gate.go`

Pure function that decides whether a classified log should be escalated to the LLM. Based on the rules table from the Phase 8 master plan.

```go
package agent

import "github.com/kaminocorp/lumber/pkg/lumber"

// ShouldEscalate returns true if a classified log event warrants LLM attention.
// Rules are hardcoded per the Phase 8 severity gate table.
// Can be made configurable per-app in a future phase.
func ShouldEscalate(event lumber.Event) bool {
    switch event.Type {
    case "ERROR":
        return true // All errors escalate
    case "PERFORMANCE":
        return true // All performance issues escalate

    case "REQUEST":
        switch event.Category {
        case "server_error", "slow_request":
            return true
        default:
            return false // success, redirect, client_error → safe
        }

    case "DEPLOY":
        return true // All deploy events are worth noting

    case "SYSTEM":
        switch event.Category {
        case "resource_alert", "config_change":
            return true
        default:
            return false // health_check, process_lifecycle, scaling_event → safe
        }

    case "ACCESS":
        switch event.Category {
        case "login_failure", "auth_failure", "permission_change", "api_key_event":
            return true
        default:
            return false // login_success, session_expired → safe
        }

    case "DATA":
        switch event.Category {
        case "migration":
            return true
        default:
            return false // query_executed, replication → safe
        }

    case "SCHEDULED":
        switch event.Category {
        case "cron_failed":
            return true
        default:
            return false // cron_started, cron_completed → safe
        }

    case "UNCLASSIFIED":
        return true // Uncertain = worth checking

    default:
        // Unknown type — escalate to be safe
        return true
    }
}
```

**Confidence backstop:** Additionally, any event with `Confidence < 0.5` should escalate regardless of type, since the classification is uncertain. This is handled in the `LumberClassifier.Classify()` method (Task 5), not in this function, because confidence is a classifier concern while this function is a domain-logic concern.

---

### Task 5 — LumberClassifier Implementation

**New file:** `backend/internal/agent/classifier_lumber.go`

The concrete `Classifier` implementation wrapping the Lumber library.

```go
package agent

import (
    "log/slog"

    "github.com/kaminocorp/lumber/pkg/lumber"
    "github.com/hejijunhao/heimdall/backend/internal/db"
)

// LumberClassifier wraps the Lumber library to classify log entries.
type LumberClassifier struct {
    engine *lumber.Lumber
}

// NewLumberClassifier creates a new classifier backed by the Lumber ONNX model.
// modelDir must contain model_quantized.onnx, vocab.txt, and 2_Dense/model.safetensors.
// Returns an error if model files are missing or ONNX runtime fails to initialize.
func NewLumberClassifier(modelDir string) (*LumberClassifier, error) {
    l, err := lumber.New(
        lumber.WithModelDir(modelDir),
        lumber.WithConfidenceThreshold(0.5),
        lumber.WithVerbosity("minimal"),
    )
    if err != nil {
        return nil, err
    }
    return &LumberClassifier{engine: l}, nil
}

func (c *LumberClassifier) Classify(logs []db.LogBuffer) ([]ClassifiedLog, int) {
    if len(logs) == 0 {
        return nil, 0
    }

    // 1. Extract text from each log's payload
    texts := ExtractTexts(logs)

    // 2. Classify via Lumber (batched ONNX inference)
    events, err := c.engine.ClassifyBatch(texts)
    if err != nil {
        slog.Error("lumber classification failed, escalating all logs", "err", err, "count", len(logs))
        // On classification failure, escalate everything (fail-open)
        flagged := make([]ClassifiedLog, len(logs))
        for i, log := range logs {
            flagged[i] = ClassifiedLog{
                Log:      log,
                Type:     "UNCLASSIFIED",
                Severity: "warning",
                Summary:  texts[i],
            }
        }
        return flagged, 0
    }

    // 3. Apply severity gate to each classified event
    var flagged []ClassifiedLog
    safeCount := 0

    for i, event := range events {
        // Confidence backstop: uncertain classifications always escalate
        if event.Confidence < 0.5 || ShouldEscalate(event) {
            flagged = append(flagged, ClassifiedLog{
                Log:        logs[i],
                Type:       event.Type,
                Category:   event.Category,
                Severity:   event.Severity,
                Confidence: event.Confidence,
                Summary:    event.Summary,
            })
        } else {
            safeCount++
        }
    }

    return flagged, safeCount
}

func (c *LumberClassifier) Close() error {
    return c.engine.Close()
}
```

**Fail-open design:** If `ClassifyBatch` errors, we escalate everything rather than silently dropping logs. This matches the `fallback` philosophy — when in doubt, let the LLM see it.

---

### Task 6 — PassthroughClassifier (Off / Fallback)

**Same file or add to:** `backend/internal/agent/classifier.go`

When the classifier is unavailable (mode = `off`, or mode = `fallback` and Lumber failed to init), use a passthrough that escalates everything.

```go
// PassthroughClassifier escalates all logs without classification.
// Used when classifier mode is "off" or when Lumber is unavailable in "fallback" mode.
type PassthroughClassifier struct{}

func (p *PassthroughClassifier) Classify(logs []db.LogBuffer) ([]ClassifiedLog, int) {
    flagged := make([]ClassifiedLog, len(logs))
    for i, log := range logs {
        flagged[i] = ClassifiedLog{
            Log:      log,
            Type:     "UNCLASSIFIED",
            Severity: "unknown",
            Summary:  ExtractText(log),
        }
    }
    return flagged, 0
}

func (p *PassthroughClassifier) Close() error { return nil }
```

---

### Task 7 — Wire Classifier into Agent

**File:** `backend/internal/agent/agent.go`

Add the classifier as a field on the `Agent` struct. The classifier is initialized in `main.go` and passed to the agent constructor.

**Updated struct and constructor:**

```go
type Agent struct {
    queries    *db.Queries
    client     *anthropic.Client
    config     *config.Config
    classifier Classifier
}

func New(queries *db.Queries, cfg *config.Config, classifier Classifier) *Agent {
    client := anthropic.NewClient(option.WithAPIKey(cfg.AnthropicKey))
    return &Agent{
        queries:    queries,
        client:     &client,
        config:     cfg,
        classifier: classifier,
    }
}
```

**File:** `backend/cmd/heimdall/main.go`

Add classifier initialization between pool creation and agent construction:

```go
// Initialize classifier based on CLASSIFIER_MODE
var classifier agent.Classifier
switch cfg.ClassifierMode {
case "on":
    c, err := agent.NewLumberClassifier(cfg.LumberModelDir)
    if err != nil {
        slog.Error("classifier required but failed to initialize", "err", err)
        os.Exit(1)
    }
    classifier = c
    slog.Info("lumber classifier initialized", "model_dir", cfg.LumberModelDir)
case "off":
    classifier = &agent.PassthroughClassifier{}
    slog.Info("classifier disabled, all logs will be escalated")
default: // "fallback"
    c, err := agent.NewLumberClassifier(cfg.LumberModelDir)
    if err != nil {
        slog.Warn("lumber classifier unavailable, falling back to passthrough", "err", err)
        classifier = &agent.PassthroughClassifier{}
    } else {
        classifier = c
        slog.Info("lumber classifier initialized", "model_dir", cfg.LumberModelDir)
    }
}

ag := agent.New(db.New(pool), cfg, classifier)
```

**Shutdown:** Add `classifier.Close()` to the shutdown sequence (before `pool.Close()`).

---

### Task 8 — Dockerfile: ONNX Runtime & Model Files

**File:** `backend/Dockerfile`

The current Dockerfile uses `alpine` (musl libc). ONNX Runtime requires glibc. Switch to `debian:bookworm-slim` for the runtime stage and add model/library download.

```dockerfile
FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /bin/heimdall ./cmd/heimdall

# Download Lumber model files and ONNX runtime
FROM golang:1.24 AS models
WORKDIR /tmp
RUN apt-get update && apt-get install -y curl && \
    mkdir -p /models/2_Dense && \
    curl -fsSL -o /models/model_quantized.onnx \
      https://huggingface.co/onnx-community/mdbr-leaf-mt-ONNX/resolve/main/onnx/model_quantized.onnx && \
    curl -fsSL -o /models/vocab.txt \
      https://huggingface.co/onnx-community/mdbr-leaf-mt-ONNX/resolve/main/vocab.txt && \
    curl -fsSL -o /models/2_Dense/model.safetensors \
      https://huggingface.co/Snowflake/mdbr-leaf-mt/resolve/main/2_Dense/model.safetensors
# ONNX Runtime library — try Lumber's Makefile first, fall back to direct ONNX Runtime release download
RUN git clone --depth 1 https://github.com/kaminocorp/lumber.git /tmp/lumber && \
    cd /tmp/lumber && make download-ort && \
    cp lib/libonnxruntime.* /usr/local/lib/

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=build /bin/heimdall /usr/local/bin/heimdall
COPY --from=models /models /opt/lumber/models
COPY --from=models /usr/local/lib/libonnxruntime.* /usr/local/lib/
RUN ldconfig
EXPOSE 8080
ENTRYPOINT ["heimdall"]
```

**Important notes:**
- The `models` stage is a separate build stage so model downloads are cached independently of code changes
- Models are placed at `/opt/lumber/models` (not `/app/models`) to avoid being shadowed by the `./backend:/app` volume mount in `docker-compose.yml`
- `LUMBER_MODEL_DIR` defaults to `/opt/lumber/models` in `config.go`, matching the COPY destination
- The ONNX runtime library fetch uses Lumber's `make download-ort` target. If this proves problematic during implementation, fall back to downloading `libonnxruntime` directly from [ONNX Runtime GitHub releases](https://github.com/microsoft/onnxruntime/releases)

---

### Task 9 — Unit Tests

**New file:** `backend/internal/agent/extract_test.go`

Test text extraction from various payload formats:

| Test case | Payload | Expected output |
|-----------|---------|-----------------|
| Standard `message` field | `{"level":"error","message":"connection refused"}` | `ERROR: connection refused` |
| Shorthand `msg` field | `{"msg":"request completed","status":200}` | `request completed` |
| `error` field | `{"error":"timeout after 30s"}` | `timeout after 30s` |
| Level prepended | `{"level":"warn","message":"high latency"}` | `WARN: high latency` |
| Supabase-style | `{"error_severity":"ERROR","msg":"deadlock detected"}` | `ERROR: deadlock detected` |
| No message field | `{"path":"/api/users","status":500}` | `{"path":"/api/users","status":500}` (raw JSON) |
| Non-JSON payload | `plain text log line` | `plain text log line` |
| Empty payload | `{}` | `{}` |

**New file:** `backend/internal/agent/severity_gate_test.go`

Test every branch of the severity gate:

| Test group | Escalated | Safe |
|------------|-----------|------|
| ERROR | `connection_failure`, `runtime_exception`, `timeout` | _(none — all escalate)_ |
| REQUEST | `server_error`, `slow_request` | `success`, `redirect`, `client_error` |
| DEPLOY | `build_started`, `deploy_failed`, `rollback` | _(none — all escalate)_ |
| SYSTEM | `resource_alert`, `config_change` | `health_check`, `process_lifecycle` |
| ACCESS | `login_failure`, `permission_change` | `login_success`, `session_expired` |
| PERFORMANCE | `latency_spike`, `db_slow_query` | _(none — all escalate)_ |
| DATA | `migration` | `query_executed`, `replication` |
| SCHEDULED | `cron_failed` | `cron_started`, `cron_completed` |
| UNCLASSIFIED | _(always)_ | _(never)_ |

These tests are pure Go — no ONNX runtime, no Docker, fast CI.

**New file:** `backend/internal/agent/classifier_test.go`

Integration test for `LumberClassifier` (requires Docker / ONNX runtime):

| Test | Input | Expected |
|------|-------|----------|
| Known error log | `"ERROR: connection refused to db-primary:5432"` | Flagged, type=ERROR |
| Known healthy request | `"GET /api/health 200 OK 3ms"` | Safe (REQUEST.success) |
| Mixed batch | 5 safe + 2 errors | `safeCount=5`, `len(flagged)=2` |
| Empty batch | `[]` | `nil, 0` |
| PassthroughClassifier | Any logs | All flagged, `safeCount=0` |

---

## File Summary

| Action | File | Task |
|--------|------|------|
| Edit | `backend/internal/config/config.go` | 1 |
| Create | `backend/internal/agent/classifier.go` | 2, 6 |
| Create | `backend/internal/agent/extract.go` | 3 |
| Create | `backend/internal/agent/severity_gate.go` | 4 |
| Create | `backend/internal/agent/classifier_lumber.go` | 5 |
| Edit | `backend/internal/agent/agent.go` | 7 |
| Edit | `backend/cmd/heimdall/main.go` | 7 |
| Edit | `backend/Dockerfile` | 8 |
| Create | `backend/internal/agent/extract_test.go` | 9 |
| Create | `backend/internal/agent/severity_gate_test.go` | 9 |
| Create | `backend/internal/agent/classifier_test.go` | 9 |

---

## Implementation Order

```
Task 1:  Config fields (ClassifierMode, LumberModelDir)
Task 2:  Classifier interface + ClassifiedLog type
Task 3:  ExtractText / ExtractTexts
Task 4:  ShouldEscalate severity gate
         ↓ (these four have no external dependencies — pure Go)
Task 5:  LumberClassifier (wraps Lumber library)
Task 6:  PassthroughClassifier
Task 7:  Wire into Agent struct + main.go init
Task 8:  Dockerfile (ONNX runtime + model files)
         ↓
Task 9:  Tests (extract, severity gate, classifier integration)
```

**Tasks 1–4** can be implemented and tested without Docker or ONNX.
**Tasks 5–8** require Lumber's ONNX model to be available.
**Task 9** spans both: unit tests (no ONNX) and integration tests (ONNX in Docker).

---

## Decisions (Pre-Resolved)

1. **Classifier is a router, not a filter.** All logs stay in `log_buffer`. Classification only determines LLM escalation.

2. **Three-way classifier mode** (`on`/`off`/`fallback`) as a global startup config via `CLASSIFIER_MODE` env var. Default: `fallback`.

3. **Text extraction** uses priority-ordered field scan (`message` → `msg` → `error` → `text` → `log` → `body`), with level prepended. Falls back to raw JSON.

4. **Severity gate** is a hardcoded pure function. Configurable per-app rules are a future phase.

5. **Fail-open on classification error.** If Lumber's `ClassifyBatch` returns an error, all logs in the batch are escalated.

6. **Dockerfile switches from Alpine to Debian** (bookworm-slim) for glibc compatibility with ONNX Runtime.

7. **Confidence backstop at 0.5.** Events below this threshold escalate regardless of type — handled in `LumberClassifier.Classify()`, not in the severity gate function.

---

## Resolved Questions

1. **Lumber `make download-ort` target** — try this first in the Docker build. If it fails (missing deps, unexpected Makefile behaviour), fall back to downloading `libonnxruntime` directly from the [ONNX Runtime GitHub releases](https://github.com/microsoft/onnxruntime/releases). Resolved during implementation.

2. **docker-compose volume mount conflict** — the backend service mounts `./backend:/app`, which would shadow any Docker-built `/app/models` directory. **Resolution:** place model files at `/opt/lumber/models` instead. This path is outside the mount and safe from shadowing. `LUMBER_MODEL_DIR` defaults to `/opt/lumber/models` accordingly.

3. **`go.mod` indirect → direct** — Lumber is currently an indirect dependency. Writing `import "github.com/kaminocorp/lumber/pkg/lumber"` in `classifier_lumber.go` will promote it to a direct dependency. **Resolution:** run `go mod tidy` after Task 5 is implemented. No risk — this is a cosmetic change in `go.mod`/`go.sum`.

---

## Post-Implementation Summary

Lumber compiles into the Heimdall binary as a Go library dependency. At startup, `main.go` checks the `CLASSIFIER_MODE` env var (`on`/`off`/`fallback`) and either initializes a `LumberClassifier` (which loads the ONNX model from `/opt/lumber/models` into memory, ~100-300ms) or falls back to a `PassthroughClassifier` that escalates everything. The classifier instance is stored on the `Agent` struct and shared across all monitoring goroutines.

When the monitor loop (Step 5) fetches a batch of logs, it calls `classifier.Classify(logs)` — a single in-process function call, no network hop. Under the hood:

- **Text extraction** — parses each `LogBuffer.Payload` JSON, pulls out the message field (`message`/`msg`/`error`/etc.), prepends the log level if present
- **ONNX inference** — Lumber's `ClassifyBatch` runs a single batched forward pass on CPU (~0.5ms/line), returning a taxonomy label and confidence score per log
- **Severity gate** — `ShouldEscalate()` applies hardcoded rules (all ERRORs escalate, REQUEST.success is safe, UNCLASSIFIED escalates, etc.)
- **Output** — returns `flagged []ClassifiedLog` (sent to Claude) and `safeCount int` (logged in heartbeat, never sent to LLM)
- **Fail-open** — if classification errors or confidence is below 0.5, logs escalate rather than being silently ignored
