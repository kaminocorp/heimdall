# Phase 5 — Agent Chat

Connects the user to the agent through the WebSocket infrastructure. User messages flow through the WebSocket to the agent loop, agent responses are sent back over the same connection, and conversations are persisted to the database with full multi-turn context.

---

## Database: User-Scope Conversations

### `backend/migrations/009_add_user_id_to_conversations.up.sql`

Added `user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE` to the `conversations` table with an index — same pattern as migration 007 for connections. Without this, conversations had no user scoping and were globally visible.

### `backend/internal/db/queries/conversations.sql`

Rewrote all queries to be user-scoped:

| Query | What changed |
|-------|-------------|
| `CreateConversation` | Added `user_id` parameter |
| `GetConversation` | → `GetConversationByUser` — scoped by `user_id` |
| `ListConversations` | → `ListConversationsByUser` — scoped by `user_id`, `LIMIT`/`OFFSET` pagination |
| `UpdateConversationMessages` | → `UpdateConversationMessagesByUser` — scoped by `user_id` |
| — | Added `UpdateConversationTitleByUser` for setting conversation titles |

Ran `sqlc generate` — `conversations.sql.go` and `models.go` regenerated. `Conversation` struct now has `UserID uuid.UUID` field.

---

## WebSocket JWT Authentication

### `backend/internal/api/middleware/auth.go`

Extracted shared JWT validation into a standalone function:

```go
func ValidateJWT(tokenStr, jwtSecret string) (uuid.UUID, error)
```

Refactored the `Auth` HTTP middleware to call `ValidateJWT` internally. Both HTTP routes and WebSocket authentication now share the same validation logic — parse HMAC-SHA256 token, check expiry, extract `sub` claim as user UUID.

**Why a query parameter?** The browser WebSocket API doesn't support custom headers on the upgrade request. The standard workaround is `?token=<jwt>`. Authentication happens *before* `websocket.Accept` — an invalid token gets an HTTP 401, never a WebSocket connection.

---

## Multi-Turn Agent Method

### `backend/internal/agent/message.go`

New domain type for chat messages stored in the conversations JSONB column:

```go
type Message struct {
    ID        string `json:"id"`
    Role      string `json:"role"`      // "user" or "assistant"
    Content   string `json:"content"`
    Timestamp string `json:"timestamp"`
}
```

### `backend/internal/agent/loop.go`

Added `RunConversation` — the multi-turn variant of `RunLoop`:

```go
func (a *Agent) RunConversation(ctx context.Context, userID uuid.UUID, history []Message, input string) (string, error)
```

Converts stored message history to `[]anthropic.MessageParam` (mapping `"user"` → `NewUserMessage`, `"assistant"` → `NewAssistantMessage`), appends the new user input, and runs the same core tool-use loop.

`RunLoop` now delegates to `RunConversation` with a nil history — preserving backward compatibility for `POST /api/agent/run`.

---

## Conversation REST Endpoints

### `backend/internal/api/handlers/conversations.go`

New handler file with two endpoints:

| Endpoint | Handler | What it does |
|----------|---------|-------------|
| `GET /api/conversations` | `ListConversations` | Returns user's conversations ordered by `updated_at DESC`. Response is a summary array `[{id, title, created_at, updated_at}]` — omits the full messages JSONB to keep list responses lightweight. |
| `GET /api/conversations/:id` | `GetConversation` | Returns a single conversation with full `messages` JSONB array. Scoped by `user_id` — returns 404 if not owned. |

### `backend/internal/api/router.go`

Added `/conversations` route group inside the JWT-protected block:

```go
r.Route("/conversations", func(r chi.Router) {
    r.Get("/", s.ListConversations)
    r.Get("/{id}", s.GetConversation)
})
```

---

## Chat Handler → Agent Integration

### `backend/internal/api/handlers/chat.go`

Replaced the placeholder `HandleChat` (which returned "Agent not yet connected") with the full agent-connected implementation.

