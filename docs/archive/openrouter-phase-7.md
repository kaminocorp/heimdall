# OpenRouter Integration — Phase 7 Completion

**Status:** ✅ Complete (post-implementation hardening pass)
**Date:** 2026-04-07
**Builds on:** [openrouter-phase-6.md](openrouter-phase-6.md)
**Validation:** `go build ./...` clean, `go vet ./...` clean, `go test ./...` all green

---

## Goal

A code-quality review of the Phase 1–6 OpenRouter integration surfaced two
issues worth fixing before shipping to production. Phase 7 is a focused
hardening pass that addresses both and documents them here so the fix history
travels with the feature rather than disappearing into `git log`.

Neither item is a correctness bug that users would have seen on day one —
they're the kind of latent issues that only bite under unusual conditions
(fallback paths that normally never fire, WebSocket clients that disconnect
mid-loop). Fixing them now is cheap; fixing them after the first mysterious
production incident is expensive.

---

## Issues addressed

### Fix 1 — Stale `defaultModelID` fallback

`backend/internal/agent/loop.go` still carried `defaultModelID = "claude-sonnet-4-5"`
from a pre-0.21 world. Every other place in the codebase that references a
default model ID uses `claude-sonnet-4-6`:

- `applications.go` (default-create, defaults response, `UpdateAppAgentConfig`
  default) — 3 sites
- `organizations.go` onboarding default
- `agent/models.go` curated `AnthropicModels` list
- `monitor.go` hardcoded test config
- Migrations `002` and `015` DB-level column default
- All test helpers and fixtures

Only `loop.go` and `provider_anthropic.go` still pointed at the older
`claude-sonnet-4-5` / `anthropic.ModelClaudeSonnet4_5` constant.

**Why it mattered, despite being dead in practice.** The agent loop populates
`params.Model` from `app_agent_config.model` (or the global `agent_config`
fallback), and both of those default to `claude-sonnet-4-6` at the DB and
handler layers. So the `loop.go` fallback would only ever fire if *both*
config reads failed — unlikely on a healthy system. But "unlikely" is the
wrong bar: if it ever did fire, the request would go to a model that isn't
even in the curated list, producing an error that's confusing to diagnose
because the ID doesn't appear anywhere visible to operators.

**Why removing the `provider_anthropic.go` fallback too.** The provider had
its own safety net: `if params.Model == "" { model = anthropic.ModelClaudeSonnet4_5 }`.
That's double defensive coding — the loop guarantees `params.Model` is
populated, so the provider fallback was both redundant and a second stale
constant to keep in sync. Deleted it and documented the contract ("params.Model
is always populated by the agent loop") in a comment, so future readers
understand why there's no fallback.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/loop.go` | `defaultModelID = "claude-sonnet-4-5"` → `"claude-sonnet-4-6"` |
| 2 | `backend/internal/agent/provider_anthropic.go` | Removed the `if params.Model == "" { ... ModelClaudeSonnet4_5 }` fallback; added a comment documenting the caller contract |

---

### Fix 2 — Producer goroutine leak on WebSocket disconnect

This is the real bug of the two. The setup:

1. `RunConversationStream` (`loop_stream.go`) spawns a goroutine that runs
   the tool-use loop and pushes `AgentEvent`s onto a buffered channel
   (`streamBufferSize = 16`).
2. The chat WebSocket handler (`chat.go`) consumes those events with
   `for ev := range events` and forwards them over the socket.
3. If the client disconnects mid-loop — closed tab, network hiccup, force
   quit — `wsjson.Write` returns an error and the handler returns, abandoning
   the channel without draining it.

**The leak.** With the consumer gone, the producer goroutine keeps emitting
events into the buffered channel. A chatty tool-use loop can comfortably
exceed 16 events (10 iterations × 2–3 tools × 2 events per tool = ~40–60 in
the worst case). Once the buffer fills, the next `events <- ev` send inside
`emitEvent` blocks **forever**. Context cancellation alone does not unblock
a blind channel send — that only happens if you explicitly select on
`ctx.Done()`.

The goroutine stays pinned until `provider.ChatCompletion` eventually returns
(either from ctx cancellation propagating through the HTTP client, or from
the model naturally finishing), but by then the next emit blocks again on
the still-full buffer. In practice the producer leaks until the process
restarts.

**Blast radius under load.** One leaked goroutine per disconnected chat
session × however many flaky mobile clients hit the service × however long
between deploys. Each leak holds onto the loop's stack (tool results,
message history, the provider response) — not huge, but cumulative, and
exactly the kind of slow memory growth that turns into an OOM at 3am two
weeks after deploy.

