# Heimdall — Database Models

## Overview

Heimdall's database stores **configuration** and **agent output** — not raw application data. Logs and events from the user's infrastructure flow through Heimdall in-stream; only the intelligence the agent produces gets persisted.

One exception: a rolling 24h `log_buffer` for internal debugging (see below).

Reference: [Scaffolding](./scaffolding.md) | [Blueprint](../plans/blueprint.md)

---

## Schema

### `connections`

Stores integration configuration — how Heimdall connects to the user's infrastructure.

```sql
CREATE TABLE connections (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,                    -- User-defined label ("Production DB", "API Logs")
    type        TEXT NOT NULL,                    -- "postgres", "webhook_logs", "syslog", "github"
    direction   TEXT NOT NULL DEFAULT 'one_way',  -- "one_way" (ingest only) or "two_way" (ingest + query)
    config      JSONB NOT NULL,                   -- Connection-specific config (host, port, credentials, repo URL, etc.)
    status      TEXT NOT NULL DEFAULT 'inactive', -- "active", "inactive", "error"
    last_seen   TIMESTAMPTZ,                      -- Last successful data received / health check
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

**Why JSONB for `config`:** Each connection type has different configuration fields. A Postgres connection needs host/port/dbname/credentials. A webhook log source needs an endpoint path and optional auth token. A GitHub connector needs a repo URL and access token. JSONB avoids a sprawling schema with nullable columns per type.

**Example `config` values:**

```json
-- type: "postgres"
{
    "host": "db.example.com",
    "port": 5432,
    "dbname": "myapp_production",
    "credentials_ref": "vault:postgres-prod"
}

-- type: "webhook_logs"
{
    "endpoint_path": "/ingest/api-logs",
    "auth_token_ref": "vault:webhook-token"
}

