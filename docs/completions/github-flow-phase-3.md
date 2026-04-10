# GitHub Connection Flow Fixes — Phase 3 Completion

**Status:** Complete
**Date:** 2026-04-10
**Plan:** [github-connection-fixes.md](../executing/github-connection-fixes.md), Phase 3
**Validation:** `vue-tsc --noEmit` clean

---

## Goal

When a user selects GitHub in the connection wizard but a GitHub connection already exists (created by a previous OAuth callback), skip the install prompt and go straight to repo selection.

---

## Root cause

`StepGitHubInstall.vue` always shows "Install GitHub App" regardless of whether an installation already exists. After deleting a GitHub connection and re-adding it (or after the first successful install), users were forced through the full install flow again even though the GitHub App is still installed on their account. This created confusion and, combined with the Phase 2 bug, led to duplicate broken connections.

---

## Changes

### `frontend/src/components/connections/wizard/ConnectionWizard.vue`

**Change 1 — Added `manage-repos` emit**

New emit signature: `'manage-repos': [connectionId: string]`. This lets the wizard signal the parent page to open the repo selector for an existing connection.

**Change 2 — Early intercept in `selectPlatform()`**

When `flowId === 'github'`, checks `store.connections` for an existing `type === 'github'` connection. If found, emits `manage-repos` with the connection ID and `close` — the wizard closes and the repo selector opens immediately. If not found, proceeds with the normal Name + Install flow.

### `frontend/src/pages/ConnectionsPage.vue`

**Change — Wire up `@manage-repos` handler on `ConnectionWizard`**

Added `@manage-repos="(id: string) => { closeWizard(); openRepoSelector(id) }"` to the wizard component. `closeWizard()` hides the modal and refreshes connections; `openRepoSelector()` opens the repo selector panel for the given connection ID. Both functions already existed — no new logic needed in the parent.

---

## User-facing behaviour change

Before:
1. User clicks "+ New Connection" → selects GitHub → enters name → sees "Install GitHub App" → gets redirected to GitHub (even though app is already installed) → creates duplicate connection

After:
1. User clicks "+ New Connection" → selects GitHub → wizard closes, repo selector opens for the existing GitHub connection → user can toggle repos immediately

The install flow only appears when no GitHub connection exists at all (first-time setup).
