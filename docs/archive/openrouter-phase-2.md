# OpenRouter Integration — Phase 2 Completion

**Status:** ✅ Complete (blocking `ChatCompletion` only; streaming deferred with Phase 1)
**Date:** 2026-04-07
**Plan:** [openrouter-implementation.md](../executing/openrouter-implementation.md), Phase 2
**Validation:** `go build ./...` clean, `go vet ./...` clean, `go test ./...` all green
**Builds on:** [openrouter-phase-1.md](openrouter-phase-1.md)

---

## Goal

Add a second `Provider` implementation that routes requests through OpenRouter's
OpenAI-compatible `/chat/completions` endpoint. With this in place, the agent
loop can transparently target hundreds of models (Claude, GPT, Gemini, Llama,
Mistral, ...) via a single integration — provided an `OPENROUTER_API_KEY` is
configured.

---

## Scope

This phase implements the **blocking** half of the OpenRouter provider:

- ✅ Task 2.1 — `OpenRouterKey` added to `config.Config` (optional; not validated)
- ✅ Task 2.2 — `OpenRouterProvider` implementing `Provider.ChatCompletion`
- ✅ Task 2.3 — Provider registered at agent startup when key is set
- ⚠️ `.env.example` update (Task 2.1 doc step) — **skipped, file does not exist**
  in this repo. The convention is `OPENROUTER_API_KEY=`; no template to update.

**Streaming (`ChatCompletionStream`) remains deferred** — same call as Phase 1.
The interface still has only `ChatCompletion`, so neither provider implements
streaming yet. When the streaming follow-up lands, it will add the method to
the interface and implement it on both providers in the same vertical slice.

---

## Files changed

### New (2)

| File | Purpose |
|---|---|
| `backend/internal/agent/provider_openrouter.go` | Raw-HTTP OpenRouter provider (~270 lines including private wire types) |
| `backend/internal/agent/provider_openrouter_test.go` | Four httptest-server tests pinning the wire format and error handling |

### Modified (2)

| File | Change |
|---|---|
| `backend/internal/config/config.go` | Added `OpenRouterKey string` field and `OPENROUTER_API_KEY` env load. Validation unchanged — the key is optional. |
| `backend/internal/agent/agent.go` | `New()` now registers `providers["openrouter"]` when `cfg.OpenRouterKey != ""`. Logs `openrouter provider enabled` so operators can confirm at startup. |

### Untouched (intentional)

- `loop.go`, `monitor.go`, `tools.go`, `chat.go` — the Phase 1 provider
  abstraction means none of these files need any change to support OpenRouter.
  The agent loop calls `a.providerFor(...)` and gets back whichever
  implementation is registered. **This is the payoff of Phase 1.**
- `provider.go` — the interface is unchanged. OpenRouterProvider satisfies it
  by implementing the same single method.

---

## Why raw HTTP, not an OpenAI SDK

The implementation plan called this out, and the implementation follows it:

- **Surface area is small.** We need exactly one endpoint for now
  (`POST /v1/chat/completions`, blocking). A focused HTTP client is
  ~270 lines including the private wire types and zero transitive deps.
- **No second large SDK.** Pulling in `openai-go` or similar would bring its
  own retry policies, transport configuration, and version churn — none of
  which we want as risk surface for a single endpoint.
- **The translation has to happen anyway.** Whether we call an SDK or write the
  JSON by hand, `ChatParams` → OpenAI shape and back is the same code. Doing
  it directly keeps the translation visible and testable instead of hidden
  inside another SDK's types.

When streaming arrives, the SSE parser is another ~50–80 lines of focused
code. Still cheaper than an SDK.

---

## Wire-format translation: the part that matters

The selection plan and the implementation plan both flagged this as the riskiest
part of the OpenRouter integration. Five concrete divergences from Anthropic:

