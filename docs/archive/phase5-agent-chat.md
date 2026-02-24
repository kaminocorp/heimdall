# Phase 5 — Agent Chat (Implementation Plan)

Connects the user to the agent through the existing WebSocket infrastructure. User messages flow through the WebSocket to the agent loop, agent responses are sent back over the same connection, and conversations are persisted to the database with full multi-turn context.

Reference: [MVP Roadmap](./mvp-roadmap.md) · [Blueprint](./blueprint.md) · [Phase 4 Completion](../completions/phase4-agent-loop.md)

---

## Prerequisites

- Phases 1–4 complete (auth, connections, webhook logs, agent loop)
- Working `RunLoop(ctx, userID, input)` with `search_logs` and `query_database` tools
- Scaffolded WebSocket hub, client, chat handler, and frontend composables

---

## Task 1 — Database: User-Scope Conversations

### Problem

The `conversations` table (migration 004) has no `user_id` column. Unlike `connections` (user-scoped since migration 007), conversations are globally visible — any authenticated user could see any conversation. Phase 5 requires user-scoped access.

### Changes

**Migration `009_add_user_id_to_conversations.up.sql`:**

```sql
ALTER TABLE conversations
    ADD COLUMN user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE;

CREATE INDEX idx_conversations_user_id ON conversations (user_id);
```

**Down migration (`009_add_user_id_to_conversations.down.sql`):**

```sql
DROP INDEX IF EXISTS idx_conversations_user_id;
ALTER TABLE conversations DROP COLUMN user_id;
```

**Rewrite sqlc queries** in `internal/db/queries/conversations.sql`:

| Query | Change |
|-------|--------|
| `CreateConversation` | Add `user_id` parameter |
| `GetConversation` | → `GetConversationByUser` — scope by `user_id` |
| `ListConversations` | → `ListConversationsByUser` — scope by `user_id`, add `LIMIT`/`OFFSET` |
| `UpdateConversationMessages` | → `UpdateConversationMessagesByUser` — scope by `user_id` |

Run `sqlc generate` to regenerate Go code.

### Verification

```bash
cd backend && sqlc generate && go build ./cmd/heimdall
```

---

## Task 2 — Backend: WebSocket JWT Authentication

### Problem

The `/ws/chat` endpoint runs outside the JWT-protected `/api` group (line 53 of `router.go`). HTTP auth middleware can't be used because the browser WebSocket API doesn't support custom headers on the upgrade request. The standard workaround is passing the JWT as a query parameter.

### Changes

**Extract shared JWT validation** in `internal/api/middleware/auth.go`:

```go
// ValidateJWT parses and validates a Supabase JWT, returning the user UUID.
func ValidateJWT(tokenStr, jwtSecret string) (uuid.UUID, error)
```

Refactor the existing `Auth` middleware to call `ValidateJWT` internally, so both HTTP and WebSocket paths share the same validation logic.

**Update `HandleChat`** in `internal/api/handlers/chat.go`:

1. Read `token` from `r.URL.Query().Get("token")`
2. If empty, reject with HTTP 401 **before** calling `websocket.Accept` (this returns a proper HTTP error to the client, not a WebSocket close frame)
3. Call `middleware.ValidateJWT(token, s.Config.SupabaseJWTSecret)`
4. If invalid, reject with HTTP 401
5. Store `userID` for use throughout the WebSocket session

### Why query param, not first-message auth?

First-message auth requires accepting the WebSocket connection before knowing if the client is authorized, which briefly exposes an unauthenticated connection. Query param auth validates *before* upgrade — an invalid token never gets a WebSocket at all. The token is protected in transit by TLS.

### Verification

```bash
# Should fail with HTTP 401 — no token
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
  http://localhost:8080/ws/chat

# Should upgrade — valid token
wscat -c "ws://localhost:8080/ws/chat?token=$JWT"
```

---

## Task 3 — Backend: Multi-Turn Agent Method

### Problem

