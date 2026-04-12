# Logs Feed Polish & Scheduled Investigations — Implementation Plan

Phase-by-phase implementation guide for two adjacent improvements:

1. **Logs feed polish** — the unified feed already exists but has bugs/UX issues (Phases 1–2).
2. **Scheduled investigations** — a new agent operating mode that runs cron-triggered prompts against pull-only connectors like Postgres and GitHub (Phases 3–4).

Each phase is a mergeable unit with a clear validation gate. File references use `path:line` notation against the codebase as of commit `b217afa` (v0.30.2).

---

## Context & Motivation

### Why phases 1–2 exist

The existing **Agent Log** page (`frontend/src/pages/AgentLogPage.vue`) is already a *unified feed* — its backend handler (`backend/internal/api/handlers/logs.go:35`) merges `log_buffer` (raw logs) and `agent_log` (agent observations) into one chronological stream, filterable by `source`, `severity`, and `connection_id`. The vision doc (`docs/vision.md:51-58`) explicitly describes this unified-feed behaviour as the design intent.

Two real bugs/UX issues remain:

- **`PruneExpiredLogs` is dead code.** The SQL function exists (`backend/internal/db/queries/log_buffer.sql:52-53`) and sqlc generates `(q *Queries) PruneExpiredLogs` (`backend/internal/db/log_buffer.sql.go:260`), but grep finds *zero* callers anywhere in `backend/`. Nothing runs it. `log_buffer` will grow unbounded in production.
- **Page name confuses users.** The page is called "Agent Log" which implies it only shows agent activity, even though it actually shows raw logs too. Users don't discover the "All sources" filter because they don't think to look on a page named "Agent Log" for raw logs.

### Why phases 3–4 exist

Today Heimdall has exactly two agent operating modes:

| Mode | Trigger | Data source | File |
|------|---------|-------------|------|
| Interactive | User opens WebSocket chat | Agent tool choice | `agent/loop.go` |
| Monitoring | Every N seconds (from `app_agent_config.schedule_interval_secs`) | `log_buffer` only (via `ListLogsSinceForApp`) | `agent/monitor.go` |

Neither mode can *autonomously probe* pull-only connectors like Postgres (`connectors/database/postgres.go`) or GitHub (`connectors/codebase/github.go`). Those connectors exist purely as **tools** the agent uses while reacting to a prompt or escalation — there is no mechanism that says "every 10 minutes, run a query against Postgres to check for slow queries" or "every hour, grep the GitHub repo for new TODOs and summarise".

Phases 3–4 add a third mode — **scheduled investigations** — implemented as cron-triggered calls into the existing `(a *Agent) RunMonitoring` entry point (`agent/loop.go:231`), which already handles provider resolution, tool registry, rate limiting, and `agent_log` emission. The new code is a scheduler that *calls* that entry point with arbitrary prompts, plus a storage layer for the schedules themselves.

### Naming decision: `investigation_schedules`

A table called `investigations` already exists (migration 003), storing incident investigation reports. To avoid confusion, the new table is **`investigation_schedules`**: each row is *"a schedule that produces an investigation"*. Investigation schedules produce entries in `agent_log` (and optionally in `investigations` / reports) when they fire.

### Decisions baked in

- Feed polish comes before the scheduler — it's ~½ day and unblocks "can I see my Supabase logs?" immediately.
- Scheduler uses `github.com/robfig/cron/v3` (de-facto Go cron parser; two years of stable releases).
- Scheduled investigations reuse `RunMonitoring` rather than introducing a parallel loop.
- Schedule scope is **app-level** — a single schedule can freely use any tool the agent knows about, which is more flexible than locking it to a specific connection.
- Phase 3 (MVP) uses a fixed 1-minute tick with `last_run_at`-based gating; Phase 4 introduces real cron parsing.
- Retention bumped from 24h → 48h per user request (one-char SQL change).
- Each phase must pass `make lint && make test` before merge.

---

## Phase 1 — Fix retention pruning & bump window to 48h

**Goal:** Wire up the dead `PruneExpiredLogs` function, extend retention to 48h, ensure `log_buffer` doesn't grow unbounded.

**Impact:** Prevents eventual production disk-fill. Zero UX change for users.

### Task 1.1 — Update the SQL to 48h

**File:** `backend/internal/db/queries/log_buffer.sql:52-53`

Change:
```sql
-- name: PruneExpiredLogs :execrows
DELETE FROM log_buffer WHERE ingested_at < now() - interval '24 hours';
```
to:
```sql
-- name: PruneExpiredLogs :execrows
DELETE FROM log_buffer WHERE ingested_at < now() - interval '48 hours';
```

Run `make sqlc-generate`. This regenerates `backend/internal/db/log_buffer.sql.go:256` with the new interval. No Go code change needed downstream.

### Task 1.2 — Add a pruner goroutine

