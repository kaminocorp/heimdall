# Neutral Canvas Colour Rebalance

**Version:** 0.41.0  
**Date:** 2026-04-13  
**Summary:** Shifted the entire UI from a green-tinted-everything aesthetic to a neutral dark canvas where green appears only on intentional interactive and brand elements.

---

## The Problem

Heimdall's "Technical Retro-Futurism" design system (v0.35.0) applied green tint to every visual layer: backgrounds, body text, borders, scrollbars, text selection, surface textures, scanlines, and fade-in animations. The result felt like wearing green-tinted glasses — the brand colour lost its impact because there was no neutral surface for the eye to rest on.

Sister platform Trajan demonstrates the correct approach: orange (its brand colour) appears only on badges, active tabs, and small interactive highlights. Surfaces, borders, and text are all neutral grays. The accent colour punctuates; it doesn't permeate.

## The Approach

**Neutral canvas, surgical green.** Make every ambient layer (backgrounds, text, borders, textures) completely neutral, then let green appear only where it communicates something — action buttons, active states, status indicators, brand elements, and decorative HUD chrome.

Green coverage drops from ~80% of pixels to ~5-10%. When it appears, it means something.

---

## Token Changes (`main.css` `:root` block)

### Backgrounds — removed green tint

| Token | Before | After |
|-------|--------|-------|
| `--bg-primary` | `#06080a` (slight green) | `#08090a` (pure neutral) |
| `--bg-surface` | `rgba(12, 16, 14, 0.55)` (green cast) | `rgba(15, 16, 18, 0.55)` (neutral) |
| `--bg-surface-hover` | `rgba(14, 20, 17, 0.75)` (green cast) | `rgba(20, 21, 24, 0.75)` (neutral) |
| `--bg-elevated` | `#0c100e` (green-black) | `#0e0f11` (neutral dark) |

### Text — removed green undertone

| Token | Before | After |
|-------|--------|-------|
| `--text-primary` | `#e8ede9` (green-tinted white) | `#e2e4e8` (cool neutral white) |
| `--text-secondary` | `#8a9e92` (muted green) | `#8a8f96` (cool slate) |
| `--text-muted` | `#4e6556` (obvious green) | `#4a4f56` (dark slate) |

### Borders — neutral structural lines

| Token | Before | After |
|-------|--------|-------|
| `--border` | `rgba(74, 122, 92, 0.18)` (green) | `rgba(140, 145, 155, 0.14)` (neutral gray) |
| `--border-hover` | `rgba(74, 122, 92, 0.32)` (green) | `rgba(140, 145, 155, 0.26)` (neutral gray) |

### Accent — kept green, slightly tightened

| Token | Before | After |
|-------|--------|-------|
| `--accent-subtle` | `rgba(74, 122, 92, 0.12)` | `rgba(74, 122, 92, 0.10)` (slightly reduced) |
| `--accent-border` | `rgba(74, 122, 92, 0.25)` | `rgba(74, 122, 92, 0.22)` (slightly reduced) |

All other accent/action/status tokens are **unchanged** — the brand green itself is not the problem, its overuse was.

---

## Global Style Changes (`main.css`)

### Scrollbar thumb — neutral

| State | Before | After |
|-------|--------|-------|
| Default | `rgba(74, 122, 92, 0.15)` | `rgba(140, 145, 155, 0.12)` |
| Hover | `rgba(74, 122, 92, 0.3)` | `rgba(140, 145, 155, 0.25)` |
| Active | `rgba(74, 122, 92, 0.5)` + green glow | `rgba(140, 145, 155, 0.40)` (no glow) |

### Text selection — neutral

Before: `rgba(74, 122, 92, 0.3)` → After: `rgba(140, 145, 155, 0.25)`

### Surface grid texture — neutral

Before: `rgba(74, 122, 92, 0.06)` gridlines → After: `rgba(140, 145, 155, 0.04)` (also slightly reduced opacity)

### Scanline overlay — neutral

Before: `rgba(74, 122, 92, 0.02)` → After: `rgba(140, 145, 155, 0.02)`

### Fade-in animation — removed green flash

The `fadeIn` keyframe previously transitioned `border-color` from green (`var(--accent)`) to `var(--border)`. This green flash on every list item entry is removed — the animation now only handles opacity and transform.

---

## AgentNebula Shader Rebalance (`AgentNebula.vue`)

The 3D particle nebula on the Connections page had hardcoded GLSL colour values that were heavily green-dominant. All three particle layers were rebalanced:

### Layer 1 — Ambient Cloud (3,000 particles)

| Colour | Before | After |
|--------|--------|-------|
| Base | `vec3(0.10, 0.20, 0.14)` (dark green) | `vec3(0.10, 0.11, 0.13)` (neutral slate) |
| Accent | `vec3(0.15, 0.35, 0.22)` (deep green) | `vec3(0.14, 0.30, 0.20)` (muted green — reduced) |
| Phosphor | `vec3(0.22, 0.50, 0.30)` (bright green) | `vec3(0.18, 0.42, 0.26)` (toned down) |
| Warm amber | `vec3(0.38, 0.30, 0.10)` | `vec3(0.30, 0.24, 0.12)` (subtler) |

Mix weights reduced: green mixes from 0.50/0.45 to 0.35/0.30, teal and amber increased.

Dormant colour shifted from green-grey `vec3(0.10, 0.13, 0.11)` to neutral `vec3(0.10, 0.10, 0.11)`.

### Layer 2 — Wisp Tendrils (1,000 particles)

Same pattern: base shifted to neutral slate `vec3(0.08, 0.09, 0.11)`, green accent reduced, teal given more weight. Dormant colour neutralised.

### Layer 3 — Core Motes (400 particles)

Candle phosphor toned down from `vec3(0.20, 0.40, 0.26)` to `vec3(0.18, 0.32, 0.22)`. Moonlit shifted from green `vec3(0.16, 0.34, 0.30)` to cool slate `vec3(0.14, 0.24, 0.30)`. Dormant colour neutralised.

**Net effect:** The nebula now reads as a moody, multi-tonal cloud that occasionally catches green highlights — rather than a green fog.

---

## What Stays Green (by design)

These elements retain their green colour because green is meaningful here — it signals brand, action, or status:

- **`--accent` / `--action`** — CTA buttons, active nav items, selected states
- **`--status-ok`** — healthy/connected status indicators
- **`--accent-glow`** / glow-pulse animations — brand atmospheric effects
- **Chrome brackets** (`.chrome-brackets`) — HUD corner decoration, uses `var(--accent)`
- **Brand wordmark glow** (`.brand-glow`) — uses `--accent-glow`
- **FlowLines** on Connections page — uses `var(--accent)` for animated connection lines

---

## Files Changed

| File | Kind | Change |
|------|------|--------|
| `frontend/src/assets/styles/main.css` | Edit | All token + global style changes above |
| `frontend/src/components/connections/AgentNebula.vue` | Edit | GLSL shader palette rebalance (3 layers) |

## Verification

| Check | Result |
|-------|--------|
| `vue-tsc --noEmit` | Clean |
| `vite build` | Clean |
| `vitest run` | 52/52 pass |
