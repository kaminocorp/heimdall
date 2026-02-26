So far we've focused primarily on setting up raw functionality and backend infrastructure. Now it's time to focus on the frontend design and user experience.

Aesthetically, I want us to go for a techno-brutalist look and feel. Look at elephantasm.com, trajancloud.com, kaminocorp.com, corpovault.com for inspiration.

Generally, the experience should feel like the user is interacting with an ultra-advanced AI system that's both powerful, exciting and almost a little intimidating. 

Think of military tech interfaces - utilitarian, functional, imposing, but also futuristic and advanced - like something from the Star Wars universe (but non-gimmickey!)

In terms of actual UI layout and components, it should be ultra-intuitive and easy to use, but also feel powerful and advanced.

Where appropriate, let's make the interface feel alive and responsive - like the AI is actively processing and responding to the user's actions.

Let's also make sure the interface is accessible and works well on all devices, so mobile and tablet users aren't left out and we don't have to maintain separate codebases for different platforms.

--

## Proposal: Techno-Brutalist Redesign

### Design Philosophy

The redesign transforms Heimdall from a generic gray utility into something that feels like military-grade AI infrastructure — dark, monospace-forward, structurally honest, and deliberately restrained. The interface should communicate: _"this system is watching everything, and it knows what it's doing."_

Three governing principles:

1. **Structure is ornament.** Borders, grids, and containers are visible and intentional — they _are_ the design, not hidden scaffolding. Think wireframe-as-aesthetic.
2. **Restraint is confidence.** One accent color. One or two typefaces. Generous negative space. The product speaks for itself.
3. **The terminal is our ancestor.** Monospace type, dark surfaces, high-contrast text, flat elements — the lineage of the command line, elevated with modern spacing and interaction.

---

### Design Tokens

#### Color Palette — "Black-Green / Military Camo"

The entire platform carries a green-black atmosphere — not just green accents on neutral dark, but backgrounds with a subtle green undertone throughout. The effect is an immersive military-terminal feel.

| Token | Value | Usage |
|-------|-------|-------|
| `--bg-primary` | `#080c08` | Page background (near-black, green undertone) |
| `--bg-surface` | `rgba(14, 24, 14, 0.5)` | Cards, panels |
| `--bg-surface-hover` | `rgba(14, 24, 14, 0.7)` | Hovered cards |
| `--bg-elevated` | `#0e150e` | Sidebar, modals |
| `--text-primary` | `#e8ede8` | Headings, primary text (warm white, slight green) |
| `--text-secondary` | `#8a9a8a` | Body text, descriptions |
| `--text-muted` | `#5a6b5a` | Timestamps, metadata |
| `--accent` | `#5a9e6a` | Primary actions, active states (muted forest green) |
| `--accent-hover` | `#4a8c5a` | Button hovers (darker) |
| `--accent-bright` | `#7ab889` | High-emphasis elements (slightly brighter sage) |
| `--accent-subtle` | `rgba(90, 158, 106, 0.1)` | Badge fills, tinted backgrounds |
| `--accent-border` | `rgba(90, 158, 106, 0.3)` | Accent-tinted borders |
| `--border` | `rgba(90, 158, 106, 0.08)` | Default borders (green-tinted) |
| `--border-hover` | `rgba(90, 158, 106, 0.18)` | Hovered borders |
| `--status-ok` | `#5a9e6a` | Active, healthy (same as accent) |
| `--status-warn` | `#c4a84a` | Warning severity (olive-gold) |
| `--status-critical` | `#c45a4a` | Error, critical (muted red, not neon) |
| `--status-info` | `#4a8aae` | Info severity (steel blue) |

The accent (`#5a9e6a`) is a muted forest/sage green — military olive-leaning, never neon. Even the status colors are desaturated to stay in the camo register: olive-gold warnings, muted reds, steel blues. The backgrounds use green-tinted near-blacks (`#080c08`, `#0e150e`) so the entire interface feels atmospherically green, not just accessorized with green dots.