**New file:** `backend/internal/agent/pruner.go`

```go
package agent

import (
    "context"
    "log/slog"
    "time"
)

const pruneInterval = 1 * time.Hour

// Prune runs a background loop that periodically deletes expired rows from log_buffer.
// The retention window is defined in the PruneExpiredLogs SQL query itself (48h).
func (a *Agent) Prune(ctx context.Context) {
    slog.Info("pruner goroutine started", "interval", pruneInterval)
    defer slog.Info("pruner goroutine stopped")

    ticker := time.NewTicker(pruneInterval)
    defer ticker.Stop()

    // Run once on startup so restarts don't delay the first prune by up to an hour.
    a.pruneTick(ctx)

    for {
        select {
        case <-ticker.C:
            a.pruneTick(ctx)
        case <-ctx.Done():
            return
        }
    }
}

func (a *Agent) pruneTick(ctx context.Context) {
    rows, err := a.queries.PruneExpiredLogs(ctx)
    if err != nil {
        slog.Error("pruner: failed to delete expired logs", "err", err)
        return
    }
    if rows > 0 {
        slog.Info("pruner: deleted expired logs", "rows", rows)
    }
}
```

### Task 1.3 — Wire pruner into agent startup

**File:** `backend/internal/agent/agent.go:75-88` (the `Start` method)

The existing `Start` method spawns exactly one goroutine (`a.Monitor(ctx)`). Add a second goroutine next to it:

```go
func (a *Agent) Start(ctx context.Context) {
    if a.cancel != nil {
        slog.Warn("agent already running, stopping previous instance before restart")
        a.Stop()
    }
    ctx, a.cancel = context.WithCancel(ctx)

    a.wg.Add(1)
    go func() {
        defer a.wg.Done()
        a.Monitor(ctx)
    }()

    a.wg.Add(1)
    go func() {
        defer a.wg.Done()
        a.Prune(ctx)
    }()
}
```

`Stop()` already waits on `a.wg`, so no changes needed there.

### Task 1.4 — Test

**New file:** `backend/internal/agent/pruner_test.go`

Test cases:
1. Insert 3 rows into `log_buffer` with `ingested_at` at `now() - 49h`, `now() - 47h`, `now()`. Call `pruneTick`. Assert exactly one row deleted (the 49h-old one).
2. Insert zero rows. Call `pruneTick`. Assert no error, no log spam.
3. Cancel the context. Assert `Prune(ctx)` returns cleanly within 1s.

Use the existing test DB setup pattern from `backend/internal/agent/monitor_test.go` (if present) or `handlers/*_test.go`. Integration tests skip automatically if `DATABASE_URL` is unset.

### Validation gate

- `make lint && make test` green
- Start a local backend with a seeded `log_buffer` containing a >48h row → observe log line `pruner: deleted expired logs rows=1` within one hour (or immediately on startup)
- Confirm `log_buffer` row count decreases

### Files touched
| File | Change |
|------|--------|
| `backend/internal/db/queries/log_buffer.sql` | 24h → 48h |
| `backend/internal/db/log_buffer.sql.go` | Regenerated by sqlc |
| `backend/internal/agent/pruner.go` | New — Prune loop |
| `backend/internal/agent/agent.go` | Wire pruner into Start |
| `backend/internal/agent/pruner_test.go` | New — unit tests |

**Scope estimate:** ~½ day including tests.

---

## Phase 2 — Rename "Agent Log" → "Activity"

**Goal:** Eliminate the mental-model mismatch where users don't realise the Agent Log page shows raw logs.

**Impact:** Pure UX — no data model changes, no backend API changes.

### Rationale

The page already shows raw logs *and* agent observations in one merged feed. The name "Agent Log" makes users assume it only shows agent output, so they don't look there for their Supabase/Vercel/webhook logs. Renaming to **"Activity"** (or "Feed") makes the unified nature obvious and matches the top-nav metaphor of other monitoring tools (Datadog uses "Events", Honeycomb uses "Traces", Grafana uses "Explore").

### Task 2.1 — Rename the page component

**File:** `frontend/src/pages/AgentLogPage.vue` → `frontend/src/pages/ActivityPage.vue`

Content changes:
- Line 34: `<h2>Agent Log</h2>` → `<h2>Activity</h2>`
- Line 35: `<p>Unified chronological feed of all system activity</p>` — keep as-is, it's already accurate

### Task 2.2 — Update the router

**File:** `frontend/src/router/index.ts` (path to verify — search for `AgentLogPage`)

Find the route importing `AgentLogPage` and:
- Update the import to `ActivityPage`
- Update the route `path` from `/agent-log` (or whatever it is) to `/activity`
- Update the route `name` to `activity`
- Add a `redirect` from the old path to `/activity` so any bookmarks and in-product links keep working

