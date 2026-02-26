# Phase 2 — Connections CRUD

Replaced stub connection handlers with real database-backed CRUD, scoped all queries to the authenticated user, and wired the frontend to the live API.

---

## Migration

### `007_add_user_id_to_connections.up.sql`

- Added `user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE` to `connections` table.
- Added index `idx_connections_user_id` on `connections.user_id` for filtered lookups.

**Why:** The original `connections` table had no ownership column. Without `user_id`, any authenticated user could see or modify any connection. Even though multi-user is out of MVP scope, this prevents data leakage and avoids a painful migration later.

---

## sqlc Queries

### `internal/db/queries/connections.sql`

Rewrote all queries to scope by `user_id`:

| Old query | New query | Change |
|-----------|-----------|--------|
| `ListConnections` | `ListConnectionsByUser` | Filters by `user_id` |
| `GetConnection` | `GetConnectionByUser` | Requires `id` AND `user_id` |
| `CreateConnection` | `CreateConnection` | Now accepts `user_id` as first param |
| `UpdateConnection` | `UpdateConnection` | WHERE clause includes `user_id` |
| `DeleteConnection` | `DeleteConnectionByUser` | Requires `id` AND `user_id` |
| `UpdateConnectionStatus` | `UpdateConnectionStatus` | Unchanged (internal/system use) |

Ran `sqlc generate` — updated `connections.sql.go` and `models.go` (added `UserID` field to `Connection` struct).

---

## Backend Handlers

### `internal/api/handlers/connections.go`

Replaced all 5 stubs with real implementations:

| Handler | Method | Behaviour |
|---------|--------|-----------|
| `ListConnections` | `GET /api/connections` | Extracts user UUID from JWT context, calls `ListConnectionsByUser`, returns JSON array. |
| `GetConnection` | `GET /api/connections/{id}` | Parses `{id}` path param, calls `GetConnectionByUser` with user scope, 404 on miss. |
| `CreateConnection` | `POST /api/connections` | Decodes JSON body (`name`, `type`, `direction`, `config`), defaults `direction` to `one_way`, `config` to `{}`, `status` to `inactive`. Returns 201. |
| `UpdateConnection` | `PUT /api/connections/{id}` | Full replacement update. Validates required fields, scoped to user. 404 if not found or not owned. |
| `DeleteConnection` | `DELETE /api/connections/{id}` | Scoped delete, returns 204 No Content. |

All handlers extract `userID` via `middleware.UserIDFromContext()` and return `401` if absent.

### Request validation

- `name` and `type` are required on create and update (400 if missing).
- `direction` defaults to `"one_way"`, `config` defaults to `{}`, `status` defaults to `"inactive"`.

### No router changes

Handler method names on `Server` are unchanged — the router (`api/router.go`) required zero modifications.

---

## Frontend

### Types — `types/connection.ts`

- Added `user_id` field to `Connection` interface (matches backend response).
- Added `CreateConnectionPayload` and `UpdateConnectionPayload` interfaces for typed API calls.

### API — `api/connections.ts`

- Updated `createConnection` and `updateConnection` to accept typed payloads instead of `Partial<Connection>`.

### Store — `stores/connections.ts`

- Added `error` ref for surfacing fetch failures.
- Added `createConnection(payload)` — calls API, prepends result to local list.
- Added `updateConnection(id, payload)` — calls API, patches local list in-place.
- Added `deleteConnection(id)` — calls API, removes from local list.

### Components

**`ConnectionForm.vue`**
- Added `direction` select field (one-way / two-way).
- Emits typed `CreateConnectionPayload` on submit.
- Emits `cancel` event for dismissing the form.
- Resets fields after successful submit.

**`ConnectionCard.vue`**
- Shows direction alongside type.
- Added Delete button that emits `delete` event with connection ID.

**`ConnectionList.vue`**
- Added empty-state message when no connections exist.
- Forwards `delete` events from cards to parent.

**`ConnectionsPage.vue`**
- "New Connection" button toggles the form.
- Handles `create` and `delete` operations with error display.
- Shows loading spinner, fetch error, or connection list.

---

## Files Changed

```
backend/migrations/007_add_user_id_to_connections.up.sql    (new)
backend/migrations/007_add_user_id_to_connections.down.sql  (new)
backend/internal/db/queries/connections.sql                 (rewritten)
backend/internal/db/connections.sql.go                      (regenerated)
backend/internal/db/models.go                               (regenerated)
backend/internal/api/handlers/connections.go                (rewritten)
frontend/src/types/connection.ts                            (expanded)
frontend/src/api/connections.ts                             (updated)
frontend/src/stores/connections.ts                          (expanded)
frontend/src/components/connections/ConnectionForm.vue      (expanded)
frontend/src/components/connections/ConnectionCard.vue      (expanded)
frontend/src/components/connections/ConnectionList.vue      (expanded)
frontend/src/pages/ConnectionsPage.vue                     (rewritten)
```
