# OpenRouter Integration — Phase 1 Completion

**Status:** ✅ Complete (provider abstraction landed; streaming deferred)
**Date:** 2026-04-07
**Plan:** [openrouter-implementation.md](../executing/openrouter-implementation.md), Phase 1
**Validation:** `go build ./...` clean, `go vet ./...` clean, `go test ./...` all green

---

## Goal

Decouple the agent loop from `anthropic-sdk-go`'s types so that a second
provider (OpenRouter, Phase 2) can plug in without touching `loop.go`,
`monitor.go`, or `tools.go`. **Zero behaviour change** — this is a pure refactor.

---

## Scope decision: streaming deferred

The implementation plan (`openrouter-implementation.md`) bundled three
streaming-related work items into Phase 1:

- `Provider.ChatCompletionStream()` interface method
- `AnthropicProvider.ChatCompletionStream()` impl using `Messages.NewStreaming()`
- `Agent.RunConversationStream()` + `chat.go` rewrite to push SSE-style chunks
  to the WebSocket
- Frontend `useAgent.ts` chunk/tool_start/tool_result/done event handling

The selection plan (`openrouter-model-selection.md` §"Decisions") **explicitly
defers streaming** ("Not now, Option A") with the rationale that streaming
"touches every layer (agent loop, WebSocket handler, frontend composable) and
would double the blast radius of this change."

**This Phase 1 commit follows the selection plan's call.** Streaming is left
for a dedicated follow-up. The `Provider` interface is intentionally minimal
today (one method) and is documented to grow with `ChatCompletionStream` when
the streaming work begins. Doing it this way means:

- Phase 1 is a true zero-behaviour-change refactor — easy to review, easy to
  roll back.
- Phase 2 (OpenRouter provider) and Phase 3 (DB migration) can land on top of
  this without taking on any frontend or WebSocket risk.
- The streaming work, when it happens, is a self-contained vertical slice that
  touches `provider.go` (interface), both providers, `chat.go`, and
  `useAgent.ts` together — which is the right granularity for that change.

What this completion doc covers: **Tasks 1.1, 1.2 (blocking only), 1.3, 1.4
(blocking only), 1.5.** Task 1.6 (`chat.go` streaming) is **not** part of this
phase.

---

## Files changed

### New (2)

| File | Purpose |
|---|---|
| `backend/internal/agent/provider.go` | Provider interface + neutral domain types (`ChatParams`, `ChatMessage`, `ContentBlock`, `ToolCall`, `ToolResult`, `ToolDef`, `ChatResponse`, `Usage`, `StopReason`) |
| `backend/internal/agent/provider_anthropic.go` | `AnthropicProvider` — only file in the package that imports `anthropic-sdk-go` |

### Modified (4)

| File | Change |
|---|---|
| `backend/internal/agent/agent.go` | `client *anthropic.Client` field replaced with `providers map[string]Provider`. New `providerFor(name string) Provider` resolver. Anthropic SDK imports removed. |
| `backend/internal/agent/loop.go` | `RunConversation` and `RunMonitoring` now build `[]ChatMessage` and call `provider.ChatCompletion()`. `extractText` takes `*ChatResponse`. All `anthropic.*` types removed. Constants `defaultModelID = "claude-sonnet-4-5"` and `defaultMaxTokens = 4096` extracted. |
| `backend/internal/agent/tools.go` | `ToolRegistry()` now returns `[]ToolDef` instead of `[]anthropic.ToolUnionParam`. Anthropic SDK import removed. `Dispatch` is unchanged. |
| `backend/internal/agent/loop_test.go` | `newTestAgent` constructs `AnthropicProvider` via the new `NewAnthropicProviderWithClient` helper and registers it under `providers[defaultProviderName]`. Tests still hit the same mock HTTP server with no other changes. |

### Untouched (intentional)

| File | Why |
|---|---|
| `backend/internal/agent/monitor.go` | Doesn't import anthropic; only calls `RunMonitoring` (which we refactored internally). The monitoring loop, rate limiter, semaphore, cursor handling all remain identical. |
| `backend/internal/api/handlers/chat.go` | Streaming not in scope. The handler still calls `s.Agent.RunConversation(...)` with the same signature and gets back the same `(string, error)`. |
| `backend/internal/agent/{tools_db,tools_logs,tools_codebase,tools_memory}.go` | Tool implementations are unchanged — they take `map[string]any` and return `(string, error)` regardless of provider. |
| `backend/internal/agent/{classifier*,emit,extract,prompt,severity_gate,message}.go` | None of these reference SDK types. |

---

## Architecture: how the pieces fit together now

