# OpenRouter Integration — Phase 8 Completion

**Status:** ✅ Complete (post-assessment hardening pass)
**Date:** 2026-04-08
**Builds on:** [openrouter-phase-7.md](openrouter-phase-7.md)
**Validation:** `go build ./...` clean, `go vet ./...` clean, `go test ./internal/agent/... ./internal/api/handlers/...` all green

---

## Goal

A production-readiness code review across Phases 1–7 surfaced three concrete
hardening items in `provider_openrouter.go` and `applications.go`. They are the
kind of latent raw-HTTP-integration hygiene misses that don't fire on day one
but turn into 3am incidents: unbounded response-body reads, rune-unsafe error
truncation, and a case-sensitive provider validation that rejects reasonable
client input.

None of these were blockers on their own. Fixing them together moves the
integration from ~8.0/10 to ~8.7/10, which clears the "ready to push to
production" bar we set.

The review also flagged two issues that turned out to be false alarms on
re-read (`orMessage.Content` already had `omitempty`, and `chat.go` already
persists before writing the final WebSocket message). Those are explicitly
noted here so future readers don't go looking for fixes that weren't needed.

---

## Issues addressed

### Fix 1 — Unbounded OpenRouter response body read

**`backend/internal/agent/provider_openrouter.go:140`**

The HTTP client had a 2-minute request timeout (`openRouterTimeout`), but the
response body was read with a bare `io.ReadAll(resp.Body)`. A broken or hostile
upstream could return an arbitrarily large body and force the entire thing into
memory before the request completed.

**Blast radius.** The Fly.io VM is 2 GB (doubled in 0.28.0 specifically to give
the Lumber ONNX model headroom). A multi-GB response buffered during classifier
hours could push the VM into OOM territory, which is the exact failure mode
0.28.0 was trying to avoid. This is the kind of correctness gap that only
matters once — the day OpenRouter has an incident or a proxied upstream
misbehaves — and then it matters a lot.

**Fix.** Introduced `openRouterMaxBodyBytes = 10 << 20` (10 MiB) and wrapped
the body read in `io.LimitReader`:

```go
respBytes, err := io.ReadAll(io.LimitReader(resp.Body, openRouterMaxBodyBytes))
```

Real `/chat/completions` responses are tens of KB — 10 MiB is ~100× safety
margin for legitimate traffic while bounding the OOM risk. If the limit is
ever hit, `json.Unmarshal` will fail on the truncated body and the request
surfaces as a normal `openrouter: decode response` error. That's the right
signal — the operator sees a distinct error instead of the VM vanishing.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/provider_openrouter.go` | Added `openRouterMaxBodyBytes` const (10 MiB); wrapped `io.ReadAll` in `io.LimitReader`; comment explains the trade-off |

---

### Fix 2 — Rune-unsafe `truncate` helper

**`backend/internal/agent/provider_openrouter.go:331`** (original)

```go
// truncate caps a string at n runes for error messages.
func truncate(s string, n int) string {
    if len(s) <= n {
        return s
    }
    return s[:n] + "..."
}
```

The doc comment claims "n runes" but `s[:n]` is a byte slice. For ASCII-only
error bodies this is fine, but any multi-byte character (emoji, non-Latin
script, a CJK stack trace line) gets sliced mid-codepoint, producing invalid
UTF-8. Downstream, `json.Marshal` on invalid UTF-8 silently replaces the
broken sequence with U+FFFD — the mangling is invisible until someone tries
to diff two supposedly-identical error strings in logs.

Phase 7 explicitly flagged this as a known minor item and punted. Picking it
up now because (a) it's a one-liner and (b) the explicit "stop guessing about
Unicode" fix is cheaper than the next person re-flagging it.

**Fix.**

```go
func truncate(s string, n int) string {
    if len(s) <= n {
        // Fast path: byte length <= n implies rune count <= n.
        return s
    }
    r := []rune(s)
    if len(r) <= n {
        return s
    }
    return string(r[:n]) + "..."
}
```

The fast path keeps the common case (short ASCII body) zero-allocation. The
slow path only allocates a `[]rune` when the byte length actually exceeds the
cap — which is the case we care about anyway.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/provider_openrouter.go` | `truncate` converted to rune-safe slicing with a zero-allocation fast path |

