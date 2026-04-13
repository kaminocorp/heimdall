# UI Refresh — Technical Retro-Futurism

## Status: Proposal

## Objective

Evolve Heimdall's visual identity from "muted techno-brutalist" to **technical retro-futurism** — higher contrast, richer colour signatures, and atmospheric depth that evokes mission-control interfaces. The goal is not a redesign but an **intensity upgrade**: the same dark-first, monospace, miltech bones with more visual authority.

### What stays unchanged
- Logo (Bifrost Starburst), wordmark, favicon
- Typefaces (JetBrains Mono + Inter), type scale, 12px floor rule
- Layout structure (sidebar w-60, page padding, card/modal sizing)
- Component architecture (no new components in this pass)
- Dark-first principle, uppercase conventions, brutalist restraint

### What changes
- Colour palette: higher saturation, wider luminance range, richer status colours
- Border/surface contrast: more visible structural lines
- Glow and luminance effects: phosphor-style glow on active/live elements
- Atmospheric textures: subtle grid/scan-line patterns on surfaces
- Animation enrichment: pulsing, breathing effects for "liveness"

---

## 1. Colour Palette Evolution

The current palette is built around feldgrau at ~12% saturation. The proposal shifts the *accent layer* to ~25-30% saturation — still military, still grey-green, but with enough chroma to register as a deliberate colour rather than a grey that happens to lean green. Backgrounds stay near-black.

### 1.1 Backgrounds — Deeper separation

The current background tiers (`#070808` → `rgba(11,13,12,0.5)` → `#0a0c0b`) are almost indistinguishable. Increase the step between tiers for clearer visual hierarchy.

| Token | Current | Proposed | Rationale |
|-------|---------|----------|-----------|
| `--bg-primary` | `#070808` | `#06080a` | Slight cool shift (towards deep blue-black rather than neutral black). Aligns with CRT/monitor housing colour. |
| `--bg-surface` | `rgba(11,13,12,0.5)` | `rgba(12,16,14,0.55)` | Slightly brighter, more visible card/panel separation |
| `--bg-surface-hover` | `rgba(11,13,12,0.7)` | `rgba(14,20,17,0.75)` | Clearer hover feedback |
| `--bg-elevated` | `#0a0c0b` | `#0c100e` | More distinct elevation layer (modals, dropdowns read as "above" the surface) |

**Net effect:** Backgrounds remain near-black but the *steps* between tiers are wider, creating depth without brightening the overall canvas.

### 1.2 Accent — From feldgrau to phosphor-feldgrau

The core identity shift. Push saturation from ~12% to ~25% and brightness up, evoking phosphor-green CRT readouts while staying in the grey-green family.

| Token | Current | Proposed | Rationale |
|-------|---------|----------|-----------|
| `--accent` | `#4d5d53` | `#4a7a5c` | Base accent: richer green, still grey-tempered. Reads as "operational green" not "grass green." |
| `--accent-hover` | `#5a6e62` | `#5a9068` | Brighter on hover — the phosphor intensifies |
| `--accent-bright` | `#6e8578` | `#6aad7a` | High-emphasis: headings, active nav indicators, key data points. Now visibly green. |
| `--accent-subtle` | `rgba(77,93,83,0.1)` | `rgba(74,122,92,0.12)` | Tinted backgrounds carry more green cast |
| `--accent-border` | `rgba(77,93,83,0.3)` | `rgba(74,122,92,0.25)` | Slightly reduced opacity since the base is more saturated — net visual weight stays similar |

**Phosphor glow token (new):**
| Token | Value | Use |
|-------|-------|-----|
| `--accent-glow` | `rgba(74,122,92,0.15)` | Box-shadow / drop-shadow for active elements, scan effects |

### 1.3 Action buttons — Bolder CTA

The current action green (`#3b8a5a`) is already the brightest element in the UI but sits close to the proposed new accent range. Push it slightly brighter and warmer to maintain separation.

| Token | Current | Proposed | Rationale |
|-------|---------|----------|-----------|
| `--action` | `#3b8a5a` | `#38a85c` | Brighter, slightly warmer green. Unmistakable as a CTA against the new accent range. |
| `--action-hover` | `#449e66` | `#42be68` | Clear hover escalation |

### 1.4 Text — Crisper contrast

