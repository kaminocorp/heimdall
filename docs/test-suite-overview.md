# Test Suite Overview

Last updated: 2026-04-05

This document catalogues every test file in the Heimdall codebase, what it covers, and where the gaps are. Use it as a reference when deciding what needs testing before a release or when adding new features.

## Running Tests

```bash
make test                              # All tests (backend + frontend)
cd backend && go test ./...            # All backend tests
cd backend && go test ./internal/agent # Single backend package
cd frontend && npm run test            # All frontend tests (vitest)
cd frontend && npm run test:watch      # Frontend watch mode
```

Backend integration tests (API handlers) require `DATABASE_URL` and skip automatically if missing.

---

## Backend (Go) — 20 test files

### API Handler Tests (7 files)

These are **integration tests** that hit a real database via the test harness in `testhelpers_test.go`.

| File | What it tests |
|------|---------------|
| `internal/api/handlers/auth_test.go` | `/api/auth/me` endpoint — user authentication retrieval |
| `internal/api/handlers/connections_test.go` | `ListConnections`, `CreateConnection` — webhook_logs connection creation |
| `internal/api/handlers/webhooks_test.go` | `IngestWebhookLogs` — webhook token auth, log ingestion |
| `internal/api/handlers/logs_test.go` | `ListLogs_Empty`, `ListLogs_WithEntries` — log retrieval with pagination |
| `internal/api/handlers/organizations_test.go` | User-org interactions, bare user creation, per-user environment setup |
| `internal/api/handlers/applications_test.go` | `ListApplications`, `CreateApplication` — app status and ownership |
| `internal/api/handlers/otlp_test.go` | `MapOTLPSeverity_ByNumber`, `MapOTLPSeverity_ByText` — OpenTelemetry severity mapping |

### Agent Tests (5 files)

Unit tests covering the AI agent pipeline — classification, escalation, tool dispatch, and monitoring.

| File | What it tests |
|------|---------------|
| `internal/agent/loop_test.go` | Agent loop execution with mock Anthropic API server |
| `internal/agent/extract_test.go` | `ExtractText` — log text extraction from various payload formats (message, msg, error, supabase error_severity) |
| `internal/agent/severity_gate_test.go` | `ShouldEscalate` — event classification by type/category (ERROR, PERFORMANCE, REQUEST, DEPLOY, SYSTEM, ACCESS, DATA) |
| `internal/agent/classifier_test.go` | `PassthroughClassifier` — log classification, flagging, empty batch handling, cleanup |
| `internal/agent/monitor_test.go` | `FormatFlaggedLogs` — formatting of classified logs for Claude escalation |
| `internal/agent/tools_test.go` | `Dispatch` — tool routing, unknown tool error handling, `search_logs` dispatch |

### Connector Tests (5 files)

Unit tests for connector initialisation, config validation, and polling logic.

| File | What it tests |
|------|---------------|
| `internal/connectors/logs/webhook_test.go` | `NewWebhook` — webhook connector initialisation |
| `internal/connectors/logs/supabase_test.go` | `NewSupabase` — valid config, default values (project_ref, access_token, poll_tables, poll_interval) |
| `internal/connectors/logs/syslog_test.go` | `NewSyslog` — defaults, invalid port/protocol validation, RFC5424 message parsing |
| `internal/connectors/database/postgres_test.go` | `New` — Postgres connector init, `NewMissingFields` — config validation |
| `internal/connectors/poller_test.go` | `Poller_StartStop`, `Poller_StopAll` — polling lifecycle with mock connector |
| `internal/connectors/codebase/github_test.go` | `New` — GitHub connector init, nil client error handling |

### Test Infrastructure (1 file)

| File | What it provides |
|------|-----------------|
| `internal/api/handlers/testhelpers_test.go` | `testEnv` struct (Pool, Queries, Server, Router, UserID, OrgID, AppID), `testSetup()` for full DB environment, `request()` HTTP helper, `createTestConnection()` factory, Chi router with context injection, cascade cleanup |

---

## Frontend (Vue 3 + TypeScript) — 5 test files

### Store Tests (4 files)

| File | What it tests |
|------|---------------|
| `stores/__tests__/connections.test.ts` | `useConnectionsStore` — fetchConnections, createConnection, error handling; mocks client.get/post/put/delete |
| `stores/__tests__/logs.test.ts` | `useLogsStore` — fetchLogs, pagination (nextPage), error handling |
| `stores/__tests__/agent.test.ts` | `useAgentStore` — fetchConfig, updateConfig (PUT + local state update) |
| `stores/__tests__/auth.test.ts` | `useAuthStore` — init, login (signInWithPassword), session management; mocks Supabase auth |

### Component Tests (1 file)

| File | What it tests |
|------|---------------|
| `components/connections/__tests__/ConnectionForm.test.ts` | Supabase type in dropdown, Supabase config field rendering (Project Reference, PAT, Poll Interval) |

### Test Infrastructure

| File | What it provides |
|------|-----------------|
| `src/test/setup.ts` | Fresh Pinia instance per test (`beforeEach`), global Axios client mock (get/post/put/delete + interceptors) |
| `vite.config.ts` (test section) | Vitest config — happy-dom environment, globals enabled, setup file path |

---

## Coverage Gaps

### Backend — well covered, some gaps

- **Covered well**: API handlers (core CRUD), agent classification pipeline, connector init/validation, polling lifecycle
- **Gaps**:
  - No tests for `agent/config` handlers (GET/PUT agent configuration)
  - No tests for `conversations` or `reports` endpoints
  - No tests for the monitoring loop end-to-end (only `FormatFlaggedLogs` helper tested)
  - `LumberClassifier` (ONNX) not directly tested (only `PassthroughClassifier`)
  - WebSocket/chat handler untested

### Frontend — sparse, stores only

- **Covered**: Pinia store data flow for 4 of the core stores
- **Gaps**:
  - No page-level tests (Dashboard, AgentChat, AgentLog, Reports, etc.)
  - No wizard flow tests (multi-step connection creation)
  - No WebSocket/chat UI tests (`useWebSocket`, `useAgent` composables)
  - No routing tests (guards, navigation, stale-asset reload handler)
  - Only 1 component test out of ~30+ components
  - No end-to-end tests

---

## Dependencies & Frameworks

### Backend

| Dependency | Purpose |
|------------|---------|
| Go `testing` | Test framework |
| `github.com/stretchr/testify` | Assertions (assert, require) |
| `net/http/httptest` | HTTP test server |
| `github.com/jackc/pgx/v5/pgxpool` | Database connection (integration tests) |
| `github.com/go-chi/chi/v5` | Router in test harness |

### Frontend

| Dependency | Purpose |
|------------|---------|
| Vitest 3.2.4 | Test framework |
| happy-dom 20.8.3 | DOM environment |
| Vue Test Utils 2.4.6 | Component mounting |
| Pinia 3.0.0 | Store testing |
| Vitest `vi` | Mocking (vi.mock, vi.fn, vi.mocked) |
