# Phase 4 — Agent Loop

Core intelligence — Claude API tool-use integration turns Heimdall from a log viewer into an AI monitoring agent that can search logs and query user databases to investigate issues.

---

## Anthropic SDK Integration

### `backend/go.mod`

Added `github.com/anthropics/anthropic-sdk-go v1.26.0` as a direct dependency. Transitive deps include `tidwall/gjson`, `tidwall/sjson`, `tidwall/pretty`, `tidwall/match` (used internally by the SDK for JSON handling).

---

## Agent Struct Refactor

### `backend/internal/agent/agent.go`

Replaced the empty stub with a real struct holding three dependencies:

| Field | Type | Purpose |
|-------|------|---------|
| `queries` | `*db.Queries` | sqlc-generated queries for Heimdall's own DB (log lookups, connection fetches, agent config) |
| `client` | `*anthropic.Client` | SDK client for Claude API calls |
| `config` | `*config.Config` | Server config (carries `AnthropicKey` for client init) |

**Constructor:** `New(queries *db.Queries, cfg *config.Config)` creates the Anthropic client via `anthropic.NewClient(option.WithAPIKey(cfg.AnthropicKey))`.

---

## Server & Router Wiring

### `backend/cmd/heimdall/main.go`

Creates the Agent in the startup sequence: `ag := agent.New(db.New(pool), cfg)` and passes it to `api.NewRouter(cfg, pool, ag)`.

### `backend/internal/api/handlers/server.go`

Added `Agent *agent.Agent` field to `Server`. Updated `NewServer` to accept `ag *agent.Agent` as a third parameter.

### `backend/internal/api/router.go`

- Updated `NewRouter` signature to `(cfg, pool, ag)`.
- Added `r.Post("/run", s.RunAgent)` in the `/agent` route group (JWT-protected).

---

## Tool Registry

### `backend/internal/agent/tools.go`

Replaced custom `ToolDefinition`/`ToolParam` types with SDK-native `anthropic.ToolUnionParam`. Only two tools are registered for MVP:

**`search_logs`** — parameters:
- `query` (string, required) — describes what to look for
- `severity` (string, optional) — filter by level
- `limit` (integer, optional) — max results (default 20, max 200)

**`query_database`** — parameters:
- `sql` (string, required) — the SQL query to execute
- `connection_id` (string, required) — UUID of the database connection

**`Dispatch` signature:** Changed from `(name, input)` to `(ctx, userID, name, input)` — user ID enables user-scoped tool queries.

**Out-of-scope tools:** `search_codebase`, `recall_similar_incidents`, `recall_lessons` removed from registry and dispatch. Stub files (`tools_codebase.go`, `tools_memory.go`) retained as empty placeholders with deferred-to-post-MVP comments.

---

## search_logs Tool

### `backend/internal/agent/tools_logs.go`

Real implementation that calls sqlc queries:

- If `severity` is provided → `ListLogsByUserAndSeverity(userID, severity, limit, 0)`
- Otherwise → `ListLogsByUser(userID, limit, 0)`

Returns JSON: `{"results": [...], "count": N}` where each result has `id`, `connection_id`, `severity`, `payload`, `ingested_at`.

**Why `query` isn't used for text search:** MVP doesn't implement server-side JSONB text search (would need a GIN index + `to_tsvector`). The `query` parameter tells Claude *what* to look for — it then filters/interprets the returned log entries itself.

---

## Postgres Connector

### `backend/internal/connectors/database/postgres.go`

Replaced stub with real implementation:

| Method | What it does |
|--------|-------------|
| `New(configJSON)` | Parses `host`, `port`, `database`, `user`, `password`, `ssl_mode` from connection config JSONB. Builds connection string. |
| `Connect(ctx)` | Creates a `pgx.Conn` with `default_transaction_read_only=on` in the connection string. |
| `Query(ctx, sql)` | Executes the query, iterates `rows.FieldDescriptions()` + `rows.Values()` to build `[]map[string]any`. |
| `Close(ctx)` | Closes the connection. |

**Read-only enforcement:** `default_transaction_read_only=on` is a PostgreSQL session parameter. The server itself rejects any write operation (INSERT, UPDATE, DELETE, DROP, etc.) with a read-only transaction error. Defense-in-depth — no SQL parsing needed.

**Connection-per-call:** No persistent pool to user databases. Fresh connection → query → close. Simple and safe for MVP frequency.

**Config defaults:** Port defaults to 5432, SSL mode defaults to `require`.

### `backend/internal/connectors/database/postgres_test.go`

Updated test to match new `New(configJSON)` constructor. Added `TestNewMissingFields` for validation coverage.

---

## query_database Tool

### `backend/internal/agent/tools_db.go`

Implementation flow:
1. Parse `sql` and `connection_id` from input
2. `GetConnectionByUser(connID, userID)` — user-scoped lookup, prevents cross-user access
3. Validate connection type is `database` or `postgres`
4. Create Postgres connector from connection config, connect, execute query, close
5. Return JSON: `{"rows": [...], "row_count": N}`

---

## Claude API Tool-Use Loop

### `backend/internal/agent/loop.go`

