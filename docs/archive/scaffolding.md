# Heimdall — Project Scaffolding

The full directory structure to scaffold for the Heimdall monorepo. This covers every file and folder to create at project initialisation.

Reference: [Blueprint](../plans/blueprint.md) | [Vision](../vision.md)

---

## Root

```
heimdall/
├── frontend/
├── backend/
├── docs/
├── docker-compose.yml
├── Makefile
├── .gitignore
└── README.md
```

### `Makefile`

Root-level task runner for common operations across both projects.

```makefile
# Targets to include:
dev-frontend     # Run frontend dev server
dev-backend      # Run backend dev server
dev              # Run both concurrently
build-frontend   # Build frontend for production
build-backend    # Build backend binary
build            # Build both
test             # Run all tests
lint             # Lint both projects
sqlc-generate    # Run sqlc generate (regenerate Go code from SQL)
migrate-up       # Run DB migrations
migrate-down     # Rollback last migration
migrate-create   # Create a new migration file
docker-up        # docker-compose up
docker-down      # docker-compose down
```

### `docker-compose.yml`

Optional — for containerised local dev if needed.

```yaml
services:
  backend:      # Go backend (hot-reload via air)
  frontend:     # Vue dev server (Vite)
```

> Note: No local Postgres. Heimdall connects directly to a remote database (Supabase or Fly.io Postgres). Connection string configured via environment variable.

### `.gitignore`

```
# Go
backend/tmp/
backend/bin/

# Node
frontend/node_modules/
frontend/dist/

# Environment
.env
.env.local
.env.*.local

# IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Docker
docker-compose.override.yml
```

---

## Frontend — `frontend/`

Vue 3 + Vite + TypeScript

```
frontend/
├── public/
│   └── favicon.ico
│
├── src/
│   ├── api/                        # API client layer
│   │   ├── client.ts               #   Base HTTP client (axios/ofetch instance)
│   │   ├── connections.ts           #   Connections API calls
│   │   ├── agent.ts                 #   Agent config API calls
│   │   ├── logs.ts                  #   Agent log API calls
│   │   └── reports.ts               #   Reports API calls
│   │
│   ├── assets/                     # Static assets
│   │   └── styles/
│   │       └── main.css             #   Global styles / Tailwind entry
│   │
│   ├── components/                 # Reusable UI components
│   │   ├── common/                 #   Shared/generic components
│   │   │   ├── AppHeader.vue
│   │   │   ├── AppSidebar.vue
│   │   │   ├── StatusBadge.vue
│   │   │   └── LoadingSpinner.vue
│   │   │
│   │   ├── connections/            #   Connection management
│   │   │   ├── ConnectionCard.vue
│   │   │   ├── ConnectionForm.vue
│   │   │   └── ConnectionList.vue
│   │   │
│   │   ├── agent/                  #   Agent chat interface
│   │   │   ├── ChatWindow.vue
│   │   │   ├── ChatMessage.vue
│   │   │   └── ChatInput.vue
│   │   │
│   │   ├── log/                    #   Agent log / master feed
│   │   │   ├── LogFeed.vue
│   │   │   ├── LogEntry.vue
│   │   │   └── LogFilters.vue
│   │   │
│   │   └── reports/                #   Incident reports
│   │       ├── ReportCard.vue
│   │       ├── ReportDetail.vue
│   │       └── ReportList.vue
│   │
│   ├── composables/                # Vue composables (shared logic)
│   │   ├── useWebSocket.ts          #   WebSocket connection management
│   │   ├── useAgent.ts              #   Agent chat state + actions
│   │   └── useAuth.ts               #   Authentication state
│   │
│   ├── layouts/                    # Page layouts
│   │   └── DefaultLayout.vue
│   │
│   ├── pages/                      # Route-level views
│   │   ├── DashboardPage.vue
│   │   ├── ConnectionsPage.vue
│   │   ├── AgentConfigPage.vue
│   │   ├── AgentChatPage.vue
│   │   ├── AgentLogPage.vue
│   │   ├── ReportsPage.vue
│   │   └── LoginPage.vue
│   │
│   ├── router/
│   │   └── index.ts                 #   Vue Router config
│   │
│   ├── stores/                     # Pinia stores
│   │   ├── auth.ts
│   │   ├── connections.ts
│   │   ├── agent.ts
│   │   └── logs.ts
│   │
│   ├── types/                      # TypeScript type definitions
│   │   ├── connection.ts
│   │   ├── agent.ts
│   │   ├── log.ts
│   │   ├── report.ts
│   │   └── api.ts                   #   API request/response types
│   │
│   ├── utils/                      # Utility functions
│   │   ├── format.ts                #   Date/time formatting, etc.
│   │   └── constants.ts             #   App-wide constants
│   │
│   ├── App.vue                     # Root component
│   └── main.ts                     # App entry point
│
├── index.html
├── vite.config.ts
├── tsconfig.json
├── tsconfig.node.json
├── eslint.config.js
├── package.json
└── tailwind.config.js
```

