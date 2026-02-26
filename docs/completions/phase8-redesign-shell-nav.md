# Phase 8 — Frontend Redesign: Shell & Navigation

Phase 2 of the techno-brutalist redesign. Replaces the flat sidebar and top header with a sectioned, branded sidebar and responsive mobile layout.

Spec: [`docs/executing/redesign-fe.md`](../executing/redesign-fe.md) → Phase 2

---

## What Changed

### Deleted

- **`AppHeader.vue`** — removed entirely. The Heimdall branding now lives in the sidebar header. No component imports AppHeader.

### `components/common/AppSidebar.vue` — Full Rewrite

**Brand header** at top:
- Pulsing green dot (`animate-pulse`) — the system's heartbeat
- "HEIMDALL" in uppercase monospace bold with wider tracking
- "Status: Active" sub-label in muted text

**Sectioned navigation** with four groups:
- OVERVIEW → Dashboard
- INFRASTRUCTURE → Connections
- AGENT → Configuration, Chat, Log
- INTELLIGENCE → Reports

Section labels are `10px` uppercase monospace with `tracking-widest` — tiny structural markers that organize without dominating.

**Active route highlighting:**
- Uses `useRoute()` for reactive route matching via computed property
- Active item gets `bg-accent-subtle` background + `text-text-primary`
- Green left-border accent bar (`w-0.5 rounded-full bg-accent`) via absolute positioning
- Inactive items: `text-text-secondary` with hover transitions

**User footer:**
- Email in muted monospace, truncated
- "SIGN OUT" in uppercase monospace, turns `status-critical` red on hover

**New props/emits:**
- `mobile?: boolean` prop — adjusts height behavior (`h-full` vs `min-h-screen`)
- `close` emit — fired on nav click so the mobile overlay can dismiss

**Styling:** `bg-bg-elevated` background, `border-r border-border`, all text in `font-mono`.

### `layouts/DefaultLayout.vue` — Full Rewrite

**Login page bypass:** When `route.name === 'login'`, renders just the `<slot>` with no sidebar or layout wrapper. The login page manages its own full-screen layout.

**Desktop layout (≥1024px / `lg`):**
- Sidebar visible in a `flex-shrink-0` aside
- Main content: `flex-1 min-w-0 p-8`
- Full dark background via `bg-bg-primary`

**Mobile layout (<1024px):**
- Sidebar hidden
- Fixed hamburger button (top-left): bordered, dark background, 3-line SVG icon
- On tap: sidebar slides in as a `Teleport`-ed overlay with:
  - `bg-black/60 backdrop-blur-sm` backdrop (click to dismiss)
  - Sidebar panel slides from left with CSS transition
- Vue `<Transition name="overlay">` handles enter/leave animations (opacity + translateX, 200ms)

### Layout Comparison

**Before:**
```
┌────────────────────────────────────────┐
│ Sidebar │ AppHeader ("Heimdall")       │
│         │──────────────────────────────│
│ • links │ <main> content               │
│ • flat  │                              │
│ • gray  │                              │
└────────────────────────────────────────┘
```

**After:**
```
┌────────────────────────────────────────┐
│ ◉ HEIMDALL     │                       │
│ Status: Active │ <main> content        │
│                │                       │
│ ── OVERVIEW    │ (full dark bg,        │
│ □ Dashboard    │  no top header)       │
│ ── AGENT       │                       │
│ □ Chat         │                       │
│ ...            │                       │
│ ───────────    │                       │
│ user@email     │                       │
│ [Sign Out]     │                       │
└────────────────────────────────────────┘
```

---

## Design Decisions

1. **`Teleport` for mobile overlay.** Renders the overlay at `<body>` level to avoid `overflow: hidden` or `z-index` stacking issues from parent containers.

2. **`min-w-0` on main content.** Prevents flex children with long content (log entries, chat messages) from blowing out the layout width — a common flex gotcha.

3. **Active state via route name, not path.** Using `route.name === 'agent-chat'` rather than path-matching avoids issues with trailing slashes, query params, or nested routes.

4. **Login bypass in DefaultLayout rather than router.** Keeps the router clean (no meta fields or nested route configs) — the layout simply checks the route name and renders raw slot content for login.

---

## Verification

- `vue-tsc -b --noEmit`: passes
- `vite build`: succeeds (823ms)
- No remaining imports of `AppHeader` in the codebase