### Task 2.3 — Update the nav

Grep for `'Agent Log'` in `frontend/src/components/` and `frontend/src/layouts/` — there's likely a sidebar or top-nav component with a hardcoded label. Update to `'Activity'`.

### Task 2.4 — Update tests & stores

Grep for `AgentLogPage` and `agent-log` across:
- `frontend/src/stores/` — the `logs` store should be unaffected but double-check
- `frontend/src/**/*.test.ts` — any component tests referencing the old name
- `docs/vision.md` — it mentions "Agent Log" as a platform section; update to "Activity" and note in the changelog that the vision's concept of a unified feed maps to the renamed page

### Task 2.5 — Update the filter default label

**File:** `frontend/src/components/log/LogFilters.vue:28`

The "Raw logs" option label is fine, but consider relabeling "Agent activity" → "Agent observations" to distinguish it from the page-level "Activity" name. Minor polish.

### Task 2.6 — Changelog entry

**File:** `docs/changelog.md`

Add a new entry at the top:

```markdown
## 0.30.3 — Activity Feed Rename & Log Retention (2026-04-11)

Two small but high-visibility fixes to the logs experience:

- **Renamed "Agent Log" → "Activity".** The page has always been a unified feed (raw logs from connectors + agent observations), but the old name implied it only showed agent output. Users were missing their own logs because they didn't look on a page called "Agent Log". Old `/agent-log` URL redirects to `/activity`.
- **Fixed `log_buffer` retention.** The `PruneExpiredLogs` SQL function existed but was never called from anywhere — `log_buffer` was growing unbounded in production. Added a background pruner goroutine running hourly, and bumped the retention window from 24h → 48h per user request.
```

### Validation gate

- Old `/agent-log` URL redirects correctly to `/activity`
- Sidebar shows "Activity" not "Agent Log"
- Filter dropdowns still work
- `cd frontend && npm run test` passes
- Manual smoke: after saving a Supabase connection, navigate to Activity and see raw log rows appear within 30 seconds

### Files touched
| File | Change |
|------|--------|
| `frontend/src/pages/AgentLogPage.vue` → `ActivityPage.vue` | Rename + header text |
| `frontend/src/router/index.ts` | Route rename + redirect |
| `frontend/src/layouts/*` or sidebar component | Nav label |
| `frontend/src/components/log/LogFilters.vue` | Minor label polish |
| `docs/vision.md` | Reflect rename |
| `docs/changelog.md` | New 0.30.3 entry |

**Scope estimate:** ~2 hours.

---

## Phase 3 — Scheduled investigations MVP (fixed 1-minute tick)

**Goal:** Prove the end-to-end loop works — a stored schedule fires, runs an agent investigation, writes to `agent_log` — using a dumb 1-minute polling tick instead of a real cron parser.

**Impact:** First cut of the new agent mode. No cron expressions yet, just a "run every N seconds" integer to start.

### Why do MVP first instead of cron parsing first?

The *interesting* unknowns here are:
1. Does `RunMonitoring` actually work well with arbitrary prompts (it was designed for classifier-flagged logs)?
2. How does the UI feel for listing/creating/editing schedules?
3. What does the `agent_log` output look like when the trigger is a schedule vs. an escalation?

None of those unknowns require cron parsing. A fixed interval is the minimum viable substrate to answer them. Cron parsing is a ~30-minute change once the rest is proven.

### Task 3.1 — Migration

**New file:** `backend/migrations/023_investigation_schedules.up.sql`

```sql
CREATE TABLE investigation_schedules (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_id          UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    prompt          TEXT NOT NULL,              -- the investigation prompt sent to the agent
    interval_secs   INTEGER NOT NULL,           -- Phase 3: plain integer interval
    cron_expr       TEXT,                       -- Phase 4: populated once cron parsing lands
    enabled         BOOLEAN NOT NULL DEFAULT true,
    last_run_at     TIMESTAMPTZ,
    last_status     TEXT,                       -- 'success' | 'error' | null
    last_error      TEXT,                       -- error message from last failed run
    last_summary    TEXT,                       -- short blurb for UI display
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_investigation_schedules_app_enabled
    ON investigation_schedules (app_id) WHERE enabled;

CREATE INDEX idx_investigation_schedules_next_run
    ON investigation_schedules (last_run_at) WHERE enabled;

ALTER TABLE investigation_schedules ENABLE ROW LEVEL SECURITY;

CREATE POLICY investigation_schedules_org ON investigation_schedules
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN users u ON u.org_id = a.org_id
            WHERE u.id = app_current_user_id()
        )
    );
```

**New file:** `backend/migrations/023_investigation_schedules.down.sql`

```sql
DROP TABLE IF EXISTS investigation_schedules;
```

Note: `interval_secs` and `cron_expr` coexist — `interval_secs` is the Phase 3 path; Phase 4 adds `cron_expr` support alongside it (not replacing). Keeps migrations forward-only.

