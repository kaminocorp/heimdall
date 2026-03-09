# Phase 8, Steps 2–3 — Monitoring State & Supporting Queries: Completion Notes

Reference: [Phase 8 Implementation Plan](../executing/phase-8-monitoring-mode.md)

---

## Step 2 — Migration 016: Monitoring State

| Action | File | Notes |
|--------|------|-------|
| Created | `backend/migrations/016_monitoring_state.up.sql` | Creates `monitoring_state` table |
| Created | `backend/migrations/016_monitoring_state.down.sql` | Drops table |

### Table created

**`monitoring_state`**
- `app_id` UUID PK (FK -> applications, CASCADE) — 1:1 with applications
- `last_monitored_at` TIMESTAMPTZ (default `now()`) — cursor position
- `updated_at` TIMESTAMPTZ

**Skip-on-resume semantics:** When monitoring is paused (`mode = 'off'`) and re-enabled, `ResetMonitoringCursor` sets `last_monitored_at = now()`. Logs that arrived while monitoring was off are skipped — no backfill, no surprise cost spikes.

---

## Step 3 — Supporting Queries

| Action | File | Queries |
|--------|------|---------|
| Created | `backend/internal/db/queries/monitoring.sql` | `GetMonitoringState`, `UpsertMonitoringState`, `ResetMonitoringCursor`, `ListActiveApplications`, `ListLogsSinceForApp` |

### Query details

**`ListActiveApplications`** — Three-way join: `applications` + `app_agent_config` + `connections` (via EXISTS subquery). Returns only apps where:
- `applications.status = 'active'`
- `app_agent_config.mode != 'off'`
- At least one connection with `status = 'active'`

Returns: `id`, `org_id`, `name`, `mode`, `schedule_interval_secs` — everything the monitor loop needs per app.

**`ListLogsSinceForApp`** — Fetches logs for an app's connections since a cursor timestamp. Joins `log_buffer` -> `connections` on `app_id`. Ordered by `ingested_at ASC` with a configurable `LIMIT` (batch size, default 200 in the monitor).

**`UpsertMonitoringState`** — Advances the cursor to a specific timestamp after processing a batch. Uses `ON CONFLICT` upsert.

**`ResetMonitoringCursor`** — Sets cursor to `now()`, used when toggling mode from off -> on.

### Generated Go code

| File | Status |
|------|--------|
| `backend/internal/db/monitoring.sql.go` | New — 5 query methods, `ListActiveApplicationsRow` and `UpsertMonitoringStateParams` structs |
| `backend/internal/db/models.go` | Regenerated — new `MonitoringState` model |

---

## Verification

- `sqlc generate` — clean
- `go build ./...` — clean
- `go vet ./...` — clean
