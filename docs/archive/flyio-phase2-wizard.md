# Fly.io Phase 2 Completion — Dual-Mode Wizard

**Date:** 2026-04-15
**Proposal:** `docs/executing/flyio-log-integration.md`
**Scope:** Tasks 2 + 3 from the proposal — wizard redesign with drain/polling mode selection

---

## What Was Done

Implemented the Fly.io connection wizard with a dual-mode flow. Users choose between **Log Drain** (recommended, near-real-time via webhook) and **API Polling** (zero-setup, uses the rewritten Phase 1 poller). The wizard dynamically switches `connectorType` and visible steps based on the mode selection.

## Files Created

| File | Purpose |
|------|---------|
| `frontend/src/components/connections/wizard/steps/StepFlyioMode.vue` | Mode selection: drain vs polling toggle |
| `frontend/src/components/connections/wizard/steps/StepFlyioAuth.vue` | Polling path: app_name + api_token credentials |
| `frontend/src/components/connections/wizard/steps/StepFlyioDrainSetup.vue` | Drain path: webhook info + Log Shipper CLI setup commands |

## Files Modified

| File | Change |
|------|--------|
| `frontend/src/components/connections/wizard/flows.ts` | Replaced Fly.io stub with full flow definition; added `getFlyioSteps()` and `getFlyioConnectorType()` helpers |
| `frontend/src/components/connections/wizard/ConnectionWizard.vue` | Added `effectiveSteps` and `effectiveConnectorType` computeds for dynamic branching; strip `flyio_mode` from config before API call |

## Wizard Flow

```
Step 1: Name           → Connection name (shared)
Step 2: Mode           → "Log Drain (recommended)" or "API Polling"

Drain path:
  Step 3: Drain Setup  → Display-only: webhook URL/token info + copy-paste CLI commands
                          Creates a webhook_logs connection (reuses existing webhook pipeline)

Polling path:
  Step 3: Auth         → Collects app_name + api_token
  Step 4: Test         → Tests the flyio connection against Fly.io Logs API
                          Creates a flyio connection (uses the Phase 1 app-level poller)
```

## Key Design Decisions

### Dynamic `connectorType` — why and how

This is the first flow in Heimdall where the connector type isn't fixed at platform selection time. The wizard previously assumed a static `connectorType` per `PlatformFlow` — fine for platforms with a single integration path, but Fly.io offers two fundamentally different architectures.

**How it works:**
1. `flows.ts` defines all possible Fly.io steps in a single flow (5 steps total)
2. `getFlyioSteps()` filters to the mode-appropriate subset at render time
3. `getFlyioConnectorType()` returns `'webhook_logs'` for drain or `'flyio'` for polling
4. `ConnectionWizard.vue` uses `effectiveSteps` and `effectiveConnectorType` computeds — these fall through to the original `selectedFlow` values for all non-Fly.io flows, keeping the change minimal

**Why not two separate flows?** The proposal explicitly calls for a single unified Fly.io entry in the platform grid. Two entries would confuse users who don't know the difference between drain and polling — the mode step explains the trade-offs in context.

### Config key cleanup

The `flyio_mode` key is wizard-internal state — the backend doesn't need it. It's stripped via destructuring before the `createConnection()` API call:
```typescript
const { flyio_mode: _, ...cleanConfig } = state.config
```

### Drain path creates `webhook_logs`, not `flyio`

The drain path doesn't use the Fly.io connector at all. It creates a standard `webhook_logs` connection — the Fly Log Shipper posts to Heimdall's existing webhook endpoint. This reuses proven infrastructure with zero new backend code.

### No test step for drain path

The drain path ends at the setup instructions step (no test step). The webhook connection is created server-side with an auto-generated token, but the Log Shipper hasn't been deployed yet — there's nothing to test until the user runs the CLI commands. The polling path has a test step because the credentials can be validated immediately.

## Component Details

### StepFlyioMode

- Two large toggle buttons with descriptions
- Drain pre-selected as default (recommended)
- "Recommended" badge on drain, "Zero setup" badge on polling
- Always valid (one option is always selected)
- Stores `flyio_mode: 'drain' | 'polling'` in `state.config`

### StepFlyioAuth

- Two form fields: app name and API token (password input)
- Follows exact pattern of `StepSupabaseAuth`
- Valid when both fields are non-empty
- Stores `app_name` and `api_token` in `state.config`

### StepFlyioDrainSetup

- Display-only (no user input), auto-valid on mount
- Follows pattern of `StepWebhookSetup`
- Shows 4 copyable CLI commands for deploying Fly Log Shipper
- Explains how the drain works (NATS → HTTP → Heimdall)
- Uses existing `CopyableField` component

## Build Status

- `vue-tsc --noEmit` — clean (0 type errors)
- `vite build` — clean (1.55s)
- `vitest run` — 52/52 tests pass

## What's Next (Phase 3+)

Per the proposal, remaining tasks:
1. **Vector webhook parser** — test whether the Fly Log Shipper's HTTP sink format auto-detects via existing parsers; add a dedicated parser if not
2. **End-to-end testing** — both drain and polling paths with real Fly.io apps