### Task 3.2 — sqlc queries

**New file:** `backend/internal/db/queries/investigation_schedules.sql`

```sql
-- name: ListEnabledSchedules :many
SELECT * FROM investigation_schedules
WHERE enabled = true
ORDER BY last_run_at ASC NULLS FIRST;

-- name: ListSchedulesByApp :many
SELECT * FROM investigation_schedules
WHERE app_id = $1
ORDER BY created_at DESC;

-- name: GetSchedule :one
SELECT * FROM investigation_schedules
WHERE id = $1;

-- name: CreateSchedule :one
INSERT INTO investigation_schedules (app_id, name, prompt, interval_secs, enabled)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateSchedule :one
UPDATE investigation_schedules
SET name = $2, prompt = $3, interval_secs = $4, enabled = $5, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteSchedule :exec
DELETE FROM investigation_schedules WHERE id = $1;

-- name: MarkScheduleRun :exec
UPDATE investigation_schedules
SET last_run_at = now(),
    last_status = $2,
    last_error  = $3,
    last_summary = $4,
    updated_at  = now()
WHERE id = $1;
```

Run `make sqlc-generate`. Sanity-check the generated Go types.

### Task 3.3 — Scheduler goroutine

**New file:** `backend/internal/agent/scheduler.go`

```go
package agent

import (
    "context"
    "log/slog"
    "time"
    "unicode/utf8"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgtype"

    "github.com/hejijunhao/heimdall/backend/internal/db"
)

const (
    schedulerTickInterval = 1 * time.Minute
    schedulerRunTimeout   = 5 * time.Minute
)

// InvestigationScheduler runs the scheduled-investigations loop.
// Every tick, it lists enabled schedules and fires those whose interval has elapsed.
func (a *Agent) InvestigationScheduler(ctx context.Context) {
    slog.Info("investigation scheduler started", "tick", schedulerTickInterval)
    defer slog.Info("investigation scheduler stopped")

    ticker := time.NewTicker(schedulerTickInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            a.schedulerTick(ctx)
        case <-ctx.Done():
            return
        }
    }
}

func (a *Agent) schedulerTick(ctx context.Context) {
    schedules, err := a.queries.ListEnabledSchedules(ctx)
    if err != nil {
        slog.Error("scheduler: list enabled schedules failed", "err", err)
        return
    }

    now := time.Now()
    for _, s := range schedules {
        if !shouldFire(s, now) {
            continue
        }
        a.runScheduledInvestigation(ctx, s)
    }
}

// shouldFire decides whether a schedule is due based on last_run_at and interval_secs.
func shouldFire(s db.InvestigationSchedule, now time.Time) bool {
    if !s.LastRunAt.Valid {
        return true // never run — fire immediately
    }
    elapsed := now.Sub(s.LastRunAt.Time)
    return elapsed >= time.Duration(s.IntervalSecs)*time.Second
}

func (a *Agent) runScheduledInvestigation(ctx context.Context, s db.InvestigationSchedule) {
    runCtx, cancel := context.WithTimeout(ctx, schedulerRunTimeout)
    defer cancel()

    // Resolve a user for agent_log attribution — reuse the same helper as Monitor.
    app, err := a.queries.GetApplicationByID(runCtx, s.AppID)
    if err != nil {
        slog.Error("scheduler: load app failed", "err", err, "schedule_id", s.ID)
        a.markRunError(ctx, s.ID, "load app: "+err.Error())
        return
    }
    userID, err := a.resolveOrgUser(runCtx, app.OrgID)
    if err != nil {
        slog.Error("scheduler: resolve user failed", "err", err, "schedule_id", s.ID)
        a.markRunError(ctx, s.ID, "resolve user: "+err.Error())
        return
    }

    appConfig, err := a.queries.GetAppAgentConfig(runCtx, s.AppID)
    if err != nil {
        slog.Warn("scheduler: app agent config missing, using default", "err", err, "app_id", s.AppID)
        appConfig = db.AppAgentConfig{AppID: s.AppID, Model: DefaultModelID}
    }

    // Rate-limit the same way monitor.go does.
    if err := a.limiter.Wait(runCtx); err != nil {
        slog.Warn("scheduler: rate limit wait interrupted", "schedule_id", s.ID, "err", err)
        return
    }

    // Reuse the monitoring entry point — it handles provider resolution,
    // tools, system prompt, and returns (assessment, severity).
    assessment, severity := a.RunMonitoring(runCtx, userID, appConfig, s.Prompt)

    summary := assessment
    if utf8.RuneCountInString(summary) > 200 {
        summary = string([]rune(summary)[:200]) + "..."
    }

    // Emit to agent_log with a distinct entry_type so the UI can render it differently.
    a.EmitLogWithSeverity(ctx, userID, nil, "scheduled_investigation", summary,
        map[string]any{
            "schedule_id":   s.ID,
            "schedule_name": s.Name,
            "app_id":        s.AppID,
            "assessment":    assessment,
            "auto_severity": severity,
        },
        severity,
    )

    a.markRunSuccess(ctx, s.ID, summary)
}

func (a *Agent) markRunSuccess(ctx context.Context, id uuid.UUID, summary string) {
    _ = a.queries.MarkScheduleRun(ctx, db.MarkScheduleRunParams{
        ID:          id,
        LastStatus:  pgtype.Text{String: "success", Valid: true},
        LastError:   pgtype.Text{},
        LastSummary: pgtype.Text{String: summary, Valid: true},
    })
}

func (a *Agent) markRunError(ctx context.Context, id uuid.UUID, errMsg string) {
    _ = a.queries.MarkScheduleRun(ctx, db.MarkScheduleRunParams{
        ID:          id,
        LastStatus:  pgtype.Text{String: "error", Valid: true},
        LastError:   pgtype.Text{String: errMsg, Valid: true},
        LastSummary: pgtype.Text{},
    })
}
```

