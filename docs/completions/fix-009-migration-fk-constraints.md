# Fix 009 — Migration Foreign Key / Constraint Gaps

**Review item:** #9 (Should Fix)
**Status:** Done

## Problem

Three constraint gaps in migration files:

1. **`agent_config` has no singleton constraint.** Multiple rows allowed but app expects exactly one.
2. **`conversations.investigation_id` FK has no `ON DELETE` clause.** Deleting an investigation causes FK violations.
3. **`log_buffer.connection_id` FK has no `ON DELETE CASCADE`.** Deleting a connection leaves undeletable log entries.

## Changes

### `backend/migrations/002_create_agent_config.up.sql`
- Changed `id` from `UUID PRIMARY KEY DEFAULT gen_random_uuid()` to `INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1)`, enforcing exactly one row.

### `backend/internal/db/queries/agent_config.sql`
- Updated `UpsertAgentConfig` to explicitly insert `id = 1` (matching the singleton constraint).
- Switched from positional params (`$1`) to `EXCLUDED.*` in the `ON CONFLICT` update clause (also addresses review item #19).

### `backend/migrations/004_create_conversations.up.sql`
- Added `ON DELETE SET NULL` to `investigation_id` FK.

### `backend/migrations/005_create_log_buffer.up.sql`
- Added `ON DELETE CASCADE` to `connection_id` FK.

## Verification
- `go build ./...` passes clean.
