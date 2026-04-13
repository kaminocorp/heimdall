# Ingestion Redesign — Phase 2: Brand Logo SVG Components

**Status:** Complete  
**Plan:** `docs/executing/ingestion-page-redesign.md`

---

## What was built

`frontend/src/components/icons/ConnectorLogo.vue` — a single Vue component that renders native brand SVG marks for all supported connection types, controlled by a `type` prop.

## Design decisions

### Single component over separate files

Eight separate logo files would each need their own import and registration. Since the SVGs are pure markup with no logic, a single component with a `type` prop and `v-if`/`v-else-if` chain keeps imports clean — one component everywhere. Vue only renders the matching branch, so no wasted DOM nodes.

### Official paths from Simple Icons

Brand SVGs sourced from [Simple Icons](https://simpleicons.org/) — a maintained collection of 3,100+ brand marks, all normalised to a **24x24 viewBox** with a **single `<path>` element**. This guarantees:
- Consistent sizing via the `size` prop
- `fill="currentColor"` works uniformly across all brands
- Recognisable, legally simplified marks (no full-colour trademark issues)

### `currentColor` pattern

Every `fill` and `stroke` references `currentColor` instead of hardcoded hex values. This means colour is controlled entirely by CSS — `class="text-accent"` for green, `class="text-text-muted"` for grey. One component, any colour context, zero colour props.

### Custom icons for protocols

Webhook and Syslog are protocols without official brand marks. Custom icons were designed:
- **Webhook** — arrow-into-bracket (stroke-based, communicates "data arriving")
- **Syslog** — terminal window with prompt chevron and cursor line

## Props

| Prop | Type | Default | Purpose |
|------|------|---------|---------|
| `type` | `string` | required | Connector type key (matches `flows.ts` `connectorType`) |
| `size` | `number` | `32` | Width and height in pixels |

## Supported types

| Type key | Brand | Source | Style |
|----------|-------|--------|-------|
| `supabase` | Supabase | Simple Icons | Lightning bolt / S mark |
| `postgres` | PostgreSQL | Simple Icons | Elephant head silhouette |
| `github` | GitHub | Simple Icons | Octocat |
| `otlp` | OpenTelemetry | Simple Icons | Telescope / signal mark |
| `webhook_logs` | Webhook | Custom | Arrow-into-bracket (stroke) |
| `syslog` | Syslog | Custom | Terminal window (stroke) |
| `datadog` | Datadog | Simple Icons | Dog silhouette |
| `mysql` | MySQL | Simple Icons | Dolphin + wordmark |
| *(fallback)* | Generic | Custom | Rounded square with node dots |

## Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/components/icons/ConnectorLogo.vue` | **New** | Brand logo component |
| `frontend/src/pages/ConnectionsPage.vue` | Edit | Temporary logo gallery preview (to be replaced in Phase 6) |

## Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean |
| Visual gallery | All 8 types render recognisable brand marks at 36px in accent green |