**Connection setup:**
1. Reads `token` from query params, validates JWT via `middleware.ValidateJWT` — rejects with HTTP 401 before upgrade if invalid
2. Reads optional `conversation_id` query param
3. If `conversation_id` provided → loads existing conversation from DB (verifies user ownership)
4. If not provided → creates a new conversation
5. Sends a `system` message to the client with the `conversation_id`

**Message read loop:**
1. Reads incoming JSON `{ "content": "..." }`
2. Builds a `chatMessage` struct with UUID, role, content, timestamp
3. Appends user message to in-memory array, persists to DB
4. Sets conversation title on first message (truncated to 50 chars)
5. Sends `{ "type": "status", "content": "thinking" }` to signal loading state
6. Converts stored messages to `[]agent.Message`, calls `agent.RunConversation`
7. Appends agent response to array, persists to DB
8. Sends agent response back over WebSocket

**Error handling:** Agent errors are sent as `{ "type": "error", "content": "..." }` — the connection stays open so the user can retry. The agent loop's own error resilience (tool errors as `isError` tool results) still applies within each invocation.

**Hub/Client not used:** The existing `ws.Hub` and `ws.Client` are broadcast primitives (one-to-many). Chat is 1:1, so the handler manages the `*websocket.Conn` directly with `wsjson.Read`/`wsjson.Write`. The Hub remains available for future broadcast features.

**Message JSONB format:**

```json
[
  { "id": "uuid", "role": "user", "content": "...", "timestamp": "2026-02-23T..." },
  { "id": "uuid", "role": "assistant", "content": "...", "timestamp": "2026-02-23T..." }
]
```

Role mapping: backend stores `"assistant"` (matches Claude API), frontend displays `"agent"`. The `toAgentHistory` helper handles the reverse mapping when loading existing conversations.

---

## Frontend: WebSocket Auth & Conversation Support

### `frontend/src/composables/useWebSocket.ts`

Updated to accept an options object:

```typescript
interface WebSocketOptions {
  token?: string
  conversationId?: string
}
export function useWebSocket(url: string, options?: WebSocketOptions)
```

Builds the WebSocket URL with query parameters: `ws://host/ws/chat?token=xxx&conversation_id=yyy`.

### `frontend/src/composables/useAgent.ts`

Expanded from a simple message relay to a full conversation-aware composable:

- Imports `useAuthStore` and passes the JWT token to `useWebSocket`
- Accepts optional `conversationId` parameter for resuming conversations
- Handles four incoming message types:
  - `type: "system"` → extracts `conversation_id` from server
  - `type: "status"` → sets `isThinking = true`
  - `type: "error"` → clears thinking state, sets error message
  - Default (has `role`) → ChatMessage, clears thinking state
- New `loadMessages(existingMessages)` for loading conversation history from REST API
- Returns `{ messages, status, conversationId, isThinking, error, sendMessage, loadMessages }`

### `frontend/src/api/conversations.ts`

New API client:

```typescript
export function listConversations() → GET /api/conversations
export function getConversation(id) → GET /api/conversations/:id
```

### `frontend/src/types/agent.ts`

Added types for the WebSocket protocol and conversation data:

- `ConversationSummary` — `{ id, title, created_at, updated_at }`
- `Conversation` — extends summary with `messages: ChatMessage[]`
- `WSSystemMessage`, `WSStatusMessage`, `WSErrorMessage` — typed WebSocket envelopes
- `WSMessage` union type

---

## Frontend: Chat UI Enhancements

### `frontend/src/pages/AgentChatPage.vue`

- Shows connection status indicator (colored dot: yellow=connecting, green=open, red=closed)
- Loads conversation history from REST API on mount if `?conversation_id=` is present
- Displays error banner when agent errors occur
- Passes `isThinking` and `disabled` props to ChatWindow

### `frontend/src/components/agent/ChatWindow.vue`

