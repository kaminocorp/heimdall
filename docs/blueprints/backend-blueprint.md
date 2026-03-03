# Backend Blueprint

A top-down walkthrough of the Heimdall Go backend — how it boots, what lives where, and how the layers connect.

---

## 1. Boot Sequence

`backend/cmd/heimdall/main.go` — the entrypoint. Everything starts here, in this order:

```
config.Load()           ← read env vars into Config struct
cfg.Validate()          ← fail fast if DATABASE_URL / ANTHROPIC_API_KEY / SUPABASE_URL missing

NewJWKSClient(url)      ← create JWKS client pointing at Supabase
jwks.Fetch(ctx)         ← pre-load ECDSA signing keys (fail if unreachable)

pgxpool.New(ctx, url)   ← establish Postgres connection pool

agent.New(queries, cfg) ← create Anthropic SDK client + Agent struct
api.NewRouter(...)      ← wire up Chi router with all handlers + middleware

http.Server.Listen()    ← bind port, timeouts: read 30s, write 5m, idle 2m

signal.Notify(SIGINT|SIGTERM)  ← block until signal, then graceful shutdown (10s)
```

The dependency chain flows in one direction:

```
Config → JWKSClient → Pool → Queries → Agent → Router → Server
```

---

## 2. Project Layout

```
backend/
├── cmd/
│   └── heimdall/
│       └── main.go                     # Entrypoint
│
├── internal/
│   ├── config/
│   │   └── config.go                   # Config struct, env loading, validation
│   │
│   ├── api/
│   │   ├── router.go                   # Chi route table + middleware wiring
│   │   ├── middleware/
│   │   │   ├── auth.go                 # JWKS client, JWT validation, Auth() middleware
│   │   │   ├── cors.go                 # CORS headers (allow-all)
│   │   │   └── logging.go             # Request logging (method, path, status, duration)
│   │   └── handlers/
│   │       ├── server.go               # Server struct — shared dependencies
│   │       ├── auth.go                 # GET /api/auth/me
│   │       ├── connections.go          # CRUD + test for /api/connections
│   │       ├── agent.go                # GET/PUT /api/agent/config, POST /api/agent/run
│   │       ├── chat.go                 # WebSocket /ws/chat (multi-turn agent chat)
│   │       ├── conversations.go        # GET /api/conversations
│   │       ├── logs.go                 # GET /api/logs (unified log feed)
│   │       ├── webhooks.go             # POST /api/webhooks/logs (log ingestion)
│   │       ├── reports.go              # Stubs for /api/reports
│   │       ├── helpers.go              # jsonError() utility
│   │       └── userqueries.go          # UserQueries() — RLS-scoped DB access
│   │
│   ├── agent/
│   │   ├── agent.go                    # Agent struct, New(), Start(), Stop()
│   │   ├── loop.go                     # RunLoop(), RunConversation() — tool-use loop
│   │   ├── tools.go                    # ToolRegistry() definitions + Dispatch()
│   │   ├── tools_logs.go              # search_logs implementation
│   │   ├── tools_db.go                # query_database implementation
│   │   ├── tools_codebase.go          # Stub (Phase 5+)
│   │   ├── tools_memory.go            # Stub (Phase 5+)
│   │   ├── prompt.go                   # System prompt + BuildSystemPrompt()
│   │   ├── emit.go                     # EmitLog() — fire-and-forget agent_log writes
│   │   ├── message.go                  # Message struct (ID, Role, Content, Timestamp)
│   │   └── monitor.go                  # Stub — continuous monitoring goroutine
│   │
│   ├── connectors/
│   │   ├── connector.go                # Connector, StreamConnector, QueryConnector interfaces
│   │   ├── registry.go                 # Thread-safe connector registry
│   │   ├── database/
│   │   │   └── postgres.go             # Read-only Postgres connector
│   │   ├── logs/
│   │   │   ├── webhook.go              # Webhook connector (stub)
│   │   │   └── syslog.go              # Syslog connector (stub)
│   │   └── codebase/
│   │       └── github.go               # GitHub connector (stub)
│   │
│   └── db/
│       ├── db.go                       # DBTX interface, Queries struct, WithTx()
│       ├── models.go                   # Generated Go types for all tables
│       ├── queries/                    # Raw SQL query definitions (input to sqlc)
│       │   ├── connections.sql
│       │   ├── conversations.sql
│       │   ├── investigations.sql
│       │   ├── log_buffer.sql
│       │   ├── agent_config.sql
│       │   ├── agent_log.sql
│       │   └── users.sql
│       ├── connections.sql.go          # Generated query functions
│       ├── conversations.sql.go
│       ├── investigations.sql.go
│       ├── log_buffer.sql.go
│       ├── agent_config.sql.go
│       ├── agent_log.sql.go
│       └── users.sql.go
│
├── migrations/                         # PostgreSQL migrations (001–013)
├── sqlc.yaml                           # sqlc codegen config
├── Dockerfile
├── go.mod
└── go.sum
```

