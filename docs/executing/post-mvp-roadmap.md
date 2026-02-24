# Heimdall — Post-MVP Roadmap

Reference: [MVP Roadmap](../executing/mvp-roadmap.md) · [Blueprint](./blueprint.md) · [Vision](../vision.md)

---

## Where We Are

The MVP (Phases 1–6) is complete. Heimdall can:
- Authenticate users via Supabase JWT
- Manage connections (CRUD, user-scoped)
- Ingest logs via webhook
- Run an AI agent with tool-use (search_logs, query_database)
- Chat with the agent via WebSocket (multi-turn, persistent conversations)
- Display a unified agent log (raw + agent activity, filterable by source)

**Build status:** Backend (`go vet ./...`) and frontend (`vue-tsc --noEmit`) both pass cleanly.

### What's Still Stubbed

These scaffolding-era stubs exist but have no real logic:
- `backend/internal/agent/monitor.go` — monitoring loop
- `backend/internal/agent/tools_codebase.go` — codebase search tool
- `backend/internal/agent/tools_memory.go` — memory recall tools
- `backend/internal/reports/generator.go` — report generation
- `backend/internal/connectors/codebase/github.go` — GitHub connector
- `backend/internal/connectors/logs/syslog.go` — syslog connector
- `backend/internal/memory/client.go` — Elephantasm client
- `frontend/src/pages/DashboardPage.vue` — dashboard (placeholder text)
- `frontend/src/pages/LoginPage.vue` — login (hardcoded token)
- `frontend/src/pages/AgentConfigPage.vue` — config (read-only, no edit UI)
- `frontend/src/pages/ReportsPage.vue` — reports (empty list, no data)

### Test Coverage

Minimal — three stub connector tests. No handler, agent, store, or integration tests.

---

## Roadmap

Phases are ordered by impact and dependency. Each phase is self-contained and shippable.

---

### Phase 7 — Polish & Hardening

**Why first:** The MVP works end-to-end, but rough edges in the frontend and missing tests make it fragile. Polish before building more features.

#### Scope

- **Dashboard page** — system overview with status cards: connection count & health, recent agent activity summary, last conversation timestamp, log ingestion rate
- **Agent config editing** — form UI on AgentConfigPage for model selection, mode toggle, system prompt override. Calls existing `PUT /api/agent/config`
- **Login page** — wire to Supabase Auth SDK (`@supabase/supabase-js`). Real email/password login, token refresh, logout
- **Error boundaries** — global error handler in Vue app, toast/notification system for transient errors
- **Loading states** — skeleton loaders for pages that fetch data on mount
- **Responsive layout** — ensure sidebar and pages work on tablet/mobile widths

#### Tests

- Backend handler tests (connections CRUD, logs, agent run, webhooks)
- Agent loop unit test (mock Claude API, verify tool dispatch)
- Frontend store tests (logs, connections, auth)

---

### Phase 8 — Monitoring Mode

**Why next:** The core vision — Heimdall watches systems 24/7, not just when a user opens chat.

#### Scope

- **Monitor goroutine** — long-running goroutine started on server boot. Polls `log_buffer` on a configurable interval (default: 60s). Passes recent logs to the agent with a monitoring-specific system prompt
- **Anomaly detection prompt** — system prompt variant that asks the agent to assess whether recent logs contain anomalies, errors, or deviations from normal patterns
- **Agent log emission** — monitoring observations written to `agent_log` with `entry_type = 'monitoring'`
- **Schedule configuration** — `agent_config.mode` supports `continuous` (always-on goroutine), `periodic` (cron-like interval), `off` (manual only). Editable from Phase 7's config UI
- **Graceful lifecycle** — monitor respects context cancellation on server shutdown

#### Database

- Add `last_monitored_at` column to `agent_config` or a separate monitoring state table to track cursor position (avoid re-processing old logs)

#### Requires

- Phase 7 (config editing UI for schedule control)

---

### Phase 9 — Report Generation

**Why next:** Completes the final vision platform section. Gives the agent's investigations a permanent, shareable output.

#### Scope

- **Investigation model** — reuse existing `investigations` table. An investigation is triggered by the monitoring loop or manually by a user in chat ("Investigate the payment errors from last night")
- **Report generator** — `reports.Generator.Generate(investigationID)` fetches investigation data, related agent log entries, and conversation history. Formats into structured sections:
  - Timeline of events
  - What was investigated (tools called, queries run)
  - Findings and diagnosis
  - Historical context (if memory is available)
  - Severity assessment
