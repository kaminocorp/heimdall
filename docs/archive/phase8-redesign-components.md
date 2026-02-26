# Phase 8 — Frontend Redesign: Shared Components

Phase 3 of the techno-brutalist redesign. Restyles all 14 Vue components to use the design token system from Phase 1. Every light-mode gray reference replaced with dark-theme token classes.

Spec: [`docs/executing/redesign-fe.md`](../executing/redesign-fe.md) → Phase 3

---

## Components Restyled

### Common (2)

**`LoadingSpinner.vue`**
- Spinner border: `border-border` track, `border-t-accent` leading edge
- Replaces `border-gray-300 / border-t-gray-900`

**`StatusBadge.vue`**
- Redesigned as ghost-fill pill with colored dot indicator
- `rounded-full` with `font-mono text-[10px] uppercase tracking-wider`
- States: `active` (green), `inactive` (muted white), `error` (red), `warning` (yellow)
- Each state uses matching `border-*/30 bg-*/10 text-*` pattern from status tokens

### Connections (3)

**`ConnectionCard.vue`**
- Card: `bg-bg-surface border-border rounded-lg` with hover border transition
- Name: uppercase monospace, type/direction as muted monospace
- Delete button: hidden by default, fades in on card hover (`opacity-0 group-hover:opacity-100`), turns red on hover

**`ConnectionForm.vue`**
- Form container: `bg-bg-surface border-border`
- Labels: uppercase monospace, `text-text-secondary`
- Inputs/selects: `bg-bg-elevated/80 border-border`, green focus ring
- Primary button: `bg-accent text-bg-primary` (dark text on green)
- Cancel button: outline style with `border-accent-border/50`

**`ConnectionList.vue`**
- Empty state: `text-text-muted font-mono`
- Layout: 2-column grid on `md+` breakpoint, single column on mobile

### Log (3)

**`LogFilters.vue`**
- Select dropdowns: dark monospace inputs matching form input style
- Shared class string via `selectClasses` const for DRY
- Added `flex-wrap` for mobile responsive wrapping

**`LogEntry.vue`**
- Agent entries: left green accent border (`border-l-2 border-l-accent`) + `bg-accent-subtle`
- Agent badge: bordered pill with `text-accent`
- Severity badges: ghost-fill style matching StatusBadge pattern (critical=red, warning=yellow, info=blue)
- Entry type label: `text-accent-bright`
- Summary text: `text-text-secondary`

**`LogFeed.vue`**
- Pagination buttons: bordered monospace uppercase with disabled opacity
- Showing count: `font-mono text-xs text-text-muted`
- Empty state: muted monospace

### Agent Chat (3)

**`ChatMessage.vue`**
- Removed bubble-chat layout (no more left/right alignment)
- Added role labels: "OPERATOR" (green accent) / "HEIMDALL" (muted) in `10px` uppercase monospace
- User messages: `bg-accent-subtle border-accent-border/50`
- Agent messages: `bg-bg-surface border-border`
- Max width: `max-w-2xl w-full` for readability

**`ChatWindow.vue`**
- Outer container: `border border-border rounded-lg bg-bg-surface` with overflow hidden
- Thinking indicator: "Processing" label with CSS scanning line animation (green gradient sweeps left-to-right, 1.5s loop)
- Replaces bouncing dots with military-tech scanning effect

**`ChatInput.vue`**
- Input: dark elevated background, monospace, green focus ring
- Send button: `bg-accent text-bg-primary uppercase tracking-wider`
- Placeholder text changes based on disabled state
- Border-top uses `border-border` token

### Reports (3)

**`ReportCard.vue`**
- Left border colored by severity (`border-l-2`): critical=red, warning=yellow, info=blue
- Severity badge: ghost-fill pill matching StatusBadge style
- Status text: muted monospace

**`ReportDetail.vue`**
- Title: monospace bold uppercase
- Key-value grid: labels in muted uppercase monospace, values in `text-text-secondary`

**`ReportList.vue`**
- No visual changes (wrapper component), spacing retained at `space-y-3`

---

## Design Patterns Applied

1. **Ghost-fill badges.** All status/severity indicators use `border-color/30 bg-color/10 text-color` — visible but not overwhelming, consistent across all contexts.

2. **Group hover reveals.** Destructive actions (delete buttons) hidden by default, revealed on card hover via `group` / `group-hover:opacity-100`. Reduces visual noise.

3. **Shared class strings.** Repeated styling (e.g., filter selects, pagination buttons) extracted into `const` variables in `<script setup>` rather than duplicating long class lists in template.

4. **Scanning line animation.** The agent thinking state uses a CSS-only horizontal gradient sweep (`translateX` keyframes) — performant, no JS re-rendering, respects `prefers-reduced-motion` via Phase 1's blanket disable.

---

## Verification

- `vue-tsc -b --noEmit`: passes
- `vite build`: succeeds (851ms)
- All 14 components now use exclusively design token classes — zero references to default gray palette remain