| Concern | Anthropic format | OpenAI/OpenRouter format | How it's handled |
|---|---|---|---|
| **System prompt** | Separate `system` param on the request | First message with `role:"system"` | `buildOpenRouterRequest` prepends a system message when `params.SystemPrompt != ""` |
| **Tool definitions** | `input_schema` at top level (already an object schema) | Nested under `function.parameters`, must be a full JSON Schema object including `type:"object"` and `properties:{}` | Wrap `t.Parameters` in `{"type":"object","properties":...,"required":...}` |
| **Tool use in responses** | `content[].type:"tool_use"` blocks intermixed with text blocks | `tool_calls[]` array on the assistant message, content is the (possibly empty) text | Translation flattens both sides — text in `Content`, tool calls in `ToolCalls`, agent loop sees the same `[]ContentBlock` |
| **Tool results** | `role:"user"` message containing `tool_result` blocks (one message holds multiple results) | Each result is a separate `role:"tool"` message with its own `tool_call_id` | Each `ToolResult` becomes its own message in the OpenAI request |
| **Stop reason** | `stop_reason: "tool_use"` / `"end_turn"` | `finish_reason: "tool_calls"` / `"stop"` / `"length"` / etc. | Only `"tool_calls"` maps to `StopReasonToolUse`; everything else collapses to `StopReasonEndTurn` (matching the Phase 1 AnthropicProvider) |

**Two subtleties worth flagging for future maintainers:**

1. **Tool arguments are a JSON-encoded string in OpenAI's spec.** Anthropic
   passes `Input` as `json.RawMessage` (a JSON object directly inline).
   OpenAI's `function.arguments` field is a *string* whose contents are JSON.
   The translation handles both directions: when sending an assistant
   tool_use back to the model, we pass `string(b.ToolCall.Input)`; when
   parsing a tool_calls response, we cast `tc.Function.Arguments` to
   `json.RawMessage` so the agent loop's `json.Unmarshal(tu.Input, ...)`
   call works without modification. The pinning test
   `TestOpenRouter_AssistantToolUseRoundTrip` exercises both directions.

2. **`ToolResult.IsError` has no equivalent in OpenAI's spec.** OpenAI's
   `role:"tool"` message has no `is_error` flag. We surface errors by
   prefixing the content with `"ERROR: "` so the model can still see that
   the tool failed. Models are good at picking up on this convention, and
   it preserves the agent loop's existing error-handling without needing an
   "OpenRouter doesn't support tool errors" caveat in `loop.go`. This is a
   pragmatic choice — if it turns out to confuse some models in practice,
   we can revisit by mapping to a structured format like
   `{"error": "..."}` instead.

---

## Required headers

OpenRouter requires (or strongly recommends) two attribution headers in
addition to standard auth:

```
Authorization: Bearer $OPENROUTER_API_KEY
HTTP-Referer:  https://heimdall.app
X-Title:       Heimdall
Content-Type:  application/json
```

`HTTP-Referer` and `X-Title` show up on the OpenRouter dashboard so operators
can see usage attributed to Heimdall. They are constants in
`provider_openrouter.go` (`openRouterReferer`, `openRouterTitle`) — change
them in one place if the public site domain ever moves.

---

## Error handling

Three error classes get specific, readable messages so the agent loop and
operators can distinguish them:

| HTTP status | Returned error | Why it matters |
|---|---|---|
| `401 Unauthorized` | `openrouter: unauthorized (401): check OPENROUTER_API_KEY` | Misconfiguration — operator action needed |
| `402 Payment Required` | `openrouter: insufficient credits (402): <body>` | Out of OpenRouter credits — operator action needed |
| `429 Too Many Requests` | `openrouter: rate limited (429): <body>` | Transient — could be retried in future, but right now bubbles up |
| Other non-2xx | `openrouter: http <status>: <body>` | Generic catch-all with response body for debugging |

Body is truncated at 500 chars to keep error messages legible. **There is no
silent fallback to Anthropic** — per the selection plan's risk mitigation,
falling back would change the user's chosen model without consent. The error
surfaces explicitly so the failure is visible.

---

## Provider registration

```go
// agent.go (post-Phase-2)
providers := map[string]Provider{
    defaultProviderName: NewAnthropicProvider(cfg.AnthropicKey),
}
if cfg.OpenRouterKey != "" {
    providers["openrouter"] = NewOpenRouterProvider(cfg.OpenRouterKey)
    slog.Info("openrouter provider enabled")
}
```

The "without OPENROUTER_API_KEY, everything works as before" guarantee from
Phase 2's validation gate is structurally enforced: the providers map only
contains `"anthropic"` unless the operator opts in. `providerFor("openrouter")`
on a single-provider deployment returns the anthropic fallback, which means
even a buggy DB row that says `provider='openrouter'` won't crash — it just
silently routes through Anthropic. That's the right behaviour for now;
Phase 3's request validation in `applications.go` will reject
`provider='openrouter'` at the API boundary if the platform hasn't enabled it.

