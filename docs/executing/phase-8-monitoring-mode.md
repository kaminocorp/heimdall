# Phase 8 — Monitoring Mode: Implementation Plan

Reference: [Post-MVP Roadmap](./post-mvp-roadmap.md) | [Vision](../vision.md) | [Changelog](../changelog.md)

---

## Goal

Make Heimdall watch systems autonomously. Logs flow in continuously, pass through a **deterministic classification pipeline** (powered by [Lumber](https://github.com/kaminocorp/lumber)), and only questionable or problematic logs are escalated to the LLM agent. If an application is running normally (BAU), the LLM is never called.

This phase also introduces the **Organization** and **Application** data models — the structural foundation for multi-app, multi-user monitoring.

---

## Architecture Overview

```
Logs ingested (webhook / future connectors)
   ↓
┌─────────────────────────────────────┐
│  Lumber Classification Pipeline     │  ← Deterministic, local, no API calls
│  Embed → Classify → Severity gate   │
└─────────────┬───────────────────────┘
              │
     ┌────────┴────────┐
     │                  │
  SAFE (BAU)       FLAGGED (warning+)
     │                  │
  Log to DB,        Log to DB,
  no LLM call       escalate to Agent
                        │
                  ┌─────┴─────┐
                  │  Agent    │  ← Claude API call
                  │  Assess   │
                  │  + Tools  │
                  └─────┬─────┘
                        │
                  Write to agent_log
                  (monitoring entry)
```

**Key principle:** The LLM is the expensive, slow path. Lumber is the cheap, fast filter. A well-behaved production app in steady state should generate zero LLM calls.

---

## Current State

| Component | Status | Notes |
|-----------|--------|-------|
| `agent/monitor.go` | Stub | Logs startup/shutdown, awaits ctx cancellation |
| `agent/agent.go` `Start()` | Stub | Logs "agent started", never called from `main.go` |
| `agent_config` table | Exists | Global singleton (id=1) — will be replaced by per-app config |
| `log_buffer` table | Exists | User-scoped, has `ingested_at` index |
| `agent_log` table | Exists | Supports arbitrary `entry_type` strings |
| `investigations` table | Exists | Has trigger/findings/status fields, mostly unused |
| Agent loop (`RunLoop`) | Working | Tool-use loop with `search_logs` + `query_database` |
| Org / App models | **Missing** | No multi-app structure exists |
| Lumber integration | **Missing** | No deterministic log classification |

---

## Implementation Steps

### Step 1 — Database: Organizations & Applications

Introduce the foundational data model for multi-app monitoring. Users belong to orgs. Apps belong to orgs. Agent config moves from a global singleton to per-application.

**Migration 014 — `organizations` and `applications`:**

```sql
-- Organizations
CREATE TABLE organizations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    slug       TEXT NOT NULL UNIQUE,              -- URL-friendly identifier, modifiable by owner/admins
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_organizations_slug ON organizations(slug);

-- Link users to orgs (a user belongs to exactly one org)
ALTER TABLE users ADD COLUMN org_id UUID REFERENCES organizations(id) ON DELETE SET NULL;
CREATE INDEX idx_users_org_id ON users(org_id);
-- org_id is nullable: users without an org haven't completed onboarding yet

-- Applications (belong to an org)
CREATE TABLE applications (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id     UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'active',   -- active, paused, archived
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_applications_org_id ON applications(org_id);

-- Connections now belong to applications (not directly to users)
ALTER TABLE connections ADD COLUMN app_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE;
CREATE INDEX idx_connections_app_id ON connections(app_id);
-- Note: existing connections table is wiped (clean break), so NOT NULL is safe
```

**Onboarding flow:** After signup, users land on an onboarding screen where they either create a new org (providing name + slug) or join an existing org (by invite — future phase). Until they have an org, they can't create apps or connections. The `users.org_id` column is nullable to represent this "not yet onboarded" state.

**Per-application agent config (migration 015):**

```sql
-- Agent config moves from global singleton to per-application
CREATE TABLE app_agent_config (
    app_id                 UUID PRIMARY KEY REFERENCES applications(id) ON DELETE CASCADE,
    model                  TEXT NOT NULL DEFAULT 'claude-sonnet-4-6',
    mode                   TEXT NOT NULL DEFAULT 'continuous',  -- continuous, periodic, off
    schedule_interval_secs INTEGER NOT NULL DEFAULT 60,         -- polling interval in seconds
    system_prompt_override TEXT,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE app_agent_config ENABLE ROW LEVEL SECURITY;
-- RLS: users can access configs for apps in their org
CREATE POLICY app_agent_config_org ON app_agent_config
    FOR ALL USING (
        app_id IN (
            SELECT a.id FROM applications a
            JOIN users u ON u.org_id = a.org_id
            WHERE u.id = app_current_user_id()
        )
    );
```

**Default values:** When a new application is created, insert a default `app_agent_config` row automatically (via application-layer logic or a trigger).

**Schedule field:** Uses a simple integer (seconds) rather than cron expressions. Both are interval-based scheduling — cron just adds day/hour specificity we don't need yet. A simple seconds field is easier to validate, display in the UI, and reason about. We can add cron later if needed.

---

### Step 2 — Database: Monitoring State (Per-Application)

Track where the monitor left off for each application, so we only process new logs.

**Migration 016 — `monitoring_state`:**

```sql
CREATE TABLE monitoring_state (
    app_id             UUID PRIMARY KEY REFERENCES applications(id) ON DELETE CASCADE,
    last_monitored_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

**Skip-on-resume semantics:** When monitoring is paused (`mode = 'off'`) and then re-enabled, the cursor is reset to `now()`. Logs that arrived while monitoring was off are skipped — no backfill. This is simple and avoids surprise cost spikes from processing a large backlog.

**sqlc queries:**

```sql
-- GetMonitoringState
SELECT * FROM monitoring_state WHERE app_id = $1;

-- UpsertMonitoringState
INSERT INTO monitoring_state (app_id, last_monitored_at, updated_at)
VALUES ($1, $2, now())
ON CONFLICT (app_id) DO UPDATE
SET last_monitored_at = $2, updated_at = now();

-- ResetMonitoringCursor (called when mode changes from off → on)
INSERT INTO monitoring_state (app_id, last_monitored_at, updated_at)
VALUES ($1, now(), now())
ON CONFLICT (app_id) DO UPDATE
SET last_monitored_at = now(), updated_at = now();
```

---

### Step 3 — Database: Supporting Queries

New sqlc queries needed for the monitoring pipeline:

```sql
-- ListActiveApplications: all apps with mode != 'off' and at least one active connection
SELECT a.id, a.org_id, a.name, c.mode, c.schedule_interval_secs
FROM applications a
JOIN app_agent_config c ON c.app_id = a.id
WHERE a.status = 'active'
  AND c.mode != 'off'
  AND EXISTS (
      SELECT 1 FROM connections conn
      WHERE conn.app_id = a.id AND conn.status = 'active'
  );

-- ListLogsSinceForApp: fetch logs for an app's connections since cursor
SELECT lb.* FROM log_buffer lb
JOIN connections c ON c.id = lb.connection_id
WHERE c.app_id = $1 AND lb.ingested_at > $2
ORDER BY lb.ingested_at ASC
LIMIT $3;
```

---

### Step 4 — Lumber Integration (In-Process)

Add Lumber as a Go library dependency. It runs inside the Heimdall process — no microservice, no network hop. The ONNX model (~23MB) is downloaded during Docker build.

```bash
go get github.com/kaminocorp/lumber
```

**Dockerfile addition:**

```dockerfile
# Download Lumber ONNX model during build
RUN curl -fsSL -o /app/models/model_quantized.onnx \
    https://huggingface.co/onnx-community/mdbr-leaf-mt-ONNX/resolve/main/onnx/model_quantized.onnx && \
    curl -fsSL -o /app/models/vocab.txt \
    https://huggingface.co/onnx-community/mdbr-leaf-mt-ONNX/resolve/main/vocab.txt
```

**New file: `backend/internal/agent/classifier.go`**

```go
package agent

import (
    "github.com/kaminocorp/lumber"
)

type LogClassifier struct {
    engine *lumber.Lumber
}

func NewLogClassifier(modelDir string) (*LogClassifier, error) {
    l, err := lumber.New(lumber.WithModelDir(modelDir))
    if err != nil {
        return nil, err
    }
    return &LogClassifier{engine: l}, nil
}

func (c *LogClassifier) Close() { c.engine.Close() }

// ClassifyBatch classifies logs and returns only those that need LLM attention.
// "Safe" logs (REQUEST/success, SYSTEM/health_check, etc.) are filtered out.
func (c *LogClassifier) ClassifyBatch(logs []LogEntry) (flagged []ClassifiedLog, safe int) {
    // 1. Extract text from each log's payload
    // 2. Call c.engine.ClassifyBatch(texts)
    // 3. Filter: only keep events with Type in {ERROR, PERFORMANCE, ACCESS(failures)}
    //    or severity >= warning
    // 4. Return flagged logs (with classification metadata) + count of safe logs
}
```

**Severity gate rules (deterministic):**

| Lumber Category | Action |
|----------------|--------|
| `ERROR/*` | Always escalate to LLM |
| `PERFORMANCE/*` | Always escalate to LLM |
| `ACCESS/login_failure`, `ACCESS/auth_failure` | Escalate |
| `ACCESS/login_success`, `ACCESS/session_expired` | Safe — skip |
| `REQUEST/server_error` | Escalate |
| `REQUEST/success`, `REQUEST/redirect` | Safe — skip |
| `REQUEST/client_error` | Safe — skip (4xx is usually caller's problem) |
| `REQUEST/slow_request` | Escalate |
| `DEPLOY/*` | Escalate (deploys are always worth noting) |
| `SYSTEM/resource_alert` | Escalate |
| `SYSTEM/health_check`, `SYSTEM/process_lifecycle` | Safe — skip |
| `SCHEDULED/cron_failed` | Escalate |
| `SCHEDULED/cron_started`, `SCHEDULED/cron_completed` | Safe — skip |
| `DATA/migration` | Escalate |
| `DATA/query_executed`, `DATA/replication` | Safe — skip |
| `UNCLASSIFIED` (confidence < 0.5) | Escalate (uncertain = worth checking) |

These rules are hardcoded initially. They can be made configurable per-app in a future phase.

---

### Step 5 — Monitor Loop Implementation

Rewrite `agent/monitor.go` with the full pipeline. The monitor runs as a long-lived goroutine, processing all active applications concurrently.

**Pipeline per monitoring cycle:**

```
1. List active applications (mode != 'off', has active connections)
2. For each app (concurrently, bounded by semaphore):
   a. Read monitoring cursor (last_monitored_at)
   b. Fetch new logs since cursor (batch limit: 200)
   c. Classify via Lumber
   d. If all safe → log heartbeat, advance cursor, done
   e. If flagged logs exist → format + send to LLM agent
   f. Write agent assessment to agent_log
   g. Advance cursor
3. Sleep for shortest interval among active apps (or a global tick)
```

**Concurrency model:**

```go
func (a *Agent) Monitor(ctx context.Context) {
    log.Info("monitoring mode started")
    defer log.Info("monitoring mode stopped")

    sem := make(chan struct{}, 10) // max 10 concurrent app monitors
    ticker := time.NewTicker(15 * time.Second) // global tick rate
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            apps := a.listActiveApplications(ctx)
            var wg sync.WaitGroup
            for _, app := range apps {
                // Check if this app's interval has elapsed
                if !a.shouldMonitor(ctx, app) {
                    continue
                }
                wg.Add(1)
                sem <- struct{}{} // acquire
                go func(app Application) {
                    defer wg.Done()
                    defer func() { <-sem }() // release
                    a.monitorApp(ctx, app)
                }(app)
            }
            wg.Wait()
        case <-ctx.Done():
            return
        }
    }
}
```

**`shouldMonitor` checks:**
- Load app's `monitoring_state.last_monitored_at`
- Compare against `app_agent_config.schedule_interval_secs`
- If `now - last_monitored_at < interval`, skip this cycle
- For `continuous` mode, always monitor (every tick)

**`monitorApp` flow:**

```go
func (a *Agent) monitorApp(ctx context.Context, app Application) {
    // 1. Get cursor
    state := a.getMonitoringState(ctx, app.ID)

    // 2. Fetch new logs
    logs := a.getLogsSinceForApp(ctx, app.ID, state.LastMonitoredAt, 200)
    if len(logs) == 0 {
        return
    }

    // 3. Classify with Lumber
    flagged, safeCount := a.classifier.ClassifyBatch(logs)

    // 4. Log heartbeat (monitor is alive, processed N logs, M safe)
    a.emitHeartbeat(ctx, app, len(logs), safeCount, len(flagged))

    // 5. If flagged logs exist, escalate to LLM
    if len(flagged) > 0 {
        input := formatFlaggedLogsForAgent(flagged)
        response, severity := a.RunMonitoring(ctx, app, input)
        a.EmitLog(ctx, app.OrgOwnerUserID, "monitoring", response, nil, severity, nil)
    }

    // 6. Advance cursor
    a.upsertMonitoringState(ctx, app.ID, logs[len(logs)-1].IngestedAt)
}
```

---

### Step 6 — Monitoring System Prompt

Create a monitoring-specific system prompt in `agent/prompt.go`. This prompt is used only when the monitor escalates flagged logs to the LLM.

```go
const monitoringSystemPrompt = `You are Heimdall, an autonomous AI monitoring agent.
You are currently in MONITORING MODE — you are not chatting with a user.

A deterministic classification pipeline has flagged the following log entries as
potentially problematic. The pipeline uses semantic classification to filter out
routine, healthy logs — what you're seeing has already been identified as unusual
or erroneous.

Your job:
1. Assess the flagged logs. Are they genuinely concerning, or false positives from the classifier?
2. If concerning, use your tools (search_logs, query_database) to investigate further context.
3. Produce a concise assessment:
   - What happened and why it matters
   - Severity: info / warning / error / critical
   - Whether further investigation is recommended

Be precise. Do NOT fabricate issues or speculate beyond what the data shows.
False alarms erode trust.`
```

---

### Step 7 — `RunMonitoring` Agent Method

Add a new method to the agent, similar to `RunLoop` but tailored for monitoring:

```go
func (a *Agent) RunMonitoring(ctx context.Context, app Application, flaggedLogs string) (string, string) {
    // 1. Use monitoringSystemPrompt (not interactive prompt)
    // 2. Load app-specific agent config (model, system_prompt_override)
    // 3. Run Claude tool-use loop (same machinery as RunLoop)
    // 4. Do NOT persist to a conversation — monitoring is sessionless
    // 5. Parse severity from response (or use Lumber's classification)
    // 6. Emit tool_call/tool_result entries to agent_log as usual
    // 7. Return (assessment text, severity)
}
```

---

### Step 8 — Wire Startup & Shutdown

**`agent/agent.go` — add lifecycle fields and methods:**

```go
type Agent struct {
    queries    *db.Queries
    client     *anthropic.Client
    config     *config.Config
    classifier *LogClassifier  // NEW
    cancel     context.CancelFunc
    wg         sync.WaitGroup
}

func (a *Agent) Start(ctx context.Context) {
    ctx, a.cancel = context.WithCancel(ctx)
    a.wg.Add(1)
    go func() {
        defer a.wg.Done()
        a.Monitor(ctx)
    }()
    log.Info("agent started, monitoring goroutine spawned")
}

func (a *Agent) Stop() {
    a.cancel()
    a.wg.Wait()
    a.classifier.Close()
    log.Info("agent stopped")
}
```

**`cmd/heimdall/main.go` — wire into startup sequence:**

```go
// After creating agent:
ag.Start(ctx)

// In shutdown sequence (before pool.Close):
ag.Stop()
```

**Lumber model files:** The ONNX model (~23MB) needs to be available at runtime. Options:
- Download during Docker build (`make download-model` in Dockerfile)
- Bundle in the repo under `backend/models/` (small enough at 23MB)
- Download on first start with a health check

---

### Step 9 — Agent Log Conventions

No schema change needed. New `entry_type` values:

| entry_type | Source | Meaning |
|------------|--------|---------|
| `observation` | Interactive chat | Agent's response in chat |
| `tool_call` | Both | Agent invoked a tool |
| `tool_result` | Both | Tool returned a result |
| `monitoring` | Monitor loop | Agent's assessment of flagged logs |
| `heartbeat` | Monitor loop | Monitor processed N logs, all safe (no LLM call) |

Heartbeat entries are lightweight — they confirm the monitor is running without cluttering the log with full agent responses. They include metadata like `{logs_processed: 47, safe: 45, flagged: 2}`.

---

### Step 10 — Frontend Changes

Minimal frontend work — surface the new data model and monitoring state.

**10a — Application selector (if multi-app):**
- Navigation or dropdown to switch between applications
- Scope all existing pages (connections, logs, agent log) to the selected app

**10b — Dashboard monitoring card:**
- Show current mode (continuous/periodic/off) per app
- Last heartbeat timestamp
- Logs processed / flagged ratio (from heartbeat entries)

**10c — Agent config page:**
- Now loads/saves per-app config via `app_agent_config`
- Schedule interval field (seconds input or preset dropdown: 30s, 1m, 5m, 15m)
- Mode toggle: continuous / periodic / off

**10d — Agent log page:**
- New badge for `monitoring` and `heartbeat` entry types
- No other changes needed — entries appear in the existing feed

**New/updated API endpoints:**

```
GET    /api/orgs                       → list user's orgs
POST   /api/orgs                       → create org
GET    /api/orgs/:id/apps              → list apps in org
POST   /api/orgs/:id/apps              → create app (+ default agent config)
GET    /api/apps/:id/agent/config      → get app agent config
PUT    /api/apps/:id/agent/config      → update app agent config
GET    /api/apps/:id/monitoring/status  → { mode, interval, last_monitored_at, is_running }
```

---

### Step 11 — Tests

| Test | What it covers |
|------|----------------|
| `classifier_test.go` | Lumber integration: known ERROR log → flagged, known 200 OK → safe |
| `monitor_test.go` — unit | Loop mechanics: skip when off, respect interval, concurrent app processing |
| `monitor_test.go` — integration | Insert logs → run monitor → verify heartbeat + monitoring entries in agent_log |
| `RunMonitoring` test | Mock Claude API, verify monitoring prompt used, tool dispatch works |
| Org/App handler tests | CRUD for organizations and applications |
| App agent config tests | Get/update per-app config, default creation on app create |

---

## Implementation Order

```
Step 1:  Migration — organizations, applications, app_agent_config
Step 2:  Migration — monitoring_state
Step 3:  Queries — active apps, logs since cursor
         ↓
Step 4:  Lumber integration — classifier.go + severity gate
Step 5:  Monitor loop — monitor.go rewrite
Step 6:  Monitoring system prompt
Step 7:  RunMonitoring agent method
         ↓
Step 8:  Wire startup/shutdown
         ↓
Step 9:  (Convention only — no code)
Step 10: Frontend — app model, config, dashboard, log badges
         ↓
Step 11: Tests
```

**Suggested PR breakdown:**

| PR | Steps | Description |
|----|-------|-------------|
| PR 1 | 1–3 | Data model: orgs, apps, configs, migrations, sqlc queries |
| PR 2 | 4 | Lumber integration + classifier with unit tests |
| PR 3 | 5–8 | Monitor loop, agent method, startup wiring |
| PR 4 | 10 | Frontend: app model, config page, dashboard, log badges |
| PR 5 | 11 | Remaining integration tests |

---

## Estimated File Changes

| Action | File |
|--------|------|
| Create | `backend/migrations/014_organizations_applications.sql` |
| Create | `backend/migrations/015_app_agent_config.sql` |
| Create | `backend/migrations/016_monitoring_state.sql` |
| Create | `backend/internal/db/queries/organizations.sql` |
| Create | `backend/internal/db/queries/applications.sql` |
| Create | `backend/internal/db/queries/app_agent_config.sql` |
| Create | `backend/internal/db/queries/monitoring.sql` |
| Create | `backend/internal/agent/classifier.go` |
| Create | `backend/internal/agent/classifier_test.go` |
| Create | `backend/internal/agent/monitor_test.go` |
| Create | `backend/internal/api/handlers/organizations.go` |
| Create | `backend/internal/api/handlers/applications.go` |
| Create | `backend/internal/api/handlers/monitoring.go` |
| Create | `frontend/src/stores/organizations.ts` |
| Create | `frontend/src/stores/applications.ts` |
| Create | `frontend/src/api/organizations.ts` |
| Create | `frontend/src/api/applications.ts` |
| Rewrite | `backend/internal/agent/monitor.go` |
| Edit | `backend/internal/agent/agent.go` (lifecycle, classifier field) |
| Edit | `backend/internal/agent/prompt.go` (monitoring prompt) |
| Edit | `backend/internal/agent/loop.go` (RunMonitoring method) |
| Edit | `backend/cmd/heimdall/main.go` (Start/Stop, classifier init) |
| Edit | `backend/internal/api/router.go` (org/app/monitoring routes) |
| Edit | `frontend/src/pages/DashboardPage.vue` (monitoring card) |
| Edit | `frontend/src/pages/AgentConfigPage.vue` (per-app config, interval field) |
| Edit | `frontend/src/pages/AgentLogPage.vue` (monitoring/heartbeat badges) |
| Regenerate | `backend/internal/db/*.sql.go` (sqlc generate) |

---

## Decisions (Resolved)

1. **Org creation flow.** Users are prompted to create a new org or join an existing one (by invite/slug). Orgs have slugs (modifiable by owner/admins — settings UI comes later). A user belongs to exactly one org (no multi-org for now).

2. **Existing data migration.** Clean break — all existing data is test data and can be wiped. No backfill migration needed.

3. **Lumber model distribution.** In-process library mode — `go get github.com/kaminocorp/lumber` imported directly into the Heimdall binary. The ONNX model (~23MB) is downloaded during Docker build via `make download-model` (or `curl` from HuggingFace). No separate microservice needed at current scale. Revisit if classification throughput becomes a bottleneck.

4. **Connection ↔ App assignment.** Connections must belong to an application (`app_id` required). The existing `user_id` scoping on connections is replaced by the org → app → connection hierarchy. All connection CRUD adapts accordingly.

5. **Lumber confidence threshold.** Start with the default 0.5 threshold. Tune later based on real-world false-positive/negative rates.