### Key decisions

- **`api/` layer** — Dedicated directory for API calls rather than scattering fetch calls across components and stores. Each file maps to a backend resource.
- **`composables/`** — `useWebSocket` is critical; it handles connection lifecycle, reconnection, and message routing for both agent chat and live log streaming.
- **`pages/`** vs **`components/`** — Pages are route-level views (one per route). Components are reusable pieces composed within pages.
- **`types/`** — Centralised TypeScript types shared across API layer, stores, and components.

---

## Backend — `backend/`

Go (standard project layout)

```
backend/
├── cmd/
│   └── heimdall/
│       └── main.go                  # Application entry point
│
├── internal/                       # Private application code
│   ├── api/                        # HTTP layer
│   │   ├── router.go               #   Route definitions
│   │   ├── middleware/
│   │   │   ├── auth.go              #   Authentication middleware
│   │   │   ├── cors.go              #   CORS configuration
│   │   │   └── logging.go           #   Request logging
│   │   └── handlers/
│   │       ├── connections.go       #   Connection CRUD handlers
│   │       ├── agent.go             #   Agent config handlers
│   │       ├── chat.go              #   WebSocket chat handler
│   │       ├── logs.go              #   Agent log handlers
│   │       ├── reports.go           #   Report handlers
│   │       └── auth.go              #   Auth handlers
│   │
│   ├── agent/                      # Agent engine
│   │   ├── agent.go                 #   Agent struct + lifecycle
│   │   ├── loop.go                  #   Core tool-use loop
│   │   ├── monitor.go               #   24/7 monitoring goroutine
│   │   ├── tools.go                 #   Tool definitions + dispatch
│   │   ├── tools_db.go              #   Database query tool impl
│   │   ├── tools_logs.go            #   Log search tool impl
│   │   ├── tools_codebase.go        #   Codebase search tool impl
│   │   ├── tools_memory.go          #   Memory recall tool impl
│   │   └── prompt.go                #   System prompt construction
│   │
│   ├── connectors/                 # External integration layer
│   │   ├── connector.go             #   Connector interface definition
│   │   ├── registry.go              #   Connector registry (manages active connections)
│   │   ├── database/
│   │   │   ├── postgres.go          #   PostgreSQL connector
│   │   │   └── postgres_test.go
│   │   ├── logs/
│   │   │   ├── webhook.go           #   Webhook-based log ingestion
│   │   │   ├── syslog.go            #   Syslog connector
│   │   │   └── webhook_test.go
│   │   └── codebase/
│   │       ├── github.go            #   GitHub API connector
│   │       └── github_test.go
│   │
│   ├── memory/                     # Elephantasm integration
│   │   ├── client.go                #   HTTP client for Elephantasm API
│   │   ├── types.go                 #   Memory/Event/Lesson structs
│   │   └── memory.go                #   High-level memory operations
│   │
│   ├── db/                         # sqlc generated code (see docs/executing/db-models.md)
│   │   ├── db.go                    #   DBTX interface + Queries struct (generated)
│   │   ├── models.go                #   Go structs for all tables (generated)
│   │   ├── connections.sql.go       #   Connection queries (generated)
│   │   ├── agent_config.sql.go      #   Agent config queries (generated)
│   │   ├── investigations.sql.go    #   Investigation queries (generated)
│   │   ├── conversations.sql.go     #   Conversation queries (generated)
│   │   ├── log_buffer.sql.go        #   Log buffer queries (generated)
│   │   └── queries/                 #   Hand-written SQL (sqlc input)
│   │       ├── connections.sql
│   │       ├── agent_config.sql
│   │       ├── investigations.sql
│   │       ├── conversations.sql
│   │       └── log_buffer.sql
│   │
│   ├── ws/                         # WebSocket management
│   │   ├── hub.go                   #   Connection hub (manages all WS clients)
│   │   ├── client.go                #   Individual client connection
│   │   └── message.go               #   Message types + routing
│   │
│   ├── reports/                    # Report generation
│   │   ├── generator.go             #   Report builder
│   │   └── templates.go             #   Report structure/templates
│   │
│   └── config/                     # Application configuration
│       └── config.go                #   Config struct + env loading
│
├── sqlc.yaml                       # sqlc configuration
├── migrations/                     # SQL migration files (see docs/executing/db-models.md)
│   ├── 001_create_connections.up.sql
│   ├── 001_create_connections.down.sql
│   ├── 002_create_agent_config.up.sql
│   ├── 002_create_agent_config.down.sql
│   ├── 003_create_investigations.up.sql
│   ├── 003_create_investigations.down.sql
│   ├── 004_create_conversations.up.sql
│   ├── 004_create_conversations.down.sql
│   ├── 005_create_log_buffer.up.sql
│   └── 005_create_log_buffer.down.sql
│
├── go.mod
└── go.sum
```