| Token | Current | Proposed | Rationale |
|-------|---------|----------|-----------|
| `--text-primary` | `#e6eae8` | `#e8ede9` | Very slight brightness bump. Near-white with cool undertone. |
| `--text-secondary` | `#8a938e` | `#8a9e92` | Slightly greener undertone — secondary text picks up the phosphor cast |
| `--text-muted` | `#576058` | `#4e6556` | Slightly more saturated muted text — still clearly "dimmed" but greener |

### 1.5 Borders — More visible structure

The most impactful single change. Current borders at 14% opacity are nearly invisible. Increasing to 20% makes the grid structure legible — the "interface chrome" of a control room.

| Token | Current | Proposed | Rationale |
|-------|---------|----------|-----------|
| `--border` | `rgba(77,93,83,0.14)` | `rgba(74,122,92,0.18)` | Visible structural lines. Panels, cards, dividers now clearly delineate space. |
| `--border-hover` | `rgba(77,93,83,0.25)` | `rgba(74,122,92,0.32)` | Hover borders are distinctly brighter |

### 1.6 Status colours — Richer signal

Status colours are the "warning lights on the control panel." Currently desaturated to match feldgrau's restraint — but in a mission-control context, status indicators should be the brightest things on screen.

| Token | Current | Proposed | Rationale |
|-------|---------|----------|-----------|
| `--status-ok` | `#4d5d53` | `#4a7a5c` | Aligned with new accent base (operational green) |
| `--status-warn` | `#c4a84a` | `#d4a832` | Richer amber. Reads as a caution light, not a dim bulb. |
| `--status-critical` | `#c45a4a` | `#d44a3a` | Hotter red. Unmistakable urgency. |
| `--status-info` | `#4a8aae` | `#4a92c4` | Brighter steel blue. CRT-cyan territory. |

**Status glow (new):** Each status colour should have a matching glow token for use in box-shadows on active indicators:

```css
--status-ok-glow: rgba(74, 122, 92, 0.2);
--status-warn-glow: rgba(212, 168, 50, 0.2);
--status-critical-glow: rgba(212, 74, 58, 0.2);
--status-info-glow: rgba(74, 146, 196, 0.2);
```

---

## 2. Glow & Luminance Effects

The defining visual signature of retro-futurism: data *emits light*. Active elements glow as if back-lit by phosphor.

### 2.1 Active glow upgrade

Current `.glow-active` is barely visible (`box-shadow: 0 0 15px rgba(77,93,83,0.06)`). Increase to a visible phosphor halo:

```css
.glow-active {
  box-shadow: 0 0 20px var(--accent-glow), 0 0 4px var(--accent-glow);
}
```

### 2.2 Status indicator pulse

Currently only the "continuous mode" dot pulses. Extend the glow-pulse to all active status indicators (connection health, monitoring state):

```css
.glow-pulse {
  animation: glowPulse 3s ease-in-out infinite;
}

@keyframes glowPulse {
  0%, 100% { box-shadow: 0 0 4px var(--accent-glow); }
  50% { box-shadow: 0 0 12px var(--accent-glow), 0 0 4px var(--accent-glow); }
}
```

The 3-second cycle is slower than the current `animate-pulse` (2s) — monitoring systems breathe slowly.

### 2.3 Sidebar brand glow

The "HEIMDALL" wordmark in the sidebar currently has a static green dot. Add a subtle text-shadow glow to the wordmark itself:

```css
.brand-glow {
  text-shadow: 0 0 20px var(--accent-glow);
}
```

This creates the effect of phosphor characters on a CRT — the text appears to emit faint light.

### 2.4 Focus ring glow

Upgrade the current `2px solid var(--accent)` focus outline to include a glow halo:

```css
:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
  box-shadow: 0 0 8px var(--accent-glow);
}
```

---

## 3. Atmospheric Textures

### 3.1 Subtle grid overlay

Add an optional CSS-only grid pattern on `--bg-surface` cards, evoking graph paper or a radar screen's coordinate grid:

```css
.surface-grid {
  background-image:
    linear-gradient(var(--border) 1px, transparent 1px),
    linear-gradient(90deg, var(--border) 1px, transparent 1px);
  background-size: 24px 24px;
}
```

**Usage:** Applied sparingly — the Activity feed container, the Dashboard overview cards, and the Agent Chat background. Not every surface. The grid should be barely perceptible at normal viewing distance but visible on close inspection.

