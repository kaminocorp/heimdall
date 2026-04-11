# GitHub Connection Flow Fixes — Implementation Plan

Three bugs prevent the GitHub App integration from working end-to-end. They share a root cause: the wizard assumes it owns connection creation, but for GitHub the **callback** creates the connection server-side, leaving the wizard out of sync.

**Bugs addressed:**
1. After installing the GitHub App, user is not redirected back to Heimdall
2. Re-adding GitHub via the wizard creates a broken connection with empty config (`missing installation_id in config`)
3. Wizard always shows "Install GitHub App" even when an installation already exists

---

## Phase 1 — Fix callback redirect

**Goal:** After GitHub installs the app and redirects to `/api/github/callback`, the user lands back on the Heimdall frontend connections page with `?github=installed`.

**Problem:** `github.go:240` redirects to `/connections?github=installed` — a bare backend path. In production the SPA catch-all may or may not serve the frontend; in dev it definitely doesn't (frontend is on `:5173`).

### Task 1.1 — Add `FRONTEND_URL` to config

**File:** `backend/internal/config/config.go`

Add a `FrontendURL` field (env: `FRONTEND_URL`, default `http://localhost:5173`). In production this would be set to the Heimdall frontend origin (e.g. `https://heimdall.fly.dev` or whatever the deployed URL is).

Check whether an existing config field already covers this (e.g. a base URL or allowed origins field) before adding a new one.

### Task 1.2 — Use `FRONTEND_URL` in callback redirect

**File:** `backend/internal/api/handlers/github.go:240`

Change:
```go
http.Redirect(w, r, "/connections?github=installed", http.StatusFound)
```
To:
```go
http.Redirect(w, r, s.Config.FrontendURL+"/connections?github=installed", http.StatusFound)
```

### Task 1.3 — Set `FRONTEND_URL` in production

Add `FRONTEND_URL` to `.env` (for local dev) and Fly.io secrets (for production). The production value depends on how the frontend is served — if it's behind the same origin as the backend (Fly serves the built SPA), the value should be that origin.

**Validation:** Install the GitHub App → browser lands on the Heimdall connections page with the success banner and repo selector open.

---

## Phase 2 — Prevent wizard from creating empty GitHub connections

**Goal:** The wizard must not call `createConnection` for GitHub flows, since the callback already creates the connection server-side.

**Problem:** `ConnectionWizard.vue:128-139` — `finish()` calls `createConnection()` when `createdConnectionId` is null. For GitHub, this creates a connection with `config: {}` (no `installation_id`), which breaks `ListGitHubRepos` and `TestGitHubConnection`.

### Task 2.1 — Skip connection creation for GitHub in the wizard

**File:** `frontend/src/components/connections/wizard/ConnectionWizard.vue`

In `finish()`, detect GitHub flows and skip `createConnection()`. The GitHub flow's last step (`StepGitHubInstall`) redirects the browser away — the wizard is abandoned. If the user somehow reaches `finish()` (e.g. clicks Done), it should just close the wizard without creating anything.

```ts
async function finish() {
  // GitHub connections are created server-side by the callback — don't duplicate.
  if (selectedFlow.value?.id === 'github') {
    emit('close')
    return
  }

  if (!createdConnectionId.value) {
    await createConnection()
    if (error.value) return
  }
  // ...
}
```

Also ensure the Install step's "Done" button is hidden or disabled. Currently `StepGitHubInstall` emits `valid: false` on mount which should disable "Continue", but the "Done" button on the last step has weaker gating (`!stepValid && isTestStep` — GitHub's last step is not a test step, so Done is always enabled).

**Fix:** Gate the Done button for GitHub flows as well. Simplest approach: have `StepGitHubInstall` never emit `valid: true`, and change the Done button's `:disabled` check from `!stepValid && isTestStep` to `!stepValid && (isTestStep || selectedFlow?.id === 'github')`. Or more cleanly, just keep the existing `valid: false` and change the disabled condition to just `!stepValid`.

