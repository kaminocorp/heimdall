# Heimdall — Incomplete Features

Post-Phase 9 audit. These are stubbed, partially implemented, or missing features that don't block core 24/7 monitoring but are needed to fulfil the full vision.

Reference: [Vision](../vision.md) · [Post-MVP Roadmap](./post-mvp-roadmap.md) · [Changelog](../changelog.md)

---

## Stubbed Features

### 1. Report Generation

**Files:** `backend/internal/reports/generator.go`
**Status:** `Generate()` returns nil — no logic.
**Frontend:** `ReportsPage.vue` calls `fetchReports()` — works but always shows an empty list.
**Vision ref:** "Produce actionable reports" — timeline, findings, historical context, severity assessment.

**What's needed:**
- Implement `Generate(investigationID)` — fetch investigation data, related agent log entries, conversation history
- Structure output: timeline, tools called, findings, severity
- Persist to `reports` table (schema may already exist via migration 003)
- Wire into monitor loop or allow manual trigger from chat ("Investigate the payment errors from last night")

---

### 2. Long-Term Memory (Elephantasm)

**Files:** `backend/internal/memory/client.go`, `backend/internal/agent/tools_memory.go`
**Status:** All methods are TODO stubs — `RecordEvent`, `QueryMemories`, `GetLessons` return immediately.
**Env vars:** `ELEPHANTASM_URL`, `ELEPHANTASM_API_KEY` — loaded in config but unused.
**Vision ref:** "Remember and learn" — past incidents, system patterns, known failure modes.

**What's needed:**
- Implement HTTP client wrapping the Elephantasm API
- Wire `recall_similar_incidents` and `recall_lessons` tools into the agent tool set
- Record events after monitoring assessments and investigations
- Enrich monitoring system prompt with recalled lessons

---

### 3. GitHub Codebase Connector

**Files:** `backend/internal/connectors/codebase/github.go`, `backend/internal/agent/tools_codebase.go`
**Status:** Placeholder — no GitHub API calls, no `search_codebase` tool logic.
**Vision ref:** "Agent queries a GitHub repo to understand relevant code during investigation."

**What's needed:**
- GitHub API client (Search API + Contents API) with PAT auth
- `search_codebase` agent tool — search files, functions, patterns in connected repos
- Connection config: `type: github`, storing `repo_url`, `access_token`, `default_branch`
- Rate limiting (5000 req/hr authenticated)

---

### 4. Syslog Connector

**Files:** `backend/internal/connectors/logs/syslog.go`
**Status:** TODO stub — no UDP/TCP listener.
**Impact:** Low priority. Webhook ingestion covers most use cases.

---

## Missing Infrastructure

### 5. Health Check Endpoint

**Status:** No `/health` or `/healthz` endpoint on the backend.
**Impact:** Docker Compose backend service can't report healthy. Blocks proper K8s readiness/liveness probes.

**What's needed:**
- `GET /health` — returns 200 if server is up, DB is reachable
- Add `healthcheck` to backend service in `docker-compose.yml`

---

### 6. Observability

**Status:** Logging via `slog` to stdout only. No metrics, no tracing.
**Impact:** Limited visibility into monitor loop performance, notification delivery rates, classifier accuracy.

**What's needed (pick as needed):**
- Prometheus metrics endpoint (`/metrics`) — monitor loop duration, logs classified/sec, notifications sent/failed
- Structured log export (JSON format for log aggregators)
- Optional: OpenTelemetry tracing for agent tool-use loops

---

## Minor Code Issues

### 7. Postgres Connector Interface Mismatch

**Files:** `backend/internal/connectors/database/postgres.go`
**Issue:** `Close(ctx context.Context) error` doesn't match the `Connector` interface's `Close() error`. `Health()` method is also missing.
**Impact:** Not called polymorphically today, but will cause a compile error if it ever is. Latent bug.

### 8. Notification Env Vars Not in .env.example

**Issue:** `RESEND_API_KEY` and `NOTIFICATION_FROM_EMAIL` aren't documented in `.env.example`.
**Impact:** Email notifications silently fail without clear guidance to operators.

---

## Priority Suggestion

| # | Feature | Impact | Effort |
|---|---------|--------|--------|
| 5 | Health check endpoint | Deployment hygiene | Small |
| 8 | Env var documentation | Operator experience | Trivial |
| 7 | Connector interface fix | Code correctness | Small |
| 1 | Report generation | Completes vision | Medium |
| 2 | Elephantasm memory | Agent learning | Medium–Large |
| 6 | Observability | Production confidence | Medium |
| 3 | GitHub connector | Investigation depth | Medium |
| 4 | Syslog connector | Niche use cases | Low priority |