### 3.2 Scan-line effect

A very subtle horizontal line pattern on elevated surfaces (modals, dropdowns), evoking CRT scan lines:

```css
.surface-scanlines {
  background-image:
    repeating-linear-gradient(
      0deg,
      transparent,
      transparent 2px,
      rgba(74, 122, 92, 0.02) 2px,
      rgba(74, 122, 92, 0.02) 4px
    );
  pointer-events: none;
}
```

Applied as a `::after` pseudo-element overlay so it doesn't interfere with content. Extremely subtle — 2% opacity.

### 3.3 Corner chrome / interface brackets

For key panels (Dashboard overview, Agent Config summary), add corner bracket decorations using `::before`/`::after` pseudo-elements:

```css
.chrome-brackets {
  position: relative;
}
.chrome-brackets::before,
.chrome-brackets::after {
  content: '';
  position: absolute;
  width: 12px;
  height: 12px;
  border-color: var(--accent);
  pointer-events: none;
}
.chrome-brackets::before {
  top: -1px;
  left: -1px;
  border-top: 1px solid;
  border-left: 1px solid;
}
.chrome-brackets::after {
  bottom: -1px;
  right: -1px;
  border-bottom: 1px solid;
  border-right: 1px solid;
}
```

This is the visual grammar of HUD overlays and targeting reticles — two corner marks that frame the content, suggesting a viewport or scan region.

---

## 4. Animation Enrichment

### 4.1 Data stream typing effect

Upgrade the current `animate-reveal` (clip-path wipe) with a green phosphor flash on reveal:

```css
@keyframes reveal {
  from {
    opacity: 0;
    clip-path: inset(0 100% 0 0);
    text-shadow: 0 0 8px var(--accent-glow);
  }
  70% {
    opacity: 1;
    clip-path: inset(0 0 0 0);
    text-shadow: 0 0 8px var(--accent-glow);
  }
  to {
    opacity: 1;
    clip-path: inset(0 0 0 0);
    text-shadow: none;
  }
}
```

The text "flares" green as it appears, then settles to normal — like a CRT character being written.

### 4.2 Card entrance stagger

Current fade-in is fine but add a subtle border-flash on entry:

```css
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(4px);
    border-color: var(--accent);
  }
  to {
    opacity: 1;
    transform: translateY(0);
    border-color: var(--border);
  }
}
```

Cards briefly flash their borders at full accent colour on entry, then fade to the structural border — like instruments coming online.

### 4.3 Scrollbar glow on interaction

```css
::-webkit-scrollbar-thumb:active {
  background: rgba(74, 122, 92, 0.5);
  box-shadow: 0 0 6px rgba(74, 122, 92, 0.3);
}
```

---

## 5. Component-Level Changes

### 5.1 Sidebar

- Wordmark: Add `.brand-glow` text-shadow
- Active nav indicator: Change from `w-0.5 bg-accent` bar to `w-0.5 bg-accent-bright` with a glow shadow (`box-shadow: 0 0 6px var(--accent-glow)`)
- Section headers ("INFRASTRUCTURE", "AGENT", etc.): Currently `text-text-muted` — change to `text-accent` so section labels carry the phosphor-green cast
- App selector: Add subtle `.surface-grid` background to the dropdown

### 5.2 Dashboard cards

- Add `.chrome-brackets` to the overview stat cards (monitoring status, connections count, recent activity)
- Apply `.surface-grid` background to the card container
- Status dots: Add `.glow-pulse` to active status indicators

### 5.3 Activity feed

- Agent-sourced entries: Increase the hover background from `hover:bg-accent-subtle/80` to `hover:bg-accent-subtle` (slightly more visible)
- Severity badges: Use the richer status colours with glow on `critical` entries
- Feed container: Apply the grid background texture

### 5.4 Agent chat

- Chat container: Apply `.surface-scanlines` overlay for CRT atmosphere
- Agent messages: Use the enhanced `animate-reveal` with phosphor flash
- "Thinking..." indicator: Add `.glow-pulse` effect

### 5.5 StatusBadge

- Active state: Add `box-shadow` using the matching `--status-*-glow` token
- The status dot: Add a subtle pulse animation for `active` state (already exists, but increase glow intensity)

### 5.6 Modals