The core agentic loop — `RunLoop(ctx, userID, input) (string, error)`:

1. **Load config** — reads `agent_config` from DB for model selection and system prompt override. Falls back to `claude-sonnet-4-5` and base system prompt if no DB row.
2. **Build messages** — initial `[UserMessage(input)]`.
3. **Loop** (max 10 iterations):
   - Call `client.Messages.New()` with model, system prompt, messages, tools.
   - `StopReason == end_turn` → extract text, return final response.
   - `StopReason == tool_use` → for each `ToolUseBlock`: unmarshal input JSON, dispatch to tool handler, collect `ToolResultBlock`s (with `isError` flag for failures), append as user message, continue loop.
4. Error if max iterations exceeded.

**Error resilience:** Tool errors are returned as `NewToolResultBlock(id, errorMsg, true)` rather than aborting the loop. Claude can gracefully handle failures (e.g., "The query failed — let me try a different approach").

**Message accumulation:** Each iteration appends the assistant's response and tool results to the messages array, giving Claude full conversation history for context.

---

## System Prompt

### `backend/internal/agent/prompt.go`

Base system prompt tells Claude it is Heimdall and describes its role. Only references the two tools actually registered (`search_logs`, `query_database`) — avoids wasting loop iterations on nonexistent tools. Supports an optional user-defined override appended after the base prompt.

---

## Server Timeouts

### `backend/cmd/heimdall/main.go`

`WriteTimeout` set to **5 minutes** (up from 30s). The `POST /api/agent/run` endpoint is synchronous — a single agent loop can make up to 10 Claude API calls with tool executions in between, easily exceeding 30 seconds. Go's `net/http` `WriteTimeout` starts counting from the end of reading request headers, so a tight timeout silently kills long-running responses with no error to the client.

---

## Agent Config Handlers

### `backend/internal/api/handlers/agent.go`

| Handler | Method | What it does |
|---------|--------|-------------|
| `GetAgentConfig` | `GET /api/agent/config` | Reads from DB via `GetAgentConfig()`, falls back to defaults (`claude-sonnet-4-5-20250929`, `continuous`) if no row |
| `UpdateAgentConfig` | `PUT /api/agent/config` | Decodes request body, calls `UpsertAgentConfig()`, returns updated config |
| `RunAgent` | `POST /api/agent/run` | Extracts userID from JWT, decodes `{"input": "..."}`, calls `Agent.RunLoop()`, returns `{"response": "..."}` |

---

## Files Changed (11 modified, 1 created)

| File | Change |
|------|--------|
| `backend/go.mod` | Added `anthropic-sdk-go v1.26.0` + transitive deps |
| `backend/internal/agent/agent.go` | Real struct fields, updated constructor |
| `backend/internal/agent/prompt.go` | System prompt scoped to registered tools only |
| `backend/cmd/heimdall/main.go` | Creates Agent, passes to router; `WriteTimeout` → 5 min for agent loop |
| `backend/internal/api/handlers/server.go` | Added `Agent` field, updated `NewServer` |
| `backend/internal/api/router.go` | Accepts Agent param, added `/agent/run` route |
| `backend/internal/agent/tools.go` | SDK-native tool types, updated Dispatch |
| `backend/internal/agent/tools_logs.go` | Real `search_logs` implementation |
| `backend/internal/connectors/database/postgres.go` | Real Postgres connector |
| `backend/internal/connectors/database/postgres_test.go` | Updated for new constructor |
| `backend/internal/agent/tools_db.go` | Real `query_database` implementation |
| `backend/internal/agent/loop.go` | Real Claude API tool-use loop |
| `backend/internal/api/handlers/agent.go` | DB-backed config + RunAgent handler |
| `backend/internal/agent/tools_codebase.go` | Cleared stub (deferred to post-MVP) |
| `backend/internal/agent/tools_memory.go` | Cleared stub (deferred to post-MVP) |

---

## Verification

### 1. Build
```bash
cd backend && go build ./cmd/heimdall
```

### 2. Vet + Tests
```bash
cd backend && go vet ./... && go test ./...
```

### 3. Agent config
```bash
# GET — returns defaults or DB row
curl http://localhost:8080/api/agent/config -H "Authorization: Bearer $JWT"

# PUT — persist to DB
curl -X PUT http://localhost:8080/api/agent/config \
  -H "Authorization: Bearer $JWT" \
  -d '{"model":"claude-sonnet-4-5-20250929","mode":"continuous"}'
```

### 4. Agent loop with search_logs
```bash
curl -X POST http://localhost:8080/api/agent/run \
  -H "Authorization: Bearer $JWT" \
  -d '{"input":"Search for any critical log entries"}'
```

### 5. Agent loop with query_database
```bash
curl -X POST http://localhost:8080/api/agent/run \
  -H "Authorization: Bearer $JWT" \
  -d '{"input":"Query connection <id>: SELECT version();"}'
```

### 6. Read-only enforcement
```bash
curl -X POST http://localhost:8080/api/agent/run \
  -H "Authorization: Bearer $JWT" \
  -d '{"input":"Run DROP TABLE users on connection <id>"}'
```
Expected: Query fails with read-only error, agent reports error gracefully.
