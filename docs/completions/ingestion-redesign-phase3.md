# Ingestion Redesign — Phase 3: Connection Bubble Component

**Status:** Complete  
**Plan:** `docs/executing/ingestion-page-redesign.md`

---

## What was built

`frontend/src/components/connections/ConnectionBubble.vue` — a clickable bubble component that represents a single connection with its native brand logo, name, type label, and status indicator.

## Design decisions

### `<button>` over `<div @click>`

The bubble is a semantic `<button>` element. This gives keyboard focusability (Tab), activation (Enter/Space), screen reader announcement as interactive, and `:focus-visible` styling — all for free. A `<div @click>` would require manually adding `tabindex`, `role`, and `@keydown` handlers.

### Staggered CSS animation via custom property

Each bubble receives `--bubble-index` via `:style`, and the CSS uses `animation-delay: calc(var(--bubble-index) * 80ms)` for staggered fade-in. This keeps all animation on the compositor thread — no JavaScript timers, no `TransitionGroup` complexity.

### Logo container with elevated background

The logo sits inside a 48×48 container with `bg-elevated` and a border, giving it a "card within a card" feel that makes the brand mark pop against the bubble background. The logo inherits `text-accent` when active, `text-text-muted` when inactive, via the `currentColor` pattern from Phase 2.

## Props

| Prop | Type | Default | Purpose |
|------|------|---------|---------|
| `connection` | `Connection` | required | The connection data |
| `testing` | `boolean` | `false` | Whether this connection is currently being tested |
| `index` | `number` | `0` | Stagger index for mount animation delay |

## Emits

| Event | Payload | Purpose |
|-------|---------|---------|
| `click` | `Connection` | Bubble was clicked (parent opens detail modal) |

## Visual states

| State | Appearance |
|-------|-----------|
| **Default** | `bg-surface`, `border-border`, muted logo |
| **Active** | Accent border, accent glow shadow, green logo |
| **Hover** | Lifts 2px, `border-hover`, deeper shadow |
| **Testing** | Accent border with pulse animation, "Testing" label |
| **Focus-visible** | 2px accent outline with offset |

## Animation

- **Mount**: Fade-in + translate-up, 350ms ease-out, staggered by 80ms per index
- **Testing pulse**: 1.5s ease-in-out infinite opacity cycle
- **`prefers-reduced-motion`**: All animations disabled, immediate full opacity

## Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/components/connections/ConnectionBubble.vue` | **New** | Bubble component |
| `frontend/src/pages/ConnectionsPage.vue` | Edit | Replaced logo gallery preview with bubble row + nebula layout |

## Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean |
| Visual | Real connections render as bubbles with brand logos, staggered animation, hover/active states |