- Apply `.surface-scanlines` as a `::after` overlay
- Header border: Use `border-accent` instead of `border-border` for more definition

---

## 6. Implementation Plan

### Phase 1 — Token update (low risk, high impact)
Update `:root` variables and `@theme` block in `main.css`. This propagates everywhere automatically via Tailwind utilities.

**Files:** `frontend/src/assets/styles/main.css`
**Estimated scope:** ~40 lines changed

### Phase 2 — Glow effects & animations
Add new CSS classes (`.glow-active`, `.glow-pulse`, `.brand-glow`, enhanced keyframes) to `main.css`.

**Files:** `frontend/src/assets/styles/main.css`
**Estimated scope:** ~60 lines added

### Phase 3 — Atmospheric textures
Add `.surface-grid`, `.surface-scanlines`, `.chrome-brackets` CSS classes.

**Files:** `frontend/src/assets/styles/main.css`
**Estimated scope:** ~50 lines added

### Phase 4 — Component application
Apply new classes to specific components. Each is a small, targeted edit.

**Files:**
- `components/common/AppSidebar.vue` — brand glow, nav indicator, section header colour
- `components/common/StatusBadge.vue` — status glow shadows
- `components/log/LogEntry.vue` — richer hover, severity glow
- `components/log/LogFeed.vue` — grid background
- `pages/DashboardPage.vue` — chrome brackets, grid, status pulse
- `pages/AgentChatPage.vue` — scanlines, reveal enhancement
- `layouts/DefaultLayout.vue` — (no changes expected)

### Phase 5 — Brand guidelines update
Update `docs/brand-guidelines.md` to reflect the new palette and design vocabulary.

---

## 7. Design Principles — What This Respects

| Principle | How this proposal respects it |
|-----------|-------------------------------|
| **Dark-first** | Backgrounds get cooler, not brighter. The canvas remains near-black. |
| **Monochrome + one accent** | Still one accent family (green). Saturation increases but no new hue introduced. Status colours are functional, not decorative. |
| **Typography over ornament** | No new decorative elements. Chrome brackets and grid textures are *structural*, reinforcing the control-panel metaphor. |
| **Brutalist restraint** | Textures are 2% opacity. Glows are subtle. Nothing is gratuitous. The atmosphere comes from *many subtle things together*, not one loud effect. |
| **Confident spacing** | No spacing changes. The enhanced colours and glows work within the existing spatial framework. |

---

## 8. What This Intentionally Does NOT Do

- **No new fonts.** JetBrains Mono and Inter are correct for this aesthetic.
- **No layout changes.** The sidebar, page structure, and component sizing stay identical.
- **No new colours outside the green family.** We don't add cyan, amber, or magenta accents — the status colours already provide those signals. The accent evolution stays within grey-green → phosphor-green.
- **No Three.js / WebGL.** All atmospheric effects are CSS-only. The existing HeroMesh canvas component is untouched.
- **No public/marketing page changes.** This proposal covers the authenticated application UI only. The landing page has its own visual treatment via HeroMesh.

---

## 9. Questions

1. **Saturation ceiling:** The proposed accent shift goes from ~12% to ~25% saturation. Should we go further (towards 35-40%, closer to true phosphor green `#00ff88` territory) or is the middle ground right? There's a spectrum between "military" and "sci-fi" and the right stop depends on taste.

2. **Grid texture scope:** The `.surface-grid` pattern could be applied to just key containers (Activity, Dashboard) or more broadly (all cards, all panels). Broader application creates a stronger mission-control feel but risks visual noise. Preference?

3. **Scan-line effect:** This is the most "stylistic" element in the proposal. It's extremely subtle (2% opacity), but some may find it gimmicky. Include or skip?

4. **Corner chrome brackets:** Same question — these are atmospheric flourishes. They add HUD character but are purely decorative. Worth the visual weight?

5. **Status glow on critical entries:** Should critical-severity items in the Activity feed get a red glow halo (like a warning light), or does that risk feeling alarming in a way that's distracting rather than informative?

6. **Animated reveal flash:** The phosphor-flash on agent message reveal is a micro-detail. It reinforces the CRT metaphor but adds animation complexity. Worth it?

7. **Brand guidelines:** Should `docs/brand-guidelines.md` be updated as part of this work, or kept as the "original spec" with a separate "retro-futurism addendum"?
