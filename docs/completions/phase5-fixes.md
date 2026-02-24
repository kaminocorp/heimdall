# Phase 5 Fixes — Error Handling Hardening

Post-implementation review of Phases 1–5 identified five error-handling gaps across the backend and frontend. All fixes target failure-mode resilience — the happy path was already correct.

---

## Backend

### `backend/internal/api/handlers/connections.go`

**Fix 1: `rand.Read` error handling**

`crypto/rand.Read` can fail under entropy exhaustion. Previously the error was silently dropped, which could produce a zero-value webhook token.

```go
// Before
rand.Read(b)

// After
if _, err := rand.Read(b); err != nil {
    jsonError(w, "failed to generate webhook token", http.StatusInternalServerError)
    return
}
```

**Fix 2: JSON marshal/unmarshal error handling**

When auto-generating a webhook token for `webhook_logs` connections, malformed config JSON was silently swallowed. Now returns 400 Bad Request if config cannot be unmarshalled, and 500 if the updated config cannot be re-marshalled.

```go
// Before
json.Unmarshal(config, &cfgMap)
config, _ = json.Marshal(cfgMap)

// After — both errors checked with early returns
if err := json.Unmarshal(config, &cfgMap); err != nil { ... }
config, err = json.Marshal(cfgMap); if err != nil { ... }
```

### `backend/internal/api/handlers/chat.go`

**Fix 3: WebSocket write error handling**

Five `wsjson.Write` calls had unchecked return values. If the client disconnects mid-conversation, the server would continue running the agent loop and persisting messages for a dead connection. Now all writes are checked — write failures in the message loop terminate the handler cleanly.

| Location | Message type | Behaviour on error |
|----------|-------------|-------------------|
| System message (connect) | `system` | Return immediately |
| Thinking indicator | `status` | Return immediately |
| Agent error | `error` | Return immediately |
| Agent response | `chatMessage` | Return immediately |
| `writeWSError` helper | `error` | Log (caller already returning) |

---

## Frontend

### `frontend/src/composables/useWebSocket.ts`

**Fix 4: Missing `onerror` handler**

The WebSocket had `onopen` and `onclose` handlers but no `onerror`. Connection failures (network errors, TLS issues, refused connections) were silent — status never updated and the UI showed a stale "connecting" state. Added `onerror` handler that transitions status to `'closed'`.

```typescript
// Added
ws.onerror = () => {
  status.value = 'closed'
}
```

### `frontend/src/types/agent.ts` + `frontend/src/composables/useAgent.ts`

**Fix 5: Type safety bypass removed**

The backend stores messages with `role: "assistant"` but the frontend `ChatMessage` type only accepted `'user' | 'agent'`. This mismatch was papered over with an `as any` cast in `useAgent.ts`. Fixed by widening the type to accept `'assistant'` and removing the cast.

```typescript
// types/agent.ts — before
role: 'user' | 'agent'

// types/agent.ts — after
role: 'user' | 'agent' | 'assistant'

// useAgent.ts — before
role: m.role === ('assistant' as any) ? 'agent' : m.role,

// useAgent.ts — after
role: m.role === 'assistant' ? 'agent' : m.role,
```

---

## Repo Hygiene

### `backend/heimdall` — compiled binary removed from git

**Fix 6: 19MB Go binary tracked in version control**

The compiled `backend/heimdall` binary was committed starting from Phase 3. Every `go build` produces a different binary (build timestamps, linker data), creating ~19MB deltas that git can't delta-compress. Added `backend/heimdall` to `.gitignore` and removed from tracking via `git rm --cached -f`.

### `frontend/vite.config.js` — duplicate config deleted

**Fix 7: Duplicate Vite configuration**

A `vite.config.js` was added alongside the existing `vite.config.ts` with identical content. Vite's resolution order prefers `.ts` over `.js`, so the `.js` file was dead weight. Deleted the duplicate — the original `.ts` file remains.

---

## Verification

Both backend and frontend compile cleanly after all fixes:

| Check | Result |
|-------|--------|
| `go build ./...` | Pass |
| `go vet ./...` | Pass |
| `vue-tsc -b --noEmit` | Pass |
