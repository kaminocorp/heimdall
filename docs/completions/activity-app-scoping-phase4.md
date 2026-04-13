# Activity Feed App-Scoping — Phase 4: Frontend

## Status: Complete

## Summary

The Activity page now sends the currently selected `app_id` to the backend
on every fetch, and re-fetches when the user switches apps in the sidebar.
This completes the end-to-end app-scoping: the backend filters by `app_id`,
and the frontend provides it.

**Plan reference:** `docs/executing/activity-app-scoping.md`, Part 4

---

## 4a — API client (`api/logs.ts`)

Added `app_id?: string` to the `listLogs` params interface. The parameter
is optional — omitting it falls back to the backend's user-scoped
behaviour (backwards compatible).

## 4b — Logs store (`stores/logs.ts`)

Imported `useAppStore` and included `currentAppId` in every `fetchLogs`
call:

```ts
app_id: appStore.currentAppId ?? undefined
```

`useAppStore()` is called inside `fetchLogs` (not at store definition
time) to avoid circular dependency issues — this matches the lazy access
pattern already used elsewhere in the codebase.

The `?? undefined` conversion ensures `null` (no app selected) becomes
`undefined` (omitted from the query string) rather than being sent as
the literal string `"null"`.

## 4c — Activity page (`ActivityPage.vue`)

Added a `watch` on `appStore.currentAppId` that resets pagination and
re-fetches logs — the same pattern used by `ConnectionsPage`,
`SchedulesPage`, `NotificationsPage`, and `DashboardPage`.

---

## Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/api/logs.ts` | Edit | Added `app_id` to params interface |
| `frontend/src/stores/logs.ts` | Edit | Import `useAppStore`, include `currentAppId` in fetch |
| `frontend/src/pages/ActivityPage.vue` | Edit | Watch `currentAppId`, re-fetch on change |

## Verification

- `vue-tsc --noEmit` — clean type check
- `vitest run` — all 50 tests pass (including logs store tests)
