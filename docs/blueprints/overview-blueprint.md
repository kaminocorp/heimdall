# Heimdall — Project Blueprint

## Overview

Heimdall is an autonomous AI monitoring agent platform for production applications. It replaces basic log alerts with an intelligent agent that has 24/7 access to infrastructure — capable of monitoring, investigating, and diagnosing issues in real time.

Reference: [Vision Document](../vision.md)

---

## Tech Stack

### Frontend — Vue 3 + Vite + TypeScript

| Layer | Choice | Rationale |
|-------|--------|-----------|
| Framework | Vue 3 (Composition API) | Lightweight, fast, good DX |
| Build tool | Vite | Fast HMR, native TS support |
| Language | TypeScript | Type safety across the frontend |
| Styling | TBD (Tailwind CSS recommended) | Utility-first, fast iteration |
| State management | Pinia | Official Vue store, simple API |
| WebSocket client | Native WebSocket / reconnecting-websocket | Real-time agent chat + live log streaming |
| HTTP client | Axios or ofetch | API calls to backend |

### Backend — Go

| Layer | Choice | Rationale |
|-------|--------|-----------|
| Language | Go | Excellent concurrency (goroutines), built for long-running services, strong performance for streaming workloads |
| HTTP framework | Chi or Echo | Lightweight, idiomatic Go routing |
| WebSocket | gorilla/websocket or nhooyr/websocket | Proven WebSocket libraries |
| LLM integration | anthropic-sdk-go | Official Anthropic Go SDK for Claude tool-use agent loop |
| Database | PostgreSQL (remote — Supabase or Fly.io) | Reliable, feature-rich, good for structured data + JSONB for flexible schemas |
| DB driver | pgx | High-performance native Go PostgreSQL driver |
| DB query layer | sqlc | SQL-first code generation — write SQL, get type-safe Go functions |
| Migrations | golang-migrate | Simple, file-based schema migrations (also serves as sqlc's schema source) |
| Config | envconfig or viper | Environment-based configuration |
| Logging | slog (stdlib) | Structured logging, zero dependencies |
| Agent memory | Elephantasm | Long-term agentic memory (see [Elephantasm Integration](./elephantasm-integration.md)) |

### Infrastructure / Tooling

| Concern | Choice | Rationale |
|---------|--------|-----------|
| Containerisation | Docker + Docker Compose | Local dev and deployment |
| CI/CD | GitHub Actions | Standard, free for public repos |
| Monorepo tooling | Makefile at root | Simple task orchestration across frontend/backend |

---

## Architecture

### High-Level Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend (Vue)                        │
│  Connections UI  │  Agent Chat  │  Agent Log  │  Reports     │
└────────┬──────────────┬──────────────┬──────────────┬───────┘
         │ HTTP         │ WebSocket    │ HTTP         │ HTTP
         ▼              ▼              ▼              ▼
┌─────────────────────────────────────────────────────────────┐
│                       Go Backend (API)                       │
│                                                              │
│  ┌──────────┐  ┌──────────────┐  ┌───────────┐  ┌────────┐ │
│  │ REST API │  │ WebSocket    │  │ Agent     │  │ Report │ │
│  │ Handlers │  │ Hub          │  │ Engine    │  │ Engine │ │
│  └──────────┘  └──────────────┘  └─────┬─────┘  └────────┘ │
│                                        │                     │
│  ┌─────────────────────────────────────┼───────────────────┐ │
│  │            Connectors Layer         │                   │ │
│  │  ┌─────────┐ ┌──────────┐ ┌────────┴───┐ ┌──────────┐ │ │
│  │  │ DB      │ │ Log      │ │ Codebase   │ │ Memory   │ │ │
│  │  │Connector│ │Connector │ │ Connector  │ │(Elephant)│ │ │
│  │  └────┬────┘ └────┬─────┘ └─────┬──────┘ └──────────┘ │ │
│  └───────┼───────────┼─────────────┼────────────────────── │ │
│          │           │             │                        │
└──────────┼───────────┼─────────────┼────────────────────────┘
           ▼           ▼             ▼
     ┌──────────┐ ┌─────────┐ ┌──────────┐
     │ User's   │ │ User's  │ │ GitHub   │
     │ Database │ │ Log     │ │ API      │
     │          │ │ Sources │ │          │
     └──────────┘ └─────────┘ └──────────┘
```

### Core Components

**REST API Handlers**
Standard CRUD for connections, agent config, reports. Serves the frontend.

**WebSocket Hub**
Manages persistent connections for:
- Agent chat (user ↔ agent real-time conversation)
- Live log streaming (log feed → frontend)

**Agent Engine**
The core intelligence layer. Responsible for:
- The agent tool-use loop (Claude API ↔ tool execution cycle)
- 24/7 monitoring mode: a background goroutine that processes incoming log streams, detects anomalies, and triggers the agent when intervention is needed
- On-demand mode: responds to user queries via the chat interface
- Scheduling: supports always-on monitoring or periodic check-ins (configurable)

**Connectors Layer**
Abstraction over external integrations. Each connector implements a common interface:
- `Connect()` — establish connection
- `Stream()` — for one-way feeds (logs)
- `Query()` — for on-demand queries (DB, codebase)
- `Health()` — connection health check

**Report Engine**
Generates incident reports from agent analysis. Stores them in the Heimdall database.

### Agent Loop (Detail)

```
          ┌────────────────────────────┐
          │   Monitoring Goroutine     │
          │   (ingests logs 24/7)      │
          └─────────┬──────────────────┘
                    │ anomaly / periodic summary
                    ▼
          ┌────────────────────────────┐
          │      Agent Loop            │
          │                            │
          │  1. Build context          │
          │     (logs, system state)   │
          │  2. Send to Claude         │◄──── User chat message
          │     (with tools)           │
          │  3. If tool_call:          │
          │     - execute tool         │
          │     - append result        │
          │     - goto 2               │
          │  4. If text response:      │
          │     - emit to chat/log     │
          │     - store in memory      │
          └────────────────────────────┘
```

### Data Flow

| Flow | Direction | Mechanism |
|------|-----------|-----------|
| Server logs → Heimdall | One-way, streaming | Log connector (syslog/webhook/file tail) |
| DB activity → Heimdall | One-way, streaming | DB connector (pg_notify / polling) |
| Agent → User's DB | On-demand | SQL query via DB connector |
| Agent → Codebase | On-demand | GitHub API via codebase connector |
| User ↔ Agent | Bidirectional | WebSocket |
| Agent → Agent Log | One-way | Internal event bus → persistence |
| Agent → Reports | One-way | Report engine triggered by agent |

---

## Monorepo Structure

```
heimdall/
├── frontend/                   # Vue 3 + Vite + TypeScript
│   ├── src/
│   │   ├── assets/
│   │   ├── components/
│   │   │   ├── connections/    # Connection management UI
│   │   │   ├── agent/          # Agent chat interface
│   │   │   ├── log/            # Agent log / master feed
│   │   │   └── reports/        # Incident reports
│   │   ├── composables/        # Vue composables (shared logic)
│   │   ├── layouts/
│   │   ├── pages/              # Route-level views
│   │   ├── router/
│   │   ├── stores/             # Pinia stores
│   │   ├── types/              # TypeScript types
│   │   ├── utils/
│   │   ├── App.vue
│   │   └── main.ts
│   ├── public/
│   ├── index.html
│   ├── vite.config.ts
│   ├── tsconfig.json
│   └── package.json
│
├── backend/                    # Go backend
│   ├── cmd/
│   │   └── heimdall/           # Main entrypoint
│   │       └── main.go
│   ├── internal/
│   │   ├── api/                # HTTP handlers + routes
│   │   │   ├── handlers/
│   │   │   ├── middleware/
│   │   │   └── router.go
│   │   ├── agent/              # Agent engine
│   │   │   ├── loop.go         # Core agent tool-use loop
│   │   │   ├── monitor.go      # 24/7 monitoring goroutine
│   │   │   ├── tools.go        # Tool definitions + dispatch
│   │   │   └── prompt.go       # System prompts
│   │   ├── connectors/         # External integrations
│   │   │   ├── connector.go    # Connector interface
│   │   │   ├── database/       # DB connector
│   │   │   ├── logs/           # Log stream connector
│   │   │   └── codebase/       # GitHub / codebase connector
│   │   ├── memory/             # Elephantasm integration
│   │   ├── models/             # Domain models / DB models
│   │   ├── reports/            # Report generation
│   │   ├── ws/                 # WebSocket hub
│   │   └── config/             # App configuration
│   ├── migrations/             # SQL migration files
│   ├── go.mod
│   └── go.sum
│
├── docs/
│   ├── vision.md
│   └── plans/
│       ├── blueprint.md        # This file
│       └── elephantasm-integration.md
│
├── docker-compose.yml          # Local dev (backend + frontend + postgres)
├── Makefile                    # Root-level task runner
└── README.md
```

---

## Key Architectural Decisions

### 1. Go for the backend
Goroutines make concurrent log ingestion and long-lived WebSocket connections trivial. The agent loop is a simple construct in Go — no framework needed. The tradeoff is a thinner AI ecosystem, but we're mostly making HTTP calls to Claude's API via the official SDK.

### 2. Connector abstraction
All external integrations implement a common interface. This keeps the agent engine decoupled from specific data sources and makes adding new connector types straightforward.

### 3. Agent engine is not a framework
The agent loop is custom code, not a LangChain-style framework. Heimdall's agent has a focused purpose (monitoring + diagnostics) with a well-defined tool set. A custom loop gives full control, better observability, and less abstraction overhead.

### 4. Elephantasm for agent memory
Long-term memory is critical for a monitoring agent — remembering past incidents, learning system patterns, and building institutional knowledge. Integrated via direct API calls from Go. See [Elephantasm Integration](./elephantasm-integration.md).

### 5. PostgreSQL as Heimdall's own database
Stores connections config, agent config, agent logs, reports, and user data. JSONB columns give flexibility for connector-specific configuration without schema sprawl.

### 6. WebSocket for real-time
Agent chat and live log streaming both use WebSockets. The Go backend manages a hub that multiplexes connections efficiently.

---

## MVP Scope (Suggested)

For a first working version, focus on:

1. **Single DB connector** (PostgreSQL) — one-way feed via pg_notify + on-demand SQL queries
2. **Single log connector** — webhook-based log ingestion (apps POST logs to Heimdall)
3. **Agent loop** — Claude-powered tool-use loop with query_database, search_logs tools
4. **Agent chat** — WebSocket-based chat UI in Vue
5. **Agent log** — chronological feed of agent observations + raw log entries
6. **Basic auth** — simple user authentication

Defer to post-MVP: codebase connector, reports, Elephantasm memory, multi-user/teams, scheduled check-in mode.
