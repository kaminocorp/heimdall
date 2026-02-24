# Heimdall — MVP Implementation Roadmap

Reference: [Blueprint](./blueprint.md) · [Vision](../vision.md)

---

## Goal

Turn the scaffolded stubs into a working end-to-end system. After this phase, a user can log in, configure a connection, send logs to Heimdall, chat with the agent, and browse the agent log.

---

## Phase 1 — Basic Auth

**Why first:** Every other feature depends on a real user session.

### Scope

- User registration endpoint (`POST /api/auth/register`)
- Login endpoint (`POST /api/auth/login`) — already routed, currently a stub
- JWT token issuance and validation
- Auth middleware enforces valid tokens on protected routes
- Frontend login page wired to real API; token stored in auth store
- Password hashing (bcrypt)

### Requires

- `users` table migration (email, hashed password, created_at)
- sqlc queries for user lookup/creation

---

## Phase 2 — Connections CRUD

**Why next:** Connections are the foundation for all data ingestion.

### Scope

- `GET /api/connections` — list all connections for the user
- `POST /api/connections` — create a new connection
- `GET /api/connections/:id` — get connection details
- `PUT /api/connections/:id` — update connection config
- `DELETE /api/connections/:id` — remove connection
- Handlers call real sqlc-generated queries (already generated)
- Frontend Connections page wired to real API

### Requires

- Handlers converted from stubs to real DB calls via sqlc
- Connection config stored as JSONB (connector-specific settings)

---

## Phase 3 — Webhook Log Connector

**Why next:** Simplest way to get real data flowing into Heimdall.

### Scope

- `POST /api/webhooks/logs` — public endpoint that accepts log payloads
- Incoming logs validated, associated with a connection, persisted to `log_buffer`
- Log entries available via `GET /api/logs` (paginated, filterable)
- Frontend log feed displays real persisted logs

### Requires

- Webhook connector implementation (replace stub)
- Log ingestion pipeline: receive → validate → persist
- Log query endpoint with filters (severity, connection, time range)

---

## Phase 4 — Agent Loop

**Why next:** The core intelligence — makes Heimdall more than a log viewer.

### Scope

- Wire up Claude API via `anthropic-sdk-go`
- Implement real tool-use loop: send message → receive tool calls → execute → return results → repeat
- Implement two real tools:
  - `search_logs` — queries `log_buffer` for matching entries
  - `query_database` — runs read-only SQL against a user's connected database
- Agent config read from DB (model, system prompt, schedule)
- Tool results fed back into the conversation loop

### Requires

- Anthropic API key in config/env
- Agent engine `Run()` method replaces stub with real Claude API calls
- Tool implementations query real data (log_buffer, user DB via connector)

---

## Phase 5 — Agent Chat (End-to-End)

**Why next:** Connects the user to the agent through the existing WebSocket infrastructure.

### Scope

- WebSocket hub routes incoming user messages to the agent loop
- Agent responses streamed back to the client via WebSocket
- Conversation persistence — messages saved to `conversations` table
- Frontend chat UI displays real agent responses
- Conversation history loaded on reconnect

### Requires

- WebSocket hub → agent engine integration
- Conversation CRUD (create, append message, list history)
- Frontend `useAgent` composable connected to real WebSocket flow

---

## Phase 6 — Agent Log

**Why next:** Gives visibility into what the agent is doing and what's happening in the system.

### Scope

- Unified chronological feed combining:
  - Raw ingested log entries (from webhook connector)
  - Agent observations (annotations, flags, investigation notes)
- `GET /api/logs` extended with source filtering (raw vs agent)
- Frontend log feed shows both raw logs and agent activity
- Agent engine emits log entries when it observes or investigates something

### Requires

- Agent log entry type/model (distinct from raw log entries)
- Agent engine hooks to emit observations during monitoring/investigation
- Log feed API supports mixed entry types

---

## Out of MVP Scope

Deferred to post-MVP (per blueprint):

- Codebase connector (GitHub integration)
- Elephantasm long-term memory
- Report generation
- Multi-user / teams
- Scheduled monitoring mode (always-on goroutine)
- DB connector streaming (pg_notify)

---

## Dependencies

```
Phase 1: Auth
    └──▶ Phase 2: Connections CRUD
              └──▶ Phase 3: Webhook Log Connector
                        └──▶ Phase 4: Agent Loop
                                  └──▶ Phase 5: Agent Chat
                                            └──▶ Phase 6: Agent Log
```

Each phase builds on the previous. Phases are sequential — later phases depend on infrastructure from earlier ones.