#### Typography

| Token | Value |
|-------|-------|
| `--font-mono` | `'JetBrains Mono', 'Fira Code', 'SF Mono', monospace` |
| `--font-sans` | `'Inter', system-ui, -apple-system, sans-serif` |

- **JetBrains Mono** as the display and UI font (nav labels, headings, badges, data). This is the brand identity typeface.
- **Inter** as the readable body font for longer text (descriptions, chat messages, report prose).
- All labels, badges, and navigation items in **uppercase monospace** with `tracking-wider`.

#### Spacing & Layout

- Container max-width: `max-w-7xl` (80rem), centered
- Section padding: `py-8` to `py-12`
- Card padding: `p-5` to `p-6`
- Border radius: `rounded-lg` (0.5rem) for cards, `rounded` (0.25rem) for inputs/badges
- Border width: `1px` everywhere — thin, structural

---

### Layout Architecture

#### Shell (DefaultLayout)

```
┌──────────────────────────────────────────────────────┐
│ SIDEBAR (w-60)          │  MAIN CONTENT AREA         │
│                         │                             │
│  ┌───────────────────┐  │  ┌──────────────────────┐  │
│  │  ◉ HEIMDALL       │  │  │  Page Header         │  │
│  │    Status: ACTIVE  │  │  │  (breadcrumb/title)  │  │
│  └───────────────────┘  │  └──────────────────────┘  │
│                         │                             │
│  ── OVERVIEW ────────   │  ┌──────────────────────┐  │
│  □ Dashboard            │  │                      │  │
│                         │  │  Page Content         │  │
│  ── INFRASTRUCTURE ──   │  │                      │  │
│  □ Connections          │  │                      │  │
│                         │  │                      │  │
│  ── AGENT ───────────   │  │                      │  │
│  □ Configuration        │  └──────────────────────┘  │
│  □ Chat                 │                             │
│  □ Log                  │                             │
│                         │                             │
│  ── INTELLIGENCE ────   │                             │
│  □ Reports              │                             │
│                         │                             │
│  ─────────────────────  │                             │
│  user@example.com       │                             │
│  [Sign Out]             │                             │
└──────────────────────────────────────────────────────┘
```

**Key changes from current layout:**
- Remove the top AppHeader entirely — Heimdall branding moves into the sidebar header
- Sidebar gets section groupings with uppercase monospace labels (OVERVIEW, INFRASTRUCTURE, AGENT, INTELLIGENCE)
- Active nav item gets a green left-border accent bar and subtle accent-tinted background
- Sidebar background: `bg-[#0e150e]` with `border-r border-[rgba(90,158,106,0.08)]`
- Main content area: full dark background, no additional header chrome

**Mobile (< 768px):** Sidebar collapses into a hamburger-triggered slide-over panel with backdrop blur overlay. Main content takes full width.

#### Page Headers

Each page gets a consistent header block:

```
CONNECTIONS                              [+ New Connection]
Manage your infrastructure integrations
────────────────────────────────────────────────────────
```

- Title: uppercase monospace, `text-2xl font-bold tracking-wider`
- Subtitle: Inter, `text-zinc-400 text-sm`
- Divider: `border-b border-white/[0.08]` with `pb-6 mb-8`
- Actions (buttons) right-aligned in the header row

---

### Component Redesign

#### Buttons

**Primary:** `bg-[#5a9e6a] text-[#080c08] font-mono text-sm uppercase tracking-wider px-4 py-2 rounded hover:bg-[#4a8c5a] transition-colors`
- Dark text on muted green for contrast
- Monospace uppercase label

**Secondary (outline):** `border border-[rgba(90,158,106,0.2)] text-[#8a9a8a] font-mono text-sm uppercase tracking-wider px-4 py-2 rounded hover:border-[#5a9e6a]/50 hover:text-[#e8ede8] transition-colors`