`RunLoop(ctx, userID, input)` takes a single string and builds a fresh `[]anthropic.MessageParam` with one user message. Each call is stateless — Claude has no context of prior exchanges. For a chat experience, Claude needs the full conversation history.

### Changes

**Add domain message type** in `internal/agent/message.go`:

```go
type Message struct {
    Role    string `json:"role"`    // "user" or "assistant"
    Content string `json:"content"`
}
```

**Add `RunConversation` method** to `internal/agent/loop.go`:

```go
func (a *Agent) RunConversation(ctx context.Context, userID uuid.UUID, history []Message, input string) (string, error)
```

Implementation:
1. Load agent config from DB (same as `RunLoop`)
2. Convert `history` to `[]anthropic.MessageParam` — map `"user"` → `NewUserMessage`, `"assistant"` → `NewAssistantMessage`
3. Append the new user input as the final message
4. Run the core loop (same iteration logic as `RunLoop`)

**Refactor `RunLoop`** to delegate:

```go
func (a *Agent) RunLoop(ctx context.Context, userID uuid.UUID, input string) (string, error) {
    return a.RunConversation(ctx, userID, nil, input)
}
```

This preserves backward compatibility for `POST /api/agent/run`.

### Verification

```bash
# Existing single-turn endpoint should still work
curl -X POST http://localhost:8080/api/agent/run \
  -H "Authorization: Bearer $JWT" \
  -d '{"input":"Search for critical log entries"}'
```

---

## Task 4 — Backend: Conversation REST Endpoints

### Problem

The frontend needs to list past conversations and load a conversation's message history (e.g., on page load or to show a conversation list).

### Changes

**New handler file** `internal/api/handlers/conversations.go`:

| Endpoint | Handler | What it does |
|----------|---------|-------------|
| `GET /api/conversations` | `ListConversations` | Returns user's conversations ordered by `updated_at DESC`. Response: `[{id, title, created_at, updated_at}]` — omits messages for list view. |
| `GET /api/conversations/:id` | `GetConversation` | Returns a single conversation with full `messages` JSONB array. Scoped by `user_id` — returns 404 if not owned. |

**Update `internal/api/router.go`** — add inside the JWT-protected group:

```go
r.Route("/conversations", func(r chi.Router) {
    r.Get("/", s.ListConversations)
    r.Get("/{id}", s.GetConversation)
})
```

### Verification

```bash
# List conversations (empty initially)
curl http://localhost:8080/api/conversations \
  -H "Authorization: Bearer $JWT"

# Get specific conversation (after one is created via WebSocket)
curl http://localhost:8080/api/conversations/<id> \
  -H "Authorization: Bearer $JWT"
```

---

## Task 5 — Backend: Chat Handler → Agent Integration

### Problem

The current `HandleChat` is a placeholder that returns `"Agent not yet connected"` for every message. This task replaces it with real agent integration, conversation persistence, and status signaling.

### Changes

**Rewrite `HandleChat`** in `internal/api/handlers/chat.go`:

**Connection setup:**
1. Authenticate via query param JWT (Task 2)
2. Read optional `conversation_id` from query params
3. If `conversation_id` provided → load existing conversation from DB (verify user owns it)
4. If not provided → create a new conversation in DB (title initially empty)
5. Send a `system` message to the client: `{ type: "system", conversation_id: "<uuid>" }`

**Message read loop:**
1. Read incoming JSON: `{ "content": "user's question" }`
2. Build a `ChatMessage` struct: `{ id, role: "user", content, timestamp }`
3. Append user message to the conversation's JSONB messages array in DB
4. Send a status message to the client: `{ "type": "status", "content": "thinking" }`
5. Load the full conversation message history from DB
6. Parse JSONB messages into `[]agent.Message` for the agent
7. Call `agent.RunConversation(ctx, userID, history, input)`
8. Build agent response: `{ id, role: "agent", content, timestamp }`
9. Append agent response to conversation messages in DB
10. Send agent response to client: `{ "id": "...", "role": "agent", "content": "...", "timestamp": "..." }`
11. Update conversation title if this is the first exchange (truncate first user message to ~50 chars)

