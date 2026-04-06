# Connections Page: Blueprint Visualization & Wizard Discard Guard

## Context

Two UX problems on the Connections page prompted this work:

1. **Lost wizard progress** — Clicking outside the connection wizard modal (or pressing Escape / the X button) immediately closed it, discarding all entered data with no warning. Users who accidentally clicked the backdrop lost multi-step form progress.

2. **Abstract connection display** — All connections were rendered as identical rectangular cards in a flat grid. There was no visual sense of _what_ each connection represents (log source vs database vs integration) or how the pieces fit together architecturally.

## What Was Done

### Feature 1: Wizard Discard Confirmation

**File modified:** `frontend/src/components/connections/wizard/ConnectionWizard.vue`

**Approach:** A gatekeeper pattern intercepts all three close paths and conditionally shows a confirmation overlay.

**Changes:**

1. **`isDirty` computed property** — Returns `true` when any of four conditions hold: a platform flow is selected, the name field has content, config entries exist, or a connection was already created in the DB mid-wizard. This covers every level of user progress.

2. **`requestClose()` gatekeeper** — Replaces all direct `handleClose()` calls. Checks `isDirty`; if clean, closes immediately. If dirty, sets `showDiscardConfirm = true` to show the overlay.

3. **Three close paths rewired:**
   - Backdrop: `@click.self="requestClose"` (was `handleClose`)
   - Escape keydown handler: calls `requestClose()` (was `handleClose()`)
   - X button: `@click="requestClose"` (was `handleClose`)

4. **Inline confirmation overlay** — Renders inside the modal (not a separate modal) using `absolute inset-0` positioning over the wizard body. Shows "Discard changes?" with Cancel and Discard buttons. The Discard button uses `bg-status-critical` for visual weight. Cancel dismisses the overlay and returns to the wizard.

5. **`relative` added to modal container** — Scopes the absolute-positioned overlay to the wizard boundaries.

**Why inline overlay instead of a separate modal?** Avoids z-index stacking complexity and keeps the wizard content visible behind the semi-transparent backdrop, reinforcing what the user is about to discard.

---

### Feature 2: Architectural Blueprint Visualization

Replaced the flat card grid with a visual infrastructure diagram showing Heimdall as a central hub with categorized zones radiating outward.

#### New Components

**`frontend/src/components/connections/ViewToggle.vue`**
- Two-button toggle group (Blueprint / List) using `v-model` pattern
- Active state uses `bg-accent/20` highlight
- Matches existing design system typography (mono, uppercase, tracking-wider)

**`frontend/src/components/connections/BlueprintNode.vue`**
- Compact connection card (single-line layout, `px-3 py-2`) — a tighter version of `ConnectionCard.vue`
- Shows: 2-letter icon badge → name (truncated) → StatusBadge → hover actions (Ping, Edit, Del, Repos)
- Delete uses the same 3-second confirmation pattern as the original ConnectionCard
- `glow-active` class on active connections for visual consistency

**`frontend/src/components/connections/BlueprintZone.vue`**
- Category container (Log Sources / Databases / Integrations)
- Inline SVG icons per category: server rack (logs), cylinder (databases), code brackets (integrations)
- Renders a vertical list of `BlueprintNode` components with staggered fade-in animation
- **Empty state:** Dashed-border placeholder button ("+ Add a log source" / "+ Connect a database" / "+ Add an integration") that emits an `add` event to open the wizard
- All CRUD events bubble up through emits

**`frontend/src/components/connections/BlueprintView.vue`**
- Main blueprint container with three-column CSS Grid layout:
  ```
  [Log Sources]     [ HEIMDALL ]     [Databases]
  [Integrations]    [   HUB    ]
  ```
- **Heimdall hub** at center: eye icon in a glowing circle (`border-2 border-accent/40`, `blur-xl animate-pulse` glow layer), "HEIMDALL" label below
- **SVG connection lines** drawn between hub and each zone using quadratic bezier paths:
  - Coordinates computed dynamically via `getBoundingClientRect()` relative to the container
  - `ResizeObserver` on the container triggers recomputation on layout changes
  - Dashed stroke (`stroke-dasharray: 6 4`) with accent color at 0.25 opacity and a subtle glow filter
  - `stroke-dashoffset` animation creates a draw-in effect on mount
  - Hidden on mobile (`hidden md:block`) since vertical stacking implies hierarchy
- **Category mapping** imports `flows` from `flows.ts` and builds `typeToCategory` / `iconMap` lookups via `Object.fromEntries()` — single source of truth, no duplication
- **Responsive:** Mobile stacks vertically (hub → log sources → databases → integrations) using grid `order` classes
- All events (delete, edit, test, manage-repos, add) forwarded to parent

#### Modified Component

**`frontend/src/pages/ConnectionsPage.vue`**

1. **Imports** added for `BlueprintView` and `ViewToggle`
2. **View mode state** with localStorage persistence:
   ```ts
   const viewMode = ref<'blueprint' | 'list'>(
     localStorage.getItem('heimdall_connections_view') || 'blueprint'
   )
   watch(viewMode, (v) => localStorage.setItem('heimdall_connections_view', v))
   ```
3. **ViewToggle** placed in the page header alongside the "+ New Connection" button
4. **Conditional rendering:** `BlueprintView` when `viewMode === 'blueprint'`, existing `ConnectionList` otherwise
5. **`@add` event** from BlueprintView's empty zone prompts wired to `openCreate()` to open the wizard

---

## Architecture Decisions

| Decision | Rationale |
|----------|-----------|
| CSS Grid (not flexbox/canvas) for layout | Grid's `grid-template-columns` + `row-span` gives clean 3-column desktop layout that degrades to single-column mobile via responsive classes |
| SVG for connection lines (not CSS borders) | Bezier curves look more like an architecture diagram; `ResizeObserver` keeps them accurate across layout shifts |
| Category mapping from `flows.ts` | Single source of truth — adding a new connector type to flows automatically places it in the correct blueprint zone |
| Inline confirmation overlay (not separate modal) | Avoids z-index stacking; keeps wizard context visible behind semi-transparent backdrop |
| localStorage for view preference | Lightweight, no API call needed, persists across sessions |

## Files Changed

| File | Status | Purpose |
|------|--------|---------|
| `frontend/src/components/connections/wizard/ConnectionWizard.vue` | Modified | Discard confirmation (isDirty, requestClose, overlay) |
| `frontend/src/components/connections/ViewToggle.vue` | New | Blueprint/List toggle button group |
| `frontend/src/components/connections/BlueprintNode.vue` | New | Compact connection card for blueprint zones |
| `frontend/src/components/connections/BlueprintZone.vue` | New | Category container with icon, nodes, empty state |
| `frontend/src/components/connections/BlueprintView.vue` | New | Hub + zones + SVG lines blueprint layout |
| `frontend/src/pages/ConnectionsPage.vue` | Modified | View toggle, conditional rendering, @add wiring |

## Verification

- **Lint:** All new/modified files pass ESLint (4 pre-existing errors in other files unchanged)
- **Tests:** All 25 frontend tests pass with no regressions
- **Manual:** `make dev-frontend` to verify both features visually
