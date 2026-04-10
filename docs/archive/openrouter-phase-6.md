# OpenRouter Integration — Phase 6 Completion

**Status:** ✅ Complete (tool-progress streaming MVP — see "Scope cut" below)
**Date:** 2026-04-07
**Plan:** [openrouter-implementation.md](../executing/openrouter-implementation.md), Phase 6
**Validation:** `go build ./...` clean, `go vet ./...` clean, `go test ./...` all green; `vue-tsc --noEmit` clean, `npm run test` all 23 tests green
**Builds on:** [openrouter-phase-5.md](openrouter-phase-5.md)
**Follow-up:** [streaming-implementation.md](../executing/streaming-implementation.md) — token-by-token streaming, the deferred 20%

---

## Scope cut: tool-progress streaming, not token streaming

The original plan called for full token-by-token LLM streaming, which
required ~600 lines spanning the `Provider` interface (new
`ChatCompletionStream` method), both provider implementations
(Anthropic SDK streaming + OpenRouter SSE parsing), the agent loop
(`RunConversationStream`), the chat handler (channel consumer), the
frontend composable (in-progress message slot), and the chat page
(active tools indicator).

We split that into two:

| | Phase 6 (this) | Streaming follow-up |
|---|---|---|
| **Scope** | Stream the loop's *progress* (which tool is running) | Stream the model's *prose* (token-by-token text) |
| **Provider changes** | None | Both providers gain a streaming method |
| **Agent loop changes** | Refactor + add `RunConversationStream` wrapper | Replace blocking calls with streaming consumer |
| **Frontend changes** | Active-tools indicator on chat page | In-progress message slot in `useAgent` |
| **WebSocket message types** | `tool_start`, `tool_result`, `message`, `error` | Adds `chunk` |
| **UX win** | User sees `searching logs · querying database` live during the wait | User also sees the agent's prose typed in real-time |
| **Lines of code** | ~150 | ~450 |

The split is intentional and the rationale is documented in
[`docs/executing/streaming-implementation.md`](../executing/streaming-implementation.md).
**That doc is the canonical "remaining 20%" plan** — re-read it before
picking up token streaming. It contains:

- Step-by-step implementation order
- Provider event-name gotchas (Anthropic SDK version sensitivity,
  OpenRouter SSE quirks)
- A non-goals section (monitoring stays blocking, no cancellation UI,
  no token-level usage tracking)
- Validation gates and a risk register
- A file touch list with current Phase 6 status next to streaming
  follow-up status, so the next implementer knows what's already wired
  up and what they're adding

The recommended smallest meaningful checkpoint is **streaming on
Anthropic only first**, with OpenRouter's `ChatCompletionStream` falling
back to a single chunk + done emitted from the blocking path. That lets
the next iteration validate the loop refactor and frontend rendering
against one provider before committing to the SSE parser.

---

## What this phase actually shipped

### New (2)

| File | Purpose |
|---|---|
| `backend/internal/agent/loop_stream.go` | `AgentEvent` type, `RunConversationStream` method, `streamBufferSize` constant. The streaming entry point that wraps the existing tool-use loop and emits progress events on a channel |
| `docs/executing/streaming-implementation.md` | Full follow-up plan for token streaming — see "Scope cut" above |

### Modified (5)

| File | Change |
|---|---|
| `backend/internal/agent/loop.go` | `RunConversation` becomes a thin wrapper around new `runConversationCore` (private). Loop body extracted unchanged except for two `emitEvent(events, ...)` calls — one before each tool dispatch (`tool_start`) and one after (`tool_result` with `IsError` set on the failure path). The blocking path passes `nil` for the events channel; the streaming path passes the producer-side of `RunConversationStream`'s channel |
| `backend/internal/api/handlers/chat.go` | Replaced the single blocking `RunConversation` call + write with a `for ev := range events` consumer of `RunConversationStream`. Forwards `tool_start` / `tool_result` events to the WebSocket as they happen; persists and writes the final assistant message on the `message` event; falls through to the existing error path on `error` |
| `frontend/src/composables/useAgent.ts` | Added `activeTools` ref. New `parsed.type === 'tool_start'` branch pushes the tool name and clears `isThinking`. New `parsed.type === 'tool_result'` branch removes the tool from the list and re-arms `isThinking` only when no other tools are still running. Error handler clears `activeTools` so a mid-loop failure doesn't strand the indicator. Final chat-message handler clears `activeTools` for the same reason |
| `frontend/src/pages/AgentChatPage.vue` | Pulls `activeTools` from `useAgent`. New indicator block above `<ChatWindow>`: muted accent-colour bar with a pulsing dot and the tool list joined by ` · `. `formatTool` helper turns `search_logs` into `search logs` for display |