**Error handling:**
- If the agent returns an error, send an error message over WebSocket: `{ "type": "error", "content": "..." }` — do NOT close the connection
- Respect context cancellation for graceful shutdown

### Message JSONB format

Messages are stored as a JSON array matching the frontend `ChatMessage` interface:

```json
[
  { "id": "uuid", "role": "user", "content": "...", "timestamp": "2026-02-23T..." },
  { "id": "uuid", "role": "agent", "content": "...", "timestamp": "2026-02-23T..." }
]
```

### Hub/Client usage

The existing `ws.Hub` and `ws.Client` types are designed for broadcasting (one-to-many). Chat is 1:1 — each WebSocket connection has its own agent conversation. The chat handler manages the connection directly using `wsjson.Read`/`wsjson.Write` without going through the Hub. The Hub remains available for future broadcast features (system notifications, monitoring alerts).

### Verification

```bash
# Connect and send a message
wscat -c "ws://localhost:8080/ws/chat?token=$JWT"
> {"content": "What logs do you see?"}
# Should receive:
# {"type": "system", "conversation_id": "..."}   (on connect)
# {"type": "status", "content": "thinking"}       (while agent works)
# {"id": "...", "role": "agent", "content": "I searched your logs and found...", "timestamp": "..."}
```

---

## Task 6 — Frontend: WebSocket Auth & Conversation Support

### Problem

The frontend connects to `/ws/chat` without authentication, doesn't pass conversation context, and has no concept of conversation IDs or thinking states.

### Changes

**Update `composables/useWebSocket.ts`:**

- Accept an options object instead of a bare URL:
  ```typescript
  interface WebSocketOptions {
    token?: string
    conversationId?: string
  }
  export function useWebSocket(url: string, options?: WebSocketOptions)
  ```
- Append query params to the URL: `url?token=xxx&conversation_id=yyy`

**Update `composables/useAgent.ts`:**

- Import `useAuthStore` and pass the JWT token to `useWebSocket`
- Accept optional `conversationId` parameter
- Add reactive state:
  - `conversationId` ref — set from the `system` message received on connect
  - `isThinking` ref — set `true` on `status/thinking`, `false` when agent response arrives
- Handle incoming message types:
  - `type: "system"` → extract `conversation_id`
  - `type: "status"` → toggle `isThinking`
  - `type: "error"` → display error to user
  - Default (has `role`) → existing ChatMessage handling
- Return `{ messages, status, conversationId, isThinking, sendMessage }`

**Add `api/conversations.ts`:**

```typescript
export function listConversations() {
  return client.get<ConversationSummary[]>('/conversations')
}

export function getConversation(id: string) {
  return client.get<Conversation>(`/conversations/${id}`)
}
```

**Add types** (extend `types/agent.ts` or create `types/conversation.ts`):

```typescript
export interface ConversationSummary {
  id: string
  title: string | null
  created_at: string
  updated_at: string
}

export interface Conversation extends ConversationSummary {
  messages: ChatMessage[]
}
```

### Verification

Open the browser, navigate to Agent Chat page, verify:
- WebSocket connects with token (check Network tab → WS → request URL has `?token=...`)
- Sending a message shows thinking state
- Agent response appears in chat

---

## Task 7 — Frontend: Chat UI Enhancements

### Problem

The chat UI is minimal — no loading states, no connection status, no conversation history support.

### Changes

**`AgentChatPage.vue`:**

- Show connection status indicator (connecting / connected / disconnected)
- On mount, if a `conversationId` route param is present, pass it to `useAgent` and load prior messages from the REST API
- Show error banner if WebSocket disconnects unexpectedly

**`ChatWindow.vue`:**

- Accept `isThinking` prop
- Show a thinking indicator (animated dots or spinner) at the bottom of the message list when `isThinking` is true
- Auto-scroll to bottom when new messages arrive or thinking state changes

**`ChatInput.vue`:**

- Disable the input and send button while `isThinking` is true (prevents sending while agent is processing)
- Disable when WebSocket status is not `open`
- Show placeholder text change: "Agent is thinking..." when disabled