---

## 3. Server Struct

All handlers hang off a single `Server` struct (`handlers/server.go`). This is the central dependency holder:

```go
type Server struct {
    Config  *config.Config         // env vars (API keys, URLs)
    Pool    *pgxpool.Pool          // Postgres connection pool (for UserQueries transactions)
    Queries *db.Queries            // sqlc-generated query functions (default, no RLS context)
    Agent   *agent.Agent           // Claude agent instance
    JWKS    *middleware.JWKSClient // JWT signing key cache
}
```

`Queries` is the default accessor used by non-user-scoped operations (webhook ingestion, agent config). User-scoped handlers call `s.UserQueries()` instead, which starts a transaction with RLS context.

---

## 4. Router & Middleware

### Middleware Chain

Applied globally to every request via `r.Use()`:

```
Request → Logging → CORS → Handler
```

| Middleware | File | What it does |
|-----------|------|-------------|
| **Logging** | `middleware/logging.go` | Wraps ResponseWriter to capture status code. Logs method, path, status, duration via `slog`. Implements `Unwrap()` so WebSocket libraries can access the underlying `net.Conn`. |
| **CORS** | `middleware/cors.go` | Sets `Access-Control-Allow-Origin: *`, allows GET/POST/PUT/DELETE/OPTIONS, allows Content-Type and Authorization headers. Returns 204 for preflight OPTIONS. |

### Auth Middleware

Applied to protected route groups via `r.Group()`:

| Component | What it does |
|-----------|-------------|
| **Auth()** | Extracts `Bearer <token>` from Authorization header → calls `ValidateJWT()` → injects `userID` (UUID) into request context → passes to handler. Returns 401 on failure. |
| **ValidateJWT()** | Parses JWT, validates ES256 signature against Supabase JWKS keys, checks expiry, extracts `sub` claim as user UUID. Shared by both HTTP middleware and WebSocket auth. |
| **JWKSClient** | Fetches and caches ECDSA public keys from Supabase `/.well-known/jwks.json`. 10-minute cache. Background refresh when stale but key exists. Force refresh on unknown `kid` (key rotation). |

### Route Table

```
router.go
├── Global middleware: Logging, CORS
│
├── /api
│   ├── POST /webhooks/logs              ← Public (webhook token auth, no JWT)
│   │
│   └── [Auth middleware group]
│       ├── /connections
│       │   ├── GET    /                 ← ListConnections
│       │   ├── POST   /                 ← CreateConnection
│       │   ├── GET    /{id}             ← GetConnection
│       │   ├── PUT    /{id}             ← UpdateConnection
│       │   ├── DELETE /{id}             ← DeleteConnection
│       │   └── POST   /{id}/test        ← TestConnection
│       │
│       ├── /agent
│       │   ├── GET    /config           ← GetAgentConfig
│       │   ├── PUT    /config           ← UpdateAgentConfig
│       │   └── POST   /run              ← RunAgent (one-shot)
│       │
│       ├── GET /logs                     ← ListLogs (unified feed)
│       │
│       ├── /conversations
│       │   ├── GET    /                 ← ListConversations
│       │   └── GET    /{id}             ← GetConversation
│       │
│       ├── GET /auth/me                  ← Me (user info)
│       │
│       └── /reports
│           ├── GET    /                 ← ListReports (stub)
│           └── GET    /{id}             ← GetReport (stub)
│
└── GET /ws/chat                          ← HandleChat (WebSocket, JWT via ?token= query param)
```