### Untouched (intentional)

- **Both `Provider` implementations.** The whole point of the scope cut
  is that streaming lives at the loop layer, not the provider layer. The
  provider interface is unchanged (`ChatCompletion` only). When token
  streaming lands, that's when `ChatCompletionStream` joins the interface.
- **`RunMonitoring` and `monitor.go`.** Monitoring stays blocking
  (loop.go:187 warning still applies). The 0.27.0 rate limiter contract
  is unchanged.
- **`RunLoop`.** Still delegates to `RunConversation`, which still
  delegates to `runConversationCore` with `nil` events. Zero behaviour
  change for callers that don't want streaming.
- **`ChatWindow.vue`.** Vue's reactivity already re-renders when message
  content changes; no changes needed for the in-progress slot today
  (and none for the future token-streaming follow-up either).
- **`tools.go`, `agent.go`, `provider.go`, both providers, sqlc, the
  database.** Nothing in the data model or provider abstraction changed.

---

## The shape of the refactor

The most important architectural change is that `runConversationCore` is
now a single shared loop body with an optional event sink:

```go
func (a *Agent) runConversationCore(
    ctx context.Context,
    userID, appID uuid.UUID,
    conversationID *uuid.UUID,
    history []Message,
    input string,
    events chan<- AgentEvent, // nil for blocking callers
) (string, error)
```

The two callers:

```go
// Blocking — what RunLoop and existing test code use.
func (a *Agent) RunConversation(...) (string, error) {
    return a.runConversationCore(ctx, ..., nil)
}

// Streaming — new entry point for the chat handler.
func (a *Agent) RunConversationStream(...) <-chan AgentEvent {
    ch := make(chan AgentEvent, streamBufferSize)
    go func() {
        defer close(ch)
        text, err := a.runConversationCore(ctx, ..., ch)
        if err != nil {
            ch <- AgentEvent{Type: "error", Content: err.Error()}
            return
        }
        ch <- AgentEvent{Type: "message", Content: text}
    }()
    return ch
}
```

And the loop body uses a nil-safe helper to emit:

```go
func emitEvent(events chan<- AgentEvent, ev AgentEvent) {
    if events == nil {
        return
    }
    events <- ev
}
```

**Why the nil-sink pattern instead of two separate loops:** the
alternative was to duplicate the entire 100-line tool-use loop body for
the streaming variant. Duplication is the source of every "fixed in one
path, forgot the other" bug, and the loop is exactly the kind of code
that gets edited frequently (tool result handling, log emission,
iteration limits). Sharing the body via an optional event sink means
any future change to tool dispatch or message accumulation
automatically applies to both blocking and streaming callers — they're
the same code path, just with one extra `emitEvent` call.

The cost of the pattern is one nil check per tool call, which is free,
and the slight conceptual overhead of "this method has an optional
side-effect." That's a fine trade for not maintaining two parallel
copies of the loop.

---

## The "re-arm thinking after the last tool" detail

This is the smallest piece of UX logic in the phase but worth calling
out because it's the kind of thing that would be tempting to skip:

```ts
if (parsed.type === 'tool_result') {
  if (parsed.tool) {
    activeTools.value = activeTools.value.filter((t) => t !== parsed.tool)
  }
  if (activeTools.value.length === 0) {
    isThinking.value = true
  }
  return
}
```

The agent's loop has three visible UX states:

1. **Initial wait** — user sent a message, agent is composing its first
   response or deciding to call tools. Show "thinking…"
2. **Tool execution** — one or more tools are in flight. Show the
   active-tools indicator.
3. **Synthesis after tools** — tools have all returned, the agent is
   composing its final reply. Show "thinking…" again.

Without the re-arm, state 3 looks like a blank pause with no indicator
at all — the user thinks the chat broke. With the re-arm, the wait is
always visually accounted for. The two indicators never appear at the
same time because `tool_start` clears `isThinking` and `tool_result`
only re-arms it if no other tools remain.

When token streaming arrives, this same flag becomes the on-switch for
the in-progress message slot — the first `chunk` event clears
`isThinking` and the in-progress message takes over. The contract is
already in place.

---

## Validation

```
$ cd backend && go build ./... && go vet ./... && go test ./...
ok    github.com/hejijunhao/heimdall/backend/internal/agent       1.146s
ok    github.com/hejijunhao/heimdall/backend/internal/api/handlers 0.851s
(... all green ...)

$ cd frontend && npx vue-tsc --noEmit && npm run test
✓ src/stores/__tests__/auth.test.ts (5 tests)
✓ src/stores/__tests__/logs.test.ts (5 tests)
✓ src/stores/__tests__/connections.test.ts (5 tests)
✓ src/components/connections/__tests__/ConnectionForm.test.ts (8 tests)
Test Files  4 passed (4)
     Tests  23 passed (23)
```