```
┌────────────────────────────────────────────────────────┐
│  agent loop  (loop.go RunConversation / RunMonitoring) │
│                                                        │
│  builds:  ChatParams { Messages, Tools, ... }          │
│  calls:   provider.ChatCompletion(ctx, params)         │
│  reads:   ChatResponse { Blocks, StopReason }          │
└──────────────────────┬─────────────────────────────────┘
                       │  speaks ONLY neutral types
                       ▼
              ┌────────────────┐
              │   Provider     │  (interface — provider.go)
              │   interface    │
              └────────┬───────┘
                       │
        ┌──────────────┴──────────────┐
        ▼                             ▼
┌───────────────────┐     ┌────────────────────────┐
│ AnthropicProvider │     │ OpenRouterProvider     │
│ (provider_        │     │ (provider_openrouter.go│
│  anthropic.go)    │     │  — Phase 2)            │
│                   │     │                        │
│ wraps             │     │ wraps                  │
│ anthropic-sdk-go  │     │ raw HTTP to            │
│                   │     │ openrouter.ai/api/v1   │
└───────────────────┘     └────────────────────────┘
```

**The contract:** every type that crosses the `Provider` interface boundary
lives in `provider.go` and has zero SDK dependencies. The agent loop has no
idea which provider it's talking to.

---

## How the translation works (AnthropicProvider)

`provider_anthropic.go` is a pure translation layer between the neutral types
and the SDK types. The mapping is mechanical and one-to-one:

| Neutral type | Anthropic SDK type |
|---|---|
| `ChatMessage{Role:"user", TextContent:"..."}` | `anthropic.NewUserMessage(anthropic.NewTextBlock(...))` |
| `ChatMessage{Role:"user", ToolResults:[...]}` | `anthropic.NewUserMessage(NewToolResultBlock(...), ...)` |
| `ChatMessage{Role:"assistant", Blocks:[...]}` | `anthropic.NewAssistantMessage(...)` with `TextBlock`/`ToolUseBlockParam` |
| `ToolDef{Name, Description, Parameters, Required}` | `anthropic.ToolUnionParam{OfTool: &anthropic.ToolParam{...}}` |
| `anthropic.StopReasonEndTurn` | `StopReasonEndTurn` |
| `anthropic.StopReasonToolUse` | `StopReasonToolUse` |
| `anthropic.TextBlock` (response) | `ContentBlock{Type:"text", Text:...}` |
| `anthropic.ToolUseBlock` (response) | `ContentBlock{Type:"tool_use", ToolCall:&ToolCall{ID,Name,Input}}` |

**Two small details worth noting:**

1. **`ToolUseBlock.Input` is preserved as `json.RawMessage`.** Both
   `ToolUseBlock` (response) and `ToolUseBlockParam` (request) accept the
   same shape, so we can round-trip the assistant's tool_use blocks back to
   the model on the next iteration without re-encoding. This matches the
   pre-refactor behaviour exactly.

2. **Unknown stop reasons collapse to `EndTurn`.** The pre-refactor loop
   already had `return extractText(resp), nil` as its final fallthrough — it
   never branched on `max_tokens`/`stop_sequence`/etc. We preserve that by
   mapping any unknown SDK stop reason to `StopReasonEndTurn` in the
   translation layer rather than adding a third enum value that nothing reads.

---

## Provider resolution

```go
// agent.go
const defaultProviderName = "anthropic"

func (a *Agent) providerFor(name string) Provider {
    if name != "" {
        if p, ok := a.providers[name]; ok {
            return p
        }
    }
    return a.providers[defaultProviderName]
}
```

Both `RunConversation` and `RunMonitoring` currently pass
`defaultProviderName` because there is no `provider` column on
`app_agent_config` yet. **Phase 3 will add that column** and replace the
hardcoded `defaultProviderName` with `appCfg.Provider`. The resolver is
designed to fall back to `"anthropic"` for any empty/unknown name, which gives
us free backwards-compat for legacy rows once the column lands.

---

## Test strategy

The existing tests in `loop_test.go`, `tools_test.go`, `monitor_test.go`,
`extract_test.go`, `severity_gate_test.go`, and `classifier_test.go` are
**unchanged in intent** — same scenarios, same assertions, same mock HTTP
server pattern. The only edit is the constructor in `newTestAgent`:

```diff
- queries := db.New(&stubDBTX{})
- return &Agent{
-     queries:    queries,
-     client:     &client,
-     config:     ...,
-     classifier: &PassthroughClassifier{},
- }
+ queries := db.New(&stubDBTX{})
+ return &Agent{
+     queries: queries,
+     providers: map[string]Provider{
+         defaultProviderName: NewAnthropicProviderWithClient(&client),
+     },
+     config:     ...,
+     classifier: &PassthroughClassifier{},
+ }
```

`NewAnthropicProviderWithClient` exists for exactly this purpose — it lets
tests inject an `anthropic.Client` already wired to a `httptest.Server`,
without rebuilding the constructor's API key path.

`newToolTestAgent` (used by `tools_test.go` and most of `monitor_test.go`)
doesn't make API calls — it only exercises `Dispatch` and the monitoring
goroutine lifecycle — so it leaves `providers` nil, which is harmless because
those tests never call `providerFor`.

