# Phase 6 — Agent Log

Unified chronological feed combining raw ingested log entries with agent observations. The agent now emits structured log entries during its tool-use loop, and the `GET /api/logs` endpoint merges both data sources into a single timeline with source filtering.

---

## Database: Agent Log Table

### `backend/migrations/010_create_agent_log.up.sql`

New `agent_log` table for agent-emitted observations:

```sql
CREATE TABLE agent_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    entry_type      TEXT NOT NULL,      -- "tool_call", "tool_result", "observation"
    summary         TEXT NOT NULL,
    detail          JSONB,
    severity        TEXT,
    conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

Two indexes: `(user_id, created_at DESC)` for the primary paginated query, `(entry_type)` for future type-based filtering.

`conversation_id` uses `ON DELETE SET NULL` — agent log entries survive conversation deletion, preserving the audit trail.

### `backend/internal/db/queries/agent_log.sql`

| Query | Purpose |
|-------|---------|
| `InsertAgentLog` | Writes an agent log entry (`:one` — returns inserted row) |
| `ListAgentLogByUser` | Paginated user-scoped list, ordered by `created_at DESC` |
| `ListAgentLogByUserAndType` | Same with `entry_type` filter |
| `CountAgentLogByUser` | Total count for pagination |

Ran `sqlc generate` — `agent_log.sql.go` and `models.go` regenerated. New `AgentLog` struct in models.

---

## Agent: Emit Hooks

### `backend/internal/agent/emit.go`

New `EmitLog` method on `Agent`:

```go
func (a *Agent) EmitLog(ctx context.Context, userID uuid.UUID, conversationID *uuid.UUID,
    entryType, summary string, detail map[string]any)
```

**Fire-and-forget design** — marshals detail to JSON, inserts via `InsertAgentLog`, and logs warnings on failure without propagating errors. This prevents observability writes from degrading the agent's primary work.

### `backend/internal/agent/loop.go`

`RunConversation` signature updated to accept `conversationID *uuid.UUID`:

```go
func (a *Agent) RunConversation(ctx context.Context, userID uuid.UUID,
    conversationID *uuid.UUID, history []Message, input string) (string, error)
```

Three emit points in the tool-use loop:

| When | Entry type | Summary | Detail |
|------|-----------|---------|--------|
| Tool dispatched | `tool_call` | `Called search_logs` | `{ "tool": "search_logs", "input": {...} }` |
| Tool succeeds | `tool_result` | `search_logs returned results` | `{ "tool": "...", "result_preview": "..." }` |
| Tool fails | `tool_result` | `search_logs failed: ...` | `{ "tool": "...", "error": "..." }` |
| Final response | `observation` | First 200 chars of agent response | `nil` |

`RunLoop` delegates to `RunConversation` with `nil` conversationID — backward compatible, but agent entries from `POST /api/agent/run` won't have a conversation link.

---

## Backend: Unified Log Endpoint

### `backend/internal/api/handlers/logs.go`

Rewrote `ListLogs` to merge two data sources into a unified feed.

**Unified response shape:**

```go
type unifiedLogEntry struct {
    ID           string          `json:"id"`
    Source       string          `json:"source"`        // "raw" or "agent"
    Timestamp    string          `json:"timestamp"`
    Severity     *string         `json:"severity"`
    SourceType   string          `json:"source_type"`   // "webhook" / "tool_call" / "tool_result" / "observation"
    ConnectionID *string         `json:"connection_id"` // null for agent entries
    Summary      string          `json:"summary"`
    Detail       json.RawMessage `json:"detail"`
}
```

**Query parameter:** `source` — accepts `all` (default), `raw`, or `agent`.

**Merge strategy:**
1. When `source=all`, both `log_buffer` and `agent_log` are queried with the same `limit`/`offset`
2. Results are merged, sorted by timestamp descending, and trimmed to `limit`
3. Total count is the sum of both sources

When `connection_id` filter is set, `source` is forced to `raw` — agent entries have no connection association.

Converter functions `rawToUnified` and `agentToUnified` normalize both source types into the common shape.

### `backend/internal/api/handlers/chat.go`

Updated `RunConversation` call to pass `&convID`:

```go
response, err := s.Agent.RunConversation(ctx, userID, &convID, history, msg.Content)
```

Agent log entries emitted during chat now link to the conversation.

---

## Frontend: Source Filtering & Agent Entry Display

### `frontend/src/types/log.ts`

Updated `LogEntry` to match the unified response shape:

```typescript
export interface LogEntry {
  id: string
  source: 'raw' | 'agent'
  timestamp: string
  severity: string | null
  source_type: string
  connection_id: string | null
  summary: string
  detail: Record<string, unknown> | null
}
```

### `frontend/src/api/logs.ts`

Added `source` parameter to `listLogs`:

```typescript
export function listLogs(params?: {
  severity?: string
  connection_id?: string
  source?: string    // new — "all", "raw", or "agent"
  limit?: number
  offset?: number
})
```

### `frontend/src/stores/logs.ts`

Added source state management:

- `source` ref — `'raw' | 'agent' | 'all'` (default `'all'`)
- `setSource(s)` — updates source and resets pagination
- `fetchLogs` now passes `source` to the API

### `frontend/src/components/log/LogFilters.vue`

Added source filter dropdown:

```html
<select v-model="source" @change="handleFilter">
  <option value="all">All sources</option>
  <option value="raw">Raw logs</option>
  <option value="agent">Agent activity</option>