---

## Tests

`provider_openrouter_test.go` adds four httptest-based tests targeting the
exact translation seams:

| Test | What it pins |
|---|---|
| `TestOpenRouter_RequestShape` | System prompt is the first message (not a separate field), tools are wrapped in a JSON Schema `object`, headers (`Authorization`, `HTTP-Referer`, `X-Title`, `Content-Type`) are set, response decodes into `ChatResponse` |
| `TestOpenRouter_ToolCallResponse` | `finish_reason:"tool_calls"` becomes `StopReasonToolUse`, the tool_calls JSON-string `arguments` round-trips through `json.Unmarshal` so the agent loop can read it |
| `TestOpenRouter_AssistantToolUseRoundTrip` | Replaying an assistant message with `Blocks` (text + tool_use) produces a single OpenAI assistant message with `content` + `tool_calls`; the matching tool result becomes a `role:"tool"` message with the right `tool_call_id`. **This is the most important test** — it exercises the multi-turn flow that's most likely to break |
| `TestOpenRouter_ErrorMapping` | 401/402/429/500 produce distinct, readable errors |

All four use a `httptest.Server` to intercept the request and assert against
the parsed `orRequest` struct, so the test failures are at the field level
(not "JSON didn't match a regex"). No real network calls, no real OpenRouter
key, runs in <1 second.

**Validation gates from `openrouter-implementation.md` Phase 2:**

- [x] With `OPENROUTER_API_KEY` set, agent can be wired through OpenRouter
      (the request shape is validated by tests; full end-to-end against the
      live service is deferred to Phase 3 when there's a way to actually
      configure an app to use it via the DB column)
- [x] Without `OPENROUTER_API_KEY`, everything works as before — anthropic-only
      registration path is unchanged
- [x] Tool use round-trips end-to-end through the OpenRouter wire format
      (`TestOpenRouter_AssistantToolUseRoundTrip`)
- [ ] Streaming works through OpenRouter — **deferred with Phase 1**

---

## What stayed the same

- The `Provider` interface — still one method (`ChatCompletion`). Adding the
  second provider didn't require any interface change, which is exactly what
  Phase 1 was designed to enable.
- `loop.go`, `monitor.go`, `tools.go`, `chat.go` — zero edits. The agent loop
  doesn't know OpenRouter exists.
- `AnthropicProvider` — untouched. Adding the OpenRouter provider had no
  cross-provider effects.
- The 0.27.0 rate limiter — still applies to `RunMonitoring` regardless of
  provider, so the cost-bound guarantee carries over to OpenRouter monitoring.

---

## Follow-ups

### Phase 3 — `app_agent_config.provider` column
This is the next step that actually lets a user select OpenRouter for an app.
Without it, OpenRouterProvider is registered but unreachable.

- Migration `022_agent_config_provider.up.sql` adding `provider TEXT NOT NULL DEFAULT 'anthropic'`
- `sqlc generate` to add `Provider` to the `AppAgentConfig` struct
- Two-line change in `loop.go` `RunConversation` and `RunMonitoring`:
  replace `a.providerFor(defaultProviderName)` with
  `a.providerFor(appCfg.Provider)`
- Handler validation in `applications.go`: accept `provider`, validate it
  against `{"anthropic", "openrouter"}`. Reject `"openrouter"` at the API
  boundary if `cfg.OpenRouterKey == ""`, rather than silently falling back

### Phase 4 — `GET /api/models`
Curated model lists with pricing/context, conditionally including
OpenRouter models when the key is set. Lives in
`backend/internal/agent/models.go`.

### Phases 5–6 — Frontend dropdown + (later) streaming
Dropdown grouped by provider with `<optgroup>`, then the streaming work
slice (provider interface method + both providers + chat.go + useAgent.ts).

### Optional: a tiny end-to-end smoke test
Phase 2's tests cover wire format but not the full agent loop end-to-end
through OpenRouter. Once Phase 3 ships and an app can be configured for
OpenRouter via the DB column, a manual smoke test (set provider in DB,
send a chat message that triggers a `search_logs` tool call, confirm the
loop completes) is the cheapest end-to-end validation. No need to automate
it before Phase 5 lands.