Note: `GetApplicationByID` may not exist in the current sqlc queries — if not, add it in the same PR or use an existing query that returns `org_id` for a given app.

### Task 3.4 — Wire scheduler into agent startup

**File:** `backend/internal/agent/agent.go:75-88` (the `Start` method, which Phase 1 already touched)

Add a third goroutine next to Monitor and Prune:

```go
a.wg.Add(1)
go func() {
    defer a.wg.Done()
    a.InvestigationScheduler(ctx)
}()
```

### Task 3.5 — HTTP handlers

**New file:** `backend/internal/api/handlers/investigation_schedules.go`

Implement five handlers:

| Method | Path | Handler | Auth |
|--------|------|---------|------|
| GET    | `/api/apps/{appId}/schedules`        | `ListSchedules`  | `authorizeApp` |
| POST   | `/api/apps/{appId}/schedules`        | `CreateSchedule` | `authorizeApp` |
| PATCH  | `/api/apps/{appId}/schedules/{id}`   | `UpdateSchedule` | `authorizeApp` + schedule ownership check |
| DELETE | `/api/apps/{appId}/schedules/{id}`   | `DeleteSchedule` | `authorizeApp` + schedule ownership check |
| POST   | `/api/apps/{appId}/schedules/{id}/run` | `RunScheduleNow` | `authorizeApp` + schedule ownership check |

All five use the existing `authorizeApp` helper to validate the app belongs to the caller's org. The PATCH/DELETE/RunNow handlers additionally assert `schedule.app_id == appID` to prevent path-mismatch confusion.

**RunScheduleNow** is just a thin wrapper that calls `a.runScheduledInvestigation(ctx, s)` synchronously and returns the `agent_log` entry ID. Useful for "test this schedule now" in the UI.

**Request shape** (create/update):

```json
{
  "name": "Check slow queries",
  "prompt": "Query pg_stat_statements for queries taking >1s in the last hour. Summarize the worst offenders.",
  "interval_secs": 600,
  "enabled": true
}
```

**Validation** on create/update:
- `name`: non-empty, ≤100 chars
- `prompt`: non-empty, ≤5000 chars (avoid prompt-injection of enormous blobs)
- `interval_secs`: ≥60 (1 minute minimum — matches the scheduler tick) and ≤86400 (1 day maximum)

Return 400 with a clear error message on validation failure.

### Task 3.6 — Router wiring

**File:** `backend/internal/api/router.go` (or wherever routes are mounted)

Mount the five routes under the existing `/api/apps/{appId}` subrouter. Follow the same pattern used for `connections`, `agent/config`, etc.

### Task 3.7 — Tests

**New file:** `backend/internal/agent/scheduler_test.go`

Test cases:
1. `shouldFire` with `LastRunAt = nil` → true.
2. `shouldFire` with `LastRunAt = now - 30s, interval = 60s` → false.
3. `shouldFire` with `LastRunAt = now - 61s, interval = 60s` → true.
4. Integration: insert a schedule with `interval_secs=60` and `last_run_at=now-2m`. Run `schedulerTick`. Assert (a) `RunMonitoring` was called — use a stub agent, (b) an `agent_log` row with `entry_type='scheduled_investigation'` exists, (c) `last_run_at` advanced.
5. `runScheduledInvestigation` propagates errors into `last_status='error'` and `last_error` without panicking.

**New file:** `backend/internal/api/handlers/investigation_schedules_test.go`

Standard handler tests for create/update/delete/list with auth failures, validation failures, ownership mismatches. Mirror the shape of `backend/internal/api/handlers/applications_test.go`.

### Validation gate