**Ghost/danger:** `text-red-400 font-mono text-sm uppercase tracking-wider hover:text-red-300 hover:bg-red-500/10 px-4 py-2 rounded transition-colors`

#### Cards

```css
border border-[rgba(90,158,106,0.08)] bg-[rgba(14,24,14,0.5)] rounded-lg p-5
hover:border-[rgba(90,158,106,0.18)] transition-colors
```

- Semi-transparent dark surface
- Border brightens on hover
- No shadows by default — shadows are reserved for elevated elements (modals, dropdowns)

#### Status Badge

Redesigned with monospace uppercase and ghost-fill styling:

- **Active:** `border border-green-500/30 bg-green-500/10 text-green-400 font-mono text-xs uppercase tracking-wider px-2 py-0.5 rounded-full`
- **Warning:** Same pattern with yellow tokens
- **Error:** Same pattern with red tokens
- **Inactive:** `border-white/[0.1] bg-white/[0.05] text-zinc-500`

#### Form Inputs

```css
bg-[#0e150e]/80 border border-[rgba(90,158,106,0.1)] rounded px-3 py-2
text-[#e8ede8] font-mono text-sm placeholder:text-[#5a6b5a]
focus:border-[#5a9e6a]/50 focus:ring-1 focus:ring-[#5a9e6a]/20
transition-colors
```

- Dark green-tinted input backgrounds that sit flush with the page
- Green accent on focus
- Monospace text inside inputs

#### Select Dropdowns

Same styling as inputs, with a custom chevron indicator.

---

### Page-by-Page Redesign

#### Login Page

Full-screen dark background. Centered card with Heimdall logo/wordmark above.

```
                    ◉ HEIMDALL
            Autonomous Monitoring Agent

         ┌────────────────────────────┐
         │  EMAIL                     │
         │  ┌──────────────────────┐  │
         │  │ operator@corp.io     │  │
         │  └──────────────────────┘  │
         │                            │
         │  PASSWORD                  │
         │  ┌──────────────────────┐  │
         │  │ ••••••••••           │  │
         │  └──────────────────────┘  │
         │                            │
         │  [████ AUTHENTICATE ████]  │
         │                            │
         │  No account? Register →    │
         └────────────────────────────┘
```

- "AUTHENTICATE" button instead of "Sign in" — language reinforces the military-tech tone
- Labels in uppercase monospace above inputs
- Card: `border border-white/[0.08] bg-zinc-900/40 backdrop-blur-sm`
- Subtle scanline or grid pattern on the background (CSS-only, no images)

#### Dashboard

Transform from placeholder into a system status overview:

```
┌─ SYSTEM STATUS ──────────────┐  ┌─ CONNECTIONS ──────────────┐
│                               │  │                            │
│  ◉ AGENT: ACTIVE              │  │  3 active · 1 inactive     │
│  Last cycle: 2m ago           │  │                            │
│  Model: claude-sonnet-4-5     │  │  ▪ prod-db      ● ACTIVE  │
│                               │  │  ▪ staging-logs ● ACTIVE  │
│                               │  │  ▪ github       ● ACTIVE  │
└───────────────────────────────┘  │  ▪ syslog       ○ INACTIVE│
                                   └────────────────────────────┘
┌─ RECENT ACTIVITY ─────────────────────────────────────────────┐
│                                                                │
│  14:32:01  OBSERVATION  Spike in 5xx errors on prod-db         │
│  14:31:58  TOOL_RESULT  search_logs returned 23 entries        │
│  14:31:55  TOOL_CALL    search_logs { severity: "critical" }   │
│  14:30:00  INFO         Routine monitoring cycle completed      │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

- Real data from existing API endpoints (agent config, connections list, recent logs)
- Each card has an uppercase monospace header label with a leading dash (the `┌─ LABEL ──` pattern)
- Activity feed uses monospace timestamps and colored type badges

#### Connections Page

Grid of connection cards with a prominent "New Connection" action.

```
┌─ CONNECTIONS ────────────────────────────── [+ NEW CONNECTION] ┐

