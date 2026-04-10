# GitHub Connection Flow Fixes — Phase 2 Completion

**Status:** Complete
**Date:** 2026-04-10
**Plan:** [github-connection-fixes.md](../executing/github-connection-fixes.md), Phase 2
**Validation:** `vue-tsc --noEmit` clean

---

## Goal

Prevent the connection wizard from creating a broken, empty-config GitHub connection. The GitHub OAuth callback (`github.go:222-233`) already creates the connection server-side with a valid `installation_id` — the wizard must not duplicate it.

---

## Root cause

The wizard's `finish()` function assumed all flows need a client-side `createConnection()` call. For most connectors (webhook, postgres, supabase) this is correct — the wizard collects config and creates the connection. But GitHub is different: the user is redirected to GitHub, and the **backend callback** creates the connection on return. By the time the user might interact with the wizard again, the connection already exists.

Two failure paths:

1. User clicks "Done" on the GitHub install step (which was incorrectly enabled) → `finish()` calls `createConnection()` with `config: {}` → creates a connection missing `installation_id` → `ListGitHubRepos` and `TestGitHubConnection` both fail with "missing installation_id in config".

2. User returns to Heimdall after GitHub install, doesn't see the connection (Phase 1 redirect bug), opens the wizard again, creates another empty connection.

---

## Changes

### `frontend/src/components/connections/wizard/ConnectionWizard.vue`

**Change 1 — Early return in `finish()` for GitHub flows**

Added a guard at the top of `finish()` that checks `selectedFlow.value?.id === 'github'`. If true, it emits `close` and returns immediately without calling `createConnection()`. The comment explains why: "GitHub connections are created server-side by the OAuth callback."

**Change 2 — Done button disabled for all non-valid last steps**

The Done button's `:disabled` condition was `!stepValid && isTestStep` — meaning it was only disabled when `stepValid` was false *and* the current step was a test step. For GitHub's install step (which is the last step but not a test step), Done was always enabled even though `StepGitHubInstall` emits `valid: false` on mount.

Changed to simply `!stepValid`. This correctly disables Done for any last step that hasn't been validated, including GitHub's install step. No impact on test steps — they already control `stepValid` via their own emit.

---

## Why this is safe

- `StepGitHubInstall` redirects the browser to GitHub on click — the wizard is abandoned (modal state is lost on navigation). The `finish()` guard is a safety net for edge cases where `finish()` is somehow reached.
- `handleClose()` only deletes connections tracked by `createdConnectionId`, which is never set for GitHub flows after this change. The callback-created connection is not touched.
- The Done button change is strictly more correct for all flows — if a step says it's not valid, Done should be disabled regardless of step type.
- Existing flows (webhook, postgres, supabase, syslog, OTLP) are unaffected because their last steps either emit `valid: true` or are test steps.

---

## Remaining work

- **Phase 1** (callback redirect): User still won't land back on the connections page after GitHub install — the backend redirects to a bare `/connections` path.
- **Phase 3** (detect existing installation): Wizard still shows "Install GitHub App" even when already installed.
- **Cleanup**: Any existing broken `github` connections with empty config in production should be manually deleted.