- `make lint && make test` green
- Integration test proves the loop: create schedule → wait 1 tick → row in `agent_log`
- `POST /schedules/{id}/run` manual test via curl hits `agent_log`
- Start dev server, create a schedule via curl, verify it fires within ~60s and produces an entry visible on the Activity page

### Files touched
| File | Change |
|------|--------|
| `backend/migrations/023_investigation_schedules.up.sql` | New |
| `backend/migrations/023_investigation_schedules.down.sql` | New |
| `backend/internal/db/queries/investigation_schedules.sql` | New — sqlc queries |
| `backend/internal/db/*_schedules.sql.go`, `models.go` | Regenerated by sqlc |
| `backend/internal/agent/scheduler.go` | New — scheduler loop |
| `backend/internal/agent/agent.go` | Third goroutine in Start |
| `backend/internal/api/handlers/investigation_schedules.go` | New — 5 handlers |
| `backend/internal/api/router.go` | Mount new routes |
| `backend/internal/agent/scheduler_test.go` | New |
| `backend/internal/api/handlers/investigation_schedules_test.go` | New |

**Scope estimate:** ~2 days. The migration and sqlc are half an hour; scheduler.go is a day including tests; handlers are half a day.

---

## Phase 4 — Real cron parsing + frontend UI

**Goal:** Replace the Phase-3 integer interval with a real cron expression, and add a frontend page so users can manage schedules without curl.

**Impact:** Turns the MVP into a usable feature.

### Task 4.1 — Add cron dependency

**File:** `backend/go.mod`

```
go get github.com/robfig/cron/v3@latest
```

Rationale: `robfig/cron/v3` is the de-facto Go cron parser. Two years of stable releases, 12k GitHub stars, no external deps, correct DST handling. Alternative `gorhill/cronexpr` is less maintained.

### Task 4.2 — Replace `interval_secs` gating with cron

**File:** `backend/internal/agent/scheduler.go`

Replace `shouldFire` with a cron-aware version:

```go
func shouldFire(s db.InvestigationSchedule, now time.Time) bool {
    if !s.LastRunAt.Valid {
        return true
    }

    // Prefer cron_expr if set; fall back to interval_secs for backward compat.
    if s.CronExpr.Valid && s.CronExpr.String != "" {
        parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
        sched, err := parser.Parse(s.CronExpr.String)
        if err != nil {
            slog.Error("scheduler: invalid cron expression",
                "schedule_id", s.ID, "cron", s.CronExpr.String, "err", err)
            return false
        }
        next := sched.Next(s.LastRunAt.Time)
        return !next.After(now)
    }

    // Legacy interval_secs path.
    elapsed := now.Sub(s.LastRunAt.Time)
    return elapsed >= time.Duration(s.IntervalSecs)*time.Second
}
```

The parser is constructed per-call for simplicity. If this loop ever shows up in profiles, cache parsed schedules keyed by cron string (tiny sync.Map).

### Task 4.3 — Update handlers to accept cron

**File:** `backend/internal/api/handlers/investigation_schedules.go`

Update request validation to accept either `interval_secs` (numeric, kept for the preset buttons) or `cron_expr` (string, advanced mode). Exactly one must be provided. Validate `cron_expr` using `cron.NewParser(...).Parse(expr)` at create/update time — return 400 if invalid, so bad expressions can never land in the DB.

### Task 4.4 — sqlc update

**File:** `backend/internal/db/queries/investigation_schedules.sql`

Update `CreateSchedule` and `UpdateSchedule` to take both `interval_secs` and `cron_expr` as parameters (cron_expr nullable). Regenerate.

### Task 4.5 — Frontend: types & store

**New file:** `frontend/src/types/schedule.ts`

```ts
export interface InvestigationSchedule {
  id: string
  app_id: string
  name: string
  prompt: string
  interval_secs: number | null
  cron_expr: string | null
  enabled: boolean
  last_run_at: string | null
  last_status: 'success' | 'error' | null
  last_error: string | null
  last_summary: string | null
  created_at: string
  updated_at: string
}
```

**New file:** `frontend/src/stores/schedules.ts`

Pinia setup store following the existing pattern in `stores/connections.ts` and `stores/logs.ts`:

```ts
export const useSchedulesStore = defineStore('schedules', () => {
  const schedules = ref<InvestigationSchedule[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function fetch(appId: string) { /* GET /api/apps/{appId}/schedules */ }
  async function create(appId: string, input: ScheduleInput) { /* POST */ }
  async function update(appId: string, id: string, input: ScheduleInput) { /* PATCH */ }
  async function remove(appId: string, id: string) { /* DELETE */ }
  async function runNow(appId: string, id: string) { /* POST /run */ }

  return { schedules, loading, error, fetch, create, update, remove, runNow }
})
```

### Task 4.6 — Frontend: schedules page

