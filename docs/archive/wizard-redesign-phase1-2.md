# Wizard Redesign — Phases 1 & 2: Flow Restructure + Platform Logos

**Status:** Complete  
**Plan:** `docs/executing/connection-wizard-redesign.md`

---

## Phase 1: Restructure `flows.ts`

### What changed

The `PlatformFlow` interface and all flow entries were updated to replace the old `category` field with a new `section` field and a `subtitle` field.

### Old taxonomy

```
category: 'log_source' | 'database' | 'generic'
```

Three flat categories grouping connectors by technical type. No indication of what *happens* when you add one.

### New taxonomy

```
section: 'platform_log' | 'direct_protocol' | 'agent_tool'
subtitle: string   // e.g. "via Syslog drain", "Query during investigations"
```

Three sections that communicate *purpose and consequence*:

| Section | Meaning | User question it answers |
|---------|---------|--------------------------|
| `platform_log` | Pick your platform, we handle the wiring | "Where do my logs come from?" |
| `direct_protocol` | Choose the transport protocol directly | "I know I want Webhook/Syslog/OTLP" |
| `agent_tool` | On-demand query access — not a log source | "What can the agent query during investigations?" |

### New platform entries (all coming soon)

| ID | Name | Section | Underlying connector | Subtitle |
|----|------|---------|---------------------|----------|
| `flyio` | Fly.io | `platform_log` | `syslog` | via Syslog drain |
| `vercel` | Vercel | `platform_log` | `webhook_logs` | via Log Drain |
| `render` | Render | `platform_log` | `syslog` | via Syslog drain |
| `railway` | Railway | `platform_log` | `webhook_logs` | via HTTP Log Drain |
| `heroku` | Heroku | `platform_log` | `syslog` | via Syslog drain |
| `aws` | AWS | `platform_log` | `otlp` | via CloudWatch + OTLP |
| `digitalocean` | DigitalOcean | `platform_log` | `syslog` | via Log Forwarding |

All have `available: false` and `steps: []`. When wired up in future, each starts the appropriate existing connector flow (syslog, webhook, or otlp) — no new backend work needed.

### Removed entry

`datadog` — removed because it didn't fit cleanly into any section and was never wired up. Can be re-added as a `platform_log` entry when Datadog support is built.

### Existing flows preserved

All 6 existing live flows (`supabase`, `webhook_logs`, `syslog`, `otlp`, `postgres`, `github`) keep their `id`, `connectorType`, `direction`, and `steps` unchanged. Only `category` was replaced with `section` and `subtitle` was added.

### Bridge fix: PlatformGrid.vue

The grid component referenced `PlatformFlow['category']` which no longer exists. Updated to use `PlatformFlow['section']` with interim section labels. This keeps the wizard functional while Phase 3 does the full grid rewrite.

---

## Phase 2: Platform Logos

### What changed

Added 7 new brand SVG marks to `ConnectorLogo.vue` for the new platform entries.

### New logos

| Type key | Brand | Source | Mark |
|----------|-------|--------|------|
| `flyio` | Fly.io | Simple Icons | Bird/flame mark |
| `vercel` | Vercel | Simple Icons | Triangle mark |
| `render` | Render | Simple Icons | Stylised R mark |
| `railway` | Railway | Simple Icons | Rail/circle mark |
| `heroku` | Heroku | Simple Icons v11 | H-in-rounded-rect mark |
| `aws` | AWS | Simple Icons v11 | "aws" wordmark + smile arrow |
| `digitalocean` | DigitalOcean | Simple Icons | Droplet mark |

Note: Heroku and AWS were removed from Simple Icons v16+ due to brand policy changes. Paths sourced from Simple Icons v11.14.0.

All logos follow the existing pattern: 24×24 viewBox, `fill="currentColor"`, single `<path>` element, tinted via CSS `color` property.

---

## Files changed

| File | Phase | Change |
|------|-------|--------|
| `frontend/src/components/connections/wizard/flows.ts` | 1 | New `section`/`subtitle` fields, 7 platform entries, removed `datadog`, removed `category` |
| `frontend/src/components/icons/ConnectorLogo.vue` | 2 | Added 7 platform logo SVGs |
| `frontend/src/components/connections/wizard/PlatformGrid.vue` | 1 (bridge) | Updated `category` → `section` references to prevent runtime break |

## Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean |
| `vitest run` | 52/52 tests pass |
| Wizard opens | Functional with interim section labels |
| All existing flows | Steps/connectorType/direction unchanged |