- Accepts `isThinking` and `disabled` props
- Shows animated bouncing dots indicator when agent is processing
- Auto-scrolls to bottom on new messages and thinking state changes via `nextTick`

### `frontend/src/components/agent/ChatInput.vue`

- Accepts `disabled` prop
- Input and send button disabled when `disabled || isThinking`
- Placeholder changes to "Agent is thinking..." when disabled
- Send button also disabled when input is empty

---

## Files Changed (15 modified, 5 new)

| File | Change |
|------|--------|
| `backend/migrations/009_add_user_id_to_conversations.up.sql` | New — adds `user_id` to conversations |
| `backend/migrations/009_add_user_id_to_conversations.down.sql` | New — down migration |
| `backend/internal/db/queries/conversations.sql` | Rewritten — user-scoped queries |
| `backend/internal/db/conversations.sql.go` | Regenerated by sqlc |
| `backend/internal/db/models.go` | Regenerated — Conversation gains UserID field |
| `backend/internal/api/middleware/auth.go` | Extracted `ValidateJWT` helper, refactored Auth middleware |
| `backend/internal/agent/message.go` | New — domain Message type |
| `backend/internal/agent/loop.go` | Added `RunConversation`, `RunLoop` delegates to it |
| `backend/internal/api/handlers/conversations.go` | New — List/Get conversation handlers |
| `backend/internal/api/handlers/chat.go` | Rewritten — real agent-connected chat handler with auth, persistence, status signaling |
| `backend/internal/api/router.go` | Added `/conversations` route group |
| `frontend/src/composables/useWebSocket.ts` | Auth query params support |
| `frontend/src/composables/useAgent.ts` | Conversation ID, thinking state, message type handling, loadMessages |
| `frontend/src/api/conversations.ts` | New — conversation API client |
| `frontend/src/types/agent.ts` | Added ConversationSummary, Conversation, WS message types |
| `frontend/src/pages/AgentChatPage.vue` | Connection status, conversation loading, error display |
| `frontend/src/components/agent/ChatWindow.vue` | Thinking indicator, auto-scroll, disabled prop |
| `frontend/src/components/agent/ChatInput.vue` | Disabled states, dynamic placeholder |
| `docs/executing/phase5-agent-chat.md` | Implementation plan |

---

## Verification

### 1. Build & vet
```bash
cd backend && go vet ./... && go build ./cmd/heimdall
```

### 2. WebSocket auth
```bash
# Should fail with HTTP 401
curl -i http://localhost:8080/ws/chat

# Should upgrade with valid token
wscat -c "ws://localhost:8080/ws/chat?token=$JWT"
```

### 3. Full chat flow
```bash
wscat -c "ws://localhost:8080/ws/chat?token=$JWT"
# ← {"type":"system","conversation_id":"<uuid>"}
> {"content": "Search for critical log entries"}
# ← {"type":"status","content":"thinking"}
# ← {"id":"...","role":"agent","content":"I searched your logs...","timestamp":"..."}
> {"content": "Tell me more about the first one"}
# ← {"type":"status","content":"thinking"}
# ← {"id":"...","role":"agent","content":"Looking at that entry...","timestamp":"..."}
```

### 4. Conversation persistence
```bash
curl http://localhost:8080/api/conversations -H "Authorization: Bearer $JWT"
curl http://localhost:8080/api/conversations/<id> -H "Authorization: Bearer $JWT"
```

### 5. Conversation resumption
```bash
wscat -c "ws://localhost:8080/ws/chat?token=$JWT&conversation_id=<id>"
> {"content": "What did we discuss earlier?"}
# Agent responds with context from prior messages
```

### 6. Frontend
- Navigate to `/agent/chat` — green status dot, input enabled
- Send message — input disables, thinking dots appear
- Agent responds — thinking clears, message appears, auto-scroll
- Refresh with `?conversation_id=<id>` — prior messages load