**Validation results:**

- `go build ./...` — clean
- `go vet ./...` — clean
- `go test ./internal/agent/...` — pass (all tests, no changes to test count)
- `go test ./...` — pass across the whole backend

The Phase 1 validation gates from `openrouter-implementation.md` are met:

- [x] All existing backend tests pass
- [x] Agent chat works identically (same `RunConversation` signature, same
      blocking call semantics — `chat.go` was not touched)
- [x] Agent monitoring works identically (rate limiter, semaphore, cursor,
      classifier pipeline all unchanged)
- [x] No `anthropic-sdk-go` types leak outside `provider_anthropic.go`
- [x] `tools.go` returns `[]ToolDef`, not `[]anthropic.ToolUnionParam`

The remaining gate from the implementation doc — "Chat shows tokens appearing
incrementally" — is explicitly out of scope and tracked in the follow-ups
below.

---

## Why this matters (beyond OpenRouter)

The provider abstraction is load-bearing for more than just adding OpenRouter:

1. **It localises SDK risk.** The Anthropic SDK has shipped breaking changes
   between minor versions before. Now there is exactly one file that needs
   updating when that happens (`provider_anthropic.go`), instead of three
   (`agent.go`, `loop.go`, `tools.go`).
2. **It makes mocking trivial.** Future tests can implement `Provider` with
   ~10 lines of fake code instead of standing up an `httptest.Server` and
   composing JSON-shaped Anthropic responses by hand. The existing mock-server
   tests are kept as-is because they were already working — but new tests
   should prefer fakes.
3. **It is a clean seam for adding cross-cutting concerns.** Retries, request
   logging, latency metrics, and (eventually) per-app cost tracking can wrap
   `Provider` as decorators without touching the loop.

---

## What stayed exactly the same (deliberate non-changes)

- The 0.27.0 rate limiter (`a.limiter.Wait(ctx)`) on the monitoring path —
  still wraps each `RunMonitoring` invocation as one atomic unit. The
  `RunMonitoring` doc-comment now also explicitly warns future contributors
  not to switch this path to streaming.
- The 10-iteration cap on the tool-use loop.
- Tool dispatch via `Agent.Dispatch`, keyed on tool name.
- Conversation persistence via `chat.go` → `persistMessages`.
- `EmitLog` calls for `tool_call` / `tool_result` / `observation` agent_log
  entries — every emit site is preserved with identical arguments.
- The default model fallback string (`claude-sonnet-4-5`) — moved to a
  constant `defaultModelID` but the value is the same one
  `anthropic.ModelClaudeSonnet4_5` resolves to.

---

## Follow-ups (in dependency order)

### Immediate next: Phase 2 — OpenRouterProvider
- New file `provider_openrouter.go` implementing `Provider` against
  `https://openrouter.ai/api/v1/chat/completions` (raw HTTP, no second SDK)
- Tool schema translation: `ToolDef` → OpenAI `function.parameters` shape
- Stop reason mapping: `finish_reason: "tool_calls"` → `StopReasonToolUse`
- Optional `OPENROUTER_API_KEY` in `config.go`; provider only registered when set

### Phase 3 — Migration `022_agent_config_provider`
- `ALTER TABLE app_agent_config ADD COLUMN provider TEXT NOT NULL DEFAULT 'anthropic'`
- Replace the `providerFor(defaultProviderName)` call sites in `RunConversation`
  and `RunMonitoring` with `providerFor(appCfg.Provider)` — that's the only
  loop.go change Phase 3 needs

### Streaming follow-up (separate from Phases 2–6)
This is the work that was sliced out of Phase 1:

- Add `ChatCompletionStream(ctx, params) (<-chan StreamEvent, error)` to the
  `Provider` interface
- Implement on both `AnthropicProvider` (via `Messages.NewStreaming()`) and
  `OpenRouterProvider` (via SSE on the `/chat/completions` endpoint with
  `"stream": true`)
- Add `Agent.RunConversationStream(...)` that runs the tool-use loop but emits
  `AgentEvent`s on a channel as text/tool calls arrive
- Update `chat.go` to consume the channel and push `chunk` / `tool_start` /
  `tool_result` / `done` messages over the WebSocket
- Update `useAgent.ts` to handle the new event types and append to an
  in-progress message
- **Critical constraint, repeated:** `RunMonitoring` MUST continue to call
  `ChatCompletion` (blocking), never `ChatCompletionStream`. The 0.27.0 rate
  limiter contract depends on it. The doc-comment in `loop.go` says so; honour
  it.

### Optional cleanups (low priority)
- Once Phase 3 lands and `providerFor(appCfg.Provider)` is in place, the
  `defaultProviderName` constant in `loop.go` becomes unused at the call sites
  and only lives in `agent.go`. Move it back to `agent.go` only.
- Consider adding a `Provider` fake (`testdata/fake_provider.go`) for future
  agent tests that don't need to exercise SDK wire-format translation.
