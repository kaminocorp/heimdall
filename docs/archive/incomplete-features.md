# Heimdall — Incomplete Features

Post-Phase 9 audit, updated 2026-04-06. These are stubbed, partially implemented, or missing features that don't block core 24/7 monitoring but are needed to fulfil the full vision.

Reference: [Vision](../vision.md) · [Post-MVP Roadmap](./post-mvp-roadmap.md) · [Changelog](../changelog.md)

---

## Unbuilt Features

### 1. Report Generation

**Status:** Not implemented. The `internal/reports` package was deleted in 0.25.0 (it contained only a `Generate()` no-op). No report generation logic exists anywhere in the codebase.
**Frontend:** `ReportsPage.vue` calls `fetchReports()` — works but always shows an empty list.
**Vision ref:** "Produce actionable reports" — timeline, findings, historical context, severity assessment.

**What's needed:**
- Implement report generation — fetch investigation data, related agent log entries, conversation history
- Structure output: timeline, tools called, findings, severity
- Persist to `reports` table (schema may already exist via migration 003)
- Wire into monitor loop or allow manual trigger from chat ("Investigate the payment errors from last night")

---

### 2. Long-Term Memory (Elephantasm)

**Status:** Not implemented. The `internal/memory` package was deleted in 0.25.0 (all methods were TODO no-ops). `ELEPHANTASM_URL` and `ELEPHANTASM_API_KEY` are loaded in config but unused.
**Vision ref:** "Remember and learn" — past incidents, system patterns, known failure modes.

**What's needed:**
- Implement HTTP client wrapping the Elephantasm API
- Wire `recall_similar_incidents` and `recall_lessons` tools into the agent tool set
- Record events after monitoring assessments and investigations
- Enrich monitoring system prompt with recalled lessons

---

## Already Implemented (removed from open list)

These were listed in the original audit but have since shipped:

| Feature | Shipped in | Notes |
|---------|-----------|-------|
| Postgres Connector Interface Mismatch | 0.26.1 | `Close` signature fixed; `Health` method added; two call sites updated |
| Health Check Endpoint | 0.26.0 | `GET /health` with DB ping; Docker Compose healthcheck wired |
| Prometheus Metrics | 0.26.0 | `GET /metrics` — monitor tick duration, logs classified, notifications sent/failed |
| JSON Structured Logging | 0.26.0 | `LOG_FORMAT=json` switches slog to `NewJSONHandler` |
| Notification Env Vars in `.env.example` | 0.26.0 | `RESEND_API_KEY`, `NOTIFICATION_FROM_EMAIL` documented |
| Syslog Connector | 0.22.0 | Full TCP/TLS listener, RFC 5424 + 3164 parsing, semaphore-capped connections, idle timeouts |
| GitHub Codebase Connector | 0.16.0 | GitHub App auth, `search_code`, `read_file`, `list_tree` via Contents + Search APIs |

**Note on OpenTelemetry tracing:** The original observability item listed OTel tracing for agent tool-use loops as optional. This has not been implemented — it remains a future option if deeper request-level tracing is needed.

---

## Priority Suggestion

| # | Feature | Impact | Effort |
|---|---------|--------|--------|
| 1 | Report generation | Completes vision | Medium |
| 2 | Elephantasm memory | Agent learning | Medium–Large |
