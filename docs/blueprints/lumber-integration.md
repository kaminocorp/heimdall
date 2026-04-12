# Lumber Integration Blueprint

How Heimdall uses Lumber — what it is, where it lives, and how log entries travel from ingestion through ONNX classification to LLM escalation.

---

## 1. What is Lumber?

Lumber is a Go library (`github.com/kaminocorp/lumber v0.10.6`) that classifies log entries using an ONNX model running via `onnxruntime_go`. It takes raw log strings, runs them through a quantized embedding model (`mdbr-leaf-mt`), and returns structured events with a type, category, severity, confidence score, and a short summary.

The underlying model is MongoDB's `mdbr-leaf-mt`, a multi-label log classifier hosted on HuggingFace. It is quantized for production inference and runs locally — no external API call.

**Lumber's output per log entry:**

| Field | Type | Example |
|-------|------|---------|
| `Type` | string | `ERROR`, `REQUEST`, `DEPLOY`, `SYSTEM`, `ACCESS`, `PERFORMANCE` (plus `UNCLASSIFIED` for off-taxonomy logs) |
| `Category` | string | `connection_failure`, `server_error`, `login_failure`, `slow_request` |
| `Severity` | string | `error`, `warning`, `info`, `debug` |
| `Confidence` | float64 | `0.87` (cosine similarity, 0.0–1.0) |
| `Summary` | string | Lumber's compacted plain-text summary of the log |

---

## 2. Model Files & Location

The model is loaded from disk at startup. Default path: `/opt/lumber/models` (configurable via `LUMBER_MODEL_DIR`).

```
/opt/lumber/models/
├── model_quantized.onnx              ← main ONNX model
├── model_quantized.onnx_data         ← external data tensors for quantized model
├── vocab.txt                         ← tokenizer vocabulary
├── tokenizer_config.json             ← tokenizer config
├── libonnxruntime.so                 ← ONNX Runtime shared library (v1.24.1)
└── 2_Dense/
    ├── model.safetensors             ← projection layer weights
    └── config.json                   ← projection layer config
```

These files are downloaded at Docker image build time (`backend/Dockerfile`):
- `model_quantized.onnx`, `vocab.txt`, `tokenizer_config.json` → HuggingFace: `onnx-community/mdbr-leaf-mt-ONNX`
- `2_Dense/` weights → HuggingFace: `MongoDB/mdbr-leaf-mt`
- `libonnxruntime.so` → GitHub: Microsoft ONNX Runtime v1.24.1

The Dockerfile uses a three-stage build: compile the Go binary, download the model files, then assemble the final runtime image.

---

## 3. Go Interface & Implementations

**File:** `backend/internal/agent/classifier.go`

```go
type Classifier interface {
    Classify(logs []db.LogBuffer) (flagged []ClassifiedLog, safeCount int)
    Close() error
}

type ClassifiedLog struct {
    Log        db.LogBuffer
    Type       string   // Lumber root type
    Category   string   // Lumber leaf category
    Severity   string   // error / warning / info / debug
    Confidence float64  // 0.0–1.0 cosine similarity
    Summary    string   // Lumber's plain-text summary
}
```

Two concrete implementations:

### LumberClassifier (ONNX-backed)

**File:** `backend/internal/agent/classifier_lumber.go`

Wraps `*lumber.Lumber`. Constructed via `NewLumberClassifier(modelDir string)`, which initialises the engine with:
- `lumber.WithModelDir(modelDir)`
- `lumber.WithVerbosity("minimal")`

The confidence threshold is **not** delegated to Lumber's `WithConfidenceThreshold` option — it is owned explicitly in `Classify` (see §6) so the policy lives in one place.

`Classify` pipeline:
1. Extract text from each log (`ExtractTexts`, capped at 1000 chars)
2. Call `engine.ClassifyBatch(texts)` → `[]lumber.Event`
3. For each event: apply confidence threshold + severity gate (see §5)
4. Return flagged logs and safe count

