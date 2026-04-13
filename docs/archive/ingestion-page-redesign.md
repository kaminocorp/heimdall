# Ingestion Page Redesign — Implementation Plan

**Status:** Not started  
**Goal:** Replace the current Blueprint/List view on the Connections page with a visual, 3D-powered layout: connection bubbles with native brand logos feeding into an ephemeral Heimdall Agent particle abstraction, with click-to-inspect detail modals.

---

## Why this exists

The current Connections page has two view modes — a "Blueprint" (3-column grid with SVG bezier lines) and a flat card list. Neither communicates the mental model that matters: **data flows from your infrastructure into the Heimdall agent**. The Blueprint view groups connections by type in a left/center/right layout, but the center "hub" is just a static circle with an eye icon. There's no visual representation of the agent as a living, processing entity — which is exactly what it is.

This redesign makes the agent tangible. Connections are clickable bubbles with recognisable brand logos (the Supabase icon, GitHub's octocat, the PostgreSQL elephant) arranged above a particle nebula that represents the Heimdall agent. The visual hierarchy is **sources above → agent below**, with animated flow lines connecting them. Clicking a connection opens a detail modal with full info, edit, test, and delete capabilities.

The 3D particle abstraction is adapted from the Elephantasm nebula reference (`docs/plans/ref-elephantasm-animation.md`) — same architectural pattern (multi-layer `THREE.Points` with GLSL simplex noise, additive blending) but retuned for Heimdall's phosphor-green retro-futurism palette and a smaller, subtler presence.

---

## Visual Architecture

```
┌─────────────────────────────────────────────────────────┐
│  Page header: "CONNECTIONS" + [+ NEW CONNECTION] button  │
├─────────────────────────────────────────────────────────┤
│                                                         │
│       ┌──────┐   ┌──────┐   ┌──────┐   ┌──────┐       │
│       │ SB   │   │ WH   │   │ GH   │   │ PG   │       │  Connection
│       │ logo │   │ logo │   │ logo │   │ logo │       │  Bubbles
│       └──┬───┘   └──┬───┘   └──┬───┘   └──┬───┘       │
│          │          │          │          │             │
│          ╰──────────┼──────────┼──────────╯             │  Flow Lines
│                     │          │                        │  (animated)
│                     ▼          ▼                        │
│            ┌────────────────────────┐                   │
│            │                        │                   │
│            │   ░░ PARTICLE NEBULA ░░│                   │  Heimdall Agent
│            │   ░░ (Three.js canvas) │                   │  3D Abstraction
│            │   ░░                   │                   │
│            │        HEIMDALL        │                   │
│            └────────────────────────┘                   │
│                                                         │
│  Empty state: "No connections yet — add your first      │
│  integration to start monitoring."                      │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Connection Bubble Anatomy

```
┌────────────────────┐
│  ┌────┐            │
│  │ 🟢 │  My Prod   │   ← Native logo (SVG), connection name
│  │logo│  Supabase   │   ← Type label, status dot
│  └────┘            │
└────────────────────┘
```

Each bubble is a Vue component rendering:
- The **native brand SVG** for the connection type (Supabase, GitHub, PostgreSQL, etc.)
- Connection **name** (truncated if long)
- **Status indicator** (green dot active, yellow warn, red error)
- Subtle glow effect when active, matching the existing `glow-active` pattern
- Click → opens detail modal

### Click-to-Inspect Detail Modal

```
┌─────────────────────────────────────────┐
│  ╳                                      │
│                                         │
│  [Logo]  My Production Database         │
│          PostgreSQL · two-way · active   │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  Host        db.example.com             │
│  Port        5432                       │
│  Database    myapp_production           │
│  User        heimdall_reader            │
│  SSL Mode    require                    │
│  Created     2026-04-10                 │
│  Last tested 2 hours ago                │
│                                         │
│  ─────────────────────────────────────  │
│                                         │
│  [Ping]  [Edit]  [Delete]               │
│                                         │
└─────────────────────────────────────────┘
```

---

## Technical Decisions

### Three.js Integration: Raw Imperative (no TresJS)

The Elephantasm reference uses React Three Fiber. We're in Vue 3. Options:

| Option | Pros | Cons |
|--------|------|------|
| **TresJS** (Vue R3F equivalent) | Declarative, Vue-native | Extra dependency, less control, smaller community |
| **Raw Three.js** in a Vue component | Zero dependencies beyond `three`, full control, matches how the Elephantasm shaders work | Imperative setup, more boilerplate |

**Decision: Raw Three.js.** The animation is self-contained (no reactive props drive the shader). A single Vue component owns a `<canvas>`, creates the Three.js scene in `onMounted`, runs `requestAnimationFrame`, and tears down in `onBeforeUnmount`. This mirrors the Elephantasm pattern most directly and avoids pulling in TresJS for one component.

**New dependency:** `three` (+ `@types/three` for dev). No other packages needed.

### Particle Aesthetic: Heimdall Palette

The Elephantasm nebula uses pearlescent iridescence (indigo/amber/rose/teal). Heimdall's palette is phosphor-green retro-futurism. The adaptation:

| Elephantasm | Heimdall |
|-------------|----------|
| Pearl base `(0.93, 0.91, 0.96)` | Pale feldgrau `(0.82, 0.90, 0.84)` |
| Indigo mood `(0.45, 0.35, 0.75)` | Deep green `(0.29, 0.48, 0.36)` — `#4a7a5c` |
| Amber mood `(0.85, 0.65, 0.35)` | Phosphor bright `(0.42, 0.74, 0.48)` — `#6aad7a` |
| Rose mood `(0.80, 0.45, 0.55)` | Teal info accent `(0.29, 0.57, 0.77)` — `#4a92c4` |
| Teal mood `(0.35, 0.70, 0.72)` | Warm amber status `(0.83, 0.66, 0.20)` — `#d4a832` muted |

The result: a breathing green-tinted nebula that feels like a powered-on monitoring system, consistent with the existing UI.

### Reduced Particle Budget

The Elephantasm runs fullscreen at 17,300 particles. The Ingestion page nebula is a smaller element (~400px tall) and shares the viewport with interactive UI. Budget:

| Layer | Elephantasm | Heimdall |
|-------|-------------|----------|
| Primary Cloud | 14,000 | 6,000 |
| Wisp Tendrils | 2,500 | 1,000 |
| Core Motes | 800 | 400 |
| **Total** | **17,300** | **7,400** |

This keeps the GPU comfortable on integrated graphics (laptops) while maintaining visual density.

### Flow Lines: CSS/SVG with Particle Hint

The "feeding" lines from bubbles to the nebula will be **animated SVG dashed paths** (similar to the current Blueprint view's approach) with a subtle particle-trail effect via CSS. Not full Three.js particle streams — that would double the rendering cost for a secondary visual. The SVG lines already have the glow filter and dash animation pattern in the existing `BlueprintView.vue`; we evolve that rather than rewrite it.

### Mobile Strategy

On viewports below `md` (768px), the 3D canvas is replaced with a **simplified static visual** — the Heimdall eye icon with a CSS glow/pulse animation (no WebGL). Connection bubbles stack vertically. The layout degrades gracefully from the spatial metaphor to a clean list, similar to how the current Blueprint view already hides SVG lines on mobile.

### Brand Logo SVGs

Each connection type gets an inline SVG component. These are small, single-colour marks — not full wordmarks.

| Type | Logo Source | Treatment |
|------|------------|-----------|
| Supabase | Official mark (the stylised "S" bolt) | Mono-colour, accent-tinted |
| PostgreSQL | Elephant head silhouette | Mono-colour |
| GitHub | Octocat silhouette | Mono-colour |
| Webhook | Custom icon (arrow-into-bracket) | Match existing style |
| Syslog | Custom icon (terminal/log) | Match existing style |
| OpenTelemetry | Official mark (the telescope) | Mono-colour |
| Datadog | Dog silhouette (coming soon) | Greyed out |
| MySQL | Dolphin silhouette (coming soon) | Greyed out |

All logos render at 32×32 inside the bubble, tinted to `--accent` when active, `--text-muted` when inactive.

---

## Implementation Phases

### Phase 1 — Three.js Agent Nebula Component

**Goal:** A standalone Vue component that renders the Heimdall particle nebula in a `<canvas>`.

**Files:**
| File | Kind | Description |
|------|------|-------------|
| `frontend/package.json` | Edit | Add `three` dependency, `@types/three` devDep |
| `frontend/src/components/connections/AgentNebula.vue` | **New** | The 3D particle component |

**Implementation detail:**

The component structure:

```vue
<template>
  <div ref="containerRef" class="agent-nebula">
    <canvas ref="canvasRef" />
    <p class="nebula-label">HEIMDALL</p>
  </div>
</template>
```

Internally:
1. `onMounted` → create `WebGLRenderer` (alpha, antialias, powerPreference), `PerspectiveCamera`, `Scene`
2. Create three `THREE.Points` layers (PrimaryCloud, WispTendrils, CoreMotes) with custom `ShaderMaterial`
3. Inline the Ashima simplex noise GLSL as a shared string constant
4. Port the vertex/fragment shaders from the Elephantasm reference, retuning:
   - Colour palette → Heimdall feldgrau/phosphor
   - Particle counts → 6000/1000/400
   - Amplitude ranges → slightly tighter (smaller visual footprint)
   - Breathing speed → slightly faster (feels more "active processing")
5. `requestAnimationFrame` loop updates uniforms (`uTime`, `uLowAmp`, `uMidAmp`, `uCoherence`) and layer rotations
6. `ResizeObserver` on the container handles canvas resize
7. `onBeforeUnmount` → dispose geometries, materials, renderer; cancel animation frame

**No OrbitControls** — this isn't interactive 3D. The nebula auto-rotates and breathes. User interaction is with the Vue layer (bubbles, modals), not the 3D scene.

**Acceptance criteria:**
- [ ] Canvas renders the particle nebula in Heimdall's green palette
- [ ] Animation runs at 60fps on a 2020 MacBook Air (integrated GPU)
- [ ] Component mounts/unmounts cleanly with no WebGL context leaks
- [ ] `vue-tsc --noEmit` and `vite build` clean
- [ ] Canvas respects container size and resizes correctly

---

### Phase 2 — Brand Logo SVG Components

**Goal:** Inline SVG components for each connection type's native logo.

**Files:**
| File | Kind | Description |
|------|------|-------------|
| `frontend/src/components/icons/ConnectorLogos.vue` | **New** | Single-file component exporting all logos as named slots/props |

**Alternative approach:** One component with a `type` prop that switches between inline SVGs. Keeps imports clean — one component, not eight separate files.

```vue
<ConnectorLogo type="supabase" :size="32" class="text-accent" />
```

**Logo sourcing:**
- Supabase, GitHub, PostgreSQL, OpenTelemetry, Datadog, MySQL → official brand SVG marks, simplified to single-path mono-colour
- Webhook, Syslog → custom icons designed to match the brand logo style
- All logos must be single-colour and accept `currentColor` for CSS colour control

**Acceptance criteria:**
- [ ] Each supported connection type renders a recognisable logo
- [ ] Logos accept `size` and inherit text colour via `currentColor`
- [ ] Coming-soon types (Datadog, MySQL) render greyed out
- [ ] `vue-tsc --noEmit` clean

---

### Phase 3 — Connection Bubble Component

**Goal:** The clickable bubble that represents a single connection.

**Files:**
| File | Kind | Description |
|------|------|-------------|
| `frontend/src/components/connections/ConnectionBubble.vue` | **New** | Bubble component |

**Props:** `connection: Connection`, `testing: boolean`

**Emits:** `click` (opens detail modal)

**Visual:**
- Rounded container with `bg-bg-surface` and `border-border`
- Brand logo (via `ConnectorLogo`) prominently displayed
- Connection name below/beside the logo
- Status dot (green/yellow/red) with glow when active
- Hover: border brightens to `border-hover`, subtle scale transform
- Testing state: pulsing accent border animation (matches existing pattern)
- Transition: fade-in with staggered delay on mount

**Acceptance criteria:**
- [ ] Renders logo, name, status for each connection type
- [ ] Click emits event (modal handled by parent)
- [ ] Hover and active states feel consistent with Heimdall's UI
- [ ] Staggered mount animation

---

### Phase 4 — Connection Detail Modal

**Goal:** Full-info modal that appears on bubble click, with edit/test/delete actions.

**Files:**
| File | Kind | Description |
|------|------|-------------|
| `frontend/src/components/connections/ConnectionDetailModal.vue` | **New** | Detail modal |

**Props:** `connection: Connection`

**Emits:** `close`, `edit`, `test`, `delete`, `manage-repos`

**Sections:**
1. **Header** — Logo + name + type/direction/status badges
2. **Configuration details** — Type-specific fields displayed as a key-value list:
   - PostgreSQL: host, port, database, user, ssl_mode
   - Supabase: project_ref, polling tables, poll interval
   - Webhook: endpoint URL, bearer token (masked)
   - Syslog: port, protocol
   - GitHub: installed repos count
   - OTLP: endpoint URL
3. **Metadata** — Created date, last tested, connection ID
4. **Action bar** — [Ping] [Edit] [Delete] buttons
   - Delete uses the same two-step confirm pattern from `ConnectionCard.vue`
   - GitHub connections also show [Manage Repos]

**Overlay:** Dark backdrop, centered modal, close on Escape or backdrop click. Same modal pattern used by `ConnectionTestModal.vue` and `ConnectionWizard.vue`.

**Acceptance criteria:**
- [ ] Shows all relevant connection info by type
- [ ] Edit button triggers parent's edit flow (opens ConnectionForm)
- [ ] Ping button triggers test (opens ConnectionTestModal)
- [ ] Delete has two-step confirmation
- [ ] Escape and backdrop click close the modal
- [ ] Sensitive fields (passwords, tokens) are masked with reveal toggle

---

### Phase 5 — Flow Lines (SVG Animated Connectors)

**Goal:** Animated lines from each connection bubble down to the nebula canvas.

**Files:**
| File | Kind | Description |
|------|------|-------------|
| `frontend/src/components/connections/FlowLines.vue` | **New** | SVG overlay computing and rendering flow paths |

**Implementation:**
- Absolutely positioned SVG layer between the bubble row and the nebula
- Each bubble has a ref; the nebula container has a ref
- `ResizeObserver` + `onMounted` computes quadratic bezier paths from each bubble's bottom-center to the nebula's top-center
- Lines are dashed (`stroke-dasharray: 6 4`), animated with `stroke-dashoffset` keyframes to create a "flowing downward" effect
- Glow filter matching existing `BlueprintView` pattern
- Lines stagger their animation start by index
- Hidden on mobile (`hidden md:block`)

**Data flow visual:** The dash animation direction is **top-to-bottom** — data visually flows from connections into the agent. This is the opposite of typical "draw-in" animations; the dashes move continuously to suggest ongoing data flow, not a one-time connection.

**Acceptance criteria:**
- [ ] Lines connect each bubble to the nebula
- [ ] Animation suggests continuous downward data flow
- [ ] Lines recompute on resize and when connections change
- [ ] Hidden on mobile
- [ ] Glow effect matches existing UI

---

### Phase 6 — Compose the New Ingestion View

**Goal:** Wire all new components into the ConnectionsPage, replacing BlueprintView and ConnectionList.

**Files:**
| File | Kind | Description |
|------|------|-------------|
| `frontend/src/pages/ConnectionsPage.vue` | Edit | Replace blueprint/list toggle with new unified view |
| `frontend/src/components/connections/IngestionView.vue` | **New** | Orchestrator component for the new layout |

**Layout structure of `IngestionView.vue`:**

```
<div class="ingestion-view">
  <!-- Bubble row: horizontally centered, wrapping -->
  <div class="bubble-row">
    <ConnectionBubble v-for="conn in connections" :connection="conn" @click="select(conn)" />
  </div>

  <!-- Flow lines SVG overlay -->
  <FlowLines :bubbles="bubbleRefs" :target="nebulaRef" />

  <!-- Agent nebula -->
  <AgentNebula ref="nebulaRef" />

  <!-- Empty state (when no connections) -->
  <div v-if="!connections.length" class="empty-state">
    ...
  </div>
</div>
```

**Changes to `ConnectionsPage.vue`:**
- Remove `ViewToggle` import and `viewMode` ref
- Remove conditional rendering of `BlueprintView` vs `ConnectionList`
- Replace with single `<IngestionView>` component
- Add `ConnectionDetailModal` with `selectedConnection` ref
- Keep all existing modal/wizard/form/error logic untouched

**Empty state:** When there are no connections, the nebula still renders (dimmer, smaller amplitude) with a centred message: "No connections yet" and the "+ New Connection" CTA. The breathing nebula without any feeding connections feels like a dormant system waiting to be activated.

**Mobile layout:**
- Bubbles stack in a 2-column grid
- Flow lines hidden
- Nebula replaced with the static Heimdall eye + CSS pulse
- Detail modal is full-screen on mobile

**Acceptance criteria:**
- [ ] New view renders correctly with 0, 1, and 5+ connections
- [ ] Clicking a bubble opens the detail modal
- [ ] All existing actions (edit, test, delete, manage repos) still work
- [ ] Wizard still opens from "+ New Connection"
- [ ] GitHub OAuth callback flow still works
- [ ] `vue-tsc --noEmit` and `vite build` clean
- [ ] Mobile layout degrades gracefully

---

### Phase 7 — Polish & Performance

**Goal:** Final tuning pass.

**Tasks:**

1. **Nebula tuning** — Adjust particle counts, amplitude ranges, colour intensities, and animation speeds based on how it looks in-situ with actual connections above it. The reference values in Phase 1 are starting points.

2. **Performance profiling** — Run Chrome DevTools Performance tab on the page. Verify:
   - 60fps sustained with nebula + bubbles
   - No significant layout thrashing from ResizeObserver
   - WebGL context doesn't leak on route navigation (mount/unmount cycle)
   - Total JS bundle impact of `three` is acceptable (tree-shaking: we only import `WebGLRenderer`, `Scene`, `PerspectiveCamera`, `Points`, `BufferGeometry`, `Float32BufferAttribute`, `ShaderMaterial`, `AdditiveBlending` — should be well under 200KB gzipped)

3. **`prefers-reduced-motion`** — When the user has reduced motion enabled:
   - Nebula freezes (no animation frame updates, static render)
   - Flow line dash animation stops
   - Bubble mount transitions are instant

4. **Empty state animation** — Nebula in dormant mode: lower particle alpha, tighter coherence range, slower breathing. Visually communicates "idle, waiting for data."

5. **Cleanup** — Remove now-unused components if no other page references them:
   - `BlueprintView.vue`, `BlueprintZone.vue`, `BlueprintNode.vue`
   - `ViewToggle.vue`
   - `ConnectionList.vue`, `ConnectionCard.vue`
   
   Verify with grep that nothing else imports them before deleting.

**Acceptance criteria:**
- [ ] Sustained 60fps on mid-range hardware
- [ ] Reduced motion respected
- [ ] No dead code left behind
- [ ] Bundle size delta documented

---

## Files Summary

### New files (8)
| File | Phase | Purpose |
|------|-------|---------|
| `components/connections/AgentNebula.vue` | 1 | Three.js particle nebula |
| `components/icons/ConnectorLogo.vue` | 2 | Brand logo SVGs by type |
| `components/connections/ConnectionBubble.vue` | 3 | Clickable connection bubble |
| `components/connections/ConnectionDetailModal.vue` | 4 | Full-info detail modal |
| `components/connections/FlowLines.vue` | 5 | SVG animated connector lines |
| `components/connections/IngestionView.vue` | 6 | New page layout orchestrator |

### Modified files (2)
| File | Phase | Change |
|------|-------|--------|
| `frontend/package.json` | 1 | Add `three` + `@types/three` |
| `frontend/src/pages/ConnectionsPage.vue` | 6 | Replace view toggle with IngestionView |

### Removed files (5, Phase 7)
| File | Reason |
|------|--------|
| `components/connections/BlueprintView.vue` | Replaced by IngestionView |
| `components/connections/BlueprintZone.vue` | Only used by BlueprintView |
| `components/connections/BlueprintNode.vue` | Only used by BlueprintZone |
| `components/connections/ViewToggle.vue` | No longer two view modes |
| `components/connections/ConnectionList.vue` | Replaced by bubble layout |
| `components/connections/ConnectionCard.vue` | Replaced by bubble + detail modal |

### Untouched files
| File | Why |
|------|-----|
| `ConnectionForm.vue` | Still used for edit flow |
| `ConnectionTestModal.vue` | Still used for ping/test |
| `ConnectionWizard.vue` + all wizard steps | Still used for creation flow |
| `GitHubRepoSelector.vue` | Still used post-install |
| `stores/connections.ts` | No API changes |
| `api/connections.ts` | No API changes |
| `types/connection.ts` | No type changes |

---

## Execution Order

Phases 1–2 are independent and can be done in parallel. Phase 3 depends on Phase 2 (needs `ConnectorLogo`). Phase 4 is independent of 1–3. Phase 5 depends on 3 (needs bubble refs). Phase 6 depends on all of 1–5. Phase 7 is final polish after 6.

```
Phase 1 (Nebula) ──────────┐
                            ├──→ Phase 5 (Flow Lines) ──→ Phase 6 (Compose) ──→ Phase 7 (Polish)
Phase 2 (Logos) → Phase 3 (Bubbles) ──┘                        ↑
                                                                │
Phase 4 (Detail Modal) ────────────────────────────────────────┘
```

---

## Risk Notes

1. **`three` bundle size** — Three.js is large (~600KB raw). With tree-shaking (Vite/Rollup), importing only the specific classes we need should reduce this significantly. If the delta exceeds 150KB gzipped, consider dynamic importing the nebula component so it doesn't block initial page load.

2. **WebGL context limits** — Browsers allow ~8–16 WebGL contexts. If users navigate quickly between routes, ensure the context is properly disposed. The `onBeforeUnmount` cleanup in Phase 1 handles this.

3. **Mobile GPU** — The "no WebGL on mobile" decision (Phase 6) sidesteps this entirely. CSS pulse fallback is zero-cost.

4. **Logo licensing** — Supabase, GitHub, PostgreSQL, and OpenTelemetry logos are permissively licensed for this kind of integration display. Verify each before using. Simplify to mono-colour single-path SVGs to avoid trademark concerns with full-colour reproductions.
