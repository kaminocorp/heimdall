# Ingestion Redesign — Phase 4: Connection Detail Modal

**Status:** Complete  
**Plan:** `docs/executing/ingestion-page-redesign.md`

---

## What was built

`frontend/src/components/connections/ConnectionDetailModal.vue` — a full-info modal that opens when a user clicks a connection bubble. Shows all connection details, type-specific configuration, metadata, and action buttons for edit/test/delete.

## Design decisions

### Close-then-delegate action pattern

When the user clicks an action (Edit, Ping, Delete), the modal emits the action event and the parent handler closes the detail modal before opening the next surface. This prevents two modals stacking and keeps the lifecycle clean. Each action in the parent follows: `selectedConnection = null` → trigger action handler.

### Two-step delete confirmation

Reuses the same pattern from `ConnectionCard.vue` — first click shows "Confirm Delete?" in red, second click executes. 3-second auto-reset timer if the user doesn't confirm. Lightweight alternative to a confirmation dialog that keeps the user in-context.

### Masked fields with reveal toggle

Passwords, access tokens, and bearer tokens display as `****` by default. A "Show/Hide" toggle per field uses a `Set<string>` of revealed field labels. Short values (<= 8 chars) are fully masked; longer values show the first 4 characters followed by asterisks for scannability without full exposure.

### Follows existing modal pattern

Matches `ConnectionTestModal.vue` exactly: fixed backdrop with `bg-black/60 backdrop-blur-sm`, `@click.self` to close on backdrop, Escape key handler via `document.addEventListener`, `border-border rounded-lg bg-bg-surface shadow-2xl`, header/body/footer sections.

## Props

| Prop | Type | Default | Purpose |
|------|------|---------|---------|
| `connection` | `Connection` | required | The connection to display |

## Emits

| Event | Payload | Purpose |
|-------|---------|---------|
| `close` | — | Modal should close |
| `edit` | `Connection` | Open edit form for this connection |
| `test` | `string` (id) | Trigger connection test |
| `delete` | `string` (id) | Delete this connection |
| `manage-repos` | `string` (id) | Open GitHub repo selector |

## Modal sections

1. **Header** — Brand logo (via `ConnectorLogo`) + name + type/direction + `StatusBadge` + close button
2. **Config details** — Type-specific key-value list:
   - PostgreSQL: host, port, database, user, password (masked), ssl_mode
   - Supabase: project_ref, access_token (masked), poll_interval, polling tables count
   - Webhook: bearer token (masked)
   - Syslog: port, protocol
   - GitHub/OTLP: minimal (actions handle the rest)
3. **Metadata** — Created date, updated date, connection ID (selectable text)
4. **Action bar** — [Repos] (GitHub only) | [Ping] | [Edit] on left, [Delete] on right

## Type-specific config rendering

The `configDetails` computed property uses a `switch` on `connection.type` to extract relevant fields from the untyped `config: Record<string, unknown>`. Each field specifies `{ label, value, masked? }`. This keeps the template generic while the logic handles per-type specifics.

## Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/components/connections/ConnectionDetailModal.vue` | **New** | Detail modal component |
| `frontend/src/pages/ConnectionsPage.vue` | Edit | Added `selectedConnection` ref, wired bubble click → detail modal, added modal to template with action event handlers |

## Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean |
| Bubble click | Opens detail modal with correct connection data |
| Escape / backdrop click | Closes modal |
| Edit action | Closes modal, opens edit form |
| Ping action | Closes modal, opens test modal |
| Delete action | Two-step confirmation, executes delete |
| Masked fields | Passwords/tokens hidden with Show/Hide toggle |