### Key decisions

- **`internal/`** — All application code lives here. This is idiomatic Go; the `internal` directory prevents other Go modules from importing Heimdall's private packages.
- **`db/`** — sqlc-generated code. You write SQL queries in `db/queries/*.sql`, run `sqlc generate`, and get type-safe Go functions + model structs. No hand-written models or store layer needed — sqlc generates both.
- **`sqlc.yaml`** — Configuration for sqlc code generation. Lives at `backend/` root, points at `migrations/` for schema and `internal/db/queries/` for query files.
- **`connectors/registry.go`** — A registry manages all active connectors at runtime. The agent engine queries the registry to access connected data sources.
- **`agent/tools_*.go`** — Each tool gets its own file. Tools are the bridge between the agent's LLM reasoning and Heimdall's connectors/data.
- **`migrations/`** — Numbered up/down SQL files for golang-migrate. Each table gets its own migration pair. Also serves as the schema source for sqlc.

---

## Initialisation Commands

Once the scaffolding is approved, the project can be initialised with:

```bash
# Frontend
cd frontend && npm create vue@latest . -- --typescript
npm install pinia vue-router axios
npm install -D tailwindcss @tailwindcss/vite

# Backend
cd backend && go mod init github.com/<org>/heimdall/backend

# Dependencies (backend)
go get github.com/go-chi/chi/v5
go get github.com/coder/websocket
go get github.com/jackc/pgx/v5
go get github.com/anthropics/anthropic-sdk-go
go get github.com/golang-migrate/migrate/v4
go get github.com/google/uuid

# Install sqlc CLI (code generation tool — not a Go dependency)
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# After writing migrations + queries, generate Go code:
cd backend && sqlc generate
```

---

## What's Not Included (intentionally)

- **Kubernetes / Helm** — premature for MVP.
- **CI/CD pipeline** — add GitHub Actions once there's something to test and deploy.
- **Monitoring of Heimdall itself** — ironic, but not needed yet.
- **Multi-tenancy** — single-user/team for MVP.
- **Users / auth tables** — deferred. MVP assumes a single operator. Added later as `006_create_users` migration.
- **Reports table** — a resolved investigation with findings *is* the report for MVP. Separate table added later if reports need their own lifecycle.