**New file:** `frontend/src/pages/SchedulesPage.vue`

Sections:
1. **Header** — "Scheduled Investigations" + "Add schedule" button
2. **List** — one card per schedule showing name, cron/interval (humanised via `cronstrue` or a small local helper), enabled toggle, last run timestamp + status indicator, last summary text (truncated), "Run now" button, "Edit" button, "Delete" button
3. **Empty state** — helpful copy explaining what schedules are ("Run investigations on a cron. Good for: slow query checks, GitHub TODO audits, daily health summaries.")
4. **Create/Edit modal** — name, prompt (textarea), interval preset buttons (Every 5 minutes, Every 15 minutes, Hourly, Daily) OR advanced cron input (toggle), enabled toggle

**New file:** `frontend/src/components/schedules/ScheduleCard.vue`
**New file:** `frontend/src/components/schedules/ScheduleModal.vue`

Use existing design-system components: `BaseSelect`, `BaseInput` or whatever `AgentConfigPage.vue` uses.

### Task 4.7 — Cron preset helpers

**New file:** `frontend/src/utils/cron-presets.ts`

```ts
export const cronPresets = [
  { label: 'Every 5 minutes',  value: '*/5 * * * *'  },
  { label: 'Every 15 minutes', value: '*/15 * * * *' },
  { label: 'Every hour',       value: '0 * * * *'    },
  { label: 'Every 6 hours',    value: '0 */6 * * *'  },
  { label: 'Daily at 9am UTC', value: '0 9 * * *'    },
]

export function humanizeCron(expr: string): string {
  const preset = cronPresets.find(p => p.value === expr)
  if (preset) return preset.label
  return expr // advanced mode — show raw
}
```

(Or pull in `cronstrue` if the user wants arbitrary expressions rendered nicely. ~20KB dep — only worth it if users will actually write custom expressions.)

### Task 4.8 — Router + nav entry

**File:** `frontend/src/router/index.ts`

Add a route `/schedules` → `SchedulesPage`. Add to the sidebar nav.

### Task 4.9 — Frontend tests

**New file:** `frontend/src/stores/schedules.test.ts`

Standard vitest store tests following the pattern in the existing stores' test files. Mock the axios client.

**New file:** `frontend/src/pages/SchedulesPage.test.ts` (optional — component tests for the empty state + render)

### Task 4.10 — Documentation

**File:** `docs/vision.md`

Add a new section under "Platform Sections" (or adjacent to Agent Chat) for **Scheduled Investigations** with a couple of example use cases.

**File:** `docs/changelog.md`

Add a `0.31.0 — Scheduled Investigations` entry describing the new agent mode, the cron scheduler, and the UI.

### Validation gate

- Create a schedule via the UI with `*/5 * * * *` + a prompt like "Query the production DB for any deadlocked transactions in the last 10 minutes"
- Observe it fire within 5 minutes
- `last_run_at`, `last_status='success'`, `last_summary` all populated
- Corresponding row visible on the Activity page with `entry_type='scheduled_investigation'`
- Disabling the schedule from the UI prevents further runs
- "Run now" fires the schedule synchronously and returns a summary
- Invalid cron expression in the modal shows an error before the request is sent
- All tests green: `make lint && make test`

### Files touched
| File | Change |
|------|--------|
| `backend/go.mod`, `go.sum` | Add `robfig/cron/v3` |
| `backend/internal/agent/scheduler.go` | `shouldFire` uses cron parser |
| `backend/internal/db/queries/investigation_schedules.sql` | CreateSchedule/UpdateSchedule take `cron_expr` |
| `backend/internal/db/*_schedules.sql.go` | Regenerated by sqlc |
| `backend/internal/api/handlers/investigation_schedules.go` | Validate + persist `cron_expr` |
| `frontend/src/types/schedule.ts` | New |
| `frontend/src/stores/schedules.ts` | New |
| `frontend/src/pages/SchedulesPage.vue` | New |
| `frontend/src/components/schedules/ScheduleCard.vue` | New |
| `frontend/src/components/schedules/ScheduleModal.vue` | New |
| `frontend/src/utils/cron-presets.ts` | New |
| `frontend/src/router/index.ts` | New route |
| `frontend/src/layouts/*` or sidebar | New nav entry |
| `frontend/src/stores/schedules.test.ts` | New |
| `docs/vision.md` | Scheduled Investigations section |
| `docs/changelog.md` | 0.31.0 entry |

**Scope estimate:** ~3 days. Backend cron swap is ~2 hours. Frontend page + modal + store + tests is the bulk. The humanised cron UX can eat time if over-polished — timebox it.

---

## Cross-cutting concerns

### Security

