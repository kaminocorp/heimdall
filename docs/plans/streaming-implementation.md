# LLM Token Streaming — Remaining Work

**Status:** Not started — follow-up to [openrouter-phase-6.md](../completions/openrouter-phase-6.md)
**Owner:** TBD
**Prereqs:** Phase 6 (tool-progress streaming MVP) merged

---

## Why this exists

Phase 6 of the OpenRouter integration plan called for full token-by-token
streaming in the chat UI. We shipped the **tool-progress streaming MVP**
instead — the user sees `searching logs · querying database` live as the
agent works, but the agent's prose still lands in one shot when the loop
returns. This document is the follow-up plan to upgrade that to true
token-level streaming without breaking the existing consumer contract.

The MVP was a deliberate scope cut: 80% of the perceived-responsiveness
win for ~20% of the work, and zero changes to either provider
implementation. Token streaming finishes the job and delivers the last
20% of the UX — incremental prose rendering during the model's
synthesis phase, which currently feels like a 2–10 second silent wait
after the tools finish.

The good news: because Phase 6's WebSocket message contract was designed
with this in mind, **the frontend already handles `chunk` events
gracefully** (it just never receives them today). Most of the work is
backend-side.

---

## What Phase 6 left in place

| Layer | Status |
|---|---|
| `AgentEvent` type with `Type` field | ✅ Built — adding a new `"chunk"` variant is one field |
| `RunConversationStream` returning `<-chan AgentEvent` | ✅ Built — channel producer can fan in token deltas alongside the existing tool events |
| Chat WebSocket handler consumes events in a `for ev := range events` loop | ✅ Built — adding a `case "chunk":` arm is mechanical |
| `useAgent.ts` watch loop dispatches on `parsed.type` | ✅ Built — adding a `parsed.type === 'chunk'` branch follows the same pattern as `tool_start` |
| Reactive message rendering in `ChatWindow` | ✅ Already re-renders as message content grows |

| Layer | Status |
|---|---|
| `Provider.ChatCompletionStream` interface method | ❌ Not yet on the interface |
| `AnthropicProvider.ChatCompletionStream` | ❌ Not implemented |
| `OpenRouterProvider.ChatCompletionStream` | ❌ Not implemented |
| Agent loop calls streaming variant | ❌ Today `runConversationCore` calls blocking `ChatCompletion` only |
| In-progress assistant message rendering | ❌ Frontend currently only renders the final message; needs an "in-progress" message slot |

---

## The plan

### Step 1 — Extend the Provider interface

`backend/internal/agent/provider.go`:

```go
type StreamEvent struct {
    Type     string         // "text_delta", "tool_call_start", "tool_call_delta", "done"
    Delta    string         // text content for "text_delta"
    ToolCall *ToolCall      // populated on "tool_call_start"
    Response *ChatResponse  // populated on "done" — full accumulated response
}

type Provider interface {
    ChatCompletion(ctx context.Context, params ChatParams) (*ChatResponse, error)
    ChatCompletionStream(ctx context.Context, params ChatParams) (<-chan StreamEvent, error)
}
```

The `done` event carries the full accumulated `*ChatResponse` so the agent
loop can dispatch tool calls without re-parsing the deltas. This keeps the
streaming layer purely additive — the loop's existing `StopReason` /
`Blocks` machinery is unchanged.

**Why a separate `StreamEvent` type instead of reusing `AgentEvent`:**
`StreamEvent` is what the *provider* emits (text deltas, tool-call deltas,
final response). `AgentEvent` is what the *agent loop* emits to the chat
handler (tool_start/result/message/error). They live at different layers
and shouldn't merge.

### Step 2 — Implement streaming on AnthropicProvider

`backend/internal/agent/provider_anthropic.go`:

Use `p.client.Messages.NewStreaming(ctx, params)` from `anthropic-sdk-go`.
The returned stream exposes events via `stream.Next() / stream.Current()`.
Spawn a goroutine that:

1. Reads events from the SDK stream.
2. Translates SDK event types to `StreamEvent`:
   - `MessageStartEvent` → ignored (just initial usage stats)
   - `ContentBlockStartEvent` for `tool_use` → `tool_call_start`
   - `ContentBlockDeltaEvent` with `text_delta` → `text_delta`
   - `ContentBlockDeltaEvent` with `input_json_delta` → `tool_call_delta`
     (accumulated locally; emitted as part of the final tool call)
   - `MessageStopEvent` → final `done` event with accumulated response
3. Closes the channel on completion or error.

Verify the exact SDK type names against the version pinned in `go.mod` —
the SDK has shifted streaming event names between versions. The closest
reference today is the test in `loop_test.go` which builds an HTTP-level
mock of the non-streaming API; for streaming you'll need a similar
SSE-level mock.