**Zero test edits required.** The blocking path through
`runConversationCore` with `events == nil` is byte-for-byte equivalent
to the old `RunConversation` body, so every existing test (including
`loop_test.go`'s mock-Anthropic-server tests) passes unchanged. The
streaming path is exercised end-to-end via the chat WebSocket handler
but has no unit test today — same calculus as Phase 4: the producer is
~30 lines wrapping a goroutine, the consumer is mechanical event
forwarding, and the most useful test is the manual smoke test below.

When token streaming lands, that's the right time to add real test
coverage for both providers' streaming paths (mock SSE for OpenRouter,
mock SDK stream for Anthropic). See the validation gates section in
the streaming follow-up doc.

### Validation gates from `openrouter-implementation.md` Phase 6

- [ ] ~~Chat shows tokens appearing incrementally (not all at once)~~
      — **deferred to streaming follow-up**, see scope cut above
- [x] Tool use shows "search logs / query database…" progress indicator
- [x] Final message is complete and matches what's persisted in the database
- [x] Resuming a conversation (loading history) still works
      (`loadMessages` is unchanged, no streaming state to hydrate)
- [x] Error messages still display correctly
      (mid-loop errors emit an `error` event that clears `activeTools`
      and surfaces via the existing error banner)

---

## Manual smoke test recipe

1. **Open the agent chat** for an app with at least one log connection.
2. **Send a prompt that forces tool use:** "What errors have we seen
   in the last hour?" The agent should call `search_logs`.
3. **Watch the indicator flow:**
   - Brief "thinking…" pulse
   - Indicator switches to `search logs` with the pulsing dot
   - Indicator disappears, "thinking…" returns briefly
   - Final agent message lands as a single block
4. **Send a multi-tool prompt:** "Compare error rates with last week
   and find the slowest queries." The agent should call `search_logs`
   and `query_database` (possibly in sequence).
5. **Watch the indicator update through multiple tools:** the active
   list should show one tool, switch to the next, etc., before the
   final synthesis.
6. **Force an error:** point the connection at an unreachable database
   and ask a query question. The `tool_result` event should fire with
   `is_error: true`, the indicator should clear, and the agent's
   response should reflect the failure (the `IsError` field is on the
   wire today but the UI just clears the indicator — styling the error
   state is an optional polish item).
7. **Resume an existing conversation** by appending
   `?conversation_id=...` to the URL. The history should load via the
   existing `loadMessages` path, with no streaming-state weirdness
   (the streaming refs are reset on every new agent turn).

---

## What stayed the same

- The 0.27.0 monitoring rate limiter — interactive chat is not
  rate-limited and never was
- `EmitLog` calls inside the loop — every tool call still writes a
  `tool_call` and `tool_result` row to `agent_log`, independent of the
  WebSocket stream
- Conversation persistence — still happens once on the `message` event,
  with the full accumulated assistant text. Streaming changes delivery,
  not storage
- `RunLoop` — unchanged signature, unchanged behaviour
- All existing tests — unchanged

---

## Follow-ups

### LLM token streaming
**See [`docs/executing/streaming-implementation.md`](../executing/streaming-implementation.md).**
This is the canonical follow-up plan and is tracked separately from
this completion doc so it doesn't get lost in the changelog. It covers
provider interface extension, both provider implementations, the loop
refactor, and the frontend in-progress message slot.

### Error-state styling for `tool_result.is_error`
The wire format already carries `is_error` on `tool_result` events
(see `chat.go`). The frontend currently ignores it and just clears
the active-tools list. A polish item: render the failed tool name in
red briefly before clearing. Skipped for now to keep the diff small.

### Backpressure tuning
`streamBufferSize = 16` (loop_stream.go:38) is sized for the typical
tool-call volume per loop iteration. If a chatty model with many tool
calls saturates the buffer and stalls the producer goroutine, bump it.
The streaming follow-up doc has more notes on this — once `chunk`
events join the stream, the buffer math changes and may need a larger
default.

### Closing the streaming-follow-up loop
When `streaming-implementation.md` is implemented, that doc moves to
`docs/completions/` and a Phase 7 completion file lands here. The
`AgentEvent` type and the consumer contract on both ends are designed
to absorb the upgrade without changing the WebSocket message names or
the chat handler's structure.
