# Phase 4 — Agent Loop

## Context

Phases 1–3 delivered auth, connections CRUD, and webhook log ingestion. Heimdall can receive and display logs, but has no intelligence. Phase 4 wires up the core agent loop — Claude API tool-use integration — turning Heimdall from a log viewer into an AI monitoring agent that can search logs and query user databases to investigate issues.

## Scope

**In scope** (per mvp-roadmap.md):
- Add `anthropic-sdk-go` dependency
- Implement real tool-use loop: message → tool calls → execute → return results → repeat
- Implement two tools: `search_logs` and `query_database`
- Agent config read from DB (model, system prompt override)
- Wire agent into server for handler access
- Test endpoint (`POST /api/agent/run`) for end-to-end verification

**Out of scope** (Phase 5+):
- WebSocket chat integration, conversation persistence, streaming responses
- `search_codebase`, `recall_similar_incidents`, `recall_lessons` tools (post-MVP)
- Scheduled monitoring goroutine (post-MVP)
- Frontend changes

---

## Tasks

### Task 1 — Add Anthropic SDK + Refactor Agent struct

**Files:**
- `backend/go.mod` / `go.sum`
- `backend/internal/agent/agent.go`

**What:** `go get github.com/anthropics/anthropic-sdk-go`. Add real fields to Agent struct:

```go
type Agent struct {
    queries *db.Queries
    client  *anthropic.Client
    config  *config.Config
}
```

Update `New()` to accept `(queries *db.Queries, cfg *config.Config)` and create the anthropic client with `option.WithAPIKey(cfg.AnthropicKey)`.

---

### Task 2 — Wire Agent into Server and Router

**Files:**
- `backend/cmd/heimdall/main.go`
- `backend/internal/api/handlers/server.go`
- `backend/internal/api/router.go`

**What:** Create `db.Queries` and Agent in main.go, thread through to Server:

- `main.go`: `ag := agent.New(db.New(pool), cfg)` → pass to `NewRouter`
- `server.go`: Add `Agent *agent.Agent` field to Server struct
- `router.go`: `NewRouter(cfg, pool, ag)` → pass to `NewServer`

---

### Task 3 — Refactor Tool Registry to SDK types

**Files:**
- `backend/internal/agent/tools.go`

**What:** Replace custom `ToolDefinition`/`ToolParam` types with `anthropic.ToolUnionParam`. Register only `search_logs` and `query_database`. Update `Dispatch` signature to `(ctx, userID, name, input)`.

`search_logs` parameters: `query` (string, required), `severity` (string, optional), `limit` (number, optional).

`query_database` parameters: `sql` (string, required), `connection_id` (string, required).

---

### Task 4 — Implement `search_logs` tool

**Files:**
- `backend/internal/agent/tools_logs.go`

**What:** New signature `toolSearchLogs(ctx, userID, input) (string, error)`. Calls existing sqlc queries:
- If `severity` filter present → `ListLogsByUserAndSeverity`
- Otherwise → `ListLogsByUser`
- Default limit 20, max 200

Returns formatted JSON of matching log entries. The `query` parameter is accepted but not used for text search in MVP (future: JSONB payload search).

---

### Task 5 — Implement Postgres connector

**Files:**
- `backend/internal/connectors/database/postgres.go`

**What:** Replace stub with real implementation:
- `New(configJSON json.RawMessage)` constructor that parses host/port/database/user/password/ssl_mode from connection config JSONB
- `Connect(ctx)` creates `pgx.Conn` with `default_transaction_read_only=on` for security
- `Query(ctx, sql)` executes query, returns `[]map[string]any` using `rows.FieldDescriptions()` + `rows.Values()`
- `Close()` closes connection
- Fresh connection per call (no persistent pool — acceptable for MVP)

---

### Task 6 — Implement `query_database` tool

**Files:**
- `backend/internal/agent/tools_db.go`

**What:** New signature `toolQueryDatabase(ctx, userID, input) (string, error)`:
1. Parse `sql` and `connection_id` from input
2. Fetch connection via `queries.GetConnectionByUser(id, userID)` — user-scoped
3. Validate connection type is `database` or `postgres`
4. Create Postgres connector from connection config, connect, execute query, close
5. Return formatted JSON results

---

### Task 7 — Implement Claude API tool-use loop

**Files:**
- `backend/internal/agent/loop.go`

**What:** Replace `RunLoop` stub. New signature: `RunLoop(ctx, userID, input) (string, error)`:

1. Load agent config from DB via `GetAgentConfig` (fallback to defaults if no row)
2. Build system prompt via `BuildSystemPrompt(override)`
3. Create initial messages: `[NewUserMessage(input)]`
4. Loop (max 10 iterations):
   - Call `client.Messages.New()` with model from config, system prompt, messages, tools
   - If `StopReason == "end_turn"` → extract text, return
   - If `StopReason == "tool_use"` → for each `ToolUseBlock`: parse input via `JSON.Input.Raw()`, dispatch to tool, collect `NewToolResultBlock(id, result, isError)`, append as user message
5. Return error if max iterations exceeded

---

### Task 8 — Wire agent config handlers to DB + add test endpoint

**Files:**
- `backend/internal/api/handlers/agent.go`
- `backend/internal/api/router.go`

**What:**
- `GetAgentConfig`: Call `s.Queries.GetAgentConfig()`, return JSON (fallback to defaults on error)
- `UpdateAgentConfig`: Decode request body, call `s.Queries.UpsertAgentConfig()`, return updated config
- New `POST /api/agent/run` handler: extract userID from JWT context, decode `{"input": "..."}`, call `s.Agent.RunLoop(ctx, userID, input)`, return `{"response": "..."}`
- Add route: `r.Post("/run", s.RunAgent)` in the `/agent` group

---

## Task Order

```
Task 1 (SDK + Agent struct)
  └──▶ Task 2 (Wire into Server)
         └──▶ Task 3 (Tool registry)
                ├──▶ Task 4 (search_logs)
                └──▶ Task 5 (Postgres connector)
                       └──▶ Task 6 (query_database)
                              └──▶ Task 7 (RunLoop)
                                     └──▶ Task 8 (Handlers + test endpoint)
```

---

## Key Design Decisions

1. **`RunLoop(ctx, userID, input)`** — userID parameter enables user-scoped tool queries. Phase 5 will pass it from JWT context in WebSocket handler.

2. **SDK-native tool types** — ToolRegistry returns `[]anthropic.ToolUnionParam` directly, no conversion layer.

3. **Read-only enforcement** — Postgres connector uses `default_transaction_read_only=on` in the connection string. Prevents mutations at the PostgreSQL session level.

4. **Connection-per-call** — No persistent pool for user databases. Simple and safe for MVP. Post-MVP can optimize.

5. **Max 10 loop iterations** — Hard cap to prevent runaway tool-use loops.

6. **Test endpoint, not WebSocket** — `POST /api/agent/run` lets us verify the loop works without Phase 5 complexity. Stateless, synchronous, JWT-protected.

---

## Files Changed (10 modified, 0 created)

| File | Change |
|------|--------|
| `backend/go.mod` | Add anthropic-sdk-go |
| `backend/internal/agent/agent.go` | Add fields, update constructor |
| `backend/cmd/heimdall/main.go` | Create Agent, pass to router |
| `backend/internal/api/handlers/server.go` | Add Agent field |
| `backend/internal/api/router.go` | Accept Agent, add `/agent/run` route |
| `backend/internal/agent/tools.go` | SDK types, updated Dispatch |
| `backend/internal/agent/tools_logs.go` | Real search_logs implementation |
| `backend/internal/connectors/database/postgres.go` | Real Postgres connector |
| `backend/internal/agent/tools_db.go` | Real query_database implementation |
| `backend/internal/agent/loop.go` | Real Claude API tool-use loop |
| `backend/internal/api/handlers/agent.go` | DB-backed config + RunAgent handler |

---

## Verification

### 1. Build
```bash
cd backend && go build ./cmd/heimdall
```

### 2. Agent config
```bash
# GET — should return defaults or DB row
curl http://localhost:8080/api/agent/config -H "Authorization: Bearer $JWT"

# PUT — persist to DB
curl -X PUT http://localhost:8080/api/agent/config \
  -H "Authorization: Bearer $JWT" \
  -d '{"model":"claude-sonnet-4-5-20250929","mode":"continuous"}'
```

### 3. Agent loop with search_logs
```bash
# Prerequisite: logs ingested via webhook
curl -X POST http://localhost:8080/api/agent/run \
  -H "Authorization: Bearer $JWT" \
  -d '{"input":"Search for any critical log entries"}'
```
Expected: Agent calls `search_logs`, returns analysis of log entries.

### 4. Agent loop with query_database
```bash
# Prerequisite: postgres connection created with valid DB credentials
curl -X POST http://localhost:8080/api/agent/run \
  -H "Authorization: Bearer $JWT" \
  -d '{"input":"Query connection <id>: SELECT version();"}'
```
Expected: Agent calls `query_database`, returns query result.

### 5. Read-only enforcement
```bash
curl -X POST http://localhost:8080/api/agent/run \
  -H "Authorization: Bearer $JWT" \
  -d '{"input":"Run DROP TABLE users on connection <id>"}'
```
Expected: Query fails with read-only error, agent reports error gracefully.