- All handlers use `authorizeApp` — no cross-org leakage
- RLS policy on `investigation_schedules` mirrors the one on `app_agent_config` — defense in depth
- Prompt length capped at 5000 chars to bound LLM cost per run
- `interval_secs` floor of 60s and cron floor of "every minute" prevents runaway scheduling (a schedule can't fire more than once per scheduler tick anyway)
- Rate limiter already gates every `RunMonitoring` call (`monitor.go:162`) — scheduled investigations inherit this for free

### Cost controls

- A user with 10 schedules running every 5 minutes = 120 LLM calls/hour. Reasonable for paid tiers; flag in PR description for review.
- Consider adding a per-app schedule count cap (e.g. 10 active schedules per app) in a follow-up if abuse becomes a concern.

### Observability

- Add a Prometheus metric `scheduled_investigations_runs_total{status="success|error"}` in `scheduler.go` — mirror the pattern in `monitor.go:139-140`
- Log `slog.Info("scheduler: ran investigation", "schedule_id", s.ID, "duration_ms", ...)` on every run for dashboard grep-ability

### Schema drift guard

Phases 3 and 4 both add migrations (023 in Phase 3, a follow-up if Phase 4 needs column changes). Given the 0.30.2 incident where migrations 020–022 were merged but not deployed, **Phase 3's PR description must explicitly include**: "Deployment step: run `make migrate-up` against production before releasing the backend". Even better if `docs/executing/schema-drift-check.md` is implemented first — then it becomes automatic.

### Backward compatibility

The `interval_secs` column introduced in Phase 3 is *not* removed in Phase 4 — it's kept as an alternative to `cron_expr` so the handler logic remains simple ("either-or"). This means Phase 3 schedules created before Phase 4 ships continue to work without migration.

---

## Execution order & dependencies

```
Phase 1 ─┐
         ├─→ Phase 2 (independent of 3–4, pure frontend polish)
         │
         └─→ Phase 3 ──→ Phase 4
              (requires agent.go change from Phase 1 — third goroutine slot)
```

Recommended merge order:
1. **Phase 1** — small, low-risk, unblocks production growth concern immediately
2. **Phase 2** — small, user-facing, can ship the same day
3. **Phase 3** — meaningful but self-contained; merge behind a feature flag if there's concern about rolling out a new goroutine
4. **Phase 4** — polish pass that turns MVP into real UX

Phases 1 and 2 can go out as a single `0.30.3` release. Phase 3 is `0.31.0-alpha` (API-only, no UI). Phase 4 is `0.31.0` (full release).

---

## Open questions to resolve before starting

These are design calls to lock in before touching code. Marked with 🔶 because they shape the data model or user experience.

1. 🔶 **Should scheduled investigations produce `reports` (incident records) or only `agent_log` entries?** The current plan writes to `agent_log` only, which means they show up on the Activity feed but don't appear in the Reports page. Reports are richer (timeline + findings + historical context) but overkill for a "every 5 minutes check pg_stats" schedule. Recommendation: `agent_log` only for Phase 3; add an optional `create_report: boolean` flag to the schedule in Phase 4 for users who want report-level output.

2. 🔶 **Should `RunScheduleNow` share the rate limiter with automatic runs?** If yes, a burst of manual "run now" clicks could starve the monitor loop. If no, users can trivially bypass rate limiting. Recommendation: yes, share the limiter. UI disables the button for ~5s after a click.

3. 🔶 **Do we need per-schedule model override?** Today `app_agent_config.model` applies to the whole app. A user might want "use Haiku for the cheap every-5-minute check and Opus for the daily deep analysis". Recommendation: not in Phase 3/4. Inherit from app config. Revisit if real users ask.

4. 🔶 **What happens to a schedule if its app is paused (`status != 'active'`)?** Recommendation: scheduler skips it (add an `app.status = 'active'` filter to `ListEnabledSchedules`). Mirrors how `ListActiveApplications` works in `monitor.go`.

5. 🔶 **Should `Phase 2` rename happen as part of `0.30.3` or separately?** Renaming routes affects bookmarks and any external docs. Recommendation: ship the rename with a 301 redirect from the old path. Low risk because the page is internal-only (behind auth).

---

## Out of scope for this plan

- **Rule-based agent log entries** — mentioned in `docs/vision.md:56` but distinct from scheduled investigations. A rule is "if X then Y" (reactive). A schedule is "every N, check and report" (proactive). Both deserve to exist eventually; keep the data models separate.
- **Investigation templates / a library of common prompts** — nice to have, but Phase 4's modal with a blank prompt textarea is sufficient for v1.
- **Schedule result history beyond `last_*` fields** — right now we only remember the most recent run. A full history table (`schedule_runs`) is a clean addition once we know users want trend views. Don't build it speculatively.
- **Cron expression DST handling edge cases** — `robfig/cron/v3` does the right thing here. Don't over-engineer.
- **Per-org global rate limiting for scheduled investigations** — the existing per-agent limiter is sufficient for v1.