┌──────────────────────┐  ┌──────────────────────┐
│  PRODUCTION DB       │  │  STAGING LOGS        │
│  postgres · two_way  │  │  webhook · one_way   │
│                      │  │                      │
│  ● ACTIVE            │  │  ● ACTIVE            │
│  Last seen: 3m ago   │  │  Last seen: 1m ago   │
│                      │  │                      │
│  [Configure] [Delete]│  │  [Configure] [Delete]│
└──────────────────────┘  └──────────────────────┘
```

- 2-column grid on desktop, single-column on mobile
- Card hover: border shifts toward `border-white/[0.15]`
- Connection type and direction as monospace labels
- Delete button is ghost/danger style — only appears on hover

**Connection Form** slides in as a panel or modal with the same dark card treatment.

#### Agent Chat

The centrepiece interaction. Should feel like a direct terminal link to an AI system.

```
┌─ AGENT CHAT ─────────────────── ● CONNECTED ─────────────────┐
│                                                                │
│  ┌─ OPERATOR ────────────────────────────────────────────┐    │
│  │  What happened with the payment service last night?    │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                │
│  ┌─ HEIMDALL ────────────────────────────────────────────┐    │
│  │  I found 47 error entries between 02:00-04:30 UTC...   │    │
│  │                                                        │    │
│  │  ▪ 23 × "connection timeout" on payment-gateway        │    │
│  │  ▪ 18 × "retry exhausted" on stripe-webhook            │    │
│  │  ▪ 6  × "deadlock detected" on orders table            │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                │
│  ┌─ ■■■ PROCESSING ─────────────────────────────────────┐    │
│  │  ░░░░░░░░░                                            │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                │
│  ┌──────────────────────────────────────────── [SEND] ──┐     │
│  │  Type your message...                                 │     │
│  └───────────────────────────────────────────────────────┘     │
└────────────────────────────────────────────────────────────────┘
```

- Messages styled as labeled blocks: `OPERATOR` (user, green accent border) and `HEIMDALL` (agent, default border)
- Thinking state: animated block with pulsing bars or a scanning-line effect
- Connection status indicator in the page header: `● CONNECTED` in green monospace
- Input area: full-width with dark input styling, green SEND button
- **Markdown rendering** in agent messages — structured like Claude Code CLI output (headers, code blocks with syntax highlighting, bullet lists, bold/italic). Uses `marked` + `highlight.js` for rich formatting.

#### Agent Log

The unified feed. Dense, data-rich, terminal-like.

```
┌─ AGENT LOG ──────────────── [Source ▾] [Severity ▾] ─────────┐
│                                                                │
│  TIMESTAMP       TYPE          SOURCE   SUMMARY                │
│  ─────────────────────────────────────────────────────────     │
│  14:32:01.003    OBSERVATION   agent    Spike in 5xx errors    │
│  14:31:58.847    TOOL_RESULT   agent    23 entries returned    │
│  14:31:55.102    TOOL_CALL     agent    search_logs(critical)  │
│  14:31:00.000    CRITICAL      raw      Connection timeout...  │
│  14:30:45.221    WARNING       raw      High memory usage...   │
│  14:30:00.000    INFO          raw      Health check passed    │
│                                                                │
│  ──────────────── Showing 1–25 of 1,847 ── [← Prev] [Next →] │
└────────────────────────────────────────────────────────────────┘
```

- Table-like layout with monospace columns
- Agent entries: green left-border accent
- Raw log entries: default border
- Severity badges: colored ghost-fill (red/yellow/blue/green)
- Expandable rows — click to see `detail` JSON in a monospace code block
- Filter dropdowns in the header row, styled as dark selects

#### Agent Config

Clean read/edit view of agent settings.

```
┌─ AGENT CONFIGURATION ─────────────────────────────────────────┐
│                                                                │
│  MODEL              claude-sonnet-4-5                          │
│  MODE               continuous                                 │
│  SCHEDULE           —                                          │
│  SYSTEM PROMPT      Default                                    │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

