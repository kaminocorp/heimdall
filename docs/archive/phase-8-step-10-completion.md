# Phase 8, Step 10 — Frontend Changes: Completion Notes

Reference: [Phase 8 Monitoring Mode Plan](../executing/phase-8-monitoring-mode.md)

---

## Overview

Surfaced the multi-application data model (Step 1) and monitoring state (Steps 2-5) in both the backend API and the frontend. Adds organization onboarding, an application selector, per-app agent configuration, monitoring status on the dashboard, and monitoring/heartbeat badges in the agent log.

**Key design decision:** Rather than nesting routes under `/apps/:appId/...`, the current app is tracked client-side via the `useAppStore` Pinia store (persisted to `localStorage`). All app-scoped API calls use the store's `currentAppId`. This avoids a massive URL restructuring and keeps the existing route layout intact.

---

## Task 1 — Database: Updated Connection & Stats Queries

| Action | File |
|--------|------|
| Edited | `backend/internal/db/queries/connections.sql` |
| Edited | `backend/internal/db/queries/stats.sql` |
| Regenerated | `backend/internal/db/connections.sql.go` |
| Regenerated | `backend/internal/db/stats.sql.go` |

**New queries:**
- `ListConnectionsByApp(app_id)` — list connections scoped to an application
- `CreateConnection` — now includes `app_id` parameter (was missing from Step 1)
- `GetAppDashboardStats(app_id)` — log count, connection count, active connections scoped to app

---

## Task 2 — Backend: Organization & Onboarding Handlers

| Action | File |
|--------|------|
| Created | `backend/internal/api/handlers/organizations.go` |

**Endpoints:**

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/api/org` | `GetOrganization` | Get current user's organization |
| POST | `/api/onboard` | `Onboard` | Create org + first app + link user + default agent config |

**Onboarding flow:** Single POST creates the full hierarchy:
1. Create organization (name + slug)
2. Link user to org (`SetUserOrg`)
3. Create default application (name from request or "My Application")
4. Create default agent config for the app (mode: `off`, model: `claude-sonnet-4-6`, interval: 60s)

Returns both the organization and application in the response. Slug uniqueness is enforced by the DB — duplicates return 409 Conflict.

---

## Task 3 — Backend: Application & Per-App Handlers

| Action | File |
|--------|------|
| Created | `backend/internal/api/handlers/applications.go` |

**Endpoints:**

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/api/apps` | `ListApplications` | List apps in user's org |
| POST | `/api/apps` | `CreateApplication` | Create app + default agent config |
| GET | `/api/apps/{appId}` | `GetApplication` | Get single application |
| GET | `/api/apps/{appId}/connections` | `ListConnectionsByApp` | List connections for an app |
| GET | `/api/apps/{appId}/agent/config` | `GetAppAgentConfig` | Get per-app agent config |
| PUT | `/api/apps/{appId}/agent/config` | `UpdateAppAgentConfig` | Update per-app agent config |
| GET | `/api/apps/{appId}/monitoring/status` | `GetMonitoringStatus` | Get monitoring mode + last check time |
| GET | `/api/apps/{appId}/stats` | `GetAppDashboardStats` | App-scoped dashboard stats |

---

## Task 4 — Backend: Connection Handler Update

| Action | File |
|--------|------|
| Edited | `backend/internal/api/handlers/connections.go` |

`CreateConnection` now requires `app_id` in the request body. Returns 400 if missing or invalid. The app_id is passed through to the `CreateConnectionParams` which includes the new column.

---

## Task 5 — Backend: Route Wiring

| Action | File |
|--------|------|
| Rewritten | `backend/internal/api/router.go` |

New route groups:
- `/api/org` — organization (GET)
- `/api/onboard` — onboarding (POST)
- `/api/apps` — application CRUD
- `/api/apps/{appId}/*` — per-app routes (connections, config, monitoring, stats)

Legacy routes (`/api/agent/config`, `/api/stats`, `/api/connections`) kept for backwards compatibility.

---

## Task 6 — Frontend: Types

| Action | File |
|--------|------|
| Created | `frontend/src/types/organization.ts` |
| Edited | `frontend/src/types/connection.ts` |

