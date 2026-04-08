# Lumber Integration Hardening (2026-04-06)

Review and hardening pass on the Lumber ONNX classifier pipeline, triggered by a post-integration audit. Six issues were identified and resolved. A reference blueprint was also produced.

---

## What Was Done

### 1. Lumber upgraded v0.9.0 → v0.10.6

**File:** `backend/go.mod`, `backend/go.sum`

**Why:** v0.9.0 was missing a `sync.Mutex` protecting ONNX inference calls in `ONNXEmbedder`. The underlying `DynamicAdvancedSession` is not thread-safe. With `maxConcurrentApps: 10` all sharing the same `*lumber.Lumber` instance and potentially calling `ClassifyBatch` concurrently, v0.9.0 had a real risk of memory corruption or crashes under load. v0.10.6 adds the mutex.

**How:** `go get github.com/kaminocorp/lumber@v0.10.6` (v0.10.5 does not exist; v0.10.6 is latest).

---

### 2. Dead DATA/SCHEDULED branches removed from severity gate

**Files:** `backend/internal/agent/severity_gate.go`, `backend/internal/agent/severity_gate_test.go`

**Why:** Lumber's current taxonomy has six root types: `ERROR`, `REQUEST`, `DEPLOY`, `SYSTEM`, `ACCESS`, `PERFORMANCE`. There is no `DATA` or `SCHEDULED` type. Any log that would semantically match those is returned as `UNCLASSIFIED` by the model — the `DATA`/`SCHEDULED` branches in `ShouldEscalate` would never execute. The escalation still happened (via the `UNCLASSIFIED → all` default rule), but the intended category-level nuance (e.g. treating `DATA.query_executed` as safe) was silently lost.

**How:** Removed the `DATA` and `SCHEDULED` switch cases from `severity_gate.go`. Added a comment documenting the six actual root types. Updated `severity_gate_test.go` to reflect that all `DATA`/`SCHEDULED` events now fall through to `default: return true` (unknown types escalate unconditionally) — moved the previously "safe" cases (`query_executed`, `replication`, `cron_started`, `cron_completed`) into the escalated list with a comment.

---

### 3. Confidence threshold ownership consolidated

**File:** `backend/internal/agent/classifier_lumber.go`

**Why:** `NewLumberClassifier` set `lumber.WithConfidenceThreshold(0.5)`, which causes Lumber to relabel low-confidence events internally. `classifier_lumber.go:56` then re-checked `event.Confidence < 0.5` explicitly. Two separate places controlling the same policy: if one is updated (e.g. tuning to 0.3 for a more sensitive deployment), the other would silently diverge.

**How:** Removed `lumber.WithConfidenceThreshold(0.5)` from `NewLumberClassifier`. The explicit `event.Confidence < 0.5` check in `Classify` is now the single source of truth. Added a comment explaining the ownership decision.

---

### 4. Rate limiter added on Claude invocations

**Files:** `backend/internal/agent/agent.go`, `backend/internal/agent/monitor.go`

**Why:** The default `CLASSIFIER_MODE` is `fallback`. If the model fails to load (wrong path, corrupted file, ORT version mismatch), `PassthroughClassifier` silently takes over and escalates every log to Claude. With `maxConcurrentApps: 10`, `logBatchLimit: 200`, and a 15-second tick, this could drive significant unintended Claude API cost with no signal beyond a startup warning log.

**How:** Added `golang.org/x/time/rate` (promoted from transitive to direct dependency). Added a `*rate.Limiter` field to `Agent`, initialised in `New()` as `rate.NewLimiter(rate.Every(2*time.Second), 5)` — 30 sustained invocations per minute, burst of 5. In `monitorApp`, `a.limiter.Wait(ctx)` is called immediately before `RunMonitoring`. If the context is cancelled while waiting, the function returns early without calling Claude.

The rate is permissive enough for normal operation (10 apps with a working classifier rarely produce 30 escalations per minute) but caps the damage if classification breaks.

---

### 5. Payload size guard added before classification

**File:** `backend/internal/agent/extract.go`

**Why:** `ExtractText` falls back to the raw JSON string when no known message field is found. Lumber's BERT-style tokenizer silently truncates at ~512 tokens, so very large JSON payloads degrade classification quality unpredictably and invisibly.

**How:** Added `maxClassifyChars = 1000` constant. On the raw JSON fallback path, the string is now truncated to 1000 UTF-8 characters before being returned. Message and severity fields extracted from known keys are not truncated (they are already short strings). Added `unicode/utf8` import.

---

### 6. Heartbeat suppressed when no flagged logs

**File:** `backend/internal/agent/monitor.go`

**Why:** `EmitLogWithSeverity` was called on every app on every 15-second tick, writing an `agent_log` entry even when all logs were safe. With multiple apps in continuous monitoring mode, this generates significant log volume in the UI with no signal value.

**How:** Moved the heartbeat `EmitLogWithSeverity` call inside the `if len(flagged) > 0` block. The agent_log entry is now only emitted when there is something to report. Go `slog` output to stdout continues for every processed batch regardless.

---

### 7. Blueprint produced

**File:** `docs/blueprints/lumber-integration.md`

Comprehensive reference document covering: what Lumber is, model files and location, the `Classifier` interface and both implementations, startup modes (`CLASSIFIER_MODE`), text extraction strategy, severity gate rules, the full data flow from ingestion to cursor advancement, key constants, relevant file map, and test instructions.

Updated after the hardening pass to reflect the actual state: v0.10.6, six-type taxonomy, consolidated threshold ownership, rate limiter, 1000-char extraction cap, and heartbeat change.

---

## Files Changed

| File | Change |
|------|--------|
| `backend/go.mod` | Lumber v0.9.0 → v0.10.6; `golang.org/x/time` promoted to direct |
| `backend/go.sum` | Updated |
| `backend/internal/agent/severity_gate.go` | DATA/SCHEDULED branches removed; taxonomy comment added |
| `backend/internal/agent/severity_gate_test.go` | DATA/SCHEDULED safe cases moved to escalated; comments updated |
| `backend/internal/agent/classifier_lumber.go` | `WithConfidenceThreshold` removed; ownership comment added |
| `backend/internal/agent/extract.go` | `maxClassifyChars = 1000`; raw JSON fallback now truncated |
| `backend/internal/agent/agent.go` | `*rate.Limiter` field added; `monitorLLMRate` var; `golang.org/x/time/rate` imported |
| `backend/internal/agent/monitor.go` | `limiter.Wait(ctx)` before `RunMonitoring`; heartbeat moved inside flagged block |
| `docs/blueprints/lumber-integration.md` | New — full Lumber integration reference |