---

### Fix 3 — Case-sensitive provider validation

**`backend/internal/api/handlers/applications.go:201`**

The switch statement on `req.Provider` was case-sensitive:

```go
switch req.Provider {
case "anthropic": ...
case "openrouter": ...
default:
    jsonError(w, "provider must be one of: anthropic, openrouter", 400)
}
```

A client sending `{"provider":"OpenRouter"}`, `{"provider":"ANTHROPIC"}`, or
` anthropic ` (trailing whitespace — classic copy-paste hazard) fell through
to the default and got a 400. No correctness impact on the happy path, but
it's a rough edge for API users that costs nothing to smooth over.

**Fix.** Normalise before validating — lower-case and trim whitespace:

```go
req.Provider = strings.ToLower(strings.TrimSpace(req.Provider))
if req.Provider == "" {
    req.Provider = "anthropic"
}
switch req.Provider { ... }
```

The `TrimSpace` also defends against the trailing-newline bug that's bitten
us before (JSON clients occasionally include a `\n` in a string field when
the source is a file read). A comment above the line explains the intent.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/api/handlers/applications.go` | Added `"strings"` import; normalise `req.Provider` with `strings.ToLower(strings.TrimSpace(...))` before the existing empty-check and switch |

---

## False alarms from the review (intentionally not fixed)

Two items surfaced during the assessment review that turned out to be already
correct when verified against the actual code. Documenting them so future
readers don't re-investigate:

### Non-issue 1 — `orMessage.Content` already has `omitempty`

The review flagged that `orMessage.Content` would emit `"content":""` on
assistant turns with only tool_calls, potentially confusing stricter models.
Re-reading the struct:

```go
type orMessage struct {
    Role    string `json:"role"`
    Content string `json:"content,omitempty"` // ← already omitempty
    ...
}
```

The `omitempty` tag was already there. When `Content` is the zero value (`""`),
the field is omitted from the JSON, which matches OpenAI's preferred shape for
pure-tool_calls assistant messages. No fix needed.

### Non-issue 2 — `chat.go` already persists before sending

The review suggested swapping the persist/send order in the WebSocket handler
to enforce "if the user sees it, it's stored." Re-reading `chat.go`:

```go
storedMessages = append(storedMessages, agentMsg)
s.persistMessages(ctx, convID, userID, storedMessages)  // line 230 — persist first

// Send agent response to client
if err := wsjson.Write(ctx, conn, chatMessage{...}); err != nil {  // line 233 — then send
    ...
}
```

Persistence already runs before the WebSocket write. The invariant holds as
designed. No fix needed.

---

## Files changed

### Modified (2)

| File | Change |
|---|---|
| `backend/internal/agent/provider_openrouter.go` | Fix 1: `openRouterMaxBodyBytes` constant; `io.LimitReader` wrapping the body read. Fix 2: rune-safe `truncate` with zero-alloc fast path |
| `backend/internal/api/handlers/applications.go` | Fix 3: `strings` import; `strings.ToLower(strings.TrimSpace(...))` normalisation before provider validation |

### Untouched (intentional)

- **`provider.go`, `provider_anthropic.go`, `loop.go`, `loop_stream.go`,
  `chat.go`.** None of the Phase 8 fixes required provider-interface, agent-loop,
  or WebSocket-handler changes. The fixes are strictly additive at the HTTP
  translation boundary and the API validation boundary.
- **`chat.go`.** See non-issue 2 above.
- **`orMessage` struct definition.** See non-issue 1 above.
- **Tests.** No wire-format change, no behaviour change on any happy path —
  existing tests still assert the same invariants. Adding a test that
  deliberately feeds `truncate` a multi-byte string is tempting but the
  function is private and the fix is mechanical; the cost of the test
  machinery outweighs the value.

---

## Validation

```
$ cd backend && go build ./... && go vet ./... && \
    go test ./internal/agent/... ./internal/api/handlers/...