### Task 2.2 — Clean up any existing broken connections

Optional but recommended: write a one-time migration or manual SQL to delete any `github` connections that have `config = '{}'` or are missing `installation_id`. Check the production database first.

**Validation:** Open wizard → GitHub → enter name → click Install → get redirected to GitHub → after install, no duplicate/empty connection exists in the DB.

---

## Phase 3 — Detect existing installation and skip to repo selection

**Goal:** When a user adds a GitHub connection but the app already has one installed, skip the install prompt and go straight to repo selection.

**Problem:** `StepGitHubInstall.vue` always shows "Install GitHub App" regardless of whether an installation already exists. After deleting and re-adding, users are forced through the full install flow again even though the GitHub App is still installed on their account.

### Task 3.1 — Add an API endpoint (or reuse existing) to check for existing GitHub connection

**Option A — Frontend-only:** Before showing `StepGitHubInstall`, check if the current app already has a `github` connection with a valid `installation_id` by fetching connections from the store. If it does, skip the install step and either close the wizard (connection already exists) or open the repo selector for that connection.

**Option B — Backend endpoint:** Add a `GET /api/apps/{appId}/github/status` endpoint that returns whether a GitHub installation exists and its connection ID. This is cleaner but more work.

Option A is simpler and sufficient — the connections are already loaded in the store.

### Task 3.2 — Modify wizard flow for GitHub when installation exists

**File:** `frontend/src/components/connections/wizard/ConnectionWizard.vue`

When the user selects GitHub from the platform grid:
1. Check `store.connections` for an existing `type === 'github'` connection
2. If found: skip the wizard, emit `close`, and open the repo selector for that connection (or show a message: "GitHub is already connected. Use the Repos button to manage repositories.")
3. If not found: proceed with the normal Name → Install flow

This could be done in `selectPlatform()`:
```ts
function selectPlatform(flowId: string) {
  if (flowId === 'github') {
    const existing = connectionsStore.connections.find(c => c.type === 'github')
    if (existing) {
      emit('close')
      // Signal parent to open repo selector — emit a new event or use a callback
      return
    }
  }
  // ... existing logic
}
```

The parent (`ConnectionsPage.vue`) would need to handle this — either via a new emit from the wizard or by checking after close.

### Task 3.3 — Alternative: make StepGitHubInstall smarter

Instead of modifying the wizard routing, make `StepGitHubInstall.vue` itself check for an existing installation and show a different UI:
- If installation exists: show "GitHub App is already installed on {account_login}. Click Done to finish." and emit `valid: true`
- If not: show the current install button

This requires passing the connections list (or the specific GitHub connection) as a prop, or having the step fetch it.

**Recommendation:** Task 3.2 is cleaner — intercept early in `selectPlatform()` rather than adding conditional logic deep in a step component.

**Validation:** Delete GitHub connection → click "+ New Connection" → select GitHub → see message that GitHub is already installed (or auto-open repo selector) instead of being asked to install again.

---

## Phase order and dependencies

Phases are independent and can be done in any order. Phase 2 is the highest priority (prevents data corruption). Phase 1 is second (broken redirect). Phase 3 is UX polish.

Recommended order: **2 → 1 → 3**

---

## Files touched

| Phase | File | Change |
|-------|------|--------|
| 1 | `backend/internal/config/config.go` | Add `FrontendURL` |
| 1 | `backend/internal/api/handlers/github.go:240` | Use full URL in redirect |
| 1 | `.env`, Fly.io secrets | Set `FRONTEND_URL` |
| 2 | `frontend/src/components/connections/wizard/ConnectionWizard.vue` | Skip create for GitHub, fix Done button gating |
| 3 | `frontend/src/components/connections/wizard/ConnectionWizard.vue` | Detect existing installation in `selectPlatform()` |
| 3 | `frontend/src/pages/ConnectionsPage.vue` | Handle repo-selector signal from wizard |
