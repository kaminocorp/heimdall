# Lumber Library API — Proposal for Heimdall Integration

This document outlines the ideal public Go API that Lumber should expose for in-process library use. Written for the Lumber developers.

---

## Context

Heimdall is a monitoring agent that ingests production logs, classifies them, and escalates problematic ones to an LLM for investigation. We want to use Lumber as the classification engine — but **in-process**, not as a CLI or sidecar. Today, all of Lumber's engine logic lives under `internal/`, so none of it is importable.

**What we need:** A public API that lets us load the ONNX model once at startup, then classify batches of raw log text on demand. We don't need Lumber's connectors, output, pipeline, dedup, or compactor — just the **embed → classify** core.

---

## Proposed Public API

Everything below would live in the top-level `lumber` package (i.e. `github.com/kaminocorp/lumber`).

### Types

```go
package lumber

// Event is the classification result for a single log line.
// Mirrors the existing internal CanonicalEvent, minus pipeline-specific fields.
type Event struct {
    Type       string  `json:"type"`       // e.g. "ERROR", "REQUEST", "DEPLOY"
    Category   string  `json:"category"`   // e.g. "connection_failure", "slow_request"
    Severity   string  `json:"severity"`   // "error", "warning", "info", "debug"
    Confidence float64 `json:"confidence"` // 0.0–1.0 cosine similarity score
    Summary    string  `json:"summary"`    // one-line summary extracted from the raw text
}
```

**Why this shape:**
- `Type` + `Category` gives the full taxonomy path (e.g. `ERROR.connection_failure`)
- `Severity` is the leaf-level severity from the taxonomy, not caller-supplied
- `Confidence` lets consumers apply their own thresholds (Heimdall escalates anything < 0.5 as uncertain)
- `Summary` is useful for human-readable agent log entries
- No `Raw` or `Timestamp` — the caller already has those, no need to echo them back

### Constructor

```go
// Option configures a Lumber engine.
type Option func(*options)

// WithConfidenceThreshold sets the minimum cosine similarity for a
// classification to be accepted. Below this, the event is UNCLASSIFIED.
// Default: 0.5
func WithConfidenceThreshold(t float64) Option

// WithVerbosity controls summary compaction level.
// Default: Standard
func WithVerbosity(v Verbosity) Option

// Lumber is the classification engine. Safe for concurrent use.
type Lumber struct { /* ... */ }

// New loads the ONNX model and pre-embeds the taxonomy.
// modelDir should contain: model_quantized.onnx, vocab.txt, projection.bin
//
// This is the expensive call (~200ms). Call once at startup.
func New(modelDir string, opts ...Option) (*Lumber, error)

// Close releases ONNX runtime resources.
func (l *Lumber) Close() error
```

**Why `modelDir` instead of individual paths:**
- Simpler API — one directory, Lumber knows its own file layout
- The three files (`model_quantized.onnx`, `vocab.txt`, `projection.bin`) are always co-located
- Aligns with how the model is downloaded (single directory target)

### Classification Methods

```go
// Classify classifies a single log line.
func (l *Lumber) Classify(text string) Event

// ClassifyBatch classifies multiple log lines in a single batched ONNX
// inference call. Significantly faster than calling Classify in a loop.
func (l *Lumber) ClassifyBatch(texts []string) []Event
```

**Why no `error` return:**
- The existing engine code already handles empty/whitespace input gracefully (returns `UNCLASSIFIED`)
- ONNX inference errors on a loaded model are effectively impossible in normal operation
- If the caller passed valid text and the model is loaded, classification always produces a result
- This makes the API dramatically easier to use — no error handling boilerplate on every call
- If the Lumber devs prefer to keep error returns for safety, that's fine too — but the current internal code suggests it's not needed post-initialization

**Why `ClassifyBatch` matters:**
- Lumber's ONNX embedder already supports batched inference (`EmbedBatch`)
- For Heimdall, each monitoring cycle classifies 50–200 logs at once
- A single batched ONNX call is much faster than N individual calls due to GPU/CPU vectorization

---

## What Stays Internal

Everything else — connectors, output formats, pipeline orchestration, dedup, streaming — stays `internal/`. The library API is **just the engine**. The CLI continues to work exactly as it does today, composing the internal pieces.

```
lumber/
  lumber.go          ← NEW: public API (Lumber struct, New, Classify, ClassifyBatch, Event)
  internal/
    engine/          ← unchanged (engine, embedder, classifier, compactor, taxonomy)
    pipeline/        ← unchanged
    connector/       ← unchanged
    output/          ← unchanged
    config/          ← unchanged
    model/           ← unchanged (or re-export Event type from here)
  cmd/lumber/        ← unchanged
```

The public `Lumber` struct would internally compose the existing `engine.Engine` (which already does embed → classify → compact). The `New` constructor would replicate the setup logic from `cmd/lumber/main.go` lines 51–73 but with a simpler surface.

---

## Example: Heimdall Integration

```go
package main

import (
    "fmt"
    "github.com/kaminocorp/lumber"
)

func main() {
    // Load once at startup (~200ms)
    l, err := lumber.New("/app/models",
        lumber.WithConfidenceThreshold(0.5),
    )
    if err != nil {
        panic(err)
    }
    defer l.Close()

    // Classify a batch of logs (single ONNX inference call)
    events := l.ClassifyBatch([]string{
        `{"level":"error","msg":"connection refused","host":"db-primary","port":5432}`,
        `{"level":"info","msg":"GET /api/health 200 OK","duration_ms":12}`,
        `{"level":"error","msg":"panic: runtime error: index out of range [3] with length 2"}`,
        `{"level":"info","msg":"cron job invoice_sync completed in 340ms"}`,
    })

    for _, e := range events {
        fmt.Printf("%-14s %-22s %-7s (%.2f) %s\n",
            e.Type, e.Category, e.Severity, e.Confidence, e.Summary)
    }
    // Output:
    // ERROR          connection_failure     error   (0.87) connection refused
    // REQUEST        success                info    (0.92) GET /api/health 200 OK
    // ERROR          runtime_exception      error   (0.91) panic: runtime error: index out of range
    // SCHEDULED      cron_completed         info    (0.85) cron job invoice_sync completed in 340ms
}
```

---

## Implementation Estimate

The public API is a thin wrapper over existing internals. The work is roughly:

1. Create `lumber.go` in the package root
2. Define `Event` type (simplified `CanonicalEvent`)
3. Define `Lumber` struct wrapping `*engine.Engine`
4. `New()` does what `cmd/lumber/main.go:51-73` does: load embedder, build taxonomy, create classifier + compactor, compose engine
5. `Classify()` delegates to `engine.Process(RawLog{Raw: text})`
6. `ClassifyBatch()` delegates to `engine.ProcessBatch([]RawLog{...})`
7. Map `CanonicalEvent` → `Event` (drop Raw, Timestamp, Count)

No refactoring of internals needed. No changes to the CLI. Just a new file that composes existing pieces behind a public facade.

---

## Nice-to-Haves (Not Blocking)

These would be useful but aren't required for Heimdall's initial integration:

- **`Event.Path() string`** — convenience method returning `"ERROR.connection_failure"` (Type + "." + Category)
- **`lumber.DefaultTaxonomy() []string`** — list all taxonomy paths, useful for documentation/validation
- **Thread safety documentation** — confirm `Lumber` is safe for concurrent use (the ONNX session likely needs to document this)