- **Report persistence** — new `reports` table (`id`, `user_id`, `investigation_id`, `title`, `content JSONB`, `severity`, `created_at`)
- **API endpoints** — `GET /api/reports` (list, user-scoped), `GET /api/reports/:id` (detail)
- **Frontend** — ReportsPage lists reports with severity badges, ReportDetail renders structured sections

#### Requires

- Phase 8 (monitoring mode generates investigations worth reporting on)

---

### Phase 10 — Elephantasm Long-Term Memory

**Why next:** Memory transforms Heimdall from a stateless responder into an agent that learns. Each incident makes it better at the next.

Reference: [Elephantasm Integration Plan](./elephantasm-integration.md)

#### Scope

- **Go HTTP client** — `internal/memory/client.go` wraps the Elephantasm API (record event, query memories, get lessons)
- **Event recording** — agent emits events to Elephantasm after investigations: what happened, what was found, what the diagnosis was
- **Memory recall tools** — wire `recall_similar_incidents` and `recall_lessons` tools to the Elephantasm client. Agent can query past incidents during investigation
- **System prompt enrichment** — monitoring prompt includes relevant lessons from memory, giving the agent evolving context
- **Memory viewer** — optional frontend page showing what the agent remembers (events → memories → lessons chain)

#### Requires

- Elephantasm API access / account
- Phase 8 (monitoring loop generates events worth remembering)

---

### Phase 11 — GitHub Codebase Connector

**Why next:** During investigation, the agent often needs to understand *why* code behaves a certain way. Codebase search closes the investigation loop.

#### Scope

- **GitHub connector** — `internal/connectors/codebase/github.go` authenticates with a personal access token, queries the GitHub Search API and Contents API
- **`search_codebase` tool** — agent can search for files, functions, or patterns in connected repositories. Returns file paths, relevant code snippets, and recent commits
- **Connection config** — `type: github`, config stores `repo_url`, `access_token`, `default_branch`
- **Rate limiting** — respect GitHub API rate limits (5000 req/hr for authenticated)

#### Requires

- Phase 7 (connection config UI for setting up GitHub repos)

---

### Phase 12 — Notifications & Escalation

**Why:** The agent watches 24/7, but findings need to reach humans where they already work.

#### Scope

- **Notification channels** — configurable per user: email, Slack webhook, Discord webhook
- **Severity thresholds** — configure which severity levels trigger notifications (e.g., only `critical` and `high`)
- **Notification preferences** — new `notification_settings` table and UI page
- **Escalation rules** — if an observation of severity `critical` goes unacknowledged for N minutes, re-notify or escalate to a secondary channel

#### Requires

- Phase 8 (monitoring mode generates observations worth notifying about)

---

## Future Considerations (Not Planned)

These are ideas from the vision doc and blueprint that don't have concrete plans yet:

| Idea | Notes |
|------|-------|
| **Multi-user / teams** | Shared workspaces, role-based access. Significant auth/permission rework |
| **DB connector streaming** | `pg_notify` or WAL-based change data capture instead of polling |
| **Syslog connector** | UDP/TCP syslog listener for traditional infrastructure |
| **Additional log connectors** | Fluentd, Vector, CloudWatch, Datadog agent forwarding |
| **Custom tool plugins** | User-defined tools the agent can call (e.g., restart service, check deploy status) |
| **Streaming agent responses** | Token-by-token streaming over WebSocket instead of full-response delivery |
| **Multi-model support** | Support for OpenAI, Gemini, or local models alongside Claude |
| **Self-hosted auth** | Alternative to Supabase for fully self-hosted deployments |

---

## Dependencies

```
Phase 7: Polish & Hardening
    ├──▶ Phase 8: Monitoring Mode
    │         ├──▶ Phase 9: Report Generation
    │         ├──▶ Phase 10: Elephantasm Memory
    │         └──▶ Phase 12: Notifications & Escalation
    └──▶ Phase 11: GitHub Codebase Connector
```

Phase 7 unblocks everything. Phases 8–12 have some interdependencies but are largely parallelizable after Phase 8.
