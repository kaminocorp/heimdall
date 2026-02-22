# impl — sqlc generate

## What

Ran `sqlc generate` to produce type-safe Go database access code from the existing migration schemas and query files.

## Why

The migration files (`backend/migrations/`), query files (`backend/internal/db/queries/`), and sqlc config (`backend/sqlc.yaml`) were all in place from the scaffolding phase, but `sqlc generate` had never been run — no generated Go files existed in `backend/internal/db/`.

## What changed

### New generated files (all auto-generated, do not edit)

- `backend/internal/db/db.go` — `DBTX` interface + `Queries` struct
- `backend/internal/db/models.go` — Go structs: `AgentConfig`, `Connection`, `Conversation`, `Investigation`, `LogBuffer`
- `backend/internal/db/connections.sql.go` — 6 queries (List, Get, Create, Update, UpdateStatus, Delete)
- `backend/internal/db/agent_config.sql.go` — 2 queries (Get, Upsert)
- `backend/internal/db/investigations.sql.go` — 7 queries (Create, Get, ListOpen, ListByDateRange, UpdateFindings, Resolve, Dismiss)
- `backend/internal/db/conversations.sql.go` — 4 queries (Create, Get, List, UpdateMessages)
- `backend/internal/db/log_buffer.sql.go` — 5 queries (Insert, ListRecent, ListByConnection, ListBySeverity, PruneExpired)

### Dependencies added to `go.mod`

- `github.com/google/uuid v1.6.0` — UUID type for generated structs
- `github.com/jackc/pgx/v5 v5.8.0` — pgx driver (DBTX interface, pgtype, pgconn)

### Notes

- Nullable JSONB columns (`findings`, `tool_trace`) generate as `[]byte` rather than `json.RawMessage`. This is functionally identical since `json.RawMessage` is a `[]byte` alias.
- Nullable text columns use `pgtype.Text`, nullable UUIDs use `pgtype.UUID` — both from pgx/v5.
- Non-nullable `TIMESTAMPTZ` maps to `time.Time`, nullable maps to `*time.Time` (per sqlc.yaml overrides).

## Verification

- `sqlc generate` — clean exit, no warnings
- `go build ./...` — passes from `backend/`