- Key-value pairs in a structured grid
- Labels in uppercase monospace, values in regular weight
- Edit mode transforms values into inputs inline

#### Reports

Card grid with severity-coded accent borders.

```
┌─ REPORTS ─────────────────────────────────────────────────────┐
│                                                                │
│  ┌─ CRITICAL ──────────────────────────────────────────┐      │
│  │  Payment service outage — 47 errors over 2.5 hours  │      │
│  │  Feb 25, 2026 · auto_detected · ● in_progress       │      │
│  └─────────────────────────────────────────────────────┘      │
│                                                                │
│  ┌─ WARNING ───────────────────────────────────────────┐      │
│  │  Memory usage trending upward on staging            │      │
│  │  Feb 24, 2026 · auto_detected · ✓ resolved          │      │
│  └─────────────────────────────────────────────────────┘      │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

- Left border color matches severity
- CRITICAL: red left border, WARNING: yellow, INFO: blue

---

### "Alive" Interface Effects

To make the interface feel like an active AI system (per the brief), these CSS-driven effects are applied selectively:

1. **Pulse dot on agent status.** A small `animate-pulse` circle next to "ACTIVE" in sidebar and dashboard — the system's heartbeat.

2. **Scanning line on thinking state.** When the agent is processing in chat, a horizontal gradient line sweeps across the thinking block — like a radar sweep or terminal cursor.

3. **Subtle border glow on active cards.** Cards that represent live/active resources get a barely-visible green box-shadow: `shadow-[0_0_15px_rgba(90,158,106,0.06)]` — just enough to feel "powered on."

4. **Typing cadence for agent messages.** Agent responses appear with a brief character-by-character reveal (CSS animation, not JS re-rendering). This is capped at ~200ms total to avoid feeling slow.

5. **Staggered fade-in on list items.** Log entries and cards fade in with a slight stagger on page load — `animation-delay` based on index, keeping total animation under 400ms.

All animations respect `prefers-reduced-motion: reduce` — they're disabled entirely for users who've opted out.

---

### Responsive Breakpoints

| Breakpoint | Behavior |
|-----------|----------|
| `< 640px` (sm) | Single column, full-width cards, collapsed sidebar (hamburger menu), chat input fixed to bottom |
| `640px–1024px` (md) | Two-column grids where appropriate, sidebar still collapsed |
| `> 1024px` (lg) | Full sidebar visible, multi-column layouts, comfortable spacing |

---

### Dependencies to Add

| Package | Purpose |
|---------|---------|
| `@fontsource/jetbrains-mono` | Self-hosted JetBrains Mono (no external font CDN) |
| `@fontsource/inter` | Self-hosted Inter |
| `marked` | Markdown → HTML rendering for agent chat messages |
| `highlight.js` | Syntax highlighting for code blocks in agent responses |

No additional UI component libraries. Everything is built with Tailwind utilities and minimal custom CSS. No component libraries (Headless UI, Radix, etc.) — the component set is small enough to own entirely.

---

### Implementation Plan

The redesign is applied incrementally, foundation-first:

**Phase 1 — Foundation**
Install fonts, set up Tailwind theme (colors, font families) in `main.css`, establish CSS custom properties. Create the dark base and verify it renders.

**Phase 2 — Shell & Navigation**
Redesign `DefaultLayout`, `AppSidebar` (with section groups, active states, mobile hamburger), remove `AppHeader`. This immediately transforms the overall feel.

**Phase 3 — Shared Components**
Restyle buttons, cards, badges, inputs, selects, loading spinner. These propagate across all pages automatically.

**Phase 4 — Pages (in priority order)**
1. Login — first impression, standalone page
2. Agent Chat — centrepiece interaction
3. Agent Log — data-dense showcase
4. Dashboard — system overview (flesh out from placeholder)
5. Connections — CRUD with new card style
6. Agent Config — simple key-value view
7. Reports — card grid with severity accents
8. 404 — quick restyle

**Phase 5 — Polish & Animation**
Add the "alive" effects: pulse dots, scanning line, staggered fades. Responsive testing across breakpoints. `prefers-reduced-motion` support.

---

### Resolved Decisions

1. **Accent color** — Military camo green (`#5a9e6a` muted forest sage). The entire platform gets green-tinted backgrounds for an immersive black-green atmosphere. All status colors desaturated to camo register.