-- type: "github"
{
    "repo": "org/myapp",
    "branch": "main",
    "access_token_ref": "vault:github-pat"
}
```

> Note: Sensitive values (passwords, tokens) should reference an external secret store, not be stored in plaintext. The `_ref` suffix convention signals this.

---

### `agent_config`

Agent behaviour configuration. Single row for MVP; becomes per-project if Heimdall supports multiple monitored applications later.

```sql
CREATE TABLE agent_config (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model       TEXT NOT NULL DEFAULT 'claude-sonnet-4-6',  -- LLM model identifier
    mode        TEXT NOT NULL DEFAULT 'continuous',          -- "continuous" or "scheduled"
    schedule    TEXT,                                         -- Cron expression (nullable, only used if mode = "scheduled")
    system_prompt_override TEXT,                              -- Optional custom instructions appended to base prompt
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

Kept deliberately minimal. Agent tuning should happen through prompt engineering and tool configuration, not through a sprawling config table.

---

### `investigations`

The core table. This is Heimdall's primary output — what the agent detected, how it investigated, and what it concluded.

```sql
CREATE TABLE investigations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trigger_type    TEXT NOT NULL,           -- "anomaly", "user_query", "scheduled_check"
    trigger_source  TEXT,                    -- Which connection or input triggered it ("connection:uuid", "chat:uuid")
    summary         TEXT NOT NULL,           -- Agent's one-line summary of the investigation
    severity        TEXT NOT NULL DEFAULT 'info',  -- "info", "warning", "critical"
    status          TEXT NOT NULL DEFAULT 'open',  -- "open", "investigating", "resolved", "dismissed"
    context         JSONB NOT NULL,          -- The initial signal that triggered investigation
    findings        JSONB,                   -- Structured diagnosis: root cause, evidence, affected services
    tool_trace      JSONB,                   -- Ordered list of tool calls the agent made and their results
    resolution      TEXT,                    -- How it was resolved (nullable until resolved)
    started_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at     TIMESTAMPTZ              -- Nullable until status = "resolved"
);

CREATE INDEX idx_investigations_status ON investigations (status);
CREATE INDEX idx_investigations_severity ON investigations (severity);
CREATE INDEX idx_investigations_started_at ON investigations (started_at DESC);
```

**`context`** — captures what the agent was looking at when it decided to investigate:

```json
{
    "log_snippet": "ERROR 2026-02-19T03:22:14Z payments-service: connection refused to stripe-api.com",
    "pattern": "5 occurrences of connection_refused in 2 minutes",
    "connection_id": "uuid-of-log-source"
}
```

**`findings`** — the agent's structured diagnosis:

```json
{
    "root_cause": "Stripe API endpoint unreachable due to DNS resolution failure",
    "affected_services": ["payments-service", "checkout-service"],
    "evidence": [
        "17 connection_refused errors in 4-minute window",
        "DNS lookup for stripe-api.com returning NXDOMAIN",
        "No code changes deployed in last 12 hours"
    ],
    "recommendation": "External dependency issue — monitor for auto-recovery, check Stripe status page"
}
```

**`tool_trace`** — ordered audit trail of what the agent did:

```json
[
    {
        "tool": "search_logs",
        "input": {"query": "connection_refused", "timeframe": "30m"},
        "output_summary": "17 matches across payments-service and checkout-service",
        "timestamp": "2026-02-19T03:24:01Z"
    },
    {
        "tool": "query_database",
        "input": {"sql": "SELECT count(*) FROM payments WHERE status = 'failed' AND created_at > now() - interval '1 hour'"},
        "output_summary": "342 failed payments in last hour (baseline: ~12)",
        "timestamp": "2026-02-19T03:24:08Z"
    }
]
```

---

### `conversations`

Chat sessions between a user and the agent.

```sql
CREATE TABLE conversations (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    investigation_id  UUID REFERENCES investigations(id),  -- Nullable; links to investigation if chat spawned one
    title             TEXT,                                  -- Auto-generated or user-set
    messages          JSONB NOT NULL DEFAULT '[]'::jsonb,    -- [{role, content, timestamp}, ...]
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_conversations_investigation ON conversations (investigation_id) WHERE investigation_id IS NOT NULL;
```

**`messages`** as JSONB array:

```json
[
    {
        "role": "user",
        "content": "What happened with payments last night?",
        "timestamp": "2026-02-19T09:15:00Z"
    },
    {
        "role": "agent",
        "content": "I detected a payment processing outage between 03:22 and 03:48 UTC...",
        "tool_calls": ["search_logs", "query_database"],
        "timestamp": "2026-02-19T09:15:12Z"
    }
]
```

**Why JSONB array instead of a separate `messages` table:** Conversations are always loaded in full (you don't paginate a chat). A JSONB array keeps it simple — one read, one write, no joins. If message volumes per conversation ever become extreme, this can be normalised later.

**`investigation_id`** — nullable FK. When a user asks the agent a question and the agent decides it warrants a formal investigation, the conversation links to the resulting investigation. Not every conversation produces an investigation.

---

### `log_buffer`

Rolling 24-hour debug buffer. Stores raw ingested logs/events for internal observability — lets you verify that the agent is seeing what it should, and catch cases where a signal came in but didn't trigger an investigation.

```sql
CREATE TABLE log_buffer (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    connection_id   UUID NOT NULL REFERENCES connections(id),
    source_type     TEXT NOT NULL,           -- "server_logs", "db_events"
    severity        TEXT,                    -- Extracted severity if parseable (nullable)
    payload         JSONB NOT NULL,          -- Raw log/event as received
    ingested_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_log_buffer_ingested_at ON log_buffer (ingested_at DESC);
CREATE INDEX idx_log_buffer_connection ON log_buffer (connection_id, ingested_at DESC);
CREATE INDEX idx_log_buffer_severity ON log_buffer (severity, ingested_at DESC) WHERE severity IS NOT NULL;
```

**Retention:** A background goroutine (or Postgres scheduled job) runs periodically and deletes rows older than 24 hours:

```sql
DELETE FROM log_buffer WHERE ingested_at < now() - interval '24 hours';
```

This table is **not user-facing** — it's a developer/operator debugging tool. It should never grow beyond ~24h of data at any given time.

---

## Entity Relationships

```
connections
    │
    ├──< log_buffer          (connection_id → connections.id)
    │
    └──  referenced in       (investigations.trigger_source, investigations.context)

agent_config                  (standalone, single row)

investigations
    │
    └──< conversations        (investigation_id → investigations.id, nullable)
```

The schema is intentionally flat. There are only two real foreign keys:
- `log_buffer.connection_id` → `connections.id`
- `conversations.investigation_id` → `investigations.id`

Everything else is loosely coupled through JSONB references (e.g. `trigger_source` containing a connection UUID). This keeps migrations simple and avoids deep join chains.

---

## Query Layer — sqlc

Heimdall uses [sqlc](https://sqlc.dev/) for type-safe database access. You write SQL queries, sqlc generates Go code.

### How it works

1. Define schemas in migration files (`backend/migrations/`)
2. Write queries in `.sql` files (`backend/internal/db/queries/`)
3. Run `sqlc generate` — it produces Go structs and functions in `backend/internal/db/`

### Configuration — `backend/sqlc.yaml`

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/db/queries/"
    schema: "migrations/"
    gen:
      go:
        package: "db"
        out: "internal/db"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_empty_slices: true
        overrides:
          - db_type: "uuid"
            go_type: "github.com/google/uuid.UUID"
          - db_type: "jsonb"
            go_type: "encoding/json.RawMessage"
          - db_type: "timestamptz"
            go_type: "time.Time"
            nullable: true
            go_type: "*time.Time"
```

### Query files

Each table gets its own query file:

```
backend/internal/db/queries/
├── connections.sql
├── agent_config.sql
├── investigations.sql
├── conversations.sql
└── log_buffer.sql
```

**Example — `investigations.sql`:**

```sql
-- name: CreateInvestigation :one
INSERT INTO investigations (
    trigger_type, trigger_source, summary, severity, status, context
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetInvestigation :one
SELECT * FROM investigations WHERE id = $1;

-- name: ListOpenInvestigations :many
SELECT * FROM investigations
WHERE status IN ('open', 'investigating')
ORDER BY started_at DESC;

-- name: ListInvestigationsByDateRange :many
SELECT * FROM investigations
WHERE started_at >= $1 AND started_at <= $2
ORDER BY started_at DESC;

-- name: UpdateInvestigationFindings :exec
UPDATE investigations
SET findings = $2, tool_trace = $3, status = $4, resolved_at = $5
WHERE id = $1;

-- name: ResolveInvestigation :exec
UPDATE investigations
SET status = 'resolved', resolution = $2, resolved_at = now()
WHERE id = $1;

-- name: DismissInvestigation :exec
UPDATE investigations
SET status = 'dismissed', resolution = $2
WHERE id = $1;
```

**Example — `log_buffer.sql`:**

```sql
-- name: InsertLogEntry :exec
INSERT INTO log_buffer (connection_id, source_type, severity, payload)
VALUES ($1, $2, $3, $4);

-- name: ListRecentLogs :many
SELECT * FROM log_buffer
WHERE ingested_at > now() - interval '1 hour'
ORDER BY ingested_at DESC;

-- name: ListLogsByConnection :many
SELECT * FROM log_buffer
WHERE connection_id = $1 AND ingested_at > now() - interval '1 hour'
ORDER BY ingested_at DESC;

-- name: ListLogsBySeverity :many
SELECT * FROM log_buffer
WHERE severity = $1 AND ingested_at > now() - interval '1 hour'
ORDER BY ingested_at DESC;

-- name: PruneExpiredLogs :execrows
DELETE FROM log_buffer WHERE ingested_at < now() - interval '24 hours';
```

### Generated output

Running `sqlc generate` produces:

```
backend/internal/db/
├── db.go              # DBTX interface + Queries struct
├── models.go          # Go structs for each table (auto-generated from schema)
├── connections.sql.go
├── agent_config.sql.go
├── investigations.sql.go
├── conversations.sql.go
└── log_buffer.sql.go
```

The generated `models.go` contains structs like:

```go
type Investigation struct {
    ID            uuid.UUID
    TriggerType   string
    TriggerSource *string
    Summary       string
    Severity      string
    Status        string
    Context       json.RawMessage
    Findings      json.RawMessage
    ToolTrace     json.RawMessage
    Resolution    *string
    StartedAt     time.Time
    ResolvedAt    *time.Time
}
```

These structs are auto-generated — you never edit them manually. If you change the schema or queries, re-run `sqlc generate`.

### JSONB handling

sqlc maps JSONB columns to `json.RawMessage` (raw bytes). For typed access in application code, define typed structs separately and unmarshal when needed:

```go
// internal/agent/types.go — hand-written, not generated

type Findings struct {
    RootCause        string   `json:"root_cause"`
    AffectedServices []string `json:"affected_services"`
    Evidence         []string `json:"evidence"`
    Recommendation   string   `json:"recommendation"`
}

// Usage
var findings Findings
json.Unmarshal(investigation.Findings, &findings)
```

This keeps the database layer clean (raw bytes in, raw bytes out) while letting application code work with typed data.

---

## Migration Files

Based on this schema, the migrations in `backend/migrations/` should be:

```
001_create_connections.up.sql
001_create_connections.down.sql
002_create_agent_config.up.sql
002_create_agent_config.down.sql
003_create_investigations.up.sql
003_create_investigations.down.sql
004_create_conversations.up.sql
004_create_conversations.down.sql
005_create_log_buffer.up.sql
005_create_log_buffer.down.sql
```

---

## What's Not Modelled (intentionally)

- **Users / auth** — deferred. MVP assumes a single operator. When auth is added, it gets its own migration (`006_create_users`), and `investigations` / `conversations` gain a `user_id` FK.
- **Reports** — in the vision doc as a separate section, but for MVP an investigation with `status = 'resolved'` and populated `findings` *is* the report. A dedicated `reports` table can be added later if reports need their own lifecycle (drafts, approvals, exports).
- **Elephantasm memory** — managed externally by Elephantasm's API. Heimdall doesn't store memory locally; it reads/writes through the memory client.
