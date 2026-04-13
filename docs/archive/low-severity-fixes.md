# Low-Severity Fixes — Issues #3 and #6

**Date:** 2026-04-12
**Assessment reference:** `docs/plans/code-assessment-2026-04-12.md`, issues #3 and #6

---

## Issue #3 — Listener Start Error Handling

### What the assessment flagged

"When a syslog listener fails to start during connection creation, the handler returns HTTP 201 (success) after setting the connection status to 'error' in the database. The client sees success while the listener is broken."

### What was actually happening

On closer inspection, the response body already included `"status": "error"` — the handler mutates `conn.Status = "error"` on the local struct before encoding. The 201 HTTP status code is correct REST semantics: the connection resource *was* created, the listener just couldn't start. The frontend receives the full connection object with the error state.

The real issue was that `UpdateConnectionStatus` errors were silently discarded:

```go
// Before — DB error silently ignored:
_ = queries.UpdateConnectionStatus(r.Context(), ...)

// After — DB error logged for observability:
if err := queries.UpdateConnectionStatus(r.Context(), ...); err != nil {
    slog.Error("failed to persist listener error status", "connection_id", conn.ID, "err", err)
}
```

### Files changed

- `backend/internal/api/handlers/connections.go` — CreateConnection and UpdateConnection syslog error paths (4 occurrences of `_ =` replaced with error logging)

---

## Issue #6 — Direct Store State Mutation

### What the assessment flagged

`ConnectionsPage.vue` directly mutated `store.loading`, `store.error`, and `store.connections` instead of calling a store action, bypassing Pinia's action semantics and spreading state-management logic into the component.

### Before

```typescript
// ConnectionsPage.vue — component manages store state directly
async function fetchAppConnections() {
  const appId = appStore.currentAppId
  if (!appId) return
  store.loading = true
  store.error = null
  try {
    store.connections = await listConnectionsByApp(appId)
  } catch (e: unknown) {
    store.error = extractApiError(e, 'Failed to load connections')
  } finally {
    store.loading = false
  }
}
```

### After

```typescript
// stores/connections.ts — new store action
async function fetchConnectionsByApp(appId: string) {
  loading.value = true
  error.value = null
  try {
    connections.value = await listConnectionsByApp(appId)
  } catch (e: unknown) {
    error.value = extractApiError(e, 'Failed to load connections')
  } finally {
    loading.value = false
  }
}

// ConnectionsPage.vue — delegates to store
async function fetchAppConnections() {
  const appId = appStore.currentAppId
  if (!appId) return
  await store.fetchConnectionsByApp(appId)
}
```

### Files changed

- `frontend/src/stores/connections.ts` — added `fetchConnectionsByApp(appId)` action, added `listConnectionsByApp` import, exported new action
- `frontend/src/pages/ConnectionsPage.vue` — replaced direct store mutation with `store.fetchConnectionsByApp()` call, removed unused `listConnectionsByApp` import

---

## Verification

- `go build ./...` + `go vet ./...` — passes
- `go test ./internal/api/handlers/` — all tests pass
- `vue-tsc --noEmit` — type-check passes
- `vitest run` — all 50 tests pass