Note: `/ws/chat` sits outside the `/api` group and outside the Auth middleware. It handles JWT validation manually from the `?token=` query parameter because the browser WebSocket API cannot set custom headers.

---

## 5. Handler Layer

### How Handlers Access the Database

Two patterns, depending on whether the data is user-scoped:

**User-scoped** (most handlers):
```go
queries, done, err := s.UserQueries(ctx, userID)
defer done()
// queries is now scoped to a transaction with SET LOCAL app.current_user_id = userID
```

**Global** (webhook ingestion, agent config):
```go
s.Queries.GetAgentConfig(ctx)  // no user context needed
```

### Connections (`connections.go`)

| Handler | Notes |
|---------|-------|
| **ListConnections** | Returns all connections for the authenticated user. |
| **CreateConnection** | Accepts name, type, direction, config. For `webhook_logs` type, auto-generates a 32-byte webhook token and merges it into config. Status defaults to `inactive`. |
| **GetConnection** | Single connection by ID. User-scoped. |
| **UpdateConnection** | Updates name, type, direction, config, status. User-scoped. |
| **DeleteConnection** | Returns 204. User-scoped. |
| **TestConnection** | For `postgres`/`database` type: creates a Postgres connector from config, calls `Connect()` + a test query, updates status to `active` or `error`. For webhook/syslog/github: auto-passes (nothing to ping). |

### Chat (`chat.go`)

The WebSocket multi-turn chat handler — the most complex handler in the backend.

**Connection flow:**
1. Extract JWT from `?token=` query param → validate → get `userID`
2. Accept WebSocket upgrade
3. Load existing conversation (if `?conversation_id=` provided) or create new one
4. Send `{type: "system", conversation_id: "..."}` to client

**Message loop (runs until disconnect):**
1. Read `{content: "..."}` from client
2. Build `chatMessage` with UUID, role "user", RFC3339 timestamp
3. Append to `storedMessages` array, persist to DB
4. Set conversation title from first user message (truncated to 50 chars)
5. Send `{type: "status", content: "thinking"}` to client
6. Convert history to `agent.Message` format (all messages except current)
7. Call `Agent.RunConversation(ctx, userID, &convID, history, input)`
8. Build `chatMessage` for agent response, append to stored, persist
9. Send `{role: "agent", content: "...", ...}` to client
10. Loop back to step 1

**WebSocket message types** (server → client):
- `system` — connection established, includes `conversation_id`
- `status` — agent state (e.g. `"thinking"`)
- `error` — error message
- `agent` — agent response (has `id`, `role`, `content`, `timestamp`)

### Logs (`logs.go`)

Serves a unified feed that merges two sources:

- **Raw logs** from `log_buffer` (source: `"raw"`)
- **Agent logs** from `agent_log` (source: `"agent"`)

Supports filters: `severity`, `connection_id`, `source` (raw|agent|all). Pagination via `limit` (default 50, max 200) and `offset`. When fetching both sources, results are merged and sorted by timestamp descending.

### Webhooks (`webhooks.go`)

Public endpoint — no JWT required. Auth is via a Bearer token that maps to a connection's `config->>'webhook_token'`.

Accepts a single JSON object or a JSON array of entries. Each entry requires `source_type` and `payload` (JSONB), with optional `severity`. Inserts into `log_buffer` with the connection's `user_id`.

### Agent Config (`agent.go`)

- **GetAgentConfig** — returns the singleton row from `agent_config` (model, mode, schedule, system_prompt_override)
- **UpdateAgentConfig** — upserts the singleton row
- **RunAgent** — one-shot execution: calls `Agent.RunLoop(ctx, userID, input)`, returns the final text response

### Conversations (`conversations.go`)

- **ListConversations** — user-scoped, returns summaries (ID, title, timestamps), paginated
- **GetConversation** — user-scoped, returns full conversation including messages JSONB

### Auth (`auth.go`)

- **Me** — returns the authenticated user's record (ID, email, created_at)

### Reports (`reports.go`)

Stubs — `ListReports` returns an empty array, `GetReport` returns 501.

---

## 6. Agent System

### Agent Struct (`agent/agent.go`)