New types: `Organization`, `Application`, `AppAgentConfig`, `MonitoringStatus`, `OnboardingPayload`, `OnboardingResponse`.

`Connection` type updated with `app_id` field. `CreateConnectionPayload` updated with required `app_id`.

---

## Task 7 — Frontend: API Layer

| Action | File |
|--------|------|
| Created | `frontend/src/api/organizations.ts` |
| Created | `frontend/src/api/applications.ts` |

`organizations.ts`: `getOrganization()`, `onboard(payload)`
`applications.ts`: `listApplications()`, `createApplication()`, `getApplication()`, `getAppAgentConfig()`, `updateAppAgentConfig()`, `getMonitoringStatus()`, `getAppStats()`, `listConnectionsByApp()`

---

## Task 8 — Frontend: App Store

| Action | File |
|--------|------|
| Created | `frontend/src/stores/app.ts` |

Central store for org + app state:

| State | Type | Description |
|-------|------|-------------|
| `organization` | `Organization \| null` | User's org |
| `applications` | `Application[]` | Apps in the org |
| `currentAppId` | `string \| null` | Selected app (persisted to localStorage) |
| `needsOnboarding` | `boolean` | True if user has no org |
| `loading` | `boolean` | Init loading state |

| Method | Description |
|--------|-------------|
| `init()` | Load org + apps, restore last selected app from localStorage |
| `selectApp(id)` | Switch current app |
| `onboard(...)` | Create org + app, set as current |
| `createApp(name)` | Create additional app |
| `reset()` | Clear state (used on logout) |

---

## Task 9 — Frontend: Router & Auth Flow

| Action | File |
|--------|------|
| Rewritten | `frontend/src/router/index.ts` |
| Edited | `frontend/src/App.vue` |
| Edited | `frontend/src/layouts/DefaultLayout.vue` |

**New route:** `/onboarding` → `OnboardingPage.vue`

**Auth guard update:** After auth check, if user `needsOnboarding` is true, redirects to `/onboarding` (unless already on onboarding or a public page).

**App.vue:** Now calls `app.init()` after `auth.init()` (only if authenticated) to load org state before the router guard evaluates.

**DefaultLayout:** Onboarding page skips the sidebar layout (same as login and public pages).

---

## Task 10 — Frontend: Onboarding Page

| Action | File |
|--------|------|
| Created | `frontend/src/pages/OnboardingPage.vue` |

Clean form with three fields: Organization Name, Slug (auto-generated from name on blur), and First Application Name. Matches the techno-brutalist design with monospace fonts, accent colors, and the animated pulse dot.

On submit, calls `app.onboard()` and redirects to `/dashboard`.

---

## Task 11 — Frontend: Sidebar App Selector

| Action | File |
|--------|------|
| Rewritten | `frontend/src/components/common/AppSidebar.vue` |

Added between the brand header and navigation sections:
- "Application" label
- `<select>` dropdown listing all apps in the org
- On change, calls `app.selectApp(id)`

User footer now shows organization name above the email. Logout calls `app.reset()` to clear app state.

---

## Task 12 — Frontend: Dashboard Page (Monitoring Card)

| Action | File |
|--------|------|
| Rewritten | `frontend/src/pages/DashboardPage.vue` |

**Changes from previous version:**
- 4-column grid (was 3): Monitoring, Agent, Connections, Log Ingestion
- Uses `getAppAgentConfig()`, `getMonitoringStatus()`, `listConnectionsByApp()`, `getAppStats()` — all app-scoped
- Monitoring card shows: mode (with status dot), interval (if periodic), last check time
- Watches `appStore.currentAppId` to reload data on app switch
- Shows app name in subtitle

---

## Task 13 — Frontend: Agent Config Page (Per-App)

| Action | File |
|--------|------|
| Rewritten | `frontend/src/pages/AgentConfigPage.vue` |