On `ClassifyBatch` error: all logs are escalated as `UNCLASSIFIED / warning` and processing continues — no crash.

### PassthroughClassifier (no-op fallback)

**File:** `backend/internal/agent/classifier.go`

Escalates every log as `UNCLASSIFIED / unknown`. No model involved. Used when Lumber is disabled or unavailable.

---

## 4. Startup & CLASSIFIER_MODE

**File:** `backend/cmd/heimdall/main.go` (lines 52–75)

The classifier is selected once at startup based on the `CLASSIFIER_MODE` env var (default: `fallback`):

| Mode | Behaviour |
|------|-----------|
| `on` | Initialise `LumberClassifier`; if it fails, exit with code 1 |
| `off` | Use `PassthroughClassifier`; all logs escalated |
| `fallback` | Try `LumberClassifier`; on failure, warn and use `PassthroughClassifier` |

The `LUMBER_MODEL_DIR` env var (default: `/opt/lumber/models`) is passed directly to `NewLumberClassifier`.

At shutdown (SIGINT/SIGTERM), `classifier.Close()` is called to release the ONNX runtime and model memory.

---

## 5. Text Extraction

**File:** `backend/internal/agent/extract.go`

Before classification, log payloads (stored as JSONB in `log_buffer`) are converted to plain strings. `ExtractText` applies this strategy:

1. Parse the JSON payload into `map[string]any`
2. Look for a message field in priority order: `message` → `msg` → `error` → `text` → `log` → `body`
3. Look for a severity prefix in: `level` → `severity` → `error_severity`
4. If both found, return `"SEVERITY: message"` (e.g. `"ERROR: connection refused"`)
5. Fallback: return the raw JSON string, truncated to **1000 characters** (`maxClassifyChars`)

The 1000-char cap on the fallback path makes the tokenizer boundary explicit. The underlying BERT-style model silently truncates at ~512 tokens; without the cap, very large JSON payloads degrade classification quality unpredictably.

`ExtractTexts` is the batch version, preserving index order for alignment with `ClassifyBatch` results.

---

## 6. Severity Gate

**File:** `backend/internal/agent/severity_gate.go`

`ShouldEscalate(event lumber.Event) bool` is the policy layer that sits on top of the model output. A log is escalated to Claude if **either**:
- Its confidence is below `0.5` (model is uncertain), **or**
- `ShouldEscalate` returns true for its type + category

Lumber's current taxonomy has **six root types**: `ERROR`, `REQUEST`, `DEPLOY`, `SYSTEM`, `ACCESS`, `PERFORMANCE`. Any log that doesn't fit is returned as `UNCLASSIFIED` by the model and escalated via the default branch. The gate rules cover exactly these six types:

| Type | Escalate categories | Safe categories |
|------|---------------------|-----------------|
| `ERROR` | all | — |
| `PERFORMANCE` | all | — |
| `REQUEST` | `server_error`, `slow_request` | `success`, `redirect`, `client_error` |
| `DEPLOY` | all | — |
| `SYSTEM` | `resource_alert`, `config_change` | `health_check`, `process_lifecycle`, `scaling_event` |
| `ACCESS` | `login_failure`, `auth_failure`, `permission_change`, `api_key_event` | `login_success`, `session_expired` |
| `UNCLASSIFIED` / unknown type | all | — |

The gate is hardcoded — tuning it requires a code change, not config. This is intentional for v1 but worth knowing as an operator.

---

## 7. Data Flow: Ingestion → Classification → Escalation