### Step 3 — Implement streaming on OpenRouterProvider

`backend/internal/agent/provider_openrouter.go`:

Send `"stream": true` in the request body. Parse the SSE response:

```
data: {"choices":[{"delta":{"content":"Hello"},"index":0}]}
data: {"choices":[{"delta":{"content":" world"},"index":0}]}
data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_abc","function":{"name":"search_logs"}}]}}]}
data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"query"}}]}}]}
data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"\":\"timeout\"}"}}]}}]}
data: {"choices":[{"finish_reason":"tool_calls","delta":{}}]}
data: [DONE]
```

Translation rules:

| SSE delta field | StreamEvent |
|---|---|
| `delta.content` | `text_delta` with `Delta = content` |
| `delta.tool_calls[].id` (first occurrence) | `tool_call_start` with empty input |
| `delta.tool_calls[].function.arguments` | `tool_call_delta` (accumulate per index) |
| `finish_reason` set | flush accumulated tool calls into the final `Response`, emit `done` |
| `[DONE]` sentinel | close channel |

**Gotchas to watch for:**

1. **Tool-call argument deltas are JSON fragments, not whole values.** The
   server sends them in pieces; you must concatenate by `tool_calls[].index`
   before parsing. If you `json.Unmarshal` a single delta in isolation, it
   will fail.
2. **`finish_reason` arrives in a delta with empty content.** Don't skip
   empty deltas — check for `finish_reason` first.
3. **`[DONE]` is a literal string, not JSON.** Match it before attempting
   to parse the data line.
4. **OpenRouter occasionally sends keepalive comments** (`: keepalive\n\n`).
   Skip lines that don't start with `data:`.
5. **Different upstream models ship deltas at different rates.** GPT-4o
   sends ~10-token chunks; Gemini sometimes sends one token at a time.
   Test with at least two models from the curated list.

### Step 4 — Refactor the agent loop

`backend/internal/agent/loop.go`:

Today `runConversationCore` calls `provider.ChatCompletion(ctx, params)`.
Add a parameter (or a sibling function) that calls
`provider.ChatCompletionStream(ctx, params)` instead and consumes the
event channel. On each `text_delta`, emit an `AgentEvent{Type: "chunk",
Content: delta}` to the existing `events` channel. On `done`, fall through
to the existing tool-dispatch / loop-iteration logic using the
accumulated `Response`.

The key insight: **the loop still iterates the same way.** Streaming only
changes how each iteration's response is delivered. Tool dispatch, the
iteration cap, the rate limiter, `EmitLog` calls — all unchanged. The
shape of the change is:

```go
// Was:
resp, err := provider.ChatCompletion(ctx, params)
// Now:
resp, err := streamAndAccumulate(ctx, provider, params, events)
```

Where `streamAndAccumulate` is a small helper that consumes the stream,
forwards `text_delta` events to `events` as `chunk` events, and returns
the final `*ChatResponse` on `done`.

**Critical:** `RunMonitoring` MUST NOT switch to streaming. Re-read the
warning in `loop.go:187-190` and the streaming-boundary callout in
`docs/executing/openrouter-implementation.md` Phase 1 Task 1.4. The
0.27.0 rate limiter assumes atomic calls; streaming muddies the contract.

### Step 5 — Frontend in-progress message slot

`frontend/src/composables/useAgent.ts`:

Add a `currentStreamingMsg` ref that holds the in-progress assistant
message. On the first `chunk`, push a new message into `messages` and
hold a reference. On subsequent `chunk`s, append to that message's
content. On the final assistant message (the existing chat-message
event), finalise the in-progress message with the server-assigned ID and
clear the ref.

```ts
if (parsed.type === 'chunk') {
  isThinking.value = false
  if (!currentStreamingMsg.value) {
    currentStreamingMsg.value = {
      id: crypto.randomUUID(),
      role: 'agent',
      content: parsed.content,
      timestamp: new Date().toISOString(),
    }
    messages.value.push(currentStreamingMsg.value)
  } else {
    currentStreamingMsg.value.content += parsed.content
  }
  return
}
```

The existing chat-message handler (which fires on the final WS message)
needs to *replace* `currentStreamingMsg` with the server's persisted
copy rather than push a new entry. This is a one-line change but easy
to forget.

### Step 6 — Chat handler emits chunks

`backend/internal/api/handlers/chat.go`:

Add a `case "chunk":` arm in the existing `for ev := range events` loop:

```go
case "chunk":
    if err := wsjson.Write(ctx, conn, map[string]string{
        "type":    "chunk",
        "content": ev.Content,
    }); err != nil {
        slog.Error("websocket write error", "err", err)
        return
    }
```

The final "message" persist + write block stays exactly as is — the
in-progress streaming is purely visual; persistence still uses the full
accumulated text from the `message` event.