```go
type Agent struct {
    queries *db.Queries          // for reading config, logs, connections
    client  *anthropic.Client    // Anthropic SDK client
    config  *config.Config       // app config (API keys, URLs)
}
```

Created at boot with `agent.New(queries, cfg)`, which instantiates the Anthropic client using the API key.

### Tool-Use Loop (`agent/loop.go`)

Two entry points:
- **`RunLoop(ctx, userID, input)`** — one-shot, no history. Delegates to `RunConversation` with empty history.
- **`RunConversation(ctx, userID, conversationID, history, input)`** — full multi-turn loop.

The loop:

```
1. Load AgentConfig from DB (model, system prompt override)
   Default model: claude-sonnet-4-5-20250929

2. Build system prompt via BuildSystemPrompt(override)

3. Convert conversation history to Claude MessageParam objects
   Append new user input as final message

4. Loop (max 10 iterations):
   │
   ├─ Call claude.Messages.New(model, system, messages, tools, max_tokens=4096)
   │
   ├─ StopReason == end_turn?
   │  └─ Extract text → EmitLog("observation") → return response
   │
   ├─ StopReason == tool_use?
   │  ├─ Append assistant message (with tool_use blocks) to messages
   │  ├─ For each tool_use block:
   │  │   ├─ Parse input JSON
   │  │   ├─ EmitLog("tool_call", {tool, input})
   │  │   ├─ Dispatch(ctx, userID, toolName, input)
   │  │   ├─ On error: return isError tool result + EmitLog("tool_result", {error})
   │  │   └─ On success: return tool result + EmitLog("tool_result", {preview})
   │  ├─ Append user message with all ToolResultBlocks to messages
   │  └─ Continue loop
   │
   └─ Other stop reason → return whatever text extracted

5. If 10 iterations exceeded → return error
```

### Tool Registry (`agent/tools.go`)

Two tools registered as Claude tool definitions:

| Tool | Required Inputs | Optional Inputs | Description |
|------|----------------|-----------------|-------------|
| **search_logs** | `query` (string) | `severity`, `limit` (default 20, max 200) | Search `log_buffer` for patterns/keywords |
| **query_database** | `sql` (string), `connection_id` (UUID) | — | Execute read-only SQL on a user's connected Postgres |

`Dispatch()` routes by tool name to the implementation function:
```go
switch name {
case "search_logs":     return a.toolSearchLogs(ctx, userID, input)
case "query_database":  return a.toolQueryDatabase(ctx, userID, input)
default:                return error("unknown tool")
}
```

### search_logs (`agent/tools_logs.go`)

Queries `log_buffer` by user ID, with optional severity filter. Returns JSON:
```json
{"results": [{id, connection_id, severity, payload, ingested_at}], "count": N}
```

### query_database (`agent/tools_db.go`)

1. Validates `connection_id` exists and belongs to user
2. Checks connection type is "database" or "postgres"
3. Creates a `database.Postgres` connector from the connection's config JSONB
4. Calls `Connect()` then `Query(ctx, sql)`
5. Returns JSON: `{"rows": [{col: val, ...}], "row_count": N}`

The connector enforces read-only via `default_transaction_read_only=on` in the connection string.

### System Prompt (`agent/prompt.go`)

```
You are Heimdall, an autonomous AI monitoring agent for production applications.

Your job is to watch, understand, investigate, and report on system health.
You have access to tools that let you search logs and query connected databases.

When you detect an anomaly:
1. Investigate using your available tools
2. Build a diagnosis based on evidence
3. Report your findings with severity assessment

You are a monitoring and diagnostics tool — you watch and diagnose,
you don't take action on systems.
```

If the user has set a `system_prompt_override` in agent config, it's appended after a blank line.

### Agent Observability (`agent/emit.go`)

`EmitLog()` is fire-and-forget — it inserts into `agent_log` but errors are logged and swallowed, never propagated. This ensures logging failures don't break agent execution.

Entry types: `"tool_call"`, `"tool_result"`, `"observation"`

Each entry includes: `user_id`, optional `conversation_id`, `entry_type`, `summary` (truncated to 200 chars for observations), `detail` (arbitrary JSON map).

---

## 7. Connectors

### Interface Hierarchy (`connectors/connector.go`)