### Verification

1. Open Agent Chat page — should show "connected" status
2. Send a message — input disables, thinking indicator appears
3. Agent responds — thinking indicator disappears, input re-enables
4. Refresh the page — conversation history loads from REST API
5. Close the backend — "disconnected" status appears, input disables

---

## Files Changed (Expected)

| File | Change |
|------|--------|
| `backend/migrations/009_add_user_id_to_conversations.up.sql` | New — adds `user_id` to conversations |
| `backend/migrations/009_add_user_id_to_conversations.down.sql` | New — down migration |
| `backend/internal/db/queries/conversations.sql` | Rewrite — user-scoped queries |
| `backend/internal/db/conversations.sql.go` | Regenerated by sqlc |
| `backend/internal/db/models.go` | Regenerated by sqlc (Conversation gains UserID field) |
| `backend/internal/api/middleware/auth.go` | Extract `ValidateJWT` helper |
| `backend/internal/agent/message.go` | New — domain Message type |
| `backend/internal/agent/loop.go` | Add `RunConversation`, refactor `RunLoop` to delegate |
| `backend/internal/api/handlers/conversations.go` | New — List/Get conversation handlers |
| `backend/internal/api/handlers/chat.go` | Rewrite — real agent-connected chat handler |
| `backend/internal/api/router.go` | Add `/conversations` routes |
| `frontend/src/composables/useWebSocket.ts` | Auth query params support |
| `frontend/src/composables/useAgent.ts` | Conversation ID, thinking state, message type handling |
| `frontend/src/api/conversations.ts` | New — conversation API client |
| `frontend/src/types/agent.ts` | Add ConversationSummary, Conversation interfaces |
| `frontend/src/pages/AgentChatPage.vue` | Connection status, conversation loading |
| `frontend/src/components/agent/ChatWindow.vue` | Thinking indicator, auto-scroll |
| `frontend/src/components/agent/ChatInput.vue` | Disabled states |

---

## Task Dependencies

```
Task 1: DB migration (user-scope conversations)
    ├──▶ Task 2: WebSocket JWT auth
    ├──▶ Task 3: Multi-turn agent method
    ├──▶ Task 4: Conversation REST endpoints
    │              │
    │              ▼
    └─────────▶ Task 5: Chat handler → agent integration
                         │
                         ▼
                Task 6: Frontend WebSocket auth & conversation support
                         │
                         ▼
                Task 7: Frontend chat UI enhancements
```

Tasks 2, 3, and 4 are independent of each other and can be done in any order after Task 1. Task 5 depends on all three. Tasks 6 and 7 are sequential after Task 5.

---

## End-to-End Verification

After all tasks are complete:

### 1. Build & test
```bash
cd backend && go vet ./... && go test ./... && go build ./cmd/heimdall
```

### 2. Full chat flow
```bash
wscat -c "ws://localhost:8080/ws/chat?token=$JWT"
> {"content": "Search for any critical log entries"}
# ← system message with conversation_id
# ← status: thinking
# ← agent response with log search results
> {"content": "Can you tell me more about the first one?"}
# ← status: thinking
# ← agent response with context from prior message (multi-turn working)
```

### 3. Conversation persistence
```bash
# List — should show the conversation just created
curl http://localhost:8080/api/conversations -H "Authorization: Bearer $JWT"

# Get — should include all messages from the chat
curl http://localhost:8080/api/conversations/<id> -H "Authorization: Bearer $JWT"
```

### 4. Conversation resumption
```bash
# Reconnect to same conversation — agent has full history
wscat -c "ws://localhost:8080/ws/chat?token=$JWT&conversation_id=<id>"
> {"content": "What did we discuss earlier?"}
# ← agent response referencing prior messages
```

### 5. Frontend
- Navigate to `/agent/chat` in browser
- Verify WebSocket connects (status indicator)
- Send a message, observe thinking state, receive response
- Refresh page — conversation history loads
- Close backend server — disconnected state shown, input disabled