---

## Non-goals

- **Streaming for monitoring mode.** Permanently out of scope. Re-read
  the boundary callout in `loop.go:187` and the rate-limiter rationale.
- **Cancellation mid-stream.** The user can navigate away, but we don't
  support an explicit "stop generating" button. The current
  blocking-via-context cancellation works the same way for streamed
  responses; no new UI needed.
- **Token-level usage tracking.** Streaming gives us a chance to capture
  per-chunk usage if the providers expose it, but that's a separate
  observability project — see the deferred items in the original
  selection plan.
- **Replacing the `tool_start` / `tool_result` events with deltas.**
  The tool progress UI is good as-is; don't refactor it just because
  text streaming arrived.

---

## Validation gates (what "done" looks like)

- [ ] `Provider.ChatCompletionStream` exists on the interface and both
      providers implement it
- [ ] `provider_anthropic_test.go` covers the streaming path with a
      mock SDK stream
- [ ] `provider_openrouter_test.go` covers the streaming path with a
      mock SSE response (multi-line `data:` events, tool-call delta
      accumulation, `[DONE]` termination, keepalive comment handling)
- [ ] Agent loop emits `chunk` events as the model produces text
- [ ] Tool dispatch still works correctly (the `done` event's
      `*ChatResponse` flows into the existing tool loop)
- [ ] Chat UI renders text as it arrives, then "snaps" to the persisted
      message on completion (the snap should be invisible — same content,
      same ID)
- [ ] `RunMonitoring` is unchanged and still calls blocking
      `ChatCompletion`
- [ ] `go test ./...` green
- [ ] Manual smoke test: chat with both an Anthropic model and an
      OpenRouter model, observe text streaming in both cases, observe
      tool progress indicators still working

---

## Risk register

| Risk | Likelihood | Mitigation |
|---|---|---|
| Anthropic SDK streaming event names differ from this doc | High | Check the SDK version in `go.mod` and read the actual exported types before writing the goroutine |
| OpenRouter tool-call argument delta accumulation has off-by-one bugs | Medium | Build the SSE mock test first, with the exact byte sequence from a real response capture |
| The "in-progress message" frontend slot races with the final persist message | Medium | Use a stable client-generated UUID for the in-progress message and reconcile by content on the final event, not by ID |
| Backpressure on a slow WebSocket reader stalls the channel buffer | Low | Channel buffer is 16 today (loop_stream.go:38); a chatty model could fill it. Bump to 64 if it becomes an issue or add a non-blocking send with drop semantics for `chunk` events specifically (text deltas are recoverable from the final message; tool events are not) |
| Streaming makes the rate limiter contract ambiguous | N/A | Doesn't apply — monitoring stays blocking. Interactive chat is not rate-limited |

---

## File touch list

| File | Phase 6 status | Streaming follow-up status |
|---|---|---|
| `backend/internal/agent/provider.go` | ChatCompletion only | Add `StreamEvent` type + `ChatCompletionStream` method |
| `backend/internal/agent/provider_anthropic.go` | Blocking impl | Add streaming impl using SDK `NewStreaming` |
| `backend/internal/agent/provider_openrouter.go` | Blocking impl | Add streaming impl using SSE parsing |
| `backend/internal/agent/loop.go` | Calls blocking ChatCompletion | Switch interactive path to `streamAndAccumulate` helper; monitoring path unchanged |
| `backend/internal/agent/loop_stream.go` | `RunConversationStream` + `AgentEvent` | Add `"chunk"` to the `Type` enum; producer goroutine forwards text deltas |
| `backend/internal/api/handlers/chat.go` | Consumes `tool_start`/`tool_result`/`message`/`error` | Add `case "chunk":` write |
| `frontend/src/composables/useAgent.ts` | Handles tool_start/tool_result | Add `currentStreamingMsg` ref + `chunk` handler + reconciliation on final message |
| `frontend/src/pages/AgentChatPage.vue` | Renders `activeTools` indicator | No changes — message rendering already reactive |
| `frontend/src/components/agent/ChatWindow.vue` | Existing markdown rendering | No changes — re-renders as content grows |

---

## Estimated diff size

Roughly 400–500 lines added across 7 files. ~60% of that is the
OpenRouter SSE parser and its tests; ~25% is the Anthropic streaming
goroutine and its tests; the remaining 15% is the frontend in-progress
slot and the loop helper.

The smallest meaningful checkpoint is **Step 1 + Step 2 alone** —
shipping streaming on Anthropic only, with OpenRouter still calling
the blocking path inside its `ChatCompletionStream` method (the
implementation can fall back to "do a non-streaming call and emit one
big chunk + done"). This lets you validate the loop refactor and
frontend rendering against a single provider before committing to the
SSE parser. Recommended order if anyone touches this in pieces.
