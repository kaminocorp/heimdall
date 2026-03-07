# Connections Blueprint

How Heimdall models and manages connections — the integrations between the platform and a user's infrastructure.

---

## Schema

The `connections` table is defined in migration `001` with additions in `007` (user scoping) and `014` (app scoping).

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

Determines what kind of infrastructure the connection links to. Stored as free-text — no DB-level constraint. The frontend enforces valid options via a `<select>` dropdown.

| Value | Label | Description |
|-------|-------|-------------|
| `postgres` | PostgreSQL | Relational database. Supports test-on-create (live TCP ping via `database.New`). Agent can run read-only SQL queries against it. |
| `webhook_logs` | Webhook Logs | Log ingestion endpoint. A `webhook_token` is auto-generated on create and stored in `config`. External systems POST logs to `/api/webhooks/logs/{token}`. |
| `syslog` | Syslog | Syslog source (not yet implemented server-side). Config captures host, port, and protocol. |
| `github` | GitHub | Codebase connection. Config captures owner, repo, and personal access token. Intended for agent code inspection during investigation. |

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

No user-provided config fields. The token is generated server-side in `CreateConnection` and used to authenticate incoming log POSTs at `/api/webhooks/logs/{token}`.

#### `syslog`

| Field | Type | Required | Default | Example |
|-------|------|----------|---------|---------|
| `host` | string | Yes | — | `syslog.example.com` |
| `port` | number | No | `514` | `514` |
| `protocol` | string | No | `udp` | `udp`, `tcp` |

#### `github`

| Field | Type | Required | Default | Example |
|-------|------|----------|---------|---------|
| `owner` | string | Yes | — | `my-org` |
| `repo` | string | Yes | — | `my-app` |
| `token` | string | Yes | — | `ghp_xxxxxxxxxxxx` |

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

---

## Key Files

| File | Role |
|------|------|
| `backend/migrations/001_create_connections.up.sql` | Initial table definition |
| `backend/migrations/007_add_user_id_to_connections.up.sql` | User scoping |
| `backend/migrations/014_organizations_applications.up.sql` | App scoping (`app_id` column) |
| `backend/internal/db/queries/connections.sql` | sqlc query definitions |
| `backend/internal/db/connections.sql.go` | Generated Go query methods |
| `backend/internal/api/handlers/connections.go` | HTTP handlers (CRUD + test) |
| `frontend/src/types/connection.ts` | TypeScript types and payload interfaces |
| `frontend/src/components/connections/ConnectionForm.vue` | Create/edit form with type-specific config fields |
