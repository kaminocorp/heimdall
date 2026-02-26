# Phase 8 — Frontend Redesign: Pages

Phase 4 of the techno-brutalist redesign. All 8 pages restyled to the dark green-black military aesthetic, with consistent page headers and full design token usage. Dashboard elevated from placeholder to real system overview.

Spec: [`docs/executing/redesign-fe.md`](../executing/redesign-fe.md) → Phase 4

---

## Pages Restyled (in priority order)

### 1. LoginPage

- Full-screen dark background with CSS grid pattern overlay (`opacity-[0.03]`, green-tinted lines at 40px intervals)
- Centered brand block: pulsing green dot + "HEIMDALL" uppercase monospace + "Autonomous Monitoring Agent" subtitle
- Login card: `border-border bg-bg-surface backdrop-blur-sm`
- Inputs: dark elevated background, monospace, green focus ring
- Submit button: "AUTHENTICATE" / "REGISTER" in uppercase tracking-widest
- Error banner: ghost-fill red (matching `status-critical` token)
- Toggle link: accent green color

### 2. AgentChatPage

- Page header: uppercase monospace title + subtitle + connection status indicator
- Status dot: animated pulse when connecting, solid green when open, red when closed
- Status label: human-readable ("Connected" / "Connecting" / "Disconnected")
- Error banner: ghost-fill `status-critical` style
- Chat window fills remaining flex space

### 3. AgentLogPage

- Page header: consistent title/subtitle/divider pattern
- Error banner: ghost-fill `status-critical` style
- All log display delegated to restyled `LogFeed` component (Phase 3)

### 4. DashboardPage — **Major Enhancement**

Transformed from 2-line placeholder into a real system status overview with three data cards:

**System Status card:**
- Agent status: pulsing green dot + "ACTIVE" label
- Model name from `agentStore.config`
- Mode from `agentStore.config`

**Connections card:**
- Summary: "N active · M inactive" count
- Lists all connections with name and `StatusBadge`

**Recent Activity card:**
- Shows 8 most recent log entries
- Each row: timestamp (tabular-nums monospace) + colored type badge + truncated summary
- Agent entries show `source_type` (tool_call, etc.) in accent green
- Raw entries show severity with matching status color

All three cards use `border-border rounded-lg bg-bg-surface` with uppercase monospace section headers.

### 5. ConnectionsPage

- Page header with "New Connection" primary button (accent green)
- Error banner: ghost-fill style
- Connection list displays in 2-column grid (Phase 3 component)

### 6. AgentConfigPage

- Page header: consistent pattern
- Config displayed as key-value pairs in a bordered card
- Labels: uppercase monospace muted
- Values: monospace primary text
- Rows separated by `border-border` dividers
- Null schedule shows em-dash, null system prompt shows "Default"

### 7. ReportsPage

- Page header: consistent pattern
- Empty state: muted monospace message
- Reports rendered via restyled `ReportList` component (Phase 3)

### 8. NotFoundPage

- Large "404" in `text-6xl font-bold text-text-muted`
- "Target not found" in uppercase monospace (military-tech copy)
- "Return to Dashboard" as outline-style button

---

## Consistent Page Header Pattern

Every authenticated page uses the same structure:

```
TITLE                                   [Optional action button]
Subtitle description text
────────────────────────────────────────────────────────────────
```

- Title: `font-mono text-2xl font-bold uppercase tracking-wider text-text-primary`
- Subtitle: `font-sans text-sm text-text-secondary mt-1`
- Divider: `pb-6 mb-8 border-b border-border`

---

## Palette Cleanup

Grep for all old Tailwind default colors (`text-gray-*`, `bg-red-*`, `text-blue-*`, etc.) returns **zero matches** across the entire `frontend/src/` directory. The codebase is fully migrated to design tokens.

---

## Verification

- `vue-tsc -b --noEmit`: passes
- `vite build`: succeeds (878ms)
- Zero remaining references to old gray/red/blue/yellow/green/purple Tailwind palette classes
