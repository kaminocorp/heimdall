# Ingestion Redesign — Phase 5: Flow Lines (SVG Animated Connectors)

**Status:** Complete  
**Plan:** `docs/executing/ingestion-page-redesign.md`

---

## What was built

`frontend/src/components/connections/FlowLines.vue` — an SVG overlay that renders animated dashed paths from each connection bubble down to the agent nebula, suggesting continuous data flow into the agent.

## Design decisions

### Continuous flow vs one-time draw-in

The existing `BlueprintView` uses a one-shot `stroke-dashoffset` animation that "draws" lines on mount — a reveal effect. FlowLines uses a **perpetual** animation: `stroke-dashoffset` cycles from 24 to 0 every 2.5 seconds, creating continuous downward motion. This communicates "data is flowing right now" rather than "these things are connected."

### Two overlapping paths per line

Each connection gets two `<path>` elements:
1. **Static base** at 8% opacity — structural reference, always visible
2. **Animated flow** at 30% opacity with glow filter — the moving data

This layering ensures connections are visible even with `prefers-reduced-motion` (animated path stops but base remains).

### Quadratic bezier with biased control point

Paths are quadratic beziers from each bubble's bottom-center to the nebula's top-center. The control point is placed at 50% horizontal and 55% vertical between source and target, creating a gentle curve that fans out from the nebula entry point rather than rigid straight lines.

### Ref collection from Vue components

`v-for` with template refs on Vue components returns component instances, not DOM elements. The pattern `:ref="(el: any) => { bubbleEls[i] = el?.$el ?? el }"` extracts the root DOM element via `$el`, falling back to the raw element for native nodes.

## Props

| Prop | Type | Default | Purpose |
|------|------|---------|---------|
| `bubbleEls` | `(HTMLElement \| null)[]` | required | Refs to each bubble's root DOM node |
| `nebulaEl` | `HTMLElement \| null` | required | Ref to the nebula container element |
| `count` | `number` | required | Connection count — triggers recompute on change |

## Exposed methods

| Method | Purpose |
|--------|---------|
| `recompute()` | Manually trigger line position recalculation |

## Animation parameters

| Parameter | Value | Purpose |
|-----------|-------|---------|
| Dash pattern | `4 8` (4px dash, 8px gap) | Sparse — hint of flow, not a solid pipe |
| Animation duration | 2.5s linear infinite | Smooth continuous downward motion |
| Stagger delay | 300ms per line index | Lines don't all start synchronized |
| Base line opacity | 0.08 | Faint structural reference |
| Flow line opacity | 0.30 | Visible but not dominant |
| Glow filter | `feGaussianBlur` stdDeviation 2.5 | Subtle phosphor glow on animated line |

## Responsive & accessibility

- **Mobile (< 768px):** Flow lines container is `display: none` — the spatial metaphor doesn't work in a stacked layout
- **`prefers-reduced-motion`:** Animation stopped, static dashoffset 0 — lines visible but frozen
- **Pointer events:** Container is `pointer-events: none` — clicks pass through to bubbles and nebula

## Recompute triggers

Lines recompute on:
1. `ResizeObserver` firing on the container (window resize, layout shift)
2. `count` prop changing (connection added/removed)
3. `nebulaEl` prop changing (nebula mounts)

## Files changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/components/connections/FlowLines.vue` | **New** | SVG flow line component |
| `frontend/src/pages/ConnectionsPage.vue` | Edit | Added FlowLines import, `bubbleEls`/`nebulaEl` refs, wrapped bubbles + nebula in relative container with FlowLines overlay |

## Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean |
| Visual | Animated dashed lines flow from each bubble into the nebula |
| Resize | Lines recompute correctly on window resize |
| Mobile | Lines hidden below 768px |