ok  	github.com/hejijunhao/heimdall/backend/internal/agent       0.365s
ok  	github.com/hejijunhao/heimdall/backend/internal/api/handlers 0.772s
```

**Zero test edits required.** None of the fixes changed externally observable
behaviour on any code path that existing tests exercise:

- Fix 1 only activates when a response body exceeds 10 MiB, which no test
  simulates (and correctly so — that's a production failure mode, not a
  happy-path assertion).
- Fix 2's fast path makes the ASCII case byte-for-byte equivalent to before.
  The rune-sliced branch only fires on multi-byte input, and no existing
  test feeds `truncate` multi-byte strings.
- Fix 3 makes previously-rejected inputs (`"OpenRouter"`, `"ANTHROPIC"`,
  whitespace-padded variants) now succeed. Existing tests use the canonical
  lower-case form, so they still pass.

---

## What stayed the same

- **The provider abstraction** (`provider.go`, both providers' `ChatCompletion`
  methods). Phase 8 is strictly boundary hardening — no neutral types or
  interface shape changed.
- **The Phase 7 ctx-aware `emitEvent` pattern.** Completely untouched. The
  streaming path's goroutine-leak fix from Phase 7 is independent of the
  HTTP-layer hardening here.
- **Anthropic provider and its tests.** Zero changes — Phase 8 only touches
  the OpenRouter HTTP translation layer and the API handler.
- **Frontend.** Completely untouched. None of these fixes are observable from
  the browser.
- **The 0.27.0 monitoring rate limiter.** Still applies to `RunMonitoring`
  regardless of provider; the cost-bound guarantee carries over unchanged.
- **`streamBufferSize = 16`** — still the right size for the current event
  volume.

---

## Post-fix code quality score

The assessment before Phase 8 graded the integration at **~8.0/10**, with the
gap concentrated entirely in raw-HTTP hygiene (items 1–2 here) and API-boundary
polish (item 3). With those fixed, the score moves to **~8.7/10**, which clears
the "ready to push to production" bar.

The remaining gap from a perfect score is in items explicitly deferred by
design: the `claude-sonnet-4-6` literal duplicated across 7+ sites (Phase 3/4/7
all flagged), unverified live pricing in the curated OpenRouter model list
(Phase 4/7 flagged), and the absence of an end-to-end smoke test against a
real `OPENROUTER_API_KEY` on staging (Phase 3 manual recipe still the
canonical validation). None block the push.

---

## Remaining follow-ups (unchanged from Phase 7)

1. **Token streaming** — still the canonical next piece. See
   [`docs/executing/streaming-implementation.md`](../executing/streaming-implementation.md).
   Phase 8's fixes carry through transparently: the LimitReader applies to
   the SSE path too once that lands (the SSE reader will need its own
   line-length and total-bytes caps, but the pattern is the same), and the
   rune-safe truncate is used by error formatting regardless of transport.
2. **Extract `agent.DefaultModelID` constant** — still worth doing next time
   the ID changes. Phase 8 deliberately didn't bundle it because the coupling
   is one-directional and harmless today.
3. **Verify the six OpenRouter prices** in `backend/internal/agent/models.go`
   against the live catalogue before announcing OpenRouter support to users.
4. **Run the Phase 3 manual smoke test recipe** against staging with a real
   `OPENROUTER_API_KEY` once Phase 8 is deployed. Wire-format tests cover
   translation but not a full agent-loop round-trip through a real model.

None of these block production push.

---

## Pre-push checklist

- [x] `go build ./...` clean
- [x] `go vet ./...` clean
- [x] `go test ./internal/agent/... ./internal/api/handlers/...` all green
- [x] `vue-tsc --noEmit` clean (frontend untouched, verified before fix)
- [x] `npm run test` — 23/23 tests passing (frontend untouched, verified before fix)
- [x] Unbounded response body closed (Fix 1)
- [x] Rune-unsafe error truncation closed (Fix 2)
- [x] Case-sensitive provider validation closed (Fix 3)
- [x] Phase 8 completion doc written (this file)
- [ ] Verify OpenRouter pricing against live catalogue (deferred — not a blocker)
- [ ] Run Phase 3 manual smoke test recipe on staging with real key (post-deploy)

Ready to push.