2. **Chat rendering** — Full markdown support (marked + highlight.js) so agent responses are as well-structured as Claude Code CLI output. Code blocks, lists, headers, bold/italic all rendered.

3. **Dashboard** — Wire up real data from existing endpoints (agent config, connections, recent logs). No new backend work needed.

---

### Logo — Heimdall Emblem

The logo should feel like a Star Wars Imperial-era military insignia: geometric, angular, symmetrical, imposing, monochrome. Combined with Heimdall's "all-seeing" surveillance identity.

Three concepts to choose from:

#### Concept A: "The Watchtower Sigil"

An eye abstracted into Imperial spoke geometry — the iris becomes a six-spoked cog, the overall silhouette is both an eye and a fortress.

- A **horizontal pointed ellipse** (almond/vesica piscis) forms the outer eye boundary
- Inside, a **circle** (the iris) with **six trapezoidal spokes** radiating from a small **central disc** (the pupil) — directly echoing the Imperial crest
- The horizontal spokes at 0/180 degrees extend to the ellipse tips, creating an unbroken "scanning line" axis
- Reads as: _"a fortress that watches"_

#### Concept B: "The Bifrost Hexagram"

Inspired by the First Order's hexagonal containment. A hexagonal frame contains a radial starburst with a lens/aperture at center.

- A **regular hexagon** as the outermost boundary (thick stroke)
- **12 pointed rays** — 6 primary (to vertices) + 6 secondary shorter rays (to edge midpoints), creating a dense radial hierarchy
- Center: **double-ring aperture motif** with a void core (the pupil/lens)
- Small **notched tips** where rays meet the hex boundary — gear-tooth effect
- Reads as: _"omnidirectional aggressive surveillance"_

#### Concept C: "The Gjallarhorn Eye" (recommended)

The most minimal and unique. A single geometric form that is simultaneously a stylized eye AND the bell of Gjallarhorn (Heimdall's horn). The most semantically layered: eye + horn + alert = monitoring agent that watches and warns.

- A **circle** (horn bell / iris) as the primary element, medium stroke
- Inside, a **smaller concentric filled circle** (pupil / horn throat)
- Above: a **heavy angular chevron** (two lines meeting at a point above center) — the severe brow. Arms extend down past the midline, framing the top half
- Below: a **mirrored thinner chevron** — the lower eyelid, lighter weight for hierarchy
- **Four diagonal lines** at 45/135/225/315 degrees radiating outward — sound waves / sight lines
- Fits within a **square bounding box** (not circular — differentiates from every Star Wars emblem while retaining geometric severity)
- Reads as: _"I see everything AND I will sound the alarm"_

| | Watchtower Sigil | Bifrost Hexagram | Gjallarhorn Eye |
|--|--|--|--|
| Shape | Pointed ellipse | Hexagon | Circle + chevrons |
| Spokes | 6 (Imperial classic) | 12 (6+6) | 4 diagonal rays |
| Eye reference | Explicit (ellipse) | Abstract (aperture) | Explicit (chevron brow) |
| Complexity | Medium | Highest | Lowest |
| Uniqueness | Moderate | High | **Highest** |
| Favicon scaling | Good | Fair | **Best** |

**Chosen: Concept B — The Bifrost Hexagram.** Dense, imposing, mechanical. The hexagonal containment and 12-ray starburst carry the most Imperial authority.