</select>
```

Emits `source` alongside existing `severity` and `connection_id` filters.

### `frontend/src/components/log/LogEntry.vue`

Visual differentiation for agent entries:

- Purple border and background tint (`border-purple-200 bg-purple-50/50`) for agent entries
- `AGENT` badge in purple for agent-sourced entries
- Human-readable labels: `tool_call` → "Tool Call", `tool_result` → "Tool Result", `observation` → "Observation"

### `frontend/src/components/log/LogFeed.vue`

Updated filter event type to include `source`:

```typescript
filter: [filters: { severity?: string; connection_id?: string; source?: string }]
```

### `frontend/src/pages/AgentLogPage.vue`

Wired source filter through:

- `handleFilter` extracts `source` from filter event, calls `logsStore.setSource`
- Remaining filters (`severity`, `connection_id`) passed through as before

---

## Files Changed (7 modified, 4 new)

| File | Change |
|------|--------|
| `backend/migrations/010_create_agent_log.up.sql` | New — agent log table with indexes |
| `backend/migrations/010_create_agent_log.down.sql` | New — down migration |
| `backend/internal/db/queries/agent_log.sql` | New — insert, list, count queries |
| `backend/internal/db/agent_log.sql.go` | New — sqlc-generated Go code |
| `backend/internal/db/models.go` | Regenerated — new `AgentLog` struct |
| `backend/internal/agent/emit.go` | New — fire-and-forget `EmitLog` method |
| `backend/internal/agent/loop.go` | Added `conversationID` param, emit hooks at tool_call/tool_result/observation |
| `backend/internal/api/handlers/chat.go` | Passes `&convID` to `RunConversation` |
| `backend/internal/api/handlers/logs.go` | Rewritten — unified feed merging raw + agent logs with source filtering |
| `frontend/src/types/log.ts` | Rewritten — unified `LogEntry` type |
| `frontend/src/api/logs.ts` | Added `source` parameter |
| `frontend/src/stores/logs.ts` | Added `source` state, `setSource` action |
| `frontend/src/components/log/LogFilters.vue` | Added source filter dropdown |
| `frontend/src/components/log/LogEntry.vue` | Agent entry styling (purple badge/border), entry type labels |
| `frontend/src/components/log/LogFeed.vue` | Updated filter event type |
| `frontend/src/pages/AgentLogPage.vue` | Wired source filter to store |

---

## Verification

### 1. Build & vet
```bash
cd backend && go vet ./... && go build ./cmd/heimdall
```

### 2. Run migration
```bash
migrate -path backend/migrations -database "$DATABASE_URL" up
```

### 3. Agent log emission (via chat)
```bash
wscat -c "ws://localhost:8080/ws/chat?token=$JWT"
> {"content": "Search for recent log entries"}
# Agent uses search_logs tool → agent_log gets tool_call, tool_result, observation entries
```

### 4. Unified feed — all sources
```bash
curl http://localhost:8080/api/logs -H "Authorization: Bearer $JWT"
# Returns mixed raw + agent entries sorted by timestamp
```

### 5. Source filtering
```bash
curl "http://localhost:8080/api/logs?source=agent" -H "Authorization: Bearer $JWT"
# Returns only agent entries (tool_call, tool_result, observation)

curl "http://localhost:8080/api/logs?source=raw" -H "Authorization: Bearer $JWT"
# Returns only raw ingested log entries
```

### 6. Frontend
- Navigate to Agent Log page
- "All sources" shows mixed feed — agent entries have purple AGENT badge
- Select "Agent activity" → only agent entries displayed
- Select "Raw logs" → only webhook-ingested entries
- Connection filter forces raw-only (agent entries have no connection)