**The fix.** Make the channel sends ctx-aware. Both `emitEvent` (called from
inside the loop body for `tool_start` / `tool_result`) and the terminal
sends in `RunConversationStream` (for `error` / `message`) now `select` on
`ctx.Done()` alongside the channel send:

```go
// loop.go — inside the tool-use loop
func emitEvent(ctx context.Context, events chan<- AgentEvent, ev AgentEvent) {
    if events == nil {
        return // blocking RunConversation path — no-op, unchanged
    }
    select {
    case events <- ev:
    case <-ctx.Done():
    }
}
```

```go
// loop_stream.go — terminal send after the loop returns
var final AgentEvent
if err != nil {
    final = AgentEvent{Type: "error", Content: err.Error()}
} else {
    final = AgentEvent{Type: "message", Content: text}
}
select {
case ch <- final:
case <-ctx.Done():
}
```

When the WebSocket handler returns early (on a write error or read error),
`r.Context()` — which is the context passed into `RunConversationStream` —
is cancelled automatically by the `net/http` machinery. That cancellation
now propagates cleanly into the producer goroutine:

1. Any in-flight `provider.ChatCompletion` call aborts (its `http.Request`
   was built with the same ctx).
2. Any `emitEvent` blocked on a full buffer unblocks via the `ctx.Done()`
   branch.
3. The goroutine returns, `defer close(ch)` fires, the channel is collected.

Dropped events are fine: the consumer is gone, so there's nothing to deliver
to anyway. The alternative — draining the channel from a separate goroutine
in `chat.go` — is more code and gives the same observable behaviour.

**Why the blocking path is untouched.** When `RunConversation` calls
`runConversationCore` with `events == nil`, `emitEvent` short-circuits at
the nil check and never reaches the `select`. The blocking path remains
byte-for-byte equivalent to pre-Phase-6 behaviour — the ctx-awareness is a
pure addition for the streaming path.

| # | File | Change |
|---|------|--------|
| 1 | `backend/internal/agent/loop.go` | `emitEvent` signature gains `ctx context.Context`; body wraps the channel send in a `select { case events <- ev: case <-ctx.Done(): }`; three call sites updated to pass `ctx` |
| 2 | `backend/internal/agent/loop_stream.go` | Terminal `error` / `message` sends in the producer goroutine wrapped in the same `select` pattern; comment explains why |

---

## Files changed

### Modified (3)

| File | Change |
|---|---|
| `backend/internal/agent/loop.go` | Fix 1: `defaultModelID` → `"claude-sonnet-4-6"`. Fix 2: `emitEvent` becomes ctx-aware, three call sites updated |
| `backend/internal/agent/provider_anthropic.go` | Fix 1: removed stale `anthropic.ModelClaudeSonnet4_5` fallback + added contract comment |
| `backend/internal/agent/loop_stream.go` | Fix 2: terminal channel sends wrapped in `select` on `ctx.Done()` |

### Untouched (intentional)

- **`chat.go`.** The ctx-aware producer means the handler doesn't need a
  separate drain goroutine. The existing `for ev := range events` consumer
  keeps working — it exits on `close(ch)` after the producer returns.
- **`monitor.go`, `RunMonitoring`.** Monitoring still uses the blocking
  path (`ChatCompletion`, no events channel). The 0.27.0 rate limiter
  contract is unchanged — this is the "monitoring stays blocking" rule
  from Phase 1/6, and it still holds.
