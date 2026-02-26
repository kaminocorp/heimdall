# Connection Test on Create

## Problem

When a user adds a connection, it always shows **"inactive"**. There's no way to know if the credentials are valid or the target is reachable until the agent tries to use the connection later — at which point the error is buried in agent logs.

## Goal

Run a connectivity test automatically after creating a connection. If the test passes, set status to `active`. If it fails, set status to `error` and surface the failure message to the user immediately.

## Design

### New endpoint

```
POST /api/connections/{id}/test
```

- Authenticated (JWT required, user-scoped)
- Fetches the connection, builds the appropriate connector, attempts to connect
- Updates the connection status to `active` or `error` via `UpdateConnectionStatus`
- Returns JSON: `{ "success": true/false, "message": "..." }`

### Test logic by connection type

| Type | What the test does |
|------|-------------------|
| `postgres` | `database.New(config)` → `Connect(ctx)` → `Close(ctx)`. 5-second timeout. |
| `webhook_logs` | No remote target to test — auto-pass, set `active`. |
| `syslog` | Stub connector — auto-pass, set `active`. |
| `github` | Stub connector — auto-pass, set `active`. |

When real connectors are implemented for syslog/github, this handler naturally gains real testing without changes — it just needs to call `Connect()`.

### Frontend flow

1. User fills out the form and clicks Create
2. Frontend calls `POST /api/connections` → connection created with `status: "inactive"`
3. Frontend immediately calls `POST /api/connections/{id}/test`
4. While testing: connection card shows a brief "Testing..." indicator
5. On success: card updates to `active` status badge
6. On failure: card updates to `error` status badge + a toast/banner shows the error message (e.g. "connection refused", "password authentication failed")

## Changes

### Backend (2 files)

1. **`internal/api/handlers/connections.go`** — Add `TestConnection` handler:
   - Extract user ID from JWT context
   - Parse connection ID from URL param
   - Fetch connection via `GetConnectionByUser` (user-scoped)
   - Switch on `conn.Type`:
     - `"postgres"`: create connector via `database.New(conn.Config)`, call `Connect()` with 5s timeout context, defer `Close()`
     - All others: auto-pass
   - On success: `UpdateConnectionStatus(id, "active")`, return `{ "success": true, "message": "Connection established" }`
   - On error: `UpdateConnectionStatus(id, "error")`, return `{ "success": false, "message": "<error detail>" }`
   - New imports: `context`, `time`, `connectors/database`

2. **`internal/api/router.go`** — Register the route:
   ```go
   r.Post("/{id}/test", s.TestConnection)
   ```

### Frontend (4 files)

3. **`src/api/connections.ts`** — Add API function:
   ```ts
   export function testConnection(id: string) {
     return client.post<{ success: boolean; message: string }>(`/connections/${id}/test`)
   }
   ```

4. **`src/stores/connections.ts`** — Add store action:
   ```ts
   async function testConnection(id: string) {
     const { data } = await connectionsApi.testConnection(id)
     // Update the local connection's status based on result
     const idx = connections.value.findIndex(c => c.id === id)
     if (idx !== -1) {
       connections.value[idx].status = data.success ? 'active' : 'error'
     }
     return data
   }
   ```

5. **`src/pages/ConnectionsPage.vue`** — Update `handleCreate`:
   - After `store.createConnection(payload)` succeeds, call `store.testConnection(conn.id)`
   - Show error message from test result if it fails

6. **`src/components/connections/ConnectionCard.vue`** — Add a "testing" visual state:
   - Accept an optional `testing` prop
   - When `testing === true`, show a subtle pulsing indicator on the status badge area
