# Connections Blueprint

How Heimdall models and manages connections — the integrations between the platform and a user's infrastructure.

---

## Schema

The `connections` table is defined in migration `001` with additions in `007` (user scoping), `008` (webhook-token index), and `014` (app scoping via `app_id`). A broader connector ecosystem (pollers, syslog TLS, GitHub App) has been layered on top via newer migrations and handler code, but the `connections` row shape itself has been stable since 014.

| Column | Type | Default | Nullable | Description |
|--------|------|---------|----------|-------------|
| `id` | `UUID` | `gen_random_uuid()` | No | Primary key |
| `name` | `TEXT` | — | No | Human-readable label (e.g. "Production DB") |
| `type` | `TEXT` | — | No | Connection type — see [Type](#type) |
| `direction` | `TEXT` | `'one_way'` | No | Data flow direction — see [Direction](#direction) |
| `config` | `JSONB` | — | No | Type-specific configuration — see [Config](#config) |
| `status` | `TEXT` | `'inactive'` | No | Lifecycle state — see [Status](#status) |
| `last_seen` | `TIMESTAMPTZ` | `NULL` | Yes | Last time data flowed through this connection |
| `user_id` | `UUID` | — | No | FK → `users(id)` CASCADE. Owner for RLS enforcement |
| `app_id` | `UUID` | — | No | FK → `applications(id)` CASCADE. Parent application |
| `created_at` | `TIMESTAMPTZ` | `now()` | No | Row creation timestamp |
| `updated_at` | `TIMESTAMPTZ` | `now()` | No | Last modification timestamp |

### Indexes

| Index | Column(s) | Purpose |
|-------|-----------|---------|
| `idx_connections_status` | `status` | Filter by lifecycle state |
| `idx_connections_user_id` | `user_id` | User-scoped queries |
| `idx_connections_app_id` | `app_id` | App-scoped queries |

---

## Column Values

### Type

Determines what kind of infrastructure the connection links to. Stored as free-text — no DB-level constraint. The frontend enforces valid options via a `<select>` dropdown. Connectors fall into three broad shapes, matching the interface hierarchy in `backend/internal/connectors/connector.go`:

- **Push/stream ingestion** (logs arrive from the outside): `webhook_logs`, `syslog`, plus the public OTLP endpoint which writes directly into `log_buffer` without its own `connections` row.
- **Pull/poller ingestion** (Heimdall fetches on a schedule, managed by `connectors.Poller`): `flyio`, `vercel`, `railway`, `mongodb`, `supabase`.
- **Query connectors** (the agent queries on-demand during investigation): `postgres` (read-only SQL), `github` (GitHub App-backed code search).

| Value | Kind | Description |
|-------|------|-------------|
| `webhook_logs` | push | Log ingestion endpoint. A `webhook_token` is auto-generated on create and stored in `config`. External systems POST to `/api/webhooks/logs` with `Authorization: Bearer <webhook_token>`. The SDKs (`@heimdall/sdk`, `heimdall-sdk`, `sdk-go`) all use this endpoint. |
| `syslog` | push | Syslog over TCP/TLS. Backed by `connectors.ListenerManager` — listeners are spawned at boot for every `active` syslog connection and rehydrated in `resumePollers`. Supports TLS when `SYSLOG_TLS_CERT` / `SYSLOG_TLS_KEY` env vars are set. |
| `flyio` | poll | Fly.io logs poller. Polled via `connectors.Poller`. |
| `vercel` | poll | Vercel logs poller. |
| `railway` | poll | Railway logs poller. |
| `mongodb` | poll | MongoDB Atlas log poller. |
| `supabase` | poll | Supabase log poller. |
| `postgres` | query | Relational database. Supports test-on-create (live TCP ping via `database.New`). Agent runs read-only SQL against it via the `query_database` tool (forces `default_transaction_read_only=on`). |
| `github` | query | Codebase connection. Backed by the **Heimdall GitHub App** (not a PAT) — config stores `installation_id`; a secondary `github_repos` table (migration 020) tracks which repos inside the installation are enabled for this connection/app. Agent uses the `search_codebase` tool to search/read code via the GitHub App client. |

### Direction

How data flows between Heimdall and the connected system.

| Value | Label | Meaning |
|-------|-------|---------|
| `one_way` | One-way (ingest only) | Data flows into Heimdall. The agent cannot query the source. |
| `two_way` | Two-way (ingest + query) | Data flows in, and the agent can also run on-demand queries against the source (e.g. SQL on a Postgres connection). |

Default: `one_way`.

### Status

Lifecycle state of the connection. Updated by the backend — not user-editable on create (always starts `inactive`).

| Value | Meaning | Set by |
|-------|---------|--------|
| `inactive` | Created but not yet verified or not currently in use | Default on `CreateConnection` |
| `active` | Verified and operational | `TestConnection` on success; `UpdateConnectionStatus` when data flows |
| `error` | Last test or operation failed | `TestConnection` on failure |

### Config

A `JSONB` object whose shape depends on `type`. No schema validation at the DB level — the frontend renders type-specific forms, and the backend treats it as opaque JSON (except for `webhook_logs` token generation).

#### `postgres`

| Field | Type | Required | Default | Example |
|-------|------|----------|---------|---------|
| `host` | string | Yes | — | `db.example.com` |
| `port` | number | No | `5432` | `5432` |
| `database` | string | Yes | — | `mydb` |
| `user` | string | Yes | — | `postgres` |
| `password` | string | Yes | — | `s3cret` |
| `ssl_mode` | string | No | `require` | `disable`, `require`, `verify-full` |

#### `webhook_logs`

| Field | Type | Required | Default | Example |
|-------|------|----------|---------|---------|
| `webhook_token` | string | Auto | Auto-generated (32 bytes hex) | `a1b2c3d4...` |

No user-provided config fields. The token is generated server-side in `CreateConnection` and used to authenticate incoming POSTs at `/api/webhooks/logs` via the `Authorization: Bearer <token>` header. The token is looked up against `connections.config->>'webhook_token'` (indexed by migration 008).

#### `syslog`

| Field | Type | Required | Default | Example |
|-------|------|----------|---------|---------|
| `host` | string | Yes | — | `0.0.0.0` (bind address for incoming syslog) |
| `port` | number | No | `6514` | `6514` (TLS) / `514` (plain) |
| `protocol` | string | No | `tls` | `tcp`, `tls` |
| `tls_cert` | string | No (auto) | From `SYSLOG_TLS_CERT` env | PEM-encoded cert |
| `tls_key` | string | No (auto) | From `SYSLOG_TLS_KEY` env | PEM-encoded key |

The listener is spawned on create (and at boot via `resumeSyslogListeners`) and managed by `connectors.ListenerManager`. Server-level TLS material is injected from config if the connection row doesn't carry its own.

#### Poll-based connectors (`flyio`, `vercel`, `railway`, `mongodb`, `supabase`)

Each has its own provider-specific config shape (API tokens, project IDs, log group names, etc.) defined in `backend/internal/connectors/logs/<provider>.go`. Common pattern: a long-lived credential + the scope of what to poll. On create, `StartPoller` spawns a goroutine that polls the provider's API and inserts matching entries into `log_buffer` under this connection's `user_id` / `connection_id`.

#### `github`

| Field | Type | Required | Default | Notes |
|-------|------|----------|---------|-------|
| `installation_id` | number | Yes | — | GitHub App installation ID returned from the `/api/github/callback` OAuth flow |
| `account_login` | string | Yes | — | GitHub org or user that installed the app |

GitHub connections are provisioned via the Heimdall **GitHub App** (not a PAT). The flow is: user hits `GET /api/github/install` → redirects to GitHub → GitHub redirects back to `GET /api/github/callback` with an installation ID → a `github` connection row is created and repos are discovered via the GitHub App client. The per-repo enablement state lives in the separate `github_repos` table (migration 020), queried via `ListEnabledGitHubReposByApp` during agent tool dispatch.

---

## Ownership & Access Control

Connections sit at the bottom of the organizational hierarchy:

```
User → Organization → Application → Connection
```

- **`user_id`** — used for RLS policy enforcement. All read/write queries filter by `user_id`.
- **`app_id`** — structural ownership. On create, the backend verifies the target app belongs to the user's org via `GetApplicationByOrgUser` before allowing the connection.
- **CASCADE deletes** — deleting a user cascades to connections (via `user_id`); deleting an application also cascades (via `app_id`).

---

## API Endpoints

All endpoints require JWT authentication.

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| `GET` | `/api/connections` | `ListConnections` | List all connections for the authenticated user |
| `GET` | `/api/connections/{id}` | `GetConnection` | Get a single connection by ID |
| `POST` | `/api/connections` | `CreateConnection` | Create a new connection (requires `app_id`) |
| `PUT` | `/api/connections/{id}` | `UpdateConnection` | Update name, type, direction, config, status |
| `DELETE` | `/api/connections/{id}` | `DeleteConnection` | Delete a connection |
| `POST` | `/api/connections/{id}/test` | `TestConnection` | Test connectivity (Postgres: live ping; others: auto-pass) |
| `GET` | `/api/connections/{id}/github/repos` | `ListGitHubRepos` | List repos discovered for a GitHub App installation |
| `PUT` | `/api/connections/{id}/github/repos` | `UpdateGitHubRepos` | Update which repos are enabled for agent code search |
| `GET` | `/api/apps/{appId}/connections` | `ListConnectionsByApp` | App-scoped list (used by the app dashboard) |
| `POST` | `/api/webhooks/logs` | `IngestWebhookLogs` | **Public** — Bearer-auth log ingestion for `webhook_logs` connections |
| `POST` | `/api/v1/logs` | `IngestOTLPLogs` | **Public** — OTLP/HTTP ingestion (no connection row required) |
| `GET` | `/api/github/install` | `InstallGitHub` | Kick off the GitHub App install flow |
| `GET` | `/api/github/callback` | `GitHubCallback` | GitHub App install callback — creates the `github` connection |

---

## Key Files

| File | Role |
|------|------|
| `backend/migrations/001_create_connections.up.sql` | Initial table definition |
| `backend/migrations/007_add_user_id_to_connections.up.sql` | User scoping |
| `backend/migrations/008_add_webhook_token_index.up.sql` | Index for webhook-token lookup |
| `backend/migrations/014_organizations_applications.up.sql` | App scoping (`app_id` column) |
| `backend/migrations/020_github_repos.up.sql` | `github_repos` table — per-repo enablement for GitHub connections |
| `backend/internal/db/queries/connections.sql` | sqlc query definitions |
| `backend/internal/db/connections.sql.go` | Generated Go query methods |
| `backend/internal/api/handlers/connections.go` | HTTP handlers (CRUD + test) |
| `backend/internal/api/handlers/webhooks.go` | `POST /api/webhooks/logs` ingestion handler |
| `backend/internal/api/handlers/github.go` | GitHub App install + callback handlers |
| `backend/internal/connectors/connector.go` | `Connector` / `StreamConnector` / `QueryConnector` interface hierarchy |
| `backend/internal/connectors/poller.go` | Goroutine pool for poll-based connectors |
| `backend/internal/connectors/listener.go` | `ListenerManager` for syslog TCP/TLS listeners |
| `backend/internal/connectors/logs/*.go` | Per-provider implementations (flyio, vercel, railway, mongodb, supabase, syslog, webhook) |
| `backend/internal/connectors/database/postgres.go` | Read-only Postgres connector used by `query_database` |
| `backend/internal/connectors/codebase/github.go` | GitHub App code-search connector used by `search_codebase` |
| `frontend/src/types/connection.ts` | TypeScript types and payload interfaces |
| `frontend/src/components/connections/ConnectionForm.vue` | Create/edit form with type-specific config fields |