- **`RunConversation` (blocking path).** Passes `nil` for events; `emitEvent`
  short-circuits at the nil check. No behaviour change.
- **All existing tests.** The ctx parameter is additive — the nil-events
  path is byte-equivalent to before, so every test that exercises
  `RunConversation` (including the mock-anthropic-server tests in
  `loop_test.go`) passes unchanged.
- **`provider.go`, `provider_openrouter.go`.** Provider interface and
  implementations have nothing to do with loop-level event sinks.

---

## Validation

```
$ cd backend && go build ./... && go vet ./... && go test ./...
ok    github.com/hejijunhao/heimdall/backend/internal/agent       0.369s
ok    github.com/hejijunhao/heimdall/backend/internal/api/handlers (cached)
(... all green ...)
```

**Zero test edits required.** Neither fix changed externally observable
behaviour on any existing code path:

- Fix 1 only affects the dead fallback path that no test exercised.
- Fix 2 is additive — the nil-events branch is unchanged, and no existing
  test deliberately fills a 16-slot stream buffer to verify the leak, so
  none of them observe a difference.

A future test that deliberately stalls the consumer and verifies the
producer unblocks on ctx cancellation would be a nice-to-have, but it
requires more machinery than the fix itself — deferred as an optional
follow-up if the streaming work in `streaming-implementation.md` adds test
coverage for the loop/consumer boundary anyway.

---

## What stayed the same

- Phase 6's `AgentEvent` contract — no wire changes, no new event types.
  The chat handler consumes the exact same events in the exact same order.
- The frontend (`useAgent.ts`, `AgentChatPage.vue`) — completely untouched.
  A leaked producer goroutine is invisible to the browser.
- `streamBufferSize = 16` — kept as-is. Once the producer is ctx-aware,
  the buffer size is a throughput-vs-backpressure knob rather than a leak
  risk, and 16 is still comfortable for normal operation.
- The 0.27.0 monitoring rate limiter — interactive chat is not rate-limited
  and never was; this fix only touches the interactive path.
- Conversation persistence in `chat.go` — still happens once on the final
  `message` event, unchanged.

---

## Remaining minor items (not addressed this phase)

These were flagged in the code review but deliberately left for follow-ups,
to keep this phase laser-focused on the two real issues:

1. **`truncate` in `provider_openrouter.go:331`** slices bytes but the doc
   comment says "n runes." Affects only error-message formatting; cosmetic.
   Fix is either a comment update or `[]rune` conversion — one-liner, not
   worth its own phase.
2. **`"claude-sonnet-4-6"` literal duplicated across 7+ sites** (handlers,
   organizations, migrations, test helpers, monitor.go, the curated model
   list). Already flagged as a follow-up in both Phase 3 and Phase 4
   completion docs. Worth an `agent.DefaultModelID` package constant the
   next time this ID changes; not urgent because the coupling is all
   one-directional (everything points at the same literal).
3. **Pricing verification** for the six entries in `agent/models.go`
   against OpenRouter's live catalogue. Still flagged from Phase 4. The
   working baseline is from `openrouter-model-selection.md`; cross-check
   before announcing OpenRouter support to users.
4. **End-to-end smoke test** against a staging environment with a real
   `OPENROUTER_API_KEY`. The Phase 3 manual smoke-test recipe is still the
   canonical validation. Wire-format tests in `provider_openrouter_test.go`
   cover translation but not the full agent-loop round-trip through a real
   model.

---

## Follow-ups

### Token streaming (deferred from Phase 6)
Still the canonical next piece. See
[`docs/executing/streaming-implementation.md`](../executing/streaming-implementation.md).
The ctx-aware emit pattern from this phase carries over cleanly: when
`chunk` events join the stream, they'll use the same `emitEvent(ctx, ...)`
helper and inherit the same leak-free semantics. No refactor of Phase 7's
work required.

### Optional cleanup
The three residual items above (bytes-vs-runes comment, duplicated model
literal, pricing verification) can ride along with any future touch to
those files. None of them block production push.