```
Connector                      Connect(), Health(), Close()
├── StreamConnector            + Stream(ctx, chan<- []byte)
└── QueryConnector             + Query(ctx, query string) (any, error)
```

### Connector Registry (`connectors/registry.go`)

Thread-safe `map[string]Connector` with `sync.RWMutex`. Methods: `Register(id, conn)`, `Get(id)`, `Remove(id)` (closes connector), `All()`. Intended for managing active runtime integrations.

### Postgres Connector (`connectors/database/postgres.go`)

The only fully implemented connector. Used by the agent's `query_database` tool.

**Config** (parsed from connection's JSONB):
```json
{"host": "", "port": 5432, "database": "", "user": "", "password": "", "ssl_mode": "require"}
```

**Security**: Connection string appends `default_transaction_read_only=on` — all queries are forced read-only at the Postgres level. Credentials are URL-encoded to handle special characters.

**Methods**:
- `New(configJSON)` — parse config, build connection string
- `Connect(ctx)` — establish `pgx.Conn`
- `Query(ctx, sql)` — execute query, return `[]map[string]any` (column name → value)
- `Close(ctx)` — close connection

### Stub Connectors

| Connector | Location | Purpose | Status |
|-----------|----------|---------|--------|
| Webhook | `connectors/logs/webhook.go` | Buffer incoming webhook payloads | Stub — webhook ingestion handled by `webhooks.go` handler directly |
| Syslog | `connectors/logs/syslog.go` | Listen for syslog messages | Stub |
| GitHub | `connectors/codebase/github.go` | Code search via GitHub API | Stub |

---

## 8. Database Layer

### sqlc Configuration (`sqlc.yaml`)

- Engine: PostgreSQL
- sqlc version: v1.30.0
- Queries: `internal/db/queries/*.sql` → generated Go in `internal/db/`
- Type overrides: `uuid` → `uuid.UUID`, `jsonb` → `json.RawMessage`, `timestamptz` → `time.Time`

### Generated Code Pattern

For each `.sql` file in `queries/`, sqlc generates a corresponding `.sql.go` file with:
- A `const` for the raw SQL string
- A `Params` struct for each query's parameters
- A method on `*Queries` that executes the query and returns typed results

The `Queries` struct wraps a `DBTX` interface (satisfied by both `*pgxpool.Pool` and `pgx.Tx`), allowing the same queries to run on a pool connection or within a transaction.

`db.New(pool)` → creates Queries on the pool (default)
`queries.WithTx(tx)` → creates Queries on a transaction (used by `UserQueries`)

### Tables

#### users
| Column | Type | Notes |
|--------|------|-------|
| id | UUID PK | Synced from Supabase `auth.users` via trigger |
| email | text | |
| created_at | timestamptz | |

#### connections
| Column | Type | Notes |
|--------|------|-------|
| id | UUID PK | |
| user_id | UUID FK → users | |
| name | text | Display name |
| type | text | `webhook_logs`, `postgres`, `database`, `syslog`, `github` |
| direction | text | `inbound`, `outbound`, `bidirectional` |
| config | jsonb | Type-specific config (host, port, webhook_token, etc.) |
| status | text | `inactive`, `active`, `error` |
| last_seen | timestamptz | Nullable |
| created_at, updated_at | timestamptz | |

#### log_buffer
| Column | Type | Notes |
|--------|------|-------|
| id | UUID PK | |
| user_id | UUID FK → users | |
| connection_id | UUID FK → connections | |
| source_type | text | `webhook` |
| severity | text | Nullable — `critical`, `error`, `warning`, `info`, `debug` |
| payload | jsonb | Raw log entry from source |
| ingested_at | timestamptz | |

#### agent_log
| Column | Type | Notes |
|--------|------|-------|
| id | UUID PK | |
| user_id | UUID FK → users | |
| conversation_id | UUID | Nullable — links to conversation if from chat |
| entry_type | text | `tool_call`, `tool_result`, `observation` |
| summary | text | Human-readable summary |
| detail | jsonb | Nullable — structured data (tool name, input, result) |
| severity | text | Nullable |
| created_at | timestamptz | |

#### conversations
| Column | Type | Notes |
|--------|------|-------|
| id | UUID PK | |
| user_id | UUID FK → users | |
| investigation_id | UUID | Nullable FK → investigations |
| title | text | Nullable — set from first user message |
| messages | jsonb | Array of `{id, role, content, timestamp}` |
| created_at, updated_at | timestamptz | |

#### investigations
| Column | Type | Notes |
|--------|------|-------|
| id | UUID PK | |
| user_id | UUID FK → users | |
| trigger_type | text | |
| trigger_source | text | Nullable |
| summary | text | |
| severity | text | |
| status | text | `open`, `investigating`, `resolved`, `dismissed` |
| context | jsonb | |
| findings | jsonb | |
| tool_trace | jsonb | |
| resolution | text | Nullable |
| started_at | timestamptz | |
| resolved_at | timestamptz | Nullable |

#### agent_config
| Column | Type | Notes |
|--------|------|-------|
| id | int PK | Singleton — always row 1 |
| model | text | Claude model ID (default: `claude-sonnet-4-5-20250929`) |
| mode | text | `continuous` |
| schedule | text | Nullable — cron expression |
| system_prompt_override | text | Nullable — appended to base system prompt |
| created_at, updated_at | timestamptz | |

---

## 9. Schema Evolution (Migrations)

```
001  Create connections table (name, type, direction, config, status)
002  Create agent_config table (singleton, model, mode, schedule)
003  Create investigations table (trigger, summary, severity, status, findings)
004  Create conversations table (messages JSONB, links to investigation)
005  Create log_buffer table (connection_id, source_type, payload)
006  Create users table + auto-sync trigger from Supabase auth.users
007  Add user_id to connections (backfill from first user + NOT NULL constraint)
008  Add index on connections.config->>'webhook_token' for fast webhook lookup
009  Add user_id to conversations (backfill + NOT NULL constraint)
010  Create agent_log table (entry_type, summary, detail, conversation_id)
011  Add user_id to investigations
012  Add user_id to log_buffer (backfill from connections.user_id)
013  Enable RLS on all 6 user-scoped tables + create app_current_user_id() helper
```

The pattern: tables were created first (001–005), then user-scoping was added retroactively (006–012), culminating in RLS enforcement (013).

---

## 10. Security Model

### Authentication

**HTTP routes**: Supabase-issued JWT in `Authorization: Bearer <token>` header. Validated via ES256 against JWKS-cached public keys.

**WebSocket**: JWT passed as `?token=` query parameter (browser WebSocket API limitation — can't set custom headers). Same `ValidateJWT()` function used for both paths.

**Webhook ingestion**: Bearer token looked up in `connections.config->>'webhook_token'` — maps to a connection and its owning user.

### Row-Level Security (RLS)

Every user-scoped table has an RLS policy:

```sql
CREATE POLICY [table]_owner ON [table]
  FOR ALL
  USING (user_id = app_current_user_id());

-- Helper function:
CREATE FUNCTION app_current_user_id() RETURNS uuid AS $$
  SELECT nullif(current_setting('app.current_user_id', true), '')::uuid;
$$ LANGUAGE sql STABLE;
```

**How it's activated from Go**:
```go
// handlers/userqueries.go
tx, _ := s.Pool.Begin(ctx)
tx.Exec(ctx, "SET LOCAL app.current_user_id = $1", userID.String())
return s.Queries.WithTx(tx)  // all subsequent queries run inside this transaction
```

**Current enforcement model**: The backend connects as the `postgres` superuser, which bypasses RLS by default. The policies protect against non-owner access paths (Supabase dashboard roles, PostgREST, direct `psql` with other roles). The `SET LOCAL` plumbing is in place so if the backend ever migrates to a non-owner app role, RLS activates automatically.

### Read-Only Database Queries

The Postgres connector appends `default_transaction_read_only=on` to the connection string. All agent queries against user databases are forced read-only at the Postgres level — writes will error.

---

## 11. Data Flows

### Log Ingestion

```
External App
  │
  │  POST /api/webhooks/logs
  │  Authorization: Bearer <webhook_token>
  │  Body: {source_type, severity?, payload}
  ▼
webhooks.go:IngestWebhookLogs()
  │
  ├─ Look up webhook_token in connections table → get connection_id + user_id
  ├─ Parse single entry or JSON array
  ├─ Validate: source_type required, payload required
  ├─ Insert into log_buffer with connection_id + user_id
  └─ Return 201 Created
         │
         ▼
    log_buffer table
         │
    ┌────┴────┐
    ▼         ▼
GET /api/logs    Agent's search_logs tool
(user-facing)    (during investigation)
```

### WebSocket Chat

```
Browser
  │
  │  GET /ws/chat?token=JWT&conversation_id=UUID
  ▼
chat.go:HandleChat()
  │
  ├─ Validate JWT → extract userID
  ├─ Accept WebSocket upgrade
  ├─ Load or create conversation
  ├─ Send {type: "system", conversation_id}
  │
  └─ Message loop:
       │
       ├─ Read {content} from client
       ├─ Build chatMessage, append to storedMessages, persist to DB
       ├─ Send {type: "status", content: "thinking"}
       │
       ├─ Agent.RunConversation(userID, convID, history, input)
       │    │
       │    └─ Tool-use loop (up to 10 iterations)
       │         ├─ Claude API call with tools
       │         ├─ If tool_use: Dispatch → search_logs / query_database
       │         │   └─ EmitLog(tool_call) + EmitLog(tool_result) → agent_log
       │         └─ If end_turn: EmitLog(observation) → return text
       │
       ├─ Build agent chatMessage, append, persist
       └─ Send {role: "agent", content, ...} to client
```

### Agent Database Query

```
Agent loop (during tool_use)
  │
  │  Dispatch("query_database", {sql, connection_id})
  ▼
tools_db.go:toolQueryDatabase()
  │
  ├─ Fetch connection by ID (verify user owns it)
  ├─ Verify type is "database" or "postgres"
  │
  ├─ database.New(connection.Config)    ← parse host/port/user/pass from JSONB
  │    └─ Build connStr with default_transaction_read_only=on
  │
  ├─ connector.Connect(ctx)             ← establish pgx.Conn to user's DB
  ├─ connector.Query(ctx, sql)          ← execute read-only SQL
  │    └─ Returns []map[string]any      ← column name → value
  ├─ connector.Close(ctx)
  │
  └─ Return JSON: {rows: [...], row_count: N}
       │
       ▼
  Back to agent loop → Claude receives tool result → continues reasoning
```

---

## 12. Configuration

### Environment Variables

| Variable | Required | Default | Purpose |
|----------|----------|---------|---------|
| `PORT` | No | `8080` | HTTP server port |
| `DATABASE_URL` | Yes | — | PostgreSQL connection string (Supabase) |
| `ANTHROPIC_API_KEY` | Yes | — | Claude API key for agent |
| `SUPABASE_URL` | Yes | — | Supabase project URL (for JWKS endpoint) |
| `ELEPHANTASM_URL` | No | — | Future: long-term memory service URL |
| `ELEPHANTASM_API_KEY` | No | — | Future: long-term memory auth key |

### Config Struct (`config/config.go`)

```go
type Config struct {
    Port           string
    DatabaseURL    string
    AnthropicKey   string
    ElephantasmURL string
    ElephantasmKey string
    SupabaseURL    string
}
```

`Load()` reads from env vars with `os.Getenv`. `Validate()` checks the three required fields.

---

## 13. Stubs & Future Work

| Component | Location | Purpose | Current State |
|-----------|----------|---------|---------------|
| **Monitor loop** | `agent/monitor.go` | Continuous 24/7 monitoring goroutine — ingest logs, detect anomalies, trigger agent | `Start()` and `Stop()` are noops |
| **GitHub connector** | `connectors/codebase/github.go` | Code search during investigation | Empty struct |
| **Syslog connector** | `connectors/logs/syslog.go` | Ingest syslog streams | Empty struct |
| **search_codebase tool** | `agent/tools_codebase.go` | Agent tool to search connected codebases | Stub |
| **Memory tools** | `agent/tools_memory.go` | `recall_similar_incidents` via Elephantasm | Stub |
| **Reports** | `handlers/reports.go` | Incident report generation and listing | Returns empty/501 |
| **Investigations** | DB queries exist | Track anomaly investigation lifecycle | Queries defined, no handler wiring |