```
Log arrives (webhook / syslog / OTLP / poller)
  └─ Stored in log_buffer (JSONB payload)

Monitor loop fires every 15s (monitor.go)
  └─ For each active app, fetch up to 200 new logs since cursor

ExtractTexts (extract.go)
  └─ Parse JSONB → plain strings, prepend severity prefix if present

Classifier.Classify (classifier_lumber.go or classifier.go)
  └─ LumberClassifier: engine.ClassifyBatch(texts) → []lumber.Event
  └─ PassthroughClassifier: mark all as UNCLASSIFIED

Severity gate (severity_gate.go)
  └─ confidence < 0.5  → escalate
  └─ ShouldEscalate()  → escalate
  └─ otherwise         → safe (increment safeCount)

Metrics (monitor.go)
  └─ heimdall_logs_classified_total{result="safe"|"flagged"} incremented

If no flagged logs → done (no agent_log entry emitted, reducing UI noise)

If flagged logs exist (capped at 50 via maxFlaggedForLLM)
  └─ formatFlaggedLogs: structured text summary for each log
     (classification metadata + timestamp + raw payload truncated to 2000 chars)

Rate limiter (agent.go)
  └─ agent.limiter.Wait(ctx) — 30/min sustained (1 per 2s), burst 5
  └─ Caps Claude API cost if classifier falls back to PassthroughClassifier

RunMonitoring (loop.go)
  └─ Claude Sonnet 4.6 invoked with formatted summary + monitoring tools
  └─ Max 10 tool-use iterations
  └─ Returns (assessment string, severity string)

Emit agent_log entry + dispatch notifications (email / Slack / Discord)

Advance cursor: monitoring_state.last_monitored_at = last log timestamp
```

---

## 8. Key Constants

Defined in `backend/internal/agent/monitor.go`:

| Constant | Value | Purpose |
|----------|-------|---------|
| `monitorTickInterval` | 15s | How often the monitor loop fires |
| `logBatchLimit` | 200 | Max logs fetched per app per tick |
| `maxFlaggedForLLM` | 50 | Cap on logs sent to Claude per tick |
| `maxPayloadChars` | 2000 | Payload truncation in LLM prompt |
| `monitorAppTimeout` | 2m | Per-app processing deadline |
| `maxConcurrentApps` | 10 | Semaphore limit on parallel app processing |

Confidence threshold: `0.5`, owned explicitly in `classifier_lumber.go:57` — **not** delegated to Lumber's `WithConfidenceThreshold` option. `NewLumberClassifier` only passes `WithModelDir` and `WithVerbosity` to the engine; the gate policy (threshold + `ShouldEscalate`) lives entirely in the Heimdall side of the boundary so it's reviewed as Heimdall code, not Lumber config.

---

## 9. Relevant Files

| File | Role |
|------|------|
| `backend/internal/agent/classifier.go` | `Classifier` interface, `ClassifiedLog` struct, `PassthroughClassifier` |
| `backend/internal/agent/classifier_lumber.go` | `LumberClassifier` wrapping `lumber.Lumber` |
| `backend/internal/agent/severity_gate.go` | `ShouldEscalate` policy rules |
| `backend/internal/agent/extract.go` | JSON payload → plain text extraction |
| `backend/internal/agent/monitor.go` | Monitor loop, pipeline orchestration, metrics, cursor |
| `backend/internal/agent/loop.go` | `RunMonitoring` — Claude invocation with flagged logs |
| `backend/internal/config/config.go` | `ClassifierMode`, `LumberModelDir` config fields |
| `backend/cmd/heimdall/main.go` | Classifier startup switch + shutdown cleanup |
| `backend/Dockerfile` | Three-stage build: compile, model download, runtime |
| `backend/internal/agent/classifier_test.go` | Unit + integration tests (skipped without `LUMBER_MODEL_DIR`) |
| `backend/internal/agent/severity_gate_test.go` | Full coverage of escalation rules |

---

## 10. Testing

**`TestPassthroughClassifier`** — verifies all logs are flagged, empty batch is safe, `Close` is a no-op.

**`TestLumberClassifier_Integration`** — skipped unless `LUMBER_MODEL_DIR` is set. Tests empty batch, known error escalation, healthy request non-escalation, and mixed batches. Requires the model files to be present.

**`TestShouldEscalate`** — 35 scenarios covering all type/category combinations; no model needed.

To run classifier tests with the real model:
```bash
LUMBER_MODEL_DIR=/opt/lumber/models go test ./internal/agent -run TestLumberClassifier
```