**Changes from previous version:**
- Loads/saves per-app config via `getAppAgentConfig()` / `updateAppAgentConfig()`
- Mode options: Continuous / Periodic / Off (was: Continuous / Scheduled / Off)
- Schedule field replaced with interval selector:
  - Preset buttons: 30s, 1m, 5m, 15m
  - Numeric input for custom values (10–86400 seconds)
  - Only shown when mode is "periodic"
- Updated model list: `claude-sonnet-4-6`, `claude-haiku-4-5-20251001`, `claude-opus-4-6`
- Watches `appStore.currentAppId` to reload on app switch

---

## Task 14 — Frontend: Agent Log Badges

| Action | File |
|--------|------|
| Edited | `frontend/src/components/log/LogEntry.vue` |

New entry type badges:
- **Monitor** — amber/warning-colored badge for `monitoring` entries (agent assessments of flagged logs)
- **Heartbeat** — green/ok-colored badge for `heartbeat` entries (monitor-alive confirmations)

Priority order: Monitor > Heartbeat > Agent (existing). The `entryTypeLabel` computed now maps both new types.

---

## Task 15 — Frontend: Connections Page (App Context)

| Action | File |
|--------|------|
| Edited | `frontend/src/pages/ConnectionsPage.vue` |
| Edited | `frontend/src/components/connections/ConnectionForm.vue` |

`ConnectionForm` now emits `Omit<CreateConnectionPayload, 'app_id'>` — the form doesn't know about app context. The page handler injects `appStore.currentAppId` when creating.

Page watches `appStore.currentAppId` to re-fetch connections on app switch.

---

## Verification

- `go build ./...` — clean
- `go vet ./...` — clean
- `go test ./internal/agent/` — all tests passing
- `npx vue-tsc --noEmit` — no TypeScript errors
- `npx vitest run` — 17 tests passing, no regressions

---

## Files Summary

| Action | File | Task |
|--------|------|------|
| Edited | `backend/internal/db/queries/connections.sql` | 1 |
| Edited | `backend/internal/db/queries/stats.sql` | 1 |
| Regenerated | `backend/internal/db/connections.sql.go` | 1 |
| Regenerated | `backend/internal/db/stats.sql.go` | 1 |
| Created | `backend/internal/api/handlers/organizations.go` | 2 |
| Created | `backend/internal/api/handlers/applications.go` | 3 |
| Edited | `backend/internal/api/handlers/connections.go` | 4 |
| Rewritten | `backend/internal/api/router.go` | 5 |
| Created | `frontend/src/types/organization.ts` | 6 |
| Edited | `frontend/src/types/connection.ts` | 6 |
| Created | `frontend/src/api/organizations.ts` | 7 |
| Created | `frontend/src/api/applications.ts` | 7 |
| Created | `frontend/src/stores/app.ts` | 8 |
| Rewritten | `frontend/src/router/index.ts` | 9 |
| Edited | `frontend/src/App.vue` | 9 |
| Edited | `frontend/src/layouts/DefaultLayout.vue` | 9 |
| Created | `frontend/src/pages/OnboardingPage.vue` | 10 |
| Rewritten | `frontend/src/components/common/AppSidebar.vue` | 11 |
| Rewritten | `frontend/src/pages/DashboardPage.vue` | 12 |
| Rewritten | `frontend/src/pages/AgentConfigPage.vue` | 13 |
| Edited | `frontend/src/components/log/LogEntry.vue` | 14 |
| Edited | `frontend/src/pages/ConnectionsPage.vue` | 15 |
| Edited | `frontend/src/components/connections/ConnectionForm.vue` | 15 |

---

## New API Endpoints Summary

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/org` | Get user's organization |
| POST | `/api/onboard` | Create org + app + link user |
| GET | `/api/apps` | List apps in user's org |
| POST | `/api/apps` | Create new application |
| GET | `/api/apps/{appId}` | Get application details |
| GET | `/api/apps/{appId}/connections` | List connections for app |
| GET | `/api/apps/{appId}/agent/config` | Get per-app agent config |
| PUT | `/api/apps/{appId}/agent/config` | Update per-app agent config |
| GET | `/api/apps/{appId}/monitoring/status` | Monitoring mode + last check |
| GET | `/api/apps/{appId}/stats` | App-scoped dashboard stats |
